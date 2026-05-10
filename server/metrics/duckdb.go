package metrics

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/marcboeker/go-duckdb"

	"github.com/YumikoKawaii/angelix/server/migrate"
	"github.com/YumikoKawaii/angelix/server/model"
)

var duckMigrations = []migrate.Migration{
	{Version: 1, SQL: `CREATE TABLE IF NOT EXISTS spans (
		member_id   VARCHAR     NOT NULL,
		trace_id    VARCHAR     NOT NULL,
		span_id     VARCHAR     NOT NULL,
		name        VARCHAR     NOT NULL,
		start_time  TIMESTAMPTZ NOT NULL,
		end_time    TIMESTAMPTZ NOT NULL,
		duration_ms BIGINT      NOT NULL,
		is_error    BOOLEAN     NOT NULL DEFAULT FALSE,
		recorded_at TIMESTAMPTZ NOT NULL
	)`},
	{Version: 2, SQL: `ALTER TABLE spans ADD COLUMN IF NOT EXISTS input_tokens BIGINT NOT NULL DEFAULT 0`},
	{Version: 3, SQL: `ALTER TABLE spans ADD COLUMN IF NOT EXISTS output_tokens BIGINT NOT NULL DEFAULT 0`},
	{Version: 4, SQL: `ALTER TABLE spans ADD COLUMN IF NOT EXISTS cache_read_tokens BIGINT NOT NULL DEFAULT 0`},
	{Version: 5, SQL: `ALTER TABLE spans ADD COLUMN IF NOT EXISTS cache_creation_tokens BIGINT NOT NULL DEFAULT 0`},
	{Version: 6, SQL: `ALTER TABLE spans ADD COLUMN IF NOT EXISTS model VARCHAR NOT NULL DEFAULT ''`},
}

// DuckDBConfig holds connection parameters.
// Fields carry kong tags so the struct can be embedded directly in a CLI/server config.
type DuckDBConfig struct {
	Path string `env:"DUCKDB_PATH" default:"metrics.duckdb" help:"DuckDB file path"`
}

type DuckDB struct {
	db *sql.DB
}

func NewDuckDB(cfg DuckDBConfig) (*DuckDB, error) {
	db, err := sql.Open("duckdb", cfg.Path)
	if err != nil {
		return nil, fmt.Errorf("duckdb open: %w", err)
	}
	if err := migrate.Run(db, duckMigrations); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("duckdb migrations: %w", err)
	}
	return &DuckDB{db: db}, nil
}

func (d *DuckDB) RecordSpans(memberID string, spans []*model.Span) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare(`
		INSERT INTO spans (member_id, trace_id, span_id, name, start_time, end_time, duration_ms, is_error, recorded_at, input_tokens, output_tokens, cache_read_tokens, cache_creation_tokens, model)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	defer stmt.Close()

	now := time.Now().UTC()
	for _, sp := range spans {
		if _, err := stmt.Exec(
			memberID, sp.TraceID, sp.SpanID, sp.Name,
			sp.StartTime, sp.EndTime, sp.DurationMs, sp.IsError, now,
			sp.InputTokens, sp.OutputTokens, sp.CacheReadTokens, sp.CacheCreationTokens, sp.Model,
		); err != nil {
			_ = tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

func (d *DuckDB) GetSpanSummary(memberID string) (*model.SpanSummary, error) {
	rows, err := d.db.Query(`
		SELECT
			name,
			COUNT(*)                    AS cnt,
			AVG(duration_ms)            AS avg_ms,
			SUM(is_error::INT)          AS errs,
			SUM(input_tokens)           AS input_tokens,
			SUM(output_tokens)          AS output_tokens,
			SUM(cache_read_tokens)      AS cache_read_tokens,
			SUM(cache_creation_tokens)  AS cache_creation_tokens
		FROM spans
		WHERE member_id = ?
		GROUP BY name
		ORDER BY cnt DESC
	`, memberID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	summary := &model.SpanSummary{MemberID: memberID}
	for rows.Next() {
		var (
			ts                  model.ToolStat
			errs                int64
			inputTokens         int64
			outputTokens        int64
			cacheReadTokens     int64
			cacheCreationTokens int64
		)
		if err := rows.Scan(&ts.Name, &ts.Count, &ts.AvgMs, &errs, &inputTokens, &outputTokens, &cacheReadTokens, &cacheCreationTokens); err != nil {
			return nil, err
		}
		ts.ErrorRate = float64(errs) / float64(ts.Count)
		ts.InputTokens = inputTokens
		ts.OutputTokens = outputTokens
		ts.CacheReadTokens = cacheReadTokens
		ts.CacheCreationTokens = cacheCreationTokens
		summary.TotalSpans += ts.Count
		summary.ErrorSpans += int(errs)
		summary.InputTokens += inputTokens
		summary.OutputTokens += outputTokens
		summary.CacheReadTokens += cacheReadTokens
		summary.CacheCreationTokens += cacheCreationTokens
		summary.TopTools = append(summary.TopTools, ts)
	}
	return summary, rows.Err()
}

func (d *DuckDB) Close() error {
	return d.db.Close()
}
