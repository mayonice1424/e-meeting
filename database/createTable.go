package createtable

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)



func CreateResetPassword(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS password_resets (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    token VARCHAR(255) NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);
	`
	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create table: %v", err)
	}
	fmt.Println("Table 'Reset Password' created successfully!")
	return nil
}



func CreatedUser(db *sql.DB) error {
	query := `
  CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY NOT NULL,
    name VARCHAR(100) DEFAULT '',  -- Kolom ini boleh NULL dan diberi string kosong default
    email VARCHAR(100) UNIQUE NOT NULL,
    username VARCHAR(100) NOT NULL,
    password VARCHAR(100) NOT NULL,
    no_hp VARCHAR(20) DEFAULT '',  -- Kolom ini boleh NULL dan diberi string kosong default
    role VARCHAR(20) DEFAULT '',  -- Kolom ini boleh NULL dan diberi string kosong default
    status VARCHAR(20) DEFAULT 'In-Active',
    language VARCHAR(50) DEFAULT '',
    profile_picture VARCHAR(100) DEFAULT '',
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
    capacity INT NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
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
    unit VARCHAR(100) NOT NULL,
    category VARCHAR(100) NOT NULL,
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
	fmt.Println("Table 'booking_room' created successfully!")
	return nil
}