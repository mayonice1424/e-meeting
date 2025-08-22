package models

import (
	"time"
)
type SuccessResponseReservationCalculation struct {
	Message string                         `json:"message"`
	Data    ReservationCalculationData     `json:"data"`
}
// ReservationCalculationData represents the calculated reservation data
type ReservationCalculationData struct {
	Rooms         []RoomCalculation `json:"rooms"`
	PersonalData  PersonalData     `json:"personalData"`
	Total         float64          `json:"total"`
}

// RoomCalculation represents the room details in the response
type RoomCalculation struct {
	Name          string  `json:"name"`
	PricePerHour  float64 `json:"pricePerHour"`
	ImageURL      string  `json:"imageURL"`
	Capacity      int     `json:"capacity"`
	Type          string  `json:"type"`
	SubTotalSnack float64 `json:"subTotalSnack"`
	SubTotalRoom  float64 `json:"subTotalRoom"`
	Snack         SnackCategory `json:"snack"`
}

// PersonalData represents the personal details of the reservation
type PersonalData struct {
	Name        string    `json:"name"`
	PhoneNumber string    `json:"phoneNumber"`
	Company     string    `json:"company"`
	StartTime   time.Time `json:"startTime"`
	EndTime     time.Time `json:"endTime"`
	Duration    int       `json:"duration"`
	Participant int       `json:"participant"`
}

// SnackCategory represents the snack details in the response
type SnackCategory struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	Unit     string  `json:"unit"`
	Price    float64 `json:"price"`
	Category string  `json:"category"`
}

type ReservationRequest struct {
	UserID     int    `json:"userID"`
	Name       string `json:"name"`
	PhoneNumber string `json:"phoneNumber"`
	Company    string `json:"company"`
	Notes      string `json:"notes"`
	Rooms      []RoomRequest `json:"rooms"`
}

type RoomRequest struct {
	ID          int       `json:"id"`
	StartTime   time.Time `json:"startTime"`
	EndTime     time.Time `json:"endTime"`
	Participant int       `json:"participant"`
	SnackID     int       `json:"snackID"`
}
type SuccessResponseCreateReservation struct {
	Message string `json:"message"`
}

type UpdateReservationStatusRequest struct {
	Status        string `json:"status"` // e.g., "confirmed", "cancelled", etc.	
}

type SuccessResponseUpdateReservationStatus struct {
	Message string `json:"message"`	
}