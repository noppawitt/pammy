package pammy

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/noppawitt/pammy/youtube"
)

type Client struct {
	dg              *discordgo.Session
	commandHandlers map[string]CommandHandleFunc
	ytClient        *youtube.Client
	hub             *Hub
}

func NewClient(token string) (*Client, error) {
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, err
	}

	ytClient := youtube.NewClient()

	b := &Client{
		dg:              dg,
		commandHandlers: make(map[string]CommandHandleFunc),
		ytClient:        ytClient,
		hub:             NewHub(dg, ytClient),
	}

	return b, nil
}

func (c *Client) Start() error {
	c.dg.AddHandler(c.handleInteractiveCreate)

	c.dg.Identify.Intents = discordgo.IntentsGuilds | discordgo.IntentsGuildMessages | discordgo.IntentsGuildVoiceStates

	err := c.dg.Open()
	if err != nil {
		return err
	}

	defer c.dg.Close()

	c.AddGlobalSlashCommand(NewAddCommand(c.hub, c.ytClient))
	c.AddGlobalSlashCommand(NewNextCommand(c.hub))
	c.AddGlobalSlashCommand(NewPauseCommand(c.hub))
	c.AddGlobalSlashCommand(NewResumeCommand(c.hub))
	c.AddGlobalSlashCommand(NewListCommand(c.hub))
	c.AddGlobalSlashCommand(NewRemoveCommand(c.hub))
	c.AddGlobalSlashCommand(NewClearCommand(c.hub))
	c.AddGlobalSlashCommand(NewResetCommand(c.hub))
	c.AddGlobalSlashCommand(NewLeaveCommand(c.hub))
	c.AddGlobalSlashCommand(NewAutoPlayCommand(c.hub))

	log.Println("Pammy is now running.")

	termCh := make(chan os.Signal, 1)
	signal.Notify(termCh, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-termCh

	log.Println("Pammy is shutting down.")

	return nil
}

func (c *Client) AddGlobalSlashCommand(cmd SlashCommand) {
	_, err := c.dg.ApplicationCommandCreate(c.dg.State.User.ID, "", cmd.Command())
	if err != nil {
		panic(err)
	}
	c.commandHandlers[cmd.Command().Name] = cmd.Handle
}

func (c *Client) handleInteractiveCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	h, ok := c.commandHandlers[i.ApplicationCommandData().Name]
	if !ok {
		return
	}

	h(s, i)
}

type CommandHandleFunc func(s *discordgo.Session, i *discordgo.InteractionCreate)

type SlashCommand interface {
	Command() *discordgo.ApplicationCommand
	Handle(s *discordgo.Session, i *discordgo.InteractionCreate)
}
