package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"tcr/internal/game"
	"tcr/internal/network"
	"tcr/internal/storage"
)

func main() {
	// Define command line flags
	addr := flag.String("addr", ":8080", "Server address to listen on (host:port)")
	mode := flag.String("mode", "online", "Server mode (online or offline)")
	configsDir := flag.String("configs", "configs", "Path to config files directory")
	dataDir := flag.String("data", "data", "Path to data files directory")
	flag.Parse()

	fmt.Println("TCR Server - Starting...")
	fmt.Printf("Address: %s\n", *addr)
	fmt.Printf("Mode: %s\n", *mode)
	fmt.Printf("Configs directory: %s\n", *configsDir)
	fmt.Printf("Data directory: %s\n", *dataDir)

	// Create data directories if they don't exist
	if err := os.MkdirAll(*dataDir, 0755); err != nil {
		log.Fatalf("Failed to create data directory: %v", err)
	}

	usersDir := filepath.Join(*dataDir, "users")
	if err := os.MkdirAll(usersDir, 0755); err != nil {
		log.Fatalf("Failed to create users directory: %v", err)
	}

	playersDir := filepath.Join(*dataDir, "players")
	if err := os.MkdirAll(playersDir, 0755); err != nil {
		log.Fatalf("Failed to create players directory: %v", err)
	}

	// Create JSON handler
	jsonHandler := storage.NewJSONHandler(*configsDir, *dataDir)

	// Run in offline mode if specified
	if *mode == "offline" {
		fmt.Println("Running in offline mode...")
		testSimpleTCR()
		return
	}

	// Create and start server
	server := network.NewServer(*addr, jsonHandler)
	err := server.Start()
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}

	// Handle Ctrl+C to gracefully shutdown server
	// In a real implementation, you'd add signal handling here
	fmt.Println("Server running. Press Ctrl+C to stop.")
	select {}
}

// testSimpleTCR tests the Simple TCR game logic in a console environment
func testSimpleTCR() {
	// Initialize storage handler for loading configs
	jsonHandler := storage.NewJSONHandler("configs", "data/players")

	// Load troop and tower specs
	troopSpecs, err := jsonHandler.LoadTroopSpecs()
	if err != nil {
		fmt.Printf("Error loading troop specs: %v\n", err)
		return
	}

	towerSpecs, err := jsonHandler.LoadTowerSpecs()
	if err != nil {
		fmt.Printf("Error loading tower specs: %v\n", err)
		return
	}

	// Create a new game session
	gameSession := game.NewGameSession("PlayerA", "PlayerB", troopSpecs, towerSpecs, jsonHandler)

	// Print initial game state
	fmt.Println("\n=== Initial Game State ===")
	fmt.Println(gameSession.GetGameStateInfo())

	// Start game loop
	reader := bufio.NewReader(os.Stdin)
	for !gameSession.GameState.IsGameOver {
		fmt.Printf("\n%s's turn. Enter command (e.g., 'd Pawn PlayerB_GUARD1'): ", gameSession.GameState.CurrentTurn)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		// Parse command
		parts := strings.Split(input, " ")
		if len(parts) < 1 {
			fmt.Println("Invalid command format.")
			continue
		}

		switch parts[0] {
		case "d", "deploy":
			if len(parts) != 3 {
				fmt.Println("Usage: d <troop_name> <target_tower_id>")
				continue
			}

			troopName := parts[1]
			targetTowerID := parts[2]

			// Deploy the troop
			resultMsg, success := gameSession.DeployTroop(gameSession.GameState.CurrentTurn, troopName, targetTowerID)
			fmt.Println(resultMsg)

			if success {
				// Print game state
				fmt.Println("\n=== Game State ===")
				fmt.Println(gameSession.GetGameStateInfo())
			}
		case "q", "quit":
			return
		case "h", "help":
			fmt.Println("Commands:")
			fmt.Println("  d <troop_name> <target_tower_id> - Deploy troop to attack a target tower")
			fmt.Println("  q - Quit the game")
			fmt.Println("  h - Show this help")
		default:
			fmt.Println("Unknown command. Type 'h' for help.")
		}
	}

	// Game over
	fmt.Println("\n=== Game Over ===")
	fmt.Printf("Winner: %s\n", gameSession.GameState.Winner)
}
