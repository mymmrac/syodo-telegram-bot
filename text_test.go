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
	}

	templates := []struct {
		key  string
		data any
	}{
		{
			"start",
			&telego.Message{
				From: &telego.User{},
			},
		},
		{
			"help",
			&telego.Message{
				From: &telego.User{},
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
