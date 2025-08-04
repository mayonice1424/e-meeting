package main

import (
	"emeeting/database"
	"fmt"

	"github.com/labstack/echo/v4"
)

func main(){
e := echo.New()
db:= database.ConnectToDatabase()
defer db.Close()

fmt.Println("Echo server is running.. ")
e.Logger.Fatal(e.Start(":8080"))

}