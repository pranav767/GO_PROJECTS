package db

import (
	"database/sql"
	"log"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

const dsn = "admin:adminpass@tcp(localhost:3306)/e-commerce?parseTime=true"

var (
	db   *sql.DB
	once sync.Once
)

// Init initializes the MySQL connection (call once at startup)
func Init() error {
	var err error
	once.Do(func() {
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

// GetDB returns the *sql.DB instance
func GetDB() *sql.DB {
	if db == nil {
		log.Printf("Warning: Database not initialized.")
		return nil
	}
	return db
}

// Close gracefully disconnects from MySQL
func Close() error {
	if db != nil {
		return db.Close()
	}
	return nil
}

// Example: CreateUser inserts a new user and returns the new ID
func CreateUser(username, passwordHash string) (int64, error) {
	result, err := db.Exec("INSERT INTO users (username, password_hash) VALUES (?, ?)", username, passwordHash)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// Example: GetUserByUsername fetches a user by username
type User struct {
	ID           int
	Username     string
	PasswordHash string
}

func GetUserByUsername(username string) (*User, error) {
	var user User
	err := db.QueryRow("SELECT id, username, password_hash FROM users WHERE username = ?", username).
		Scan(&user.ID, &user.Username, &user.PasswordHash)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
