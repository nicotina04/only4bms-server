package main

import (
	"log"
	"net/http"

	"github.com/nicotina04/only4bms-server/internal/api"
	"github.com/nicotina04/only4bms-server/internal/lobby"
	"github.com/nicotina04/only4bms-server/internal/ws"
)

func main() {
	cfg := LoadConfig()

	log.Printf("Only4BMS Server starting on :%s", cfg.Port)
	log.Printf("Songs directory: %s", cfg.SongsDir)

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
