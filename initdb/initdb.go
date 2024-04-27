package initdb

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func InitDB() {
	var err error
	log.Println("Initializing database connection...")
	connStr := os.Getenv("DATABASE_URL")

	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Error can't connect to database: %s", err)
	}

	log.Println("Database connection successfully established.")
}
