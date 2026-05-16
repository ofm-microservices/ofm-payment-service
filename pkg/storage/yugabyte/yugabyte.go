package db

import (
	"fmt"
	"time"

	"payment-service/config"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

var connectDB = sqlx.Connect

// Open creates the payment-service YugabyteDB connection pool.
func Open(cfg config.DBConfig) (*sqlx.DB, error) {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Name,
		cfg.SSLMode,
	)

	dbx, err := connectDB("pgx", dsn)
	if err != nil {
		return nil, WrapOpenDBError(err)
	}
	dbx.SetMaxOpenConns(cfg.MaxOpenConns)
	dbx.SetMaxIdleConns(cfg.MaxIdleConns)
	dbx.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	dbx.SetConnMaxIdleTime(2 * time.Minute)
	return dbx, nil
}
