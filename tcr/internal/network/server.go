package network

import (
	"fmt"
	"log"
	"net"
	"sync"
	"tcr/internal/game"
	"tcr/internal/models"
	"tcr/internal/storage"
)

// Client represents a connected client
type Client struct {
	Username string
	Conn     net.Conn
	PlayerID string
	InGame   bool
}

// GameSession represents a game session between two clients
type GameSession struct {
	GameEngine *game.GameSession
	PlayerA    *Client
	PlayerB    *Client
}

// GameServer represents the TCP game server
type GameServer struct {
	Addr          string
	Listener      net.Listener
	Clients       map[string]*Client      // map of username to client
	GameSessions  map[string]*GameSession // map of session ID to game session
	WaitingPlayer *Client                 // Player waiting for a match
	TroopSpecs    []models.TroopSpec
	TowerSpecs    []models.TowerSpec
	JSONHandler   *storage.JSONHandler
	mutex         sync.Mutex
}

// NewServer creates a new game server
func NewServer(addr string, jsonHandler *storage.JSONHandler) *GameServer {
	return &GameServer{
		Addr:         addr,
		Clients:      make(map[string]*Client),
		GameSessions: make(map[string]*GameSession),
		JSONHandler:  jsonHandler,
	}
}

// Start starts the server and begins listening for connections
func (s *GameServer) Start() error {
	// Load game specifications
	var err error
	s.TroopSpecs, err = s.JSONHandler.LoadTroopSpecs()
	if err != nil {
		return fmt.Errorf("failed to load troop specs: %v", err)
	}

	s.TowerSpecs, err = s.JSONHandler.LoadTowerSpecs()
	if err != nil {
		return fmt.Errorf("failed to load tower specs: %v", err)
	}

	log.Printf("Loaded %d troop specs and %d tower specs", len(s.TroopSpecs), len(s.TowerSpecs))

	// Start listening
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
	// Create a new client
	client := &Client{
		Conn:   conn,
		InGame: false,
	}

	// Cleanup when this function exits
	defer func() {
		// Remove client from game session if in one
		s.mutex.Lock()
		if client.Username != "" {
			if client.InGame {
				// Handle game cleanup if in a game
				s.handlePlayerDisconnect(client)
			}
			delete(s.Clients, client.Username)
		}
		s.mutex.Unlock()

		// Close connection
		conn.Close()
		log.Printf("Connection from %s closed", conn.RemoteAddr())
	}()

	// Read and handle messages
	for {
		var message models.GenericMessage
		err := ReadMessage(conn, &message)
		if err != nil {
			log.Printf("Error reading message: %v", err)
			break
		}

		// Handle message based on type
		switch message.Type {
		case models.MsgTypeLoginRequest:
			s.handleLogin(client, message.Payload)
		case models.MsgTypeRegisterRequest:
			s.handleRegister(client, message.Payload)
		case models.MsgTypeDeployTroopCommand:
			s.handleDeployTroop(client, message.Payload)
		default:
			log.Printf("Unknown message type: %s", message.Type)
		}
	}
}

// handleRegister handles a registration request
func (s *GameServer) handleRegister(client *Client, payload interface{}) {
	// Parse register payload
	registerPayload, ok := payload.(map[string]interface{})
	if !ok {
		sendError(client.Conn, "Invalid registration payload")
		return
	}

	// Extract username and password
	username, ok := registerPayload["username"].(string)
	if !ok || username == "" {
		sendError(client.Conn, "Invalid username")
		return
	}

	password, ok := registerPayload["password"].(string)
	if !ok || password == "" {
		sendError(client.Conn, "Invalid password")
		return
	}

	// Check if username is already taken
	if s.JSONHandler.UserExists(username) {
		// Send registration failure response
		registerResponse := models.RegisterResponsePayload{
			Success: false,
			Message: "Username already taken",
		}

		responseMsg := models.GenericMessage{
			Type:    models.MsgTypeRegisterResponse,
			Payload: registerResponse,
		}

		WriteMessage(client.Conn, responseMsg)
		return
	}

	// Save user data
	userData := storage.UserData{
		Username: username,
		Password: password,
	}

	err := s.JSONHandler.SaveUserData(userData)
	if err != nil {
		log.Printf("Error saving user data: %v", err)
		sendError(client.Conn, "Failed to register user")
		return
	}

	log.Printf("New user registered: %s", username)

	// Send registration success response
	registerResponse := models.RegisterResponsePayload{
		Success: true,
		Message: fmt.Sprintf("Successfully registered as %s", username),
	}

	responseMsg := models.GenericMessage{
		Type:    models.MsgTypeRegisterResponse,
		Payload: registerResponse,
	}

	WriteMessage(client.Conn, responseMsg)
}

// handleLogin handles a login request
func (s *GameServer) handleLogin(client *Client, payload interface{}) {
	// Parse login payload
	loginPayload, ok := payload.(map[string]interface{})
	if !ok {
		sendError(client.Conn, "Invalid login payload")
		return
	}

	// Extract username and password
	username, ok := loginPayload["username"].(string)
	if !ok || username == "" {
		sendError(client.Conn, "Invalid username")
		return
	}

	password, ok := loginPayload["password"].(string)
	if !ok {
		// For backward compatibility, allow empty password
		password = ""
	}

	// Check if user exists
	if !s.JSONHandler.UserExists(username) {
		sendError(client.Conn, "User does not exist")
		return
	}

	// Validate password
	userData, err := s.JSONHandler.LoadUserData(username)
	if err != nil {
		log.Printf("Error loading user data: %v", err)
		sendError(client.Conn, "Login failed")
		return
	}

	if userData.Password != password {
		sendError(client.Conn, "Invalid password")
		return
	}

	// Check if username is already logged in
	s.mutex.Lock()
	_, exists := s.Clients[username]
	if exists {
		s.mutex.Unlock()
		sendError(client.Conn, "User already logged in")
		return
	}

	// Update client info and add to clients map
	client.Username = username
	client.PlayerID = username // Using username as player ID for now
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

	err = WriteMessage(client.Conn, responseMsg)
	if err != nil {
		log.Printf("Error sending login response: %v", err)
		return
	}

	// Try to match the player
	s.tryMatchPlayer(client)
}

// tryMatchPlayer attempts to match a player with another waiting player
func (s *GameServer) tryMatchPlayer(client *Client) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// If no waiting player, set this client as waiting
	if s.WaitingPlayer == nil {
		s.WaitingPlayer = client
		log.Printf("Player %s is waiting for a match", client.Username)

		// Send notification to client
		message := "Waiting for another player to join..."
		notificationMsg := models.GenericMessage{
			Type: models.MsgTypeErrorNotification, // Reusing error for notifications
			Payload: models.ErrorNotificationPayload{
				ErrorMessage: message,
			},
		}
		WriteMessage(client.Conn, notificationMsg)
		return
	}

	// Don't match player with themselves (if reconnecting)
	if s.WaitingPlayer.Username == client.Username {
		log.Printf("Player %s reconnected, still waiting for opponent", client.Username)
		return
	}

	// We have a match!
	playerA := s.WaitingPlayer
	playerB := client
	s.WaitingPlayer = nil

	log.Printf("Matched players %s and %s", playerA.Username, playerB.Username)

	// Create a new game session
	s.createGameSession(playerA, playerB)
}

// createGameSession creates a new game session between two players
func (s *GameServer) createGameSession(playerA, playerB *Client) {
	// Create game engine
	gameEngine := game.NewGameSession(playerA.Username, playerB.Username, s.TroopSpecs, s.TowerSpecs)

	// Create game session
	sessionID := fmt.Sprintf("%s_vs_%s", playerA.Username, playerB.Username)
	session := &GameSession{
		GameEngine: gameEngine,
		PlayerA:    playerA,
		PlayerB:    playerB,
	}

	// Update client states
	playerA.InGame = true
	playerB.InGame = true

	// Add session to map
	s.GameSessions[sessionID] = session

	log.Printf("Created game session %s", sessionID)

	// Send game start notifications to both players
	s.sendGameStartNotifications(session)
}

// sendGameStartNotifications sends game start notifications to both players
func (s *GameServer) sendGameStartNotifications(session *GameSession) {
	playerA := session.PlayerA
	playerB := session.PlayerB
	gameEngine := session.GameEngine

	// Prepare game start notification for player A
	playerAState := s.createPlayerState(gameEngine.GameState.PlayerA)
	playerBState := s.createPlayerState(gameEngine.GameState.PlayerB)

	// Notification for Player A
	playerANotification := models.GameStartNotificationPayload{
		OpponentUsername: playerB.Username,
		YourPlayerInfo:   playerAState,
		GameMode:         "SIMPLE", // For now, only Simple TCR is supported
	}

	playerAMsg := models.GenericMessage{
		Type:    models.MsgTypeGameStartNotification,
		Payload: playerANotification,
	}

	// Notification for Player B
	playerBNotification := models.GameStartNotificationPayload{
		OpponentUsername: playerA.Username,
		YourPlayerInfo:   playerBState,
		GameMode:         "SIMPLE", // For now, only Simple TCR is supported
	}

	playerBMsg := models.GenericMessage{
		Type:    models.MsgTypeGameStartNotification,
		Payload: playerBNotification,
	}

	// Send notifications
	err := WriteMessage(playerA.Conn, playerAMsg)
	if err != nil {
		log.Printf("Error sending game start notification to %s: %v", playerA.Username, err)
	}

	err = WriteMessage(playerB.Conn, playerBMsg)
	if err != nil {
		log.Printf("Error sending game start notification to %s: %v", playerB.Username, err)
	}

	// Send initial game state update to both players
	s.broadcastGameState(session, "")

	// Send turn notification to the first player
	s.sendTurnNotification(session)
}

// createPlayerState creates a PlayerState from a game.Player
func (s *GameServer) createPlayerState(player *game.Player) models.PlayerState {
	// Create tower states
	kingTowerState := models.TowerState{
		ID:        player.KingTower.ID,
		Type:      player.KingTower.Spec.Type,
		CurrentHP: player.KingTower.CurrentHP,
		MaxHP:     player.KingTower.Spec.BaseHP,
		Attack:    player.KingTower.CurrentATK,
		Defense:   player.KingTower.CurrentDEF,
		Destroyed: player.KingTower.Destroyed,
	}

	guardTower1State := models.TowerState{
		ID:        player.GuardTower1.ID,
		Type:      player.GuardTower1.Spec.Type,
		CurrentHP: player.GuardTower1.CurrentHP,
		MaxHP:     player.GuardTower1.Spec.BaseHP,
		Attack:    player.GuardTower1.CurrentATK,
		Defense:   player.GuardTower1.CurrentDEF,
		Destroyed: player.GuardTower1.Destroyed,
	}

	guardTower2State := models.TowerState{
		ID:        player.GuardTower2.ID,
		Type:      player.GuardTower2.Spec.Type,
		CurrentHP: player.GuardTower2.CurrentHP,
		MaxHP:     player.GuardTower2.Spec.BaseHP,
		Attack:    player.GuardTower2.CurrentATK,
		Defense:   player.GuardTower2.CurrentDEF,
		Destroyed: player.GuardTower2.Destroyed,
	}

	// Create troop states
	troopStates := make([]models.TroopState, len(player.Troops))
	for i, troop := range player.Troops {
		troopStates[i] = models.TroopState{
			Name:    troop.Spec.Name,
			HP:      troop.CurrentHP,
			Attack:  troop.CurrentATK,
			Defense: troop.CurrentDEF,
		}
	}

	return models.PlayerState{
		Username:    player.Username,
		KingTower:   kingTowerState,
		GuardTower1: guardTower1State,
		GuardTower2: guardTower2State,
		Troops:      troopStates,
	}
}

// broadcastGameState sends the current game state to both players
func (s *GameServer) broadcastGameState(session *GameSession, lastActionLog string) {
	gameEngine := session.GameEngine

	// Create game state update
	stateUpdate := models.GameStateUpdatePayload{
		PlayerA:       s.createPlayerState(gameEngine.GameState.PlayerA),
		PlayerB:       s.createPlayerState(gameEngine.GameState.PlayerB),
		CurrentTurn:   gameEngine.GameState.CurrentTurn,
		LastActionLog: lastActionLog,
	}

	gameStateMsg := models.GenericMessage{
		Type:    models.MsgTypeGameStateUpdate,
		Payload: stateUpdate,
	}

	// Send to both players
	WriteMessage(session.PlayerA.Conn, gameStateMsg)
	WriteMessage(session.PlayerB.Conn, gameStateMsg)
}

// sendTurnNotification sends a turn notification to the current player
func (s *GameServer) sendTurnNotification(session *GameSession) {
	gameEngine := session.GameEngine
	currentTurn := gameEngine.GameState.CurrentTurn

	// Create turn notification
	turnNotification := models.TurnNotificationPayload{
		CurrentTurnUsername: currentTurn,
	}

	turnMsg := models.GenericMessage{
		Type:    models.MsgTypeTurnNotification,
		Payload: turnNotification,
	}

	// Determine which client should receive the notification
	var currentPlayer *Client
	if currentTurn == session.PlayerA.Username {
		currentPlayer = session.PlayerA
	} else {
		currentPlayer = session.PlayerB
	}

	// Send notification
	WriteMessage(currentPlayer.Conn, turnMsg)
}

// getSessionForPlayer finds the game session for a given player
func (s *GameServer) getSessionForPlayer(client *Client) *GameSession {
	for _, session := range s.GameSessions {
		if session.PlayerA.Username == client.Username || session.PlayerB.Username == client.Username {
			return session
		}
	}
	return nil
}

// handleDeployTroop handles a deploy troop command
func (s *GameServer) handleDeployTroop(client *Client, payload interface{}) {
	// Check if player is in a game
	s.mutex.Lock()
	session := s.getSessionForPlayer(client)
	s.mutex.Unlock()

	if session == nil {
		sendError(client.Conn, "You are not in a game")
		return
	}

	// Parse deploy troop payload
	deployPayload, ok := payload.(map[string]interface{})
	if !ok {
		sendError(client.Conn, "Invalid deploy troop payload")
		return
	}

	// Extract troop name and target tower ID
	troopName, ok := deployPayload["troopName"].(string)
	if !ok || troopName == "" {
		sendError(client.Conn, "Invalid troop name")
		return
	}

	targetTowerID, ok := deployPayload["targetTowerID"].(string)
	if !ok || targetTowerID == "" {
		sendError(client.Conn, "Invalid target tower ID")
		return
	}

	// Pass command to the game engine
	actionResultMsg, success := session.GameEngine.DeployTroop(client.Username, troopName, targetTowerID)

	// Send action result to the player
	actionResult := models.ActionResultPayload{
		Success: success,
		Action:  fmt.Sprintf("Deploy %s to %s", troopName, targetTowerID),
		Message: actionResultMsg,
	}

	resultMsg := models.GenericMessage{
		Type:    models.MsgTypeActionResult,
		Payload: actionResult,
	}

	WriteMessage(client.Conn, resultMsg)

	// If successful, broadcast updated game state to both players
	if success {
		s.broadcastGameState(session, actionResultMsg)

		// Check if game is over
		if session.GameEngine.GameState.IsGameOver {
			s.handleGameOver(session)
			return
		}

		// If game continues, send turn notification
		s.sendTurnNotification(session)
	}
}

// handleGameOver handles game over events
func (s *GameServer) handleGameOver(session *GameSession) {
	winner := session.GameEngine.GameState.Winner

	// Create game over notification
	gameOverPayload := models.GameOverNotificationPayload{
		WinnerUsername: winner,
		Reason:         "King Tower destroyed",
	}

	gameOverMsg := models.GenericMessage{
		Type:    models.MsgTypeGameOverNotification,
		Payload: gameOverPayload,
	}

	// Send to both players
	WriteMessage(session.PlayerA.Conn, gameOverMsg)
	WriteMessage(session.PlayerB.Conn, gameOverMsg)

	// Clean up game session
	s.mutex.Lock()
	session.PlayerA.InGame = false
	session.PlayerB.InGame = false
	for id, gs := range s.GameSessions {
		if gs == session {
			delete(s.GameSessions, id)
			break
		}
	}
	s.mutex.Unlock()
}

// handlePlayerDisconnect handles a player disconnecting from a game
func (s *GameServer) handlePlayerDisconnect(client *Client) {
	session := s.getSessionForPlayer(client)
	if session == nil {
		return
	}

	// Determine the other player
	var otherPlayer *Client
	if client.Username == session.PlayerA.Username {
		otherPlayer = session.PlayerB
	} else {
		otherPlayer = session.PlayerA
	}

	// Send game over notification to the other player
	if otherPlayer != nil && otherPlayer.InGame {
		gameOverPayload := models.GameOverNotificationPayload{
			WinnerUsername: otherPlayer.Username,
			Reason:         fmt.Sprintf("%s disconnected", client.Username),
		}

		gameOverMsg := models.GenericMessage{
			Type:    models.MsgTypeGameOverNotification,
			Payload: gameOverPayload,
		}

		WriteMessage(otherPlayer.Conn, gameOverMsg)
		otherPlayer.InGame = false
	}

	// Clean up game session
	for id, gs := range s.GameSessions {
		if gs == session {
			delete(s.GameSessions, id)
			break
		}
	}
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
