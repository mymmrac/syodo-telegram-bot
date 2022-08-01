package main

import (
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"

	"github.com/mymmrac/syodo-telegram-bot/config"
	"github.com/mymmrac/syodo-telegram-bot/logger"
)

// Handler represents update handler
type Handler struct {
	cfg  *config.Config
	log  logger.Logger
	bh   *th.BotHandler
	text Text
}

// NewHandler creates new Handler
func NewHandler(cfg *config.Config, log logger.Logger, bh *th.BotHandler, text Text) *Handler {
	return &Handler{
		cfg:  cfg,
		log:  log,
		bh:   bh,
		text: text,
	}
}

// RegisterHandlers registers all handlers in bot handler
func (h *Handler) RegisterHandlers() {
	h.bh.HandleMessage(h.startCmd, th.CommandEqual("start"))
}

func (h *Handler) startCmd(bot *telego.Bot, message telego.Message) {
	chatID := message.Chat.ID
	_, err := bot.SendMessage(
		tu.Message(tu.ID(chatID), h.text.Get("start", message)).
			WithParseMode(telego.ModeHTML),
	)
	if err != nil {
		h.log.Errorf("Send start message: %s", err)
	}
}
