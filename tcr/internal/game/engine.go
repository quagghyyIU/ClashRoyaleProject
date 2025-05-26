package game

import (
	"fmt"
	"math/rand"
	"tcr/internal/models"
	"time"
)

// GameSession represents a game session between two players
type GameSession struct {
	GameState  *GameState
	TroopSpecs []models.TroopSpec // Available troops for both players
	TowerSpecs []models.TowerSpec // Available towers for both players
}

// NewGameSession creates a new game session with two players
func NewGameSession(playerAName, playerBName string, troopSpecs []models.TroopSpec, towerSpecs []models.TowerSpec) *GameSession {
	// Initialize random seed
	rand.NewSource(time.Now().UnixNano())

	// Create two players
	playerA := NewPlayer(playerAName)
	playerB := NewPlayer(playerBName)

	// Initialize the game session
	gs := &GameSession{
		TroopSpecs: troopSpecs,
		TowerSpecs: towerSpecs,
	}

	// Assign towers to players
	gs.assignTowersToPlayers(playerA, playerB)

	// Assign random troops to players
	gs.assignTroopsToPlayers(playerA, playerB)

	// Create game state
	gs.GameState = NewGameState(playerA, playerB)

	return gs
}

// assignTowersToPlayers assigns towers to both players
func (gs *GameSession) assignTowersToPlayers(playerA, playerB *Player) {
	// Find tower specs by type
	var kingTowerSpec, guardTower1Spec, guardTower2Spec *models.TowerSpec

	for i := range gs.TowerSpecs {
		switch gs.TowerSpecs[i].Type {
		case "KING":
			kingTowerSpec = &gs.TowerSpecs[i]
		case "GUARD1":
			guardTower1Spec = &gs.TowerSpecs[i]
		case "GUARD2":
			guardTower2Spec = &gs.TowerSpecs[i]
		}
	}

	// Assign towers to player A
	playerA.KingTower = NewTowerInstance(kingTowerSpec, playerA.Username, playerA.Level)
	playerA.GuardTower1 = NewTowerInstance(guardTower1Spec, playerA.Username, playerA.Level)
	playerA.GuardTower2 = NewTowerInstance(guardTower2Spec, playerA.Username, playerA.Level)

	// Assign towers to player B
	playerB.KingTower = NewTowerInstance(kingTowerSpec, playerB.Username, playerB.Level)
	playerB.GuardTower1 = NewTowerInstance(guardTower1Spec, playerB.Username, playerB.Level)
	playerB.GuardTower2 = NewTowerInstance(guardTower2Spec, playerB.Username, playerB.Level)
}

// assignTroopsToPlayers assigns random troops to both players
func (gs *GameSession) assignTroopsToPlayers(playerA, playerB *Player) {
	// Filter out special-only troops (like Queen)
	regularTroops := make([]models.TroopSpec, 0)
	specialTroops := make([]models.TroopSpec, 0)

	for _, troop := range gs.TroopSpecs {
		if troop.IsSpecialOnly {
			specialTroops = append(specialTroops, troop)
		} else {
			regularTroops = append(regularTroops, troop)
		}
	}

	// Both players should always have access to special troops
	// For regular troops, randomly select 3 unique troops for each player
	regularTroopIndices := rand.Perm(len(regularTroops))

	// Ensure we have enough regular troops
	if len(regularTroopIndices) < 6 {
		// If we don't have 6 unique regular troops, just use what we have and repeat some
		for i := len(regularTroopIndices); i < 6; i++ {
			regularTroopIndices = append(regularTroopIndices, i%len(regularTroops))
		}
	}

	// Assign 3 regular troops to player A
	for i := 0; i < 3; i++ {
		troopSpec := regularTroops[regularTroopIndices[i]]
		troopInstance := NewTroopInstance(&troopSpec, fmt.Sprintf("%s_troop_%d", playerA.Username, i), playerA.Level)
		playerA.Troops = append(playerA.Troops, troopInstance)
	}

	// Assign 3 regular troops to player B
	for i := 0; i < 3; i++ {
		troopSpec := regularTroops[regularTroopIndices[i+3]]
		troopInstance := NewTroopInstance(&troopSpec, fmt.Sprintf("%s_troop_%d", playerB.Username, i), playerB.Level)
		playerB.Troops = append(playerB.Troops, troopInstance)
	}

	// Add special troops (like Queen) to both players
	for i, troopSpec := range specialTroops {
		// Add to player A
		troopInstanceA := NewTroopInstance(&troopSpec, fmt.Sprintf("%s_special_%d", playerA.Username, i), playerA.Level)
		playerA.Troops = append(playerA.Troops, troopInstanceA)

		// Add to player B
		troopInstanceB := NewTroopInstance(&troopSpec, fmt.Sprintf("%s_special_%d", playerB.Username, i), playerB.Level)
		playerB.Troops = append(playerB.Troops, troopInstanceB)
	}
}

// DeployTroop handles the deployment of a troop by a player
// Returns a message describing what happened and whether the action was successful
func (gs *GameSession) DeployTroop(playerUsername, troopName, targetTowerID string) (string, bool) {
	// Check if game is already over
	if gs.GameState.IsGameOver {
		return "Game is already over.", false
	}

	// Get the player who is deploying the troop
	var actingPlayer *Player
	if playerUsername == gs.GameState.PlayerA.Username {
		actingPlayer = gs.GameState.PlayerA
	} else if playerUsername == gs.GameState.PlayerB.Username {
		actingPlayer = gs.GameState.PlayerB
	} else {
		return "Invalid player username.", false
	}

	// Check if it's the player's turn
	if gs.GameState.CurrentTurn != playerUsername && !gs.GameState.CanContinueAttacking {
		return "It's not your turn.", false
	}

	// Find the troop in the player's hand
	troop, troopIndex := FindTroopInHand(actingPlayer, troopName)
	if troop == nil {
		return "Troop not found in your hand.", false
	}

	// Handle Queen's special ability (or other special-only troops)
	if troop.Spec.IsSpecialOnly {
		// Apply special ability
		abilityResult := ApplySpecialAbility(actingPlayer, troop.Spec)

		// Remove Queen from hand after use
		actingPlayer.Troops = append(actingPlayer.Troops[:troopIndex], actingPlayer.Troops[troopIndex+1:]...)

		// End turn (even if continue attacking was true)
		if !gs.GameState.CanContinueAttacking {
			gs.GameState.SwitchTurn()
		} else {
			gs.GameState.CanContinueAttacking = false
		}

		return abilityResult, true
	}

	// Validate the target
	if !IsValidTarget(actingPlayer, targetTowerID, gs.GameState) {
		return "Invalid target tower.", false
	}

	// Get the target tower
	var targetTower *TowerInstance
	opponentPlayer := gs.GameState.GetOpponentPlayer()

	if targetTowerID == opponentPlayer.KingTower.ID {
		targetTower = opponentPlayer.KingTower
	} else if targetTowerID == opponentPlayer.GuardTower1.ID {
		targetTower = opponentPlayer.GuardTower1
	} else if targetTowerID == opponentPlayer.GuardTower2.ID {
		targetTower = opponentPlayer.GuardTower2
	}

	// Calculate damage
	damage := CalculateDamage(troop.CurrentATK, targetTower.CurrentDEF)

	// Apply damage to target tower
	targetTower.CurrentHP -= damage

	// Check if tower is destroyed
	towerDestroyed := false
	if targetTower.CurrentHP <= 0 {
		targetTower.CurrentHP = 0
		targetTower.Destroyed = true
		towerDestroyed = true
		gs.GameState.LastDestroyedTowerID = targetTowerID
	}

	// Remove the troop from the player's hand after use
	actingPlayer.Troops = append(actingPlayer.Troops[:troopIndex], actingPlayer.Troops[troopIndex+1:]...)

	// Check win condition
	if targetTower == opponentPlayer.KingTower && targetTower.Destroyed {
		gs.GameState.SetWinner(actingPlayer.Username)
		return fmt.Sprintf("%s's troop %s dealt %d damage to %s and destroyed it! %s wins the game!",
			actingPlayer.Username, troopName, damage, targetTowerID, actingPlayer.Username), true
	}

	// If a tower was destroyed, the player can continue attacking
	if towerDestroyed {
		gs.GameState.CanContinueAttacking = true
		return fmt.Sprintf("%s's troop %s dealt %d damage to %s and destroyed it! You can attack again.",
			actingPlayer.Username, troopName, damage, targetTowerID), true
	}

	// If not continuing attack, switch turn
	if !gs.GameState.CanContinueAttacking {
		gs.GameState.SwitchTurn()
	} else {
		gs.GameState.CanContinueAttacking = false
	}

	return fmt.Sprintf("%s's troop %s dealt %d damage to %s (HP remaining: %d).",
		actingPlayer.Username, troopName, damage, targetTowerID, targetTower.CurrentHP), true
}

// GetGameStateInfo returns a string with the current game state for console display
func (gs *GameSession) GetGameStateInfo() string {
	if gs.GameState.IsGameOver {
		return fmt.Sprintf("Game Over! Winner: %s\n", gs.GameState.Winner)
	}

	playerA := gs.GameState.PlayerA
	playerB := gs.GameState.PlayerB

	info := fmt.Sprintf("Current Turn: %s\n\n", gs.GameState.CurrentTurn)

	// Player A info
	info += fmt.Sprintf("Player A (%s):\n", playerA.Username)
	info += fmt.Sprintf("  King Tower: HP=%d/%d, ATK=%d, DEF=%d, Destroyed=%v\n",
		playerA.KingTower.CurrentHP, playerA.KingTower.Spec.BaseHP,
		playerA.KingTower.CurrentATK, playerA.KingTower.CurrentDEF, playerA.KingTower.Destroyed)
	info += fmt.Sprintf("  Guard Tower 1: HP=%d/%d, ATK=%d, DEF=%d, Destroyed=%v\n",
		playerA.GuardTower1.CurrentHP, playerA.GuardTower1.Spec.BaseHP,
		playerA.GuardTower1.CurrentATK, playerA.GuardTower1.CurrentDEF, playerA.GuardTower1.Destroyed)
	info += fmt.Sprintf("  Guard Tower 2: HP=%d/%d, ATK=%d, DEF=%d, Destroyed=%v\n",
		playerA.GuardTower2.CurrentHP, playerA.GuardTower2.Spec.BaseHP,
		playerA.GuardTower2.CurrentATK, playerA.GuardTower2.CurrentDEF, playerA.GuardTower2.Destroyed)

	info += "  Available Troops:\n"
	for _, troop := range playerA.Troops {
		info += fmt.Sprintf("    %s: HP=%d, ATK=%d, DEF=%d\n",
			troop.Spec.Name, troop.CurrentHP, troop.CurrentATK, troop.CurrentDEF)
	}

	// Player B info
	info += fmt.Sprintf("\nPlayer B (%s):\n", playerB.Username)
	info += fmt.Sprintf("  King Tower: HP=%d/%d, ATK=%d, DEF=%d, Destroyed=%v\n",
		playerB.KingTower.CurrentHP, playerB.KingTower.Spec.BaseHP,
		playerB.KingTower.CurrentATK, playerB.KingTower.CurrentDEF, playerB.KingTower.Destroyed)
	info += fmt.Sprintf("  Guard Tower 1: HP=%d/%d, ATK=%d, DEF=%d, Destroyed=%v\n",
		playerB.GuardTower1.CurrentHP, playerB.GuardTower1.Spec.BaseHP,
		playerB.GuardTower1.CurrentATK, playerB.GuardTower1.CurrentDEF, playerB.GuardTower1.Destroyed)
	info += fmt.Sprintf("  Guard Tower 2: HP=%d/%d, ATK=%d, DEF=%d, Destroyed=%v\n",
		playerB.GuardTower2.CurrentHP, playerB.GuardTower2.Spec.BaseHP,
		playerB.GuardTower2.CurrentATK, playerB.GuardTower2.CurrentDEF, playerB.GuardTower2.Destroyed)

	info += "  Available Troops:\n"
	for _, troop := range playerB.Troops {
		info += fmt.Sprintf("    %s: HP=%d, ATK=%d, DEF=%d\n",
			troop.Spec.Name, troop.CurrentHP, troop.CurrentATK, troop.CurrentDEF)
	}

	return info
}
