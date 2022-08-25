package main

type OrderProduct struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	Price      int    `json:"price"`
	Amount     int    `json:"amount"`
	CategoryID string `json:"categoryID"`
}

type OrderRequest struct {
	AppData              string         `json:"appData"`
	Products             []OrderProduct `json:"products"`
	DoNotCall            bool           `json:"doNotCall"`
	NoNapkins            bool           `json:"noNapkins"`
	CutleryCount         int            `json:"cutleryCount"`
	TrainingCutleryCount int            `json:"trainingCutleryCount"`
	Comment              string         `json:"comment"`
}
