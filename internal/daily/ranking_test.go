package daily

import (
	"testing"
	"time"
)

func TestValidateSubmission(t *testing.T) {
	valid := SubmitRequest{
		PlayerName: "alice",
		Seed:       1773532800,
		Score:      900000,
		Accuracy:   95.0,
		Judgments:  map[string]int{"PERFECT": 190, "GREAT": 10, "GOOD": 0, "MISS": 0},
		MaxCombo:   200,
	}

	if err := ValidateSubmission(valid); err != nil {
		t.Errorf("valid submission rejected: %v", err)
	}

	tests := []struct {
		name   string
		modify func(SubmitRequest) SubmitRequest
	}{
		{"empty player name", func(r SubmitRequest) SubmitRequest { r.PlayerName = ""; return r }},
		{"player name too long", func(r SubmitRequest) SubmitRequest { r.PlayerName = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"; return r }},
		{"score too high", func(r SubmitRequest) SubmitRequest { r.Score = 1_000_001; return r }},
		{"negative score", func(r SubmitRequest) SubmitRequest { r.Score = -1; return r }},
		{"accuracy too high", func(r SubmitRequest) SubmitRequest { r.Accuracy = 100.1; return r }},
		{"negative accuracy", func(r SubmitRequest) SubmitRequest { r.Accuracy = -1; return r }},
		{"zero notes", func(r SubmitRequest) SubmitRequest { r.Judgments = map[string]int{}; return r }},
		{"too many notes", func(r SubmitRequest) SubmitRequest {
			r.Judgments = map[string]int{"PERFECT": 10001}
			return r
		}},
		{"negative judgment", func(r SubmitRequest) SubmitRequest {
			r.Judgments = map[string]int{"PERFECT": -1}
			return r
		}},
		{"combo exceeds notes", func(r SubmitRequest) SubmitRequest { r.MaxCombo = 201; return r }},
		{"negative combo", func(r SubmitRequest) SubmitRequest { r.MaxCombo = -1; return r }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := tt.modify(valid)
			if err := ValidateSubmission(req); err == nil {
				t.Error("expected validation error")
			}
		})
	}
}

func TestPeriodRange(t *testing.T) {
	tests := []struct {
		name      string
		period    string
		ref       time.Time
		wantStart string
		wantEnd   string
	}{
		{
			name:      "daily",
			period:    "daily",
			ref:       time.Date(2026, 3, 13, 0, 0, 0, 0, time.UTC),
			wantStart: "2026-03-13",
			wantEnd:   "2026-03-13",
		},
		{
			name:      "weekly - Wednesday",
			period:    "weekly",
			ref:       time.Date(2026, 3, 11, 0, 0, 0, 0, time.UTC), // Wednesday
			wantStart: "2026-03-09",                                   // Monday
			wantEnd:   "2026-03-15",                                   // Sunday
		},
		{
			name:      "weekly - Monday",
			period:    "weekly",
			ref:       time.Date(2026, 3, 9, 0, 0, 0, 0, time.UTC), // Monday
			wantStart: "2026-03-09",
			wantEnd:   "2026-03-15",
		},
		{
			name:      "weekly - Sunday",
			period:    "weekly",
			ref:       time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC), // Sunday
			wantStart: "2026-03-09",
			wantEnd:   "2026-03-15",
		},
		{
			name:      "monthly",
			period:    "monthly",
			ref:       time.Date(2026, 3, 13, 0, 0, 0, 0, time.UTC),
			wantStart: "2026-03-01",
			wantEnd:   "2026-03-31",
		},
		{
			name:      "monthly - February",
			period:    "monthly",
			ref:       time.Date(2026, 2, 15, 0, 0, 0, 0, time.UTC),
			wantStart: "2026-02-01",
			wantEnd:   "2026-02-28",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end := periodRange(tt.period, tt.ref)
			if start != tt.wantStart {
				t.Errorf("start: got %s, want %s", start, tt.wantStart)
			}
			if end != tt.wantEnd {
				t.Errorf("end: got %s, want %s", end, tt.wantEnd)
			}
		})
	}
}
