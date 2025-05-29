package game

import "tcr/internal/shared" // Added import for shared constants

// GameState holds all information for a single game
type GameState struct {
	// Pointers to both players
	PlayerA *Player
	PlayerB *Player

	// Current turn - stores the Username of the player whose turn it is
	CurrentTurn string

	// Game status
	IsGameOver bool
	Winner     string // Empty if no winner yet, or PlayerA/PlayerB's username if there's a winner

	// Track the last target destroyed (for the "Continue Attacking" rule)
	LastDestroyedTowerID string
	CanContinueAttacking bool

	// Log of the last action taken for client display
	LastActionLog string
}

// NewGameState creates a new game state with the given players
func NewGameState(playerA, playerB *Player) *GameState {
	return &GameState{
		PlayerA:              playerA,
		PlayerB:              playerB,
		CurrentTurn:          playerA.Username, // PlayerA starts by default
		IsGameOver:           false,
		Winner:               "",
		LastDestroyedTowerID: "",
		CanContinueAttacking: false,
		LastActionLog:        "",
	}
}

// GetCurrentPlayer returns the player whose turn it is now
func (gs *GameState) GetCurrentPlayer() *Player {
	if gs.CurrentTurn == gs.PlayerA.Username {
		return gs.PlayerA
	}
	return gs.PlayerB
}

// GetOpponentPlayer returns the player who is not currently taking their turn
func (gs *GameState) GetOpponentPlayer() *Player {
	if gs.CurrentTurn == gs.PlayerA.Username {
		return gs.PlayerB
	}
	return gs.PlayerA
}

// SwitchTurn changes the turn to the other player and regenerates mana for the new current player.
func (gs *GameState) SwitchTurn() {
	// Determine the player whose turn it will become
	var nextPlayer *Player
	if gs.CurrentTurn == gs.PlayerA.Username {
		gs.CurrentTurn = gs.PlayerB.Username
		nextPlayer = gs.PlayerB
	} else {
		gs.CurrentTurn = gs.PlayerA.Username
		nextPlayer = gs.PlayerA
	}

	// Regenerate mana for the player whose turn it now is
	if nextPlayer != nil {
		nextPlayer.CurrentMana += shared.ManaRegenRate
		if nextPlayer.CurrentMana > shared.MaxMana {
			nextPlayer.CurrentMana = shared.MaxMana
		}
	}

	gs.CanContinueAttacking = false
}

// SetWinner sets the winner of the game and marks the game as over
func (gs *GameState) SetWinner(winnerUsername string) {
	gs.IsGameOver = true
	gs.Winner = winnerUsername
}
