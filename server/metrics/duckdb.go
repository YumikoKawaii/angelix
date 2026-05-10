package metrics

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/marcboeker/go-duckdb"

	"github.com/YumikoKawaii/angelix/server/model"
)

const duckSchema = `
CREATE TABLE IF NOT EXISTS spans (
    member_id   VARCHAR     NOT NULL,
    trace_id    VARCHAR     NOT NULL,
    span_id     VARCHAR     NOT NULL,
    name        VARCHAR     NOT NULL,
    start_time  TIMESTAMPTZ NOT NULL,
    end_time    TIMESTAMPTZ NOT NULL,
    duration_ms BIGINT      NOT NULL,
    is_error    BOOLEAN     NOT NULL DEFAULT FALSE,
    recorded_at TIMESTAMPTZ NOT NULL
);
`

type DuckDB struct {
	db *sql.DB
}

func NewDuckDB(path string) (*DuckDB, error) {
	db, err := sql.Open("duckdb", path)
	if err != nil {
		return nil, fmt.Errorf("duckdb open: %w", err)
	}
	if _, err := db.Exec(duckSchema); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("duckdb schema: %w", err)
	}
	return &DuckDB{db: db}, nil
}

func (d *DuckDB) RecordSpans(memberID string, spans []*model.Span) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare(`
		INSERT INTO spans (member_id, trace_id, span_id, name, start_time, end_time, duration_ms, is_error, recorded_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
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
			COUNT(*)         AS cnt,
			AVG(duration_ms) AS avg_ms,
			SUM(is_error)    AS errs
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
			ts   model.ToolStat
			errs int
		)
		if err := rows.Scan(&ts.Name, &ts.Count, &ts.AvgMs, &errs); err != nil {
			return nil, err
		}
		ts.ErrorRate = float64(errs) / float64(ts.Count)
		summary.TotalSpans += ts.Count
		summary.ErrorSpans += errs
		summary.TopTools = append(summary.TopTools, ts)
	}
	return summary, rows.Err()
}

func (d *DuckDB) Close() error {
	return d.db.Close()
}
