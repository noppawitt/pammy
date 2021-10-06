package pammy

import (
	"fmt"
	"net/url"

	"github.com/bwmarrin/discordgo"
	"github.com/noppawitt/pammy/youtube"
)

type AddCommand struct {
	hub      *Hub
	ytClient *youtube.Client
}

func NewAddCommand(hub *Hub, ytClient *youtube.Client) *AddCommand {
	return &AddCommand{
		hub:      hub,
		ytClient: ytClient,
	}
}

func (c *AddCommand) Command() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        "add",
		Description: "Add music from youtube",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "search-term",
				Description: "Search term or Youtube URL",
				Required:    true,
			},
		},
	}
}

func (c *AddCommand) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) {
	respondText(s, i.Interaction, "Pammy is thinking...")

	vs := userVoiceState(s.State, i.GuildID, i.Member.User.ID)
	if vs == nil {
		updateResponse(s, i.Interaction, "You need to join voice channel first")
		return
	}

	bot, ok := c.hub.GetBot(i.GuildID)
	if !ok {
		bot = NewBot(i.GuildID, s, c.ytClient, nil)
		c.hub.SetBot(bot, i.GuildID)
	}

	if bot.ChannelID() != "" && bot.ChannelID() != vs.ChannelID {
		updateResponse(s, i.Interaction, "Pammy is singing in other voice channel")
		return
	}

	query := i.ApplicationCommandData().Options[0].StringValue()

	var videoURL = query

	_, err := url.ParseRequestURI(query)
	if err != nil {
		result, err := c.ytClient.SearchOne(query)
		if err != nil {
			updateResponse(s, i.Interaction, "Cannot get result for: "+query)
			return
		}

		videoURL = result.ID
	}

	video, err := c.ytClient.GetVideo(videoURL)
	if err != nil {
		updateResponse(s, i.Interaction, "Cannot get video")
		return
	}

	track := &Track{
		ID:     video.ID,
		Name:   video.Title,
		Length: video.Duration,
	}

	if bot.ChannelID() == "" {
		err := bot.JoinVoiceChannel(vs.ChannelID)
		if err != nil {
			updateResponse(s, i.Interaction, "Cannot join voice channel")
			return
		}
	}

	bot.Add(track)

	updateResponse(s, i.Interaction, fmt.Sprintf("Added `%s`", track.Name))
}

func userVoiceState(s *discordgo.State, guildID, userID string) *discordgo.VoiceState {
	g, err := s.Guild(guildID)
	if err != nil {
		return nil
	}

	for _, vs := range g.VoiceStates {
		if vs.UserID == userID {
			return vs
		}
	}

	return nil
}

type NextCommand struct {
	hub *Hub
}

func NewNextCommand(hub *Hub) *NextCommand {
	return &NextCommand{
		hub: hub,
	}
}

func (c *NextCommand) Command() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        "next",
		Description: "Skip to the next track",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "n",
				Description: "Number of tracks to skip",
				Required:    false,
			},
		},
	}
}

func (c *NextCommand) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) {
	bot, ok := c.hub.GetBot(i.GuildID)
	if !ok {
		respondAddMusicFirst(s, i.Interaction)
		return
	}

	n := 1
	if len(i.ApplicationCommandData().Options) > 0 {
		n = int(i.ApplicationCommandData().Options[0].IntValue())
	}

	err := bot.Next(n)
	if err != nil {
		respondTextPrivate(s, i.Interaction, "Cannot skip next")
		return
	}

	if bot.CurrentTrackIndex() >= bot.TotalTracks() {
		respondText(s, i.Interaction, "End of queue")
	} else {
		respondText(s, i.Interaction, fmt.Sprintf("Skipped to track #%d", bot.CurrentTrackIndex()+1))
	}
}

type PauseCommand struct {
	hub *Hub
}

func NewPauseCommand(hub *Hub) *PauseCommand {
	return &PauseCommand{
		hub: hub,
	}
}

func (c *PauseCommand) Command() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        "pause",
		Description: "Pause the current playing track",
	}
}

func (c *PauseCommand) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) {
	bot, ok := c.hub.GetBot(i.GuildID)
	if !ok {
		respondAddMusicFirst(s, i.Interaction)
		return
	}

	if bot.State() != BotStatePlaying {
		respondTextPrivate(s, i.Interaction, "No music playing")
		return
	}

	err := bot.Pause()
	if err != nil {
		respondTextPrivate(s, i.Interaction, "Cannot pause")
		return
	}

	respondText(s, i.Interaction, "Player paused")
}

type ResumeCommand struct {
	hub *Hub
}

func NewResumeCommand(hub *Hub) *ResumeCommand {
	return &ResumeCommand{
		hub: hub,
	}
}

func (c *ResumeCommand) Command() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        "resume",
		Description: "Resume the current playing track",
	}
}

func (c *ResumeCommand) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) {
	bot, ok := c.hub.GetBot(i.GuildID)
	if !ok {
		respondAddMusicFirst(s, i.Interaction)
		return
	}

	if bot.State() != BotStatePaused {
		respondTextPrivate(s, i.Interaction, "Music is now playing")
		return
	}

	err := bot.Resume()
	if err != nil {
		respondTextPrivate(s, i.Interaction, "Cannot resume")
		return
	}

	respondText(s, i.Interaction, "Resuming player")
}

type RemoveCommand struct {
	hub *Hub
}

func NewRemoveCommand(hub *Hub) *RemoveCommand {
	return &RemoveCommand{
		hub: hub,
	}
}

func (c *RemoveCommand) Command() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        "remove",
		Description: "Remove track from the playlist",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "track-no",
				Description: "Track No.",
				Required:    true,
			},
		},
	}
}

func (c *RemoveCommand) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) {
	bot, ok := c.hub.GetBot(i.GuildID)
	if !ok {
		respondTextPrivate(s, i.Interaction, "Pammy didn't join any channels")
		return
	}

	trackNo := int(i.ApplicationCommandData().Options[0].IntValue())

	err := bot.Remove(trackNo - 1)
	if err != nil {
		switch err {
		case ErrTrackNotFound:
			respondTextPrivate(s, i.Interaction, "Track not found")
		case ErrCannotRemovePlayingTrack:
			respondTextPrivate(s, i.Interaction, "Cannot remove the playing track")
		}
		return
	}

	respondText(s, i.Interaction, fmt.Sprintf("Removed track #%d", trackNo))
}

type ResetCommand struct {
	hub *Hub
}

func NewResetCommand(hub *Hub) *ResetCommand {
	return &ResetCommand{
		hub: hub,
	}
}

func (c *ResetCommand) Command() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        "reset",
		Description: "Reset the player",
	}
}

func (c *ResetCommand) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) {
	bot, ok := c.hub.GetBot(i.GuildID)
	if !ok {
		respondAddMusicFirst(s, i.Interaction)
		return
	}

	bot.Reset()

	respondText(s, i.Interaction, "Player is reset")
}

type ListCommand struct {
	hub *Hub
}

func NewListCommand(hub *Hub) *ListCommand {
	return &ListCommand{
		hub: hub,
	}
}

func (c *ListCommand) Command() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        "list",
		Description: "List tracks in playlist",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "page",
				Description: "Page",
				Required:    false,
			},
		},
	}
}

func (c *ListCommand) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) {
	bot, ok := c.hub.GetBot(i.GuildID)
	if !ok {
		respondTextPrivate(s, i.Interaction, "Pammy didn't join any channels")
		return
	}

	page := 0
	if len(i.ApplicationCommandData().Options) > 0 {
		page = int(i.ApplicationCommandData().Options[0].IntValue())
	}

	trackPage := bot.List(page, 10)

	respondText(s, i.Interaction, trackPage.DisplayText())
}

type LeaveCommand struct {
	hub *Hub
}

func NewLeaveCommand(hub *Hub) *LeaveCommand {
	return &LeaveCommand{
		hub: hub,
	}
}

func (c *LeaveCommand) Command() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        "leave",
		Description: "Leave the player",
	}
}

func (c *LeaveCommand) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) {
	bot, ok := c.hub.GetBot(i.GuildID)
	if !ok {
		respondTextPrivate(s, i.Interaction, "Pammy didn't join any channels")
		return
	}

	bot.Close()
	c.hub.RemoveBot(i.GuildID)

	respondText(s, i.Interaction, "Seeya!")
}

func respondAddMusicFirst(s *discordgo.Session, i *discordgo.Interaction) error {
	return respondTextPrivate(s, i, "Add music with `/play {search-term}` first")
}

func respondText(s *discordgo.Session, i *discordgo.Interaction, content string) error {
	return s.InteractionRespond(i, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
		},
	})
}

func respondTextPrivate(s *discordgo.Session, i *discordgo.Interaction, content string) error {
	return s.InteractionRespond(i, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
			Flags:   1 << 6,
		},
	})
}

func updateResponse(s *discordgo.Session, i *discordgo.Interaction, content string) error {
	_, err := s.InteractionResponseEdit(s.State.User.ID, i, &discordgo.WebhookEdit{
		Content: content,
	})

	return err
}
