## System Architecture

This document outlines the architecture of the Text-Based Clash Royale (TCR) application. The system is designed as a client-server model, enabling two players to connect and engage in a text-based strategy game.

### 1. Overview

The TCR system consists of two main components:

1.  **TCR Server (`cmd/server`):** A Go application responsible for managing game logic, player connections, and game state.
2.  **TCR Client (`cmd/client`):** A Go command-line application that players use to connect to the server, send commands, and receive game updates.

Communication between the client and server is facilitated over **TCP/IP**, using **JSON** as the message format for data exchange.

```mermaid
graph TD
    subgraph Player 1 Machine
        Client1[TCR Client 1]
    end
    subgraph Player 2 Machine
        Client2[TCR Client 2]
    end
    subgraph Server Machine
        ServerApp[TCR Server]
        ServerApp --- G1(Game Session 1)
        G1 --- P1Data[Player 1 State]
        G1 --- P2Data[Player 2 State]
        GameConfigs[Game Configs (troops.json, towers.json)] -.-> ServerApp
        PlayerDataStore[Player Data (players/*.json) for Enhanced TCR] <-.-> ServerApp
    end

    Client1 -- TCP/JSON --> ServerApp
    Client2 -- TCP/JSON --> ServerApp
    ServerApp -- TCP/JSON --> Client1
    ServerApp -- TCP/JSON --> Client2

    style Client1 fill:#D6EAF8,stroke:#333,stroke-width:2px
    style Client2 fill:#D6EAF8,stroke:#333,stroke-width:2px
    style ServerApp fill:#D5F5E3,stroke:#333,stroke-width:2px
    style G1 fill:#FCF3CF,stroke:#333,stroke-width:1px
```

*(This Mermaid diagram provides a visual. You can generate an image from this or include the code block if your Markdown renderer supports Mermaid.)*

### 2. Components in Detail

#### 2.1. TCR Server

The server is the authoritative component of the game. Its primary responsibilities include:

*   **Connection Management (`internal/network/server.go`):**
    *   Listens for incoming TCP connections from clients on a configured port.
    *   Manages each connected client in a separate goroutine, allowing concurrent player interactions.
    *   Handles basic client authentication (username for session identification).
*   **Player Matchmaking:**
    *   Pairs two authenticated clients to start a new game session.
*   **Game Session Management (`internal/game/engine.go`):**
    *   Instantiates and manages `GameSession` objects for each active game.
    *   Each `GameSession` encapsulates the state and logic for one match between two players.
    *   Processes game actions (e.g., troop deployment, attacks) received from clients.
    *   Enforces game rules (Simple TCR and Enhanced TCR) via `internal/game/rules.go`.
    *   Updates the game state (`internal/game/state.go`) based on actions and rules.
*   **Game Logic (`internal/game/`):**
    *   **Entities (`entities.go`):** Defines structures for Players, Towers, and Troops, including their stats and current state.
    *   **Combat (`combat.go`):** Implements damage calculation, including CRIT logic for Enhanced TCR.
    *   **Rules (`rules.go`):** Validates player actions against game rules (targeting, deployment conditions, mana costs).
    *   **Special Abilities:** Handles unique troop abilities (e.g., Queen's heal).
*   **State Synchronization:**
    *   Broadcasts game state updates and event notifications to connected clients in a game session to ensure both players have a consistent view.
*   **Data Persistence (`internal/storage/json_handler.go`):**
    *   Loads initial game specifications (troop and tower stats) from `configs/*.json` files at startup.
    *   For Enhanced TCR, loads and saves player profiles (EXP, level) from/to `data/players/*.json`.
*   **Concurrency:** Utilizes goroutines for handling multiple client connections and game sessions concurrently. Mutexes are used within `GameSession` to protect shared game state from race conditions.

#### 2.2. TCR Client

The client application provides the user interface for players. Its responsibilities include:

*   **Server Connection (`internal/network/client.go`):**
    *   Establishes a TCP connection to the TCR server.
    *   Handles sending the player's chosen username for identification.
*   **User Interface (`cmd/client/main.go`):**
    *   Provides a text-based command-line interface (CLI).
    *   Prompts the user for actions (e.g., deploying troops, selecting targets).
    *   Displays game information received from the server (e.g., tower health, player mana, game events, opponent's actions).
*   **Command Sending:**
    *   Translates user input into structured JSON messages (`internal/models/messages.go`).
    *   Sends these command messages to the server.
*   **Update Receiving:**
    *   Receives JSON messages from the server (e.g., game state updates, turn notifications, error messages).
    *   Parses these messages and updates the display accordingly.

#### 2.3. Network Protocol (`internal/network/protocol.go`)

*   **Transport Protocol:** TCP/IP is used for reliable, ordered delivery of messages between clients and the server.
*   **Message Format:** JSON is used to serialize and deserialize data for network transmission. This provides a human-readable and flexible format.
    *   Standardized message structures are defined in `internal/models/messages.go`.
*   **Message Framing:** A mechanism (e.g., sending message length before the message body, or using a delimiter) is implemented to ensure complete JSON messages are read from the TCP stream.

#### 2.4. Configuration and Data Files

*   **`configs/troops.json` & `configs/towers.json`:**
    *   Store the base specifications (HP, ATK, DEF, Mana Cost, EXP rewards, Special Abilities) for all troops and towers.
    *   Loaded by the server at startup.
*   **`data/players/` (for Enhanced TCR):**
    *   Stores individual player data (e.g., `player_username.json`) containing their EXP and level.
    *   Accessed by the server to load player progress and save updates.

### 3. Game Flow (High-Level)

1.  **Server Startup:** The server starts, loads game configurations (troops, towers).
2.  **Client Connection:** Players launch client applications, which connect to the server and send their usernames.
3.  **Matchmaking:** The server pairs two connected clients.
4.  **Game Initialization:** A new `GameSession` is created on the server for the paired players. Initial game state (towers, randomly assigned troops for Simple TCR) is set up.
5.  **Gameplay (Simple TCR):**
    *   Server notifies players whose turn it is.
    *   The active player's client sends a command (e.g., deploy troop).
    *   Server validates and processes the command, updates the game state.
    *   Server broadcasts the updated game state/action result to both clients.
    *   Turns alternate until a win condition is met.
6.  **Gameplay (Enhanced TCR):**
    *   Game runs for a 3-minute timer.
    *   Players continuously send commands (deploy, attack) if they have mana/valid actions.
    *   Server processes commands, updates mana, checks for CRITs, awards EXP for destroyed units, and updates player levels.
    *   Server broadcasts frequent game state updates.
    *   Winner determined by King Tower destruction or most towers destroyed at timeout.
7.  **Game End:** Server notifies clients of the game outcome. For Enhanced TCR, player EXP and levels are updated and saved.
8.  **Client Disconnection:** Clients can disconnect, or the server handles disconnections.

### 4. Key Design Choices

*   **Go Language:** Chosen for its strong support for concurrency (goroutines, channels), networking capabilities, and efficient performance.
*   **TCP Protocol:** Selected for its reliability and ordered message delivery, simplifying state management compared to UDP for this turn-based/session-based game.
*   **JSON for Communication:** Offers a human-readable, widely supported, and flexible data interchange format.
*   **Centralized Server Logic:** The server is the single source of truth for game state and rules, preventing cheating and ensuring consistency.
*   **Modular Design:** The codebase is organized into internal packages (`game`, `network`, `models`, `storage`) to promote separation of concerns and maintainability.