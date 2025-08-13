package uploadController

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
	"os"
	"path/filepath"
	"emeeting/models"
	"strings"
	"io"
)

// UploadHandler godoc
// @Summary Upload file
// @Description Upload an image file and save it to the server
// @Tags upload
// @Accept multipart/form-data
// @Produce json
// @Param Authorization header string true "Bearer <JWT Token>"
// @Param file formData file true "File to upload"
// @Success 200 {object} models.UploadResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/upload [post]
func UploadHandler(c echo.Context) error {
	fmt.Println("UploadHandler called")
	const maxSize = 1 << 20

	err := c.Request().ParseMultipartForm(maxSize)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"message": "file size too large",
		})
	}

	file, fileHeader, err := c.Request().FormFile("file")
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"message": "file not found",
		})
	}
	defer file.Close()

	allowedExtensions := []string{".jpg", ".jpeg", ".png"}
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if !contains(allowedExtensions, ext) {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"message": "file type is not supported",
		})
	}

	uploadDir := "./temp"
	err = os.MkdirAll(uploadDir, os.ModePerm)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": "internal server error",
		})
	}

	filePath := filepath.Join(uploadDir, fileHeader.Filename)

	_, err = os.Stat(filePath)
	if err == nil {
		err := os.Remove(filePath)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"message": "failed to delete existing file",
			})
		}
	}

	outFile, err := os.Create(filePath)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": "internal server error",
		})
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, file)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": "internal server error",
		})
	}

	imageURL := fmt.Sprintf("http://localhost:8080/temp/%s", filepath.Base(filePath))

	response := models.UploadResponse{
		Message: "upload file success",
	}
	response.Data.ImageURL = imageURL

	return c.JSON(http.StatusOK, response)
}

// Fungsi untuk memeriksa apakah ekstensi file sesuai dengan daftar yang diizinkan
func contains(allowed []string, ext string) bool {
	for _, a := range allowed {
		if ext == a {
			return true
		}
	}
	return false
}
