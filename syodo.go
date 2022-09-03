package main

import (
	"encoding/json"
	"fmt"

	"github.com/valyala/fasthttp"

	"github.com/mymmrac/syodo-telegram-bot/config"
)

const (
	contentTypeJSON = "application/json"
	authHeader      = "x-api-key"
)

const (
	shippingTypeDelivery   = "Доставка"
	shippingTypeSelfPickup = "Самовивіз"
)

// SyodoService represents a type to interact with Syodo API
type SyodoService struct {
	cfg    *config.Config
	client *fasthttp.Client
}

// NewSyodoService creates new SyodoService
func NewSyodoService(cfg *config.Config) *SyodoService {
	return &SyodoService{
		cfg:    cfg,
		client: &fasthttp.Client{},
	}
}

type priceRequestOrder struct {
	ID         string `json:"id"`
	CategoryID string `json:"category_id"`
	Title      string `json:"title"`
	Amount     int    `json:"qty"`
}

type priceRequestDelivery struct {
	Type string `json:"type"`
	Zone string `json:"serviceArea"`
}

type priceRequest struct {
	Order           []priceRequestOrder  `json:"order"`
	DeliveryDetails priceRequestDelivery `json:"deliveryDetails"`
}

type priceResponse struct {
	Delivery int `json:"delivery"`
	Discount int `json:"discount"`
}

// CalculatePrice returns calculated price depending on order details and delivery zone
func (s *SyodoService) CalculatePrice(order OrderDetails, zone DeliveryZone, selfPickup bool) (int, error) {
	requestOrder := make([]priceRequestOrder, len(order.Request.Products))
	for i, p := range order.Request.Products {
		requestOrder[i] = priceRequestOrder{
			ID:         p.ID,
			CategoryID: p.CategoryID,
			Title:      p.Title,
			Amount:     p.Amount,
		}
	}

	shippingType := shippingTypeDelivery
	if selfPickup {
		shippingType = shippingTypeSelfPickup
	}

	priceReq := &priceRequest{
		Order: requestOrder,
		DeliveryDetails: priceRequestDelivery{
			Type: shippingType,
			Zone: zone,
		},
	}

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.SetRequestURI(s.cfg.App.SyodoAPIURL + "/price")
	req.Header.SetContentType(contentTypeJSON)
	req.Header.SetMethod(fasthttp.MethodPost)
	req.Header.Set(authHeader, s.cfg.App.SyodoAPIKey)

	data, err := json.Marshal(priceReq)
	if err != nil {
		return 0, fmt.Errorf("encode body: %w", err)
	}
	req.SetBodyRaw(data)

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	err = s.client.Do(req, resp)
	if err != nil {
		return 0, fmt.Errorf("call syodo: %w", err)
	}

	if statusCode := resp.StatusCode(); statusCode != fasthttp.StatusOK {
		return 0, fmt.Errorf("call syodo status: %d", statusCode)
	}

	var priceResp priceResponse
	if err = json.Unmarshal(resp.Body(), &priceResp); err != nil {
		return 0, fmt.Errorf("decode body: %w", err)
	}

	return priceResp.Delivery - priceResp.Discount, nil
}
