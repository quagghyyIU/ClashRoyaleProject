package game

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

// SwitchTurn changes the turn to the other player
func (gs *GameState) SwitchTurn() {
	if gs.CurrentTurn == gs.PlayerA.Username {
		gs.CurrentTurn = gs.PlayerB.Username
	} else {
		gs.CurrentTurn = gs.PlayerA.Username
	}
	gs.CanContinueAttacking = false
}

// SetWinner sets the winner of the game and marks the game as over
func (gs *GameState) SetWinner(winnerUsername string) {
	gs.IsGameOver = true
	gs.Winner = winnerUsername
}
