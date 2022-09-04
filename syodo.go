package main

import (
	"encoding/json"
	"fmt"
	"net/url"

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

func (s *SyodoService) call(path string, method string, data, result any) error {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	apiURL, err := url.JoinPath(s.cfg.App.SyodoAPIURL, path)
	if err != nil {
		return fmt.Errorf("join path: %w", err)
	}

	req.SetRequestURI(apiURL)

	req.Header.SetMethod(method)
	req.Header.SetContentType(contentTypeJSON)
	req.Header.Set(authHeader, s.cfg.App.SyodoAPIKey)

	if data != nil {
		jsonData, err := json.Marshal(data)
		if err != nil {
			return fmt.Errorf("encode data: %w", err)
		}

		req.SetBodyRaw(jsonData)
	}

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	if err := s.client.Do(req, resp); err != nil {
		return fmt.Errorf("call syodo: %w", err)
	}

	if statusCode := resp.StatusCode(); statusCode != fasthttp.StatusOK {
		return fmt.Errorf("call syodo bad status: %d", statusCode)
	}

	if err := json.Unmarshal(resp.Body(), result); err != nil {
		return fmt.Errorf("decode result: %w", err)
	}

	return nil
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

	var priceResp priceResponse
	if err := s.call("/price", fasthttp.MethodPost, priceReq, &priceResp); err != nil {
		return 0, fmt.Errorf("price API: %w", err)
	}

	return priceResp.Delivery - priceResp.Discount, nil
}
