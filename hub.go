package pammy

import (
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/noppawitt/pammy/youtube"
)

type Hub struct {
	mu sync.RWMutex

	bots map[string]*Bot
}

func NewHub(dg *discordgo.Session, ytClient *youtube.Client) *Hub {
	return &Hub{
		bots: make(map[string]*Bot),
	}
}

func (h *Hub) SetBot(bot *Bot, guildID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.bots[guildID] = bot
}

func (h *Hub) GetBot(guildID string) (*Bot, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	bot, ok := h.bots[guildID]
	return bot, ok
}

func (h *Hub) RemoveBot(guildID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.bots, guildID)
}
