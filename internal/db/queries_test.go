package db

import (
	"context"
	"strings"
	"testing"
)

func setupTestStore(t *testing.T) *Store {
	t.Helper()
	database, err := Open(":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { database.Close() })

	if err := Migrate(database); err != nil {
		t.Fatalf("Migrate: %v", err)
	}
	return NewStore(database)
}

func TestUpsertAndGetDailyCourse(t *testing.T) {
	store := setupTestStore(t)
	ctx := context.Background()

	course := DailyCourse{Date: "2026-03-13", Seed: 1773532800, Difficulty: "INTERMEDIATE"}
	if err := store.UpsertDailyCourse(ctx, course); err != nil {
		t.Fatalf("UpsertDailyCourse: %v", err)
	}

	got, err := store.GetDailyCourse(ctx, "2026-03-13")
	if err != nil {
		t.Fatalf("GetDailyCourse: %v", err)
	}
	if got == nil {
		t.Fatal("expected course, got nil")
	}
	if got.Seed != 1773532800 || got.Difficulty != "INTERMEDIATE" {
		t.Errorf("unexpected course: %+v", got)
	}
}

func TestGetDailyCourseNotFound(t *testing.T) {
	store := setupTestStore(t)
	ctx := context.Background()

	got, err := store.GetDailyCourse(ctx, "2099-01-01")
	if err != nil {
		t.Fatalf("GetDailyCourse: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil, got %+v", got)
	}
}

func TestSubmitScoreAndRanking(t *testing.T) {
	store := setupTestStore(t)
	ctx := context.Background()

	date := "2026-03-13"

	// Submit three scores
	scores := []DailyRanking{
		{Date: date, PlayerName: "alice", Score: 900000, Accuracy: 95.0, MaxCombo: 200, Judgments: `{"PERFECT":190,"GREAT":10}`},
		{Date: date, PlayerName: "bob", Score: 950000, Accuracy: 98.0, MaxCombo: 250, Judgments: `{"PERFECT":245,"GREAT":5}`},
		{Date: date, PlayerName: "carol", Score: 800000, Accuracy: 90.0, MaxCombo: 150, Judgments: `{"PERFECT":140,"GREAT":10}`},
	}
	for _, s := range scores {
		if err := store.SubmitScore(ctx, s); err != nil {
			t.Fatalf("SubmitScore(%s): %v", s.PlayerName, err)
		}
	}

	// Check daily ranking order
	entries, err := store.GetDailyRanking(ctx, date)
	if err != nil {
		t.Fatalf("GetDailyRanking: %v", err)
	}
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}
	if entries[0].PlayerName != "bob" || entries[1].PlayerName != "alice" || entries[2].PlayerName != "carol" {
		t.Errorf("unexpected ranking order: %v", entries)
	}

	// Check rank for alice
	rank, total, err := store.GetPlayerRank(ctx, date, "alice")
	if err != nil {
		t.Fatalf("GetPlayerRank: %v", err)
	}
	if rank != 2 || total != 3 {
		t.Errorf("expected rank 2 of 3, got rank %d of %d", rank, total)
	}
}

func TestDuplicateSubmission(t *testing.T) {
	store := setupTestStore(t)
	ctx := context.Background()

	r := DailyRanking{Date: "2026-03-13", PlayerName: "alice", Score: 900000, Accuracy: 95.0, MaxCombo: 200, Judgments: "{}"}
	if err := store.SubmitScore(ctx, r); err != nil {
		t.Fatalf("first submit: %v", err)
	}
	err := store.SubmitScore(ctx, r)
	if err == nil {
		t.Fatal("expected error on duplicate submission")
	}
	if !strings.Contains(err.Error(), "UNIQUE constraint") {
		t.Errorf("expected UNIQUE constraint error, got: %v", err)
	}
}

func TestAggregatedRanking(t *testing.T) {
	store := setupTestStore(t)
	ctx := context.Background()

	// Submit scores across multiple days
	entries := []DailyRanking{
		{Date: "2026-03-10", PlayerName: "alice", Score: 900000, Accuracy: 95.0, MaxCombo: 200, Judgments: "{}"},
		{Date: "2026-03-11", PlayerName: "alice", Score: 910000, Accuracy: 96.0, MaxCombo: 210, Judgments: "{}"},
		{Date: "2026-03-10", PlayerName: "bob", Score: 950000, Accuracy: 98.0, MaxCombo: 250, Judgments: "{}"},
	}
	for _, e := range entries {
		if err := store.SubmitScore(ctx, e); err != nil {
			t.Fatalf("SubmitScore: %v", err)
		}
	}

	ranking, err := store.GetAggregatedRanking(ctx, "2026-03-10", "2026-03-16")
	if err != nil {
		t.Fatalf("GetAggregatedRanking: %v", err)
	}
	if len(ranking) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(ranking))
	}
	// Alice has higher total (900000 + 910000 = 1810000) vs Bob (950000)
	if ranking[0].PlayerName != "alice" {
		t.Errorf("expected alice first in weekly, got %s", ranking[0].PlayerName)
	}
	if ranking[0].PlayCount != 2 {
		t.Errorf("expected play_count 2, got %d", ranking[0].PlayCount)
	}
}

func TestGetRecentCourses(t *testing.T) {
	store := setupTestStore(t)
	ctx := context.Background()

	courses := []DailyCourse{
		{Date: "2026-03-11", Seed: 100, Difficulty: "BEGINNER"},
		{Date: "2026-03-12", Seed: 200, Difficulty: "INTERMEDIATE"},
		{Date: "2026-03-13", Seed: 300, Difficulty: "ADVANCED"},
	}
	for _, c := range courses {
		if err := store.UpsertDailyCourse(ctx, c); err != nil {
			t.Fatalf("UpsertDailyCourse: %v", err)
		}
	}

	got, err := store.GetRecentCourses(ctx, 2)
	if err != nil {
		t.Fatalf("GetRecentCourses: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2, got %d", len(got))
	}
	if got[0].Date != "2026-03-13" || got[1].Date != "2026-03-12" {
		t.Errorf("unexpected order: %v", got)
	}
}
