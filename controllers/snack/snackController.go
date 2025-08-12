package snackController

import (
	configDb "emeeting/config"
	"emeeting/models"
	"fmt"
	"net/http"
	"github.com/labstack/echo/v4"
)

// GetSnacks godoc
// @Summary Endpoint for snack by id
// @Description Snack
// @Tags snacks
// @Accept json
// @Produce json
// @Success 200 {object} models.SuccessResponseSnack
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /snack [get]
func GetSnacks(c echo.Context) error {
	fmt.Println("Claims: test0")
	db:= configDb.ConnectToDatabase()
	defer db.Close()
	var snacks []models.Snack
	query := "SELECT id, name, price, unit, category FROM snack_category"
	rows, err := db.Query(query)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "failed to retrieve snacks: " + err.Error()})
	}

	defer rows.Close()
	fmt.Println("Claims: test5", rows)

	for rows.Next() {
		var snack models.Snack
		if err := rows.Scan(&snack.ID, &snack.Name, &snack.Price, &snack.Unit, &snack.Category); err != nil {
			return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "failed to scan snack: " + err.Error()})
		}
		snacks = append(snacks, snack)
	}

	if err := rows.Err(); err != nil {
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "error iterating over snacks: " + err.Error()})
	}
	if len(snacks) == 0 {
		return c.JSON(http.StatusNotFound, models.ErrorResponse{Message: "no snacks found"})
	}
	return c.JSON(http.StatusOK, models.SuccessResponseSnack{Data: snacks, Message: "snacks retrieved successfully"})
}

