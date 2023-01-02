package main

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/fasthttp/router"
	"github.com/mymmrac/memkey"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
	"github.com/valyala/fasthttp"
	"googlemaps.github.io/maps"

	"github.com/mymmrac/syodo-telegram-bot/config"
	"github.com/mymmrac/syodo-telegram-bot/logger"
)

const (
	currency      = "UAH"
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
		syodo:      NewSyodoService(cfg, log),
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
	h.bh.HandlePreCheckoutQuery(h.preCheckout)
	h.bh.HandleMessage(h.successPayment, th.SuccessPayment())
	h.bh.HandleMessage(h.unknown)

	h.rtr.POST("/order", func(ctx *fasthttp.RequestCtx) {
		h.log.Debugf("Received order request: `%s`", string(ctx.PostBody()))
		h.orderHandler(ctx)
	})

	h.rtr.GET("/order", func(ctx *fasthttp.RequestCtx) {
		//nolint:errcheck
		_, _ = ctx.WriteString(strconv.Itoa(h.orderStore.Len()))
	})
}

//nolint:funlen,gocognit,cyclop
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

	if order.Name == "" || len(order.Phone) != 13 ||
		(order.DeliveryType == deliveryTypeDelivery && (order.Address == "" || order.City == "")) {
		h.log.Errorf("Bad order info: %+v", order)
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}

	var (
		price    PriceResponse
		location maps.LatLng
	)

	switch order.DeliveryType {
	case deliveryTypeDelivery:
		location, err = h.delivery.CalculateLocation(order)
		if err != nil {
			break
		}

		price, err = h.syodo.CalculatePriceDelivery(order.Products, location, order.Promotion)
	case "self_pickup_1", "self_pickup_2":
		price, err = h.syodo.CalculatePriceSelfPickup(order.Products, order.Promotion)
	default:
		h.log.Errorf("Unknown delivery type: %s", err)
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}

	if err != nil {
		h.log.Errorf("Calculate price: %s", err)
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		return
	}

	h.invalidateOldOrders()
	orderKey := h.storeOrder(order, price.ServiceArea)

	link, err := h.bot.CreateInvoiceLink(&telego.CreateInvoiceLinkParams{
		Title:         "Замовлення #" + orderKey,
		Description:   h.data.Text("orderDescription"),
		Payload:       orderKey,
		ProviderToken: h.cfg.App.ProviderToken,
		Currency:      currency,
		Prices:        h.constructPrices(order, price),
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

func (h *Handler) constructPrices(order OrderRequest, price PriceResponse) []telego.LabeledPrice {
	prices := make([]telego.LabeledPrice, 0, len(order.Products))
	for _, p := range order.Products {
		prices = append(prices, telego.LabeledPrice{
			Label:  fmt.Sprintf("%s %d ✕ %s", emojiByCategoryID(p.CategoryID), p.Amount, p.Title),
			Amount: p.Amount * p.Price,
		})
	}

	if order.CutleryCount > 0 {
		prices = append(prices, tu.LabeledPrice(fmt.Sprintf("🥢 %d ✕ Прибори", order.CutleryCount), 0))
	}
	if order.TrainingCutleryCount > 0 {
		prices = append(prices, tu.LabeledPrice(fmt.Sprintf("🥢 %d ✕ Навчальні прибори", order.TrainingCutleryCount), 0))
	}
	if !order.NoNapkins {
		prices = append(prices, tu.LabeledPrice("🧻 Серветки", 0))
	}

	if price.Delivery != 0 {
		if order.DeliveryType == deliveryTypeDelivery {
			prices = append(prices, tu.LabeledPrice(h.labelByZone(price.ServiceArea), price.Delivery))
		} else {
			prices = append(prices, tu.LabeledPrice("👋 Самовивіз", price.Delivery))
		}
	}

	switch order.Promotion {
	case promo4Plus1:
		prices = append(prices, tu.LabeledPrice("🎟 Акція 4+1", -price.Discount))
	case promoSelfPickup:
		prices = append(prices, tu.LabeledPrice("🎟 Самовивіз -10%", -price.Discount))
	}

	return prices
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

func (h *Handler) labelByZone(zone DeliveryZone) string {
	switch zone {
	case ZoneGreen:
		return "🛵 Доставка у зелену зону"
	case ZoneYellow:
		return "🛵 Доставка у жовту зону"
	case ZoneRed:
		return "🛵 Доставка у червону зону"
	default:
		// No shipping option
		h.log.Errorf("Unknown zone: %q", zone)
		return "<UNKNOWN>"
	}
}

func (h *Handler) preCheckout(bot *telego.Bot, query telego.PreCheckoutQuery) {
	order, ok := h.getOrder(query.InvoicePayload)
	if !ok {
		h.log.Errorf("Order not found: %s", query.InvoicePayload)
		h.failPreCheckout(query.ID, h.data.Text("orderNotFoundError"))
		return
	}

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

	if err := h.syodo.SuccessPayment(payment, order.ExternalOrderID); err != nil {
		h.log.Errorf("Success payment: %s", err)

		_, err = bot.SendMessage(tu.Message(tu.ID(chatID), h.data.Text("successPaymentOrderFailedError")))
		if err != nil {
			h.log.Errorf("Send success payment error message: %s", err)
			return
		}
		return
	}

	h.orderStore.Delete(payment.InvoicePayload)

	_, err := bot.SendMessage(tu.Message(tu.ID(chatID), h.data.Temp("successPayment", order)).
		WithParseMode(telego.ModeHTML))
	if err != nil {
		h.log.Errorf("Send success payment message: %s", err)
		return
	}
}
