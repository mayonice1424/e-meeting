package models

type CreateRoom struct {
	Name            string `json:"name"`
	PricePerHour    string `json:"price_per_hour"`
	ImageURL        string `json:"image_url"`
	Capacity        string `json:"capacity"`
	Type            string `json:"type"`
}