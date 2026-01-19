package main

import (
	"fmt"
	"net/http"
	"sync"

	"golang.org/x/net/websocket"
)

type Connections struct {
	m     sync.Mutex
	store map[string]*websocket.Conn
}

type Server struct {
	conns Connections
}

func NewServer() *Server {
	return &Server{
		conns: Connections{
			store: make(map[string]*websocket.Conn),
		},
	}
}

func (s *Server) handleConn(ws *websocket.Conn) {
	params := ws.Request().URL.Query()
	name := params.Get("name")

	if name == "" {
		name = "Anonymous"
	}

	s.conns.m.Lock()
	s.conns.store[name] = ws
	s.conns.m.Unlock()

	fmt.Printf("New connection: %s (Remote Addr: %s)\n", name, ws.RemoteAddr())

	s.receiveMsg(ws, name)

	s.conns.m.Lock()
	delete(s.conns.store, name)
	s.conns.m.Unlock()
}

func (s *Server) receiveMsg(ws *websocket.Conn, name string) {
	buf := make([]byte, 1024)
	for {
		n, err := ws.Read(buf)
		if err != nil {
			fmt.Println("Connection closed:", ws.RemoteAddr())
			break
		}

		msg := buf[:n]
		s.broadcastMsg(msg, name)
	}
}

func (s *Server) broadcastMsg(msg []byte, sender string) {
	for name, ws := range s.conns.store {
		if sender == name {
			continue
		}

		go func(ws *websocket.Conn) {
			if _, err := ws.Write(msg); err != nil {
				fmt.Println("Broadcast error:", err)
			}
		}(ws)
	}
}

func (s *Server) sendDirectMsg(msg []byte, recipient string) {}

func main() {
	server := NewServer()

	http.Handle("/ws", websocket.Handler(server.handleConn))

	fmt.Println("Server starting on port 8080...")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println("Server connection error:", err)
	}
}
