package auth

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	// "strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
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

func ValidateRefreshToken(c echo.Context) error {
	tokenString := c.Request().Header.Get("Authorization")
	if tokenString == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"message": "authorization header is missing"})
	}
	if tokenString[:7] != "Bearer " {
		return c.JSON(http.StatusUnauthorized, map[string]string{"message": "invalid token format"})
	}

	tokenString = tokenString[7:]
	fmt.Println("INI AUTH HEADER", tokenString)
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return jwtKey, nil
	})
	fmt.Println("Token String with bearer:", token)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"message": "failed to parse token: " + err.Error()})
	}
	if !token.Valid {
		return c.JSON(http.StatusUnauthorized, map[string]string{"message": "token is not valid"})
	}
	if _, ok := token.Claims.(jwt.MapClaims); !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"message": "invalid token"})
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"message": "invalid token claims"})
	}
	email, ok := claims["email"].(string)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"message": "email claim is missing"})
	}
	username, ok := claims["username"].(string)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"message": "username claim is missing"})
	}	
	userId, ok := claims["userId"].(float64)
	if !ok {	

		return c.JSON(http.StatusUnauthorized, map[string]string{"message": "userId claim is missing"})
	}	
	role, ok := claims["role"].(string)
	if !ok {	
		return c.JSON(http.StatusUnauthorized, map[string]string{"message": "role claim is missing"})
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email":    email,
		"username": username,
		"userId":   userId,
		"role":     role,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	})
	accessTokenString, err := accessToken.SignedString(jwtKey)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "error generating access token: " + err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{
		"accessToken":  accessTokenString,
	})
}


func ValidateTokenJWT(c echo.Context) (jwt.MapClaims, error) {
	tokenString := c.Request().Header.Get("Authorization")
	if tokenString == "" {
		return nil, errors.New("authorization header is missing")
	}
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return jwtKey, nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}
	if !token.Valid {
		return nil, errors.New("token is not valid")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}


func AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		claims, err := ValidateTokenJWT(c)
	if err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{"message": err.Error()})
		}
		c.Set("userClaims", claims)
		return next(c)
	}
}