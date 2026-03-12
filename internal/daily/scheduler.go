package daily

import (
	"context"
	"fmt"
	"time"

	"github.com/nicotina04/only4bms-server/internal/db"
)

// TodayCourseResponse is the JSON response for GET /api/daily/today.
type TodayCourseResponse struct {
	Date       string `json:"date"`
	Seed       int64  `json:"seed"`
	Difficulty string `json:"difficulty"`
}

// Service manages daily course lifecycle.
type Service struct {
	store     *db.Store
	resetHour int
}

// NewService creates a new daily course service.
func NewService(store *db.Store, resetHour int) *Service {
	return &Service{store: store, resetHour: resetHour}
}

// EffectiveDate returns the effective course date, accounting for resetHour.
// If current UTC hour < resetHour, the effective date is yesterday.
func (s *Service) EffectiveDate(now time.Time) string {
	utc := now.UTC()
	if utc.Hour() < s.resetHour {
		utc = utc.AddDate(0, 0, -1)
	}
	return utc.Format("2006-01-02")
}

// effectiveTime returns the effective date as a time.Time (midnight UTC).
func (s *Service) effectiveTime(now time.Time) time.Time {
	utc := now.UTC()
	if utc.Hour() < s.resetHour {
		utc = utc.AddDate(0, 0, -1)
	}
	y, m, d := utc.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

// GetTodayCourse returns today's course, creating it in DB if it doesn't exist.
func (s *Service) GetTodayCourse(ctx context.Context) (*TodayCourseResponse, error) {
	now := time.Now()
	dateStr := s.EffectiveDate(now)
	dateTime := s.effectiveTime(now)

	// Check if course already exists
	course, err := s.store.GetDailyCourse(ctx, dateStr)
	if err != nil {
		return nil, fmt.Errorf("get daily course: %w", err)
	}

	if course == nil {
		// Create today's course
		course = &db.DailyCourse{
			Date:       dateStr,
			Seed:       SeedForDate(dateTime),
			Difficulty: DifficultyForDate(dateTime),
		}
		if err := s.store.UpsertDailyCourse(ctx, *course); err != nil {
			return nil, fmt.Errorf("upsert daily course: %w", err)
		}
	}

	return &TodayCourseResponse{
		Date:       course.Date,
		Seed:       course.Seed,
		Difficulty: course.Difficulty,
	}, nil
}

// GetHistory returns recent daily courses.
func (s *Service) GetHistory(ctx context.Context, days int) ([]db.DailyCourse, error) {
	return s.store.GetRecentCourses(ctx, days)
}
