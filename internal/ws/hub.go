package ws

import (
	"encoding/json"
	"log"
	"net/http"
	"sync/atomic"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// IncomingMessage pairs a raw message with the client that sent it.
type IncomingMessage struct {
	Client *Client
	Raw    []byte
}

// MessageHandler processes incoming WebSocket messages.
type MessageHandler func(client *Client, env Envelope)

// Hub manages WebSocket clients and message routing.
type Hub struct {
	Register   chan *Client
	Unregister chan *Client
	Incoming   chan *IncomingMessage

	handler  MessageHandler
	nextID   atomic.Int32
	password string
}

// NewHub creates a new Hub.
func NewHub(password string, handler MessageHandler) *Hub {
	return &Hub{
		Register:   make(chan *Client, 16),
		Unregister: make(chan *Client, 16),
		Incoming:   make(chan *IncomingMessage, 256),
		handler:    handler,
		password:   password,
	}
}

// Run starts the Hub's main event loop.
func (h *Hub) Run() {
	for msg := range h.Incoming {
		var env Envelope
		if err := json.Unmarshal(msg.Raw, &env); err != nil {
			log.Printf("invalid message from %s: %v", msg.Client.Name, err)
			continue
		}
		h.handler(msg.Client, env)
	}
}

// HandleWS upgrades HTTP to WebSocket and registers the client.
func (h *Hub) HandleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("ws upgrade error: %v", err)
		return
	}

	id := int(h.nextID.Add(1))
	name := r.URL.Query().Get("name")
	if name == "" {
		name = "Player"
	}

	client := &Client{
		ID:   id,
		Name: name,
		Hub:  h,
		Conn: conn,
		Send: make(chan []byte, 64),
	}

	// Password check
	password := r.URL.Query().Get("password")
	if h.password != "" && password != h.password {
		client.SendMessage("join_error", JoinErrorData{Message: "Invalid password"})
		conn.Close()
		return
	}

	h.Register <- client

	go client.WritePump()
	go client.ReadPump()
}
