package main

import (
	"bufio"
	"fmt"
	"net/http"
	"sync"

	"golang.org/x/net/websocket"
)

type Connections struct {
	m     sync.Mutex
	store map[*websocket.Conn]bool
}

type Server struct {
	conns Connections
}

func NewServer() *Server {
	return &Server{
		conns: Connections{
			store: make(map[*websocket.Conn]bool),
		},
	}
}

func (s *Server) handleConn(ws *websocket.Conn) {
	s.conns.m.Lock()
	defer s.conns.m.Unlock()

	fmt.Println("New incoming connection:", ws.RemoteAddr())

	s.conns.store[ws] = true

	go s.receiveMsg(ws)
}

func (s *Server) receiveMsg(ws *websocket.Conn) {
	scanner := bufio.NewScanner(ws)

	for scanner.Scan() {
		msg := scanner.Text()
		fmt.Printf("Received message from %v: %s", ws.RemoteAddr(), msg)

		ws.Write([]byte("Thank you!"))
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading from connection:", err)
		return
	}
}

func main() {
	server := NewServer()

	http.Handle("/ws", websocket.Handler(server.handleConn))

	fmt.Println("Server starting on port 8080...")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println("Server connection error:", err)
	}
}
