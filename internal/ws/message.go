package ws

import "encoding/json"

// Envelope is the top-level JSON wrapper for all WS messages.
type Envelope struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data,omitempty"`
}

// --- Client → Server payloads ---

type JoinData struct {
	Name     string `json:"name"`
	Password string `json:"password,omitempty"`
}

type SelectSongData struct {
	SongID        string        `json:"song_id"`
	BMSFile       string        `json:"bms_file"`
	MatchSettings MatchSettings `json:"match_settings"`
}

type MatchSettings struct {
	Speed     float64  `json:"speed"`
	Modifiers []string `json:"modifiers"`
	Buffs     []string `json:"buffs"`
	Debuffs   []string `json:"debuffs"`
}

type SyncScoreData struct {
	Combo     int            `json:"combo"`
	Judgments map[string]int `json:"judgments"`
}

// --- Server → Client payloads ---

type JoinSuccessData struct {
	PlayerID int `json:"player_id"`
	HostID   int `json:"host_id"`
}

type JoinErrorData struct {
	Message string `json:"message"`
}

type PlayerInfo struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type LobbyStateData struct {
	Players        []PlayerInfo  `json:"players"`
	HostID         int           `json:"host_id"`
	SelectedSongID string        `json:"selected_song_id,omitempty"`
	SelectedBMSFile string       `json:"selected_bms_file,omitempty"`
	ReadyPlayers   []int         `json:"ready_players"`
}

type StartGameData struct {
	StartTimeOffset int           `json:"start_time_offset"`
	SongID          string        `json:"song_id"`
	BMSFile         string        `json:"bms_file"`
	MatchSettings   MatchSettings `json:"match_settings"`
}

type OpponentScoreData struct {
	Combo     int            `json:"combo"`
	Judgments map[string]int `json:"judgments"`
}

// NewEnvelope creates an Envelope with a marshalled data payload.
func NewEnvelope(msgType string, data any) ([]byte, error) {
	raw, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return json.Marshal(Envelope{Type: msgType, Data: raw})
}
