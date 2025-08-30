package reservation

import (
	"database/sql"
	configDb "emeeting/config"
	"emeeting/models"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
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
	if startTimeParsed.After(endTimeParsed) {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "End time cannot be earlier than start time"})
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
	if participant > room.Capacity {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Participant count exceeds room capacity"})
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

	if currentStatus == "Cancelled" || currentStatus == "Paid" {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "reservation already cancelled/paid"})
	}

	updateQuery := "UPDATE data_personal_reservation SET reservation_status = $1 WHERE id = $2"
	_, err = db.Exec(updateQuery, request.Status, reservationID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "Internal server error"})
	}
	return c.JSON(http.StatusOK, models.SuccessResponseUpdateReservationStatus{
		Message: "Update status successfully",
	})
}

// GetReservationSchedule godoc
// @Summary Get reservation schedule
// @Description Retrieve the reservation schedule within a specified date range with pagination
// @Tags reservation
// @Accept json
// @Produce json
// @Param start_date query string true "Start date (YYYY-MM-DD)"
// @Param end_date query string true "End date (YYYY-MM-DD)"
// @Param page query int false "Page number (default is 1)"
// @Param Authorization header string true "Bearer <JWT Token>"
// @Success 200 {object} models.SuccessResponseReservationSchedule
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/reservations/schedules [get]
func GetReservationSchedule(c echo.Context) error {
	startDateStr := c.QueryParam("start_date")
	endDateStr := c.QueryParam("end_date")
	pageStr := c.QueryParam("page")

	layout := "2006-01-02"
	startDate, err := time.Parse(layout, startDateStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Invalid start_date format, expected YYYY-MM-DD"})
	}
	endDate, err := time.Parse(layout, endDateStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Invalid end_date format, expected YYYY-MM-DD"})
	}

	page := 1
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	pageSize := 10
	offset := (page - 1) * pageSize

	db := configDb.ConnectToDatabase()
	defer db.Close()

	var totalData int
	countQuery := `
		SELECT COUNT(r.id) 
		FROM data_personal_reservation r
		LEFT JOIN data_booking_room s ON r.id = s.reservation_id
		WHERE s.start_date::date >= $1 AND s.end_date::date <= $2
	`
	err = db.QueryRow(countQuery, startDate, endDate).Scan(&totalData)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "Internal server error"})
	}

	query := `
		SELECT r.id, m.name, r.company_name, s.start_date, s.end_date, r.reservation_status
		FROM data_personal_reservation r
		LEFT JOIN data_booking_room s ON r.id = s.reservation_id
		LEFT JOIN room m ON m.id = s.id_room
		WHERE s.start_date::date >= $1 AND s.end_date::date <= $2
		ORDER BY s.start_date
		LIMIT $3 OFFSET $4
	`
	rows, err := db.Query(query, startDate, endDate, pageSize, offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "Internal server error"})
	}
	defer rows.Close()

	var schedules []models.ReservationSchedule
	for rows.Next() {
		var schedule models.ReservationSchedule
		var scheduleDetail models.ScheduleDetail
		err := rows.Scan(&schedule.ID, &schedule.RoomName, &schedule.CompanyName, &scheduleDetail.StartTime, &scheduleDetail.EndTime, &scheduleDetail.Status)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "Error scanning reservation"})
		}

		var startParsed, endParsed time.Time
		startParsed, err = time.Parse("2006-01-02 15:04:05", scheduleDetail.StartTime)
		if err != nil {
			startParsed, err = time.Parse(time.RFC3339, scheduleDetail.StartTime)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "Error parsing reservation start time"})
			}
		}
		endParsed, err = time.Parse("2006-01-02 15:04:05", scheduleDetail.EndTime)
		if err != nil {
			endParsed, err = time.Parse(time.RFC3339, scheduleDetail.EndTime)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "Error parsing reservation end time"})
			}
		}

		if endParsed.Before(time.Now()) {
			scheduleDetail.Status = "Done"
		} else if startParsed.Before(time.Now()) && endParsed.After(time.Now()) {
			scheduleDetail.Status = "In Progress"
		} else {
			scheduleDetail.Status = "Up Coming"
		}

		schedule.Schedules = append(schedule.Schedules, scheduleDetail)
		schedules = append(schedules, schedule)
	}

	response := models.SuccessResponseReservationSchedule{
		Message:      "Reservation schedule fetched successfully",
		Reservations: schedules,
		Page:         page,
		PageSize:     pageSize,
		TotalPage:    (totalData + pageSize - 1) / pageSize,
		TotalData:    totalData,
	}
	return c.JSON(http.StatusOK, response)
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

// GetDashboard godoc
// @Summary Get dashboard data for reservations
// @Description Retrieve dashboard data including total rooms, total visitors, total reservations, and total omzet within a specified date range.
// @Tags dashboard
// @Accept json
// @Produce json
// @Param startDate query string true "Start date (YYYY-MM-DD)"
// @Param endDate query string true "End date (YYYY-MM-DD)"
// @Param Authorization header string true "Bearer <JWT Token>"
// @Success 200 {object} models.SuccessResponseDashboard
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/dashboard [get]
func GetDashboard(c echo.Context) error {
	startDateStr := c.QueryParam("startDate")
	endDateStr := c.QueryParam("endDate")

	layout := "2006-01-02"
	startDate, err := time.Parse(layout, startDateStr)
	if err != nil {
		fmt.Print("Error parsing start date: ", err)
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Invalid start date format, expected YYYY-MM-DD"})
	}

	endDate, err := time.Parse(layout, endDateStr)
	if err != nil {
		fmt.Print("Error parsing end date: ", err)
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Invalid end date format, expected YYYY-MM-DD"})
	}

	if startDate.After(endDate) {
		fmt.Print("Invalid date range: ", startDate, " - ", endDate)
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Start date must be smaller than end date"})
	}

	db := configDb.ConnectToDatabase()
	defer db.Close()

	var totalRooms int
	roomCountQuery := "SELECT COUNT(*) FROM room"
	err = db.QueryRow(roomCountQuery).Scan(&totalRooms)
	if err != nil {
		fmt.Print("Error counting rooms: ", err)
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "Error counting rooms"})
	}

	var totalVisitors int
	visitorCountQuery := `
		SELECT SUM(b.total_participant)
		FROM data_personal_reservation r
		LEFT JOIN data_booking_room b ON r.id = b.reservation_id
		WHERE r.reservation_status = 'Paid' AND b.start_date::date >= $1 AND b.end_date::date <= $2
	`
	err = db.QueryRow(visitorCountQuery, startDate, endDate).Scan(&totalVisitors)
	if err != nil {
		fmt.Print("Error counting visitors: ", err)
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "No Visitor Found in reservation status Paid"})
	}

	var totalReservations int
	reservationCountQuery := `
		SELECT COUNT(*)
		FROM data_personal_reservation r
		LEFT JOIN data_booking_room b ON r.id = b.reservation_id
		WHERE r.reservation_status = 'Paid' AND b.start_date::date >= $1 AND b.end_date::date <= $2
	`
	err = db.QueryRow(reservationCountQuery, startDate, endDate).Scan(&totalReservations)
	if err != nil {
		fmt.Print("Error counting reservations: ", err)
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "Error counting reservations"})
	}

	var totalOmzet float64
	omzetQuery := `
		SELECT SUM(b.room_price * b.duration)
		FROM data_personal_reservation r
		LEFT JOIN data_booking_room b ON r.id = b.reservation_id
		WHERE r.reservation_status = 'Paid' AND b.start_date::date >= $1 AND b.end_date::date <= $2
	`
	err = db.QueryRow(omzetQuery, startDate, endDate).Scan(&totalOmzet)
	if err != nil {
		fmt.Print("Error calculating omzet: ", err)
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "Error calculating omzet"})
	}

	roomQuery := `
		SELECT m.id, m.name, COALESCE(SUM(b.room_price * b.duration), 0) as omzet
		FROM room m
		LEFT JOIN data_booking_room b ON m.id = b.id_room AND b.reservation_id IN
			(SELECT id FROM data_personal_reservation WHERE reservation_status = 'Paid')
		GROUP BY m.id
		ORDER BY m.id ASC
	`
	rows, err := db.Query(roomQuery)
	if err != nil {
		fmt.Print("Error fetching room data: ", err)
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "Error fetching room data"})
	}
	defer rows.Close()

	var rooms []models.RoomDashboard
	for rows.Next() {
		var room models.RoomDashboard
		var omzet float64
		err := rows.Scan(&room.ID, &room.Name, &omzet)
		if err != nil {
			fmt.Print("Error scanning room data: ", err)
			return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "Error scanning room data"})
		}

		room.Omzet = omzet
		room.PercentageOfUsage = math.Round((omzet/totalOmzet)*100*100) / 100

		rooms = append(rooms, room)
	}

	response := models.SuccessResponseDashboard{
		Message: "Get dashboard data success",
		Data: models.DashboardData{
			TotalRoom:        totalRooms,
			TotalVisitor:     totalVisitors,
			TotalReservation: totalReservations,
			TotalOmzet:       totalOmzet,
			Rooms:            rooms,
		},
	}
	return c.JSON(http.StatusOK, response)
}



// GetHistory godoc
// @Summary Get reservation history
// @Description Retrieve the reservation history within a specified date range with pagination
// @Tags reservation
// @Accept json
// @Produce json
// @Param start_date query string false "Start date (YYYY-MM-DD)"
// @Param end_date query string false "End date (YYYY-MM-DD)"
// @Param type query string false "Reservation type"
// @Param status query string false "Reservation status"
// @Param page query int false "Page number (default is 1)"
// @Param Authorization header string true "Bearer <JWT Token>"
// @Success 200 {object} models.SuccessResponseHistory
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/reservation/history [get]
func GetHistory(c echo.Context) error {
    fmt.Println("Fetching reservation history")
    claimsInterface := c.Get("userClaims")
    if claimsInterface == nil {
        return c.JSON(http.StatusUnauthorized, models.ErrorResponse{Message: "Unauthorized"})
    }

	claims, ok := claimsInterface.(jwt.MapClaims)
	if !ok {
		return c.JSON(http.StatusUnauthorized, models.ErrorResponse{Message: "Invalid claims"})
	}
	role := claims["role"].(string)
	var userId string
	if id, ok := claims["userId"].(float64); ok {
		userId = fmt.Sprintf("%.0f", id) 
	} else if id, ok := claims["userId"].(string); ok {
		userId = id
	} else {
		return c.JSON(http.StatusUnauthorized, models.ErrorResponse{Message: "Invalid userId format"})
	}
	_ = userId

    startDateStr := c.QueryParam("start_date")
    endDateStr := c.QueryParam("end_date")
    pageStr := c.QueryParam("page")
    typeStr := c.QueryParam("type")
    statusStr := c.QueryParam("status")
    fmt.Println("Query Parameters - start_date:", startDateStr, "end_date:", endDateStr, "page:", pageStr, "type:", typeStr, "status:", statusStr)

    layout := "2006-01-02"
    var startDate time.Time
    var endDate time.Time

    if startDateStr == "" {
        startDate = time.Now()
    } else {
        var err error
        startDate, err = time.Parse(layout, startDateStr)
        if err != nil {
            return c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Invalid startDate format"})
        }
    }

    if endDateStr == "" {
        endDate = time.Now()
    } else {
        var err error
        endDate, err = time.Parse(layout, endDateStr)
        if err != nil {
            return c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Invalid endDate format"})
        }
    }

    startDateStrFormatted := startDate.Format("2006-01-02")
    endDateStrFormatted := endDate.Format("2006-01-02")

    page := 1
    if pageStr != "" {
        if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
            page = p
        }
    }
    pageSize := 10
    offset := (page - 1) * pageSize

    db := configDb.ConnectToDatabase()
    defer db.Close()

	var totalData int
	var query string
	var rows *sql.Rows

	if role == "Admin" {
		countQuery := `
			SELECT COUNT(dpr.id)
			FROM data_booking_room dbr
			JOIN data_personal_reservation dpr ON dbr.reservation_id = dpr.id
			LEFT JOIN room r ON r.id = dbr.id_room
			WHERE 1 = 1
		`

		args := []interface{}{}

		if startDateStr != "" {
			countQuery += " AND dbr.start_date::date >= $1"
			args = append(args, startDateStrFormatted)
		}
		if endDateStr != "" {
			countQuery += " AND dbr.end_date::date <= $2"
			args = append(args, endDateStrFormatted)
		}

		if typeStr != "" {
			countQuery += " AND r.type = $3"
			args = append(args, typeStr)
		}

		if statusStr != "" {
			countQuery += " AND dpr.reservation_status = $4"
			args = append(args, statusStr)
		}

		err := db.QueryRow(countQuery, args...).Scan(&totalData)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "Error fetching total data"})
		}

		query = `
			SELECT dpr.id, dpr.name, dpr.no_hp, dpr.company_name, dpr.total_invoice, dpr.reservation_status, dpr.recervation_create_date, dpr.reservation_update_date
			FROM data_personal_reservation dpr
			JOIN data_booking_room dbr ON dbr.reservation_id = dpr.id
			LEFT JOIN room r ON r.id = dbr.id_room
			WHERE 1 = 1
		`

		args = []interface{}{}
		if startDateStr != "" {
			query += " AND dbr.start_date::date >= $1"
			args = append(args, startDateStrFormatted)
		}
		if endDateStr != "" {
			query += " AND dbr.end_date::date <= $2"
			args = append(args, endDateStrFormatted)
		}

		if typeStr != "" {
			query += " AND r.type = $3"
			args = append(args, typeStr)
		}

		if statusStr != "" {
			query += " AND dpr.reservation_status = $4"
			args = append(args, statusStr)
		}

		limitPos := len(args) + 1
		offsetPos := len(args) + 2

		query += fmt.Sprintf(`
			ORDER BY dpr.recervation_create_date
			LIMIT $%d OFFSET $%d
		`, limitPos, offsetPos)

		args = append(args, pageSize, offset)

		rows, err = db.Query(query, args...)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "Error fetching reservation data"})
		}
	} else {
		// Similar logic for non-Admin role
	}

    defer rows.Close()

    var reservations []models.PersonalReservationHistory
    for rows.Next() {
        var reservation models.PersonalReservationHistory
        err := rows.Scan(
            &reservation.ID,
            &reservation.Name,
            &reservation.PhoneNumber,
            &reservation.Company,
            &reservation.TotalInvoice,
            &reservation.ReservationStatus,
            &reservation.CreatedAt,
            &reservation.UpdatedAt,
        )
        if err != nil {
            return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "Error scanning reservation"})
        }

        var rooms []models.RoomDetail
        roomQuery := `
            SELECT r.id, r.price_per_hour, r.name, r.type, r.capacity, dbr.sub_total_room_price, dbr.sub_total_snack_price, s.id AS snack_id, s.name AS snack_name, s.unit, s.price, s.category
            FROM room r
            JOIN data_booking_room dbr ON r.id = dbr.id_room
            LEFT JOIN snack_category s ON s.id = dbr.snack_id
            WHERE dbr.reservation_id = $1
        `
        roomRows, err := db.Query(roomQuery, reservation.ID)
        if err != nil {
            return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "Error fetching room data"})
        }
        defer roomRows.Close()

        for roomRows.Next() {
            var room models.RoomDetail
            var snack models.SnackCategory
            err := roomRows.Scan(&room.ID, &room.Price, &room.Name, &room.Type, &room.Capacity, &room.SubTotalRoom, &room.SubTotalSnack, &snack.ID, &snack.Name, &snack.Unit, &snack.Price, &snack.Category)
            if err != nil {
                return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "Error scanning room and snack data"})
            }
            room.Snack = snack
            rooms = append(rooms, room)
        }
        reservation.Rooms = rooms
        reservations = append(reservations, reservation)
    }

    response := models.SuccessResponseHistory{
        Message:  "Reservation history fetched successfully",
        Data:     reservations,
        Page:     page,
        PageSize: pageSize,
        TotalPage: (totalData + pageSize - 1) / pageSize,
        TotalData: totalData,
    }

    return c.JSON(http.StatusOK, response)
}
