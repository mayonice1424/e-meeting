package main

import (
	"emeeting/config"
	"emeeting/database"
	"emeeting/middleware"
	user "emeeting/routes/user"
	snack "emeeting/routes/snack"

	"fmt"
	"log"

	_ "emeeting/docs"

	"github.com/labstack/echo/v4"
	echoSwagger "github.com/swaggo/echo-swagger"
)

func main() {
	e := echo.New()
	db := configDb.ConnectToDatabase()
	defer db.Close()
	e.GET("/swagger/*", echoSwagger.WrapHandler)
	e.GET("/validate-token", func(c echo.Context) error {
		claims, err := auth.ValidateTokenJWT(c)
		if err != nil {
			return c.JSON(401, map[string]string{"error": err.Error()})
		}
		return c.JSON(200, claims)
	})
	e.GET("/validate-refresh-token", auth.ValidateRefreshToken)
	user.RegisterRoutes(e)
	snack.SnackRoutes(e)
	err := createtable.CreatedUser(db)
	if err != nil {
		log.Fatal("Error creating table: ", err)
	}
	err = createtable.CreateResetPassword(db)
	if err != nil {
		log.Fatal("Error creating table: ", err)
	}
	err = createtable.CreateDataPersonalReservation(db)
	if err != nil {
		log.Fatal("Error creating 'data_personal_reservation' table: ", err)
	}
	err = createtable.CreateDataRoom(db)
	if err != nil {
		log.Fatal("Error creating 'room' table: ", err)
	}
	err = createtable.CreateDataSnack(db)
	if err != nil {
		log.Fatal("Error creating 'snack' table: ", err)
	}
	err = createtable.CreateDataBookingRoom(db)
	if err != nil {
		log.Fatal("Error creating 'booking' table: ", err)
	}
	fmt.Println("Echo server is running..")
	e.Logger.Fatal(e.Start(":8080"))
}
