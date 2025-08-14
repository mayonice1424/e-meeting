package roomController

import (
	// "database/sql"
	configDb "emeeting/config"
	"emeeting/models"
	"fmt"
	"net/http"
	"os"
	"strings"
	"io"
	"path/filepath"
	"strconv"

	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"
	// "fmt"
)

// GetRooms godoc
// @Summary Endpoint Get all rooms with pagination
// @Description Get all rooms with pagination, name, type, picture, price_per_hour, and capacity
// @Tags rooms
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer <JWT Token>"
// @Param name query string false "Room name"
// @Param type query string false "Room type"
// @Param capacity query int false "Room capacity"
// @Param page query int false "Page number" default(1)
// @Param pageSize query int false "Number of records per page" default(10)
// @Success 200 {object} models.SuccessResponseRoom
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/rooms [get]
func GetRooms(c echo.Context) error {
	db := configDb.ConnectToDatabase()
	defer db.Close()

	// Read query parameters
	name := c.QueryParam("name")
	roomType := c.QueryParam("type")
	capacity := c.QueryParam("capacity")
	pageStr := c.QueryParam("page")
	pageSizeStr := c.QueryParam("pageSize")

	// Default values for pagination
	page := 1
	pageSize := 10

	// Parse page and pageSize query params
	if pageStr != "" {
		page, _ = strconv.Atoi(pageStr)
		if page <= 0 {
			page = 1
		}
	}
	if pageSizeStr != "" {
		pageSize, _ = strconv.Atoi(pageSizeStr)
		if pageSize <= 0 {
			pageSize = 10
		}
	}

	// Calculate offset
	offset := (page - 1) * pageSize

	// Building query to fetch filtered rooms with pagination
	query := "SELECT id, name, type, picture, price_per_hour, capacity, created_at, updated_at FROM room WHERE 1=1"

	// Add filters based on query params
	var queryParams []interface{}
	paramCount := 1

	if name != "" {
		query += " AND name ILIKE '%' || $1 || '%'"
		queryParams = append(queryParams, name)
		paramCount++
	}
	if roomType != "" {
		query += " AND type = $" + strconv.Itoa(paramCount)
		queryParams = append(queryParams, roomType)
		paramCount++
	}
	if capacity != "" {
		query += " AND capacity = $" + strconv.Itoa(paramCount)
		queryParams = append(queryParams, capacity)
		paramCount++
	}

	// Add pagination
	query += " LIMIT $" + strconv.Itoa(paramCount) + " OFFSET $" + strconv.Itoa(paramCount+1)
	queryParams = append(queryParams, pageSize, offset)

	// Execute the query
	rows, err := db.Query(query, queryParams...)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "Internal Server Error"})
	}
	defer rows.Close()

	var rooms []models.RoomById
	for rows.Next() {
		var room models.RoomById
		if err := rows.Scan(&room.ID, &room.Name, &room.Type, &room.Picture, &room.PricePerHour, &room.Capacity, &room.CreatedAt, &room.UpdatedAt); err != nil {
			return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "Internal Server Error"})
		}
		rooms = append(rooms, room)
	}

	if err := rows.Err(); err != nil {
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "Internal Server Error"})
	}

	if rooms == nil {
		return c.JSON(http.StatusNotFound, models.ErrorResponse{Message: "No rooms found"})
	}

	var totalRooms int
	err = db.QueryRow("SELECT COUNT(*) FROM room WHERE 1=1").Scan(&totalRooms)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "Internal Server Error"})
	}

	totalPages := (totalRooms + pageSize - 1) / pageSize

	return c.JSON(http.StatusOK, models.SuccessResponseRooms{
		Data:      rooms,
		Message:   "Rooms retrieved successfully",
		Page:      page,
		PageSize:  pageSize,
		TotalPage: totalPages,
		TotalData: totalRooms,
	})
}


// func GetRoomByID(c echo.Context) error {
// 	// Ambil parameter ID dari URL
// 	id := c.Param("id")
// 	if id == "" {
// 		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Room ID is required"})
// 	}
// 	// Inisialisasi objek Room
// 	room := models.RoomById{}
// 	// Validasi ID
// 	if _, err := strconv.Atoi(id); err != nil {
// 		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid Room ID"})
// 	}
// 	// Query database untuk mencari Room berdasarkan ID
// 	db := configDb.ConnectToDatabase()
// 	defer db.Close()
// 	query := "SELECT id, name, type, picture, price_per_hour, capacity FROM room WHERE id = $1"
// 	err := db.QueryRow(query, id).Scan(&room.ID, &room.Name, &room.Type, &room.Picture, &room.PricePerHour, &room.Capacity)
// 	if err != nil {
// 		if err == sql.ErrNoRows {
// 			return c.JSON(http.StatusNotFound, map[string]string{"message": "Room not found"})
// 		}
// 		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to retrieve room: " + err.Error()})
// 	}

// 	// Return data Room dalam format JSON
// 	return c.JSON(http.StatusOK, room)
// }

// CreateRoom godoc
// @Summary Endpoint create a new room
// @Description Create a new room with name, type, picture, price_per_hour, and capacity
// @Tags rooms
// @Accept json
// @Produce json
// @Param room body models.CreateRoom true "Room object"
// @Param Authorization header string true "Bearer <JWT Token>"
// @Success 201 {object} models.SuccessResponseRoom
// @Failure 400 {object} models.ErrorResponse
// @Router /api/v1/rooms [post]
func CreateRoom(c echo.Context) error {
	fmt.Println("CreateRoom called1")
	db := configDb.ConnectToDatabase()
	defer db.Close()
	var newRoom models.CreateRoom
	if err := c.Bind(&newRoom); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Invalid input"})
	}
	query := "INSERT INTO room (name, type, picture, price_per_hour, capacity) VALUES ($1, $2, $3, $4, $5)"
	_, err := db.Exec(query, newRoom.Name, newRoom.Type, newRoom.Picture, newRoom.PricePerHour, newRoom.Capacity)
	if err != nil {
		return c.JSON(http.StatusNotFound, models.ErrorResponse{Message: "Failed to create room: " + err.Error()})
	}

	return c.JSON(http.StatusCreated, models.SuccessResponseRoom{
		Message: "Room created successfully",
	})
}


// UpdateRoom godoc
// @Summary Endpoint for updating room by id
// @Description Update room by id with id in the URL path
// @Tags rooms
// @Accept json
// @Produce json
// @Param id path int true "Room ID"
// @Param Authorization header string true "Bearer <JWT Token>"
// @Param room body models.CreateRoom true "Room update object"
// @Success 200 {object} models.SuccessResponseRoom
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/rooms/{id} [put]
func UpdateRoom(c echo.Context) error {
	db := configDb.ConnectToDatabase()
	defer db.Close()
	id := c.Param("id")
	room := new(models.RoomById)
	if err := c.Bind(room); err != nil {
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "Invalid input"})
	}
	updateFields := []string{}
	updateValues := []interface{}{}
	paramCount := 1

	if room.Name != "" {
		updateFields = append(updateFields, fmt.Sprintf("name=$%d", paramCount))
		updateValues = append(updateValues, room.Name)
		paramCount++
	}
	if room.Type != "" {
		updateFields = append(updateFields, fmt.Sprintf("type=$%d", paramCount))
		updateValues = append(updateValues, room.Type)
		paramCount++
	}
	if room.Picture != "" {
		if strings.Contains(room.Picture, "/temp/") {
			fileName := filepath.Base(room.Picture)
			sourcePath := filepath.Join("./temp", fileName)
			destPath := filepath.Join("./uploads", fileName)
			sourcePath = filepath.ToSlash(sourcePath)
			destPath = filepath.ToSlash(destPath)
			fmt.Println("Source Path:", sourcePath)
			fmt.Println("Destination Path:", destPath)
			err := moveFile(sourcePath, destPath)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "Failed to move file: " + err.Error()})
			}
			room.Picture = fmt.Sprintf("http://localhost:8080/uploads/%s", fileName)
		}
		updateFields = append(updateFields, fmt.Sprintf("picture=$%d", paramCount))
		updateValues = append(updateValues, room.Picture)
		paramCount++
	}
	if room.PricePerHour != 0 {
		updateFields = append(updateFields, fmt.Sprintf("price_per_hour=$%d", paramCount))
		updateValues = append(updateValues, room.PricePerHour)
		paramCount++
	}
	if room.Capacity != 0 {
		updateFields = append(updateFields, fmt.Sprintf("capacity=$%d", paramCount))
		updateValues = append(updateValues, room.Capacity)
		paramCount++
	}
	if len(updateFields) == 0 {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "No fields to update"})
	}
	updateFields = append(updateFields, fmt.Sprintf("updated_at = CURRENT_TIMESTAMP"))

	query := fmt.Sprintf("UPDATE room SET %s WHERE id=$%d", strings.Join(updateFields, ", "), paramCount)
	configDb.ConnectToDatabase()
	defer db.Close()
	updateValues = append(updateValues, id)
	_, err := db.Exec(query, updateValues...)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "Failed to update room "})
	}
	return c.JSON(http.StatusOK, models.SuccessResponseRoom{Message: "Room updated successfully"})
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

func moveFile(sourcePath string, destPath string) error {
	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}



	return nil
}