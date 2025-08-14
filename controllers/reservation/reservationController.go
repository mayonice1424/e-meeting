package reservation

import (
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
					Name:       room.Name,
					PricePerHour: float64(room.PricePerHour),
					ImageURL:   room.Picture,
					Capacity:   room.Capacity,
					Type:       room.Type,
					SubTotalSnack: subTotalSnack,
					SubTotalRoom: subTotalRoom,
					Snack:      snack,
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