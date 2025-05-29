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
	InGame       bool
	MyTurn       bool
	OpponentName string
	GameMode     string
	GameOver     bool
	MessageCh    chan models.GenericMessage
	DisconnectCh chan error
}

// NewClient creates a new game client
func NewClient(addr string) *GameClient {
	return &GameClient{
		Addr:         addr,
		Connected:    false,
		LoggedIn:     false,
		InGame:       false,
		MyTurn:       false,
		GameOver:     false,
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
		c.InGame = false
		c.MyTurn = false
		return err
	}
	return nil
}

// Login sends a login request to the server
func (c *GameClient) Login(username, password string) error {
	if !c.Connected {
		return fmt.Errorf("not connected to server")
	}

	c.Username = username

	// Prepare login request
	loginPayload := models.LoginRequestPayload{
		Username: username,
		Password: password,
	}

	message := models.GenericMessage{
		Type:    models.MsgTypeLoginRequest,
		Payload: loginPayload,
	}

	// Send login request
	return WriteMessage(c.conn, message)
}

// Register sends a registration request to the server
func (c *GameClient) Register(username, password string) error {
	if !c.Connected {
		return fmt.Errorf("not connected to server")
	}

	// Prepare registration request
	registerPayload := models.RegisterRequestPayload{
		Username: username,
		Password: password,
	}

	message := models.GenericMessage{
		Type:    models.MsgTypeRegisterRequest,
		Payload: registerPayload,
	}

	// Send registration request
	return WriteMessage(c.conn, message)
}

// DeployTroop sends a deploy troop command to the server
func (c *GameClient) DeployTroop(troopName, targetTowerID string) error {
	if !c.Connected || !c.InGame {
		return fmt.Errorf("not in game")
	}

	// Prepare deploy troop command
	deployPayload := models.DeployTroopCommandPayload{
		TroopName:     troopName,
		TargetTowerID: targetTowerID,
	}

	message := models.GenericMessage{
		Type:    models.MsgTypeDeployTroopCommand,
		Payload: deployPayload,
	}

	// Send deploy troop command
	return WriteMessage(c.conn, message)
}

// SendSkipTurnCommand sends a skip turn command to the server
func (c *GameClient) SendSkipTurnCommand() error {
	if !c.Connected || !c.InGame {
		return fmt.Errorf("not in game")
	}

	// No payload needed for skip, just the message type
	message := models.GenericMessage{
		Type:    models.MsgTypeSkipTurnCommand, // New message type
		Payload: nil,                           // No specific payload needed for skip
	}

	// Send skip turn command
	return WriteMessage(c.conn, message)
}

// listen listens for messages from the server
func (c *GameClient) listen() {
	defer func() {
		c.Connected = false
		c.LoggedIn = false
		c.InGame = false
		c.MyTurn = false
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
		case models.MsgTypeRegisterResponse:
			// Just forward to channel for display
		case models.MsgTypeGameStartNotification:
			c.handleGameStartNotification(message.Payload)
		case models.MsgTypeGameStateUpdate:
			c.handleGameStateUpdate(message.Payload)
		case models.MsgTypeTurnNotification:
			c.handleTurnNotification(message.Payload)
		case models.MsgTypeActionResult:
			// Just forward to channel for display
		case models.MsgTypeGameOverNotification:
			c.handleGameOverNotification(message.Payload)
		}

		// Forward all messages to channel for processing by the main client
		c.MessageCh <- message
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

	// Update client state based on login success
	if success {
		c.LoggedIn = true

		// Extract player ID if available
		if playerID, ok := responseMap["playerId"].(string); ok {
			c.PlayerID = playerID
		}
	} else {
		// Reset logged in state on failed login
		c.LoggedIn = false
		c.PlayerID = ""
	}
}

// handleGameStartNotification handles a game start notification from the server
func (c *GameClient) handleGameStartNotification(payload interface{}) {
	// Parse game start notification payload
	notifMap, ok := payload.(map[string]interface{})
	if !ok {
		return
	}

	// Extract opponent username and game mode
	if opponentName, ok := notifMap["opponentUsername"].(string); ok {
		c.OpponentName = opponentName
	}

	if gameMode, ok := notifMap["gameMode"].(string); ok {
		c.GameMode = gameMode
	}

	// Set in-game flag
	c.InGame = true
}

// handleGameStateUpdate handles a game state update from the server
func (c *GameClient) handleGameStateUpdate(payload interface{}) {
	// No state to update in the client structure, just forward to main for display
}

// handleTurnNotification handles a turn notification from the server
func (c *GameClient) handleTurnNotification(payload interface{}) {
	// Parse turn notification payload
	turnMap, ok := payload.(map[string]interface{})
	if !ok {
		return
	}

	// Extract current turn username
	currentTurn, ok := turnMap["currentTurnUsername"].(string)
	if !ok {
		return
	}

	// Check if it's this client's turn
	c.MyTurn = (currentTurn == c.Username)
}

// handleGameOverNotification handles a game over notification from the server
func (c *GameClient) handleGameOverNotification(payload interface{}) {
	// Reset game state
	c.InGame = false
	c.MyTurn = false
	c.GameOver = true
}
