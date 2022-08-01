package main

import (
	"fmt"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"

	"github.com/mymmrac/syodo-telegram-bot/config"
	"github.com/mymmrac/syodo-telegram-bot/logger"
)

// Handler represents update handler
type Handler struct {
	cfg *config.Config
	log logger.Logger
	bh  *th.BotHandler
}

// NewHandler creates new Handler
func NewHandler(cfg *config.Config, log logger.Logger, bh *th.BotHandler) *Handler {
	return &Handler{
		cfg: cfg,
		log: log,
		bh:  bh,
	}
}

// RegisterHandlers registers all handlers in bot handler
func (h *Handler) RegisterHandlers() {
	h.bh.HandleMessage(h.startCmd, th.CommandEqual("start"))
}

func (h *Handler) startCmd(bot *telego.Bot, message telego.Message) {
	chatID := message.Chat.ID
	_, err := bot.SendMessage(tu.Message(tu.ID(chatID), fmt.Sprintf("Hello %s", message.From.FirstName)))
	if err != nil {
		h.log.Errorf("Send message: %s", err)
	}
}
