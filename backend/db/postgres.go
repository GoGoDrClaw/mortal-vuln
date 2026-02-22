package db

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/lib/pq"
)

var PG *sql.DB

func InitPostgres() error {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			envOr("PGHOST", "postgres"),
			envOr("PGPORT", "5432"),
			envOr("PGUSER", "ctf"),
			envOr("PGPASSWORD", "ctf"),
			envOr("PGDATABASE", "ctf"),
		)
	}

	d, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("open: %w", err)
	}
	if err := d.Ping(); err != nil {
		return fmt.Errorf("ping postgres: %w", err)
	}
	PG = d
	log.Println("✅ PostgreSQL connected")
	return initSchema()
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func initSchema() error {
	_, err := PG.Exec(`
		CREATE TABLE IF NOT EXISTS sessions (
			id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			save_code    CHAR(8) UNIQUE NOT NULL,
			nickname     VARCHAR(50) NOT NULL,
			character    VARCHAR(20) NOT NULL,
			created_at   TIMESTAMPTZ DEFAULT NOW(),
			last_seen_at TIMESTAMPTZ DEFAULT NOW()
		);

		CREATE TABLE IF NOT EXISTS task_completions (
			id           SERIAL PRIMARY KEY,
			session_id   UUID NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
			task_id      INTEGER NOT NULL,
			task_name    VARCHAR(100) NOT NULL,
			points       INTEGER NOT NULL,
			details      TEXT DEFAULT '',
			completed_at TIMESTAMPTZ DEFAULT NOW(),
			CONSTRAINT uq_session_task UNIQUE(session_id, task_id)
		);

		CREATE INDEX IF NOT EXISTS idx_tc_session    ON task_completions(session_id);
		CREATE INDEX IF NOT EXISTS idx_tc_task_id    ON task_completions(task_id);
		CREATE INDEX IF NOT EXISTS idx_sessions_code ON sessions(save_code);
	`)
	if err != nil {
		return fmt.Errorf("schema: %w", err)
	}
	// Migrations — safe to run repeatedly
	for _, q := range []string{
		`ALTER TABLE task_completions ADD COLUMN IF NOT EXISTS first_blood BOOLEAN DEFAULT FALSE`,
		`ALTER TABLE task_completions ADD COLUMN IF NOT EXISTS combo_bonus INTEGER DEFAULT 0`,
		`CREATE UNIQUE INDEX IF NOT EXISTS uq_sessions_nickname_character ON sessions(lower(nickname), character)`,
	} {
		if _, e := PG.Exec(q); e != nil {
			return fmt.Errorf("migration: %w", e)
		}
	}
	log.Println("✅ PostgreSQL schema ready")
	return nil
}

// codeAlphabet excludes visually ambiguous chars: 0/O, 1/I/L, U, 5/S, A
const codeAlphabet = "BCDFGHJKMNPQRTVWXYZ2346789"

// GenerateSaveCode returns an 8-character random code from codeAlphabet.
func GenerateSaveCode() (string, error) {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	n := byte(len(codeAlphabet))
	out := make([]byte, 8)
	for i, v := range b {
		out[i] = codeAlphabet[v%n]
	}
	return string(out), nil
}

// Session holds player session metadata.
type Session struct {
	ID        string
	SaveCode  string
	Nickname  string
	Character string
}

// DisplayCode formats the save code as XXXX-XXXX.
func (s *Session) DisplayCode() string {
	if len(s.SaveCode) == 8 {
		return s.SaveCode[:4] + "-" + s.SaveCode[4:]
	}
	return s.SaveCode
}

// ErrDuplicateNicknameChar is returned when the (nickname, character) pair already exists.
var ErrDuplicateNicknameChar = fmt.Errorf("nickname_character_taken")

// CreateSession inserts a new session, retrying on save_code collisions.
func CreateSession(nickname, character string) (*Session, error) {
	for i := 0; i < 5; i++ {
		code, err := GenerateSaveCode()
		if err != nil {
			return nil, err
		}
		var s Session
		err = PG.QueryRow(`
			INSERT INTO sessions (save_code, nickname, character)
			VALUES ($1, $2, $3)
			RETURNING id, save_code, nickname, character
		`, code, nickname, character).Scan(&s.ID, &s.SaveCode, &s.Nickname, &s.Character)
		if err == nil {
			return &s, nil
		}
		// Unique violation on nickname+character — not a retry-able error
		if strings.Contains(err.Error(), "uq_sessions_nickname_character") {
			return nil, ErrDuplicateNicknameChar
		}
		// Unique violation on save_code — retry with new code
		if strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "duplicate") {
			continue
		}
		return nil, err
	}
	return nil, fmt.Errorf("failed to generate unique save code")
}

// GetSessionByCode looks up a session by code (normalises: uppercase, strips hyphens).
func GetSessionByCode(code string) (*Session, error) {
	normalized := strings.ToUpper(strings.ReplaceAll(code, "-", ""))
	var s Session
	err := PG.QueryRow(`
		UPDATE sessions SET last_seen_at = NOW()
		WHERE save_code = $1
		RETURNING id, save_code, nickname, character
	`, normalized).Scan(&s.ID, &s.SaveCode, &s.Nickname, &s.Character)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// GetSessionByID loads a session by UUID.
func GetSessionByID(id string) (*Session, error) {
	var s Session
	err := PG.QueryRow(`
		SELECT id, save_code, nickname, character FROM sessions WHERE id = $1
	`, id).Scan(&s.ID, &s.SaveCode, &s.Nickname, &s.Character)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}
