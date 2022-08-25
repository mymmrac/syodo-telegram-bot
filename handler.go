package main

import (
	"encoding/json"
	"fmt"

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
				URL: h.cfg.Settings.WebAppURL,
			},
		},
	})
	if err != nil {
		h.log.Fatalf("Set bot menu button: %v", err)
	}

	h.bh.HandleMessage(h.startCmd, th.CommandEqual("start"))
	h.bh.HandleMessage(h.helpCmd, th.CommandEqual("help"))
	h.bh.HandleShippingQuery(h.shipping)
	h.bh.HandlePreCheckoutQuery(h.preCheckout)

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

type OrderProduct struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	Price      int    `json:"price"`
	Amount     int    `json:"amount"`
	CategoryID string `json:"categoryID"`
}

type OrderRequest struct {
	AppData              string         `json:"appData"`
	Products             []OrderProduct `json:"products"`
	DoNotCall            bool           `json:"doNotCall"`
	NoNapkins            bool           `json:"noNapkins"`
	CutleryCount         int            `json:"cutleryCount"`
	TrainingCutleryCount int            `json:"trainingCutleryCount"`
	Comment              string         `json:"comment"`
}

func (h *Handler) orderHandler(ctx *fasthttp.RequestCtx) {
	data := ctx.PostBody()

	var order OrderRequest
	if err := json.Unmarshal(data, &order); err != nil {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}

	_, err := tu.ValidateWebAppData(h.bot.Token(), order.AppData)
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusForbidden)
		return
	}

	prices := make([]telego.LabeledPrice, len(order.Products))
	for i, p := range order.Products {
		prices[i] = telego.LabeledPrice{
			Label:  fmt.Sprintf("%s %d 𐄂 %s", emojiByCategoryID(p.CategoryID), p.Amount, p.Title),
			Amount: p.Amount * p.Price,
		}
	}

	link, err := h.bot.CreateInvoiceLink(&telego.CreateInvoiceLinkParams{
		Title:                     "SYODO",
		Description:               "Замовлення",
		Payload:                   "TODO", // TODO: Fill
		ProviderToken:             h.cfg.Settings.ProviderToken,
		Currency:                  "UAH",
		Prices:                    prices,
		NeedName:                  true,
		NeedPhoneNumber:           true,
		NeedShippingAddress:       true,
		SendPhoneNumberToProvider: true,
		IsFlexible:                true,
	})
	if err != nil || link == nil || *link == "" {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		return
	}

	_, _ = ctx.WriteString(*link)
	ctx.SetStatusCode(fasthttp.StatusOK)
}

func emojiByCategoryID(id string) string {
	switch id {
	case "13": // Суші
		return "🍣"
	case "7": // Роли
		return "🍱"
	case "8": // Сети
		return "🍱"
	case "9": // Напої
		return "🥤"
	case "10": // Соуси
		return "🍥"
	case "11": // Десерти
		return "🍡"
	default:
		return "🍱"
	}
}

func (h *Handler) shipping(bot *telego.Bot, query telego.ShippingQuery) {
	err := bot.AnswerShippingQuery(tu.ShippingQuery(query.ID, true,
		tu.ShippingOption("shipping_regular", "Доставка курєром",
			tu.LabeledPrice("🛵 Доставка звичайна", 6500),
		),
		tu.ShippingOption("take_away", "Самовивіз",
			tu.LabeledPrice("👋 Самовивіз", -1000),
		),
	))
	if err != nil {
		h.log.Errorf("Answer shipping: %s", err)
		return
	}
}

func (h *Handler) preCheckout(bot *telego.Bot, query telego.PreCheckoutQuery) {
	err := bot.AnswerPreCheckoutQuery(tu.PreCheckoutQuery(query.ID, true))
	if err != nil {
		h.log.Errorf("Answer pre checkout: %s", err)
		return
	}
}
