package db

import (
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	"payment-service/config"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/yugabytedb"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type migrator interface{ Up() error }

var newMigrator = func(sourceURL, databaseURL string) (migrator, error) { return migrate.New(sourceURL, databaseURL) }

// RunMigrations applies the payment-service Yugabyte schema migrations.
func RunMigrations(cfg config.DBConfig) error {
	dsn := fmt.Sprintf(
		"yugabytedb://%s:%s@%s:%d/%s?sslmode=%s&x-migrations-table=%s",
		url.QueryEscape(cfg.User),
		url.QueryEscape(cfg.Password),
		cfg.Host,
		cfg.Port,
		cfg.Name,
		cfg.SSLMode,
		url.QueryEscape(cfg.MigrationsTable),
	)

	migrationPath := cfg.MigrationsPath
	if after, ok := strings.CutPrefix(migrationPath, "file://"); ok {
		if !filepath.IsAbs(after) {
			abs, err := filepath.Abs(after)
			if err != nil {
				return WrapResolveMigrationsPathError(err)
			}
			migrationPath = "file://" + abs
		}
	}

	m, err := newMigrator(migrationPath, dsn)
	if err != nil {
		return WrapCreateMigratorError(err)
	}
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return WrapRunMigrationsError(err)
	}
	return nil
}
