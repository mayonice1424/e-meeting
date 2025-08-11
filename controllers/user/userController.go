package userController

import (
	"database/sql"
	configDb "emeeting/config"
	"emeeting/models"
	"emeeting/utils"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

var jwtKey []byte

func init() {
    err := godotenv.Load()
    if err != nil {
        log.Fatal("Error loading .env file")
    }
    jwtKey = []byte(os.Getenv("SECRET_KEY"))
    if len(jwtKey) == 0 {
        log.Fatal("SECRET_KEY is not set in the environment")
    }
}
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
	return c.JSON(http.StatusCreated, models.SuccessResponse{
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
	var role string
	err = db.QueryRow("SELECT id, password, email,role FROM users WHERE username = $1", loginUser.Username).Scan(&userId, &storedPassword, &email, &role)
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
		"role" : role,
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
		"role" : role,
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
	fmt.Println("JWT KEY", string(jwtKey))
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return "", fmt.Errorf("error generating token: %v", err)
	}
	return tokenString, nil
}

// RequestResetPassword godoc
// @Summary Endpoint for user reset password
// @Description Reset Password User with email
// @Tags Reset Password
// @Accept json
// @Produce json
// @Param user body models.EmailResetPassword true "User reset password object"
// @Success 200 {object} models.SuccessResponseResetPassword
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /password/reset [post]
func RequestResetPassword(c echo.Context) error {
	var email models.EmailResetPassword
	err:=c.Bind(&email)
	if err != nil {
		return c.JSON(http.StatusForbidden, models.ErrorResponse{Message: "Email not found"})
	}
	token, err := utils.GenerateResetToken(email.Email)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, "Email not found")
	}

	err = utils.SendResetEmail(email.Email, token)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "Internal Server Error")
	}
	tokenData := models.TokenResetPassword{Token: token}
	return c.JSON(http.StatusOK, models.SuccessResponseResetPassword{
		Data:    tokenData,
	})
}

// ResetPassword godoc
// @Summary Endpoint for user reset password
// @Description Reset Password User with email
// @Tags Reset Password
// @Accept json
// @Produce json
// @Param id path int true "User ID" // Add this for the ID in the URL path
// @Param Authorization header string true "Bearer <JWT Token>" 
// @Param user body models.ResetPasswordById true "User reset password object"
// @Success 200 {object} models.SuccessResponseResetPassword
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /password/reset/{id} [put]
func ResetPassword(c echo.Context) error {
	userId := c.Param("id")
	if userId == "" {
			return c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "User ID is required"})
	}
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		return c.JSON(http.StatusUnauthorized, models.ErrorResponse{Message: "Authorization header is required"})
	}
	var passwordPayload models.ResetPasswordById
	err := c.Bind(&passwordPayload)
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Invalid request payload"})
	}
	if !PasswordValidation(passwordPayload.NewPassword) {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Password must contain at least one lowercase letter, one uppercase letter, one digit, and one special character"})
	}
	if passwordPayload.NewPassword != passwordPayload.ConfirmPassword{
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "Password and Confirm Password do not match"})
	}
	valid, userID, err := utils.VerifyResetToken(authHeader)
	if err != nil || !valid {
		return c.JSON(http.StatusBadRequest, "Invalid or expired token")
	}
	err = utils.ResetPassword(userID, passwordPayload.NewPassword)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "Error resetting password")
	}
	return c.JSON(http.StatusOK, "Password successfully reset")
}


// UserById godoc
// @Summary Endpoint for user by id
// @Description User by id with id in the URL path
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param Authorization header string true "Bearer <JWT Token>" 
// @Success 200 {object} models.SuccessResponseUser
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/users/{id} [get]
func UserById(c echo.Context) error {
	db := configDb.ConnectToDatabase()
	defer db.Close()
	idParam  := c.Param("id")
	urlID, err := strconv.Atoi(idParam)
	claims, ok := c.Get("userClaims").(jwt.MapClaims)
	if !ok {
		return c.JSON(http.StatusUnauthorized, models.ErrorResponse{Message: "user not found"})
	}
	fmt.Println("User ID from token:", claims["userId"])
	fmt.Println("User ID from params:", urlID)

	userIdClaim, ok := claims["userId"]
	if !ok {
		return c.JSON(http.StatusUnauthorized, models.ErrorResponse{Message: "User ID not found in token"})
	}

	var userIdFromToken int
	switch v := userIdClaim.(type) {
	case float64:
		userIdFromToken = int(v)
	case string:
		idParsed, err := strconv.Atoi(v)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, models.ErrorResponse{Message: "unauthorized"})
		}
		userIdFromToken = idParsed
	default:
		return c.JSON(http.StatusUnauthorized, models.ErrorResponse{Message: "unauthorized"})
	}

	if userIdFromToken != urlID {
		return c.JSON(http.StatusForbidden, models.ErrorResponse{Message: "unauthorized"})
	}
	if urlID == 0 {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{Message: "User ID is required"})
	}
	var user models.User
	query := "SELECT id, name, email, username, no_hp, role, status, language, profile_picture, created_at, updated_at FROM users WHERE id = $1"
	err = db.QueryRow(query, urlID).Scan(&user.ID, &user.Name, &user.Email, &user.Username, &user.No_HP, &user.Role, &user.Status, &user.Language, &user.Profile_Picture, &user.Created_At, &user.Updated_At)
	if err != nil {
		fmt.Println("Error fetching user details:", err)
		if errors.Is(err, sql.ErrNoRows) {
			return c.JSON(http.StatusNotFound, models.ErrorResponse{Message: "User not found"})
		}
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: "internal server error"})
	}
	return c.JSON(http.StatusOK, models.SuccessResponseUser{
		Data:    []models.User{user},
		Message: "User retrieved successfully",
	})
}