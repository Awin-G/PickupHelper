package repository

import (
	"context"
	"fmt"

	"pickup-helper/internal/config"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

// NewMySQL opens a sqlx.DB connection and configures pool parameters.
func NewMySQL(cfg *config.Config) (*sqlx.DB, error) {
	db, err := sqlx.Connect("mysql", cfg.MySQLDSN())
	if err != nil {
		return nil, fmt.Errorf("connect mysql: %w", err)
	}
	db.SetMaxOpenConns(cfg.MySQL.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MySQL.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.MySQL.ConnMaxLifetime)
	return db, nil
}

// PingMySQL verifies the database connection is alive.
func PingMySQL(ctx context.Context, db *sqlx.DB) error {
	if db == nil {
		return fmt.Errorf("nil db")
	}
	return db.PingContext(ctx)
}
