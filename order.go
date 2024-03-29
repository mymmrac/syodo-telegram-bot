package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/mymmrac/memkey"
	"googlemaps.github.io/maps"
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
	Name                 string         `json:"name"`
	Phone                string         `json:"phone"`
	DeliveryType         string         `json:"deliveryType"`
	Location             maps.LatLng    `json:"-"`
	Promotion            string         `json:"promotion"`
	City                 string         `json:"city"`
	Address              string         `json:"address"`
	Entrance             string         `json:"entrance"`
	ECode                string         `json:"eCode"`
	Floor                string         `json:"floor"`
	Apartment            string         `json:"apartment"`
}

// OrderDetails represents full order info
type OrderDetails struct {
	OrderID         string       `json:"orderID"`
	ExternalOrderID string       `json:"externalOrderID"`
	Request         OrderRequest `json:"request"`
	OrderURL        string       `json:"orderURL"`
	ServiceArea     string       `json:"serviceArea"`
	TotalAmount     float64      `json:"totalAmount"`
	CreatedAt       time.Time    `json:"createdAt"`
}

func (h *Handler) storeOrder(order OrderRequest, area string) string {
	var orderKey string
	for orderKey == "" || h.orderStore.Has(orderKey) {
		//nolint:gosec
		orderKey = fmt.Sprintf("%06d", rand.Intn(orderKeyBound))
	}

	memkey.Set(h.orderStore, orderKey, OrderDetails{
		OrderID:     orderKey,
		Request:     order,
		ServiceArea: area,
		CreatedAt:   time.Now().UTC(),
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
	ttlTime := time.Now().UTC().Add(-h.cfg.Settings.OrderTTL)

	for _, e := range memkey.Entries[OrderDetails](h.orderStore) {
		if ttlTime.After(e.Value.CreatedAt) {
			h.orderStore.Delete(e.Key)
		}
	}
}
