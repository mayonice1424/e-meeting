package snackcontroller

import (
	"emeeting/models/"
	"net/http"

	"github.com/labstack/echo/v4"
	echoSwagger "github.com/swaggo/echo-swagger"
)

type Snack struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Price int    `json:"Price"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}

var snacks = []models.Snack{
	{ID: 1, Name: "Chips", Price: 10000},
	{ID: 2, Name: "Cookies", Price: 15000},
}

func main() {
	e := echo.New()

	e.GET("/swagger/*", echoSwagger.WrapHandler)
	e.GET("/snacks", GetSnacks)
	// e.GET("/snacks/:id", GetSnackByID)
	// e.POST("/snacks", createSnack)
	// e.PUT("/snacks/:id", UpdateSnack)
	// e.DELETE("/snacks/:id", DeleteSnack)

	e.Start(":8080")
}

// Get all snacks
func GetSnacks(c echo.Context) error {
	if len(snacks) == 0 {
		return c.JSON(http.StatusNotFound, ErrorResponse{Message: "No snacks found"})
	}
	return c.JSON(http.StatusOK, snacks)
}

// Get snack by ID
// func getSnackByID(c echo.Context) error {
// 	id := c.Param("id")
// 	IDInt, err := strconv.Atoi(id)
// 	if err != nil {
// 		return c.JSON(http.StatusBadRequest, ErrorResponse{Message: "Invalid snack ID"})
// 	}
// }

// 	for _, snack := range snacks {
// 		if snack.ID == id {
// 			c.JSON(http.StatusOK, snack)
// 			return
// 	}

// 	return c.JSON(http.StatusNotFound, ErrorResponse{Message: "Snack not found"})
// }
// // Create a new snack
// func createSnack(c echo.Context) error {
// 	var newSnack Snack
// 	err := c.Bind(&newSnack)
// 	if err != nil {
// 		return c.JSON(http.StatusBadRequest, ErrorResponse{Message: "Invalid request payload"})
// 	}

// 	if newSnack.Name == "" {
// 		return c.JSON(http.StatusBadRequest, ErrorResponse{Message: "Name is required"})
// 	}

// 	snacks = append(snacks, newSnack)
// 	return c.JSON(http.StatusCreated, newSnack)
// }

// // Update a snack
// func UpdateSnack(c echo.Context) error {
// 	id := c.Param("id")            // ambil parameter id yang akan diupdate
// 	IDInt, err := strconv.Atoi(id) // konversi string ke integer
// 	if err != nil {
// 		return c.JSON(http.StatusBadRequest, ErrorResponse{Message: "Invalid snack ID"})
// 	}

// 	var updatedSnack Snack       // buat variabel untuk menyimpan data snack yang akan diupdate
// 	err = c.Bind(&updatedSnack) // bind data dari request body ke variabel updatedSnack
// 	if err != nil {
// 		return c.JSON(http.StatusBadRequest, ErrorResponse{Message: "Invalid request payload"})
// 	}

// 	// cari user dengan ID yang sesuai
// 	for i, snack := range snacks {
// 		if user.ID == IDInt {
// 			snacks[i] = updatedUser // update data snack
// 			return c.JSON(http.StatusOK, updatedSnack)
// 		}
// 	}
// 	return c.JSON(http.StatusNotFound, ErrorResponse{Message: "Snack not found"})
// }

// // Delete a snack
// func DeleteSnack(c echo.Context) error {
// 	id := c.Param("id")            // ambil parameter id yang akan dihapus
// 	IDInt, err := strconv.Atoi(id) // konversi string ke integer
// 	if err != nil {
// 		return c.JSON(http.StatusBadRequest, ErrorResponse{Message: "Invalid snack ID"})
// 	}

// 	// cari snack dengan ID yang sesuai
// 	// jika ditemukan, hapus snack tersebut dari slice snacks
// 	for i, snack := range snacks {
// 		if snack.ID == IDInt {
// 			snacks = append(snacks[:i], snacks[i+1:]...) // hapus snack dari slice
// 			// menghapus snack dengan cara menggabungkan slice sebelum dan sesudah index yang dihapus
// 			// sehingga snack dengan ID tersebut tidak ada lagi di slice snacks
// 			// mengembalikan status 204 No Content sebagai respons
// 			// karena tidak ada konten yang dikembalikan setelah penghapusan
// 			return c.NoContent(http.StatusNoContent)
// 		}
// 	}
// 	return c.JSON(http.StatusNotFound, ErrorResponse{Message: "Snack not found"})
// }
