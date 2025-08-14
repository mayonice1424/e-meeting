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
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

type SuccessResponseRoom struct {
	Message string      `json:"message"`
}

type SuccessResponseRooms struct {
	Message   string      `json:"message"`
	Data      []RoomById  `json:"data"`
	Page      int         `json:"page"`      
	PageSize  int         `json:"pageSize"`   
	TotalPage int         `json:"totalPage"`  
	TotalData int         `json:"totalData"` 
}