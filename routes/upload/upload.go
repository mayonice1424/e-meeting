package router

import (
	"emeeting/controllers/upload"
	auth "emeeting/middleware"
	"github.com/labstack/echo/v4"
)

func RegisterUploadRoutes(e *echo.Echo) {
	group:=e.Group("/api/v1")
	group.Use(auth.AuthMiddleware)
	group.POST("/upload", uploadController.UploadHandler)
}
