package lobby

import (
	"encoding/json"
	"log"

	"github.com/nicotina04/only4bms-server/internal/ws"
)

// Manager handles lobby lifecycle and routes WebSocket messages.
type Manager struct {
	lobby *Lobby
}

// NewManager creates a Manager with a single lobby (MVP).
func NewManager() *Manager {
	return &Manager{
		lobby: New(),
	}
}

// HandleRegister processes a new client connection.
func (m *Manager) HandleRegister(c *ws.Client) {
	if !m.lobby.AddPlayer(c) {
		c.SendMessage("join_error", ws.JoinErrorData{Message: "Lobby is full"})
		c.Close()
		return
	}

	log.Printf("player joined: %s (id=%d)", c.Name, c.ID)

	c.SendMessage("join_success", ws.JoinSuccessData{
		PlayerID: c.ID,
		HostID:   m.lobby.HostID,
	})

	m.broadcastLobbyState()
}

// HandleUnregister processes a client disconnection.
func (m *Manager) HandleUnregister(c *ws.Client) {
	m.lobby.RemovePlayer(c.ID)
	log.Printf("player left: %s (id=%d)", c.Name, c.ID)

	if !m.lobby.IsEmpty() {
		m.broadcastLobbyState()
	}
}

// HandleMessage routes an incoming WebSocket message to the appropriate handler.
func (m *Manager) HandleMessage(c *ws.Client, env ws.Envelope) {
	switch env.Type {
	case "select_song":
		m.handleSelectSong(c, env.Data)
	case "ready":
		m.handleReady(c)
	case "sync_score":
		m.handleSyncScore(c, env.Data)
	default:
		log.Printf("unknown message type: %s from %s", env.Type, c.Name)
	}
}

func (m *Manager) handleSelectSong(c *ws.Client, raw json.RawMessage) {
	var data ws.SelectSongData
	if err := json.Unmarshal(raw, &data); err != nil {
		log.Printf("invalid select_song data: %v", err)
		return
	}

	if !m.lobby.SelectSong(c.ID, data.SongID, data.BMSFile, data.MatchSettings) {
		c.SendMessage("join_error", ws.JoinErrorData{Message: "Cannot select song: not host or wrong state"})
		return
	}

	log.Printf("host %s selected song: %s / %s", c.Name, data.SongID, data.BMSFile)
	m.broadcastLobbyState()
}

func (m *Manager) handleReady(c *ws.Client) {
	allReady := m.lobby.SetReady(c.ID)

	log.Printf("player %s is ready", c.Name)
	m.broadcastLobbyState()

	if allReady {
		m.lobby.mu.RLock()
		startData := ws.StartGameData{
			StartTimeOffset: 3000,
			SongID:          m.lobby.SelectedSongID,
			BMSFile:         m.lobby.SelectedBMSFile,
			MatchSettings:   m.lobby.MatchSettings,
		}
		m.lobby.mu.RUnlock()

		log.Printf("all players ready, starting game: %s", startData.SongID)
		m.lobby.Broadcast("start_game", startData)
	}
}

func (m *Manager) handleSyncScore(c *ws.Client, raw json.RawMessage) {
	var data ws.SyncScoreData
	if err := json.Unmarshal(raw, &data); err != nil {
		log.Printf("invalid sync_score data: %v", err)
		return
	}

	m.lobby.BroadcastExcept(c.ID, "opponent_score", ws.OpponentScoreData{
		Combo:     data.Combo,
		Judgments: data.Judgments,
	})
}

func (m *Manager) broadcastLobbyState() {
	state := m.lobby.GetLobbyState()
	m.lobby.Broadcast("lobby_state", state)
}

// WSHandler returns the MessageHandler for the Hub.
func (m *Manager) WSHandler() ws.MessageHandler {
	return m.HandleMessage
}

// RegisterClient handles a new connection from the Hub.
func (m *Manager) RegisterClient(c *ws.Client) {
	m.HandleRegister(c)
}

// UnregisterClient handles a disconnection from the Hub.
func (m *Manager) UnregisterClient(c *ws.Client) {
	m.HandleUnregister(c)
}
