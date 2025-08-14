package router

import (
	auth "emeeting/middleware"
	reservationController "emeeting/controllers/reservation"

	"github.com/labstack/echo/v4"
)

func ReservationRoutes(e *echo.Echo){
	group := e.Group("/api/v1")
	group.Use(auth.AuthMiddleware)
	group.GET("/reservation/calculation", reservationController.ReservationCalculation)
	// group.POST("/reservations", reservationController.)
}