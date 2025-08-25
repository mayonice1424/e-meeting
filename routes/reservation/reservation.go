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
	group.GET("/reservations/schedules", reservationController.GetReservationSchedule)
	group.GET("/rooms/:id/reservation", reservationController.GetRoomsReservationSchedule)
	group.GET("/dashboard", reservationController.GetDashboard, auth.AuthAdminRoleMiddleware)
	// group.POST("/reservations", reservationController.)
}
