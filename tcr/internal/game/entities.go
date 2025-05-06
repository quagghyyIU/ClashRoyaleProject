package game

import (
	"tcr/internal/models"
)

// Player represents a player in the game
type Player struct {
	Username    string
	KingTower   *TowerInstance
	GuardTower1 *TowerInstance
	GuardTower2 *TowerInstance
	Troops      []*TroopInstance // Available troops (hand)

	// Enhanced TCR features
	CurrentEXP              int
	Level                   int
	CurrentMana             int
	RequiredEXPForNextLevel int
}

// TowerInstance represents a tower instance in the game with current stats
type TowerInstance struct {
	Spec       *models.TowerSpec
	ID         string // e.g., "PlayerA_KingTower"
	CurrentHP  int
	CurrentATK int
	CurrentDEF int
	Destroyed  bool
}

// TroopInstance represents a troop instance in the game with current stats
type TroopInstance struct {
	Spec       *models.TroopSpec
	ID         string // Unique identifier
	CurrentHP  int
	CurrentATK int
	CurrentDEF int
}

// NewPlayer creates a new player with initialized values
func NewPlayer(username string) *Player {
	return &Player{
		Username:                username,
		Troops:                  make([]*TroopInstance, 0),
		Level:                   1,
		CurrentEXP:              0,
		CurrentMana:             0,
		RequiredEXPForNextLevel: 100, // Base EXP required for level 2
	}
}

// NewTowerInstance creates a new tower instance from a tower spec
func NewTowerInstance(spec *models.TowerSpec, playerUsername string, playerLevel int) *TowerInstance {
	levelMultiplier := 1.0 + float64(playerLevel-1)*0.1

	return &TowerInstance{
		Spec:       spec,
		ID:         playerUsername + "_" + spec.Type,
		CurrentHP:  int(float64(spec.BaseHP) * levelMultiplier),
		CurrentATK: int(float64(spec.BaseATK) * levelMultiplier),
		CurrentDEF: int(float64(spec.BaseDEF) * levelMultiplier),
		Destroyed:  false,
	}
}

// NewTroopInstance creates a new troop instance from a troop spec
func NewTroopInstance(spec *models.TroopSpec, id string, playerLevel int) *TroopInstance {
	levelMultiplier := 1.0 + float64(playerLevel-1)*0.1

	return &TroopInstance{
		Spec:       spec,
		ID:         id,
		CurrentHP:  int(float64(spec.BaseHP) * levelMultiplier),
		CurrentATK: int(float64(spec.BaseATK) * levelMultiplier),
		CurrentDEF: int(float64(spec.BaseDEF) * levelMultiplier),
	}
}
