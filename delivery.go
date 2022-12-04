package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/mymmrac/telego"
	"googlemaps.github.io/maps"

	"github.com/mymmrac/syodo-telegram-bot/config"
	"github.com/mymmrac/syodo-telegram-bot/logger"
)

// DeliveryZone represents delivery zones
type DeliveryZone = string

// Defined delivery zones
const (
	ZoneUnknown DeliveryZone = ""
	ZoneGreen   DeliveryZone = "green"
	ZoneYellow  DeliveryZone = "yellow"
	ZoneRed     DeliveryZone = "red"
)

// SelfPickup represents self-pickup delivery method with -10% promotion
const SelfPickup = "self_pickup"

// SelfPickup4Plus1 represents self-pickup delivery method with 4+1 promotion
const SelfPickup4Plus1 = "self_pickup_4_plus_1"

// DeliveryStrategy represents model of calculation delivery zones by addresses
type DeliveryStrategy struct {
	cfg    *config.Config
	log    logger.Logger
	client *maps.Client
}

// NewDeliveryStrategy creates new DeliveryStrategy
func NewDeliveryStrategy(cfg *config.Config, log logger.Logger) (*DeliveryStrategy, error) {
	client, err := maps.NewClient(maps.WithAPIKey(cfg.App.GoogleMapsAPIKey))
	if err != nil {
		return nil, fmt.Errorf("create maps client: %w", err)
	}

	return &DeliveryStrategy{
		cfg:    cfg,
		log:    log,
		client: client,
	}, nil
}

// CalculateLocation returns delivery location by its address
func (s *DeliveryStrategy) CalculateLocation(shipping telego.ShippingAddress) maps.LatLng {
	country := strings.ToLower(shipping.CountryCode)
	if country != "ua" {
		s.log.Errorf("Calculate zone: non UA country: %s", country)
		return maps.LatLng{}
	}

	ctx, cancel := context.WithTimeout(context.Background(), s.cfg.Settings.RequestTimeout)
	defer cancel()

	results, err := s.client.Geocode(ctx, &maps.GeocodingRequest{
		Components: map[maps.Component]string{
			maps.ComponentCountry:            country,
			maps.ComponentLocality:           shipping.City,
			maps.ComponentAdministrativeArea: shipping.State,
			maps.ComponentRoute:              strings.TrimSpace(shipping.StreetLine1 + " " + shipping.StreetLine2),
		},
		Bounds:   approximateBounds,
		Region:   strings.ToLower(shipping.CountryCode),
		Language: "uk",
	})
	if err != nil {
		s.log.Errorf("Calculate zone: geocode for %+v, error: %s", shipping, err)
		return maps.LatLng{}
	}

	if len(results) == 0 {
		s.log.Debugf("Calculate zone: no address found for %+v", shipping)
		return maps.LatLng{}
	}

	chosenResult := results[0]
	location := chosenResult.Geometry.Location

	s.log.Debugf("Calculate zone: chosen address for %+v was: location: %s, address: %s",
		shipping, location.String(), chosenResult.FormattedAddress)

	return location
}

//nolint:gomnd
var approximateBounds = &maps.LatLngBounds{
	NorthEast: maps.LatLng{
		Lat: 50.061937,
		Lng: 24.386862,
	},
	SouthWest: maps.LatLng{
		Lat: 48.71841570388124,
		Lng: 23.471838912294967,
	},
}
