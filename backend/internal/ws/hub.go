package ws

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 4096
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Allow all origins; tighten in prod
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Message is a board event broadcast to clients.
type Message struct {
	BoardID string          `json:"board_id"`
	Type    string          `json:"type"` // card.created, card.moved, column.updated, ...
	Payload json.RawMessage `json:"payload"`
}

// Client is a single WebSocket connection.
type Client struct {
	hub     *Hub
	boardID string
	conn    *websocket.Conn
	send    chan []byte
}

// Hub maintains the set of active clients and broadcasts messages.
type Hub struct {
	// clients maps boardID → set of clients
	clients    map[string]map[*Client]struct{}
	broadcast  chan Message
	register   chan *Client
	unregister chan *Client
	done       chan struct{}
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[string]map[*Client]struct{}),
		broadcast:  make(chan Message, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		done:       make(chan struct{}),
	}
}

// Run processes hub events; call in a dedicated goroutine.
func (h *Hub) Run() {
	for {
		select {
		case <-h.done:
			// Close all client send channels so write pumps exit
			for _, clients := range h.clients {
				for c := range clients {
					close(c.send)
				}
			}
			return

		case c := <-h.register:
			if h.clients[c.boardID] == nil {
				h.clients[c.boardID] = make(map[*Client]struct{})
			}
			h.clients[c.boardID][c] = struct{}{}

		case c := <-h.unregister:
			if board := h.clients[c.boardID]; board != nil {
				if _, ok := board[c]; ok {
					delete(board, c)
					close(c.send)
					if len(board) == 0 {
						delete(h.clients, c.boardID)
					}
				}
			}

		case msg := <-h.broadcast:
			data, err := json.Marshal(msg)
			if err != nil {
				log.Printf("ws: marshal error: %v", err)
				continue
			}
			for c := range h.clients[msg.BoardID] {
				select {
				case c.send <- data:
				default:
					// Slow client: drop and disconnect
					close(c.send)
					delete(h.clients[msg.BoardID], c)
				}
			}
		}
	}
}

// Broadcast queues a message for fan-out to all board subscribers.
func (h *Hub) Broadcast(msg Message) {
	h.broadcast <- msg
}

// Stop signals the hub to shut down.
func (h *Hub) Stop() { close(h.done) }

// ServeWS upgrades the connection and registers the client for the given boardID.
func (h *Hub) ServeWS(w http.ResponseWriter, r *http.Request, boardID string) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("ws upgrade: %v", err)
		return
	}
	c := &Client{hub: h, boardID: boardID, conn: conn, send: make(chan []byte, 256)}
	h.register <- c
	go c.writePump()
	go c.readPump()
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	for {
		// We only need to read to detect disconnects and handle pongs
		if _, _, err := c.conn.ReadMessage(); err != nil {
			break
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case msg, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(msg)
			// Batch any queued messages
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}
			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
