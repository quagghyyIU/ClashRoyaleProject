package storage

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"tcr/internal/models"
)

// JSONHandler handles JSON file operations
type JSONHandler struct {
	ConfigDir string
	DataDir   string
	mutex     sync.Mutex
}

// UserData represents user account data
type UserData struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// NewJSONHandler creates a new JSON handler
func NewJSONHandler(configDir, dataDir string) *JSONHandler {
	return &JSONHandler{
		ConfigDir: configDir,
		DataDir:   dataDir,
	}
}

// LoadTroopSpecs loads troop specifications from a JSON file
func (h *JSONHandler) LoadTroopSpecs() ([]models.TroopSpec, error) {
	filePath := filepath.Join(h.ConfigDir, "troops.json")
	file, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read troop specs file: %w", err)
	}

	var troops []models.TroopSpec
	err = json.Unmarshal(file, &troops)
	if err != nil {
		return nil, fmt.Errorf("failed to parse troop specs JSON: %w", err)
	}

	return troops, nil
}

// LoadTowerSpecs loads tower specifications from a JSON file
func (h *JSONHandler) LoadTowerSpecs() ([]models.TowerSpec, error) {
	filePath := filepath.Join(h.ConfigDir, "towers.json")
	file, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read tower specs file: %w", err)
	}

	var towers []models.TowerSpec
	err = json.Unmarshal(file, &towers)
	if err != nil {
		return nil, fmt.Errorf("failed to parse tower specs JSON: %w", err)
	}

	return towers, nil
}

// SaveUserData saves user login data to a JSON file
func (h *JSONHandler) SaveUserData(user UserData) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	// Ensure the users directory exists
	usersDir := filepath.Join(h.DataDir, "users")
	if err := os.MkdirAll(usersDir, 0755); err != nil {
		return fmt.Errorf("failed to create users directory: %w", err)
	}

	// Create the file path
	filePath := filepath.Join(usersDir, user.Username+".json")

	// Marshal user data to JSON
	data, err := json.MarshalIndent(user, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal user data: %w", err)
	}

	// Write the file
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write user data file: %w", err)
	}

	return nil
}

// LoadUserData loads user login data from a JSON file
func (h *JSONHandler) LoadUserData(username string) (UserData, error) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	// Get the file path
	usersDir := filepath.Join(h.DataDir, "users")
	filePath := filepath.Join(usersDir, username+".json")

	// Read the file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return UserData{}, fmt.Errorf("failed to read user data file: %w", err)
	}

	// Unmarshal JSON to user data
	var user UserData
	if err := json.Unmarshal(data, &user); err != nil {
		return UserData{}, fmt.Errorf("failed to parse user data JSON: %w", err)
	}

	return user, nil
}

// UserExists checks if a user already exists
func (h *JSONHandler) UserExists(username string) bool {
	usersDir := filepath.Join(h.DataDir, "users")
	filePath := filepath.Join(usersDir, username+".json")
	_, err := os.Stat(filePath)
	return err == nil
}

// SavePlayerData saves player data to a JSON file (for Enhanced TCR)
func (h *JSONHandler) SavePlayerData(username string, currentEXP, level int) error {
	playerData := struct {
		Username   string `json:"username"`
		CurrentEXP int    `json:"currentEXP"`
		Level      int    `json:"level"`
	}{
		Username:   username,
		CurrentEXP: currentEXP,
		Level:      level,
	}

	data, err := json.MarshalIndent(playerData, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling player data: %w", err)
	}

	filename := fmt.Sprintf("player_%s.json", username)
	filePath := filepath.Join(h.DataDir, filename)

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("error writing player data file: %w", err)
	}

	return nil
}

// LoadPlayerData loads player data from a JSON file (for Enhanced TCR)
func (h *JSONHandler) LoadPlayerData(username string) (int, int, error) {
	filename := fmt.Sprintf("player_%s.json", username)
	filePath := filepath.Join(h.DataDir, filename)

	log.Printf("[LOADPLAYERDATA_DEBUG] Attempting to load player data for: Username='%s', FullPath='%s'", username, filePath)

	data, err := os.ReadFile(filePath)
	if err != nil {
		log.Printf("[LOADPLAYERDATA_DEBUG] os.ReadFile error for '%s': %v. os.IsNotExist(err): %t", filePath, err, os.IsNotExist(err))
		if os.IsNotExist(err) {
			// If file doesn't exist, return default values
			return 0, 1, nil
		}
		return 0, 1, fmt.Errorf("error reading player data file: %w", err)
	}

	log.Printf("[LOADPLAYERDATA_DEBUG] Successfully read data for '%s', length: %d", filePath, len(data))

	var playerData struct {
		Username   string `json:"username"`
		CurrentEXP int    `json:"currentEXP"`
		Level      int    `json:"level"`
	}

	if err := json.Unmarshal(data, &playerData); err != nil {
		log.Printf("[LOADPLAYERDATA_DEBUG] JSON unmarshal error for '%s': %v", filePath, err)
		return 0, 1, fmt.Errorf("error unmarshaling player data: %w", err)
	}

	log.Printf("[LOADPLAYERDATA_DEBUG] Successfully unmarshaled data for '%s': EXP=%d, Level=%d", filePath, playerData.CurrentEXP, playerData.Level)
	return playerData.CurrentEXP, playerData.Level, nil
}
