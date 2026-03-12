package db

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

// Open opens a SQLite database with WAL mode and recommended pragmas.
func Open(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	// SQLite pragmas for performance and correctness
	pragmas := []string{
		"PRAGMA journal_mode=WAL",
		"PRAGMA busy_timeout=5000",
		"PRAGMA foreign_keys=ON",
	}
	for _, p := range pragmas {
		if _, err := db.Exec(p); err != nil {
			db.Close()
			return nil, fmt.Errorf("exec pragma %q: %w", p, err)
		}
	}

	// Single writer connection for SQLite
	db.SetMaxOpenConns(1)

	return db, nil
}

// Migrate creates the required tables if they do not exist.
func Migrate(db *sql.DB) error {
	const schema = `
	CREATE TABLE IF NOT EXISTS daily_courses (
		date       TEXT PRIMARY KEY,
		seed       INTEGER NOT NULL,
		difficulty TEXT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS daily_rankings (
		id           INTEGER PRIMARY KEY AUTOINCREMENT,
		date         TEXT NOT NULL,
		player_name  TEXT NOT NULL,
		score        INTEGER NOT NULL,
		accuracy     REAL NOT NULL,
		max_combo    INTEGER NOT NULL,
		judgments    TEXT NOT NULL,
		submitted_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(date, player_name)
	);

	CREATE INDEX IF NOT EXISTS idx_daily_rank ON daily_rankings(date, score DESC);
	CREATE INDEX IF NOT EXISTS idx_weekly ON daily_rankings(player_name, date);
	`

	if _, err := db.Exec(schema); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	return nil
}
