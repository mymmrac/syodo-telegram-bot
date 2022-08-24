package main

import (
	"github.com/fasthttp/router"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
	"github.com/valyala/fasthttp"

	"github.com/mymmrac/syodo-telegram-bot/config"
	"github.com/mymmrac/syodo-telegram-bot/logger"
)

// Handler represents update handler
type Handler struct {
	cfg  *config.Config
	log  logger.Logger
	bot  *telego.Bot
	bh   *th.BotHandler
	rtr  *router.Router
	data TextData
}

// NewHandler creates new Handler
func NewHandler(cfg *config.Config, log logger.Logger, bot *telego.Bot, bh *th.BotHandler, rtr *router.Router,
	textData TextData) *Handler {
	return &Handler{
		cfg:  cfg,
		log:  log,
		bot:  bot,
		bh:   bh,
		rtr:  rtr,
		data: textData,
	}
}

// RegisterHandlers registers all handlers in bot handler
func (h *Handler) RegisterHandlers() {
	err := h.bot.SetMyCommands(&telego.SetMyCommandsParams{
		Commands: []telego.BotCommand{
			{Command: "start", Description: h.data.Text("startDescription")},
			{Command: "help", Description: h.data.Text("helpDescription")},
		},
	})
	if err != nil {
		h.log.Fatalf("Set bot commands: %v", err)
	}

	err = h.bot.SetChatMenuButton(&telego.SetChatMenuButtonParams{
		MenuButton: &telego.MenuButtonWebApp{
			Type: telego.ButtonTypeWebApp,
			Text: h.data.Text("menuButton"),
			WebApp: telego.WebAppInfo{
				URL: h.data.Text("webAppURL"),
			},
		},
	})
	if err != nil {
		h.log.Fatalf("Set bot menu button: %v", err)
	}

	h.bh.HandleMessage(h.startCmd, th.CommandEqual("start"))
	h.bh.HandleMessage(h.helpCmd, th.CommandEqual("help"))

	h.rtr.POST("/order", func(ctx *fasthttp.RequestCtx) {
		h.log.Infof("Received order request: `%s`", string(ctx.PostBody()))
		h.orderHandler(ctx)
	})
}

func (h *Handler) startCmd(bot *telego.Bot, message telego.Message) {
	chatID := message.Chat.ID
	_, err := bot.SendMessage(
		tu.Message(tu.ID(chatID), h.data.Temp("start", message)).
			WithParseMode(telego.ModeHTML).
			WithReplyMarkup(tu.InlineKeyboard(
				tu.InlineKeyboardRow(
					tu.InlineKeyboardButton(h.data.Text("menuButton")).
						WithWebApp(&telego.WebAppInfo{URL: h.data.Text("webAppURL")}),
				),
			)),
	)
	if err != nil {
		h.log.Errorf("Send start message: %s", err)
	}
}

func (h *Handler) helpCmd(bot *telego.Bot, message telego.Message) {
	chatID := message.Chat.ID
	_, err := bot.SendMessage(
		tu.Message(tu.ID(chatID), h.data.Temp("help", message)).
			WithParseMode(telego.ModeHTML).
			WithReplyMarkup(tu.InlineKeyboard(
				tu.InlineKeyboardRow(
					tu.InlineKeyboardButton(h.data.Text("siteButtonText")).
						WithURL(h.data.Text("siteURL")),
				),
				tu.InlineKeyboardRow(
					tu.InlineKeyboardButton(h.data.Text("instagramButtonText")).
						WithURL(h.data.Text("instagramURL")),
					tu.InlineKeyboardButton(h.data.Text("facebookButtonText")).
						WithURL(h.data.Text("facebookURL")),
				),
			)),
	)
	if err != nil {
		h.log.Errorf("Send help message: %s", err)
	}
}

func (h *Handler) orderHandler(ctx *fasthttp.RequestCtx) {
	_, _ = ctx.WriteString("ok")
}
