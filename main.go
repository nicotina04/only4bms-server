package main

import (
	"log"
	"net/http"

	"github.com/nicotina04/only4bms-server/internal/api"
	"github.com/nicotina04/only4bms-server/internal/daily"
	"github.com/nicotina04/only4bms-server/internal/db"
	"github.com/nicotina04/only4bms-server/internal/lobby"
	"github.com/nicotina04/only4bms-server/internal/ws"
)

func main() {
	cfg := LoadConfig()

	log.Printf("Only4BMS Server starting on :%s", cfg.Port)
	log.Printf("Songs directory: %s", cfg.SongsDir)

	// Database
	database, err := db.Open(cfg.DBPath)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer database.Close()

	if err := db.Migrate(database); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}
	log.Printf("Database: %s", cfg.DBPath)

	store := db.NewStore(database)
	dailyService := daily.NewService(store, cfg.DailyResetHour)

	// Lobby manager (single lobby for MVP)
	mgr := lobby.NewManager()

	// WebSocket hub
	hub := ws.NewHub(cfg.ServerPassword, mgr.WSHandler())

	// Wire up register/unregister events
	go func() {
		for {
			select {
			case client := <-hub.Register:
				mgr.RegisterClient(client)
			case client := <-hub.Unregister:
				mgr.UnregisterClient(client)
			}
		}
	}()
	go hub.Run()

	// HTTP routes
	mux := http.NewServeMux()

	// Song API
	songsHandler := api.NewSongsHandler(cfg.SongsDir)
	songsHandler.RegisterRoutes(mux)

	// Daily course API
	dailyHandler := api.NewDailyHandler(dailyService)
	dailyHandler.RegisterRoutes(mux)

	// WebSocket endpoint
	mux.HandleFunc("GET /ws", hub.HandleWS)

	// Health check
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	addr := ":" + cfg.Port
	log.Printf("Listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
