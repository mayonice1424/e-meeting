package router

import (
	roomController "emeeting/controllers/room"
	auth "emeeting/middleware"

	"github.com/labstack/echo/v4"
)

func RoomRoutes(e *echo.Echo) {
	group := e.Group("/api/v1")
	group.Use(auth.AuthMiddleware)
	group.POST("/rooms", roomController.CreateRoom, auth.AuthAdminRoleMiddleware)
	group.GET("/rooms", roomController.GetRooms)
	group.PUT("/rooms/:id", roomController.UpdateRoom, auth.AuthAdminRoleMiddleware)
	group.DELETE("/rooms/:id", roomController.DeleteRoom, auth.AuthAdminRoleMiddleware)


	// e.GET("/rooms", roomController.GetRooms)
	// e.GET("/rooms/:id", roomController.GetRoomByID)
	// e.POST("/rooms", roomController.CreateRoom)
	// e.DELETE("/rooms/:id", roomController.DeleteRoom)
}
