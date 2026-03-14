// Package migrations embeds SQL migration files and provides a helper to run
// them against a PostgreSQL database using golang-migrate.
package migrations

import (
	"embed"
	"errors"
	"fmt"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5" // pgx5 driver
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed *.sql
var sqlFiles embed.FS

// Run applies all pending UP migrations embedded in this package.
// dsn must be a postgres:// or pgx5:// connection string.
// It is idempotent: already-applied migrations produce no error.
func Run(dsn string, log *slog.Logger) error {
	src, err := iofs.New(sqlFiles, ".")
	if err != nil {
		return fmt.Errorf("migrations: create iofs source: %w", err)
	}

	// golang-migrate's pgx/v5 driver is registered under the "pgx5" scheme.
	// Rewrite postgres:// → pgx5:// so the correct driver is selected.
	pgx5DSN := "pgx5" + dsn[len("postgres"):]

	m, err := migrate.NewWithSourceInstance("iofs", src, pgx5DSN)
	if err != nil {
		return fmt.Errorf("migrations: create migrator: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrations: up: %w", err)
	}

	version, dirty, _ := m.Version()
	log.Info("migrations applied", slog.Uint64("version", uint64(version)), slog.Bool("dirty", dirty))
	return nil
}
