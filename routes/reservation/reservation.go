package router

import (
	auth "emeeting/middleware"

	"github.com/labstack/echo/v4"
)

func ReservationRoutes(e *echo.Echo){
	group := e.Group("/api/v1")
	group.Use(auth.AuthMiddleware)

	// group.POST("/reservations", reservationController.)
}