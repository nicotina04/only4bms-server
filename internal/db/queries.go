package db

import (
	"context"
	"database/sql"
	"time"
)

// DailyCourse represents a row in the daily_courses table.
type DailyCourse struct {
	Date       string
	Seed       int64
	Difficulty string
}

// DailyRanking represents a row in the daily_rankings table.
type DailyRanking struct {
	ID          int64
	Date        string
	PlayerName  string
	Score       int
	Accuracy    float64
	MaxCombo    int
	Judgments   string
	SubmittedAt time.Time
}

// RankingEntry represents an aggregated ranking row (weekly/monthly).
type RankingEntry struct {
	Rank        int     `json:"rank"`
	PlayerName  string  `json:"player_name"`
	TotalScore  int64   `json:"total_score"`
	PlayCount   int     `json:"play_count"`
	AvgAccuracy float64 `json:"avg_accuracy"`
}

// Store wraps a sql.DB and provides query methods.
type Store struct {
	db *sql.DB
}

// NewStore creates a new Store.
func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

// UpsertDailyCourse inserts or replaces a daily course.
func (s *Store) UpsertDailyCourse(ctx context.Context, course DailyCourse) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT OR REPLACE INTO daily_courses (date, seed, difficulty) VALUES (?, ?, ?)`,
		course.Date, course.Seed, course.Difficulty,
	)
	return err
}

// GetDailyCourse retrieves a daily course by date.
func (s *Store) GetDailyCourse(ctx context.Context, date string) (*DailyCourse, error) {
	var c DailyCourse
	err := s.db.QueryRowContext(ctx,
		`SELECT date, seed, difficulty FROM daily_courses WHERE date = ?`, date,
	).Scan(&c.Date, &c.Seed, &c.Difficulty)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// SubmitScore inserts a daily ranking entry.
// Returns an error if the player already submitted for that date (UNIQUE constraint).
func (s *Store) SubmitScore(ctx context.Context, r DailyRanking) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO daily_rankings (date, player_name, score, accuracy, max_combo, judgments)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		r.Date, r.PlayerName, r.Score, r.Accuracy, r.MaxCombo, r.Judgments,
	)
	return err
}

// GetPlayerRank returns the rank and total players for a given date and player.
func (s *Store) GetPlayerRank(ctx context.Context, date, playerName string) (rank int, totalPlayers int, err error) {
	err = s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM daily_rankings WHERE date = ?`, date,
	).Scan(&totalPlayers)
	if err != nil {
		return 0, 0, err
	}

	// Rank = number of players with a higher score + 1
	err = s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) + 1 FROM daily_rankings
		 WHERE date = ? AND score > (SELECT score FROM daily_rankings WHERE date = ? AND player_name = ?)`,
		date, date, playerName,
	).Scan(&rank)
	return rank, totalPlayers, err
}

// GetDailyRanking returns all rankings for a specific date, ordered by score descending.
func (s *Store) GetDailyRanking(ctx context.Context, date string) ([]RankingEntry, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT player_name, score, 1 as play_count, accuracy
		 FROM daily_rankings WHERE date = ? ORDER BY score DESC`, date,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []RankingEntry
	rank := 1
	for rows.Next() {
		var e RankingEntry
		if err := rows.Scan(&e.PlayerName, &e.TotalScore, &e.PlayCount, &e.AvgAccuracy); err != nil {
			return nil, err
		}
		e.Rank = rank
		rank++
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

// GetAggregatedRanking returns aggregated rankings for a date range (weekly/monthly).
func (s *Store) GetAggregatedRanking(ctx context.Context, startDate, endDate string) ([]RankingEntry, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT player_name, SUM(score) as total_score, COUNT(*) as play_count,
		        AVG(accuracy) as avg_accuracy
		 FROM daily_rankings
		 WHERE date BETWEEN ? AND ?
		 GROUP BY player_name
		 ORDER BY total_score DESC`,
		startDate, endDate,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []RankingEntry
	rank := 1
	for rows.Next() {
		var e RankingEntry
		if err := rows.Scan(&e.PlayerName, &e.TotalScore, &e.PlayCount, &e.AvgAccuracy); err != nil {
			return nil, err
		}
		e.Rank = rank
		rank++
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

// GetRecentCourses returns the most recent daily courses.
func (s *Store) GetRecentCourses(ctx context.Context, limit int) ([]DailyCourse, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT date, seed, difficulty FROM daily_courses ORDER BY date DESC LIMIT ?`, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var courses []DailyCourse
	for rows.Next() {
		var c DailyCourse
		if err := rows.Scan(&c.Date, &c.Seed, &c.Difficulty); err != nil {
			return nil, err
		}
		courses = append(courses, c)
	}
	return courses, rows.Err()
}
