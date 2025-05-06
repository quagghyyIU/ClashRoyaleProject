## Application PDU (Protocol Data Unit) Description

Communication between the TCR Client and Server is achieved by exchanging JSON-formatted messages over TCP. Each message, or Protocol Data Unit (PDU), follows a general structure containing a `Type` field to indicate the nature of the message, and a `Payload` field carrying the specific data for that message type.

### 1. General Message Structure

All PDUs adhere to the following basic JSON structure:

```json
{
  "Type": "MESSAGE_TYPE_STRING",
  "Payload": {
    // Payload content varies based on MESSAGE_TYPE_STRING
  }
}
```

*   **`Type` (string):** A unique string identifier for the kind of message being sent (e.g., `"LOGIN_REQUEST"`, `"GAME_STATE_UPDATE"`).
*   **`Payload` (object):** A JSON object containing the data specific to the message type. If no additional data is needed for a particular message type, the payload can be an empty object (`{}`) or `null`.

### 2. Message Framing

While JSON defines the content, TCP is a stream-based protocol. To ensure the receiver can correctly parse complete JSON messages from the stream, a simple message framing protocol is used:
*Each JSON PDU is preceded by its length as a fixed-size header (e.g., a 4-byte integer representing the length of the upcoming JSON string in bytes).*
*(Alternatively, if you plan to use a newline delimiter: Each JSON PDU is sent as a single line, terminated by a newline character (`\n`). The receiver reads data line by line.)*

**Note:** Choose *one* framing method. Length-prefixing is generally more robust for arbitrary JSON. Newline delimiting is simpler if your JSON never contains unescaped newlines. I'll assume **length-prefixing** for the examples below, but state your choice clearly.

### 3. Common Message Types and Payloads

Below are descriptions of common PDUs used in the TCR application.

#### 3.1. Client-to-Server Messages

##### 3.1.1. `LOGIN_REQUEST`
*   **Type:** `"LOGIN_REQUEST"`
*   **Direction:** Client -> Server
*   **Purpose:** Sent by the client to initiate a session with the server, providing the player's username.
*   **Payload:**
    ```json
    {
      "Username": "player123"
    }
    ```
    *   `Username` (string): The desired username for the player.

##### 3.1.2. `DEPLOY_TROOP_COMMAND`
*   **Type:** `"DEPLOY_TROOP_COMMAND"`
*   **Direction:** Client -> Server
*   **Purpose:** Sent by the client to deploy one of their available troops to target an opponent's tower.
*   **Payload:**
    ```json
    {
      "TroopName": "Pawn", // Name of the troop from the player's hand/available troops
      "TargetTowerID": "PlayerB_GuardTower1" // Unique ID of the opponent's tower to target
    }
    ```
    *   `TroopName` (string): The name of the troop spec the player wishes to deploy (e.g., "Pawn", "Queen").
    *   `TargetTowerID` (string): A unique identifier for the target tower (e.g., constructed as `OpponentUsername_TowerType` like "PlayerB_KingTower", or a simpler ID assigned by the server).

##### 3.1.3. `ATTACK_COMMAND` (For Enhanced TCR or if troops persist and can be commanded)
*   **Type:** `"ATTACK_COMMAND"`
*   **Direction:** Client -> Server
*   **Purpose:** (Primarily for Enhanced TCR) Sent by the client to command an already deployed troop to attack a target.
*   **Payload:**
    ```json
    {
      "AttackingTroopInstanceID": "troop_instance_123", // ID of the player's troop on the field
      "TargetTowerID": "PlayerB_GuardTower2"
    }
    ```
    *   `AttackingTroopInstanceID` (string): Unique ID of the player's troop already on the field.
    *   `TargetTowerID` (string): Unique ID of the opponent's tower to target.
    *(Note: For Simple TCR, deploying a troop might immediately lead to its attack, making a separate `ATTACK_COMMAND` less necessary for the initial deployment action.)*

##### 3.1.4. `CLIENT_READY` (Optional, for explicit game start synchronization)
*   **Type:** `"CLIENT_READY"`
*   **Direction:** Client -> Server
*   **Purpose:** Sent by the client after receiving initial game setup information, indicating they are ready to start the game.
*   **Payload:** `{}` (Empty or null)

#### 3.2. Server-to-Client Messages

##### 3.2.1. `LOGIN_RESPONSE`
*   **Type:** `"LOGIN_RESPONSE"`
*   **Direction:** Server -> Client
*   **Purpose:** Sent by the server in response to a `LOGIN_REQUEST`.
*   **Payload:**
    ```json
    {
      "Success": true,
      "Message": "Login successful. Waiting for opponent...",
      "PlayerID": "player123_sessionXYZ" // Optional: A server-assigned unique ID for the session/player
    }
    ```
    or on failure:
    ```json
    {
      "Success": false,
      "Message": "Username already taken or invalid."
    }
    ```
    *   `Success` (boolean): Indicates if the login was successful.
    *   `Message` (string): A descriptive message for the user.
    *   `PlayerID` (string, optional): A unique ID assigned by the server.

##### 3.2.2. `GAME_START_NOTIFICATION`
*   **Type:** `"GAME_START_NOTIFICATION"`
*   **Direction:** Server -> Client (to both players)
*   **Purpose:** Notifies clients that a game has been found and is starting. Includes initial game setup.
*   **Payload:**
    ```json
    {
      "OpponentUsername": "player_opponent",
      "YourPlayerInfo": { // Info specific to the receiving player
        "Username": "player_self",
        "AssignedTroops": [ // For Simple TCR
          {"Name": "Pawn", "HP": 50, "ATK": 150, "DEF": 100},
          {"Name": "Rook", "HP": 250, "ATK": 200, "DEF": 200},
          {"Name": "Knight", "HP": 200, "ATK": 300, "DEF": 150}
        ],
        "Mana": 5 // For Enhanced TCR
      },
      "InitialGameState": { // Shared initial state, see GAME_STATE_UPDATE
        // ... (subset of GameStateUpdate focusing on initial tower HPs etc.)
      },
      "GameMode": "SimpleTCR" // or "EnhancedTCR"
    }
    ```
    *   `OpponentUsername` (string): The username of the opponent.
    *   `YourPlayerInfo` (object): Information specific to the receiving player.
        *   `AssignedTroops` (array of objects, Simple TCR): List of troops assigned to the player, with their base stats.
        *   `Mana` (int, Enhanced TCR): Starting mana.
    *   `InitialGameState` (object): A snapshot of the initial game state, similar to `GAME_STATE_UPDATE`.
    *   `GameMode` (string): Indicates if it's "SimpleTCR" or "EnhancedTCR".

##### 3.2.3. `GAME_STATE_UPDATE`
*   **Type:** `"GAME_STATE_UPDATE"`
*   **Direction:** Server -> Client (to both players)
*   **Purpose:** Sent by the server to update clients on the current state of the game. This is a key message.
*   **Payload:**
    ```json
    {
      "PlayerA": { // Player who is 'Player A' in this game session
        "Username": "player_A_username",
        "Level": 1, // Enhanced TCR
        "CurrentEXP": 0, // Enhanced TCR
        "RequiredEXPForNextLevel": 100, // Enhanced TCR
        "Mana": 7, // Enhanced TCR
        "Towers": [
          {"ID": "PlayerA_KingTower", "Type": "KingTower", "CurrentHP": 2000, "MaxHP": 2000},
          {"ID": "PlayerA_GuardTower1", "Type": "GuardTower1", "CurrentHP": 1000, "MaxHP": 1000, "Destroyed": false},
          {"ID": "PlayerA_GuardTower2", "Type": "GuardTower2", "CurrentHP": 1000, "MaxHP": 1000, "Destroyed": false}
        ],
        "AvailableTroops": [ // Player's hand/available troops for deployment (Simple TCR)
          {"Name": "Pawn", "Count": 1}, // Or just list names if only one of each
          {"Name": "Bishop", "Count": 1}
        ],
        "DeployedTroops": [ // Troops currently on the field (more relevant for Enhanced TCR)
          {"InstanceID": "troop_abc", "Name": "Rook", "CurrentHP": 250, "Target": "PlayerB_GuardTower1"}
        ]
      },
      "PlayerB": { // Similar structure for Player B
        "Username": "player_B_username",
        // ... same fields as PlayerA
      },
      "CurrentTurn": "player_A_username", // Simple TCR: Username of the player whose turn it is
      "GameTimer": 150, // Enhanced TCR: Remaining game time in seconds
      "LastActionLog": "Player A deployed Pawn targeting Player B's GuardTower1. Damage: 50." // Optional: a human-readable log of the last significant action
    }
    ```
    *   This payload contains a comprehensive snapshot of the game. You might send partial updates for efficiency later, but a full state is simpler initially.
    *   `MaxHP` is useful for client display (e.g., health bars).
    *   `Destroyed` flag for towers.

##### 3.2.4. `ACTION_RESULT`
*   **Type:** `"ACTION_RESULT"`
*   **Direction:** Server -> Client (potentially to both, or just the acting client)
*   **Purpose:** Provides feedback on a specific action taken by a player.
*   **Payload:**
    ```json
    {
      "Success": true,
      "Action": "DEPLOY_TROOP",
      "Message": "Pawn deployed successfully!",
      "DamageDealt": 50, // If applicable
      "Crit": false, // If applicable (Enhanced TCR)
      "HealAmount": 300, // If Queen's ability
      "TargetHPAfter": 950 // If applicable
    }
    ```
    or on failure:
    ```json
    {
      "Success": false,
      "Action": "DEPLOY_TROOP",
      "Message": "Invalid target: GuardTower1 must be destroyed first."
    }
    ```
    *   `Success` (boolean): Whether the action was valid and processed.
    *   `Action` (string): The type of action this result refers to.
    *   `Message` (string): Human-readable feedback.
    *   Other fields provide specific details of the outcome.

##### 3.2.5. `TURN_NOTIFICATION` (Primarily for Simple TCR)
*   **Type:** `"TURN_NOTIFICATION"`
*   **Direction:** Server -> Client (to both players)
*   **Purpose:** Indicates whose turn it is.
*   **Payload:**
    ```json
    {
      "CurrentTurnUsername": "player_who_should_act"
    }
    ```

##### 3.2.6. `GAME_OVER_NOTIFICATION`
*   **Type:** `"GAME_OVER_NOTIFICATION"`
*   **Direction:** Server -> Client (to both players)
*   **Purpose:** Announces the end of the game and the result.
*   **Payload:**
    ```json
    {
      "WinnerUsername": "player_winner", // Can be null or a special value like "DRAW"
      "Reason": "King Tower destroyed.", // Or "Time ran out. Most towers destroyed."
      "AwardedEXP": 30 // EXP awarded for this match (Enhanced TCR)
    }
    ```

##### 3.2.7. `ERROR_NOTIFICATION`
*   **Type:** `"ERROR_NOTIFICATION"`
*   **Direction:** Server -> Client
*   **Purpose:** Sent by the server to inform the client of an error not tied to a specific action result (e.g., internal server error, opponent disconnected).
*   **Payload:**
    ```json
    {
      "ErrorMessage": "Opponent has disconnected. Game cannot continue."
    }
    ```

### 4. Go Structs (Example)

In your Go code (`internal/models/messages.go`), these would translate to structs:

```go
package models

// GenericMessage is the base structure for all messages.
type GenericMessage struct {
    Type    string      `json:"Type"`
    Payload interface{} `json:"Payload"` // Use specific payload structs for unmarshaling
}

// LoginRequestPayload is the payload for LOGIN_REQUEST.
type LoginRequestPayload struct {
    Username string `json:"Username"`
}

// DeployTroopCommandPayload is the payload for DEPLOY_TROOP_COMMAND.
type DeployTroopCommandPayload struct {
    TroopName     string `json:"TroopName"`
    TargetTowerID string `json:"TargetTowerID"`
}

// GameStateUpdatePayload defines the structure for game state updates.
// (This would be a large struct with nested PlayerState, TowerState, etc.)
type TowerState struct {
    ID          string `json:"ID"`
    Type        string `json:"Type"`
    CurrentHP   int    `json:"CurrentHP"`
    MaxHP       int    `json:"MaxHP"`
    Destroyed   bool   `json:"Destroyed"`
}

type PlayerState struct {
    Username                string       `json:"Username"`
    Level                   int          `json:"Level,omitempty"` // omitempty for fields not in SimpleTCR
    CurrentEXP              int          `json:"CurrentEXP,omitempty"`
    RequiredEXPForNextLevel int          `json:"RequiredEXPForNextLevel,omitempty"`
    Mana                    int          `json:"Mana,omitempty"`
    Towers                  []TowerState `json:"Towers"`
    // ... other fields like AvailableTroops, DeployedTroops
}

type GameStateUpdatePayload struct {
    PlayerA         PlayerState `json:"PlayerA"`
    PlayerB         PlayerState `json:"PlayerB"`
    CurrentTurn     string      `json:"CurrentTurn,omitempty"` // Simple TCR
    GameTimer       int         `json:"GameTimer,omitempty"`   // Enhanced TCR
    LastActionLog   string      `json:"LastActionLog,omitempty"`
}

// ... other payload structs ...
```

**Note on `Payload interface{}`:** When decoding on the server or client, you'll typically unmarshal into `GenericMessage` first to get the `Type`. Then, based on the `Type`, you'll unmarshal `Payload` (which might still be a `json.RawMessage`) into the specific payload struct.