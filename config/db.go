package configDb

import (
	"database/sql" // query builder
	"fmt"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"log"
	"os"
)

var (
	dbHost            string
	dbPort            string
	dbUser            string
	dbPassword        string
	dbName            string
	JwtExpirationTime string
	JwtRefreshTime    string
	DB                *sql.DB
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	dbHost = os.Getenv("DB_HOST")
	dbPort = os.Getenv("DB_PORT")
	dbUser = os.Getenv("DB_USER")
	dbPassword = os.Getenv("DB_PASSWORD")
	dbName = os.Getenv("DB_NAME")
	JwtExpirationTime = os.Getenv("JWT_EXPIRATION_TIME")
	JwtRefreshTime = os.Getenv("JWT_REFRESH_EXPIRATION_TIME_SECONDS")
}

func ConnectToDatabase() *sql.DB {
	connStr := fmt.Sprintf("user=%s dbname=%s password=%s host=%s port=%s sslmode=disable", dbUser, dbName, dbPassword, dbHost, dbPort)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Database Connection Failed: ", err)
	}
	err = db.Ping()
	if err != nil {
		log.Fatal("Failed to ping the database: ", err)
	}
	fmt.Println("Database Connected Successfully")

	DB = db
	return db
}
