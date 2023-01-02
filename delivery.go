package main

import (
	"context"
	"fmt"

	"googlemaps.github.io/maps"

	"github.com/mymmrac/syodo-telegram-bot/config"
	"github.com/mymmrac/syodo-telegram-bot/logger"
)

// DeliveryZone represents delivery zones
type DeliveryZone = string

// Defined delivery zones
const (
	ZoneGreen  DeliveryZone = "green"
	ZoneYellow DeliveryZone = "yellow"
	ZoneRed    DeliveryZone = "red"
)

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
func (s *DeliveryStrategy) CalculateLocation(order OrderRequest) (maps.LatLng, error) {
	ctx, cancel := context.WithTimeout(context.Background(), s.cfg.Settings.RequestTimeout)
	defer cancel()

	results, err := s.client.Geocode(ctx, &maps.GeocodingRequest{
		Components: map[maps.Component]string{
			maps.ComponentCountry:  "ua",
			maps.ComponentLocality: order.City,
			maps.ComponentRoute:    order.Address,
		},
		Bounds:   approximateBounds,
		Region:   "ua",
		Language: "uk",
	})
	if err != nil {
		return maps.LatLng{}, fmt.Errorf("geocode for %+v, error: %w", order, err)
	}

	if len(results) == 0 {
		return maps.LatLng{}, fmt.Errorf("no address found for %+v", order)
	}

	chosenResult := results[0]
	location := chosenResult.Geometry.Location

	s.log.Debugf("Calculate zone: chosen address for %+v was: location: %s, address: %s",
		order, location.String(), chosenResult.FormattedAddress)

	return location, nil
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
