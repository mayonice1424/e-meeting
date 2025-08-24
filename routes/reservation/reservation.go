package router

import (
	reservationController "emeeting/controllers/reservation"
	auth "emeeting/middleware"

	"github.com/labstack/echo/v4"
)

func ReservationRoutes(e *echo.Echo) {
	group := e.Group("/api/v1")
	group.Use(auth.AuthMiddleware)
	group.GET("/reservation/calculation", reservationController.ReservationCalculation)
	group.POST("/reservation", reservationController.CreateReservation)
	group.PUT("/reservation/status/:id", reservationController.UpdateReservationStatus)
	group.GET("/reservation/:id", reservationController.GetReservationById)
	// group.POST("/reservations", reservationController.)
}
