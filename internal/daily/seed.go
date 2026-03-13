package daily

import "time"

var difficulties = [4]string{"BEGINNER", "INTERMEDIATE", "ADVANCED", "ORDEAL"}

// SeedForDate returns a deterministic seed for the given date.
func SeedForDate(date time.Time) int64 {
	y, m, d := date.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC).Unix()
}

// DifficultyForDate returns the rotating difficulty for the given date.
func DifficultyForDate(date time.Time) string {
	return difficulties[date.YearDay()%4]
}
