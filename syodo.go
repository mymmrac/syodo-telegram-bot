package main

import (
	"crypto/sha1" //nolint:gosec
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/mymmrac/telego"
	"github.com/valyala/fasthttp"
	"googlemaps.github.io/maps"

	"github.com/mymmrac/syodo-telegram-bot/config"
	"github.com/mymmrac/syodo-telegram-bot/logger"
)

const (
	contentTypeJSON = "application/json"
	contentTypeURL  = "application/x-www-form-urlencoded"
	authHeader      = "x-api-key"
)

const (
	deliveryTypeDelivery = "delivery"

	shippingTypeDelivery   = "Доставка"
	shippingTypeSelfPickup = "Самовивіз"

	promo4Plus1     = "4+1"
	promoSelfPickup = "Самовивіз"
)

// SyodoService represents a type to interact with Syodo API
type SyodoService struct {
	cfg      *config.Config
	log      logger.Logger
	client   *fasthttp.Client
	timezone *time.Location
}

// NewSyodoService creates new SyodoService
func NewSyodoService(cfg *config.Config, log logger.Logger) *SyodoService {
	loc, err := time.LoadLocation("Europe/Kiev")
	assert(err == nil, fmt.Errorf("load timezone: %w", err))

	return &SyodoService{
		cfg:      cfg,
		log:      log,
		client:   &fasthttp.Client{},
		timezone: loc,
	}
}

func (s *SyodoService) callJSON(path, method string, data, result any) error {
	var jsonData []byte
	if data != nil {
		var err error
		jsonData, err = json.Marshal(data)
		if err != nil {
			return fmt.Errorf("encode data: %w", err)
		}
	}

	return s.call(path, method, contentTypeJSON, jsonData, result)
}

func (s *SyodoService) callURL(path, method, data string, result any) error {
	return s.call(path, method, contentTypeURL, []byte(data), result)
}

func (s *SyodoService) call(path, method, contentType string, data []byte, result any) error {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	apiURL, err := url.JoinPath(s.cfg.App.SyodoAPIURL, path)
	if err != nil {
		return fmt.Errorf("join path: %w", err)
	}

	req.SetRequestURI(apiURL)

	req.Header.SetMethod(method)
	req.Header.SetContentType(contentType)
	req.Header.Set(authHeader, s.cfg.App.SyodoAPIKey)

	if data != nil {
		req.SetBodyRaw(data)
	}

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	if err = s.client.DoTimeout(req, resp, s.cfg.Settings.RequestTimeout); err != nil {
		return fmt.Errorf("call syodo: %w", err)
	}

	if statusCode := resp.StatusCode(); statusCode != fasthttp.StatusOK {
		return fmt.Errorf("call syodo bad status: %d", statusCode)
	}

	if result != nil {
		body := resp.Body()
		s.log.Debugf("Request to %q: data: %s, response %s", path, string(data), string(body))

		if err = json.Unmarshal(body, result); err != nil {
			return fmt.Errorf("decode result: %w", err)
		}
	}

	return nil
}

type orderDTO struct {
	ID         string `json:"id"`
	CategoryID string `json:"category_id"`
	Title      string `json:"title"`
	Amount     int    `json:"qty"`
}

type pointDTO struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type deliveryDTO struct {
	Type  string   `json:"type"`
	Point pointDTO `json:"point,omitempty"`
}

type priceRequest struct {
	Order             []orderDTO  `json:"order"`
	DeliveryDetails   deliveryDTO `json:"deliveryDetails"`
	SelectedPromotion string      `json:"selectedPromotion"`
}

// PriceResponse represents calculated price of order
type PriceResponse struct {
	Delivery    int    `json:"delivery"`
	Discount    int    `json:"discount"`
	ServiceArea string `json:"service_area"`
}

// CalculatePriceDelivery returns calculated price depending on order details and delivery zone
func (s *SyodoService) CalculatePriceDelivery(
	products []OrderProduct, location maps.LatLng, promotion string,
) (PriceResponse, error) {
	return s.calculatePrice(products, shippingTypeDelivery, location, promotion)
}

// CalculatePriceSelfPickup returns calculated price depending on order details
func (s *SyodoService) CalculatePriceSelfPickup(products []OrderProduct, promotion string) (PriceResponse, error) {
	resp, err := s.calculatePrice(products, shippingTypeSelfPickup, maps.LatLng{}, promotion)
	return resp, err
}

func (s *SyodoService) calculatePrice(
	products []OrderProduct, shippingType string, location maps.LatLng, promotion string,
) (PriceResponse, error) {
	requestOrder := orderToDTO(products)

	priceReq := &priceRequest{
		Order: requestOrder,
		DeliveryDetails: deliveryDTO{
			Type: shippingType,
			Point: pointDTO{
				Lat: location.Lat,
				Lng: location.Lng,
			},
		},
		SelectedPromotion: promotion,
	}

	var priceResp PriceResponse
	if err := s.callJSON("/price", fasthttp.MethodPost, priceReq, &priceResp); err != nil {
		return PriceResponse{}, fmt.Errorf("price API: %w", err)
	}

	return priceResp, nil
}

func orderToDTO(products []OrderProduct) []orderDTO {
	dto := make([]orderDTO, len(products))
	for i, p := range products {
		dto[i] = orderDTO{
			ID:         p.ID,
			CategoryID: p.CategoryID,
			Title:      p.Title,
			Amount:     p.Amount,
		}
	}

	return dto
}

type contactDTO struct {
	Name  string `json:"name"`
	Phone string `json:"phone"`
}

type deliveryDetailsDTO struct {
	Type           string `json:"type"`
	Date           string `json:"date"`
	Time           string `json:"time"`
	Comments       string `json:"comments"`
	Address        string `json:"address"`
	Entrance       string `json:"entrance"`
	Apt            string `json:"apt"`
	ECode          string `json:"eCode"`
	ServiceArea    string `json:"serviceArea"`
	PickupLocation string `json:"pickupLocation"`
	DontCall       bool   `json:"dontCall"`
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
//
//nolint:cyclop
func (s *SyodoService) Checkout(order *OrderDetails) error {
	if order == nil {
		return errors.New("nil order checkout")
	}

	var deliveryType string
	if order.Request.DeliveryType == deliveryTypeDelivery {
		deliveryType = shippingTypeDelivery
	} else {
		deliveryType = shippingTypeSelfPickup
	}

	var pickupLocation string
	switch order.Request.DeliveryType {
	case "self_pickup_1":
		pickupLocation = "1"
	case "self_pickup_2":
		pickupLocation = "2"
	}

	checkoutReq := checkoutRequest{
		Description: fmt.Sprintf("Замовлення з Telegram: %s, #%s",
			time.Now().In(s.timezone).Format("2006-01-02 15:04"), order.OrderID),
		Currency: currency,
		Language: "ua",
		ContactDetails: contactDTO{
			Name:  order.Request.Name,
			Phone: order.Request.Phone,
		},
		DeliveryDetails: deliveryDetailsDTO{
			Type:           deliveryType,
			DontCall:       order.Request.DoNotCall,
			Comments:       order.Request.Comment,
			Address:        order.Request.Address + ", м. " + order.Request.City,
			Entrance:       order.Request.Entrance,
			Apt:            order.Request.Apartment,
			ECode:          order.Request.ECode,
			ServiceArea:    order.ServiceArea,
			PickupLocation: pickupLocation,
		},
		PaymentDetails: paymentDTO{
			PaymentMethod: "Онлайн",
		},
		Info: infoDTO{
			NoNapkins:       order.Request.NoNapkins,
			Persons:         order.Request.CutleryCount,
			TrainingPersons: order.Request.TrainingCutleryCount,
		},
		OrderDetails:      orderToDTO(order.Request.Products),
		SelectedPromotion: order.Request.Promotion,
	}

	var checkoutResp checkoutResponse
	if err := s.callJSON("/payments/checkout", fasthttp.MethodPost, checkoutReq, &checkoutResp); err != nil {
		return fmt.Errorf("checkout API: %w", err)
	}

	if !s.cfg.Settings.TestMode {
		signature := sign(checkoutResp.Data, s.cfg.App.LiqPayPrivetKeyEnv)
		if signature != checkoutResp.Signature {
			return fmt.Errorf("checkout signature does not match")
		}
	}

	data, err := base64.StdEncoding.DecodeString(checkoutResp.Data)
	if err != nil {
		return fmt.Errorf("decode data: %w", err)
	}
	s.log.Debugf("Checkout data: %s", string(data))

	var checkout checkoutDTO
	if err = json.Unmarshal(data, &checkout); err != nil {
		return fmt.Errorf("unmarshal data: %w", err)
	}

	order.ExternalOrderID = checkout.OrderID
	if s.cfg.Settings.TestMode {
		order.OrderURL = strings.Replace(checkout.ResultURL, "APP_LIQ_PAY_RESULT_URL",
			"https://www.syodo.com.ua/ua/success", 1)
	} else {
		order.OrderURL = checkout.ResultURL
	}
	order.TotalAmount = checkout.Amount

	return nil
}

type successPaymentDTO struct {
	PayType                 string `json:"paytype"`
	Status                  string `json:"status"`
	ProviderPaymentChargeID string `json:"liqpay_order_id"`
	OrderID                 string `json:"transaction_id"`
	TotalAmount             int    `json:"amount"`
	ExternalOrderID         string `json:"order_id"`
}

// SuccessPayment confirm success payment in Syodo
func (s *SyodoService) SuccessPayment(payment *telego.SuccessfulPayment, externalOrderID string) error {
	successPayment := successPaymentDTO{
		PayType:                 "telegram",
		Status:                  "success",
		ProviderPaymentChargeID: payment.ProviderPaymentChargeID,
		OrderID:                 payment.InvoicePayload,
		TotalAmount:             payment.TotalAmount,
		ExternalOrderID:         externalOrderID,
	}

	dataJSON, err := json.Marshal(successPayment)
	if err != nil {
		return fmt.Errorf("marshal JSON: %w", err)
	}

	data := base64.StdEncoding.EncodeToString(dataJSON)
	signature := sign(data, s.cfg.App.LiqPayPrivetKeyEnv)
	fullData := fmt.Sprintf("signature=%s&data=%s", signature, data)

	s.log.Debugf("Payments callback data: %s", fullData)

	if err = s.callURL("/payments/callback", fasthttp.MethodPost, fullData, nil); err != nil {
		return fmt.Errorf("success payment API: %w", err)
	}

	return nil
}

func sign(data, key string) string {
	//nolint:gosec
	hash := sha1.New()
	hash.Write([]byte(key))
	hash.Write([]byte(data))
	hash.Write([]byte(key))
	return base64.StdEncoding.EncodeToString(hash.Sum(nil))
}
