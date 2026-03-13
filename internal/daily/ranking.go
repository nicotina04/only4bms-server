package daily

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/nicotina04/only4bms-server/internal/db"
)

// SubmitRequest is the JSON body for POST /api/daily/submit.
type SubmitRequest struct {
	PlayerName string         `json:"player_name"`
	Seed       int64          `json:"seed"`
	Score      int            `json:"score"`
	Accuracy   float64        `json:"accuracy"`
	Judgments  map[string]int `json:"judgments"`
	MaxCombo   int            `json:"max_combo"`
}

// SubmitResponse is the JSON response for POST /api/daily/submit.
type SubmitResponse struct {
	Rank         int `json:"rank"`
	TotalPlayers int `json:"total_players"`
}

// RankingResponse is the JSON response for GET /api/daily/ranking.
type RankingResponse struct {
	Period    string           `json:"period"`
	StartDate string          `json:"start_date"`
	EndDate   string          `json:"end_date"`
	Entries   []db.RankingEntry `json:"entries"`
}

var (
	ErrInvalidSubmission = errors.New("invalid submission")
	ErrSeedMismatch      = errors.New("seed does not match today's course")
)

// ValidateSubmission checks the submission for basic anti-cheat rules.
func ValidateSubmission(req SubmitRequest) error {
	if req.PlayerName == "" || len(req.PlayerName) > 32 {
		return fmt.Errorf("%w: player_name must be 1-32 characters", ErrInvalidSubmission)
	}
	if req.Score < 0 || req.Score > 1_000_000 {
		return fmt.Errorf("%w: score must be 0-1000000", ErrInvalidSubmission)
	}
	if req.Accuracy < 0 || req.Accuracy > 100 {
		return fmt.Errorf("%w: accuracy must be 0-100", ErrInvalidSubmission)
	}

	totalNotes := 0
	for _, v := range req.Judgments {
		if v < 0 {
			return fmt.Errorf("%w: negative judgment count", ErrInvalidSubmission)
		}
		totalNotes += v
	}
	if totalNotes == 0 || totalNotes > 10000 {
		return fmt.Errorf("%w: total notes must be 1-10000", ErrInvalidSubmission)
	}
	if req.MaxCombo < 0 || req.MaxCombo > totalNotes {
		return fmt.Errorf("%w: max_combo must be 0-%d", ErrInvalidSubmission, totalNotes)
	}

	return nil
}

// SubmitScore validates and stores a score submission.
func (s *Service) SubmitScore(ctx context.Context, req SubmitRequest) (*SubmitResponse, error) {
	if err := ValidateSubmission(req); err != nil {
		return nil, err
	}

	// Verify seed matches today's course
	now := time.Now()
	dateStr := s.EffectiveDate(now)
	course, err := s.store.GetDailyCourse(ctx, dateStr)
	if err != nil {
		return nil, fmt.Errorf("get daily course: %w", err)
	}
	if course == nil || course.Seed != req.Seed {
		return nil, ErrSeedMismatch
	}

	// Serialize judgments to JSON
	judgmentsJSON, err := json.Marshal(req.Judgments)
	if err != nil {
		return nil, fmt.Errorf("marshal judgments: %w", err)
	}

	// Store the score
	ranking := db.DailyRanking{
		Date:       dateStr,
		PlayerName: req.PlayerName,
		Score:      req.Score,
		Accuracy:   req.Accuracy,
		MaxCombo:   req.MaxCombo,
		Judgments:  string(judgmentsJSON),
	}
	if err := s.store.SubmitScore(ctx, ranking); err != nil {
		return nil, err
	}

	// Get rank
	rank, total, err := s.store.GetPlayerRank(ctx, dateStr, req.PlayerName)
	if err != nil {
		return nil, fmt.Errorf("get rank: %w", err)
	}

	return &SubmitResponse{Rank: rank, TotalPlayers: total}, nil
}

// GetRanking returns rankings for the specified period.
func (s *Service) GetRanking(ctx context.Context, period string, refDate time.Time) (*RankingResponse, error) {
	startDate, endDate := periodRange(period, refDate)

	var entries []db.RankingEntry
	var err error

	if period == "daily" {
		entries, err = s.store.GetDailyRanking(ctx, startDate)
	} else {
		entries, err = s.store.GetAggregatedRanking(ctx, startDate, endDate)
	}
	if err != nil {
		return nil, err
	}

	if entries == nil {
		entries = []db.RankingEntry{}
	}

	return &RankingResponse{
		Period:    period,
		StartDate: startDate,
		EndDate:   endDate,
		Entries:   entries,
	}, nil
}

// periodRange calculates the start and end dates for a ranking period.
func periodRange(period string, ref time.Time) (start, end string) {
	ref = ref.UTC()
	switch period {
	case "weekly":
		// Find Monday of the week
		weekday := ref.Weekday()
		if weekday == time.Sunday {
			weekday = 7
		}
		monday := ref.AddDate(0, 0, -int(weekday-time.Monday))
		sunday := monday.AddDate(0, 0, 6)
		return monday.Format("2006-01-02"), sunday.Format("2006-01-02")
	case "monthly":
		first := time.Date(ref.Year(), ref.Month(), 1, 0, 0, 0, 0, time.UTC)
		last := first.AddDate(0, 1, -1)
		return first.Format("2006-01-02"), last.Format("2006-01-02")
	default: // daily
		d := ref.Format("2006-01-02")
		return d, d
	}
}
