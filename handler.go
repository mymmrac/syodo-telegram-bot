package main

import (
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"

	"github.com/mymmrac/syodo-telegram-bot/config"
	"github.com/mymmrac/syodo-telegram-bot/logger"
)

type Handler struct {
	cfg *config.Config
	log logger.Logger
	bh  *th.BotHandler
}

func NewHandler(cfg *config.Config, log logger.Logger, bh *th.BotHandler) *Handler {
	return &Handler{
		cfg: cfg,
		log: log,
		bh:  bh,
	}
}

func (h *Handler) RegisterHandlers() {
	h.bh.HandleMessage(h.createInvoice, th.CommandEqual("invoice"))
	h.bh.HandlePreCheckoutQuery(h.preCheckout)
	h.bh.HandleMessage(h.successPayment, th.SuccessPayment())
}

const currency = "UAH"

func (h *Handler) createInvoice(bot *telego.Bot, message telego.Message) {
	chatID := message.Chat.ID
	_, err := bot.SendInvoice(tu.Invoice(
		tu.ID(chatID),
		"Test Invoice",
		"Hello World!",
		"test-invoice",
		h.cfg.Settings.ProviderToken,
		currency,
		tu.LabeledPrice("🍱 Bento", 2000),
		tu.LabeledPrice("🍣 Sushi", 1000),
		tu.LabeledPrice("🚚 Delivery", 0),
		tu.LabeledPrice("🎟 Coupon", -1500),
	).
		WithNeedPhoneNumber().
		WithNeedShippingAddress().
		WithMaxTipAmount(5000).
		WithSuggestedTipAmounts(1000, 3000, 5000))
	if err != nil {
		h.log.Error("Send invoice: ", err)
	}
}

func (h *Handler) preCheckout(bot *telego.Bot, query telego.PreCheckoutQuery) {
	err := bot.AnswerPreCheckoutQuery(tu.PreCheckoutQuery(query.ID, true))
	if err != nil {
		h.log.Error("Answer pre checkout: ", err)
	}
}

func (h *Handler) successPayment(bot *telego.Bot, message telego.Message) {
	chatID := message.Chat.ID
	payment := message.SuccessfulPayment

	_, err := bot.SendMessage(tu.MessageWithEntities(tu.ID(chatID),
		tu.Entity("Thanks for ordering, shipping to: "),
		tu.Entity(payment.OrderInfo.ShippingAddress.StreetLine1).Bold(),
	))
	if err != nil {
		h.log.Error("Send message: ", err)
	}
}
