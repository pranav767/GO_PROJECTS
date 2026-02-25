package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var (
	db   *sql.DB
	once sync.Once
)

// getDSN builds the MySQL DSN from environment variables with defaults
func getDSN() string {
	user := os.Getenv("DB_USER")
	if user == "" {
		user = "root"
	}
	pass := os.Getenv("DB_PASS")
	if pass == "" {
		pass = "adminpass"
	}
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "localhost"
	}
	port := os.Getenv("DB_PORT")
	if port == "" {
		port = "3306"
	}
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "leaderboard_system"
	}
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", user, pass, host, port, dbName)
}

func InitDB() error {
	var err error
	once.Do(func() {
		dsn := getDSN()
		db, err = sql.Open("mysql", dsn)
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
