package pammy

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/noppawitt/dca"
	"github.com/noppawitt/pammy/youtube"
)

var (
	ErrNotInVoiceChannel        = errors.New("not in voice channel")
	ErrEmptyTracks              = errors.New("empty tracks")
	ErrTrackNotFound            = errors.New("track not found")
	ErrCannotRemovePlayingTrack = errors.New("cannot remove the playing track")
)

type Track struct {
	ID       string
	Name     string
	Duration time.Duration
}

type BotState uint

const (
	BotStateWaitForTrack BotState = iota
	BotStatePlaying
	BotStatePaused
)

type Bot struct {
	mu sync.RWMutex

	guidID         string
	voiceChannelID string
	textChannelID  string

	tracks                []*Track
	currentTrackIdx       int
	state                 BotState
	autoDiscoverNextTrack bool

	errCh  chan error
	skipCh chan chan struct{}
	stopCh chan struct{}

	ytClient   *youtube.Client
	dg         *discordgo.Session
	vc         *discordgo.VoiceConnection
	streamSess *dca.StreamingSession
	encodeSess *dca.EncodeSession
}

func NewBot(guidID string, dg *discordgo.Session, ytClient *youtube.Client, errCh chan error) *Bot {
	return &Bot{
		guidID:          guidID,
		tracks:          nil,
		currentTrackIdx: 0,
		state:           BotStateWaitForTrack,
		errCh:           errCh,
		skipCh:          make(chan chan struct{}),
		stopCh:          make(chan struct{}),
		ytClient:        ytClient,
		dg:              dg,
	}
}

func (b *Bot) JoinVoiceChannel(channelID string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	vc, err := b.dg.ChannelVoiceJoin(b.guidID, channelID, false, true)
	if err != nil {
		return err
	}

	b.voiceChannelID = channelID
	b.vc = vc

	return nil
}

func (b *Bot) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.state != BotStateWaitForTrack {
		b.stopCh <- struct{}{}
	}

	close(b.skipCh)
	close(b.stopCh)

	if b.vc != nil {
		b.vc.Disconnect()
		b.vc.Close()
	}

	if b.errCh != nil {
		close(b.errCh)
	}
}

func (b *Bot) play() {
	if b.vc == nil {
		b.sendError(ErrNotInVoiceChannel)
		return
	}

	b.mu.Lock()
	b.state = BotStatePlaying
	b.mu.Unlock()
	b.vc.Speaking(true)
	defer func() {
		b.mu.Lock()
		b.state = BotStateWaitForTrack
		b.mu.Unlock()
		b.vc.Speaking(false)
	}()

	for b.currentTrackIdx < len(b.tracks) {
		var err error

		video, err := b.ytClient.GetVideo(b.tracks[b.currentTrackIdx].ID)
		if err != nil {
			b.sendError(err)
		}

		streamURL, err := b.ytClient.GetAudioStreamURL(video)
		if err != nil {
			b.sendError(err)
		}

		_, sendErr := b.dg.ChannelMessageSend(b.textChannelID, fmt.Sprintf("Playing `%s`", video.Title))
		if sendErr != nil {
			log.Println(sendErr)
		}

		var skipped chan struct{}
		if err != nil {
			b.dg.ChannelMessageSend(b.textChannelID, fmt.Sprintf("Cannot play `%s`, skipping...", video.Title))
		} else {
			dca.Logger = log.New(ioutil.Discard, "", 0)
			b.encodeSess, err = dca.EncodeFile(streamURL, dca.StdEncodeOptions)
			if err != nil {
				b.sendError(err)
				return
			}

			done := make(chan error)
			b.streamSess = dca.NewStream(b.encodeSess, b.vc, done)

			select {
			case err := <-done:
				b.stop()
				if err != nil && err != io.EOF {
					b.sendError(err)
					return
				}
			case skipped = <-b.skipCh:
				b.stop()
			case <-b.stopCh:
				b.stop()
				return
			}
		}

		if b.currentTrackIdx == len(b.tracks)-1 && b.autoDiscoverNextTrack {
			err = b.discoverNextTrack()
			if err != nil {
				log.Println("cannot discover next track: ", err)
			}
		}

		b.mu.Lock()
		b.currentTrackIdx++
		b.mu.Unlock()

		if skipped != nil {
			skipped <- struct{}{}
		}
	}
}

func (b *Bot) stop() {
	b.streamSess.SetPaused(true)
	b.encodeSess.Stop()
}

func (b *Bot) Pause() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.state != BotStatePlaying {
		return errors.New("cannot pause: no music is playing")
	}

	b.state = BotStatePaused
	b.streamSess.SetPaused(true)

	return nil
}

func (b *Bot) Resume() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.state != BotStatePaused {
		return errors.New("cannot resume: music is playing")
	}

	b.state = BotStatePlaying
	b.streamSess.SetPaused(false)
	return nil
}

func (b *Bot) Remove(idx int) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if idx < 0 || idx >= len(b.tracks) {
		return ErrTrackNotFound
	}

	if b.currentTrackIdx == idx {
		return ErrCannotRemovePlayingTrack
	}

	b.tracks = append(b.tracks[:idx], b.tracks[idx+1:]...)

	return nil
}

func (b *Bot) Next(n int) error {
	return b.GoTo(b.currentTrackIdx + n)
}

func (b *Bot) Prev(n int) error {
	return b.GoTo(b.currentTrackIdx - n)
}

func (b *Bot) GoTo(idx int) error {
	b.mu.Lock()

	if len(b.tracks) == 0 {
		return ErrEmptyTracks
	}

	if idx < 0 {
		idx = 0
	}

	if idx > len(b.tracks) {
		idx = len(b.tracks)
	}

	b.currentTrackIdx = idx

	if b.state == BotStateWaitForTrack {
		if idx == len(b.tracks) {
			if b.autoDiscoverNextTrack {
				return b.discoverNextTrack()
			}
		} else {
			go b.play()
		}
		return nil
	}

	// Skipping in playing state is done by:
	// 1. Set the current track index to the one before the wanted track.
	// 2. Send skip signal to termninate the current playing track then wait the playloop to iterate to the next track.
	b.currentTrackIdx--
	done := make(chan struct{})
	b.skipCh <- done

	b.mu.Unlock()

	<-done
	close(done)

	return nil
}

func (b *Bot) Add(track *Track) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.tracks = append(b.tracks, track)
	if b.state == BotStateWaitForTrack {
		go b.play()
	}
}

// Clear clears the upcomming tracks (tracks queue) then return a total removed tracks
// Set all to true to clear all tracks except the current playing track.
func (b *Bot) Clear(all bool) int {
	b.mu.Lock()
	defer b.mu.Unlock()

	if len(b.tracks) == 0 {
		return 0
	}

	var total int

	if all {
		if b.state == BotStateWaitForTrack {
			// remove all tracks
			total = len(b.tracks)
			b.tracks = nil
		} else {
			// remove all tracks except the current playing track
			total = len(b.tracks) - 1
			b.tracks = b.tracks[b.currentTrackIdx : b.currentTrackIdx+1]
		}
	} else {
		if b.state == BotStateWaitForTrack {
			// no queue
			return 0
		}
		// remove upcomming tracks
		total = len(b.tracks[b.currentTrackIdx+1:])
		b.tracks = b.tracks[b.currentTrackIdx:]
	}

	b.currentTrackIdx = 0

	return total
}

func (b *Bot) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.state != BotStateWaitForTrack {
		b.stopCh <- struct{}{}
	}

	b.currentTrackIdx = 0
	b.state = BotStateWaitForTrack
	b.tracks = nil
	b.autoDiscoverNextTrack = false
}

func (b *Bot) VoiceChannelID() string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.voiceChannelID
}

func (b *Bot) TextChannelID() string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.textChannelID
}

func (b *Bot) SetTextChannelID(id string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.textChannelID = id
}

func (b *Bot) State() BotState {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.state
}

func (b *Bot) CurrentTrackIndex() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.currentTrackIdx
}

func (b *Bot) TotalTracks() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.tracks)
}

func (b *Bot) sendError(err error) {
	if b.errCh != nil {
		b.errCh <- err
	}
}

type TrackPage struct {
	TrackInfos  []string
	Page        int
	PageSize    int
	TotalPage   int
	TotalTracks int
	TotalQueued int
}

func (tp TrackPage) DisplayText() string {
	if len(tp.TrackInfos) == 0 {
		return "No tracks"
	}

	s := fmt.Sprintf("Total tracks: %d (%d queued)\n```", tp.TotalTracks, tp.TotalQueued)

	for _, info := range tp.TrackInfos {
		s += info + "\n"
	}

	s += fmt.Sprintf("```Page %d of %d", tp.Page, tp.TotalPage)

	return s
}

func (b *Bot) List(page, pageSize int) TrackPage {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if pageSize == 0 {
		pageSize = 10
	}

	if page == 0 {
		// display the page with the playing track
		page = (b.currentTrackIdx / pageSize) + 1
	}

	start := (page - 1) * pageSize
	end := start + pageSize
	if end > len(b.tracks) {
		end = len(b.tracks)
	}

	var infos []string
	for i := start; i < end; i++ {
		info := fmt.Sprintf("%d %s  %s", i+1, b.tracks[i].Name, b.tracks[i].Duration)
		if i == b.currentTrackIdx {
			info += "  [Playing]"
		}
		infos = append(infos, info)
	}

	trackPage := TrackPage{
		TrackInfos:  infos,
		Page:        page,
		PageSize:    pageSize,
		TotalPage:   (len(b.tracks) / pageSize) + 1,
		TotalTracks: len(b.tracks),
	}

	if b.state == BotStateWaitForTrack {
		trackPage.TotalQueued = 0
	} else {
		trackPage.TotalQueued = len(b.tracks[b.currentTrackIdx+1:])
	}

	return trackPage
}

func (b *Bot) AutoDiscoverNextTrack() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.autoDiscoverNextTrack
}

func (b *Bot) SetAutoDiscoverNextTrack(v bool) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.autoDiscoverNextTrack = v
}

func (b *Bot) discoverNextTrack() error {
	b.dg.ChannelMessageSend(b.textChannelID, "Discovering next music...")

	lastTrackIdx := len(b.tracks) - 1
	if lastTrackIdx < 0 {
		return ErrEmptyTracks
	}

	videos, err := b.ytClient.GetSuggestedVideos(b.tracks[lastTrackIdx].ID)
	if err != nil {
		return err
	}

	if len(videos) == 0 {
		return errors.New("no tracks discovered")
	}

	// TODO: improve suggested track selection
	video := videos[expRandInt(len(videos))]

	track := &Track{
		ID:       video.ID,
		Name:     video.Title,
		Duration: video.Duration,
	}

	go b.Add(track)

	return nil
}

func expRandInt(n int) int {
	x := -rand.Float64()
	// y is in range (0,1]
	y := math.Pow(math.E, 4*x)
	return int(math.Floor(y * float64(n)))
}
