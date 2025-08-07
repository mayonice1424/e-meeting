package userController

import (
	configDb "emeeting/config"
	"emeeting/models"
	"fmt"
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
// @Param user body models.CreateUser true "User object"
// @Success 201 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Router /users/register [post]
func UserRegister(c echo.Context) error {
	db := configDb.ConnectToDatabase()
	defer db.Close()

	var newUser models.CreateUser
	var responseUser models.User
	var data *int
	data = nil
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
	
	err = db.QueryRow(`
		INSERT INTO users (email, password, username) 
		VALUES ($1, $2, $3) 
		RETURNING id`, 
		newUser.Email, 
		hashedPassword, 
		newUser.Username).Scan(&responseUser.ID)
	if err != nil {
		log.Println("Error inserting user:", err)
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "Error creating user"})
	}
	// query:= "select id, name, email, username, no_hp, role, status, language, profile_picture, created_at, updated_at  from users where id = $1"
	// err = db.QueryRow(query,responseUser.ID).Scan( &responseUser.ID, &responseUser.Name, &responseUser.Email, &responseUser.Username, &responseUser.No_HP, &responseUser.Role, &responseUser.Status, &responseUser.Language, &responseUser.Profile_Picture, &responseUser.Created_At, &responseUser.Updated_At  )
	// 	if err != nil {
	// 	log.Println("Error fetching user details:", err)
	// 	return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "Error retrieving user details"})
	// }
	return c.JSON(http.StatusOK, models.SuccessResponse{
		Data:    data,
		Message: "User created successfully",
	})
}


// UserLogin godoc
// @Summary Endpoint for user login
// @Description Login user with username and password
// @Tags users
// @Accept json
// @Produce json
// @Param user body models.LoginUser true "User login object"
// @Success 200 {object} models.SuccessResponseLogin
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /users/login [post]
func UserLogin(c echo.Context) error {
	db := configDb.ConnectToDatabase()
	defer db.Close()
	var loginUser models.LoginUser
	var token models.Token
	err := c.Bind(&loginUser)
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Login Failed"})
	}

	var storedPassword string
	var userId int
	var email string
	err = db.QueryRow("SELECT id, password, email FROM users WHERE username = $1", loginUser.Username).Scan(&userId, &storedPassword, &email)
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Login Failed"})
	}
	err = bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(loginUser.Password))
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Login Failed"})
	}
accessClaimsMap := jwt.MapClaims{
		"email":    email,       
		"username": loginUser.Username,
		"userId":   userId,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	}

	accessToken, err := generateJWTToken(accessClaimsMap)
	if err != nil {
		log.Println("Error generating access token:", err)
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "Error generating access token"})
	}

	refreshClaims := jwt.MapClaims{
		"email":    email,       
		"username": loginUser.Username,
		"userId":   userId,
		"exp":      time.Now().Add(30 * 24 * time.Hour).Unix(), 
	}

	refreshToken, err := generateJWTToken(refreshClaims)
	if err != nil {
		log.Println("Error generating refresh token:", err)
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "Error generating refresh token"})
	}

	token.Access_Token = accessToken
	token.Refresh_Token = refreshToken

	return c.JSON(http.StatusOK, models.SuccessResponseLogin{
		Data:    token,
		Message: "Login successful",
	})
}
func generateJWTToken(claims jwt.MapClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return "", fmt.Errorf("Error generating token: %v", err)
	}
	return tokenString, nil
}