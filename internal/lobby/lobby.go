package lobby

import (
	"sync"

	"github.com/nicotina04/only4bms-server/internal/ws"
)

// State represents the lobby state machine.
type State int

const (
	StateWaiting     State = iota // waiting for players
	StateSelecting                // host is selecting a song
	StateDownloading              // players are downloading the song
	StatePlaying                  // match in progress
)

const MinPlayers = 2

// Lobby holds the state for a match room.
type Lobby struct {
	mu sync.RWMutex

	MaxPlayers int
	Players    map[int]*ws.Client // player_id → client
	HostID     int
	State      State

	SelectedSongID  string
	SelectedBMSFile string
	MatchSettings   ws.MatchSettings
	ReadyPlayers    map[int]bool
}

// New creates a new empty lobby with the given max player capacity.
func New(maxPlayers int) *Lobby {
	return &Lobby{
		MaxPlayers:   maxPlayers,
		Players:      make(map[int]*ws.Client),
		ReadyPlayers: make(map[int]bool),
		State:        StateWaiting,
	}
}

// AddPlayer adds a client to the lobby. Returns false if full.
func (l *Lobby) AddPlayer(c *ws.Client) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	if len(l.Players) >= l.MaxPlayers {
		return false
	}

	l.Players[c.ID] = c

	// First player becomes host
	if len(l.Players) == 1 {
		l.HostID = c.ID
	}

	// Allow song selection once minimum players have joined
	if len(l.Players) >= MinPlayers && l.State == StateWaiting {
		l.State = StateSelecting
	}

	return true
}

// RemovePlayer removes a client and handles host succession.
func (l *Lobby) RemovePlayer(id int) {
	l.mu.Lock()
	defer l.mu.Unlock()

	delete(l.Players, id)
	delete(l.ReadyPlayers, id)

	if len(l.Players) == 0 {
		return
	}

	// Host succession
	if l.HostID == id {
		for pid := range l.Players {
			l.HostID = pid
			break
		}
	}

	// Reset to waiting if we lost a player
	l.resetState()
}

// SelectSong sets the selected song (host only).
func (l *Lobby) SelectSong(playerID int, songID, bmsFile string, settings ws.MatchSettings) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	if playerID != l.HostID || l.State != StateSelecting {
		return false
	}

	l.SelectedSongID = songID
	l.SelectedBMSFile = bmsFile
	l.MatchSettings = settings
	l.ReadyPlayers = make(map[int]bool)
	l.State = StateDownloading

	return true
}

// SetReady marks a player as ready. Returns true if all players are ready.
func (l *Lobby) SetReady(playerID int) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.State != StateDownloading {
		return false
	}

	l.ReadyPlayers[playerID] = true

	if len(l.ReadyPlayers) == len(l.Players) && len(l.Players) >= MinPlayers {
		l.State = StatePlaying
		return true
	}

	return false
}

// FinishMatch resets the lobby back to selecting state.
func (l *Lobby) FinishMatch() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.resetState()
}

func (l *Lobby) resetState() {
	l.SelectedSongID = ""
	l.SelectedBMSFile = ""
	l.MatchSettings = ws.MatchSettings{}
	l.ReadyPlayers = make(map[int]bool)

	if len(l.Players) >= MinPlayers {
		l.State = StateSelecting
	} else {
		l.State = StateWaiting
	}
}

// IsEmpty returns true if no players are in the lobby.
func (l *Lobby) IsEmpty() bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return len(l.Players) == 0
}

// Broadcast sends a message to all players in the lobby.
func (l *Lobby) Broadcast(msgType string, data any) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	for _, c := range l.Players {
		c.SendMessage(msgType, data)
	}
}

// BroadcastExcept sends a message to all players except one.
func (l *Lobby) BroadcastExcept(excludeID int, msgType string, data any) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	for _, c := range l.Players {
		if c.ID != excludeID {
			c.SendMessage(msgType, data)
		}
	}
}

// GetLobbyState builds the lobby state payload.
func (l *Lobby) GetLobbyState() ws.LobbyStateData {
	l.mu.RLock()
	defer l.mu.RUnlock()

	players := make([]ws.PlayerInfo, 0, len(l.Players))
	for _, c := range l.Players {
		players = append(players, ws.PlayerInfo{ID: c.ID, Name: c.Name})
	}

	ready := make([]int, 0, len(l.ReadyPlayers))
	for pid := range l.ReadyPlayers {
		ready = append(ready, pid)
	}

	return ws.LobbyStateData{
		Players:         players,
		HostID:          l.HostID,
		SelectedSongID:  l.SelectedSongID,
		SelectedBMSFile: l.SelectedBMSFile,
		ReadyPlayers:    ready,
	}
}
