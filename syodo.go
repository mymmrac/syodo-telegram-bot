package main

import (
	"github.com/valyala/fasthttp"

	"github.com/mymmrac/syodo-telegram-bot/config"
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

// CalculatePrice returns calculated price depending on order details and delivery zone
func (s *SyodoService) CalculatePrice(order OrderDetails, zone DeliveryZone, selfPickup bool) (int, error) {
	return 0, nil
}
