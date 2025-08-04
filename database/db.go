package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func ConnectToDatabase() *sql.DB {
	// Ganti localhost dengan db, nama service dari Docker Compose
	connStr := "user=postgres dbname=postgres password=polkmn1234 host=localhost port=5431 sslmode=disable"  // Menggunakan port 5432 di container
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Database Connection Failed: ", err)
	}
	err = db.Ping()
	if err != nil {
		log.Fatal("Failed to ping the database: ", err)
	}
	fmt.Println("Database Connected Successfully")
	return db
}


func CreatedUser(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY NOT NULL,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    username VARCHAR(100) NOT NULL,
    password VARCHAR(100) NOT NULL,
    no_hp VARCHAR(20) NOT NULL,
    role VARCHAR(20) NOT NULL,
    status BOOLEAN DEFAULT false,
    language VARCHAR(50) NOT NULL,
    profile_picture VARCHAR(100) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP 
);
	`
	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create table: %v", err)
	}
	fmt.Println("Table 'users' created successfully!")
	return nil
}


func CreateDataPersonalReservation(db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS data_personal_reservation (
			id SERIAL PRIMARY KEY NOT NULL,
			id_user INT REFERENCES users(id) ON DELETE CASCADE NOT NULL,
			recervation_create_date TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
			reservation_update_date TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			name VARCHAR(50) NOT NULL,
			no_hp VARCHAR(20) NOT NULL,
			company_name VARCHAR(100) NOT NULL,
			reservation_status VARCHAR(10) NOT NULL,
			note VARCHAR(200) ,
			total_invoice FLOAT NOT NULL
		)
	`
	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create table: %v", err)
	}
	fmt.Println("Table 'data_personal_reservation' created successfully!")
	return nil
}
func CreateDataRoom(db *sql.DB) error {
		query := `
		CREATE TABLE IF NOT EXISTS room (
    id SERIAL PRIMARY KEY NOT NULL,
    name VARCHAR(100) NOT NULL,
    type VARCHAR(20) NOT NULL,
    picture VARCHAR(100) NOT NULL,
    price_per_hour INT NOT NULL,
    capacity INT NOT NULL
		)
		`
	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create table: %v", err)
	}
	fmt.Println("Table 'room' created successfully!")
	return nil
}
func CreateDataSnack(db *sql.DB) error {
		query := `
		CREATE TABLE IF NOT EXISTS snack_category (
    id SERIAL PRIMARY KEY NOT NULL,
    name VARCHAR(100) NOT NULL,
    satuan VARCHAR(100) NOT NULL,
    price INT NOT NULL
		)
		`
	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create table: %v", err)
	}
	fmt.Println("Table 'snack_category' created successfully!")
	return nil
}




func CreateDataBookingRoom(db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS data_booking_room (
    id SERIAL PRIMARY KEY NOT NULL,
    id_room INT NOT NULL,
    reservation_id INT NOT NULL,
    snack_id INT,
    start_date TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    end_date TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    total_participant INT NOT NULL,
    duration FLOAT NOT NULL,
    room_price FLOAT NOT NULL,
    total_snack INT,
    sub_total_invoice FLOAT NOT NULL,
    snack_price FLOAT,
    sub_total_room_price FLOAT NOT NULL,
    sub_total_snack_price FLOAT,
    FOREIGN KEY (id_room) REFERENCES room(id) ON DELETE CASCADE,
    FOREIGN KEY (snack_id) REFERENCES snack_category(id) ON DELETE CASCADE	
	)
	`
	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create table: %v", err)
	}
	fmt.Println("Table 'booking room' created successfully!")
	return nil
}