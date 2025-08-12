package models

type Snack struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Price    int    `json:"price"`
	Unit     string `json:"unit"`
	Category string `json:"category"`
}

type SuccessResponseSnack struct {
	Data    []Snack `json:"data"`
	Message string  `json:"message"`
}


type ErrorResponseSnack struct {
	Message string `json:"message"`
}