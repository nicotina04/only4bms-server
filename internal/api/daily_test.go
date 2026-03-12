package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nicotina04/only4bms-server/internal/daily"
	"github.com/nicotina04/only4bms-server/internal/db"
)

func setupDailyTestHandler(t *testing.T) (*DailyHandler, *http.ServeMux) {
	t.Helper()

	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { database.Close() })

	if err := db.Migrate(database); err != nil {
		t.Fatalf("Migrate: %v", err)
	}

	store := db.NewStore(database)
	service := daily.NewService(store, 0)
	handler := NewDailyHandler(service)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)
	return handler, mux
}

func TestGetToday(t *testing.T) {
	_, mux := setupDailyTestHandler(t)

	req := httptest.NewRequest("GET", "/api/daily/today", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp daily.TodayCourseResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Date == "" || resp.Seed == 0 {
		t.Errorf("unexpected response: %+v", resp)
	}
}

func TestGetTodayIdempotent(t *testing.T) {
	_, mux := setupDailyTestHandler(t)

	// Call twice - should return same result
	var first, second daily.TodayCourseResponse
	for i, target := range []*daily.TodayCourseResponse{&first, &second} {
		req := httptest.NewRequest("GET", "/api/daily/today", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("call %d: expected 200, got %d", i+1, w.Code)
		}
		if err := json.NewDecoder(w.Body).Decode(target); err != nil {
			t.Fatalf("call %d: decode: %v", i+1, err)
		}
	}

	if first.Seed != second.Seed || first.Date != second.Date {
		t.Errorf("responses differ: %+v vs %+v", first, second)
	}
}

func TestSubmitScoreAndRanking(t *testing.T) {
	_, mux := setupDailyTestHandler(t)

	// First, get today's course to know the seed
	req := httptest.NewRequest("GET", "/api/daily/today", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	var course daily.TodayCourseResponse
	json.NewDecoder(w.Body).Decode(&course)

	// Submit a score
	submission := daily.SubmitRequest{
		PlayerName: "test_player",
		Seed:       course.Seed,
		Score:      900000,
		Accuracy:   95.5,
		Judgments:  map[string]int{"PERFECT": 190, "GREAT": 8, "GOOD": 2, "MISS": 0},
		MaxCombo:   198,
	}
	body, _ := json.Marshal(submission)
	req = httptest.NewRequest("POST", "/api/daily/submit", bytes.NewReader(body))
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("submit: expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var result daily.SubmitResponse
	json.NewDecoder(w.Body).Decode(&result)
	if result.Rank != 1 || result.TotalPlayers != 1 {
		t.Errorf("expected rank 1 of 1, got rank %d of %d", result.Rank, result.TotalPlayers)
	}

	// Check ranking
	req = httptest.NewRequest("GET", "/api/daily/ranking", nil)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("ranking: expected 200, got %d", w.Code)
	}

	var ranking daily.RankingResponse
	json.NewDecoder(w.Body).Decode(&ranking)
	if len(ranking.Entries) != 1 {
		t.Errorf("expected 1 entry, got %d", len(ranking.Entries))
	}
}

func TestSubmitDuplicate(t *testing.T) {
	_, mux := setupDailyTestHandler(t)

	// Get today's seed
	req := httptest.NewRequest("GET", "/api/daily/today", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	var course daily.TodayCourseResponse
	json.NewDecoder(w.Body).Decode(&course)

	submission := daily.SubmitRequest{
		PlayerName: "test_player",
		Seed:       course.Seed,
		Score:      900000,
		Accuracy:   95.0,
		Judgments:  map[string]int{"PERFECT": 190, "GREAT": 10},
		MaxCombo:   200,
	}
	body, _ := json.Marshal(submission)

	// First submission
	req = httptest.NewRequest("POST", "/api/daily/submit", bytes.NewReader(body))
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("first submit: expected 200, got %d: %s", w.Code, w.Body.String())
	}

	// Duplicate submission
	req = httptest.NewRequest("POST", "/api/daily/submit", bytes.NewReader(body))
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusConflict {
		t.Errorf("duplicate submit: expected 409, got %d: %s", w.Code, w.Body.String())
	}
}

func TestSubmitInvalidScore(t *testing.T) {
	_, mux := setupDailyTestHandler(t)

	submission := daily.SubmitRequest{
		PlayerName: "test_player",
		Seed:       12345,
		Score:      2_000_000, // exceeds max
		Accuracy:   95.0,
		Judgments:  map[string]int{"PERFECT": 190, "GREAT": 10},
		MaxCombo:   200,
	}
	body, _ := json.Marshal(submission)

	req := httptest.NewRequest("POST", "/api/daily/submit", bytes.NewReader(body))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestGetRankingInvalidPeriod(t *testing.T) {
	_, mux := setupDailyTestHandler(t)

	req := httptest.NewRequest("GET", "/api/daily/ranking?period=yearly", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestGetHistory(t *testing.T) {
	_, mux := setupDailyTestHandler(t)

	// Create a course first
	req := httptest.NewRequest("GET", "/api/daily/today", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	// Get history
	req = httptest.NewRequest("GET", "/api/daily/history", nil)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestGetHistoryInvalidDays(t *testing.T) {
	_, mux := setupDailyTestHandler(t)

	req := httptest.NewRequest("GET", "/api/daily/history?days=50", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}
