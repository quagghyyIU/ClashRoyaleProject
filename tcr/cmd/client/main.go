package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
	"tcr/internal/models"
	"tcr/internal/network"
)

func main() {
	// Define command line flags
	addr := flag.String("addr", "localhost:8080", "Server address to connect to (host:port)")
	flag.Parse()

	fmt.Println("TCR Client - Phase 2")
	fmt.Println("=====================")

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
	go handleServerMessages(client)

	// Prompt for username
	fmt.Print("Enter your username: ")
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)

	// Send login request
	err = client.Login(username)
	if err != nil {
		fmt.Printf("Error sending login request: %v\n", err)
		client.Disconnect()
		return
	}

	// Main input loop
	for client.Connected {
		// Read user input
		fmt.Print("> ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		// Process commands
		if input == "quit" || input == "exit" {
			break
		}

		// Currently only login is implemented, other commands will be added in Phase 3
		fmt.Println("Command not recognized. Available commands: quit")
	}

	// Disconnect from server
	client.Disconnect()
	fmt.Println("Disconnected from server. Goodbye!")
}

// handleServerMessages handles messages received from the server
func handleServerMessages(client *network.GameClient) {
	// Handle disconnect notification
	go func() {
		err := <-client.DisconnectCh
		fmt.Printf("\nDisconnected from server: %v\n", err)
		fmt.Print("> ")
	}()

	// Handle messages
	for {
		select {
		case message := <-client.MessageCh:
			handleMessage(message)
		}
	}
}

// handleMessage processes a message received from the server
func handleMessage(message models.GenericMessage) {
	switch message.Type {
	case models.MsgTypeLoginResponse:
		handleLoginResponse(message.Payload)
	case models.MsgTypeErrorNotification:
		handleErrorNotification(message.Payload)
	default:
		fmt.Printf("\nReceived message of type: %s\n", message.Type)
		fmt.Print("> ")
	}
}

// handleLoginResponse processes a login response from the server
func handleLoginResponse(payload interface{}) {
	// Parse login response payload
	responseMap, ok := payload.(map[string]interface{})
	if !ok {
		fmt.Println("\nInvalid login response payload")
		return
	}

	// Extract success status and message
	success, _ := responseMap["success"].(bool)
	message, _ := responseMap["message"].(string)

	// Display result
	if success {
		fmt.Printf("\nLogin successful: %s\n", message)
	} else {
		fmt.Printf("\nLogin failed: %s\n", message)
	}

	fmt.Print("> ")
}

// handleErrorNotification processes an error notification from the server
func handleErrorNotification(payload interface{}) {
	// Parse error notification payload
	errorMap, ok := payload.(map[string]interface{})
	if !ok {
		fmt.Println("\nInvalid error notification payload")
		return
	}

	// Extract error message
	errorMessage, _ := errorMap["errorMessage"].(string)

	// Display error
	fmt.Printf("\nServer error: %s\n", errorMessage)
	fmt.Print("> ")
}
