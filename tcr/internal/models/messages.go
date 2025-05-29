package models

// Message types
const (
	// Basic connection messages
	MsgTypeLoginRequest      = "LOGIN_REQUEST"
	MsgTypeLoginResponse     = "LOGIN_RESPONSE"
	MsgTypeRegisterRequest   = "REGISTER_REQUEST"
	MsgTypeRegisterResponse  = "REGISTER_RESPONSE"
	MsgTypeErrorNotification = "ERROR_NOTIFICATION"

	// Game-related messages (Phase 3)
	MsgTypeDeployTroopCommand    = "DEPLOY_TROOP_COMMAND"
	MsgTypeGameStartNotification = "GAME_START_NOTIFICATION"
	MsgTypeGameStateUpdate       = "GAME_STATE_UPDATE"
	MsgTypeActionResult          = "ACTION_RESULT"
	MsgTypeTurnNotification      = "TURN_NOTIFICATION"
	MsgTypeGameOverNotification  = "GAME_OVER_NOTIFICATION"
	MsgTypeSkipTurnCommand       = "SKIP_TURN_COMMAND"
)

// GenericMessage is the wrapper for all network messages
type GenericMessage struct {
	Type    string      `json:"type"`    // Message type
	Payload interface{} `json:"payload"` // Message payload
}

// LoginRequestPayload is the payload for a login request
type LoginRequestPayload struct {
	Username string `json:"username"` // Username for login
	Password string `json:"password"` // Password for login
}

// RegisterRequestPayload is the payload for a registration request
type RegisterRequestPayload struct {
	Username string `json:"username"` // Username to register
	Password string `json:"password"` // Password to register
}

// RegisterResponsePayload is the payload for a registration response
type RegisterResponsePayload struct {
	Success bool   `json:"success"` // Whether registration was successful
	Message string `json:"message"` // Success or error message
}

// LoginResponsePayload is the payload for a login response
type LoginResponsePayload struct {
	Success  bool   `json:"success"`  // Whether login was successful
	Message  string `json:"message"`  // Success or error message
	PlayerID string `json:"playerId"` // Optional player ID
}

// ErrorNotificationPayload is the payload for an error notification
type ErrorNotificationPayload struct {
	ErrorMessage string `json:"errorMessage"` // Error message
}

// Phase 3 message payloads

// DeployTroopCommandPayload is sent by client to deploy a troop
type DeployTroopCommandPayload struct {
	TroopName     string `json:"troopName"`     // Name of the troop to deploy
	TargetTowerID string `json:"targetTowerID"` // ID of the target tower
}

// TowerState represents the current state of a tower
type TowerState struct {
	ID        string `json:"id"`        // Tower ID
	Type      string `json:"type"`      // Tower type (KING, GUARD1, GUARD2)
	CurrentHP int    `json:"currentHP"` // Current health points
	MaxHP     int    `json:"maxHP"`     // Maximum health points
	Attack    int    `json:"attack"`    // Attack value
	Defense   int    `json:"defense"`   // Defense value
	Destroyed bool   `json:"destroyed"` // Whether the tower is destroyed
}

// TroopState represents the current state of a troop
type TroopState struct {
	Name     string `json:"name"`     // Troop name
	HP       int    `json:"hp"`       // Health points
	Attack   int    `json:"attack"`   // Attack value
	Defense  int    `json:"defense"`  // Defense value
	ManaCost int    `json:"manaCost"` // Mana cost of the troop (from TroopSpec)
}

// PlayerState represents the current state of a player
type PlayerState struct {
	Username                string       `json:"username"`                // Player's username
	KingTower               TowerState   `json:"kingTower"`               // King tower state
	GuardTower1             TowerState   `json:"guardTower1"`             // Guard tower 1 state
	GuardTower2             TowerState   `json:"guardTower2"`             // Guard tower 2 state
	Troops                  []TroopState `json:"troops"`                  // Available troops
	Level                   int          `json:"level"`                   // Player's current level
	CurrentEXP              int          `json:"currentEXP"`              // Player's current EXP
	RequiredEXPForNextLevel int          `json:"requiredEXPForNextLevel"` // EXP needed for next level
	CurrentMana             int          `json:"currentMana"`             // Player's current mana
	MaxMana                 int          `json:"maxMana"`                 // Player's maximum mana (e.g., 10)
}

// GameStartNotificationPayload is sent by server to notify clients that a game is starting
type GameStartNotificationPayload struct {
	OpponentUsername string      `json:"opponentUsername"` // Opponent's username
	YourPlayerInfo   PlayerState `json:"yourPlayerInfo"`   // Your player info
	GameMode         string      `json:"gameMode"`         // Game mode (SIMPLE or ENHANCED)
}

// GameStateUpdatePayload is sent by server to update clients on the current game state
type GameStateUpdatePayload struct {
	PlayerA       PlayerState `json:"playerA"`       // Player A state
	PlayerB       PlayerState `json:"playerB"`       // Player B state
	CurrentTurn   string      `json:"currentTurn"`   // Username of player whose turn it is
	LastActionLog string      `json:"lastActionLog"` // Optional description of last action
}

// ActionResultPayload is sent by server to notify client of the result of their action
type ActionResultPayload struct {
	Success bool   `json:"success"` // Whether the action was successful
	Action  string `json:"action"`  // Description of the action
	Message string `json:"message"` // Result message
}

// TurnNotificationPayload is sent by server to notify client that it's their turn
type TurnNotificationPayload struct {
	CurrentTurnUsername string `json:"currentTurnUsername"` // Username of player whose turn it is
}

// GameOverNotificationPayload is sent by server to notify clients that the game is over
type GameOverNotificationPayload struct {
	WinnerUsername string `json:"winnerUsername"` // Username of the winner
	Reason         string `json:"reason"`         // Reason for game end
}
