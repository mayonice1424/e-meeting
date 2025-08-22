package reservation

import (
	"database/sql"
	configDb "emeeting/config"
	"emeeting/models"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
)

// ReservationCalculation godoc
// @Summary Endpoint for calculating reservation
// @Description Calculate the total cost for a room reservation based on room, snack, participant count, and time
// @Tags reservation
// @Accept json
// @Produce json
// @Param room_id query int true "Room ID"
// @Param snack_id query int false "Snack ID"
// @Param startTime query string true "Start time (YYYY-MM-DD HH:MM:SS)"
// @Param endTime query string true "End time (YYYY-MM-DD HH:MM:SS)"
// @Param participant query int true "Number of participants"
// @Param user_id query int true "User ID"
// @Param name query string true "User name"
// @Param phoneNumber query string true "Phone number"
// @Param company query string true "Company name"
// @Param Authorization header string true "Bearer <JWT Token>"
// @Success 200 {object} models.SuccessResponseReservationCalculation
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/reservation/calculation [get]
func ReservationCalculation(c echo.Context) error {
	fmt.Println("ReservationCalculation called")
	roomID, err := strconv.Atoi(c.QueryParam("room_id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Invalid room ID"})
	}
	snackID, err := strconv.Atoi(c.QueryParam("snack_id"))
	if err != nil {
		snackID = 0 // If no snack is provided, set to 0 (no snack)
	}
	startTime := c.QueryParam("startTime")
	endTime := c.QueryParam("endTime")
	participant, err := strconv.Atoi(c.QueryParam("participant"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Invalid participant count"})
	}
	_, err = strconv.Atoi(c.QueryParam("user_id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Invalid user ID"})
	}
	name := c.QueryParam("name")
	phoneNumber := c.QueryParam("phoneNumber")
	company := c.QueryParam("company")
	if startTime == "" || endTime == "" {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Start and end time are required"})
	}
	startTimeParsed, err := time.Parse("2006-01-02 15:04:05", startTime)
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Invalid start time format"})
	}
	endTimeParsed, err := time.Parse("2006-01-02 15:04:05", endTime)
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Invalid end time format"})
	}
	duration := endTimeParsed.Sub(startTimeParsed).Hours()
	db := configDb.ConnectToDatabase()
	defer db.Close()

	var room models.RoomById
	queryRoom := "SELECT id, name, price_per_hour, capacity, picture FROM room WHERE id = $1"
	err = db.QueryRow(queryRoom, roomID).Scan(&room.ID, &room.Name, &room.PricePerHour, &room.Capacity, &room.Picture)
	if err != nil {
		return c.JSON(http.StatusNotFound, models.ErrorResponse{Message: "Room not found"})
	}

	var snack models.SnackCategory
	var subTotalSnack float64
	if snackID != 0 {
		querySnack := "SELECT id, name, price FROM snack_category WHERE id = $1"
		err = db.QueryRow(querySnack, snackID).Scan(&snack.ID, &snack.Name, &snack.Price)
		if err != nil {
			return c.JSON(http.StatusNotFound, models.ErrorResponse{Message: "Snack not found"})
		}
		subTotalSnack = snack.Price * float64(participant) // Calculate snack total for participants
	}

	subTotalRoom := float64(room.PricePerHour) * duration

	total := subTotalRoom + subTotalSnack

	response := models.SuccessResponseReservationCalculation{
		Message: "Reservation calculation successful",
		Data: models.ReservationCalculationData{
			Rooms: []models.RoomCalculation{
				{
					Name:          room.Name,
					PricePerHour:  float64(room.PricePerHour),
					ImageURL:      room.Picture,
					Capacity:      room.Capacity,
					Type:          room.Type,
					SubTotalSnack: subTotalSnack,
					SubTotalRoom:  subTotalRoom,
					Snack:         snack,
				},
			},
			PersonalData: models.PersonalData{
				Name:        name,
				PhoneNumber: phoneNumber,
				Company:     company,
				StartTime:   startTimeParsed,
				EndTime:     endTimeParsed,
				Duration:    int(duration),
				Participant: participant,
			},
			Total: total,
		},
	}

	return c.JSON(http.StatusOK, response)
}

// CreateReservation godoc
// @Summary Create a new reservation
// @Description Create a reservation with room details
// @Tags reservation
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer <JWT Token>"
// @Param reservation body models.ReservationRequest true "Reservation Information"
// @Success 201 {object} models.SuccessResponseCreateReservation
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/reservation [post]
func CreateReservation(c echo.Context) error {
	var reservation models.ReservationRequest
	if err := c.Bind(&reservation); err != nil {
		fmt.Println("Error binding reservation request:", err)
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Invalid request body"})
	}
	db := configDb.ConnectToDatabase()
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "Database error"})
	}

	var total_invoice float64
	var roomPrice float64
	var snackPrice float64

	for _, room := range reservation.Rooms {
		var existingBookingStatus string
		roomBookingQuery := `
		SELECT r.reservation_status 
		FROM data_booking_room b
		JOIN data_personal_reservation r
    ON b.reservation_id = r.id
		WHERE b.id_room = $1
    AND (b.start_date, b.end_date) OVERLAPS ($2, $3)  -- Pastikan tipe data sudah sesuai
    AND r.reservation_status IN ('Booked', 'Paid')`
		err := tx.QueryRow(roomBookingQuery, room.ID, room.StartTime, room.EndTime).Scan(&existingBookingStatus)
		if err == nil {
			tx.Rollback()
			return c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: fmt.Sprintf("Room %d is already booked or paid for the requested time", room.ID)})
		} else if err != sql.ErrNoRows {
			tx.Rollback()
			return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "Failed to check room availability"})
		}

		if room.StartTime.After(room.EndTime) {
			tx.Rollback()
			return c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "End time cannot be earlier than start time"})
		}
		var roomCapacity int
		roomQuery := `SELECT capacity FROM room WHERE id = $1`
		err = tx.QueryRow(roomQuery, room.ID).Scan(&roomCapacity)
		if err != nil {
			if err == sql.ErrNoRows {
				fmt.Println("Room not found:", room.ID)
				tx.Rollback()
				return c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: fmt.Sprintf("Room ID %d not found", room.ID)})
			}
			tx.Rollback()
			fmt.Println("Error fetching room capacity:", err)
			return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "Failed to fetch room capacity"})
		}

		if room.Participant > roomCapacity {
			tx.Rollback()
			return c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Message: fmt.Sprintf("The number of participants (%d) exceeds the room capacity (%d)", room.Participant, roomCapacity),
			})
		}

		err = tx.QueryRow(`SELECT price_per_hour FROM room WHERE id = $1`, room.ID).Scan(&roomPrice)
		if err != nil {
			tx.Rollback()
			fmt.Println("Error fetching room data:", err)
			return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "Failed to fetch room data"})
		}

		err = tx.QueryRow(`SELECT price FROM snack_category WHERE id = $1`, room.SnackID).Scan(&snackPrice)
		if err != nil {
			tx.Rollback()
			fmt.Println("Error fetching snack data:", err)
			return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "Failed to fetch snack data"})
		}

		duration := int(room.EndTime.Sub(room.StartTime).Hours())

		roomTotalPrice := roomPrice * float64(duration)

		snackTotalPrice := float64(room.Participant) * snackPrice

		total_invoice += roomTotalPrice + snackTotalPrice
	}

	reservationQuery := `
		INSERT INTO data_personal_reservation (id_user, name, no_hp, company_name, reservation_status, note, total_invoice)
		VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`
	var reservationID int
	err = tx.QueryRow(reservationQuery, reservation.UserID, reservation.Name, reservation.PhoneNumber, reservation.Company, "Booked", reservation.Notes, total_invoice).Scan(&reservationID)
	if err != nil {
		tx.Rollback()
		fmt.Println("Error creating reservation:", err)
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "Failed to create reservation"})
	}

	for _, room := range reservation.Rooms {
		duration := int(room.EndTime.Sub(room.StartTime).Hours())

		roomQuery := `
			INSERT INTO data_booking_room (id_room, reservation_id, snack_id, start_date, end_date, total_participant, duration, room_price, total_snack, sub_total_invoice, snack_price, sub_total_room_price, sub_total_snack_price)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`

		_, err := tx.Exec(roomQuery, room.ID, reservationID, room.SnackID, room.StartTime, room.EndTime, room.Participant, duration, roomPrice, float64(room.Participant)*snackPrice, total_invoice, snackPrice, roomPrice*float64(duration), snackPrice*float64(room.Participant))
		if err != nil {
			tx.Rollback()
			return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "Failed to book room"})
		}
	}

	err = tx.Commit()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "Failed to commit reservation"})
	}

	return c.JSON(http.StatusCreated, models.SuccessResponseCreateReservation{
		Message: "Reservation created successfully",
	})
}

// UpdateReservationStatus godoc
// @Summary Update reservation status
// @Description Update the status of a reservation
// @Tags reservation
// @Accept json
// @Produce json
// @Param id path int true "Reservation ID"
// @Param Authorization header string true "Bearer <JWT Token>"
// @Param reservation body models.UpdateReservationStatusRequest true "New reservation status"
// @Success 200 {object} models.SuccessResponseUpdateReservationStatus
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/reservation/status/{id} [put]
func UpdateReservationStatus(c echo.Context) error {
	reservationID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Invalid reservation ID"})
	}

	var request models.UpdateReservationStatusRequest
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Invalid request body"})
	}

	db := configDb.ConnectToDatabase()
	defer db.Close()

	// Get current status
	var currentStatus string
	query := "SELECT reservation_status FROM data_personal_reservation WHERE id = $1"
	err = db.QueryRow(query, reservationID).Scan(&currentStatus)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, models.ErrorResponse{Message: "url not found"})
		}
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "Internal server error"})
	}

	// Check if already canceled or paid
	if currentStatus == "Cancelled" || currentStatus == "Paid" {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "reservation already cancelled/paid"})
	}

	// Update Status
	updateQuery := "UPDATE data_personal_reservation SET reservation_status = $1 WHERE id = $2"
	_, err = db.Exec(updateQuery, request.Status, reservationID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "Internal server error"})
	}
	return c.JSON(http.StatusOK, models.SuccessResponseUpdateReservationStatus{
		Message: "Update status successfully",
	})
}
