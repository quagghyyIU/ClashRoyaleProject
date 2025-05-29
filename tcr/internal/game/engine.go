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

	// Create player A
	playerA := NewPlayer(playerAName) // Initializes with defaults (Lvl 1, 0 EXP, etc)
	profileA, errA := jsonHandler.LoadPlayerData(playerAName)
	if errA != nil {
		log.Printf("Error loading player data for %s: %v. Using default/initial stats.", playerAName, errA)
		// Keep default NewPlayer stats, but ensure RequiredEXP is set based on level 1
		playerA.RequiredEXPForNextLevel = shared.CalculateRequiredEXP(playerA.Level) // Should be 100 for level 1
	} else {
		playerA.Level = profileA.Level
		playerA.CurrentEXP = profileA.CurrentEXP
		playerA.RequiredEXPForNextLevel = profileA.RequiredEXPForNextLevel
		// If loaded profile had 0 for RequiredEXP (e.g. old format or error), recalculate
		if playerA.RequiredEXPForNextLevel == 0 {
			playerA.RequiredEXPForNextLevel = shared.CalculateRequiredEXP(playerA.Level)
		}
	}
	playerA.CurrentMana = shared.InitialMana // Initialize Mana for Enhanced TCR

	// Create player B
	playerB := NewPlayer(playerBName)
	profileB, errB := jsonHandler.LoadPlayerData(playerBName)
	if errB != nil {
		log.Printf("Error loading player data for %s: %v. Using default/initial stats.", playerBName, errB)
		playerB.RequiredEXPForNextLevel = shared.CalculateRequiredEXP(playerB.Level)
	} else {
		playerB.Level = profileB.Level
		playerB.CurrentEXP = profileB.CurrentEXP
		playerB.RequiredEXPForNextLevel = profileB.RequiredEXPForNextLevel
		if playerB.RequiredEXPForNextLevel == 0 {
			playerB.RequiredEXPForNextLevel = shared.CalculateRequiredEXP(playerB.Level)
		}
	}
	playerB.CurrentMana = shared.InitialMana // Initialize Mana for Enhanced TCR

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

	// Check mana cost for non-special troops
	if !troop.Spec.IsSpecialOnly {
		if actingPlayer.CurrentMana < troop.Spec.ManaCost {
			return fmt.Sprintf("Not enough mana to deploy %s. Requires %d, you have %d.", troopName, troop.Spec.ManaCost, actingPlayer.CurrentMana), false
		}
		// Deduct mana only if it's a regular troop and has mana cost
		actingPlayer.CurrentMana -= troop.Spec.ManaCost
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

	// Calculate damage using Enhanced Combat with Crit Chance
	// For now, using DefaultTroopCritChance from shared constants for all troops.
	// This could be made troop-specific later by fetching crit chance from troop.Spec if available.
	damage, didCrit := CalculateDamageEnhanced(troop.CurrentATK, targetTower.CurrentDEF, float64(shared.DefaultTroopCritChance))

	var critMessage string
	if didCrit {
		critMessage = " (CRITICAL HIT!)"
	}

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

		destructionMessage = fmt.Sprintf("%s's troop %s dealt %d damage%s to %s and destroyed it!",
			actingPlayer.Username, troopName, damage, critMessage, targetTowerID)
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
		gameOverMsg := gs.HandleGameOver(actingPlayer.Username, false) // false because it's not a draw
		return destructionMessage + " " + gameOverMsg, true
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

	return fmt.Sprintf("%s's troop %s dealt %d damage%s to %s (HP remaining: %d).",
		actingPlayer.Username, troopName, damage, critMessage, targetTowerID, targetTower.CurrentHP), true
}

// SkipTurn handles a player skipping their turn, granting them bonus mana.
// Returns a message describing what happened and whether the action was successful.
func (gs *GameSession) SkipTurn(playerUsername string) (string, bool) {
	// Check if game is already over
	if gs.GameState.IsGameOver {
		return "Game is already over.", false
	}

	// Get the player who is skipping the turn
	var actingPlayer *Player
	if playerUsername == gs.GameState.PlayerA.Username {
		actingPlayer = gs.GameState.PlayerA
	} else if playerUsername == gs.GameState.PlayerB.Username {
		actingPlayer = gs.GameState.PlayerB
	} else {
		return "Invalid player username.", false
	}

	// Check if it's the player's turn
	if gs.GameState.CurrentTurn != playerUsername {
		return "It's not your turn.", false
	}

	// Calculate mana gain for skipping (1.5x ManaRegenRate)
	// Integer arithmetic: ManaRegenRate + ManaRegenRate / 2
	manaGainOnSkip := shared.ManaRegenRate + (shared.ManaRegenRate / 2)

	oldMana := actingPlayer.CurrentMana
	actingPlayer.CurrentMana += manaGainOnSkip
	if actingPlayer.CurrentMana > shared.MaxMana {
		actingPlayer.CurrentMana = shared.MaxMana
	}
	gainedMana := actingPlayer.CurrentMana - oldMana

	// Prepare message
	skipMessage := fmt.Sprintf("%s skipped their turn and gained %d mana.", actingPlayer.Username, gainedMana)
	log.Printf(skipMessage) // Server-side log

	// Switch turn to the other player.
	// The SwitchTurn() method in state.go will handle giving the *next* player their normal ManaRegenRate.
	gs.GameState.SwitchTurn()
	gs.GameState.LastActionLog = skipMessage // Update last action for client display

	return skipMessage, true
}

// HandleGameOver processes end-of-game logic, including EXP awards and saving player data.
func (gs *GameSession) HandleGameOver(winnerUsername string, isDraw bool) string {
	gs.GameState.IsGameOver = true
	finalMessage := ""

	playerA := gs.GameState.PlayerA
	playerB := gs.GameState.PlayerB

	if isDraw {
		gs.GameState.Winner = "DRAW"
		finalMessage = "The game is a DRAW!"
		log.Printf("Game ended in a draw between %s and %s.", playerA.Username, playerB.Username)

		// Award draw EXP to both players
		playerA.CurrentEXP += shared.DrawEXPReward
		playerB.CurrentEXP += shared.DrawEXPReward
		levelUpMsgA := gs.HandleExperienceAndLevelUp(playerA) // This also saves data
		levelUpMsgB := gs.HandleExperienceAndLevelUp(playerB) // This also saves data

		if levelUpMsgA != "" {
			finalMessage += "\nPlayer A: " + levelUpMsgA
		}
		if levelUpMsgB != "" {
			finalMessage += "\nPlayer B: " + levelUpMsgB
		}

	} else {
		gs.GameState.Winner = winnerUsername
		var winningPlayer *Player
		var losingPlayer *Player // For potential future use, e.g. different save logic

		if winnerUsername == playerA.Username {
			winningPlayer = playerA
			losingPlayer = playerB
		} else {
			winningPlayer = playerB
			losingPlayer = playerA
		}

		finalMessage = fmt.Sprintf("Game Over! Winner: %s!", winnerUsername)
		log.Printf("Game ended. Winner: %s. Loser: %s.", winningPlayer.Username, losingPlayer.Username)

		// Award win EXP to the winner
		winningPlayer.CurrentEXP += shared.WinEXPReward
		levelUpMsgWinner := gs.HandleExperienceAndLevelUp(winningPlayer) // Saves winner's data
		// Save losing player's data as well (they might have gained EXP from destroying units)
		_ = gs.HandleExperienceAndLevelUp(losingPlayer) // We call this to ensure loser's data (like EXP from units) is saved.

		if levelUpMsgWinner != "" {
			finalMessage += "\n" + levelUpMsgWinner
		}
	}

	// Note: HandleExperienceAndLevelUp already saves individual player data.
	// If global game state needed saving, it would be done here.
	return finalMessage
}

// HandleExperienceAndLevelUp checks for player level up and updates stats accordingly.
// It returns a message if the player leveled up, otherwise an empty string.
func (gs *GameSession) HandleExperienceAndLevelUp(player *Player) string {
	leveledUp := false
	levelUpMessage := ""
	for player.CurrentEXP >= player.RequiredEXPForNextLevel && player.RequiredEXPForNextLevel > 0 { // Add check for > 0 to prevent infinite loop if misconfigured
		player.Level++
		player.CurrentEXP -= player.RequiredEXPForNextLevel
		if player.CurrentEXP < 0 {
			player.CurrentEXP = 0
		}
		player.RequiredEXPForNextLevel = shared.CalculateRequiredEXP(player.Level)
		leveledUp = true
	}
	if leveledUp {
		levelUpMessage = fmt.Sprintf("%s leveled up to Level %d! Next level at %d EXP.",
			player.Username, player.Level, player.RequiredEXPForNextLevel)
		log.Println(levelUpMessage) // Server-side log
	}

	// Always save player data after EXP change (level up or not)
	if gs.JSONHandler != nil {
		playerProfile := storage.PlayerProfile{
			Username:                player.Username,
			Level:                   player.Level,
			CurrentEXP:              player.CurrentEXP,
			RequiredEXPForNextLevel: player.RequiredEXPForNextLevel,
		}
		if err := gs.JSONHandler.SavePlayerData(playerProfile); err != nil {
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
