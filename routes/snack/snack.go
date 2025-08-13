package routes

import (
	snackController "emeeting/controllers/snack"
	auth "emeeting/middleware"
	"github.com/labstack/echo/v4"
)
func SnackRoutes(e *echo.Echo) {
	group := e.Group("/api/v1")
	group.Use(auth.AuthMiddleware)
	group.GET("/snack", snackController.GetSnacks)
}