package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func setupTestSongs(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()

	// Create song_01 with metadata and files
	song01 := filepath.Join(dir, "song_01")
	os.MkdirAll(song01, 0755)

	meta := SongMeta{ID: "song_01", Title: "Test Song", Artist: "Test Artist", Level: "5", BPM: "120"}
	metaJSON, _ := json.Marshal(meta)
	os.WriteFile(filepath.Join(song01, "metadata.json"), metaJSON, 0644)
	os.WriteFile(filepath.Join(song01, "chart.bms"), []byte("BMS_DATA"), 0644)
	os.WriteFile(filepath.Join(song01, "kick.wav"), []byte("WAV_DATA"), 0644)

	// Create song_02 with metadata
	song02 := filepath.Join(dir, "song_02")
	os.MkdirAll(song02, 0755)

	meta2 := SongMeta{ID: "song_02", Title: "Another Song", Artist: "Another Artist", Level: "8", BPM: "160"}
	metaJSON2, _ := json.Marshal(meta2)
	os.WriteFile(filepath.Join(song02, "metadata.json"), metaJSON2, 0644)
	os.WriteFile(filepath.Join(song02, "hard.bms"), []byte("BMS_DATA_2"), 0644)

	return dir
}

func newTestMux(songsDir string) *http.ServeMux {
	mux := http.NewServeMux()
	h := NewSongsHandler(songsDir)
	h.RegisterRoutes(mux)
	return mux
}

func TestListSongs(t *testing.T) {
	dir := setupTestSongs(t)
	mux := newTestMux(dir)

	req := httptest.NewRequest("GET", "/api/songs", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var songs []SongMeta
	if err := json.Unmarshal(w.Body.Bytes(), &songs); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if len(songs) != 2 {
		t.Errorf("expected 2 songs, got %d", len(songs))
	}
}

func TestListSongs_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	mux := newTestMux(dir)

	req := httptest.NewRequest("GET", "/api/songs", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var songs []SongMeta
	json.Unmarshal(w.Body.Bytes(), &songs)
	if len(songs) != 0 {
		t.Errorf("expected 0 songs, got %d", len(songs))
	}
}

func TestGetSongManifest(t *testing.T) {
	dir := setupTestSongs(t)
	mux := newTestMux(dir)

	req := httptest.NewRequest("GET", "/api/songs/song_01", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var manifest SongManifest
	if err := json.Unmarshal(w.Body.Bytes(), &manifest); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if manifest.ID != "song_01" {
		t.Errorf("expected song_01, got %s", manifest.ID)
	}

	// Should have chart.bms and kick.wav (metadata.json excluded)
	if len(manifest.Files) != 2 {
		t.Errorf("expected 2 files, got %d: %v", len(manifest.Files), manifest.Files)
	}
}

func TestGetSongManifest_NotFound(t *testing.T) {
	dir := setupTestSongs(t)
	mux := newTestMux(dir)

	req := httptest.NewRequest("GET", "/api/songs/nonexistent", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestDownloadFile(t *testing.T) {
	dir := setupTestSongs(t)
	mux := newTestMux(dir)

	req := httptest.NewRequest("GET", "/api/songs/song_01/download/chart.bms", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	if w.Body.String() != "BMS_DATA" {
		t.Errorf("expected BMS_DATA, got %q", w.Body.String())
	}
}

func TestDownloadFile_PathTraversal(t *testing.T) {
	dir := setupTestSongs(t)
	mux := newTestMux(dir)

	tests := []struct {
		name string
		path string
	}{
		{"dot dot in id", "/api/songs/../etc/download/passwd"},
		{"dot dot in filename", "/api/songs/song_01/download/../../etc/passwd"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, req)

			if w.Code == http.StatusOK {
				t.Errorf("expected non-200 for path traversal attempt, got %d", w.Code)
			}
		})
	}
}

func TestIsValidSongID(t *testing.T) {
	tests := []struct {
		id    string
		valid bool
	}{
		{"song_01", true},
		{"", false},
		{"..", false},
		{"../etc", false},
		{"song/sub", false},
		{"song\\sub", false},
	}

	for _, tt := range tests {
		if got := isValidSongID(tt.id); got != tt.valid {
			t.Errorf("isValidSongID(%q) = %v, want %v", tt.id, got, tt.valid)
		}
	}
}

func TestIsValidFilename(t *testing.T) {
	tests := []struct {
		name  string
		valid bool
	}{
		{"chart.bms", true},
		{"kick.wav", true},
		{"", false},
		{"..", false},
		{"../secret", false},
		{"sub/file", false},
	}

	for _, tt := range tests {
		if got := isValidFilename(tt.name); got != tt.valid {
			t.Errorf("isValidFilename(%q) = %v, want %v", tt.name, got, tt.valid)
		}
	}
}
