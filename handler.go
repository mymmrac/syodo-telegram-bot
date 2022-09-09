package main

import (
	"encoding/json"
	"fmt"
	"sync"
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
	syodo      *SyodoService
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
		syodo:      NewSyodoService(cfg),
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
		h.log.Errorf("Unmarshal order request: %s", err)
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}

	_, err := tu.ValidateWebAppData(h.bot.Token(), order.AppData)
	if err != nil {
		h.log.Errorf("Invalid web app data: %q", order.AppData)
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
		h.log.Errorf("Create invoice link: %q, %s", link, err)
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
	case "7", "8", "14": // Роли, Сети, Без лактози
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
	order, ok := h.getOrder(query.InvoicePayload)
	if !ok {
		h.log.Errorf("Order not found: %s", query.InvoicePayload)
		h.failShipping(query.ID, h.data.Text("orderNotFoundError"))
		return
	}

	var (
		wg      sync.WaitGroup
		options []telego.ShippingOption

		zone             DeliveryZone
		label            string
		priceDelivery    int
		priceDeliveryErr error

		priceSelfPickup    int
		priceSelfPickupErr error
	)

	wg.Add(1)
	go func() {
		defer wg.Done()

		zone = h.delivery.CalculateZone(query.ShippingAddress)

		priceDelivery, priceDeliveryErr = h.syodo.CalculatePrice(order, zone, false)
		if priceDeliveryErr != nil {
			return
		}

		label = labelByZone(zone)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		priceSelfPickup, priceSelfPickupErr = h.syodo.CalculatePrice(order, "", true)
	}()

	wg.Wait()
	if priceDeliveryErr != nil || priceSelfPickupErr != nil {
		h.log.Errorf("Calculate price: %s, %s", priceDeliveryErr, priceSelfPickupErr)
		h.failShipping(query.ID, h.data.Text("calculateShippingPriceError"))
		return
	}

	if zone != ZoneUnknown {
		options = append(options, tu.ShippingOption(zone, "Доставка курєром",
			tu.LabeledPrice(label, priceDelivery),
		))
	}

	options = append(options, tu.ShippingOption(SelfPickup, "Самовивіз",
		tu.LabeledPrice("👋 Самовивіз (-10%)", priceSelfPickup),
	))

	err := bot.AnswerShippingQuery(tu.ShippingQuery(query.ID, true, options...))
	if err != nil {
		h.log.Errorf("Answer shipping: %s", err)
		return
	}
}

func labelByZone(zone DeliveryZone) string {
	switch zone {
	case ZoneGreen:
		return "🛵 Доставка у зелену зону"
	case ZoneYellow:
		return "🛵 Доставка у жовту зону"
	case ZoneRed:
		return "🛵 Доставка у червону зону"
	default:
		// No shipping option
		return "<UNKNOWN>"
	}
}

func (h *Handler) failShipping(queryID, failureReason string) {
	err := h.bot.AnswerShippingQuery(tu.ShippingQuery(queryID, false).WithErrorMessage(failureReason))
	if err != nil {
		h.log.Errorf("Answer shipping (failure): %s", err)
		return
	}
}

func (h *Handler) preCheckout(bot *telego.Bot, query telego.PreCheckoutQuery) {
	_, ok := DeliveryMethodIDs[query.ShippingOptionID]
	if !ok {
		h.log.Errorf("Unknown delivery method: %s", query.ShippingOptionID)
		h.failPreCheckout(query.ID, h.data.Text("orderDeliveryError"))
		return
	}

	info := query.OrderInfo
	if info == nil || info.ShippingAddress == nil || info.Name == "" || info.PhoneNumber == "" {
		h.log.Errorf("Bad order info: %+v", info)
		h.failPreCheckout(query.ID, h.data.Text("orderInfoError"))
		return
	}

	order, ok := h.getOrder(query.InvoicePayload)
	if !ok {
		h.log.Errorf("Order not found: %s", query.InvoicePayload)
		h.failPreCheckout(query.ID, h.data.Text("orderNotFoundError"))
		return
	}

	order.OrderInfo = info
	order.ShippingOptionID = query.ShippingOptionID

	if err := h.syodo.Checkout(&order); err != nil {
		h.log.Errorf("Checkout: %s", err)
		h.failPreCheckout(query.ID, h.data.Text("orderCheckoutError"))
		return
	}
	h.log.Debugf("Order checkout: %+v", order)

	h.updateOrder(order)

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

	order, ok := h.getOrder(payment.InvoicePayload)
	if !ok {
		h.log.Errorf("Order not found: %s", payment.InvoicePayload)

		_, err := bot.SendMessage(tu.Message(tu.ID(chatID), h.data.Text("successPaymentOrderNotFoundError")))
		if err != nil {
			h.log.Errorf("Send success payment error message: %s", err)
			return
		}
		return
	}

	_, err := bot.SendMessage(tu.Message(tu.ID(chatID), h.data.Temp("successPayment", order)).
		WithParseMode(telego.ModeHTML))
	if err != nil {
		h.log.Errorf("Send success payment message: %s", err)
		return
	}
}
