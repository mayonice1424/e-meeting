package models

// Removed because it was unused

type Reservation_rooms struct {
	ReservationID int    `json:"reservation_id"`
	RoomID        int    `json:"room_id"`
	StartTime     string `json:"start_time"`
	EndTime       string `json:"end_time"`
	Participant   int    `json:"participant"`
	SnackID       int    `json:"snack_id"`
}

type ReservationRequest struct {
	Name        string              `json:"name"`
	PhoneNumber string              `json:"phone_number"`
	Company     string              `json:"company"`
	Notes       string              `json:"notes"`
	Rooms       []Reservation_rooms `json:"rooms"`
}

type SuccessResponseReservations struct {
	Message string `json:"message"`
}
