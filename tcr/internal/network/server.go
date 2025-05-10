package network

import (
	"fmt"
	"log"
	"net"
	"sync"
	"tcr/internal/models"
)

// Client represents a connected client
type Client struct {
	Username string
	Conn     net.Conn
}

// GameServer represents the TCP game server
type GameServer struct {
	Addr     string
	Listener net.Listener
	Clients  map[string]*Client // map of username to client
	mutex    sync.Mutex
}

// NewServer creates a new game server
func NewServer(addr string) *GameServer {
	return &GameServer{
		Addr:    addr,
		Clients: make(map[string]*Client),
	}
}

// Start starts the server and begins listening for connections
func (s *GameServer) Start() error {
	var err error
	s.Listener, err = net.Listen("tcp", s.Addr)
	if err != nil {
		return err
	}

	log.Printf("Server started on %s", s.Addr)

	for {
		conn, err := s.Listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v", err)
			continue
		}

		log.Printf("New connection from %s", conn.RemoteAddr())

		// Handle the client in a new goroutine
		go s.handleClient(conn)
	}
}

// Stop stops the server
func (s *GameServer) Stop() error {
	// Close all client connections
	s.mutex.Lock()
	for _, client := range s.Clients {
		client.Conn.Close()
	}
	s.mutex.Unlock()

	// Close the listener
	if s.Listener != nil {
		return s.Listener.Close()
	}
	return nil
}

// handleClient handles a client connection
func (s *GameServer) handleClient(conn net.Conn) {
	defer conn.Close()

	// Temporary client until login
	client := &Client{
		Conn: conn,
	}

	for {
		// Read message from client
		var message models.GenericMessage
		err := ReadMessage(conn, &message)
		if err != nil {
			log.Printf("Error reading message: %v", err)
			break
		}

		// Process message based on type
		switch message.Type {
		case models.MsgTypeLoginRequest:
			s.handleLogin(client, message.Payload)
		default:
			// Send error for unknown message type
			errorPayload := models.ErrorNotificationPayload{
				ErrorMessage: fmt.Sprintf("Unknown message type: %s", message.Type),
			}
			errorMsg := models.GenericMessage{
				Type:    models.MsgTypeErrorNotification,
				Payload: errorPayload,
			}
			err = WriteMessage(conn, errorMsg)
			if err != nil {
				log.Printf("Error sending error notification: %v", err)
			}
		}
	}

	// Remove client from server's clients map if they were logged in
	if client.Username != "" {
		s.mutex.Lock()
		delete(s.Clients, client.Username)
		s.mutex.Unlock()
		log.Printf("Client %s disconnected", client.Username)
	} else {
		log.Printf("Client disconnected before logging in")
	}
}

// handleLogin handles a login request
func (s *GameServer) handleLogin(client *Client, payload interface{}) {
	// Parse login payload
	loginPayload, ok := payload.(map[string]interface{})
	if !ok {
		sendError(client.Conn, "Invalid login payload")
		return
	}

	// Extract username
	username, ok := loginPayload["username"].(string)
	if !ok || username == "" {
		sendError(client.Conn, "Invalid username")
		return
	}

	// Check if username is already taken
	s.mutex.Lock()
	_, exists := s.Clients[username]
	if exists {
		s.mutex.Unlock()
		sendError(client.Conn, "Username already taken")
		return
	}

	// Update client info and add to clients map
	client.Username = username
	s.Clients[username] = client
	s.mutex.Unlock()

	log.Printf("Client logged in as %s", username)

	// Send login success response
	loginResponse := models.LoginResponsePayload{
		Success:  true,
		Message:  fmt.Sprintf("Successfully logged in as %s", username),
		PlayerID: username,
	}

	responseMsg := models.GenericMessage{
		Type:    models.MsgTypeLoginResponse,
		Payload: loginResponse,
	}

	err := WriteMessage(client.Conn, responseMsg)
	if err != nil {
		log.Printf("Error sending login response: %v", err)
	}

	// Check if we can start a game (if we have at least 2 clients)
	// This is just placeholder for now - we'll expand this in Phase 3
	s.mutex.Lock()
	if len(s.Clients) >= 2 {
		log.Printf("Have %d clients, enough to start a game", len(s.Clients))
		// In Phase 3, we'll implement game creation here
	}
	s.mutex.Unlock()
}

// sendError sends an error notification to the client
func sendError(conn net.Conn, errorMessage string) {
	errorPayload := models.ErrorNotificationPayload{
		ErrorMessage: errorMessage,
	}
	errorMsg := models.GenericMessage{
		Type:    models.MsgTypeErrorNotification,
		Payload: errorPayload,
	}
	err := WriteMessage(conn, errorMsg)
	if err != nil {
		log.Printf("Error sending error notification: %v", err)
	}
}
