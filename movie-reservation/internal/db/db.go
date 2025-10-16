package db

import (
	"database/sql"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

const DSN = "root:adminpass@tcp(localhost:3306)/movie_reservation?parseTime=true"

var (
	db   *sql.DB
	once sync.Once
)

func InitDB() error {
	var err error
	once.Do(func() {
		db, err = sql.Open("mysql", DSN)
		if err != nil {
			return
		}
		db.SetConnMaxLifetime(time.Minute * 3)
		db.SetMaxOpenConns(10)
		db.SetMaxIdleConns(10)
		err = db.Ping()
		if err != nil {
			db = nil
		}
	})
	return err
}

// GetDB returns a *sql.DB instance
func GetDB() *sql.DB {
	if db == nil {
		log.Printf("DataBase not initialized.")
		return nil
	}
	return db
}

func Close() error {
	if db != nil {
		return db.Close()
	}
	return nil
}

// ResetDatabase conditionally wipes all data when DB_RESET_ON_START is set to a truthy value.
// It truncates tables (not dropping schema) so that seeding starts from a clean slate.
// Uses foreign_key_checks disable/enable to maintain referential integrity during truncation.
func ResetDatabase() error {
	flag := strings.ToLower(os.Getenv("DB_RESET_ON_START"))
	if flag == "" || flag == "0" || flag == "false" || flag == "no" {
		return nil
	}
	if db == nil {
		return nil
	}
	log.Println("⚠️  DB_RESET_ON_START enabled: wiping data before initialization ...")
	tables := []string{
		"reservation_seats",
		"seat_reservations",
		"reservations",
		"showtimes",
		"seats",
		"theaters",
		"movies",
		"genres",
		"users",
	}
	// Disable FK checks
	if _, err := db.Exec("SET FOREIGN_KEY_CHECKS = 0"); err != nil {
		return err
	}
	for _, t := range tables {
		if _, err := db.Exec("TRUNCATE TABLE " + t); err != nil {
			log.Printf("TRUNCATE %s failed: %v", t, err)
		}
	}
	// Re-enable FK checks
	if _, err := db.Exec("SET FOREIGN_KEY_CHECKS = 1"); err != nil {
		return err
	}
	log.Println("✅ Database cleared.")
	return nil
}
