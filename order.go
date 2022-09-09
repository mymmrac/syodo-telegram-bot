package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/mymmrac/memkey"
	"github.com/mymmrac/telego"
)

// OrderProduct represents a single item in order
type OrderProduct struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	Price      int    `json:"price"`
	Amount     int    `json:"amount"`
	CategoryID string `json:"categoryID"`
}

// OrderRequest represents order info sent by user
type OrderRequest struct {
	AppData              string         `json:"appData"`
	Products             []OrderProduct `json:"products"`
	DoNotCall            bool           `json:"doNotCall"`
	NoNapkins            bool           `json:"noNapkins"`
	CutleryCount         int            `json:"cutleryCount"`
	TrainingCutleryCount int            `json:"trainingCutleryCount"`
	Comment              string         `json:"comment"`
}

// OrderDetails represents full order info
type OrderDetails struct {
	OrderID          string            `json:"orderID"`
	ExternalOrderID  string            `json:"externalOrderID"`
	Request          OrderRequest      `json:"request"`
	OrderInfo        *telego.OrderInfo `json:"orderInfo"`
	ShippingOptionID string            `json:"shippingOptionID"`
	OrderURL         string            `json:"orderURL"`
	TotalAmount      int               `json:"totalAmount"`
	CreatedAt        time.Time         `json:"createdAt"`
}

func (h *Handler) storeOrder(order OrderRequest) string {
	var orderKey string
	for orderKey == "" || h.orderStore.Has(orderKey) {
		//nolint:gosec
		orderKey = fmt.Sprintf("%06d", rand.Intn(orderKeyBound))
	}

	memkey.Set(h.orderStore, orderKey, OrderDetails{
		OrderID:   orderKey,
		Request:   order,
		CreatedAt: time.Now().UTC(),
	})

	return orderKey
}

func (h *Handler) getOrder(key string) (OrderDetails, bool) {
	return memkey.Get[OrderDetails](h.orderStore, key)
}

func (h *Handler) updateOrder(order OrderDetails) {
	memkey.Set(h.orderStore, order.OrderID, order)
}

func (h *Handler) invalidateOldOrders() {
	ttlTime := time.Now().UTC().Add(-orderTTL)

	for _, e := range memkey.Entries[OrderDetails](h.orderStore) {
		if ttlTime.After(e.Value.CreatedAt) {
			h.orderStore.Delete(e.Key)
		}
	}
}
