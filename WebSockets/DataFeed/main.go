package main

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"golang.org/x/net/websocket"
)

type Server struct {
	connections map[*websocket.Conn]bool
}

func NewServer() *Server {
	return &Server{connections: make(map[*websocket.Conn]bool)}
}

func (s *Server) handleWs(ws *websocket.Conn) {
	fmt.Println("new incoming conenction from client:", ws.RemoteAddr())
	s.connections[ws] = true // not concurrent safe, should use mutex
	s.readLoop(ws)
}

func (s *Server) handleLiveDataFeed(ws *websocket.Conn) {
	fmt.Println("new incoming live-data-feed subscription from client:", ws.RemoteAddr())
	for {
		payload := fmt.Sprintf("live feed data:->%s\n", time.Now().String())
		ws.Write([]byte(payload))
		time.Sleep(2 * time.Second)
	}
}

func (s *Server) readLoop(ws *websocket.Conn) {
	buf := make([]byte, 1024)
	for {
		n, err := ws.Read(buf)
		if err != nil {
			if err == io.EOF {
				s.connections[ws] = false // close connection
				break
			}
			fmt.Println("read error:", err)
			continue
		}
		msg := []byte(fmt.Sprintf("user %s wrote: ", ws.RemoteAddr()) + string(buf[:n]))
		s.broadcast(msg)
	}
}

func (s *Server) broadcast(b []byte) {
	for ws := range s.connections {
		go func(ws *websocket.Conn) {
			if _, err := ws.Write(b); err != nil {
				fmt.Println("write error:", err)
			}

		}(ws)
	}
}

func main() {
	server := NewServer()
	http.Handle("/chat", websocket.Handler(server.handleWs))
	http.Handle("/live-feed", websocket.Handler(server.handleLiveDataFeed))
	http.ListenAndServe(":3000", nil)
}
