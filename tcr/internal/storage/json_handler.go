package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"tcr/internal/models"
)

// JSONHandler handles loading and saving JSON data
type JSONHandler struct {
	ConfigsPath string
	PlayersPath string
}

// NewJSONHandler creates a new JSONHandler with the given paths
func NewJSONHandler(configsPath, playersPath string) *JSONHandler {
	return &JSONHandler{
		ConfigsPath: configsPath,
		PlayersPath: playersPath,
	}
}

// LoadTroopSpecs loads troop specifications from troops.json
func (h *JSONHandler) LoadTroopSpecs() ([]models.TroopSpec, error) {
	troopsFilePath := filepath.Join(h.ConfigsPath, "troops.json")
	data, err := os.ReadFile(troopsFilePath)
	if err != nil {
		return nil, fmt.Errorf("error reading troops.json: %w", err)
	}

	var troops []models.TroopSpec
	if err := json.Unmarshal(data, &troops); err != nil {
		return nil, fmt.Errorf("error unmarshaling troops data: %w", err)
	}

	return troops, nil
}

// LoadTowerSpecs loads tower specifications from towers.json
func (h *JSONHandler) LoadTowerSpecs() ([]models.TowerSpec, error) {
	towersFilePath := filepath.Join(h.ConfigsPath, "towers.json")
	data, err := os.ReadFile(towersFilePath)
	if err != nil {
		return nil, fmt.Errorf("error reading towers.json: %w", err)
	}

	var towers []models.TowerSpec
	if err := json.Unmarshal(data, &towers); err != nil {
		return nil, fmt.Errorf("error unmarshaling towers data: %w", err)
	}

	return towers, nil
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
	filePath := filepath.Join(h.PlayersPath, filename)

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("error writing player data file: %w", err)
	}

	return nil
}

// LoadPlayerData loads player data from a JSON file (for Enhanced TCR)
func (h *JSONHandler) LoadPlayerData(username string) (int, int, error) {
	filename := fmt.Sprintf("player_%s.json", username)
	filePath := filepath.Join(h.PlayersPath, filename)

	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// If file doesn't exist, return default values
			return 0, 1, nil
		}
		return 0, 1, fmt.Errorf("error reading player data file: %w", err)
	}

	var playerData struct {
		Username   string `json:"username"`
		CurrentEXP int    `json:"currentEXP"`
		Level      int    `json:"level"`
	}

	if err := json.Unmarshal(data, &playerData); err != nil {
		return 0, 1, fmt.Errorf("error unmarshaling player data: %w", err)
	}

	return playerData.CurrentEXP, playerData.Level, nil
}
