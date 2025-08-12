package roomController

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

// Room struct
type Room struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// In-memory data
var rooms = []Room{
	{ID: 1, Name: "Room A"},
	{ID: 2, Name: "Room B"},
}

func main() {
	e := echo.New()

	// Routes
	e.GET("/rooms", getRooms)
	e.GET("/rooms/:id", getRoomByID)
	e.POST("/rooms", createRoom)
	e.PUT("/rooms/:id", updateRoom)
	e.DELETE("/rooms/:id", deleteRoom)

	// Start server
	e.Start(":8080")
}

// Handlers
func getRooms(c echo.Context) error {
	return c.JSON(http.StatusOK, rooms)
}

func getRoomByID(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid ID"})
	}

	for _, room := range rooms {
		if room.ID == id {
			return c.JSON(http.StatusOK, room)
		}
	}
	return c.JSON(http.StatusNotFound, map[string]string{"message": "Room not found"})
}

func createRoom(c echo.Context) error {
	room := new(Room)
	if err := c.Bind(room); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid input"})
	}

	room.ID = len(rooms) + 1
	rooms = append(rooms, *room)
	return c.JSON(http.StatusCreated, room)
}

func updateRoom(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid ID"})
	}

	room := new(Room)
	if err := c.Bind(room); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid input"})
	}

	for i, r := range rooms {
		if r.ID == id {
			rooms[i].Name = room.Name
			return c.JSON(http.StatusOK, rooms[i])
		}
	}
	return c.JSON(http.StatusNotFound, map[string]string{"message": "Room not found"})
}

func deleteRoom(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid ID"})
	}

	for i, room := range rooms {
		if room.ID == id {
			rooms = append(rooms[:i], rooms[i+1:]...)
			return c.JSON(http.StatusOK, map[string]string{"message": "Room deleted"})
		}
	}
	return c.JSON(http.StatusNotFound, map[string]string{"message": "Room not found"})
}
