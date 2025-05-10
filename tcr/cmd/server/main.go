package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"tcr/internal/game"
	"tcr/internal/network"
	"tcr/internal/storage"
)

func main() {
	// Define command line flags
	mode := flag.String("mode", "network", "Run mode: 'network' for network server or 'offline' for offline test")
	addr := flag.String("addr", ":8080", "Network address to listen on (host:port)")
	flag.Parse()

	fmt.Println("Starting TCR Server...")

	// Run in the specified mode
	if *mode == "offline" {
		fmt.Println("Running Simple TCR Offline Test")
		testSimpleTCR()
	} else {
		fmt.Printf("Starting network server on %s\n", *addr)
		runNetworkServer(*addr)
	}
}

// runNetworkServer starts the TCP server and listens for connections
func runNetworkServer(addr string) {
	// Create the server
	server := network.NewServer(addr)

	// Set up signal handler for graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Start the server in a goroutine
	go func() {
		err := server.Start()
		if err != nil {
			fmt.Printf("Server error: %v\n", err)
			close(sigCh) // Signal to exit
		}
	}()

	// Wait for termination signal
	<-sigCh
	fmt.Println("\nShutting down server...")
	server.Stop()
	fmt.Println("Server shutdown complete")
}

// mapTowerID maps a simple numeric ID to the actual tower ID
func mapTowerID(opponentUsername string, towerNumID string) string {
	switch towerNumID {
	case "1", "g1", "guard1":
		return opponentUsername + "_GUARD1"
	case "2", "g2", "guard2":
		return opponentUsername + "_GUARD2"
	case "3", "k", "king":
		return opponentUsername + "_KING"
	default:
		return "" // Invalid ID
	}
}

// getOpponentUsername returns the opponent's username
func getOpponentUsername(currentPlayer string, playerA, playerB string) string {
	if currentPlayer == playerA {
		return playerB
	}
	return playerA
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
	gameSession := game.NewGameSession("PlayerA", "PlayerB", troopSpecs, towerSpecs)

	// Print initial game state
	fmt.Println("\n=== Initial Game State ===")
	fmt.Println(gameSession.GetGameStateInfo())

	// Start game loop
	reader := bufio.NewReader(os.Stdin)

	for !gameSession.GameState.IsGameOver {
		currentPlayer := gameSession.GameState.CurrentTurn
		opponentPlayer := getOpponentUsername(currentPlayer, gameSession.GameState.PlayerA.Username, gameSession.GameState.PlayerB.Username)

		fmt.Println("\n=== Commands ===")
		fmt.Println("  d <troop_name> <tower_number> - Deploy a troop to attack a tower")
		fmt.Println("  Tower numbers: 1=Guard1, 2=Guard2, 3=King")
		fmt.Println("  quit - Exit the game")
		fmt.Printf("\n%s's turn> ", currentPlayer)

		// Read user input
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "quit" {
			break
		}

		// Parse command
		parts := strings.Split(input, " ")
		if len(parts) == 0 {
			continue
		}

		// Check if command is deploy or d
		if parts[0] == "deploy" || parts[0] == "d" {
			if len(parts) != 3 {
				fmt.Println("Invalid deploy command. Usage: d <troop_name> <tower_number>")
				continue
			}

			troopName := parts[1]
			towerID := mapTowerID(opponentPlayer, parts[2])

			if towerID == "" {
				fmt.Println("Invalid tower number. Use 1 for Guard1, 2 for Guard2, 3 for King.")
				continue
			}

			message, success := gameSession.DeployTroop(currentPlayer, troopName, towerID)
			fmt.Println(message)

			if success {
				// Print updated game state
				fmt.Println("\n=== Updated Game State ===")
				fmt.Println(gameSession.GetGameStateInfo())
			}
		} else {
			fmt.Println("Unknown command. Available commands: d (deploy), quit")
		}
	}

	// Game over
	if gameSession.GameState.IsGameOver {
		fmt.Println("\n=== GAME OVER ===")
		fmt.Printf("Winner: %s\n", gameSession.GameState.Winner)
	}
}
