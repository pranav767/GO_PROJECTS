package db

import (
	"database/sql"
	"log"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

const DSN = "root:adminpass@tcp(localhost:3306)/image_processing"

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
