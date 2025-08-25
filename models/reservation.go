package models

import (
	"time"
)
type SuccessResponseReservationCalculation struct {
	Message string                         `json:"message"`
	Data    ReservationCalculationData     `json:"data"`
}

type PersonalReservation struct {
    ID                  int       `json:"id"`
    Name                string    `json:"name"`
    PhoneNumber         string    `json:"phoneNumber"`
    Company             string    `json:"company"`
    ReservationStatus   string    `json:"reservationStatus"`
    Notes               string    `json:"notes"`
    TotalInvoice        float64   `json:"totalInvoice"`
    StartTime           time.Time `json:"start_time"`
    EndTime             time.Time `json:"end_time"`
    Duration            int       `json:"duration"`  // Durasi dalam jam
    Participant         int       `json:"participant"`
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

type RoomCalculationReservation struct {
    Name          string    `json:"name"`
    PricePerHour  float64   `json:"pricePerHour"`
    ImageURL      string    `json:"imageURL"`
    Capacity      int       `json:"capacity"`
    Type          string    `json:"type"`
    SubTotalSnack float64   `json:"subTotalSnack"`
    SubTotalRoom  float64   `json:"subTotalRoom"`
    Snack         SnackCategory `json:"snack"`
    StartTime     time.Time `json:"start_time"`
    EndTime       time.Time `json:"end_time"`
    Duration      int       `json:"duration"`   
    Participant   int       `json:"participant"`
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

type ReservationCalculationReservation struct {
	Rooms         []RoomCalculationReservation `json:"rooms"`  // Ubah ke RoomCalculationReservation
	PersonalData  PersonalData                 `json:"personalData"`
	Total         float64                      `json:"total"`
}
type ReservationSchedule struct {
	ID          int                `json:"id"`
	RoomName    string             `json:"roomName"`
	CompanyName string             `json:"companyName"`
	Schedules   []ScheduleDetail   `json:"schedule"` // Nested schedules array
}

type ScheduleDetail struct {
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
	Status    string `json:"status"` 
}

type SuccessResponseReservationSchedule struct {
	Message   string                 `json:"message"`
	Reservations []ReservationSchedule `json:"reservations"`
	TotalData int                    `json:"totalData"`
	Page      int                    `json:"page"`
	PageSize  int                    `json:"pageSize"`
	TotalPage int                    `json:"totalPage"`
}

type RoomsReservationSchedule struct {
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
	Status    string `json:"status"` 
}

type SuccessResponseRoomsReservationSchedule struct {
	RoomName    string                 `json:"roomName"`
	Schedule    []RoomsReservationSchedule `json:"schedule"`
	TotalBooked int                    `json:"totalBooked"`
}

type RoomDashboard struct {
	ID                   int     `json:"id"`
	Name                 string  `json:"name"`
	Omzet                float64 `json:"omzet"`
	PercentageOfUsage    float64 `json:"percentageOfUsage"`
}

type DashboardData struct {
	TotalRoom      int             `json:"totalRoom"`
	TotalVisitor   int             `json:"totalVisitor"`
	TotalReservation int           `json:"totalReservation"`
	TotalOmzet      float64        `json:"totalOmzet"`
	Rooms           []RoomDashboard `json:"rooms"`
}

type SuccessResponseDashboard struct {
	Message string       `json:"message"`
	Data    DashboardData `json:"data"`
}

