package router

import (
	roomController "emeeting/controllers/room"
	auth "emeeting/middleware"
	"github.com/labstack/echo/v4"
)

func RoomRoutes(e *echo.Echo) {
	group := e.Group("/api/v1")
	group.Use(auth.AuthMiddleware)
	group.POST("/rooms", roomController.CreateRoom)
	// e.GET("/rooms", roomController.GetRooms)
	// e.GET("/rooms/:id", roomController.GetRoomByID)
	// e.POST("/rooms", roomController.CreateRoom)
	// e.PUT("/rooms/:id", roomController.UpdateRoom)
	// e.DELETE("/rooms/:id", roomController.DeleteRoom)
}
