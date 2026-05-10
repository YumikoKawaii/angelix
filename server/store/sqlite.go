package store

import (
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"

	"github.com/YumikoKawaii/angelix/server/model"
)

const schema = `
CREATE TABLE IF NOT EXISTS members (
    id         TEXT PRIMARY KEY,
    name       TEXT NOT NULL,
    email      TEXT NOT NULL UNIQUE,
    token      TEXT NOT NULL UNIQUE,
    api_key    TEXT NOT NULL,
    created_at DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS spans (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    member_id   TEXT    NOT NULL REFERENCES members(id) ON DELETE CASCADE,
    trace_id    TEXT    NOT NULL,
    span_id     TEXT    NOT NULL,
    name        TEXT    NOT NULL,
    start_time  DATETIME NOT NULL,
    end_time    DATETIME NOT NULL,
    duration_ms INTEGER NOT NULL,
    is_error    INTEGER NOT NULL DEFAULT 0,
    recorded_at DATETIME NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_spans_member ON spans(member_id);
CREATE INDEX IF NOT EXISTS idx_spans_name   ON spans(member_id, name);
`

type SQLite struct {
	db *sql.DB
}

func NewSQLite(path string) (*SQLite, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	db.SetMaxOpenConns(1) // SQLite is single-writer
	if _, err := db.Exec(schema); err != nil {
		return nil, fmt.Errorf("run schema: %w", err)
	}
	return &SQLite{db: db}, nil
}

func (s *SQLite) Close() error { return s.db.Close() }

// Member management

func (s *SQLite) CreateMember(m *model.Member) error {
	_, err := s.db.Exec(
		`INSERT INTO members (id, name, email, token, api_key, created_at) VALUES (?,?,?,?,?,?)`,
		m.ID, m.Name, m.Email, m.Token, m.APIKey, m.CreatedAt,
	)
	return err
}

func (s *SQLite) GetMemberByToken(token string) (*model.Member, error) {
	return s.scanMember(s.db.QueryRow(
		`SELECT id, name, email, token, api_key, created_at FROM members WHERE token = ?`, token,
	))
}

func (s *SQLite) GetMemberByID(id string) (*model.Member, error) {
	return s.scanMember(s.db.QueryRow(
		`SELECT id, name, email, token, api_key, created_at FROM members WHERE id = ?`, id,
	))
}

func (s *SQLite) ListMembers() ([]*model.Member, error) {
	rows, err := s.db.Query(
		`SELECT id, name, email, token, api_key, created_at FROM members ORDER BY created_at`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []*model.Member
	for rows.Next() {
		m, err := s.scanMember(rows)
		if err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	return members, rows.Err()
}

func (s *SQLite) DeleteMember(id string) error {
	res, err := s.db.Exec(`DELETE FROM members WHERE id = ?`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("member not found")
	}
	return nil
}

type scanner interface {
	Scan(dest ...any) error
}

func (s *SQLite) scanMember(row scanner) (*model.Member, error) {
	var m model.Member
	err := row.Scan(&m.ID, &m.Name, &m.Email, &m.Token, &m.APIKey, &m.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("member not found")
	}
	return &m, err
}

// Telemetry

func (s *SQLite) RecordSpans(memberID string, spans []*model.Span) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare(
		`INSERT INTO spans (member_id, trace_id, span_id, name, start_time, end_time, duration_ms, is_error, recorded_at)
		 VALUES (?,?,?,?,?,?,?,?,?)`,
	)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	defer stmt.Close()

	for _, sp := range spans {
		isErr := 0
		if sp.IsError {
			isErr = 1
		}
		if _, err := stmt.Exec(
			memberID, sp.TraceID, sp.SpanID, sp.Name,
			sp.StartTime, sp.EndTime, sp.DurationMs, isErr,
			time.Now().UTC(),
		); err != nil {
			_ = tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

func (s *SQLite) GetSpanSummary(memberID string) (*model.SpanSummary, error) {
	rows, err := s.db.Query(
		`SELECT name, COUNT(*) AS cnt,
		        AVG(duration_ms) AS avg_ms,
		        SUM(is_error) AS errs
		 FROM spans WHERE member_id = ?
		 GROUP BY name
		 ORDER BY cnt DESC`,
		memberID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	summary := &model.SpanSummary{MemberID: memberID}
	for rows.Next() {
		var ts model.ToolStat
		var errs int
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
