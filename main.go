package main

import (
	"emeeting/config"
	"emeeting/database"
	"emeeting/routes/user"
	"fmt"
	"log"

	"github.com/labstack/echo/v4"
	echoSwagger "github.com/swaggo/echo-swagger"
	_ "emeeting/docs"
)

func main() {
	e := echo.New()
	db := configDb.ConnectToDatabase()
	defer db.Close() 
	e.GET("/swagger/*", echoSwagger.WrapHandler)
	routes.RegisterRoutes(e)
	err := createtable.CreatedUser(db)
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
