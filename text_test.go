package main

import (
	"testing"

	"github.com/mymmrac/telego"
)

func TestTextData(t *testing.T) {
	data, err := LoadTextData("text.toml")
	if err != nil {
		t.Error(err)
	}

	keys := []string{
		"startDescription",
		"helpDescription",
		"siteButtonText",
		"instagramButtonText",
		"facebookButtonText",
		"siteURL",
		"instagramURL",
		"facebookURL",
		"menuButton",
		"orderNotFoundError",
		"orderDeliveryError",
		"orderDescription",
	}

	templates := []struct {
		key  string
		data any
	}{
		{
			key: "start",
			data: telego.Message{
				From: &telego.User{},
			},
		},
		{
			key: "help",
			data: telego.Message{
				From: &telego.User{},
			},
		},
		{
			key: "successPayment",
			data: telego.SuccessfulPayment{
				OrderInfo: &telego.OrderInfo{
					ShippingAddress: &telego.ShippingAddress{},
				},
			},
		},
	}

	for _, text := range keys {
		_ = data.Text(text)
	}

	for _, temp := range templates {
		_ = data.Temp(temp.key, temp.data)
	}
}
