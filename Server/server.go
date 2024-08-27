package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
)

type Client struct {
	conn net.Conn
	name string
}

type Server struct {
	listenAddr string
	ln         net.Listener
	quitch     chan struct{}
	msgch      chan string
	clients    map[string]*Client
}

func NewServer(listenAddr string) *Server {
	return &Server{
		listenAddr: listenAddr,
		quitch:     make(chan struct{}),
		msgch:      make(chan string, 10),
		clients:    make(map[string]*Client),
	}
}

func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.listenAddr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %v", s.listenAddr, err)
	}
	defer ln.Close()
	s.ln = ln

	log.Printf("Server listening on %s", s.listenAddr)

	go s.acceptLoop()
	go s.messageLoop()

	<-s.quitch

	return nil
}

func (s *Server) acceptLoop() {
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			log.Printf("Accept error: %v", err)
			continue
		}

		client := &Client{conn: conn}
		go s.handleClient(client)
	}
}

func (s *Server) handleClient(client *Client) {
	defer func() {
		s.removeClient(client)
		client.conn.Close()
	}()

	scanner := bufio.NewScanner(client.conn)
	if scanner.Scan() {
		client.name = strings.TrimSpace(scanner.Text())
		s.clients[client.name] = client
		s.broadcast(fmt.Sprintf("%s has joined the chat", client.name))
	}

	for scanner.Scan() {
		msg := scanner.Text()
		if strings.HasPrefix(msg, "@") {
			parts := strings.SplitN(msg, " ", 2)
			if len(parts) < 2 {
				continue
			}
			targetName := strings.TrimPrefix(parts[0], "@")
			message := parts[1]
			if targetClient, ok := s.clients[targetName]; ok {
				fmt.Fprintf(targetClient.conn, "%s (private): %s\n", client.name, message)
			} else {
				fmt.Fprintf(client.conn, "User %s not found\n", targetName)
			}
		} else {
			s.broadcast(fmt.Sprintf("%s: %s", client.name, msg))
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Read error from %s: %v", client.conn.RemoteAddr(), err)
	}
}

func (s *Server) broadcast(message string) {
	for _, client := range s.clients {
		fmt.Fprintf(client.conn, "%s\n", message)
	}
}

func (s *Server) removeClient(client *Client) {
	delete(s.clients, client.name)
	s.broadcast(fmt.Sprintf("%s has left the chat", client.name))
}

func (s *Server) messageLoop() {
	for msg := range s.msgch {
		s.broadcast(msg)
	}
}

func main() {
	server := NewServer(":3000")
	if err := server.Start(); err != nil {
		log.Fatal("Server error:", err)
	}
}
