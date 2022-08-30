package main

import (
	"context"
	"fmt"

	"github.com/golang/geo/s2"
	"github.com/mymmrac/telego"
	"googlemaps.github.io/maps"

	"github.com/mymmrac/syodo-telegram-bot/config"
	"github.com/mymmrac/syodo-telegram-bot/logger"
)

// DeliveryZone represents delivery zones
type DeliveryZone string

// Defined delivery zones
const (
	ZoneUnknown DeliveryZone = ""
	ZoneGreen   DeliveryZone = "green"
	ZoneYellow  DeliveryZone = "yellow"
	ZoneRed     DeliveryZone = "red"
)

// SelfPickup represents self pickup delivery method
const SelfPickup = "self_pickup"

// DeliveryMethodIDs lists all delivery methods
var DeliveryMethodIDs = map[string]struct{}{
	string(ZoneGreen):  {},
	string(ZoneYellow): {},
	string(ZoneRed):    {},
	SelfPickup:         {},
}

// DeliveryStrategy represents model of calculation delivery zones by addresses
type DeliveryStrategy struct {
	log    logger.Logger
	client *maps.Client

	greenPolygon  *s2.Polygon
	yellowPolygon *s2.Polygon
	redPolygon    *s2.Polygon
}

// NewDeliveryStrategy creates new DeliveryStrategy
func NewDeliveryStrategy(cfg *config.Config, log logger.Logger) (*DeliveryStrategy, error) {
	client, err := maps.NewClient(maps.WithAPIKey(cfg.App.GoogleMapsToken))
	if err != nil {
		return nil, fmt.Errorf("create maps client: %w", err)
	}

	greenPolygon := s2.PolygonFromLoops([]*s2.Loop{s2.LoopFromPoints([]s2.Point{
		s2.PointFromLatLng(s2.LatLngFromDegrees(0, 0)),
	})})

	return &DeliveryStrategy{
		log:    log,
		client: client,

		greenPolygon: greenPolygon,
	}, nil
}

// CalculateZone returns delivery zone by its address
func (s *DeliveryStrategy) CalculateZone(shipping telego.ShippingAddress) DeliveryZone {
	ctx := context.Background()
	results, err := s.client.Geocode(ctx, &maps.GeocodingRequest{
		Address:      "",
		Components:   nil,
		Bounds:       nil,
		Region:       "",
		LatLng:       nil,
		ResultType:   nil,
		LocationType: nil,
		PlaceID:      "",
		Language:     "",
		Custom:       nil,
	})
	if err != nil {
		s.log.Errorf("Calculate zone: geocode for %+v, error: %s", shipping, err)
		return ZoneUnknown
	}

	if len(results) == 0 {
		s.log.Debugf("Calculate zone: no address found for %+v", shipping)
		return ZoneUnknown
	}

	chosenResult := results[0]
	s.log.Debugf("Calculate zone: chosen address for %+v was: %s", shipping, chosenResult.FormattedAddress)

	location := chosenResult.Geometry.Location
	point := s2.PointFromLatLng(s2.LatLngFromDegrees(location.Lat, location.Lng))

	if s.greenPolygon.ContainsPoint(point) {
		return ZoneGreen
	}

	if s.yellowPolygon.ContainsPoint(point) {
		return ZoneYellow
	}

	if s.redPolygon.ContainsPoint(point) {
		return ZoneRed
	}

	return ZoneUnknown
}
