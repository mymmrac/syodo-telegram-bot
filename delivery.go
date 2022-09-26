package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/golang/geo/s2"
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

// SelfPickup represents self pickup delivery method with -10% promotion
const SelfPickup = "self_pickup"

// SelfPickup4Plus1 represents self pickup delivery method with 4+1 promotion
const SelfPickup4Plus1 = "self_pickup_4_plus_1"

// DeliveryMethodIDs lists all delivery methods
var DeliveryMethodIDs = map[string]struct{}{
	ZoneGreen:        {},
	ZoneYellow:       {},
	ZoneRed:          {},
	SelfPickup:       {},
	SelfPickup4Plus1: {},
}

// DeliveryStrategy represents model of calculation delivery zones by addresses
type DeliveryStrategy struct {
	cfg    *config.Config
	log    logger.Logger
	client *maps.Client

	greenPolygon  *s2.Polygon
	yellowPolygon *s2.Polygon
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

		greenPolygon:  constructPolygon(greenAreaPoints),
		yellowPolygon: constructPolygon(yellowAreaPoints),
	}, nil
}

// CalculateZone returns delivery zone by its address
func (s *DeliveryStrategy) CalculateZone(shipping telego.ShippingAddress) DeliveryZone {
	country := strings.ToLower(shipping.CountryCode)
	if country != "ua" {
		s.log.Errorf("Calculate zone: non UA country: %s", country)
		return ZoneUnknown
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
		return ZoneUnknown
	}

	if len(results) == 0 {
		s.log.Debugf("Calculate zone: no address found for %+v", shipping)
		return ZoneUnknown
	}

	chosenResult := results[0]
	location := chosenResult.Geometry.Location

	s.log.Debugf("Calculate zone: chosen address for %+v was: location: %s, address: %s",
		shipping, location.String(), chosenResult.FormattedAddress)

	point := s2.PointFromLatLng(s2.LatLngFromDegrees(location.Lat, location.Lng))

	if s.greenPolygon.ContainsPoint(point) {
		return ZoneGreen
	}

	if s.yellowPolygon.ContainsPoint(point) {
		return ZoneYellow
	}

	return ZoneRed
}

func constructPolygon(latLngPoints []maps.LatLng) *s2.Polygon {
	points := make([]s2.Point, len(latLngPoints))
	for i, p := range latLngPoints {
		points[i] = s2.PointFromLatLng(s2.LatLngFromDegrees(p.Lat, p.Lng))
	}

	polygon := s2.PolygonFromLoops([]*s2.Loop{s2.LoopFromPoints(points)})
	polygon.Invert()
	return polygon
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

//nolint:gomnd
var greenAreaPoints = []maps.LatLng{
	{Lat: 49.778702, Lng: 23.980260},
	{Lat: 49.779991, Lng: 23.976215},
	{Lat: 49.781509, Lng: 23.976307},
	{Lat: 49.785319, Lng: 23.978164},
	{Lat: 49.801215, Lng: 23.980265},
	{Lat: 49.801792, Lng: 23.973725},
	{Lat: 49.804527, Lng: 23.969418},
	{Lat: 49.808931, Lng: 23.967124},
	{Lat: 49.814717, Lng: 23.955665},
	{Lat: 49.822046, Lng: 23.978470},
	{Lat: 49.821869, Lng: 23.988816},
	{Lat: 49.821536, Lng: 24.005082},
	{Lat: 49.826604, Lng: 24.009824},
	{Lat: 49.819949, Lng: 24.020654},
	{Lat: 49.810505, Lng: 24.047324},
	{Lat: 49.806584, Lng: 24.047000},
	{Lat: 49.803530, Lng: 24.047263},
	{Lat: 49.795389, Lng: 24.054545},
	{Lat: 49.790065, Lng: 24.034081},
	{Lat: 49.785749, Lng: 24.033867},
	{Lat: 49.783298, Lng: 24.028115},
	{Lat: 49.780526, Lng: 24.027478},
	{Lat: 49.778248, Lng: 24.025505},
	{Lat: 49.775246, Lng: 24.025282},
}

//nolint:gomnd
var yellowAreaPoints = []maps.LatLng{
	{Lat: 49.778702, Lng: 23.980260},
	{Lat: 49.779991, Lng: 23.976215},
	{Lat: 49.781509, Lng: 23.976307},
	{Lat: 49.785319, Lng: 23.978164},
	{Lat: 49.801215, Lng: 23.980265},
	{Lat: 49.801792, Lng: 23.973725},
	{Lat: 49.804527, Lng: 23.969418},
	{Lat: 49.808931, Lng: 23.967124},
	{Lat: 49.814717, Lng: 23.955665},
	{Lat: 49.818809, Lng: 23.969699},
	{Lat: 49.821845, Lng: 23.967545},
	{Lat: 49.824083, Lng: 23.946517},
	{Lat: 49.827042, Lng: 23.947324},
	{Lat: 49.834094, Lng: 23.986771},
	{Lat: 49.836923, Lng: 24.001425},
	{Lat: 49.837410, Lng: 24.001908},
	{Lat: 49.838450, Lng: 24.000833},
	{Lat: 49.842022, Lng: 23.997889},
	{Lat: 49.845856, Lng: 23.996328},
	{Lat: 49.845381, Lng: 24.002248},
	{Lat: 49.846416, Lng: 24.005670},
	{Lat: 49.846727, Lng: 24.006220},
	{Lat: 49.843001, Lng: 24.015702},
	{Lat: 49.842879, Lng: 24.018582},
	{Lat: 49.844922, Lng: 24.026596},
	{Lat: 49.839412, Lng: 24.030534},
	{Lat: 49.840661, Lng: 24.035876},
	{Lat: 49.839880, Lng: 24.038529},
	{Lat: 49.840033, Lng: 24.041216},
	{Lat: 49.840825, Lng: 24.046452},
	{Lat: 49.836115, Lng: 24.069121},
	{Lat: 49.835258, Lng: 24.069484},
	{Lat: 49.831127, Lng: 24.069353},
	{Lat: 49.823070, Lng: 24.077181},
	{Lat: 49.816338, Lng: 24.079489},
	{Lat: 49.810172, Lng: 24.085803},
	{Lat: 49.810350, Lng: 24.081295},
	{Lat: 49.808931, Lng: 24.080687},
	{Lat: 49.807531, Lng: 24.079591},
	{Lat: 49.807130, Lng: 24.078352},
	{Lat: 49.806476, Lng: 24.077306},
	{Lat: 49.805205, Lng: 24.077197},
	{Lat: 49.803696, Lng: 24.076488},
	{Lat: 49.802390, Lng: 24.073310},
	{Lat: 49.800527, Lng: 24.074029},
	{Lat: 49.798623, Lng: 24.072147},
	{Lat: 49.794635, Lng: 24.073525},
	{Lat: 49.789424, Lng: 24.074993},
	{Lat: 49.785752, Lng: 24.076810},
	{Lat: 49.783936, Lng: 24.075482},
	{Lat: 49.782271, Lng: 24.068724},
	{Lat: 49.780808, Lng: 24.068565},
	{Lat: 49.780835, Lng: 24.055429},
	{Lat: 49.781585, Lng: 24.051762},
	{Lat: 49.785707, Lng: 24.051276},
	{Lat: 49.785534, Lng: 24.048430},
	{Lat: 49.788966, Lng: 24.047567},
	{Lat: 49.788928, Lng: 24.044085},
	{Lat: 49.792141, Lng: 24.041710},
	{Lat: 49.790065, Lng: 24.034081},
	{Lat: 49.785749, Lng: 24.033867},
	{Lat: 49.783298, Lng: 24.028115},
	{Lat: 49.780526, Lng: 24.027478},
	{Lat: 49.778248, Lng: 24.025505},
	{Lat: 49.775246, Lng: 24.025282},
}
