package migrate

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

type Migration struct {
	Version int
	SQL     string
}

const createTableSQL = `CREATE TABLE IF NOT EXISTS schema_migrations (
    version    INTEGER PRIMARY KEY,
    applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
)`

// Run applies pending migrations to a database/sql database (SQLite, DuckDB).
func Run(db *sql.DB, migrations []Migration) error {
	if _, err := db.Exec(createTableSQL); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}
	for _, m := range migrations {
		var count int
		if err := db.QueryRow(`SELECT COUNT(*) FROM schema_migrations WHERE version = ?`, m.Version).Scan(&count); err != nil {
			return fmt.Errorf("check migration %d: %w", m.Version, err)
		}
		if count > 0 {
			continue
		}
		if _, err := db.Exec(m.SQL); err != nil {
			return fmt.Errorf("migration %d: %w", m.Version, err)
		}
		if _, err := db.Exec(`INSERT INTO schema_migrations (version) VALUES (?)`, m.Version); err != nil {
			return fmt.Errorf("record migration %d: %w", m.Version, err)
		}
		slog.Info("applied migration", "version", m.Version)
	}
	return nil
}

// RunClickHouse applies pending migrations using the ClickHouse native driver.
func RunClickHouse(conn driver.Conn, migrations []Migration) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := conn.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
		    version    UInt32,
		    applied_at DateTime DEFAULT now()
		) ENGINE = MergeTree()
		ORDER BY version
	`); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	for _, m := range migrations {
		var count uint64
		if err := conn.QueryRow(ctx, `SELECT count() FROM schema_migrations WHERE version = ?`, uint32(m.Version)).Scan(&count); err != nil {
			return fmt.Errorf("check migration %d: %w", m.Version, err)
		}
		if count > 0 {
			continue
		}
		if err := conn.Exec(ctx, m.SQL); err != nil {
			return fmt.Errorf("migration %d: %w", m.Version, err)
		}
		if err := conn.Exec(ctx, `INSERT INTO schema_migrations (version) VALUES (?)`, uint32(m.Version)); err != nil {
			return fmt.Errorf("record migration %d: %w", m.Version, err)
		}
		slog.Info("applied migration", "version", m.Version)
	}
	return nil
}
