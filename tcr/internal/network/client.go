package network

import (
	"fmt"
	"net"
	"tcr/internal/models"
)

// GameClient represents the TCP game client
type GameClient struct {
	Addr         string
	conn         net.Conn
	Username     string
	Connected    bool
	LoggedIn     bool
	PlayerID     string
	MessageCh    chan models.GenericMessage
	DisconnectCh chan error
}

// NewClient creates a new game client
func NewClient(addr string) *GameClient {
	return &GameClient{
		Addr:         addr,
		Connected:    false,
		LoggedIn:     false,
		MessageCh:    make(chan models.GenericMessage, 10),
		DisconnectCh: make(chan error, 1),
	}
}

// Connect connects to the server
func (c *GameClient) Connect() error {
	var err error
	c.conn, err = net.Dial("tcp", c.Addr)
	if err != nil {
		return err
	}

	c.Connected = true

	// Start listening for messages
	go c.listen()

	return nil
}

// Disconnect disconnects from the server
func (c *GameClient) Disconnect() error {
	if c.conn != nil {
		err := c.conn.Close()
		c.Connected = false
		c.LoggedIn = false
		return err
	}
	return nil
}

// Login sends a login request to the server
func (c *GameClient) Login(username string) error {
	if !c.Connected {
		return fmt.Errorf("not connected to server")
	}

	c.Username = username

	// Prepare login request
	loginPayload := models.LoginRequestPayload{
		Username: username,
	}

	message := models.GenericMessage{
		Type:    models.MsgTypeLoginRequest,
		Payload: loginPayload,
	}

	// Send login request
	return WriteMessage(c.conn, message)
}

// listen listens for messages from the server
func (c *GameClient) listen() {
	defer func() {
		c.Connected = false
		c.LoggedIn = false
		c.DisconnectCh <- fmt.Errorf("disconnected from server")
	}()

	for {
		var message models.GenericMessage
		err := ReadMessage(c.conn, &message)
		if err != nil {
			return
		}

		// Process message based on type
		switch message.Type {
		case models.MsgTypeLoginResponse:
			c.handleLoginResponse(message.Payload)
		default:
			// Forward message to channel for processing by the main client
			c.MessageCh <- message
		}
	}
}

// handleLoginResponse handles a login response from the server
func (c *GameClient) handleLoginResponse(payload interface{}) {
	// Parse login response payload
	responseMap, ok := payload.(map[string]interface{})
	if !ok {
		return
	}

	// Extract success status
	success, ok := responseMap["success"].(bool)
	if !ok {
		return
	}

	// If login successful, update client state
	if success {
		c.LoggedIn = true

		// Extract player ID if available
		if playerID, ok := responseMap["playerId"].(string); ok {
			c.PlayerID = playerID
		}
	}

	// Forward message to channel for processing by the main client
	c.MessageCh <- models.GenericMessage{
		Type:    models.MsgTypeLoginResponse,
		Payload: payload,
	}
}
