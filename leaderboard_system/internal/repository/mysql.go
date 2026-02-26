package repository

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// MySQLConfig holds MySQL connection parameters.
type MySQLConfig struct {
	User            string
	Password        string
	Host            string
	Port            string
	DBName          string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// NewMySQL creates a new MySQL connection with connection pooling.
func NewMySQL(cfg MySQLConfig) (*sql.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName)
	return openAndConfigureMySQL(dsn, cfg)
}

// NewMySQLFromDSN creates a MySQL connection from a raw DSN string.
func NewMySQLFromDSN(dsn string) (*sql.DB, error) {
	return openAndConfigureMySQL(dsn, MySQLConfig{})
}

// openAndConfigureMySQL opens a MySQL connection and applies pool settings.
func openAndConfigureMySQL(dsn string, cfg MySQLConfig) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("opening mysql: %w", err)
	}

	maxOpenConns := cfg.MaxOpenConns
	if maxOpenConns == 0 {
		maxOpenConns = 10
	}
	maxIdleConns := cfg.MaxIdleConns
	if maxIdleConns == 0 {
		maxIdleConns = 10
	}
	connMaxLifetime := cfg.ConnMaxLifetime
	if connMaxLifetime == 0 {
		connMaxLifetime = 3 * time.Minute
	}

	db.SetConnMaxLifetime(connMaxLifetime)
	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("pinging mysql: %w", err)
	}

	return db, nil
}
