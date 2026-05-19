package db

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
)

type Config struct {
	Driver   string
	Host     string
	Port     int
	Name     string
	User     string
	Password string
}

func Open(cfg Config) (*sqlx.DB, error) {
	var dsn string
	driver := cfg.Driver

	switch driver {
	case "mysql":
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=true&loc=Local",
			cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name)
	case "sqlite3":
		dsn = cfg.Name
	default:
		return nil, fmt.Errorf("unsupported db driver: %s", driver)
	}

	conn, err := sqlx.Open(driver, dsn)
	if err != nil {
		return nil, fmt.Errorf("db open: %w", err)
	}

	conn.SetMaxOpenConns(25)
	conn.SetMaxIdleConns(5)
	conn.SetConnMaxLifetime(5 * time.Minute)

	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("db ping: %w", err)
	}

	return conn, nil
}
