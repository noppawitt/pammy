package pammy

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
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
	ID     string
	Name   string
	Length time.Duration
}

type BotState uint

const (
	BotStateWaitForTrack BotState = iota
	BotStatePlaying
	BotStatePaused
)

type Bot struct {
	mu sync.RWMutex

	guidID    string
	channelID string

	tracks          []*Track
	currentTrackIdx int
	state           BotState
	errCh           chan error
	skipCh          chan struct{}
	stopCh          chan struct{}

	youtubeClient *youtube.Client
	dg            *discordgo.Session
	vc            *discordgo.VoiceConnection
	streamSess    *dca.StreamingSession
	encodeSess    *dca.EncodeSession
}

func NewBot(guidID string, dg *discordgo.Session, youtubeClient *youtube.Client, errCh chan error) *Bot {
	return &Bot{
		guidID:          guidID,
		tracks:          nil,
		currentTrackIdx: 0,
		state:           BotStateWaitForTrack,
		errCh:           errCh,
		skipCh:          make(chan struct{}),
		stopCh:          make(chan struct{}),
		youtubeClient:   youtubeClient,
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

	b.channelID = channelID
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
	}

	b.state = BotStatePlaying
	b.vc.Speaking(true)
	defer func() {
		b.state = BotStateWaitForTrack
		b.vc.Speaking(false)
	}()

	for b.currentTrackIdx < len(b.tracks) {
		b.mu.Lock()

		video, err := b.youtubeClient.GetVideo(b.tracks[b.currentTrackIdx].ID)
		if err != nil {
			b.sendError(err)
			return
		}

		streamURL, err := b.youtubeClient.GetStreamURL(video, video.Formats.FindByItag(249))
		if err != nil {
			b.sendError(err)
			return
		}

		dca.Logger = log.New(ioutil.Discard, "", 0)
		b.encodeSess, err = dca.EncodeFile(streamURL, dca.StdEncodeOptions)
		if err != nil {
			b.sendError(err)
			return
		}

		done := make(chan error)
		b.streamSess = dca.NewStream(b.encodeSess, b.vc, done)

		b.mu.Unlock()

		select {
		case err = <-done:
			b.stop()
			b.mu.Lock()
			if err != nil && err != io.EOF {
				b.sendError(err)
				return
			}
			b.currentTrackIdx++
			b.mu.Unlock()
		case <-b.skipCh:
			b.stop()
		case <-b.stopCh:
			b.stop()
			return
		}
	}
}

func (b *Bot) stop() {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.streamSess.SetPaused(true)
	b.encodeSess.Cleanup()
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
	defer b.mu.Unlock()

	if len(b.tracks) == 0 {
		return ErrEmptyTracks
	}

	if idx < 0 {
		idx = 0
	}

	if idx >= len(b.tracks) {
		idx = len(b.tracks)
	}

	b.currentTrackIdx = idx
	b.skipCh <- struct{}{}

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

func (b *Bot) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.state != BotStateWaitForTrack {
		b.stopCh <- struct{}{}
	}

	b.currentTrackIdx = 0
	b.state = BotStateWaitForTrack
	b.tracks = nil
}

func (b *Bot) ChannelID() string {
	return b.channelID
}

func (b *Bot) State() BotState {
	return b.state
}

func (b *Bot) CurrentTrackIndex() int {
	return b.currentTrackIdx
}

func (b *Bot) TotalTracks() int {
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
}

func (tp TrackPage) DisplayText() string {
	if len(tp.TrackInfos) == 0 {
		return "No tracks"
	}

	s := fmt.Sprintf("Queue (%d tracks)\n```", tp.TotalTracks)

	for _, info := range tp.TrackInfos {
		s += info + "\n"
	}

	s += fmt.Sprintf("```Page %d of %d", tp.Page, tp.TotalPage)

	return s
}

func (b *Bot) List(page, pageSize int) TrackPage {
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
		info := fmt.Sprintf("%d %s  %s", i+1, b.tracks[i].Name, b.tracks[i].Length)
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

	return trackPage
}
