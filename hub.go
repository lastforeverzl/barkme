package main

import (
	"log"
	"sync"

	"github.com/lastforeverzl/barkme/message"
)

// Hub maintains the set of active clients and broadcasts messages to the
// sessions.
type Hub struct {
	sessions   map[*Session]bool
	broadcast  chan message.Envelope
	register   chan *Session
	unregister chan *Session
	rwmutex    *sync.RWMutex
	env        *Env
}

func newHub(env *Env) *Hub {
	return &Hub{
		sessions:   make(map[*Session]bool),
		broadcast:  make(chan message.Envelope),
		register:   make(chan *Session),
		unregister: make(chan *Session),
		rwmutex:    &sync.RWMutex{},
		env:        env,
	}
}

func (h *Hub) run() {
	for {
		select {
		case session := <-h.register:
			h.rwmutex.Lock()
			h.sessions[session] = true
			h.rwmutex.Unlock()
			log.Printf("register a session")
		case session := <-h.unregister:
			if _, ok := h.sessions[session]; ok {
				h.rwmutex.Lock()
				delete(h.sessions, session)
				close(session.output)
				h.rwmutex.Unlock()
				log.Printf("unregister a session")
			}
		case message := <-h.broadcast:
			h.rwmutex.RLock()
			log.Printf("sessions sum: %v", len(h.sessions))
			go h.env.db.UpdateUserAction(message)
			for session := range h.sessions {
				if session.device != message.Username {
					session.writeMessage(message)
				}

			}
			h.rwmutex.RUnlock()
		}
	}
}
