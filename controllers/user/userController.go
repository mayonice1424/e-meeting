package userController

import (
	configDb "emeeting/config"
	"emeeting/models"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

var jwtKey = []byte(os.Getenv("SECRET_KEY")) 
func PasswordValidation(password string) bool {
	lowercaseRegex := `[a-z]`
	uppercaseRegex := `[A-Z]`
	digitRegex := `\d`
	specialCharRegex := `[!@#$%^&*()_+{}|:<>?]`
	minLengthRegex := `.{8,}`

	lowercaseRe := regexp.MustCompile(lowercaseRegex)
	uppercaseRe := regexp.MustCompile(uppercaseRegex)
	digitRe := regexp.MustCompile(digitRegex)
	specialCharRe := regexp.MustCompile(specialCharRegex)
	minLengthRe := regexp.MustCompile(minLengthRegex)

	return lowercaseRe.MatchString(password) &&
		uppercaseRe.MatchString(password) &&
		digitRe.MatchString(password) &&
		specialCharRe.MatchString(password) &&
		minLengthRe.MatchString(password)
}
// @Summary Endpoint create a new user
// @Descrition Create a new user with username, email, password
// @Tags users 
// @Accept json
// @Produce json
// @Param user body models.User true "User object"
// @Success 201 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Router /users/register [post]
func UserRegister(c echo.Context) error {
	db := configDb.ConnectToDatabase()
	defer db.Close()

	var newUser models.User
	err := c.Bind(&newUser)
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Invalid request payload"})
	}
	if !PasswordValidation(newUser.Password) {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Password must contain at least one lowercase letter, one uppercase letter, one digit, and one special character"})
	}
	if newUser.Password != newUser.Confirm_Password {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Password and Confirm Password do not match"})
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newUser.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "Error hashing password"})
	}

	_, err = db.Exec(`
			INSERT INTO users (email, password, username) 
			VALUES ($1, $2, $3)`,
			newUser.Email, 
			hashedPassword, 
			newUser.Username)
	if err != nil {
		log.Println("Error inserting user:", err)
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "Error creating user"})
	}

	claims := &jwt.StandardClaims{
		Subject:   newUser.Username, 
		ExpiresAt: time.Now().Add(24 * time.Hour).Unix(), 
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(jwtKey)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "Error generating token"})
	}

	return c.JSON(http.StatusOK, models.SuccessResponse{
		Data:    models.Token{Token: signedToken},
		Message: "User created successfully",
	})
}
