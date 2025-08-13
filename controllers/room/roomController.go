package roomController

<<<<<<< HEAD
import (
	"database/sql"
	configDb "emeeting/config"
	"emeeting/models"
	"net/http"
	"strconv"
=======
// import (
// 	"net/http"
// 	"strconv"
// 	"database/sql"
// 	configDb "emeeting/config"
// 	"fmt"
>>>>>>> adec5aa7faf61884b208b07fb9c0a278beec7930

// 	"github.com/labstack/echo/v4"
// )

<<<<<<< HEAD
// Handlers
func GetRooms(c echo.Context) error {
	db := configDb.ConnectToDatabase()
	defer db.Close()
	var rooms []models.Room
	query := "SELECT id, name, type, picture, price_per_hour, capacity FROM room"
	rows, err := db.Query(query)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to retrieve rooms: " + err.Error()})
	}
	defer rows.Close()
	for rows.Next() {
		var room models.Room
		if err := rows.Scan(&room.ID, &room.Name, &room.Type, &room.Picture, &room.PricePerHour, &room.Capacity); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to scan room: " + err.Error()})
		}
		rooms = append(rooms, room)
	}
	if err := rows.Err(); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Error iterating over rooms: " + err.Error()})
	}
	if len(rooms) == 0 {
		return c.JSON(http.StatusNotFound, map[string]string{"message": "No rooms found"})
	}
	return c.JSON(http.StatusOK, rooms)
}

func GetRoomByID(c echo.Context) error {
	// Ambil parameter ID dari URL
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Room ID is required"})
	}
	// Inisialisasi objek Room
	room := models.Room{}
	// Validasi ID
	if _, err := strconv.Atoi(id); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid Room ID"})
	}
	// Query database untuk mencari Room berdasarkan ID
	db := configDb.ConnectToDatabase()
	defer db.Close()
	query := "SELECT id, name, type, picture, price_per_hour, capacity FROM room WHERE id = $1"
	err := db.QueryRow(query, id).Scan(&room.ID, &room.Name, &room.Type, &room.Picture, &room.PricePerHour, &room.Capacity)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, map[string]string{"message": "Room not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to retrieve room: " + err.Error()})
	}

	// Return data Room dalam format JSON
	return c.JSON(http.StatusOK, room)
}
func CreateRoom(c echo.Context) error {
	db := configDb.ConnectToDatabase()
	defer db.Close()
	room := new(models.Room)
	if err := c.Bind(room); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Invalid input"})
	}
	query := "INSERT INTO rooms (name, type, picture, price_per_hour, capacity) VALUES (, ) RETURNING id"
	err := db.QueryRow(query, room.Name, room.Capacity).Scan(&room.ID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"message": "Failed to create room"})
	}

	return c.JSON(http.StatusOK, room)
}

func UpdateRoom(c echo.Context) error {
	db := configDb.ConnectToDatabase()
	defer db.Close()
	id := c.Param("id")
	room := new(models.Room)
	if err := c.Bind(room); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Invalid input"})
	}

	query := "UPDATE rooms SET name=$1, capacity=$2 WHERE id=$3"
	_, err := db.Exec(query, room.Name, room.Capacity, id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"message": "Failed to update room"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Room updated successfully"})
}

func DeleteRoom(c echo.Context) error {
	db := configDb.ConnectToDatabase()
	defer db.Close()
	id := c.Param("id")
	query := "DELETE FROM rooms WHERE id=$1"
	_, err := db.Exec(query, id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"message": "Failed to delete room"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Room deleted successfully"})
}
=======
// func CreateRoom(c echo.Context) error

// }

>>>>>>> adec5aa7faf61884b208b07fb9c0a278beec7930
