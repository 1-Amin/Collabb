package ws

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestHubBroadcast(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	defer hub.Stop()

	// Start a test WS server
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hub.ServeWS(w, r, "board-1")
	}))
	defer srv.Close()

	url := "ws" + strings.TrimPrefix(srv.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	// Give register goroutine time to run
	time.Sleep(50 * time.Millisecond)

	payload, _ := json.Marshal(map[string]string{"key": "value"})
	hub.Broadcast(Message{BoardID: "board-1", Type: "test.event", Payload: payload})

	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, msg, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	var got Message
	if err := json.Unmarshal(msg, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.Type != "test.event" {
		t.Errorf("want type test.event, got %s", got.Type)
	}
}

func TestHubIsolateBoards(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	defer hub.Stop()

	makeConn := func(boardID string) *websocket.Conn {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			hub.ServeWS(w, r, boardID)
		}))
		t.Cleanup(srv.Close)
		url := "ws" + strings.TrimPrefix(srv.URL, "http")
		conn, _, err := websocket.DefaultDialer.Dial(url, nil)
		if err != nil {
			t.Fatalf("dial: %v", err)
		}
		return conn
	}

	c1 := makeConn("board-A")
	c2 := makeConn("board-B")
	defer c1.Close()
	defer c2.Close()
	time.Sleep(50 * time.Millisecond)

	payload, _ := json.Marshal("hello")
	hub.Broadcast(Message{BoardID: "board-A", Type: "ping", Payload: payload})

	// c1 should receive; c2 should not (timeout)
	c1.SetReadDeadline(time.Now().Add(time.Second))
	if _, _, err := c1.ReadMessage(); err != nil {
		t.Errorf("c1 expected message: %v", err)
	}

	c2.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
	if _, _, err := c2.ReadMessage(); err == nil {
		t.Error("c2 should not receive board-A message")
	}
}
