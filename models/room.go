package models

type Room struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Type         string `json:"type"`
	Picture      string `json:"picture"`
	PricePerHour int    `json:"price_per_hour"`
	Capacity     int    `json:"capacity"`
}
