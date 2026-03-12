package api

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type SongMeta struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Artist string `json:"artist"`
	Level  string `json:"level"`
	BPM    string `json:"bpm"`
}

type SongManifest struct {
	ID    string   `json:"id"`
	Files []string `json:"files"`
}

// SongsHandler serves song-related HTTP endpoints.
type SongsHandler struct {
	SongsDir string
}

// NewSongsHandler creates a new SongsHandler.
func NewSongsHandler(songsDir string) *SongsHandler {
	return &SongsHandler{SongsDir: songsDir}
}

// RegisterRoutes registers song API routes on the given mux.
func (h *SongsHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/songs", h.ListSongs)
	mux.HandleFunc("GET /api/songs/{id}", h.GetSongManifest)
	mux.HandleFunc("GET /api/songs/{id}/download/{filename}", h.DownloadFile)
}

// ListSongs returns all available songs.
func (h *SongsHandler) ListSongs(w http.ResponseWriter, r *http.Request) {
	entries, err := os.ReadDir(h.SongsDir)
	if err != nil {
		http.Error(w, "failed to read songs directory", http.StatusInternalServerError)
		return
	}

	songs := make([]SongMeta, 0)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		meta, err := loadSongMeta(filepath.Join(h.SongsDir, entry.Name()))
		if err != nil {
			continue
		}
		songs = append(songs, *meta)
	}

	writeJSON(w, http.StatusOK, songs)
}

// GetSongManifest returns the file list for a specific song.
func (h *SongsHandler) GetSongManifest(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	songDir := filepath.Join(h.SongsDir, id)

	if !isValidSongID(id) || !dirExists(songDir) {
		http.Error(w, "song not found", http.StatusNotFound)
		return
	}

	entries, err := os.ReadDir(songDir)
	if err != nil {
		http.Error(w, "failed to read song directory", http.StatusInternalServerError)
		return
	}

	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() && entry.Name() != "metadata.json" {
			files = append(files, entry.Name())
		}
	}

	writeJSON(w, http.StatusOK, SongManifest{ID: id, Files: files})
}

// DownloadFile serves a specific file from a song directory.
func (h *SongsHandler) DownloadFile(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	filename := r.PathValue("filename")

	if !isValidSongID(id) || !isValidFilename(filename) {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	filePath := filepath.Join(h.SongsDir, id, filename)

	// Prevent path traversal
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}
	absSongsDir, _ := filepath.Abs(h.SongsDir)
	if !strings.HasPrefix(absPath, absSongsDir) {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}

	http.ServeFile(w, r, filePath)
}

func loadSongMeta(songDir string) (*SongMeta, error) {
	data, err := os.ReadFile(filepath.Join(songDir, "metadata.json"))
	if err != nil {
		return nil, err
	}
	var meta SongMeta
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, err
	}
	// Use directory name as ID if not set
	if meta.ID == "" {
		meta.ID = filepath.Base(songDir)
	}
	return &meta, nil
}

func isValidSongID(id string) bool {
	return id != "" && !strings.Contains(id, "..") && !strings.Contains(id, "/") && !strings.Contains(id, "\\")
}

func isValidFilename(name string) bool {
	return name != "" && !strings.Contains(name, "..") && !strings.Contains(name, "/") && !strings.Contains(name, "\\")
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
