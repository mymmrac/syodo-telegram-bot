package main

import "time"

// OrderProduct represents a single item in order
type OrderProduct struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	Price      int    `json:"price"`
	Amount     int    `json:"amount"`
	CategoryID string `json:"categoryID"`
}

// OrderRequest represents order info sent by user
type OrderRequest struct {
	AppData              string         `json:"appData"`
	Products             []OrderProduct `json:"products"`
	DoNotCall            bool           `json:"doNotCall"`
	NoNapkins            bool           `json:"noNapkins"`
	CutleryCount         int            `json:"cutleryCount"`
	TrainingCutleryCount int            `json:"trainingCutleryCount"`
	Comment              string         `json:"comment"`
}

// OrderDetails represents full order info
type OrderDetails struct {
	Request   OrderRequest `json:"request"`
	CreatedAt time.Time    `json:"createdAt"`
}
