package main

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

// Session wrapper around websocket connections.
type Session struct {
	conn    *websocket.Conn
	output  chan Envelope
	open    bool
	rwmutex *sync.RWMutex
	hub     *Hub
	device  string
}

func (s *Session) writeMessage(msg Envelope) {
	if s.closed() {
		log.Printf("error: tried to write to closed a session")
		return
	}
	select {
	case s.output <- msg:
	default:
		log.Printf("error: session message buffer is full")
	}
}

func (s *Session) closed() bool {
	s.rwmutex.RLock()
	defer s.rwmutex.RUnlock()

	return !s.open
}

func (s *Session) readPump() {
	defer func() {
		s.hub.unregister <- s
		s.conn.Close()
	}()
	s.conn.SetReadLimit(maxMessageSize)
	s.conn.SetReadDeadline(time.Now().Add(pongWait))
	s.conn.SetPongHandler(func(string) error {
		s.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	for {
		// _, message, err := s.conn.ReadMessage()
		var envelope Envelope
		err := s.conn.ReadJSON(&envelope)
		log.Printf("envelope: %v", envelope)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error in readPump: %v", err)
			}
			break
		}
		// envelope := Envelope{t: websocket.TextMessage, Msg: message}
		s.hub.broadcast <- envelope
	}
}

func (s *Session) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		s.conn.Close()
	}()
	for {
		select {
		case msg, ok := <-s.output:
			s.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				s.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			// err := s.conn.WriteMessage(websocket.TextMessage, []byte(msg.Msg))
			err := s.conn.WriteJSON(msg)
			if err != nil {
				log.Printf("error in writePump: %v", err)
				return
			}

		case <-ticker.C:
			s.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := s.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func wsHandler(hub *Hub, w http.ResponseWriter, r *http.Request) {
	log.Printf("device: %v", r.URL.Query().Get("device"))
	conn, err := upgrader.Upgrade(w, r, w.Header())
	if err != nil {
		log.Printf("upgrader.Upgrade error: %v", err)
		return
	}
	session := &Session{
		conn:    conn,
		output:  make(chan Envelope, 256),
		open:    true,
		rwmutex: &sync.RWMutex{},
		hub:     hub,
		device:  r.URL.Query().Get("device"),
	}
	hub.register <- session
	go session.readPump()
	go session.writePump()
}
