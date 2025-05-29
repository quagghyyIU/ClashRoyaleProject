package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
	"tcr/internal/models"
	"tcr/internal/network"
	"time"
)

// Game state variables
var (
	gameMode         string
	opponentUsername string
	myPlayerState    models.PlayerState
	opponentState    models.PlayerState
	currentTurn      string
	lastActionLog    string
	// Store detailed game state
	myTroops         []models.TroopState
	myKingTower      models.TowerState
	myGuardTower1    models.TowerState
	myGuardTower2    models.TowerState
	enemyKingTower   models.TowerState
	enemyGuardTower1 models.TowerState
	enemyGuardTower2 models.TowerState
	// Authentication status flags
	loginSuccess        bool
	registrationSuccess bool
	authErrorMessage    string
)

func init() {
	// Initialize with empty data to avoid nil pointer errors
	myPlayerState = models.PlayerState{
		Username:                "You",
		KingTower:               models.TowerState{Type: "KING"},
		GuardTower1:             models.TowerState{Type: "GUARD1"},
		GuardTower2:             models.TowerState{Type: "GUARD2"},
		Troops:                  make([]models.TroopState, 0),
		Level:                   1,
		CurrentEXP:              0,
		RequiredEXPForNextLevel: 100,
		CurrentMana:             0,
		MaxMana:                 10,
	}
	opponentState = models.PlayerState{
		Username:                "Opponent",
		KingTower:               models.TowerState{Type: "KING"},
		GuardTower1:             models.TowerState{Type: "GUARD1"},
		GuardTower2:             models.TowerState{Type: "GUARD2"},
		Troops:                  make([]models.TroopState, 0),
		Level:                   1,
		CurrentEXP:              0,
		RequiredEXPForNextLevel: 100,
		CurrentMana:             0,
		MaxMana:                 10,
	}
	myKingTower = models.TowerState{
		Type:      "KING",
		CurrentHP: 0,
		MaxHP:     0,
		Attack:    0,
		Defense:   0,
		Destroyed: false,
	}
	myGuardTower1 = models.TowerState{
		Type:      "GUARD1",
		CurrentHP: 0,
		MaxHP:     0,
		Attack:    0,
		Defense:   0,
		Destroyed: false,
	}
	myGuardTower2 = models.TowerState{
		Type:      "GUARD2",
		CurrentHP: 0,
		MaxHP:     0,
		Attack:    0,
		Defense:   0,
		Destroyed: false,
	}
	enemyKingTower = models.TowerState{
		Type:      "KING",
		CurrentHP: 0,
		MaxHP:     0,
		Attack:    0,
		Defense:   0,
		Destroyed: false,
	}
	enemyGuardTower1 = models.TowerState{
		Type:      "GUARD1",
		CurrentHP: 0,
		MaxHP:     0,
		Attack:    0,
		Defense:   0,
		Destroyed: false,
	}
	enemyGuardTower2 = models.TowerState{
		Type:      "GUARD2",
		CurrentHP: 0,
		MaxHP:     0,
		Attack:    0,
		Defense:   0,
		Destroyed: false,
	}
	myTroops = make([]models.TroopState, 0)
}

func main() {
	// Define command line flags
	addr := flag.String("addr", "localhost:8080", "Server address to connect to (host:port)")
	flag.Parse()

	fmt.Println("TCR Client - Phase 3")
	fmt.Println("===================")
	fmt.Println("Welcome to Text-based Clash Royale!")
	fmt.Println("Please log in or register to play.")
	fmt.Println("===================")

	// Create the client
	client := network.NewClient(*addr)

	// Connect to the server
	fmt.Printf("Connecting to server at %s...\n", *addr)
	err := client.Connect()
	if err != nil {
		fmt.Printf("Error connecting to server: %v\n", err)
		return
	}

	fmt.Println("Connected to server!")

	// Set up reader for user input
	reader := bufio.NewReader(os.Stdin)

	// Start goroutine to handle messages from server
	stopCh := make(chan struct{})
	go handleServerMessages(client, stopCh)

	// Authentication loop
	authenticated := false
	for !authenticated && client.Connected {
		// Reset authentication flags
		loginSuccess = false
		registrationSuccess = false
		authErrorMessage = ""

		// Ask user if they want to register or login
		displayAuthPrompt()
		choiceStr, _ := reader.ReadString('\n')
		choiceStr = strings.TrimSpace(choiceStr)

		// Get username and password
		fmt.Print("Enter your username: ")
		username, _ := reader.ReadString('\n')
		username = strings.TrimSpace(username)

		fmt.Print("Enter your password: ")
		password, _ := reader.ReadString('\n')
		password = strings.TrimSpace(password)

		// Process login or registration
		if choiceStr == "2" {
			// Register new account
			err = client.Register(username, password)
			if err != nil {
				fmt.Printf("Error sending registration request: %v\n", err)
				continue
			}

			// Wait for registration response
			waitTime := 0
			for waitTime < 3 && !registrationSuccess && authErrorMessage == "" {
				time.Sleep(1 * time.Second)
				waitTime++
			}

			if !registrationSuccess {
				// If authErrorMessage is already set by a server response, don't print the generic timeout
				if authErrorMessage == "" {
					fmt.Println("\n⚠️  Registration timed out. Please try again.")
				} else {
					// Server already provided an error message via handleRegisterResponse or handleErrorNotification
					// fmt.Printf("\n⚠️  REGISTRATION FAILED: %s\n", authErrorMessage) // This line is now effectively handled by the handlers
				}
				continue
			}

			// Now login with the new account
			fmt.Println("\n✅ Registration successful! Logging in with your new account...")
			err = client.Login(username, password)
			if err != nil {
				fmt.Printf("Error sending login request: %v\n", err)
				continue
			}
		} else {
			// Login with existing account
			err = client.Login(username, password)
			if err != nil {
				fmt.Printf("Error sending login request: %v\n", err)
				continue
			}
		}

		// Wait for login response
		waitTime := 0
		for waitTime < 3 && !loginSuccess && authErrorMessage == "" {
			time.Sleep(1 * time.Second)
			waitTime++
		}

		if !loginSuccess {
			// If authErrorMessage is already set by a server response, don't print the generic timeout
			if authErrorMessage == "" {
				fmt.Println("\n⚠️  Login timed out. Please try again.")
			} else {
				// Server already provided an error message via handleLoginResponse or handleErrorNotification
				// fmt.Printf("\n⚠️  LOGIN FAILED: %s\n", authErrorMessage) // This line is now effectively handled by the handlers
			}
			continue
		}

		authenticated = true
		// fmt.Println("\n✅ Successfully authenticated!") // Removed as handleLoginResponse already prints a success message
		fmt.Println("Entering lobby - waiting for another player to connect...")

		// Now that user is authenticated, display available commands
		fmt.Println("\n=== Available Commands ===")
		fmt.Println("Commands available in lobby:")
		fmt.Println("  help - Show this help information")
		fmt.Println("  quit - Exit the game")
		fmt.Println("Commands available in game:")
		fmt.Println("  d <troop_name> - Deploy a troop (auto-targets enemy towers in sequence)")
		fmt.Println("  status - Display current game status")
		fmt.Println("  help - Display help information")
		fmt.Println("  quit - Exit the game")
		fmt.Println("========================")
	}

	// Check if we're still connected after authentication attempts
	if !client.Connected {
		fmt.Println("\n❌ Disconnected from server. Please restart the client and try again.")
		return
	}

	// Main input loop
	for client.Connected {
		var input string
		if !client.InGame { // In Lobby
			fmt.Print("> ")
			input, _ = reader.ReadString('\n')

			// If the game started while we were waiting for lobby input,
			// immediately re-evaluate client.InGame at the top of the loop.
			if client.InGame {
				continue
			}

			input = strings.TrimSpace(input)

			if input == "quit" || input == "exit" {
				break
			}
			if input == "help" {
				fmt.Println("\n==============================================")
				fmt.Println("📋 LOBBY COMMANDS 📋")
				fmt.Println("==============================================")
				fmt.Println("  help   - Show this help information")
				fmt.Println("  quit   - Exit the game")
				fmt.Println("")
				fmt.Println("You are in the lobby waiting for a game to start.")
				fmt.Println("Please wait for another player to connect...")
				fmt.Println("==============================================\n")
			} else {
				fmt.Println("Waiting for a game to start... (type 'help' for lobby commands or 'quit' to exit)")
			}
			continue // Loop back for next lobby input
		}

		// --- In Game ---
		// Refresh hand/target info if it's our turn, right before prompting
		if client.MyTurn {
			// displayPlayerHandAndTargetInfo(&myPlayerState) // Moved to handleTurnNotification or specific command handlers
			fmt.Printf("Your turn - Enter command (d <troop_name>, status, help, quit): ")
		} else {
			fmt.Print("(Waiting for opponent... Type status, help, or quit): ")
		}
		input, _ = reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "quit" || input == "exit" {
			break
		}

		if input == "status" {
			displayGameStatus(client, &myPlayerState, &opponentState, currentTurn, opponentUsername)
			// displayPlayerHandAndTargetInfo() // displayGameStatus might call this if it's our turn
			continue
		}

		if input == "help" {
			displayHelp()
			// displayPlayerHandAndTargetInfo() // Help shouldn't necessarily redisplay hand
			continue
		}

		// Commands below are only processed if it's the player's turn
		if !client.MyTurn {
			fmt.Println("It's not your turn. Type 'status', 'help', or 'quit'.")
			continue
		}

		// Parse and execute command (Player's Turn Only)
		parts := strings.Split(input, " ")
		if len(parts) < 1 {
			continue
		}

		switch parts[0] {
		case "d", "deploy":
			if len(parts) < 2 {
				fmt.Println("Usage: d <troop_name>")
				continue
			}

			troopName := parts[1]

			// Automatically determine the target tower
			var targetTowerID string

			// For Queen, we don't need a target tower for the ability to work
			if troopName == "Queen" {
				// Any target ID will work for Queen since server ignores it for special troops
				// But we need to provide a syntactically correct ID
				targetTowerID = opponentUsername + "_KING"
			} else {
				// Auto-select target based on game rules
				// Guard Tower 1 must be destroyed before targeting Guard Tower 2 or King
				if !opponentState.GuardTower1.Destroyed {
					targetTowerID = opponentState.GuardTower1.ID
				} else if !opponentState.GuardTower2.Destroyed {
					targetTowerID = opponentState.GuardTower2.ID
				} else if !opponentState.KingTower.Destroyed {
					targetTowerID = opponentState.KingTower.ID
				} else {
					// All towers destroyed? This shouldn't happen as game should be over
					fmt.Println("Error: Can't find a valid target tower.")
					continue
				}

				// Make sure we have a valid ID
				if targetTowerID == "" {
					// Construct a plausible ID if we somehow don't have it from server
					if !opponentState.GuardTower1.Destroyed {
						targetTowerID = opponentUsername + "_GUARD1"
					} else if !opponentState.GuardTower2.Destroyed {
						targetTowerID = opponentUsername + "_GUARD2"
					} else {
						targetTowerID = opponentUsername + "_KING"
					}
				}
			}

			// Show clear feedback about what we're targeting
			var targetDesc string
			if strings.Contains(targetTowerID, "KING") {
				targetDesc = "King Tower"
			} else if strings.Contains(targetTowerID, "GUARD1") {
				targetDesc = "Guard Tower 1"
			} else if strings.Contains(targetTowerID, "GUARD2") {
				targetDesc = "Guard Tower 2"
			} else {
				targetDesc = targetTowerID // fallback
			}
			fmt.Printf("\n>>> DEPLOYING %s to attack %s <<<\n", strings.ToUpper(troopName), targetDesc)

			// Send deploy command
			err := client.DeployTroop(troopName, targetTowerID)
			if err != nil {
				fmt.Printf("Error sending deploy command: %v\n", err)
			}
			continue // Wait for server updates before re-prompting
		case "skip":
			fmt.Println("\n>>> SKIPPING TURN <<<\n")
			err := client.SendSkipTurnCommand() // We'll need to define this method in network/client.go
			if err != nil {
				fmt.Printf("Error sending skip command: %v\n", err)
			}
			continue // Wait for server updates
		default:
			fmt.Println("Unknown command. Type 'help' for list of commands.")
		}
	}

	// Signal message handler to stop
	close(stopCh)

	// Disconnect from server
	client.Disconnect()
	fmt.Println("Disconnected from server. Goodbye!")
}

// handleServerMessages handles messages received from the server
func handleServerMessages(client *network.GameClient, stopCh <-chan struct{}) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Recovered in handleServerMessages: %v\n", r)
			// Optionally, decide if client should be marked as disconnected or attempt to re-establish
		}
		fmt.Println("Message handler stopped.")
	}()

	for {
		select {
		case <-stopCh:
			return
		case message := <-client.MessageCh:
			// Process different message types
			switch message.Type {
			case models.MsgTypeLoginResponse:
				handleLoginResponse(message.Payload)
			case models.MsgTypeRegisterResponse:
				handleRegisterResponse(message.Payload)
			case models.MsgTypeErrorNotification:
				handleErrorNotification(message.Payload)
			case models.MsgTypeGameStartNotification:
				handleGameStartNotification(client, message.Payload)
			case models.MsgTypeGameStateUpdate:
				handleGameStateUpdate(client, message.Payload)
			case models.MsgTypeTurnNotification:
				handleTurnNotification(client, message.Payload)
			case models.MsgTypeActionResult:
				handleActionResult(client, message.Payload)
			case models.MsgTypeGameOverNotification:
				handleGameOverNotification(message.Payload)
			}
		case err := <-client.DisconnectCh:
			fmt.Printf("Disconnected from server: %v\n", err)
			return
		}
	}
}

// handleLoginResponse handles a login response from the server
func handleLoginResponse(payload interface{}) {
	// Parse login response payload
	responseMap, ok := payload.(map[string]interface{})
	if !ok {
		fmt.Println("Error parsing login response")
		return
	}

	// Extract success and message
	success, _ := responseMap["success"].(bool)
	message, _ := responseMap["message"].(string)

	if success {
		fmt.Printf("\n✅ Login successful: %s\n", message)
		loginSuccess = true
		authErrorMessage = "" // Clear any previous auth error on successful login
	} else {
		fmt.Printf("\n⚠️ Login failed: %s\n", message)
		authErrorMessage = message
		loginSuccess = false // Ensure loginSuccess is false
	}
}

// handleRegisterResponse handles a registration response from the server
func handleRegisterResponse(payload interface{}) {
	// Parse registration response payload
	responseMap, ok := payload.(map[string]interface{})
	if !ok {
		fmt.Println("Error parsing registration response")
		return
	}

	// Extract success and message
	success, _ := responseMap["success"].(bool)
	message, _ := responseMap["message"].(string)

	if success {
		fmt.Printf("\n✅ Registration successful: %s\n", message)
		registrationSuccess = true
		authErrorMessage = "" // Clear any previous auth error on successful registration
	} else {
		fmt.Printf("\n⚠️ Registration failed: %s\n", message)
		authErrorMessage = message
		registrationSuccess = false // Ensure registrationSuccess is false
	}
}

// handleErrorNotification handles an error notification from the server
// This can be for login, registration, or other general errors.
func handleErrorNotification(payload interface{}) {
	// Parse error notification payload
	errorMap, ok := payload.(map[string]interface{})
	if !ok {
		fmt.Println("Error parsing error notification")
		return
	}

	// Extract error message
	errorMessage, _ := errorMap["errorMessage"].(string)
	fmt.Printf("\n⚠️  Server message: %s\n", errorMessage)

	// Set error message for authentication process if it's an auth-related error
	// We infer it's auth-related if loginSuccess or registrationSuccess are currently being processed (i.e., not yet true)
	if !loginSuccess && !registrationSuccess {
		authErrorMessage = errorMessage
	}
}

// handleGameStartNotification handles a game start notification from the server
func handleGameStartNotification(c *network.GameClient, payload interface{}) {
	c.InGame = true
	fmt.Println("\n==============================================")
	fmt.Println("⚔️  GAME STARTING! ⚔️")
	fmt.Println("==============================================")

	gameStartMap, ok := payload.(map[string]interface{})
	if !ok {
		fmt.Println("Error parsing game start notification")
		return
	}

	// Extract opponent username and game mode
	opponentUsername, _ = gameStartMap["opponentUsername"].(string)
	gameMode, _ = gameStartMap["gameMode"].(string)

	// Parse your player info
	if pInfo, ok := gameStartMap["yourPlayerInfo"].(map[string]interface{}); ok {
		myPlayerState = parsePlayerState(pInfo, "You")
	} else {
		fmt.Println("Error parsing yourPlayerInfo from GameStartNotification")
	}

	fmt.Printf("You are playing against: %s\n", opponentUsername)
	fmt.Printf("Game Mode: %s\n", gameMode)
	// Initial game status display will be handled by the first GameStateUpdate
	// displayGameStatus(c, &myPlayerState, &opponentState, currentTurn, opponentUsername)
}

// parsePlayerState converts a map to models.PlayerState
func parsePlayerState(playerMap map[string]interface{}, defaultUsername string) models.PlayerState {
	ps := models.PlayerState{Username: defaultUsername} // Default username if not in map

	if username, ok := playerMap["username"].(string); ok {
		ps.Username = username
	}
	if kt, ok := playerMap["kingTower"].(map[string]interface{}); ok {
		ps.KingTower = parseTowerState(kt)
	}
	if gt1, ok := playerMap["guardTower1"].(map[string]interface{}); ok {
		ps.GuardTower1 = parseTowerState(gt1)
	}
	if gt2, ok := playerMap["guardTower2"].(map[string]interface{}); ok {
		ps.GuardTower2 = parseTowerState(gt2)
	}
	if troopsList, ok := playerMap["troops"].([]interface{}); ok {
		ps.Troops = make([]models.TroopState, len(troopsList))
		for i, t := range troopsList {
			if troopMap, ok := t.(map[string]interface{}); ok {
				ps.Troops[i] = parseTroopState(troopMap)
			}
		}
	}
	// Parse new fields for Enhanced TCR
	if level, ok := playerMap["level"].(float64); ok { // JSON numbers are often float64
		ps.Level = int(level)
	}
	if currentEXP, ok := playerMap["currentEXP"].(float64); ok {
		ps.CurrentEXP = int(currentEXP)
	}
	if requiredEXP, ok := playerMap["requiredEXPForNextLevel"].(float64); ok {
		ps.RequiredEXPForNextLevel = int(requiredEXP)
	}
	if currentMana, ok := playerMap["currentMana"].(float64); ok {
		ps.CurrentMana = int(currentMana)
	}
	if maxMana, ok := playerMap["maxMana"].(float64); ok {
		ps.MaxMana = int(maxMana)
	}

	return ps
}

func parseTowerState(towerMap map[string]interface{}) models.TowerState {
	ts := models.TowerState{}
	if id, ok := towerMap["id"].(string); ok {
		ts.ID = id
	}
	if typ, ok := towerMap["type"].(string); ok {
		ts.Type = typ
	}
	if hp, ok := towerMap["currentHP"].(float64); ok { // JSON numbers are float64
		ts.CurrentHP = int(hp)
	}
	if maxHp, ok := towerMap["maxHP"].(float64); ok {
		ts.MaxHP = int(maxHp)
	}
	if atk, ok := towerMap["attack"].(float64); ok {
		ts.Attack = int(atk)
	}
	if def, ok := towerMap["defense"].(float64); ok {
		ts.Defense = int(def)
	}
	if destroyed, ok := towerMap["destroyed"].(bool); ok {
		ts.Destroyed = destroyed
	}
	return ts
}

func parseTroopState(troopMap map[string]interface{}) models.TroopState {
	trs := models.TroopState{}
	if name, ok := troopMap["name"].(string); ok {
		trs.Name = name
	}
	if hp, ok := troopMap["hp"].(float64); ok {
		trs.HP = int(hp)
	}
	if atk, ok := troopMap["attack"].(float64); ok {
		trs.Attack = int(atk)
	}
	if def, ok := troopMap["defense"].(float64); ok {
		trs.Defense = int(def)
	}
	// Attempt to parse ManaCost - this depends on server sending it in TroopState
	if manaCost, ok := troopMap["manaCost"].(float64); ok {
		trs.ManaCost = int(manaCost)
	}
	return trs
}

// handleGameStateUpdate handles a game state update from the server
func handleGameStateUpdate(c *network.GameClient, payload interface{}) {
	gameStateMap, ok := payload.(map[string]interface{})
	if !ok {
		fmt.Println("Error parsing game state update")
		return
	}

	// Update current turn and last action log
	if ct, ok := gameStateMap["currentTurn"].(string); ok {
		currentTurn = ct
		if currentTurn == myPlayerState.Username { // Compare with updated myPlayerState.Username
			c.MyTurn = true
		} else {
			c.MyTurn = false
		}
	}
	if lal, ok := gameStateMap["lastActionLog"].(string); ok {
		lastActionLog = lal
		fmt.Printf("\n--- Server Log: %s ---\n", lastActionLog)
	}

	// Update player states
	if pA, ok := gameStateMap["playerA"].(map[string]interface{}); ok {
		// Determine if playerA is me or opponent based on Username
		usernameA, _ := pA["username"].(string)
		if usernameA == myPlayerState.Username {
			myPlayerState = parsePlayerState(pA, myPlayerState.Username)
		} else {
			opponentState = parsePlayerState(pA, opponentUsername)
		}
	}
	if pB, ok := gameStateMap["playerB"].(map[string]interface{}); ok {
		usernameB, _ := pB["username"].(string)
		if usernameB == myPlayerState.Username {
			myPlayerState = parsePlayerState(pB, myPlayerState.Username)
		} else {
			opponentState = parsePlayerState(pB, opponentUsername)
		}
	}

	// Display updated game status
	fmt.Println("\n--- Game State Updated ---")
	displayGameStatus(c, &myPlayerState, &opponentState, currentTurn, opponentUsername)

	// If it's my turn now, the TurnNotification handler will display hand and prompt.
	// If it's not my turn, or game is over, display appropriate message.
	if !c.MyTurn && !c.GameOver {
		fmt.Print("(Waiting for opponent... Type status, help, or quit): ")
	} else if c.GameOver {
		// Game over message is handled by handleGameOverNotification
		// We might want a generic prompt here or nothing if quit is the only option
		fmt.Print("> ") // Generic prompt after game over
	}
	// The prompt for action when it is our turn is now handled by handleTurnNotification
	// or if the user types a command that doesn't end their turn (like a failed deploy).
}

// handleTurnNotification handles a turn notification from the server
func handleTurnNotification(client *network.GameClient, payload interface{}) {
	turnNotifMap, ok := payload.(map[string]interface{})
	if !ok {
		fmt.Println("Error parsing turn notification")
		return
	}

	if ctUser, ok := turnNotifMap["currentTurnUsername"].(string); ok {
		currentTurn = ctUser // Update currentTurn
		fmt.Printf("\n--- It's now %s's turn. ---\n", currentTurn)
		if currentTurn == myPlayerState.Username { // Compare with updated myPlayerState.Username
			client.MyTurn = true
			// Display hand and prompt only if game is not over
			if !client.GameOver {
				displayPlayerHandAndTargetInfo(&myPlayerState)
				fmt.Print("Your turn - Enter command (d <troop_name>, status, help, quit): ")
			}
		} else {
			client.MyTurn = false
			if !client.GameOver {
				fmt.Print("(Waiting for opponent... Type status, help, or quit): ")
			}
		}
	}
}

func displayPlayerHandAndTargetInfo(player *models.PlayerState) {
	fmt.Printf("\n--- Mana: %d/%d ---\n", player.CurrentMana, player.MaxMana)
	fmt.Println("--- Your Hand ---")
	if len(player.Troops) == 0 {
		fmt.Println("  Your hand is empty!")
	} else {
		for _, troop := range player.Troops {
			// Attempt to display ManaCost - requires server to send it in TroopState
			if troop.ManaCost > 0 {
				fmt.Printf("  - %s (ATK:%d DEF:%d HP:%d Mana:%d)\n", troop.Name, troop.Attack, troop.Defense, troop.HP, troop.ManaCost)
			} else {
				fmt.Printf("  - %s (ATK:%d DEF:%d HP:%d)\n", troop.Name, troop.Attack, troop.Defense, troop.HP)
			}
		}
	}
	fmt.Println("-----------------")

	// Display target info (simplified, relies on auto-targeting logic in main loop)
	// This could be enhanced to show specific target details based on opponentState
	fmt.Println("--- Target Info ---")
	if !opponentState.GuardTower1.Destroyed {
		fmt.Printf("  Next auto-target: Opponent's Guard Tower 1 (ID: %s)\n", opponentState.GuardTower1.ID)
	} else if !opponentState.GuardTower2.Destroyed {
		fmt.Printf("  Next auto-target: Opponent's Guard Tower 2 (ID: %s)\n", opponentState.GuardTower2.ID)
	} else if !opponentState.KingTower.Destroyed {
		fmt.Printf("  Next auto-target: Opponent's King Tower (ID: %s)\n", opponentState.KingTower.ID)
	} else {
		fmt.Println("  All opponent towers destroyed or game is over.")
	}
	fmt.Println("-------------------")
}

// handleActionResult handles an action result from the server
func handleActionResult(client *network.GameClient, payload interface{}) {
	// Parse action result payload
	resultMap, ok := payload.(map[string]interface{})
	if !ok {
		fmt.Println("Error parsing action result")
		return
	}

	// Extract result information
	success, _ := resultMap["success"].(bool)
	message, _ := resultMap["message"].(string)

	// Print result
	if success {
		fmt.Printf("Action successful: %s\n", message)
		// If action was successful, MyTurn will be updated by a subsequent TurnNotification or GameStateUpdate
		// No need to set client.MyTurn here.
	} else {
		fmt.Printf("Action failed: %s\n", message)
		fmt.Println("Please try again.")
		// If the action failed for a reason that doesn't end the turn (e.g. invalid troop),
		// ensure it's still the player's turn so they can retry.
		client.MyTurn = true
	}
}

// handleGameOverNotification handles a game over notification from the server
func handleGameOverNotification(payload interface{}) {
	gameOverMap, ok := payload.(map[string]interface{})
	if !ok {
		fmt.Println("Error parsing game over notification")
		return
	}
	winner, _ := gameOverMap["winnerUsername"].(string)
	reason, _ := gameOverMap["reason"].(string)

	fmt.Println("\n==============================================")
	fmt.Println("GAME OVER!")
	if winner != "" && winner != "DRAW" {
		fmt.Printf("Winner: %s\n", winner)
	} else if winner == "DRAW" {
		fmt.Println("Result: It's a DRAW!")
	}
	fmt.Printf("Reason: %s\n", reason)
	fmt.Println("==============================================")
	fmt.Println("Thank you for playing! You can type 'quit' to exit.")
	// Set a flag to stop prompting for turns or actions.
	// This should be handled by the main loop checking client.Connected and client.GameOver (if we add such a flag)
	// For now, client.MyTurn will be false if game over notification is processed after a turn notification.
	// A more robust solution is a specific client.GameOver flag.
}

// displayGameStatus displays the current game status in a more readable format
func displayGameStatus(c *network.GameClient, me *models.PlayerState, opp *models.PlayerState, turnUser string, oppUser string) {
	fmt.Println("\n==============================================")
	fmt.Println("           GAME STATUS")
	fmt.Println("==============================================")
	fmt.Printf("Current Turn: %s\n", turnUser)
	if gameMode != "" { // Display game mode if known
		fmt.Printf("Game Mode: %s\n", gameMode)
	}
	if lastActionLog != "" {
		fmt.Printf("Last Action: %s\n", lastActionLog)
	}
	fmt.Println("----------------------------------------------")

	// Display Your Info (me)
	fmt.Printf("YOUR INFO (%s):\n", me.Username)
	fmt.Printf("  Level: %d\n", me.Level)
	fmt.Printf("  EXP: %d / %d\n", me.CurrentEXP, me.RequiredEXPForNextLevel)
	fmt.Printf("  Mana: %d / %d\n", me.CurrentMana, me.MaxMana)
	fmt.Println("  Towers:")
	fmt.Printf("    - King Tower   (ID: %s): HP=%d/%d %s\n", me.KingTower.ID, me.KingTower.CurrentHP, me.KingTower.MaxHP, formatDestroyedStatus(me.KingTower.Destroyed))
	fmt.Printf("    - Guard Tower 1 (ID: %s): HP=%d/%d %s\n", me.GuardTower1.ID, me.GuardTower1.CurrentHP, me.GuardTower1.MaxHP, formatDestroyedStatus(me.GuardTower1.Destroyed))
	fmt.Printf("    - Guard Tower 2 (ID: %s): HP=%d/%d %s\n", me.GuardTower2.ID, me.GuardTower2.CurrentHP, me.GuardTower2.MaxHP, formatDestroyedStatus(me.GuardTower2.Destroyed))
	// fmt.Println("  Hand:") // Hand info will be shown by displayPlayerHandAndTargetInfo when it's player's turn
	// if len(me.Troops) == 0 {
	// 	fmt.Println("    Your hand is empty!")
	// } else {
	// 	for _, troop := range me.Troops {
	// 		if troop.ManaCost > 0 { // Check if ManaCost is available
	// 			fmt.Printf("    - %s (ATK:%d DEF:%d HP:%d Mana:%d)\n", troop.Name, troop.Attack, troop.Defense, troop.HP, troop.ManaCost)
	// 		} else {
	// 			fmt.Printf("    - %s (ATK:%d DEF:%d HP:%d)\n", troop.Name, troop.Attack, troop.Defense, troop.HP)
	// 		}
	// 	}
	// }
	fmt.Println("----------------------------------------------")

	// Display Opponent Info (opp)
	fmt.Printf("OPPONENT INFO (%s):\n", oppUser) // Use oppUser which is opponentUsername
	fmt.Printf("  Level: %d\n", opp.Level)       // Display opponent's level
	// Opponent's EXP and Mana are not typically shown, but towers are.
	fmt.Println("  Towers:")
	fmt.Printf("    - King Tower   (ID: %s): HP=%d/%d %s\n", opp.KingTower.ID, opp.KingTower.CurrentHP, opp.KingTower.MaxHP, formatDestroyedStatus(opp.KingTower.Destroyed))
	fmt.Printf("    - Guard Tower 1 (ID: %s): HP=%d/%d %s\n", opp.GuardTower1.ID, opp.GuardTower1.CurrentHP, opp.GuardTower1.MaxHP, formatDestroyedStatus(opp.GuardTower1.Destroyed))
	fmt.Printf("    - Guard Tower 2 (ID: %s): HP=%d/%d %s\n", opp.GuardTower2.ID, opp.GuardTower2.CurrentHP, opp.GuardTower2.MaxHP, formatDestroyedStatus(opp.GuardTower2.Destroyed))
	// We don't usually show opponent's hand
	fmt.Println("==============================================")
}

// formatDestroyedStatus returns a string indicating if a tower is destroyed
func formatDestroyedStatus(destroyed bool) string {
	if destroyed {
		return "(DESTROYED ☠️)"
	}
	return ""
}

// displayHelp displays the help information
func displayHelp() {
	fmt.Println("\n==============================================")
	fmt.Println("📋 GAME COMMANDS 📋")
	fmt.Println("==============================================")
	fmt.Println("Game Actions:")
	fmt.Println("  d <troop_name> - Deploy a troop to attack")
	fmt.Println("    - Auto-targets enemy towers in sequence")
	fmt.Println("    - Example: d Knight (deploys Knight to attack)")
	fmt.Println("    - Example: d Queen (deploys Queen to heal your lowest HP tower)")
	fmt.Println("  skip           - Skip your turn and gain bonus mana")
	fmt.Println("")
	fmt.Println("Information Commands:")
	fmt.Println("  status - Display detailed game status (towers, troops, etc.)")
	fmt.Println("  help   - Display this help information")
	fmt.Println("")
	fmt.Println("System Commands:")
	fmt.Println("  quit   - Exit the game")
	fmt.Println("==============================================")
	fmt.Println("Commands available IN GAME:")
	fmt.Println("----------------------------------------------")
	fmt.Println("  d <troop_name> - Deploy a troop (auto-targets enemy towers in sequence)")
	fmt.Println("                   Example: d Pawn")
	fmt.Println("                   (Queen will automatically heal your lowest HP tower)")
	fmt.Println("  skip           - Skip your turn and gain bonus mana")
	fmt.Println("  status         - Display current game status")
	fmt.Println("  help           - Display this help information")
	fmt.Println("  quit           - Forfeit the current game and exit")
	fmt.Println("==============================================")
}

// displayAuthPrompt displays the authentication options to the user
func displayAuthPrompt() {
	fmt.Println("\n==== Authentication Required ====")
	fmt.Println("1) Login with existing account")
	fmt.Println("2) Register a new account")
	fmt.Println("==============================")
	fmt.Print("Enter your choice (1/2): ")
}
