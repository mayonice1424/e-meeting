package utils

import (
	"crypto/rand"
	configDb "emeeting/config"
	"emeeting/models"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/gomail.v2"
)
var emailUser string 
var emailPassword string 
func init() {
    err := godotenv.Load()
    if err != nil {
        log.Fatal("Error loading .env file")
    }
		emailUser  = string(os.Getenv("EMAIL_USER"))
		emailPassword = string(os.Getenv("EMAIL_PASSWORD"))
}
func GenerateRandomString(length int) (string, error) {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}
func GenerateResetToken(email string) (string, error) {
	token, err := GenerateRandomString(32)
	if err != nil {
		log.Println("Error generating token:", err)
		return "", err
	}

	expiresAt := time.Now().Add(1 * time.Hour)
	userID, err := getUserIDByEmail(email)
	if err != nil {
		return "", err
	}

	passwordReset := models.PasswordReset{
		UserID:    userID,
		Token:     token,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}
	err = savePasswordReset(passwordReset)
	if err != nil {
		log.Println("Error saving password reset token:", err)
		return "", err
	}
	return token, nil
}
func savePasswordReset(reset models.PasswordReset) error {
	db := configDb.ConnectToDatabase()
	defer db.Close()
	_, err := db.Exec("INSERT INTO password_resets (user_id, token, expires_at, created_at) VALUES ($1, $2, $3, $4)", 
		reset.UserID, reset.Token, reset.ExpiresAt, reset.CreatedAt)
	return err
}

func getUserIDByEmail(email string) (int, error) {
	var userID int
	db := configDb.ConnectToDatabase()
	defer db.Close()
	err := db.QueryRow("SELECT id FROM users WHERE email = $1", email).Scan(&userID)
	if err != nil {
		return 0, errors.New("user not found")
	}
	return userID, nil
}

func SendResetEmail(userEmail, token string) error {
	resetLink := fmt.Sprintf("https://localhost:8080/reset-password?token=%s", token)
	emailContent := fmt.Sprintf("Klik link ini untuk mereset password Anda: %s", resetLink)
	emailUser := os.Getenv("EMAIL_USER")
  emailPassword := os.Getenv("EMAIL_PASSWORD")
	m := gomail.NewMessage()
	m.SetHeader("From", "mayonica1424@gmail.com") 
	m.SetHeader("To", userEmail)
	m.SetHeader("Subject", "Reset Password")
	m.SetBody("text/html", emailContent)
	fmt.Println("INI USER ",emailUser, emailPassword)
	d := gomail.NewDialer("smtp.gmail.com", 465, emailUser, emailPassword)
	err := d.DialAndSend(m)
	if err != nil {
		log.Println("Failed to send email:", err)
		return err
	}

	return nil
}

func VerifyResetToken(token string) (bool, int, error) {
	var userID int
	var expiresAt time.Time
	db := configDb.ConnectToDatabase()
	defer db.Close()

	err := db.QueryRow("SELECT user_id, expires_at FROM password_resets WHERE token = $1", token).Scan(&userID, &expiresAt)
	if err != nil {
		return false, 0, errors.New("invalid token")
	}
	fmt.Println("INI USER ID", userID)
	fmt.Println("INI EXPIRES AT", expiresAt)
	if expiresAt.Before(time.Now()) {
		return false, 0, errors.New("token expired")
	}
	return true, userID, nil
}

func ResetPassword(userID int, newPassword string) error {
	db := configDb.ConnectToDatabase()
	defer db.Close()
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = db.Exec("UPDATE users SET password = $1 WHERE id = $2", hashedPassword, userID)
	if err != nil {
		return err
	}
	_, err = db.Exec("DELETE FROM password_resets WHERE user_id = $1", userID)
	return err
}