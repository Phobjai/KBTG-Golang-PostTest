package tax

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/lib/pq"
)

var db *sql.DB

func InitDB() {
	var err error
	log.Println("Initializing database connection...")
	connStr := os.Getenv("DATABASE_URL")

	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Error can't connect to database: %s", err)
	}

	// Attempt to ping the database to verify connection

	log.Println("Database connection successfully established.")
}
