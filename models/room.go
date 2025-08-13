package models

type CreateRoom struct {
	Name         string `json:"name"`
	Type         string `json:"type"`
	Picture      string `json:"picture"`
	PricePerHour int    `json:"price_per_hour"`
	Capacity     int    `json:"capacity"`
}

type RoomById struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Type         string `json:"type"`
	Picture      string `json:"picture"`
	PricePerHour int    `json:"price_per_hour"`
	Capacity     int    `json:"capacity"`
}

type SuccessResponseRoom struct {
	Message string      `json:"message"`
}

type SuccessResponseRooms struct {
	Message string      `json:"message"`
	Data    []RoomById  `json:"data"`
}