package main

import (
	"fmt"
	"log"
	"path/filepath"
	"tcr/internal/storage"
)

func main() {
	fmt.Println("Text-Based Clash Royale (TCR) - Server")
	fmt.Println("=======================================")
	fmt.Println("Server initialized - Phase 0")

	// Initialize paths
	configsPath := filepath.Join(".", "configs")
	playersPath := filepath.Join(".", "data", "players")

	// Initialize JSON handler for loading game data
	jsonHandler := storage.NewJSONHandler(configsPath, playersPath)

	// Load troop and tower specifications
	troops, err := jsonHandler.LoadTroopSpecs()
	if err != nil {
		log.Fatalf("Failed to load troop specifications: %v", err)
	}

	towers, err := jsonHandler.LoadTowerSpecs()
	if err != nil {
		log.Fatalf("Failed to load tower specifications: %v", err)
	}

	// Log successful loading of specifications
	log.Printf("Successfully loaded %d troop specifications", len(troops))
	log.Printf("Successfully loaded %d tower specifications", len(towers))

	// Print out loaded troops and towers for verification
	fmt.Println("\nLoaded Troops:")
	for _, troop := range troops {
		fmt.Printf("- %s (HP: %d, ATK: %d, DEF: %d, Mana: %d, EXP: %d)\n",
			troop.Name, troop.BaseHP, troop.BaseATK, troop.BaseDEF, troop.ManaCost, troop.DestroyEXP)
		if troop.SpecialAbility != "" {
			fmt.Printf("  Special Ability: %s\n", troop.SpecialAbility)
		}
	}

	fmt.Println("\nLoaded Towers:")
	for _, tower := range towers {
		fmt.Printf("- %s (Type: %s, HP: %d, ATK: %d, DEF: %d, CRIT: %d%%, EXP: %d)\n",
			tower.Name, tower.Type, tower.BaseHP, tower.BaseATK, tower.BaseDEF,
			tower.CritChancePercent, tower.DestroyEXP)
	}

	// This is a placeholder for the server implementation in future phases
	log.Println("Server application initialized in Phase 0")
}
