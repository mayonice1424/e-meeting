package routes

import (
	userController "emeeting/controllers/user"

	"github.com/labstack/echo/v4"
)

func RegisterRoutes(e *echo.Echo) {
	e.POST("/users/register", userController.UserRegister)
}