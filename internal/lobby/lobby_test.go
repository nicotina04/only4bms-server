package lobby

import (
	"testing"

	"github.com/nicotina04/only4bms-server/internal/ws"
)

// newTestClient creates a minimal Client for testing (no real WS connection).
func newTestClient(id int, name string) *ws.Client {
	return &ws.Client{
		ID:   id,
		Name: name,
		Send: make(chan []byte, 64),
	}
}

func TestAddPlayer_FirstBecomesHost(t *testing.T) {
	l := New()
	c1 := newTestClient(1, "P1")

	if !l.AddPlayer(c1) {
		t.Fatal("expected AddPlayer to succeed")
	}
	if l.HostID != 1 {
		t.Errorf("expected host=1, got %d", l.HostID)
	}
	if l.State != StateWaiting {
		t.Errorf("expected StateWaiting, got %d", l.State)
	}
}

func TestAddPlayer_SecondTransitionsToSelecting(t *testing.T) {
	l := New()
	l.AddPlayer(newTestClient(1, "P1"))
	l.AddPlayer(newTestClient(2, "P2"))

	if l.State != StateSelecting {
		t.Errorf("expected StateSelecting, got %d", l.State)
	}
}

func TestAddPlayer_ThirdRejected(t *testing.T) {
	l := New()
	l.AddPlayer(newTestClient(1, "P1"))
	l.AddPlayer(newTestClient(2, "P2"))

	if l.AddPlayer(newTestClient(3, "P3")) {
		t.Fatal("expected third player to be rejected")
	}
}

func TestRemovePlayer_HostSuccession(t *testing.T) {
	l := New()
	l.AddPlayer(newTestClient(1, "P1"))
	l.AddPlayer(newTestClient(2, "P2"))

	l.RemovePlayer(1) // host leaves

	if l.HostID != 2 {
		t.Errorf("expected host=2 after succession, got %d", l.HostID)
	}
	if l.State != StateWaiting {
		t.Errorf("expected StateWaiting after player left, got %d", l.State)
	}
}

func TestRemovePlayer_BothLeave_Empty(t *testing.T) {
	l := New()
	l.AddPlayer(newTestClient(1, "P1"))
	l.AddPlayer(newTestClient(2, "P2"))

	l.RemovePlayer(1)
	l.RemovePlayer(2)

	if !l.IsEmpty() {
		t.Error("expected lobby to be empty")
	}
}

func TestSelectSong_OnlyHost(t *testing.T) {
	l := New()
	l.AddPlayer(newTestClient(1, "P1"))
	l.AddPlayer(newTestClient(2, "P2"))

	settings := ws.MatchSettings{Speed: 2.0}

	// Non-host should fail
	if l.SelectSong(2, "song_01", "hard.bms", settings) {
		t.Fatal("non-host should not be able to select song")
	}

	// Host should succeed
	if !l.SelectSong(1, "song_01", "hard.bms", settings) {
		t.Fatal("host should be able to select song")
	}

	if l.State != StateDownloading {
		t.Errorf("expected StateDownloading, got %d", l.State)
	}
	if l.SelectedSongID != "song_01" {
		t.Errorf("expected song_01, got %s", l.SelectedSongID)
	}
}

func TestSelectSong_WrongState(t *testing.T) {
	l := New()
	c1 := newTestClient(1, "P1")
	l.AddPlayer(c1)
	// Only 1 player → StateWaiting

	if l.SelectSong(1, "song_01", "hard.bms", ws.MatchSettings{}) {
		t.Fatal("should not select song in Waiting state")
	}
}

func TestReady_AllReadyStartsGame(t *testing.T) {
	l := New()
	l.AddPlayer(newTestClient(1, "P1"))
	l.AddPlayer(newTestClient(2, "P2"))
	l.SelectSong(1, "song_01", "hard.bms", ws.MatchSettings{})

	if l.SetReady(1) {
		t.Fatal("should not start with only one ready")
	}

	if !l.SetReady(2) {
		t.Fatal("expected all-ready to return true")
	}

	if l.State != StatePlaying {
		t.Errorf("expected StatePlaying, got %d", l.State)
	}
}

func TestReady_WrongState(t *testing.T) {
	l := New()
	l.AddPlayer(newTestClient(1, "P1"))
	l.AddPlayer(newTestClient(2, "P2"))
	// State is Selecting, not Downloading

	if l.SetReady(1) {
		t.Fatal("should not be able to ready in Selecting state")
	}
}

func TestFinishMatch_ResetsToSelecting(t *testing.T) {
	l := New()
	l.AddPlayer(newTestClient(1, "P1"))
	l.AddPlayer(newTestClient(2, "P2"))
	l.SelectSong(1, "song_01", "hard.bms", ws.MatchSettings{})
	l.SetReady(1)
	l.SetReady(2)

	l.FinishMatch()

	if l.State != StateSelecting {
		t.Errorf("expected StateSelecting after finish, got %d", l.State)
	}
	if l.SelectedSongID != "" {
		t.Error("expected song selection to be cleared")
	}
	if len(l.ReadyPlayers) != 0 {
		t.Error("expected ready players to be cleared")
	}
}

func TestGetLobbyState(t *testing.T) {
	l := New()
	l.AddPlayer(newTestClient(1, "P1"))
	l.AddPlayer(newTestClient(2, "P2"))
	l.SelectSong(1, "song_01", "hard.bms", ws.MatchSettings{Speed: 3.0})
	l.SetReady(1)

	state := l.GetLobbyState()

	if len(state.Players) != 2 {
		t.Errorf("expected 2 players, got %d", len(state.Players))
	}
	if state.HostID != 1 {
		t.Errorf("expected host=1, got %d", state.HostID)
	}
	if state.SelectedSongID != "song_01" {
		t.Errorf("expected song_01, got %s", state.SelectedSongID)
	}
	if len(state.ReadyPlayers) != 1 {
		t.Errorf("expected 1 ready player, got %d", len(state.ReadyPlayers))
	}
}

func TestBroadcast(t *testing.T) {
	l := New()
	c1 := newTestClient(1, "P1")
	c2 := newTestClient(2, "P2")
	l.AddPlayer(c1)
	l.AddPlayer(c2)

	l.Broadcast("lobby_state", ws.LobbyStateData{
		Players: []ws.PlayerInfo{{ID: 1, Name: "P1"}, {ID: 2, Name: "P2"}},
		HostID:  1,
	})

	// Both clients should have a message in their Send channel
	select {
	case msg := <-c1.Send:
		if len(msg) == 0 {
			t.Error("expected c1 to receive non-empty message")
		}
	default:
		t.Error("c1 did not receive broadcast")
	}

	select {
	case msg := <-c2.Send:
		if len(msg) == 0 {
			t.Error("expected c2 to receive non-empty message")
		}
	default:
		t.Error("c2 did not receive broadcast")
	}
}

func TestBroadcastExcept(t *testing.T) {
	l := New()
	c1 := newTestClient(1, "P1")
	c2 := newTestClient(2, "P2")
	l.AddPlayer(c1)
	l.AddPlayer(c2)

	l.BroadcastExcept(1, "opponent_score", ws.OpponentScoreData{Combo: 50})

	// c2 should have a message, c1 should not
	select {
	case msg := <-c2.Send:
		if len(msg) == 0 {
			t.Error("expected c2 to receive message")
		}
	default:
		t.Error("c2 did not receive message")
	}

	select {
	case <-c1.Send:
		t.Error("c1 should not have received the message")
	default:
		// expected
	}
}
