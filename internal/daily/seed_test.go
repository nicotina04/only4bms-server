package daily

import (
	"testing"
	"time"
)

func TestSeedForDateDeterministic(t *testing.T) {
	date := time.Date(2026, 3, 13, 0, 0, 0, 0, time.UTC)
	seed1 := SeedForDate(date)
	seed2 := SeedForDate(date)
	if seed1 != seed2 {
		t.Errorf("seed not deterministic: %d != %d", seed1, seed2)
	}

	// Same date with different time should produce same seed
	dateWithTime := time.Date(2026, 3, 13, 15, 30, 0, 0, time.UTC)
	seed3 := SeedForDate(dateWithTime)
	if seed1 != seed3 {
		t.Errorf("seed should ignore time: %d != %d", seed1, seed3)
	}
}

func TestSeedForDateDifferentDays(t *testing.T) {
	day1 := time.Date(2026, 3, 13, 0, 0, 0, 0, time.UTC)
	day2 := time.Date(2026, 3, 14, 0, 0, 0, 0, time.UTC)
	if SeedForDate(day1) == SeedForDate(day2) {
		t.Error("different days should produce different seeds")
	}
}

func TestSeedForDateKnownValue(t *testing.T) {
	date := time.Date(2026, 3, 9, 0, 0, 0, 0, time.UTC)
	expected := int64(1773014400) // 2026-03-09 00:00:00 UTC Unix
	got := SeedForDate(date)
	if got != expected {
		t.Errorf("expected seed %d, got %d", expected, got)
	}
}

func TestDifficultyForDateRotation(t *testing.T) {
	// Check that 4 consecutive days cycle through all difficulties
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	seen := map[string]bool{}
	for i := 0; i < 4; i++ {
		d := base.AddDate(0, 0, i)
		diff := DifficultyForDate(d)
		seen[diff] = true
	}
	if len(seen) != 4 {
		t.Errorf("expected 4 unique difficulties, got %d: %v", len(seen), seen)
	}
}

func TestDifficultyForDateConsistency(t *testing.T) {
	date := time.Date(2026, 3, 13, 0, 0, 0, 0, time.UTC)
	d1 := DifficultyForDate(date)
	d2 := DifficultyForDate(date)
	if d1 != d2 {
		t.Errorf("difficulty not consistent: %s != %s", d1, d2)
	}
}
