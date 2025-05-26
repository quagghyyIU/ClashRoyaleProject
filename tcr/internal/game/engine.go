package game

import (
	"fmt"
	"log"
	"math/rand"
	"tcr/internal/models"
	"tcr/internal/shared"
	"tcr/internal/storage"
	"time"
)

// GameSession represents a game session between two players
type GameSession struct {
	GameState   *GameState
	TroopSpecs  []models.TroopSpec   // Available troops for both players
	TowerSpecs  []models.TowerSpec   // Available towers for both players
	JSONHandler *storage.JSONHandler // Added to save player data
}

// NewGameSession creates a new game session with two players
func NewGameSession(playerAName, playerBName string, troopSpecs []models.TroopSpec, towerSpecs []models.TowerSpec, jsonHandler *storage.JSONHandler) *GameSession {
	// Initialize random seed
	rand.NewSource(time.Now().UnixNano())

	// Create two players
	playerA := NewPlayer(playerAName)
	var err error
	playerA.CurrentEXP, playerA.Level, err = jsonHandler.LoadPlayerData(playerAName)
	if err != nil {
		log.Printf("Error loading player data for %s: %v. Using default stats.", playerAName, err)
		// Keep default NewPlayer stats (Level 1, 0 EXP)
	}
	playerA.RequiredEXPForNextLevel = shared.CalculateRequiredEXP(playerA.Level)

	playerB := NewPlayer(playerBName)
	playerB.CurrentEXP, playerB.Level, err = jsonHandler.LoadPlayerData(playerBName)
	if err != nil {
		log.Printf("Error loading player data for %s: %v. Using default stats.", playerBName, err)
		// Keep default NewPlayer stats (Level 1, 0 EXP)
	}
	playerB.RequiredEXPForNextLevel = shared.CalculateRequiredEXP(playerB.Level)

	// Initialize the game session
	gs := &GameSession{
		TroopSpecs:  troopSpecs,
		TowerSpecs:  towerSpecs,
		JSONHandler: jsonHandler, // Store the handler
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

		// Replenish one troop (even for Queen, as she is consumed)
		gs.replenishTroopForPlayer(actingPlayer)

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
	var destructionMessage string // To store the base destruction message
	if targetTower.CurrentHP <= 0 {
		targetTower.CurrentHP = 0
		targetTower.Destroyed = true
		towerDestroyed = true
		gs.GameState.LastDestroyedTowerID = targetTowerID

		// Award EXP for destroying the tower
		actingPlayer.CurrentEXP += targetTower.Spec.DestroyEXP
		levelUpMessage := gs.HandleExperienceAndLevelUp(actingPlayer)

		destructionMessage = fmt.Sprintf("%s's troop %s dealt %d damage to %s and destroyed it!",
			actingPlayer.Username, troopName, damage, targetTowerID)
		if levelUpMessage != "" {
			destructionMessage += " " + levelUpMessage
		}
	}

	// Remove the troop from the player's hand after use
	actingPlayer.Troops = append(actingPlayer.Troops[:troopIndex], actingPlayer.Troops[troopIndex+1:]...)

	// Replenish one troop
	gs.replenishTroopForPlayer(actingPlayer)

	// Check win condition
	if targetTower == opponentPlayer.KingTower && targetTower.Destroyed {
		gs.GameState.SetWinner(actingPlayer.Username)
		// Award end-game EXP - we'll do this in handleGameOver or similar later
		return fmt.Sprintf("%s %s wins the game!", destructionMessage, actingPlayer.Username), true
	}

	// If a tower was destroyed, the player can continue attacking
	if towerDestroyed {
		gs.GameState.CanContinueAttacking = true
		return fmt.Sprintf("%s You can attack again.", destructionMessage), true
	}

	// If not continuing attack, switch turn
	if !gs.GameState.CanContinueAttacking {
		gs.GameState.SwitchTurn()
	} else {
		// Player was continuing an attack, but didn't destroy another tower.
		// Their bonus turn ends now.
		gs.GameState.CanContinueAttacking = false
		gs.GameState.SwitchTurn()
	}

	return fmt.Sprintf("%s's troop %s dealt %d damage to %s (HP remaining: %d).",
		actingPlayer.Username, troopName, damage, targetTowerID, targetTower.CurrentHP), true
}

// HandleExperienceAndLevelUp checks for player level up and updates stats accordingly.
// It returns a message if the player leveled up, otherwise an empty string.
func (gs *GameSession) HandleExperienceAndLevelUp(player *Player) string {
	leveledUp := false
	levelUpMessage := ""
	for player.CurrentEXP >= player.RequiredEXPForNextLevel {
		player.Level++
		player.CurrentEXP -= player.RequiredEXPForNextLevel
		// It's good practice to ensure CurrentEXP doesn't become negative if it's exactly RequiredEXPForNextLevel
		if player.CurrentEXP < 0 {
			player.CurrentEXP = 0
		}
		player.RequiredEXPForNextLevel = shared.CalculateRequiredEXP(player.Level) // Use shared util
		leveledUp = true
	}
	if leveledUp {
		levelUpMessage = fmt.Sprintf("%s leveled up to Level %d! Next level at %d EXP.",
			player.Username, player.Level, player.RequiredEXPForNextLevel)
		fmt.Println(levelUpMessage) // For server-side logging for now
	}

	// Always save player data after EXP change (level up or not)
	if gs.JSONHandler != nil {
		err := gs.JSONHandler.SavePlayerData(player.Username, player.CurrentEXP, player.Level)
		if err != nil {
			log.Printf("Error saving player data for %s after EXP update: %v", player.Username, err)
		}
	} else {
		log.Printf("Warning: JSONHandler is nil in GameSession. Cannot save player data for %s.", player.Username)
	}

	return levelUpMessage
}

// replenishTroopForPlayer adds a new distinct troop to the player's hand
func (gs *GameSession) replenishTroopForPlayer(player *Player) {
	if len(gs.TroopSpecs) == 0 {
		return // No troops defined to replenish from
	}

	// Create a map of troops currently in hand for quick lookup
	handTroopNames := make(map[string]bool)
	for _, troopInstance := range player.Troops {
		handTroopNames[troopInstance.Spec.Name] = true
	}

	// Find available troop specs not currently in hand
	availableToReplenish := make([]models.TroopSpec, 0)
	for _, spec := range gs.TroopSpecs {
		if !spec.IsSpecialOnly { // Typically replenish with regular troops
			if !handTroopNames[spec.Name] {
				availableToReplenish = append(availableToReplenish, spec)
			}
		}
	}

	// If all regular troops are in hand, or no regular troops are available to replenish,
	// we might allow adding a duplicate or a special troop if desired.
	// For now, if no distinct regular troop is available, do nothing to avoid duplicates.
	if len(availableToReplenish) == 0 {
		// Fallback: if no distinct regular troops, try adding any troop not in hand (including special)
		// This part can be adjusted based on desired game mechanics for replenishment limits.
		for _, spec := range gs.TroopSpecs {
			if !handTroopNames[spec.Name] {
				availableToReplenish = append(availableToReplenish, spec)
			}
		}
		if len(availableToReplenish) == 0 {
			// If still no troops (e.g., player has all defined troops), do nothing.
			return
		}
	}

	// Select a random troop from the available ones
	randIndex := rand.Intn(len(availableToReplenish))
	newTroopSpec := availableToReplenish[randIndex]

	// Add the new troop to the player's hand
	newTroopInstance := NewTroopInstance(&newTroopSpec, fmt.Sprintf("%s_troop_%d", player.Username, len(player.Troops)+1), player.Level)
	player.Troops = append(player.Troops, newTroopInstance)

	// It might be good to send a message to the client that a troop has been replenished.
	// For now, the next GameStateUpdate will show the new troop.
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
