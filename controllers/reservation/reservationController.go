package reservationController

import (
	configDb "emeeting/config"
	"emeeting/models"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

// GetReservations godoc
// @Summary Endpoint Get all reservations a reservation
// @Description Get all reservations with name, phoneNumber, company, notes, room
// @Tags reservations
// @Accept json
// @Produce json
// @param reservation body models.ReservationRequest true "Reservation Request"
// @Param Authorization header string true "Bearer <JWT Token>"
// @Success 201 {object} models.SuccessResponseReservations
// @Failure 400 {object} models.ErrorResponse
// @Router /api/v1/reservations [post]
func CreateReservation(c echo.Context) error {
	db := configDb.ConnectToDatabase()
	defer db.Close()
	var reservation models.ReservationRequest
	if err := c.Bind(&reservation); err != nil {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Invalid request body"})
	}

	// Start transaction
	tx, err := db.Begin()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Message: "internal server error",
		})
	}

	// Insert ke tabel reservation
	query := `
	INSERT INTO reservation (name, phone_number, company, notes) 
	VALUES ($1, $2, $3, $4) 
	RETURNING id`
	var reservationID int
	err = tx.QueryRow(query, reservation.Name, reservation.PhoneNumber, reservation.Company, reservation.Notes).Scan(&reservationID)
	fmt.Println("reservationID:", reservationID)
	if err != nil {
		_ = tx.Rollback()
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "internal server error" + err.Error()})
	}

	// Insert data ke tabel reservation_rooms
	for _, room := range reservation.Rooms {
		if room.RoomID == 0 {
			_ = tx.Rollback()
			return c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Message: "room not found / room has been booked",
			})
		}

		_, err := tx.Exec(`
		INSERT INTO reservation_rooms (reservation_id, room_id, start_time, end_time, participant, snack_id) VALUES ($1, $2, $3, $4, $5, $6)
	`, reservationID, room.RoomID, room.StartTime, room.EndTime, room.Participant, room.SnackID)

		if err != nil {
			_ = tx.Rollback()
			return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "internal server error" + err.Error()})

		}
	}

	// Commit transaksi
	err = tx.Commit()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Message: "internal server error",
		})
	}

	// Berhasil
	return c.JSON(http.StatusCreated, models.SuccessResponse{Message: "Reservation successfully created"})

}
