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
		querySnack := "SELECT id, name, price, unit, category FROM snack_category WHERE id = $1"
		err = db.QueryRow(querySnack, snackID).Scan(&snack.ID, &snack.Name, &snack.Price, &snack.Unit, &snack.Category)
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
	roomIDMap := make(map[int]bool)
	for _, room := range reservation.Rooms {
		if roomIDMap[room.ID] {
			tx.Rollback()
			return c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Message: fmt.Sprintf("Room ID %d is duplicated in the request", room.ID),
			})
		}
		roomIDMap[room.ID] = true
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

// ReservationSchedule godoc
// @Summary Get reservation schedule by ID
// @Description Retrieve reservation schedule details by reservation ID
// @Tags reservation
// @Accept json
// @Produce json
// @Param id path int true "Reservation ID"
// @Param Authorization header string true "Bearer <JWT Token>"
// @Success 200 {object} models.SuccessResponseReservationSchedule
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/reservation/schedule/{id} [get]
func ReservationSchedule(c echo.Context) error {
	// ambil parameter dari path
	reservationID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Invalid reservation ID"})
	}

	db := configDb.ConnectToDatabase()
	defer db.Close()
	// Query: ambil jadwal reservasi berdasarkan reservationID
	query := `
		SELECT r.id, r.name, b.start_date, b.end_date, b.total_participant
		FROM data_booking_room b
		JOIN room r ON b.id_room = r.id
		WHERE b.reservation_id = $1
	`

	rows, err := db.Query(query, reservationID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Message: "Internal server error (query failed)",
		})
	}
	defer rows.Close()

	var schedules []models.ReservationSchedule
	// Loop untuk membaca hasil query dan masukkan ke slice schedules
	for rows.Next() {
		var schedule models.ReservationSchedule
		err := rows.Scan(
			&schedule.RoomID,
			&schedule.RoomName,
			&schedule.StartTime,
			&schedule.EndTime,
			&schedule.ParticipantCount,
		)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "Failed to scan row"})
		}
		schedules = append(schedules, schedule)
	}

	if err = rows.Err(); err != nil {
		// jika error saat iterasi hasil query, return 500
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "Internal server error"})
	}
	if len(schedules) == 0 {
		return c.JSON(http.StatusNotFound, models.ErrorResponse{Message: "No reservation found for the given ID"})
	}
	// Buat respone dalam format sesuai models
	return c.JSON(http.StatusOK, models.SuccessResponseReservationSchedule{Data: schedules, Message: "Reservation schedule retrieved successfully"})
}

// GetRoomsReservationSchedule godoc
// @Summary Get room reservation schedule
// @Description Retrieve the reservation schedule for a specific room on a given date
// @Tags reservation
// @Accept json
// @Produce json
// @Param id query int true "Room ID"
// @Param start_date query string true "Date (YYYY-MM-DD)"
// @Param Authorization header string true "Bearer <JWT Token>"
// @Success 200 {object} models.SuccessResponseRoomsReservationSchedule
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/rooms/{id}/reservation [get]
func GetRoomsReservationSchedule(c echo.Context) error {
	idRoom, err := strconv.Atoi(c.QueryParam("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Invalid room ID"})
	}
	startDate := c.QueryParam("start_date")

	if startDate == "" {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Start Date is required"})
	}

	db := configDb.ConnectToDatabase()
	defer db.Close()

	var roomName string
	err = db.QueryRow("SELECT name FROM room WHERE id = $1", idRoom).Scan(&roomName)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, models.ErrorResponse{Message: "Url not found"})
		} else if err != nil {
			fmt.Println("Error querying room name:", err)
			return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "Internal server error"})
		}
	}

	rows, err := db.Query(`
		SELECT dbr.start_date, dbr.end_date, dpr.reservation_status
		FROM data_booking_room dbr
		LEFT JOIN data_personal_reservation dpr ON dbr.reservation_id = dpr.id
		WHERE dbr.id_room = $1 AND DATE(dbr.start_date) = $2
		ORDER BY dbr.start_date ASC
	`, idRoom, startDate)
	if err != nil {
		fmt.Println("Error querying reservation schedule:", err)
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "Internal server error"})
	}
	defer rows.Close()
	var schedules []models.RoomsReservationSchedule
	for rows.Next() {
		var schedule models.RoomsReservationSchedule
		err := rows.Scan(&schedule.StartTime, &schedule.EndTime, &schedule.Status)
		if err != nil {
			fmt.Println("Error scanning reservation schedule:", err)	
			return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "Internal server error"})
		}
		schedules = append(schedules, schedule)
	}
	return c.JSON(http.StatusOK, models.SuccessResponseRoomsReservationSchedule{
		RoomName:    roomName,
		Schedule:    schedules,
		TotalBooked: len(schedules),
	})
}


// GetReservationById godoc
// @Summary Get reservation by ID
// @Description Get the details of a reservation by its ID
// @Tags reservation
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer <JWT Token>"
// @Param id path int true "Reservation ID"
// @Success 200 {object} models.ReservationCalculationReservation
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/reservation/{id} [get]
func GetReservationById(c echo.Context) error {
	reservationID := c.Param("id")
	if reservationID == "" {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Reservation ID is required"})
	}

	db := configDb.ConnectToDatabase()
	defer db.Close()

	var reservation models.PersonalReservation
	reservationQuery := `
		SELECT dpr.name, dpr.no_hp, dpr.company_name, dpr.reservation_status, dpr.note, dpr.total_invoice, dbr.duration, dbr.start_date, dbr.end_date, dbr.total_participant
		FROM data_personal_reservation dpr left join data_booking_room dbr on dpr.id = dbr.reservation_id
		WHERE dpr.id = $1
	`
	err := db.QueryRow(reservationQuery, reservationID).Scan(
		&reservation.Name, &reservation.PhoneNumber, &reservation.Company, &reservation.ReservationStatus, &reservation.Notes, &reservation.TotalInvoice, &reservation.Duration, &reservation.StartTime, &reservation.EndTime, &reservation.Participant,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, models.ErrorResponse{Message: "Reservation not found"})
		}
		fmt.Println("Error fetching reservation data:", err)
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "Error fetching reservation data"})
	}

	var rooms []models.RoomCalculationReservation
	roomQuery := `
		SELECT r.name, r.price_per_hour, r.picture, r.capacity, r.type, b.start_date, b.end_date, b.total_participant, b.snack_id
		FROM data_booking_room b
		JOIN room r ON b.id_room = r.id
		WHERE b.reservation_id = $1
	`
	rows, err := db.Query(roomQuery, reservationID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "Error fetching room data"})
	}
	defer rows.Close()

	for rows.Next() {
		var room models.RoomCalculationReservation
		var snackID int
		err := rows.Scan(
			&room.Name, &room.PricePerHour, &room.ImageURL, &room.Capacity, &room.Type, &room.StartTime, &room.EndTime, &room.Participant, &snackID,
		)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "Error reading room data"})
		}

		var snack models.SnackCategory
		if snackID != 0 {
			snackQuery := `
				SELECT id, name, price, unit, category
				FROM snack_category
				WHERE id = $1
			`
			err = db.QueryRow(snackQuery, snackID).Scan(&snack.ID, &snack.Name, &snack.Price, &snack.Unit, &snack.Category)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "Error fetching snack data"})
			}
		}

		duration := room.EndTime.Sub(room.StartTime).Hours()
		subTotalRoom := room.PricePerHour * duration
		subTotalSnack := float64(room.Participant) * snack.Price

		room.SubTotalRoom = subTotalRoom
		room.SubTotalSnack = subTotalSnack
		room.Snack = snack

		rooms = append(rooms, room)
	}

	var total float64
	for _, room := range rooms {
		total += room.SubTotalRoom + room.SubTotalSnack
	}

	var roomCalcs []models.RoomCalculation
	for _, rr := range rooms {
		roomCalcs = append(roomCalcs, models.RoomCalculation{
			Name:          rr.Name,
			PricePerHour:  rr.PricePerHour,
			ImageURL:      rr.ImageURL,
			Capacity:      rr.Capacity,
			Type:          rr.Type,
			SubTotalRoom:  rr.SubTotalRoom,
			SubTotalSnack: rr.SubTotalSnack,
			Snack:         rr.Snack,
		})
	}

	response := models.SuccessResponseReservationCalculation{
		Message: "Reservation data fetched successfully",
		Data: models.ReservationCalculationData{
			Rooms: roomCalcs,
			PersonalData: models.PersonalData{
				Name:        reservation.Name,
				PhoneNumber: reservation.PhoneNumber,
				Company:     reservation.Company,
				StartTime:   reservation.StartTime,
				EndTime:     reservation.EndTime,
				Duration:    int(reservation.Duration),
				Participant: reservation.Participant,
			},
			Total: total,
		},
	}

	return c.JSON(http.StatusOK, response)
}