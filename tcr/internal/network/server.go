package network

import (
	"fmt"
	"log"
	"net"
	"sync"
	"tcr/internal/game"
	"tcr/internal/models"
	"tcr/internal/shared"
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
		case models.MsgTypeSkipTurnCommand:
			s.handleSkipTurn(client, message.Payload)
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
	gameEngine := game.NewGameSession(playerA.Username, playerB.Username, s.TroopSpecs, s.TowerSpecs, s.JSONHandler)

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
		MaxHP:     player.KingTower.MaxHP,
		Attack:    player.KingTower.CurrentATK,
		Defense:   player.KingTower.CurrentDEF,
		Destroyed: player.KingTower.Destroyed,
	}

	guardTower1State := models.TowerState{
		ID:        player.GuardTower1.ID,
		Type:      player.GuardTower1.Spec.Type,
		CurrentHP: player.GuardTower1.CurrentHP,
		MaxHP:     player.GuardTower1.MaxHP,
		Attack:    player.GuardTower1.CurrentATK,
		Defense:   player.GuardTower1.CurrentDEF,
		Destroyed: player.GuardTower1.Destroyed,
	}

	guardTower2State := models.TowerState{
		ID:        player.GuardTower2.ID,
		Type:      player.GuardTower2.Spec.Type,
		CurrentHP: player.GuardTower2.CurrentHP,
		MaxHP:     player.GuardTower2.MaxHP,
		Attack:    player.GuardTower2.CurrentATK,
		Defense:   player.GuardTower2.CurrentDEF,
		Destroyed: player.GuardTower2.Destroyed,
	}

	// Create troop states
	troopStates := make([]models.TroopState, len(player.Troops))
	for i, troop := range player.Troops {
		troopStates[i] = models.TroopState{
			Name:     troop.Spec.Name,
			HP:       troop.CurrentHP,
			Attack:   troop.CurrentATK,
			Defense:  troop.CurrentDEF,
			ManaCost: troop.Spec.ManaCost,
		}
	}

	return models.PlayerState{
		Username:                player.Username,
		KingTower:               kingTowerState,
		GuardTower1:             guardTower1State,
		GuardTower2:             guardTower2State,
		Troops:                  troopStates,
		Level:                   player.Level,
		CurrentEXP:              player.CurrentEXP,
		RequiredEXPForNextLevel: player.RequiredEXPForNextLevel,
		CurrentMana:             player.CurrentMana,
		MaxMana:                 shared.MaxMana,
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

// handleSkipTurn handles a skip turn command
func (s *GameServer) handleSkipTurn(client *Client, payload interface{}) {
	// Check if player is in a game
	s.mutex.Lock()
	session := s.getSessionForPlayer(client)
	s.mutex.Unlock()

	if session == nil {
		sendError(client.Conn, "You are not in a game")
		return
	}

	// No payload to parse for skip turn, just the client's username is needed.

	// Pass command to the game engine
	actionResultMsg, success := session.GameEngine.SkipTurn(client.Username)

	// Send action result to the player who skipped
	actionResult := models.ActionResultPayload{
		Success: success,
		Action:  "Skip Turn",
		Message: actionResultMsg,
	}

	resultMsg := models.GenericMessage{
		Type:    models.MsgTypeActionResult, // Use ActionResult to inform the skipper
		Payload: actionResult,
	}

	WriteMessage(client.Conn, resultMsg)

	// If successful, broadcast updated game state to both players
	if success {
		// The actionResultMsg from SkipTurn already contains the log like "Player X skipped..."
		// This will be set as LastActionLog in GameEngine.SkipTurn, so broadcastGameState will pick it up.
		s.broadcastGameState(session, actionResultMsg)

		// Check if game is over (unlikely for a skip, but good practice)
		if session.GameEngine.GameState.IsGameOver {
			s.handleGameOver(session)
			return
		}

		// If game continues, send turn notification to the *next* player
		s.sendTurnNotification(session)
	}
}

// handleGameOver handles game over events
func (s *GameServer) handleGameOver(session *GameSession) {
	winnerUsername := session.GameEngine.GameState.Winner // Username of the winner
	var winningPlayer *game.Player
	var losingPlayer *game.Player

	// Determine player objects
	if winnerUsername == session.PlayerA.Username {
		winningPlayer = session.GameEngine.GameState.PlayerA
		losingPlayer = session.GameEngine.GameState.PlayerB
	} else if winnerUsername == session.PlayerB.Username {
		winningPlayer = session.GameEngine.GameState.PlayerB
		losingPlayer = session.GameEngine.GameState.PlayerA
	} else {
		// This case implies a draw or game ended without a clear winner from GameState
		// For now, we assume a winner is always set if King Tower is destroyed.
		// If draw mechanics are added, this section will need adjustment.
		log.Printf("Game Over for session %s. No clear winner or draw declared in GameState.Winner. Current EXP not changed for win/loss.", session.GameEngine.GameState.PlayerA.Username+"_vs_"+session.GameEngine.GameState.PlayerB.Username)
		// Save data for both players even if no win/loss EXP is awarded
		pA := session.GameEngine.GameState.PlayerA
		pB := session.GameEngine.GameState.PlayerB
		if err := s.JSONHandler.SavePlayerData(storage.PlayerProfile{
			Username: pA.Username, Level: pA.Level, CurrentEXP: pA.CurrentEXP, RequiredEXPForNextLevel: pA.RequiredEXPForNextLevel,
		}); err != nil {
			log.Printf("Error saving player data for %s: %v", pA.Username, err)
		}
		if err := s.JSONHandler.SavePlayerData(storage.PlayerProfile{
			Username: pB.Username, Level: pB.Level, CurrentEXP: pB.CurrentEXP, RequiredEXPForNextLevel: pB.RequiredEXPForNextLevel,
		}); err != nil {
			log.Printf("Error saving player data for %s: %v", pB.Username, err)
		}
		// Proceed with standard game over notification without win/loss EXP messages
	}

	if winningPlayer != nil { // If there was a winner
		log.Printf("Player %s won. Awarding %d EXP.", winnerUsername, shared.WinEXPReward)
		// winningPlayer.CurrentEXP += shared.WinEXPReward // This is now handled by GameEngine.HandleGameOver
		// levelUpMsgWinner := session.GameEngine.HandleExperienceAndLevelUp(winningPlayer) // This is also handled by GameEngine.HandleGameOver
		// if levelUpMsgWinner != "" {
		// 	log.Printf("Post-game level up for winner %s: %s", winnerUsername, levelUpMsgWinner)
		// 	// This message could be appended to the GameOverNotification or sent separately
		// }

		// GameEngine.HandleGameOver already calls HandleExperienceAndLevelUp which saves player data.
		// So, explicit saves here might be redundant if GameEngine.HandleGameOver is the sole authority.
		// However, the original handleGameOver in engine.go was designed to be called BY the network layer (like this one)
		// and engine.HandleGameOver itself doesn't award match EXP, it relies on the caller.

		// The GameEngine's HandleGameOver (which awards EXP and then calls HandleExperienceAndLevelUp for saving)
		// should be the one called if we want the engine to manage this. The current server.handleGameOver seems to duplicate some logic.

		// For now, assuming the server's handleGameOver is orchestrating and gameEngine.HandleGameOver was called by DeployTroop.
		// The main thing is that player data (winningPlayer, losingPlayer) should be saved after any EXP changes.
		// The call to GameEngine.HandleExperienceAndLevelUp (which saves) is inside GameEngine.HandleGameOver.
		// If GameEngine.DeployTroop calls GameEngine.HandleGameOver, then saves are done.
		// This server.handleGameOver might be for cases where the game ends not due to a troop deployment (e.g. disconnect).

		// Let's simplify: if gameEngine.HandleGameOver was called (e.g. from DeployTroop), data is saved.
		// If this server.handleGameOver is for other reasons (like disconnect), we save here.
		// The original gameEngine.HandleGameOver modified player stats and called its own HandleExperienceAndLevelUp.

		// Reconciling: game.HandleGameOver in engine.go *is* called by DeployTroop and does EXP and saving.
		// This network.handleGameOver should primarily be for sending notifications and session cleanup.
		// It should rely on the game.Player objects already having their EXP/Level updated by the engine.

		// Save winning player's already updated profile (updated by engine)
		if err := s.JSONHandler.SavePlayerData(storage.PlayerProfile{
			Username: winningPlayer.Username, Level: winningPlayer.Level, CurrentEXP: winningPlayer.CurrentEXP, RequiredEXPForNextLevel: winningPlayer.RequiredEXPForNextLevel,
		}); err != nil {
			log.Printf("Error saving player data for winner %s: %v", winningPlayer.Username, err)
		}

		// Save loser's data as well (their EXP might have changed from destroying units, and engine would have updated it)
		if losingPlayer != nil {
			if err := s.JSONHandler.SavePlayerData(storage.PlayerProfile{
				Username: losingPlayer.Username, Level: losingPlayer.Level, CurrentEXP: losingPlayer.CurrentEXP, RequiredEXPForNextLevel: losingPlayer.RequiredEXPForNextLevel,
			}); err != nil {
				log.Printf("Error saving player data for loser %s: %v", losingPlayer.Username, err)
			}
		}
	} // Add logic for DrawEXPReward if draw state is possible and distinct from no winner.

	// Create game over notification (original logic)
	gameOverPayload := models.GameOverNotificationPayload{
		WinnerUsername: winnerUsername,         // This remains the same
		Reason:         "King Tower destroyed", // Or other reason
	}

	gameOverMsg := models.GenericMessage{
		Type:    models.MsgTypeGameOverNotification,
		Payload: gameOverPayload,
	}

	// Send to both players
	if session.PlayerA != nil && session.PlayerA.Conn != nil {
		WriteMessage(session.PlayerA.Conn, gameOverMsg)
	}
	if session.PlayerB != nil && session.PlayerB.Conn != nil {
		WriteMessage(session.PlayerB.Conn, gameOverMsg)
	}

	// Clean up game session (original logic)
	s.mutex.Lock()
	if session.PlayerA != nil {
		session.PlayerA.InGame = false
	}
	if session.PlayerB != nil {
		session.PlayerB.InGame = false
	}
	// Use a safe way to identify the session before deleting
	// This assumes session IDs are formed by player names, which might need to be more robust
	var sessionID string
	if session.PlayerA != nil && session.PlayerB != nil {
		sessionID = session.PlayerA.Username + "_vs_" + session.PlayerB.Username
	} else if session.PlayerA != nil {
		// Fallback or handle case where PlayerB might be nil (e.g. disconnected earlier)
		sessionID = session.PlayerA.Username + "_vs_unknown"
	} else if session.PlayerB != nil {
		sessionID = "unknown_vs_" + session.PlayerB.Username
	}
	// else if both are nil, sessionID remains empty, delete might not work as expected

	if sessionID != "" {
		delete(s.GameSessions, sessionID)
		log.Printf("Cleaned up game session: %s", sessionID)
	} else {
		log.Printf("Could not determine session ID for cleanup. Session players: A=%v, B=%v", session.PlayerA, session.PlayerB)
	}
	s.mutex.Unlock()
}

// handlePlayerDisconnect handles a player disconnecting from a game
func (s *GameServer) handlePlayerDisconnect(client *Client) {
	session := s.getSessionForPlayer(client)
	if session == nil {
		log.Printf("Player %s disconnected, was not in an active game session.", client.Username)
		return
	}

	log.Printf("Player %s disconnected from game session.", client.Username)

	// Save the disconnecting player's data
	var disconnectedPlayerGameObj *game.Player
	if client.Username == session.GameEngine.GameState.PlayerA.Username {
		disconnectedPlayerGameObj = session.GameEngine.GameState.PlayerA
	} else if client.Username == session.GameEngine.GameState.PlayerB.Username {
		disconnectedPlayerGameObj = session.GameEngine.GameState.PlayerB
	}

	if disconnectedPlayerGameObj != nil {
		if err := s.JSONHandler.SavePlayerData(storage.PlayerProfile{
			Username: disconnectedPlayerGameObj.Username, Level: disconnectedPlayerGameObj.Level, CurrentEXP: disconnectedPlayerGameObj.CurrentEXP, RequiredEXPForNextLevel: disconnectedPlayerGameObj.RequiredEXPForNextLevel,
		}); err != nil {
			log.Printf("Error saving player data for disconnecting player %s: %v", disconnectedPlayerGameObj.Username, err)
		} else {
			log.Printf("Saved player data for disconnecting player %s (Level: %d, EXP: %d)", disconnectedPlayerGameObj.Username, disconnectedPlayerGameObj.Level, disconnectedPlayerGameObj.CurrentEXP)
		}
	} else {
		log.Printf("Could not find game object for disconnecting player %s to save data.", client.Username)
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
