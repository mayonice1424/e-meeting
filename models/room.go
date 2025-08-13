package models

<<<<<<< HEAD
type Room struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Type         string `json:"type"`
	Picture      string `json:"picture"`
	PricePerHour int    `json:"price_per_hour"`
	Capacity     int    `json:"capacity"`
}
=======
type CreateRoom struct {
	Name            string `json:"name"`
	PricePerHour    string `json:"price_per_hour"`
	ImageURL        string `json:"image_url"`
	Capacity        string `json:"capacity"`
	Type            string `json:"type"`
}
>>>>>>> adec5aa7faf61884b208b07fb9c0a278beec7930
