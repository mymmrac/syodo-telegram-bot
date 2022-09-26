package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"time"

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

	shippingPromo4Plus1     = "4+1"
	shippingPromoSelfPickup = "Самовивіз"
)

// SyodoService represents a type to interact with Syodo API
type SyodoService struct {
	cfg      *config.Config
	client   *fasthttp.Client
	timezone *time.Location
}

// NewSyodoService creates new SyodoService
func NewSyodoService(cfg *config.Config) *SyodoService {
	loc, err := time.LoadLocation("Europe/Kiev")
	assert(err == nil, fmt.Errorf("load timezone: %w", err))

	return &SyodoService{
		cfg:      cfg,
		client:   &fasthttp.Client{},
		timezone: loc,
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
		var jsonData []byte
		jsonData, err = json.Marshal(data)
		if err != nil {
			return fmt.Errorf("encode data: %w", err)
		}

		req.SetBodyRaw(jsonData)
	}

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	if err = s.client.Do(req, resp); err != nil {
		return fmt.Errorf("call syodo: %w", err)
	}

	if statusCode := resp.StatusCode(); statusCode != fasthttp.StatusOK {
		return fmt.Errorf("call syodo bad status: %d", statusCode)
	}

	if err = json.Unmarshal(resp.Body(), result); err != nil {
		return fmt.Errorf("decode result: %w", err)
	}

	return nil
}

type orderDTO struct {
	ID         string `json:"id"`
	CategoryID string `json:"category_id"`
	Title      string `json:"title"`
	Amount     int    `json:"qty"`
}

type deliveryDTO struct {
	Type string `json:"type"`
	Zone string `json:"serviceArea"`
}

type priceRequest struct {
	Order             []orderDTO  `json:"order"`
	DeliveryDetails   deliveryDTO `json:"deliveryDetails"`
	SelectedPromotion string      `json:"selectedPromotion"`
}

type priceResponse struct {
	Delivery int `json:"delivery"`
	Discount int `json:"discount"`
}

func orderToDTO(order OrderDetails) []orderDTO {
	dto := make([]orderDTO, len(order.Request.Products))
	for i, p := range order.Request.Products {
		dto[i] = orderDTO{
			ID:         p.ID,
			CategoryID: p.CategoryID,
			Title:      p.Title,
			Amount:     p.Amount,
		}
	}

	return dto
}

// CalculatePrice returns calculated price depending on order details and delivery zone
func (s *SyodoService) CalculatePrice(order OrderDetails, zone DeliveryZone, selfPickup bool) (int, error) {
	requestOrder := orderToDTO(order)

	shippingType := shippingTypeDelivery
	shippingPromo := shippingPromo4Plus1
	if selfPickup {
		shippingType = shippingTypeSelfPickup
		shippingPromo = shippingPromoSelfPickup
	}

	priceReq := &priceRequest{
		Order: requestOrder,
		DeliveryDetails: deliveryDTO{
			Type: shippingType,
			Zone: zone,
		},
		SelectedPromotion: shippingPromo,
	}

	var priceResp priceResponse
	if err := s.call("/price", fasthttp.MethodPost, priceReq, &priceResp); err != nil {
		return 0, fmt.Errorf("price API: %w", err)
	}

	return priceResp.Delivery - priceResp.Discount, nil
}

type contactDTO struct {
	Name  string `json:"name"`
	Phone string `json:"phone"`
}

type deliveryDetailsDTO struct {
	Type        string `json:"type"`
	Date        string `json:"date"`
	Time        string `json:"time"`
	DontCall    bool   `json:"dontCall"`
	Comments    string `json:"comments"`
	Address     string `json:"address"`
	Entrance    string `json:"entrance"`
	Apt         string `json:"apt"`
	ECode       string `json:"eCode"`
	ServiceArea string `json:"serviceArea"`
}

type paymentDTO struct {
	PaymentMethod string `json:"paymentMethod"`
	RestFrom      string `json:"restFrom"`
}

type infoDTO struct {
	NoNapkins       bool `json:"noNapkins"`
	Persons         int  `json:"persons"`
	TrainingPersons int  `json:"trainingPersons"`
}

type checkoutRequest struct {
	Description       string             `json:"description"`
	Currency          string             `json:"currency"`
	Language          string             `json:"language"`
	ContactDetails    contactDTO         `json:"contactDetails"`
	DeliveryDetails   deliveryDetailsDTO `json:"deliveryDetails"`
	PaymentDetails    paymentDTO         `json:"paymentDetails"`
	Info              infoDTO            `json:"info"`
	OrderDetails      []orderDTO         `json:"orderDetails"`
	SelectedPromotion string             `json:"selectedPromotion"`
}

type checkoutResponse struct {
	Data      string `json:"data"`
	Signature string `json:"signature"`
	OrderID   string `json:"orderId"`
}

type checkoutDTO struct {
	OrderID            string  `json:"order_id"`
	PublicKey          string  `json:"public_key"`
	Version            string  `json:"version"`
	Action             string  `json:"action"`
	Amount             float64 `json:"amount"`
	Currency           string  `json:"currency"`
	Description        string  `json:"description"`
	Language           string  `json:"language"`
	ProductDescription string  `json:"product_description"`
	ExpiredDate        string  `json:"expired_date"`
	ResultURL          string  `json:"result_url"`
	ServerURL          string  `json:"server_url"`
	SenderAddress      string  `json:"sender_address"`
	SenderCity         string  `json:"sender_city"`
	SenderFirstName    string  `json:"sender_first_name"`
	Info               string  `json:"info"`
	Alg                string  `json:"alg"`
}

// Checkout registers order in Syodo services
func (s *SyodoService) Checkout(order *OrderDetails) error {
	if order == nil {
		return errors.New("nil order checkout")
	}

	requestOrder := orderToDTO(*order)

	area := order.ShippingOptionID
	deliveryType := shippingTypeDelivery
	shippingPromo := shippingPromo4Plus1
	if order.ShippingOptionID == SelfPickup {
		area = ""
		deliveryType = shippingTypeSelfPickup
		shippingPromo = shippingPromoSelfPickup
	}

	shipping := order.OrderInfo.ShippingAddress
	address := shipping.StreetLine1
	if shipping.StreetLine2 != "" {
		address += " " + shipping.StreetLine2
	}
	address += ", " + shipping.City
	if shipping.State != "" {
		address += ", " + shipping.State
	}

	checkoutReq := checkoutRequest{
		Description: fmt.Sprintf("Замовлення з Telegram: %s, %s",
			time.Now().In(s.timezone).Format("2006-01-02 15:04"), order.OrderID),
		Currency: currency,
		Language: "ua",
		ContactDetails: contactDTO{
			Name:  order.OrderInfo.Name,
			Phone: order.OrderInfo.PhoneNumber,
		},
		DeliveryDetails: deliveryDetailsDTO{
			Type:        deliveryType,
			DontCall:    order.Request.DoNotCall,
			Comments:    order.Request.Comment,
			Address:     address,
			ServiceArea: area,
		},
		PaymentDetails: paymentDTO{
			PaymentMethod: "Онлайн",
		},
		Info: infoDTO{
			NoNapkins:       order.Request.NoNapkins,
			Persons:         order.Request.CutleryCount,
			TrainingPersons: order.Request.TrainingCutleryCount,
		},
		OrderDetails:      requestOrder,
		SelectedPromotion: shippingPromo,
	}

	var checkoutResp checkoutResponse
	if err := s.call("/payments/checkout", fasthttp.MethodPost, checkoutReq, &checkoutResp); err != nil {
		return fmt.Errorf("checkout API: %w", err)
	}

	data, err := base64.StdEncoding.DecodeString(checkoutResp.Data)
	if err != nil {
		return fmt.Errorf("decode data: %w", err)
	}

	var checkout checkoutDTO
	if err = json.Unmarshal(data, &checkout); err != nil {
		return fmt.Errorf("unmarshal data: %w", err)
	}

	order.ExternalOrderID = checkout.OrderID
	order.OrderURL = checkout.ResultURL
	order.TotalAmount = checkout.Amount

	return nil
}
