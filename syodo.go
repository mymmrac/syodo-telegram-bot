package main

import (
	"github.com/valyala/fasthttp"

	"github.com/mymmrac/syodo-telegram-bot/config"
)

type SyodoService struct {
	cfg    *config.Config
	client *fasthttp.Client
}

func NewSyodoService(cfg *config.Config) *SyodoService {
	return &SyodoService{
		cfg:    cfg,
		client: &fasthttp.Client{},
	}
}

func (s *SyodoService) CalculatePrice(order OrderDetails, zone DeliveryZone, selfPickup bool) (int, error) {
	return 0, nil
}
