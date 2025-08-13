package routes

import (
	userController "emeeting/controllers/user"
	auth "emeeting/middleware"

	"github.com/labstack/echo/v4"
)

func RegisterRoutes(e *echo.Echo) {
	e.POST("/users/register", userController.UserRegister)
	e.POST("/users/login", userController.UserLogin)
	e.POST("/password/reset", userController.RequestResetPassword)
	e.PUT("/password/reset/:id", userController.ResetPassword)
	group := e.Group("/api/v1")
	group.Use(auth.AuthMiddleware)
	group.GET("/users/:id", userController.UserById)
	group.PUT("/users/:id", userController.UserUpdate)
}