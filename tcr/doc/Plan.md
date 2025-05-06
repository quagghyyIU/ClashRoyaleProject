**Detailed Project Plan: Text-Based Clash Royale (TCR)**

**Assumptions Recap (to be integrated into relevant phases):**

1.  **Tower CRIT Values (Appendix):** The `CritChancePercent` in `TowerSpec` will be stored but **initially unused for the tower's own attacks**. CRIT damage in Enhanced TCR applies to the *attacking troop*.
2.  **Troop "Queen":** Deployed, performs her heal, and is then **consumed/removed**. Her `BaseHP`, `BaseATK`, `BaseDEF` will be 0 in `troops.json`, and `IsSpecialOnly` will be `true`.
3.  **"EXP" Values (Appendix):** These are rewards for destroying specific units, awarded *immediately* during Enhanced TCR, in addition to end-of-match EXP.

---

**1. Core Philosophy & Technology Stack**

*   **Language:** Go
*   **Protocol:** TCP
*   **Data Format (Network):** JSON
*   **Data Format (Storage):** JSON
*   **Concurrency:** Goroutines and Channels

---

**2. Project Structure (Directory Layout)**

```
tcr/
├── cmd/
│   ├── server/
│   │   └── main.go
│   └── client/
│       └── main.go
├── internal/
│   ├── game/
│   │   ├── engine.go
│   │   ├── entities.go
│   │   ├── combat.go
│   │   ├── rules.go
│   │   └── state.go
│   ├── models/
│   │   ├── messages.go
│   │   └── specs.go
│   ├── network/
│   │   ├── server.go
│   │   ├── client.go
│   │   └── protocol.go
│   ├── storage/
│   │   └── json_handler.go
│   └── shared/
│       └── constants.go
│       └── utils.go
├── configs/
│   ├── troops.json
│   └── towers.json
├── data/
│   └── players/
└── go.mod
└── go.sum
```

---

**3. Core Components & Modules**

*   **`models/specs.go`:**
    *   `TroopSpec`: `Name` (string), `BaseHP` (int), `BaseATK` (int), `BaseDEF` (int), `ManaCost` (int), `DestroyEXP` (int), `SpecialAbility` (string, e.g., "HEAL_LOWEST_HP_TOWER_300"), `IsSpecialOnly` (bool).
    *   `TowerSpec`: `Name` (string), `Type` (string), `BaseHP` (int), `BaseATK` (int), `BaseDEF` (int), `CritChancePercent` (int), `DestroyEXP` (int).
*   **`configs/troops.json` & `configs/towers.json`:**
    *   Populate accurately from Appendix. Queen: `BaseHP: 0, BaseATK: 0, BaseDEF: 0, IsSpecialOnly: true`.
*   **`game/entities.go`:**
    *   `Player`: `Username` (string), `KingTower`, `GuardTower1`, `GuardTower2` (*`TowerInstance`), `AvailableTroops` ([]`TroopInstance` or []`specs.TroopSpec` for hand).
        *   *(Enhanced)*: `CurrentEXP` (int), `Level` (int), `CurrentMana` (int), `RequiredEXPForNextLevel` (int).
    *   `TowerInstance`: `Spec` (`specs.TowerSpec`), `CurrentHP` (int), `EffectiveATK` (int), `EffectiveDEF` (int), `ID` (string).
    *   `TroopInstance`: `Spec` (`specs.TroopSpec`), `CurrentHP` (int), `EffectiveATK` (int), `EffectiveDEF` (int), `InstanceID` (string, if they persist on field).
*   **`game/combat.go`:**
    *   `CalculateDamage(attackerEffectiveATK int, defenderEffectiveDEF int) int`: Implements `DMG = ATK_A - DEF_B (if ≥ 0)`.
    *   *(Enhanced)* `CalculateDamageEnhanced(attackerEffectiveATK int, defenderEffectiveDEF int, attackerCritChancePercent float64) (damage int, didCrit bool)`: Uses a global constant for `attackerCritChancePercent` (e.g., 15-20%) initially. `DMG = ATK_A or (ATK_A * 1.2 if CRIT) - DEF_B`.
*   **`game/rules.go`:**
    *   `IsValidTarget(attackingPlayer Player, targetTowerID string, gameState state.GameState) bool`: Implements G1 -> (G2 or King) logic.
    *   `CanDeployTroop(player Player, troopName string, gameState state.GameState) bool`: Checks hand, mana (Enhanced).
    *   `ApplySpecialAbility(actingPlayer *Player, friendlyTowers []*TowerInstance, opponentTowers []*TowerInstance, troopSpec specs.TroopSpec, gameState *state.GameState)`: Handles Queen's heal.
*   **`game/engine.go`:**
    *   `GameSession`: Manages `GameState`.
        *   Handles EXP awards (unit destruction, match end) and leveling logic for Enhanced TCR.
        *   Handles Queen deployment: call `ApplySpecialAbility`, then remove Queen from play.
        *   Leveling stat calculation: `EffectiveStat = BaseStat * (1 + (Player.Level-1)*0.10)`. (Level 1 = base stats).
        *   `RequiredEXPForNextLevel` calculation: `current_req * 1.10`.

---

**4. Development Phases (Detailed Breakdown)**

**Phase 0: Project Setup & Foundational Structures**
*   **Objective:** Establish project structure, define core data models from Appendix, and enable loading of game configurations.
*   **Grading Component:** Foundational for all.
*   **Tasks:**
    1.  Initialize Go module (`go mod init tcr`).
    2.  Create the directory structure as outlined in Section 2.
    3.  **Define Structs (`models/specs.go`):**
        *   Implement `TroopSpec` and `TowerSpec` structs precisely matching the fields detailed in Section 3 (Core Components).
    4.  **Create Config Files (`configs/`):**
        *   Create `troops.json` and `towers.json`.
        *   Populate these files meticulously with data from the Appendix.
            *   For Queen: `{"Name": "Queen", "BaseHP": 0, "BaseATK": 0, "BaseDEF": 0, "ManaCost": 5, "DestroyEXP": 30, "SpecialAbility": "HEAL_LOWEST_HP_TOWER_300", "IsSpecialOnly": true}`.
            *   Ensure all other troops and towers are accurately represented.
    5.  **Implement Config Loading (`storage/json_handler.go`):**
        *   Create `LoadTroopSpecs(filePath string) ([]specs.TroopSpec, error)`.
        *   Create `LoadTowerSpecs(filePath string) ([]specs.TowerSpec, error)`.
        *   Include basic error handling (file not found, JSON parsing errors).
    6.  **Define Game Entities (`game/entities.go`):**
        *   Implement initial `Player`, `TowerInstance`, and `TroopInstance` structs. Focus on fields needed for Simple TCR first (e.g., `Username`, `CurrentHP`). `EffectiveATK/DEF` can be same as base initially.
*   **Documentation:**
    *   Start `README.md`: Project title, brief description.
    *   Add comments to `models/specs.go` explaining each field.
    *   Document the structure of `configs/troops.json` and `configs/towers.json` (either in `README.md` or a separate `CONFIG_GUIDE.md`).

**Phase 1: Simple TCR - Core Game Logic (Offline Simulation)**
*   **Objective:** Implement the complete game logic for Simple TCR, testable without networking.
*   **Grading Component: Simple TCR (40 points - Core Logic)**
*   **Tasks:**
    1.  **Implement Basic Combat (`game/combat.go`):**
        *   Create `CalculateDamage(attackerEffectiveATK int, defenderEffectiveDEF int) int`.
    2.  **Define Game State (`game/state.go`):**
        *   Create `GameState` struct to hold all information for a single game (pointers to both `Player` objects, current turn, etc.).
    3.  **Implement Game Initialization (`game/engine.go`):**
        *   Function to create a new `GameState` for two players.
        *   Assign 3 `TowerInstance`s to each player (King, Guard1, Guard2) based on loaded `TowerSpec`s.
        *   Randomly assign 3 unique `TroopSpec`s (from loaded specs) to each player's hand (`AvailableTroops`).
    4.  **Implement Core Turn Actions (`game/engine.go`, `game/rules.go`):**
        *   **Troop Deployment:** Logic for a player to select a troop from their hand and a target.
        *   **Target Validation (`game/rules.go`):** Implement `IsValidTarget` for Simple TCR (G1 must be destroyed before G2 or King).
        *   **Attack Execution:**
            *   Selected troop attacks the target tower. Apply damage using `CalculateDamage`.
            *   Update tower's `CurrentHP`.
            *   If tower HP <= 0, mark it as destroyed.
        *   **"Continue Attacking" Rule:** If a troop destroys a tower, it gets an immediate second attack phase in the *same* player's turn against a *new valid* target (if available and chosen).
        *   **Queen's Special Ability (`game/rules.go`, `game/engine.go`):**
            *   Implement `ApplySpecialAbility`. If Queen is deployed, this function is called.
            *   It should find the friendly tower with the lowest current HP (percentage might be better if max HPs vary significantly beyond Guard/King) and heal it by 300 (not exceeding max HP).
            *   After healing, the Queen is removed from the game (not added as a persistent `TroopInstance` on the field).
    5.  **Implement Win Condition (`game/engine.go`):**
        *   Check after each attack if an opponent's King Tower is destroyed.
    6.  **Turn Management (`game/engine.go`):**
        *   Logic to switch turns between players.
    7.  **Offline Testing:**
        *   Create a temporary `main_test.go` or a function in `cmd/server/main.go` to simulate a full game flow by calling engine functions directly. Print state changes to console. Test various scenarios, including Queen's deployment and win conditions.
*   **Documentation:**
    *   Document functions in `game/engine.go`, `game/combat.go`, `game/rules.go`, `game/state.go`.
    *   Explain the Simple TCR turn flow and rules implemented in `README.md`.

**Phase 2: Basic Networking & Client-Server Shell**
*   **Objective:** Establish basic TCP client-server communication, including connection, message framing, and simple user identification.
*   **Grading Component: Simple TCR (40 points - Network Foundation)**
*   **Tasks:**
    1.  **Implement Protocol (`network/protocol.go`):**
        *   Implement `Encode(v interface{}) ([]byte, error)` (JSON marshal).
        *   Implement `Decode(data []byte, v interface{}) error` (JSON unmarshal).
        *   **Implement Message Framing:** Decide and implement:
            *   **Length-prefixing (Recommended):** Send a fixed-size (e.g., 4-byte) header indicating the length of the JSON message before sending the JSON string. The receiver reads the length, then reads that many bytes.
            *   *Or Newline Delimiter:* Send each JSON message as a single line ending with `\n`. Simpler, but JSON must not contain unescaped newlines.
    2.  **Define Basic Network Messages (`models/messages.go`):**
        *   `GenericMessage { Type string; Payload interface{} }`.
        *   `LoginRequestPayload { Username string }`.
        *   `LoginResponsePayload { Success bool; Message string; PlayerID string (optional) }`.
        *   `ErrorNotificationPayload { ErrorMessage string }`.
    3.  **Implement Basic Server (`network/server.go`, `cmd/server/main.go`):**
        *   Listen for TCP connections on a configurable port.
        *   For each new connection, spawn a goroutine to handle it.
        *   In the handler goroutine:
            *   Read `LOGIN_REQUEST`.
            *   Store username with connection (e.g., in a map).
            *   Send `LOGIN_RESPONSE`.
        *   Basic matchmaking: Store the first two connected (and logged-in) clients for a potential game.
    4.  **Implement Basic Client (`network/client.go`, `cmd/client/main.go`):**
        *   Connect to server (IP/port from args or hardcoded).
        *   Prompt user for username.
        *   Send `LOGIN_REQUEST`.
        *   Receive and display `LOGIN_RESPONSE` or `ERROR_NOTIFICATION`.
        *   Keep connection open to receive further messages.
    5.  **Testing:** Test connection, login, and simple message exchange. Ensure message framing works correctly.
*   **Documentation:**
    *   Document `network/protocol.go` including the chosen message framing technique.
    *   Document the initial PDU structures in `README.md` or `PDU_DESCRIPTION.md`.
    *   Add basic usage instructions for client/server to `README.md`.

**Phase 3: Integrating Simple TCR with Networking**
*   **Objective:** Make the Simple TCR game fully playable over the network between two clients.
*   **Grading Component: Simple TCR (40 points - Completion)**
*   **Tasks:**
    1.  **Expand Network Messages (`models/messages.go`):**
        *   `DeployTroopCommandPayload { TroopName string; TargetTowerID string }`.
        *   `GameStartNotificationPayload { OpponentUsername string; YourPlayerInfo object; InitialGameState object; GameMode string }`.
        *   `GameStateUpdatePayload { PlayerA PlayerState; PlayerB PlayerState; CurrentTurn string; LastActionLog string (optional) }` (define `PlayerState`, `TowerState` within).
        *   `ActionResultPayload { Success bool; Action string; Message string; ...details }`.
        *   `TurnNotificationPayload { CurrentTurnUsername string }`.
        *   `GameOverNotificationPayload { WinnerUsername string; Reason string }`.
    2.  **Server-Side Game Integration (`network/server.go`, `game/engine.go`):**
        *   When two players are matched:
            *   Create a `GameSession` instance using the Simple TCR logic from Phase 1.
            *   Send `GameStartNotification` to both clients with initial setup (their troops, opponent, initial tower HPs).
        *   In client handler goroutines:
            *   Listen for `DeployTroopCommand`.
            *   Pass command to the `GameSession`.
            *   `GameSession` validates (player's turn, valid troop/target using `game/rules.go`) and executes the action (updating its internal `GameState`).
            *   `GameSession` returns an action result.
        *   After processing an action:
            *   Send `ActionResultPayload` to the acting client.
            *   Broadcast `GameStateUpdatePayload` to both clients in the session.
            *   If game over, send `GameOverNotificationPayload`.
            *   Send `TurnNotificationPayload` to indicate next turn.
    3.  **Client-Side UI and Logic (`cmd/client/main.go`, `network/client.go`):**
        *   On receiving `GameStartNotification`: Display initial game info.
        *   On receiving `GameStateUpdate`: Parse and display the current state (tower HPs, own troops, opponent's visible state, whose turn).
        *   On receiving `TurnNotification` (if it's this client's turn): Prompt for action (e.g., "deploy <troop> <target>").
        *   Parse user input and construct `DeployTroopCommand`. Send to server.
        *   Display `ActionResultPayload` and `GameOverNotificationPayload`.
    4.  **Testing:** Play full Simple TCR games between two client instances. Test all rules, Queen's ability, win conditions over the network.
*   **Documentation:**
    *   Update PDU descriptions with all new messages.
    *   Document the client command syntax.
    *   Provide clear instructions in `README.md` on how to run a 2-player Simple TCR game.

**Phase 4: Enhanced TCR - Core Mechanics Implementation**
*   **Objective:** Implement Continuous Play, Mana System, Critical Damage, and basic EXP/Leveling.
*   **Grading Components:**
    *   **Continuous Play (10 points)**
    *   **Mana System (10 points)**
    *   **Critical Damage Logic (5 points)**
    *   **Leveling System (10 points - Partial: EXP gain, basic leveling)**
*   **Tasks:**
    1.  **Update Entities & Specs (`models/specs.go`, `game/entities.go`):**
        *   Add `ManaCost` to `TroopSpec`. Ensure `troops.json` reflects this.
        *   Add `CurrentEXP`, `Level`, `CurrentMana`, `RequiredEXPForNextLevel` to `Player` entity.
    2.  **Player Data Persistence (`storage/json_handler.go`):**
        *   Implement `LoadPlayerData(username string) (game.PlayerProfile, error)` (define `PlayerProfile` struct for stored data like EXP, Level).
        *   Implement `SavePlayerData(playerProfile game.PlayerProfile) error`.
        *   Server loads this on login (if Enhanced mode), saves on game end/disconnect. Create `data/players/` directory.
    3.  **Critical Damage (`game/combat.go`):**
        *   Implement `CalculateDamageEnhanced`.
        *   **Decision:** Use a global constant (e.g., in `shared/constants.go`) for `AttackerCritChancePercent` (e.g., `DefaultTroopCritChance = 0.15` for 15%). CRIT damage is `ATK_A * 1.2`.
    4.  **Game Engine Modifications for Enhanced TCR (`game/engine.go`):**
        *   **Game Mode Selection:** Server needs a way to start an Enhanced TCR game (e.g., config flag, or if player data exists).
        *   **Game Timer:** Implement a 3-minute timer for the game session.
        *   **Continuous Play:** Remove turn-based logic. Server processes commands from either player as they arrive (if valid).
        *   **Mana System:**
            *   Player `CurrentMana` starts at 5.
            *   Regenerate 1 mana per second, up to a max of 10.
            *   Troop deployment consumes mana based on `TroopSpec.ManaCost`. Check in `CanDeployTroop`.
        *   **EXP Awards:**
            *   When a unit (troop or tower) is destroyed, award its `TroopSpec.DestroyEXP` or `TowerSpec.DestroyEXP` to the opposing player. Update `Player.CurrentEXP`.
            *   At game end: Win: 30 EXP, Draw: 10 EXP. Update `Player.CurrentEXP`.
        *   **Basic Leveling Logic:**
            *   After any EXP gain, check if `Player.CurrentEXP >= Player.RequiredEXPForNextLevel`.
            *   If so: `Player.Level++`, `Player.CurrentEXP -= Player.RequiredEXPForNextLevel` (or reset EXP for new level), `Player.RequiredEXPForNextLevel = int(float64(Player.RequiredEXPForNextLevel) * 1.10)`.
            *   Initialize `RequiredEXPForNextLevel` for Level 1 to a base value (e.g., 100).
        *   **Update Win Conditions:** King Tower destroyed first, OR most towers destroyed when timer runs out. If tower counts are equal at timeout, it's a draw.
    5.  **Network Message Updates (`models/messages.go`):**
        *   `GameStateUpdatePayload` needs to include `Mana`, `Level`, `EXP`, `GameTimer`.
        *   Consider an `AttackCommandPayload` if troops persist and can be re-targeted. For now, deploying a troop might still be its primary attack action.
    6.  **Client UI Updates (`cmd/client/main.go`):**
        *   Display Mana, EXP, Level, Game Timer.
        *   Allow continuous command input (respecting mana).
    7.  **Testing:** Test all new Enhanced TCR mechanics: timer, mana regen/cost, CRITs, EXP from units, match-end EXP, basic level-ups.
*   **Documentation:**
    *   Document Enhanced TCR rules in `README.md`.
    *   Detail new PDU fields.
    *   Explain player data storage (`data/players/*.json`).
    *   Document CRIT chance implementation.

**Phase 5: Enhanced TCR - Leveling System Stat Scaling & Full Integration**
*   **Objective:** Complete the Leveling System by implementing stat scaling and ensure all Enhanced features work cohesively.
*   **Grading Component: Leveling System (10 points - Completion)**
*   **Tasks:**
    1.  **Implement Stat Scaling (`game/engine.go` or when instances are created/used):**
        *   When a player deploys a troop, its `EffectiveATK` and `EffectiveDEF` (and `CurrentHP` from `BaseHP`) are calculated: `EffectiveStat = BaseStat * (1 + (Player.Level-1)*0.10)`. (Level 1 uses base stats).
        *   Player's tower stats (`EffectiveATK`, `EffectiveDEF`, `MaxHP`) are also scaled by their owner's level using the same formula. This applies when damage is calculated against them or if they were to attack.
    2.  **Refine `RequiredEXPForNextLevel`:** Ensure the 10% increase per level is correctly applied and stored.
    3.  **Comprehensive Testing:**
        *   Test stat scaling: verify troops/towers have increased stats at higher player levels.
        *   Test Queen's heal with scaled tower MaxHP.
        *   Test full game loops of Enhanced TCR with leveling, ensuring EXP and stats progress correctly.
        *   Test edge cases for win/draw conditions.
    4.  **Code Review & Refinement:** Review game logic for clarity, efficiency, and correctness. Improve error handling.
*   **Documentation:**
    *   Detail the stat scaling formula and its application in `README.md`.
    *   Finalize all game rule descriptions.

**Phase 6: Documentation & Final Polish**
*   **Objective:** Produce comprehensive documentation and ensure the project is well-polished.
*   **Grading Component: Documentation (20 points)**
*   **Tasks:**
    1.  **Code Comments:** Ensure all public functions, structs, and complex logic blocks have clear GoDoc comments.
    2.  **Finalize `README.md`:**
        *   Project Overview.
        *   System Architecture (as drafted previously).
        *   Application PDU Description (as drafted previously).
        *   Detailed Build and Run Instructions (for server and client).
        *   Explanation of Game Rules (Simple and Enhanced, clearly distinct).
        *   Features Implemented (can map to grading criteria).
        *   Directory Structure Overview.
        *   Known Limitations or Future Improvement Ideas.
    3.  **User Guide (Simple):** Briefly explain how to play the game using the client commands.
    4.  **Review All Documentation:** Check for clarity, accuracy, and completeness.
    5.  **Final Testing:** One last round of testing for major functionalities.
    6.  **Code Cleanup:** Remove dead code, format consistently.

**Phase 7: Bonus Features (Optional)**
*   **Objective:** Implement additional features beyond the core requirements if time permits.
*   **Grading Component: Bonus Features (5 points)**
*   **Tasks (Examples):**
    *   Add more troops/towers with unique abilities from your own design.
    *   Implement troop-specific CRIT chances (add `CritChancePercent` to `TroopSpec`).
    *   Basic spectator mode.
    *   Persistent chat between players in a game.
    *   More sophisticated CLI (e.g., using a library like `cobra`).
*   **Documentation:** Document any bonus features implemented.

---