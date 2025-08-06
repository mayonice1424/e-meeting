package main

import (
	"emeeting/database"
	"fmt"
	"log"

	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()
	db := database.ConnectToDatabase()
	defer db.Close()

	e.GET("/generate-token", auth.GenerateTokenJWT)

	err := database.CreatedUser(db)
	if err != nil {
		log.Fatal("Error creating table: ", err)
	}
	err = database.CreateDataPersonalReservation(db)
	if err != nil {
		log.Fatal("Error creating 'data_personal_reservation' table: ", err)
	}
	err = database.CreateDataRoom(db)
	if err != nil {
		log.Fatal("Error creating 'room' table: ", err)
	}
	err = database.CreateDataSnack(db)
	if err != nil {
		log.Fatal("Error creating 'snack' table: ", err)
	}
	err = database.CreateDataBookingRoom(db)
	if err != nil {
		log.Fatal("Error creating 'booking' table: ", err)
	}
	fmt.Println("Echo server is running..")
	e.Logger.Fatal(e.Start(":8080"))
}
