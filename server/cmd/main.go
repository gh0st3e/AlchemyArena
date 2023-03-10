package main

import (
	"golang.org/x/net/websocket"
	"io"
	"log"
	"net/http"
)

type Server struct {
	conns map[*websocket.Conn]int
	room  map[int]map[*websocket.Conn]bool
	ind   int
}

func NewServer() *Server {
	return &Server{
		conns: make(map[*websocket.Conn]int),
		room:  make(map[int]map[*websocket.Conn]bool),
		ind:   1,
	}
}

func (s *Server) handleWS(ws *websocket.Conn) {
	log.Printf("New connection: %s\n", ws.RemoteAddr())
	s.conns[ws] = s.ind
	if len(s.conns)%2 == 0 {
		s.ind++
	}
	s.readLoop(ws)

}

func (s *Server) readLoop(ws *websocket.Conn) {
	buf := make([]byte, 1024)
	for {
		n, err := ws.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Printf("Read error: %s\n", err)
			continue
		}

		msg := buf[:n]
		log.Println(string(msg))
		s.broadcast(ws, msg)

	}
}

func (s *Server) broadcast(ws *websocket.Conn, b []byte) {
	ind := s.conns[ws]
	for ws := range s.conns {
		if s.conns[ws] == ind {
			go func(ws *websocket.Conn) {
				if _, err := ws.Write(b); err != nil {
					log.Printf("Write broadcast error: %s", err)
				}
			}(ws)
		}
	}
}

func (s *Server) handleHTTP(w http.ResponseWriter, r *http.Request) {
	// Разрешить все источники
	w.Header().Set("Access-Control-Allow-Origin", "*")
	// Разрешить заголовки и методы CORS
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	// Добавить CSP заголовок
	w.Header().Set("Content-Security-Policy", "connect-src ws://localhost:3000")

	if r.Method == "GET" && r.URL.Path == "/" {
		http.ServeFile(w, r, "test.html")
	} else {
		http.NotFound(w, r)
	}
}

func main() {
	server := NewServer()
	http.HandleFunc("/", server.handleHTTP)
	http.Handle("/ws", websocket.Handler(server.handleWS))
	err := http.ListenAndServe("", nil)
	if err != nil {
		log.Printf("listen error: %s", err)
	}
}
