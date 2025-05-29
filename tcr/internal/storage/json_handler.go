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

// PlayerProfile represents the data for a player that is persisted.
type PlayerProfile struct {
	Username                string `json:"username"`
	Level                   int    `json:"level"`
	CurrentEXP              int    `json:"currentEXP"`
	RequiredEXPForNextLevel int    `json:"requiredEXPForNextLevel"`
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

// SavePlayerData saves player profile data to a JSON file.
// The data is stored in <DataDir>/players/<username>.json.
func (h *JSONHandler) SavePlayerData(profile PlayerProfile) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	playersDataDir := filepath.Join(h.DataDir, "players")
	if err := os.MkdirAll(playersDataDir, 0755); err != nil {
		return fmt.Errorf("failed to create player data directory '%s': %w", playersDataDir, err)
	}

	filePath := filepath.Join(playersDataDir, profile.Username+".json")

	data, err := json.MarshalIndent(profile, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling player profile for %s: %w", profile.Username, err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("error writing player profile file for %s: %w", profile.Username, err)
	}
	log.Printf("Player data for %s saved to %s", profile.Username, filePath)
	return nil
}

// LoadPlayerData loads a player's profile from a JSON file.
// It looks for <DataDir>/players/<username>.json.
// If the file doesn't exist, it returns a default profile for a new level 1 player.
func (h *JSONHandler) LoadPlayerData(username string) (PlayerProfile, error) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	playersDataDir := filepath.Join(h.DataDir, "players")
	filePath := filepath.Join(playersDataDir, username+".json")

	log.Printf("[LOADPLAYERDATA_DEBUG] Attempting to load player data for: Username='%s', FullPath='%s'", username, filePath)

	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("[LOADPLAYERDATA_DEBUG] Player data file not found for '%s'. Returning default profile.", username)
			// Return default profile for a new player
			return PlayerProfile{
				Username:                username,
				Level:                   1,
				CurrentEXP:              0,
				RequiredEXPForNextLevel: 100, // Base EXP for level 1 to level up, as per plan
			}, nil
		}
		return PlayerProfile{}, fmt.Errorf("error reading player data file '%s': %w", filePath, err)
	}

	log.Printf("[LOADPLAYERDATA_DEBUG] Successfully read data for '%s', length: %d", filePath, len(data))

	var profile PlayerProfile
	if err := json.Unmarshal(data, &profile); err != nil {
		log.Printf("[LOADPLAYERDATA_DEBUG] JSON unmarshal error for '%s': %v", filePath, err)
		return PlayerProfile{}, fmt.Errorf("error unmarshaling player data from '%s': %w", filePath, err)
	}

	// Ensure username in profile matches requested username, or fill if empty (older format handling if any)
	if profile.Username == "" {
		profile.Username = username
	}

	log.Printf("[LOADPLAYERDATA_DEBUG] Successfully unmarshaled data for '%s': EXP=%d, Level=%d, ReqEXP=%d",
		filePath, profile.CurrentEXP, profile.Level, profile.RequiredEXPForNextLevel)
	return profile, nil
}
