package ws

import (
	"encoding/json"
	"testing"
)

func TestNewEnvelope(t *testing.T) {
	data := JoinSuccessData{PlayerID: 1, HostID: 1}
	raw, err := NewEnvelope("join_success", data)
	if err != nil {
		t.Fatalf("NewEnvelope error: %v", err)
	}

	var env Envelope
	if err := json.Unmarshal(raw, &env); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if env.Type != "join_success" {
		t.Errorf("expected type join_success, got %s", env.Type)
	}

	var result JoinSuccessData
	if err := json.Unmarshal(env.Data, &result); err != nil {
		t.Fatalf("unmarshal data error: %v", err)
	}

	if result.PlayerID != 1 || result.HostID != 1 {
		t.Errorf("unexpected data: %+v", result)
	}
}

func TestNewEnvelope_LobbyState(t *testing.T) {
	data := LobbyStateData{
		Players:         []PlayerInfo{{ID: 1, Name: "P1"}, {ID: 2, Name: "P2"}},
		HostID:          1,
		SelectedSongID:  "song_01",
		SelectedBMSFile: "hard.bms",
		ReadyPlayers:    []int{1},
	}

	raw, err := NewEnvelope("lobby_state", data)
	if err != nil {
		t.Fatalf("NewEnvelope error: %v", err)
	}

	var env Envelope
	json.Unmarshal(raw, &env)

	var result LobbyStateData
	json.Unmarshal(env.Data, &result)

	if len(result.Players) != 2 {
		t.Errorf("expected 2 players, got %d", len(result.Players))
	}
	if result.SelectedSongID != "song_01" {
		t.Errorf("expected song_01, got %s", result.SelectedSongID)
	}
}

func TestNewEnvelope_OpponentScore(t *testing.T) {
	data := OpponentScoreData{
		Combo:     150,
		Judgments: map[string]int{"PERFECT": 140, "GREAT": 8, "GOOD": 2, "MISS": 0},
	}

	raw, err := NewEnvelope("opponent_score", data)
	if err != nil {
		t.Fatalf("NewEnvelope error: %v", err)
	}

	var env Envelope
	json.Unmarshal(raw, &env)

	var result OpponentScoreData
	json.Unmarshal(env.Data, &result)

	if result.Combo != 150 {
		t.Errorf("expected combo 150, got %d", result.Combo)
	}
	if result.Judgments["PERFECT"] != 140 {
		t.Errorf("expected PERFECT=140, got %d", result.Judgments["PERFECT"])
	}
}

func TestEnvelope_Unmarshal_ClientMessage(t *testing.T) {
	// Simulate a raw client message
	raw := `{"type":"select_song","data":{"song_id":"song_01","bms_file":"hard.bms","match_settings":{"speed":2.0,"modifiers":["RANDOM"],"buffs":[],"debuffs":[]}}}`

	var env Envelope
	if err := json.Unmarshal([]byte(raw), &env); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if env.Type != "select_song" {
		t.Errorf("expected select_song, got %s", env.Type)
	}

	var data SelectSongData
	if err := json.Unmarshal(env.Data, &data); err != nil {
		t.Fatalf("unmarshal data error: %v", err)
	}

	if data.SongID != "song_01" {
		t.Errorf("expected song_01, got %s", data.SongID)
	}
	if data.MatchSettings.Speed != 2.0 {
		t.Errorf("expected speed 2.0, got %f", data.MatchSettings.Speed)
	}
	if len(data.MatchSettings.Modifiers) != 1 || data.MatchSettings.Modifiers[0] != "RANDOM" {
		t.Errorf("unexpected modifiers: %v", data.MatchSettings.Modifiers)
	}
}

func TestEnvelope_SyncScore(t *testing.T) {
	raw := `{"type":"sync_score","data":{"combo":10,"judgments":{"PERFECT":10,"GREAT":0,"GOOD":0,"MISS":0}}}`

	var env Envelope
	json.Unmarshal([]byte(raw), &env)

	var data SyncScoreData
	json.Unmarshal(env.Data, &data)

	if data.Combo != 10 {
		t.Errorf("expected combo 10, got %d", data.Combo)
	}
	if data.Judgments["PERFECT"] != 10 {
		t.Errorf("expected PERFECT=10, got %d", data.Judgments["PERFECT"])
	}
}
