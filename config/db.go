

package configDb

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