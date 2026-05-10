package store

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"

	"github.com/YumikoKawaii/angelix/server/migrate"
	"github.com/YumikoKawaii/angelix/server/model"
)

var migrations = []migrate.Migration{
	{Version: 1, SQL: `CREATE TABLE IF NOT EXISTS members (
		id         TEXT PRIMARY KEY,
		name       TEXT NOT NULL,
		email      TEXT NOT NULL UNIQUE,
		token      TEXT NOT NULL UNIQUE,
		api_key    TEXT NOT NULL,
		created_at DATETIME NOT NULL
	)`},
}

type SQLite struct {
	db *sql.DB
}

func NewSQLite(path string) (*SQLite, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	db.SetMaxOpenConns(1)
	if err := migrate.Run(db, migrations); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("migrations: %w", err)
	}
	return &SQLite{db: db}, nil
}

func (s *SQLite) Close() error { return s.db.Close() }

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
