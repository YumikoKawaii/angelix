package store

import (
	"database/sql"
	"fmt"
	"time"

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
		api_key    TEXT NOT NULL DEFAULT '',
		created_at DATETIME NOT NULL
	)`},
	{Version: 2, SQL: `CREATE TABLE IF NOT EXISTS credentials (
		id         TEXT PRIMARY KEY,
		name       TEXT NOT NULL,
		api_key    TEXT NOT NULL,
		created_at DATETIME NOT NULL
	)`},
	{Version: 3, SQL: `CREATE TABLE IF NOT EXISTS member_credentials (
		member_id     TEXT NOT NULL REFERENCES members(id),
		credential_id TEXT NOT NULL REFERENCES credentials(id),
		assigned_at   DATETIME NOT NULL,
		PRIMARY KEY (member_id, credential_id)
	)`},
	{Version: 4, SQL: `ALTER TABLE credentials RENAME COLUMN api_key TO access_token`},
	{Version: 5, SQL: `ALTER TABLE credentials ADD COLUMN IF NOT EXISTS refresh_token TEXT NOT NULL DEFAULT ''`},
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

// ── Members ──────────────────────────────────────────────────────────────────

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
	_, _ = s.db.Exec(`DELETE FROM member_credentials WHERE member_id = ?`, id)
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

// ── Credentials ──────────────────────────────────────────────────────────────

func (s *SQLite) CreateCredential(c *model.Credential) error {
	_, err := s.db.Exec(
		`INSERT INTO credentials (id, name, access_token, refresh_token, created_at) VALUES (?,?,?,?,?)`,
		c.ID, c.Name, c.AccessToken, c.RefreshToken, c.CreatedAt,
	)
	return err
}

func (s *SQLite) ListCredentials() ([]*model.Credential, error) {
	rows, err := s.db.Query(
		`SELECT id, name, access_token, refresh_token, created_at FROM credentials ORDER BY created_at`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var creds []*model.Credential
	for rows.Next() {
		var c model.Credential
		if err := rows.Scan(&c.ID, &c.Name, &c.AccessToken, &c.RefreshToken, &c.CreatedAt); err != nil {
			return nil, err
		}
		creds = append(creds, &c)
	}
	return creds, rows.Err()
}

func (s *SQLite) DeleteCredential(id string) error {
	_, _ = s.db.Exec(`DELETE FROM member_credentials WHERE credential_id = ?`, id)
	res, err := s.db.Exec(`DELETE FROM credentials WHERE id = ?`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("credential not found")
	}
	return nil
}

// ── Assignments ───────────────────────────────────────────────────────────────

func (s *SQLite) AssignCredential(memberID, credentialID string) error {
	_, err := s.db.Exec(
		`INSERT OR IGNORE INTO member_credentials (member_id, credential_id, assigned_at) VALUES (?,?,?)`,
		memberID, credentialID, time.Now().UTC(),
	)
	return err
}

func (s *SQLite) UnassignCredential(memberID, credentialID string) error {
	_, err := s.db.Exec(
		`DELETE FROM member_credentials WHERE member_id = ? AND credential_id = ?`,
		memberID, credentialID,
	)
	return err
}

func (s *SQLite) ListMemberCredentials(memberID string) ([]*model.Credential, error) {
	rows, err := s.db.Query(`
		SELECT c.id, c.name, c.access_token, c.refresh_token, c.created_at
		FROM credentials c
		JOIN member_credentials mc ON mc.credential_id = c.id
		WHERE mc.member_id = ?
		ORDER BY mc.assigned_at
	`, memberID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var creds []*model.Credential
	for rows.Next() {
		var c model.Credential
		if err := rows.Scan(&c.ID, &c.Name, &c.AccessToken, &c.RefreshToken, &c.CreatedAt); err != nil {
			return nil, err
		}
		creds = append(creds, &c)
	}
	return creds, rows.Err()
}

// ── Helpers ───────────────────────────────────────────────────────────────────

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
