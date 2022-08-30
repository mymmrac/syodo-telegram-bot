package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/fasthttp/router"
	"github.com/mymmrac/memkey"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
	"github.com/valyala/fasthttp"

	"github.com/mymmrac/syodo-telegram-bot/config"
	"github.com/mymmrac/syodo-telegram-bot/logger"
)

const (
	currency      = "UAH"
	orderTTL      = time.Hour * 4
	orderKeyBound = 1_000_000
)

// Handler represents update handler
type Handler struct {
	cfg        *config.Config
	log        logger.Logger
	bot        *telego.Bot
	bh         *th.BotHandler
	rtr        *router.Router
	data       TextData
	orderStore *memkey.Store[string]
	delivery   *DeliveryStrategy
}

// NewHandler creates new Handler
func NewHandler(cfg *config.Config, log logger.Logger, bot *telego.Bot, bh *th.BotHandler, rtr *router.Router,
	textData TextData, delivery *DeliveryStrategy,
) *Handler {
	return &Handler{
		cfg:        cfg,
		log:        log,
		bot:        bot,
		bh:         bh,
		rtr:        rtr,
		data:       textData,
		orderStore: &memkey.Store[string]{},
		delivery:   delivery,
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
				URL: h.cfg.App.WebAppURL,
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
	h.bh.HandleMessage(h.successPayment, th.SuccessPayment())

	h.rtr.POST("/order", func(ctx *fasthttp.RequestCtx) {
		h.log.Infof("Received order request: `%s`", string(ctx.PostBody()))
		h.orderHandler(ctx)
	})
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

	h.invalidateOldOrders()
	orderKey := h.storeOrder(order)

	prices := make([]telego.LabeledPrice, 0, len(order.Products))
	for _, p := range order.Products {
		prices = append(prices, telego.LabeledPrice{
			Label:  fmt.Sprintf("%s %d 𐄂 %s", emojiByCategoryID(p.CategoryID), p.Amount, p.Title),
			Amount: p.Amount * p.Price,
		})
	}

	if order.CutleryCount > 0 {
		prices = append(prices, tu.LabeledPrice(fmt.Sprintf("🥢 %d 𐄂 Прибори", order.CutleryCount), 0))
	}
	if order.TrainingCutleryCount > 0 {
		prices = append(prices, tu.LabeledPrice(fmt.Sprintf("🥢 %d 𐄂 Навчальні прибори", order.TrainingCutleryCount), 0))
	}
	if !order.NoNapkins {
		prices = append(prices, tu.LabeledPrice("🧻 Серветки", 0))
	}

	link, err := h.bot.CreateInvoiceLink(&telego.CreateInvoiceLinkParams{
		Title:                     "Замовлення #" + orderKey,
		Description:               h.data.Text("orderDescription"),
		Payload:                   orderKey,
		ProviderToken:             h.cfg.App.ProviderToken,
		Currency:                  currency,
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

	//nolint:errcheck
	_, _ = ctx.WriteString(*link)
	ctx.SetStatusCode(fasthttp.StatusOK)
}

func emojiByCategoryID(id string) string {
	switch id {
	case "13": // Суші
		return "🍣"
	case "7", "8": // Роли, Сети
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
	_, ok := h.getOrder(query.InvoicePayload)

	if !ok {
		err := bot.AnswerShippingQuery(tu.ShippingQuery(query.ID, false).
			WithErrorMessage(h.data.Text("orderNotFoundError")))
		if err != nil {
			h.log.Errorf("Answer shipping: %s", err)
			return
		}

		return
	}

	var options []telego.ShippingOption
	zone := h.delivery.CalculateZone(query.ShippingAddress)

	switch zone {
	case ZoneGreen:
		options = append(options, tu.ShippingOption(string(ZoneGreen), "Доставка курєром",
			tu.LabeledPrice("🛵 Доставка у зелену зону", h.cfg.App.Prices.RegularDelivery),
		))
	case ZoneYellow:
		options = append(options, tu.ShippingOption(string(ZoneYellow), "Доставка курєром",
			tu.LabeledPrice("🛵 Доставка у жовту зону", h.cfg.App.Prices.RegularDelivery),
		))
	case ZoneRed:
		options = append(options, tu.ShippingOption(string(ZoneRed), "Доставка курєром",
			tu.LabeledPrice("🛵 Доставка у червону зону", h.cfg.App.Prices.RegularDelivery),
		))
	default:
		// No shipping option
	}

	options = append(options, tu.ShippingOption(SelfPickup, "Самовивіз",
		tu.LabeledPrice("👋 Самовивіз", h.cfg.App.Prices.SelfPickup),
	))

	err := bot.AnswerShippingQuery(tu.ShippingQuery(query.ID, true, options...))
	if err != nil {
		h.log.Errorf("Answer shipping: %s", err)
		return
	}
}

func (h *Handler) preCheckout(bot *telego.Bot, query telego.PreCheckoutQuery) {
	_, ok := DeliveryMethodIDs[query.ShippingOptionID]
	if !ok {
		h.failPreCheckout(query.ID, h.data.Text("orderDeliveryError"))
		return
	}

	_, ok = h.getOrder(query.InvoicePayload)
	if !ok {
		h.failPreCheckout(query.ID, h.data.Text("orderNotFoundError"))
		return
	}

	err := bot.AnswerPreCheckoutQuery(tu.PreCheckoutQuery(query.ID, true))
	if err != nil {
		h.log.Errorf("Answer pre checkout: %s", err)
		return
	}
}

func (h *Handler) failPreCheckout(queryID, failureReason string) {
	err := h.bot.AnswerPreCheckoutQuery(tu.PreCheckoutQuery(queryID, false).WithErrorMessage(failureReason))
	if err != nil {
		h.log.Errorf("Answer pre checkout (failure): %s", err)
		return
	}
}

func (h *Handler) successPayment(bot *telego.Bot, message telego.Message) {
	chatID := message.Chat.ID
	payment := message.SuccessfulPayment

	_, err := bot.SendMessage(tu.Message(tu.ID(chatID), h.data.Temp("successPayment", payment)))
	if err != nil {
		h.log.Errorf("Send success payment message: %s", err)
		return
	}
}
