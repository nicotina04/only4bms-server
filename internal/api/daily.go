package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/nicotina04/only4bms-server/internal/daily"
)

// DailyHandler serves daily course HTTP endpoints.
type DailyHandler struct {
	service *daily.Service
}

// NewDailyHandler creates a new DailyHandler.
func NewDailyHandler(service *daily.Service) *DailyHandler {
	return &DailyHandler{service: service}
}

// RegisterRoutes registers daily API routes on the given mux.
func (h *DailyHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/daily/today", h.GetToday)
	mux.HandleFunc("GET /api/daily/ranking", h.GetRanking)
	mux.HandleFunc("POST /api/daily/submit", h.SubmitScore)
	mux.HandleFunc("GET /api/daily/history", h.GetHistory)
}

// GetToday returns today's daily course info.
func (h *DailyHandler) GetToday(w http.ResponseWriter, r *http.Request) {
	course, err := h.service.GetTodayCourse(r.Context())
	if err != nil {
		http.Error(w, "failed to get today's course", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, course)
}

// GetRanking returns rankings for the specified period.
func (h *DailyHandler) GetRanking(w http.ResponseWriter, r *http.Request) {
	period := r.URL.Query().Get("period")
	if period == "" {
		period = "daily"
	}
	if period != "daily" && period != "weekly" && period != "monthly" {
		http.Error(w, "invalid period: must be daily, weekly, or monthly", http.StatusBadRequest)
		return
	}

	refDate := time.Now().UTC()
	if dateStr := r.URL.Query().Get("date"); dateStr != "" {
		parsed, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			http.Error(w, "invalid date format, use YYYY-MM-DD", http.StatusBadRequest)
			return
		}
		refDate = parsed
	}

	ranking, err := h.service.GetRanking(r.Context(), period, refDate)
	if err != nil {
		http.Error(w, "failed to get ranking", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, ranking)
}

// SubmitScore handles score submission.
func (h *DailyHandler) SubmitScore(w http.ResponseWriter, r *http.Request) {
	var req daily.SubmitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	result, err := h.service.SubmitScore(r.Context(), req)
	if err != nil {
		if errors.Is(err, daily.ErrInvalidSubmission) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if errors.Is(err, daily.ErrSeedMismatch) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		// UNIQUE constraint violation = duplicate submission
		if strings.Contains(err.Error(), "UNIQUE constraint") {
			http.Error(w, "already submitted for today", http.StatusConflict)
			return
		}
		http.Error(w, "failed to submit score", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// GetHistory returns recent daily courses.
func (h *DailyHandler) GetHistory(w http.ResponseWriter, r *http.Request) {
	days := 7
	if d := r.URL.Query().Get("days"); d != "" {
		parsed, err := strconv.Atoi(d)
		if err != nil || parsed < 1 || parsed > 30 {
			http.Error(w, "days must be 1-30", http.StatusBadRequest)
			return
		}
		days = parsed
	}

	courses, err := h.service.GetHistory(r.Context(), days)
	if err != nil {
		http.Error(w, "failed to get history", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, courses)
}
