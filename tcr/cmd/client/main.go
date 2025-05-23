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
				fmt.Println("==============================================")
			} else {
				fmt.Println("Waiting for a game to start... (type 'help' for lobby commands or 'quit' to exit)")
			}
			continue // Loop back for next lobby input
		}

		// --- In Game ---
		if client.MyTurn {
			fmt.Print("Your turn - Enter command (d <troop_name>, status, help, quit): ")
		} else {
			fmt.Print("(Waiting for opponent... Type status, help, or quit): ")
		}
		input, _ = reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "quit" || input == "exit" {
			break
		}

		if input == "status" {
			displayGameStatus(client)
			displayPlayerHandAndTargetInfo()
			continue
		}

		if input == "help" {
			displayHelp()
			displayPlayerHandAndTargetInfo()
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
				if !enemyGuardTower1.Destroyed {
					targetTowerID = enemyGuardTower1.ID
				} else if !enemyGuardTower2.Destroyed {
					targetTowerID = enemyGuardTower2.ID
				} else if !enemyKingTower.Destroyed {
					targetTowerID = enemyKingTower.ID
				} else {
					// All towers destroyed? This shouldn't happen as game should be over
					fmt.Println("Error: Can't find a valid target tower.")
					continue
				}

				// Make sure we have a valid ID
				if targetTowerID == "" {
					// Construct a plausible ID if we somehow don't have it from server
					if !enemyGuardTower1.Destroyed {
						targetTowerID = opponentUsername + "_GUARD1"
					} else if !enemyGuardTower2.Destroyed {
						targetTowerID = opponentUsername + "_GUARD2"
					} else {
						targetTowerID = opponentUsername + "_KING"
					}
				}
			}

			// Show clear feedback about what we're targeting
			var targetDesc string
			switch {
			case targetTowerID == enemyGuardTower1.ID || strings.HasSuffix(targetTowerID, "_GUARD1"):
				targetDesc = "Guard Tower 1"
			case targetTowerID == enemyGuardTower2.ID || strings.HasSuffix(targetTowerID, "_GUARD2"):
				targetDesc = "Guard Tower 2"
			case targetTowerID == enemyKingTower.ID || strings.HasSuffix(targetTowerID, "_KING"):
				targetDesc = "King Tower"
			default:
				targetDesc = targetTowerID
			}

			if troopName == "Queen" {
				fmt.Printf("Deploying Queen to heal your lowest HP tower\n")
			} else {
				fmt.Printf("Auto-targeting: Deploying %s to attack %s\n", troopName, targetDesc)
			}

			// Send deploy command
			err := client.DeployTroop(troopName, targetTowerID)
			if err != nil {
				fmt.Printf("Error sending deploy command: %v\n", err)
			}
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
				handleGameStartNotification(message.Payload)
			case models.MsgTypeGameStateUpdate:
				handleGameStateUpdate(client, message.Payload)
			case models.MsgTypeTurnNotification:
				handleTurnNotification(client, message.Payload)
			case models.MsgTypeActionResult:
				handleActionResult(message.Payload)
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
func handleGameStartNotification(payload interface{}) {
	// Parse game start notification payload
	notifMap, ok := payload.(map[string]interface{})
	if !ok {
		fmt.Println("Error parsing game start notification")
		return
	}

	// Extract opponent username and game mode
	opponentUsername, _ = notifMap["opponentUsername"].(string)
	gameMode, _ = notifMap["gameMode"].(string)

	// Extract player info
	playerInfoMap, ok := notifMap["yourPlayerInfo"].(map[string]interface{})
	if ok {
		// Store relevant player info
		username, _ := playerInfoMap["username"].(string)

		// Get towers
		if kingTower, ok := playerInfoMap["kingTower"].(map[string]interface{}); ok {
			myKingTower = parseTowerState(kingTower)
		}
		if guardTower1, ok := playerInfoMap["guardTower1"].(map[string]interface{}); ok {
			myGuardTower1 = parseTowerState(guardTower1)
		}
		if guardTower2, ok := playerInfoMap["guardTower2"].(map[string]interface{}); ok {
			myGuardTower2 = parseTowerState(guardTower2)
		}

		// Get troops
		if troops, ok := playerInfoMap["troops"].([]interface{}); ok {
			myTroops = make([]models.TroopState, 0, len(troops))
			for _, t := range troops {
				if troopMap, ok := t.(map[string]interface{}); ok {
					myTroops = append(myTroops, parseTroopState(troopMap))
				}
			}
		}

		fmt.Printf("\n===============================\n")
		fmt.Printf("🎮 GAME STARTED! 🎮\n")
		fmt.Printf("===============================\n")
		fmt.Printf("Game Mode: %s\n", gameMode)
		fmt.Printf("You: %s\n", username)
		fmt.Printf("Opponent: %s\n", opponentUsername)

		// Display initial state
		fmt.Println("\n=== Your Towers ===")
		fmt.Printf("King Tower: HP %d/%d, ATK %d, DEF %d\n",
			myKingTower.CurrentHP, myKingTower.MaxHP, myKingTower.Attack, myKingTower.Defense)
		fmt.Printf("Guard Tower 1: HP %d/%d, ATK %d, DEF %d\n",
			myGuardTower1.CurrentHP, myGuardTower1.MaxHP, myGuardTower1.Attack, myGuardTower1.Defense)
		fmt.Printf("Guard Tower 2: HP %d/%d, ATK %d, DEF %d\n",
			myGuardTower2.CurrentHP, myGuardTower2.MaxHP, myGuardTower2.Attack, myGuardTower2.Defense)

		fmt.Println("\n=== Your Available Troops ===")
		for _, troop := range myTroops {
			fmt.Printf("%s: HP %d, ATK %d, DEF %d\n",
				troop.Name, troop.HP, troop.Attack, troop.Defense)
		}

		fmt.Println("\nWaiting for the first turn...")
		fmt.Println("Type 'help' during the game to see available commands.")
	}
}

// parseTowerState parses a tower state from a map
func parseTowerState(towerMap map[string]interface{}) models.TowerState {
	id, _ := towerMap["id"].(string)
	towerType, _ := towerMap["type"].(string)
	currentHP, _ := towerMap["currentHP"].(float64)
	maxHP, _ := towerMap["maxHP"].(float64)
	attack, _ := towerMap["attack"].(float64)
	defense, _ := towerMap["defense"].(float64)
	destroyed, _ := towerMap["destroyed"].(bool)

	return models.TowerState{
		ID:        id,
		Type:      towerType,
		CurrentHP: int(currentHP),
		MaxHP:     int(maxHP),
		Attack:    int(attack),
		Defense:   int(defense),
		Destroyed: destroyed,
	}
}

// parseTroopState parses a troop state from a map
func parseTroopState(troopMap map[string]interface{}) models.TroopState {
	name, _ := troopMap["name"].(string)
	hp, _ := troopMap["hp"].(float64)
	attack, _ := troopMap["attack"].(float64)
	defense, _ := troopMap["defense"].(float64)

	return models.TroopState{
		Name:    name,
		HP:      int(hp),
		Attack:  int(attack),
		Defense: int(defense),
	}
}

// handleGameStateUpdate handles a game state update from the server
func handleGameStateUpdate(client *network.GameClient, payload interface{}) {
	// Parse game state update payload
	stateMap, ok := payload.(map[string]interface{})
	if !ok {
		fmt.Println("Error parsing game state update")
		return
	}

	// Extract current turn and last action log
	currentTurn, _ = stateMap["currentTurn"].(string)
	lastActionLog, _ = stateMap["lastActionLog"].(string)

	// Extract player states
	// Try to find our player and opponent in playerA or playerB
	playerA, hasPlayerA := stateMap["playerA"].(map[string]interface{})
	playerB, hasPlayerB := stateMap["playerB"].(map[string]interface{})

	if hasPlayerA && hasPlayerB {
		playerAUsername, _ := playerA["username"].(string)
		// Determine which is our player and which is opponent
		var myPlayerMap, enemyPlayerMap map[string]interface{}
		if playerAUsername == client.Username {
			myPlayerMap = playerA
			enemyPlayerMap = playerB
		} else {
			myPlayerMap = playerB
			enemyPlayerMap = playerA
		}

		// Update my player state
		if kingTower, ok := myPlayerMap["kingTower"].(map[string]interface{}); ok {
			myKingTower = parseTowerState(kingTower)
		}
		if guardTower1, ok := myPlayerMap["guardTower1"].(map[string]interface{}); ok {
			myGuardTower1 = parseTowerState(guardTower1)
		}
		if guardTower2, ok := myPlayerMap["guardTower2"].(map[string]interface{}); ok {
			myGuardTower2 = parseTowerState(guardTower2)
		}

		// Update enemy player state
		if kingTower, ok := enemyPlayerMap["kingTower"].(map[string]interface{}); ok {
			enemyKingTower = parseTowerState(kingTower)
		}
		if guardTower1, ok := enemyPlayerMap["guardTower1"].(map[string]interface{}); ok {
			enemyGuardTower1 = parseTowerState(guardTower1)
		}
		if guardTower2, ok := enemyPlayerMap["guardTower2"].(map[string]interface{}); ok {
			enemyGuardTower2 = parseTowerState(guardTower2)
		}

		// Update troops
		if troops, ok := myPlayerMap["troops"].([]interface{}); ok {
			myTroops = make([]models.TroopState, 0, len(troops))
			for _, t := range troops {
				if troopMap, ok := t.(map[string]interface{}); ok {
					myTroops = append(myTroops, parseTroopState(troopMap))
				}
			}
		}
	}

	// Log the update
	if lastActionLog != "" {
		fmt.Printf("\n>>> %s\n", lastActionLog)
	}
}

// handleTurnNotification handles a turn notification from the server
func handleTurnNotification(client *network.GameClient, payload interface{}) {
	// Parse turn notification payload
	turnMap, ok := payload.(map[string]interface{})
	if !ok {
		fmt.Println("Error parsing turn notification")
		return
	}

	// Extract current turn username
	username, _ := turnMap["currentTurnUsername"].(string)
	fmt.Printf("\n=== It's %s's turn ===\n", username)

	// If it's my turn, display available troops and valid targets
	if username == client.Username {
		displayPlayerHandAndTargetInfo()
	}
}

// displayPlayerHandAndTargetInfo displays the player's current hand and auto-target
func displayPlayerHandAndTargetInfo() {
	fmt.Println("\n=== Your Available Troops ===")
	if len(myTroops) == 0 {
		fmt.Println("No troops available at the moment.")
	} else {
		for _, troop := range myTroops {
			fmt.Printf("%s: HP %d, ATK %d, DEF %d\n",
				troop.Name, troop.HP, troop.Attack, troop.Defense)
		}
	}

	fmt.Println("\n=== Current Target ===")

	// Display the current auto-target based on enemy tower status
	if !enemyGuardTower1.Destroyed {
		fmt.Printf("Auto-targeting: %s (Guard Tower 1): HP %d/%d, DEF %d\n",
			enemyGuardTower1.ID, enemyGuardTower1.CurrentHP, enemyGuardTower1.MaxHP, enemyGuardTower1.Defense)
	} else if !enemyGuardTower2.Destroyed {
		fmt.Printf("Auto-targeting: %s (Guard Tower 2): HP %d/%d, DEF %d\n",
			enemyGuardTower2.ID, enemyGuardTower2.CurrentHP, enemyGuardTower2.MaxHP, enemyGuardTower2.Defense)
	} else if !enemyKingTower.Destroyed {
		fmt.Printf("Auto-targeting: %s (King Tower): HP %d/%d, DEF %d\n",
			enemyKingTower.ID, enemyKingTower.CurrentHP, enemyKingTower.MaxHP, enemyKingTower.Defense)
	} else {
		fmt.Println("All enemy towers are destroyed!")
	}

	// Special note for Queen
	hasQueen := false
	for _, troop := range myTroops {
		if troop.Name == "Queen" {
			hasQueen = true
			break
		}
	}

	if hasQueen {
		fmt.Println("\nNOTE: The Queen will heal your lowest HP tower when deployed")
		// Display player's own tower HPs for context if Queen is available
		var kingStatus, guard1Status, guard2Status string
		if myKingTower.Destroyed {
			kingStatus = " (DESTROYED)"
		}
		if myGuardTower1.Destroyed {
			guard1Status = " (DESTROYED)"
		}
		if myGuardTower2.Destroyed {
			guard2Status = " (DESTROYED)"
		}

		fmt.Printf("Your tower HPs: King=%d/%d%s, Guard1=%d/%d%s, Guard2=%d/%d%s\n",
			myKingTower.CurrentHP, myKingTower.MaxHP, kingStatus,
			myGuardTower1.CurrentHP, myGuardTower1.MaxHP, guard1Status,
			myGuardTower2.CurrentHP, myGuardTower2.MaxHP, guard2Status)
	}
}

// handleActionResult handles an action result from the server
func handleActionResult(payload interface{}) {
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
	} else {
		fmt.Printf("Action failed: %s\n", message)
		fmt.Println("Please try again.")
	}
}

// handleGameOverNotification handles a game over notification from the server
func handleGameOverNotification(payload interface{}) {
	// Parse game over notification payload
	overMap, ok := payload.(map[string]interface{})
	if !ok {
		fmt.Println("Error parsing game over notification")
		return
	}

	// Extract winner and reason
	winner, _ := overMap["winnerUsername"].(string)
	reason, _ := overMap["reason"].(string)

	fmt.Printf("\n=== GAME OVER ===\n")
	fmt.Printf("Winner: %s\n", winner)
	fmt.Printf("Reason: %s\n", reason)
	fmt.Println("Starting a new game when another player connects...")
}

// displayGameStatus displays the current game status
func displayGameStatus(client *network.GameClient) {
	fmt.Printf("\n=== Game Status ===\n")
	fmt.Printf("Game Mode: %s\n", gameMode)
	fmt.Printf("You: %s\n", client.Username)
	fmt.Printf("Opponent: %s\n", client.OpponentName)
	fmt.Printf("Current Turn: %s\n", currentTurn)
	if lastActionLog != "" {
		fmt.Printf("Last Action: %s\n", lastActionLog)
	}

	fmt.Println("\n=== Your Towers ===")
	var kingStatus, guard1Status, guard2Status string
	if myKingTower.Destroyed {
		kingStatus = " (DESTROYED)"
	} else {
		kingStatus = ""
	}
	if myGuardTower1.Destroyed {
		guard1Status = " (DESTROYED)"
	} else {
		guard1Status = ""
	}
	if myGuardTower2.Destroyed {
		guard2Status = " (DESTROYED)"
	} else {
		guard2Status = ""
	}

	fmt.Printf("King Tower: HP %d/%d, ATK %d, DEF %d%s\n",
		myKingTower.CurrentHP, myKingTower.MaxHP, myKingTower.Attack, myKingTower.Defense, kingStatus)
	fmt.Printf("Guard Tower 1: HP %d/%d, ATK %d, DEF %d%s\n",
		myGuardTower1.CurrentHP, myGuardTower1.MaxHP, myGuardTower1.Attack, myGuardTower1.Defense, guard1Status)
	fmt.Printf("Guard Tower 2: HP %d/%d, ATK %d, DEF %d%s\n",
		myGuardTower2.CurrentHP, myGuardTower2.MaxHP, myGuardTower2.Attack, myGuardTower2.Defense, guard2Status)

	fmt.Println("\n=== Opponent Towers ===")
	var enemyKingStatus, enemyGuard1Status, enemyGuard2Status string
	if enemyKingTower.Destroyed {
		enemyKingStatus = " (DESTROYED)"
	} else {
		enemyKingStatus = ""
	}
	if enemyGuardTower1.Destroyed {
		enemyGuard1Status = " (DESTROYED)"
	} else {
		enemyGuard1Status = ""
	}
	if enemyGuardTower2.Destroyed {
		enemyGuard2Status = " (DESTROYED)"
	} else {
		enemyGuard2Status = ""
	}

	fmt.Printf("King Tower: HP %d/%d, DEF %d%s\n",
		enemyKingTower.CurrentHP, enemyKingTower.MaxHP, enemyKingTower.Defense, enemyKingStatus)
	fmt.Printf("Guard Tower 1: HP %d/%d, DEF %d%s\n",
		enemyGuardTower1.CurrentHP, enemyGuardTower1.MaxHP, enemyGuardTower1.Defense, enemyGuard1Status)
	fmt.Printf("Guard Tower 2: HP %d/%d, DEF %d%s\n",
		enemyGuardTower2.CurrentHP, enemyGuardTower2.MaxHP, enemyGuardTower2.Defense, enemyGuard2Status)

	fmt.Println("\n=== Your Available Troops ===")
	if len(myTroops) == 0 {
		fmt.Println("No troops available")
	} else {
		for _, troop := range myTroops {
			fmt.Printf("%s: HP %d, ATK %d, DEF %d\n",
				troop.Name, troop.HP, troop.Attack, troop.Defense)
		}
	}
}

// displayHelp displays help information
func displayHelp() {
	fmt.Println("\n==============================================")
	fmt.Println("📋 GAME COMMANDS 📋")
	fmt.Println("==============================================")
	fmt.Println("Game Actions:")
	fmt.Println("  d <troop_name> - Deploy a troop to attack")
	fmt.Println("    - Auto-targets enemy towers in sequence")
	fmt.Println("    - Example: d Knight (deploys Knight to attack)")
	fmt.Println("    - Example: d Queen (deploys Queen to heal your lowest HP tower)")
	fmt.Println("")
	fmt.Println("Information Commands:")
	fmt.Println("  status - Display detailed game status (towers, troops, etc.)")
	fmt.Println("  help   - Display this help information")
	fmt.Println("")
	fmt.Println("System Commands:")
	fmt.Println("  quit   - Exit the game")
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
