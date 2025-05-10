# Text-Based Clash Royale (TCR)

A text-based implementation of a simplified Clash Royale-like game using Go, TCP networking, and JSON for data exchange.

## Project Overview

TCR is a strategic card game where players deploy troops to attack the opponent's towers. The game has two modes:

1. **Simple TCR**: Turn-based gameplay with basic mechanics
2. **Enhanced TCR**: Real-time gameplay with mana system, critical hits, and player progression

## Game Entities

### Towers
- **King Tower**: Main defense structure, destroying it results in victory
- **Guard Towers (1 & 2)**: Secondary defense structures that must be targeted in a specific order

### Troops
The game features various troops with different stats and abilities:
- **Pawn**: Basic unit with low HP
- **Knight**: Balanced unit with moderate HP and attack
- **Bishop**: Strategic unit with good attack
- **Rook**: Defensive unit with high HP
- **Prince**: Powerful unit with high HP, attack, and defense
- **Queen**: Special unit that heals the friendly tower with the lowest HP

## Configuration

Game data is stored in JSON configuration files under the `configs/` directory:
- `troops.json`: Contains troop specifications
- `towers.json`: Contains tower specifications

## Development Status

Currently in Phase 2 of development, which includes:
- Project structure setup ✅
- Core data structures and configuration files ✅
- Basic loading utilities ✅
- Complete game logic implementation for Simple TCR ✅
- Offline testing functionality ✅
- Basic TCP client-server communication ✅
- Login and user identification ✅

## Simple TCR Game Rules

In Simple TCR mode:
1. Players take turns deploying troops to attack opponent towers
2. The Guard Tower 1 must be destroyed before Guard Tower 2 or King Tower can be targeted
3. When a troop destroys a tower, the player gets an immediate second attack
4. The Queen troop can be deployed to heal the friendly tower with the lowest HP percentage
5. The game ends when a player's King Tower is destroyed

## How to Run

### Phase 1: Offline Testing

To test the Simple TCR game logic in offline mode:

1. Build and run the server with the offline flag:
   ```bash
   cd tcr
   go build ./cmd/server
   ./server -mode offline
   ```

2. Follow the on-screen prompts to play the game:
   - The game starts with Player A's turn
   - Use the `d <troop_name> <tower_number>` command to attack
   - Tower numbers: 1=Guard1, 2=Guard2, 3=King
   - Examples: `d Pawn 1` to attack Guard Tower 1, `d Knight 3` to attack King Tower (if valid)
   - To deploy the Queen and use her healing ability, use: `d Queen 1` (the target number doesn't matter)

### Phase 2: Network Testing

To test the client-server networking:

1. Start the server:
   ```bash
   cd tcr
   go build ./cmd/server
   ./server -addr :8080
   ```

2. In a separate terminal, start a client:
   ```bash
   cd tcr
   go build ./cmd/client
   ./client -addr localhost:8080
   ```

3. Enter a username when prompted.
4. The client will connect to the server and log in.
5. You can start additional clients to test multiple connections.

### Phase 3: Full Networked Game

To play the Simple TCR game over the network in Phase 3:

1. Start the server:
   ```bash
   cd tcr
   go build ./cmd/server
   ./server -addr :8080
   ```

2. Start two clients in separate terminals:
   ```bash
   cd tcr
   go build ./cmd/client
   ./client -addr localhost:8080
   ```

3. Choose to register a new account or login with an existing account.
   - Enter your username and password when prompted.
   - User accounts are stored securely on the server.

4. Once two players are connected, a game will automatically start.
5. Players take turns deploying troops to attack the opponent's towers.
6. Available commands during the game:
   - `d <troop_name>` - Deploy a troop (auto-targeting is enabled)
     - Example: `d Knight` to deploy Knight to attack the next valid target
     - Example: `d Queen` to deploy the Queen and heal your lowest HP tower
   - `status` - Display the current game status
   - `help` - Display available commands
   - `quit` - Exit the game

7. Game rules apply as described above:
   - The client automatically targets enemy towers in the correct sequence:
     - Guard Tower 1 must be destroyed before attacking Guard Tower 2 or King Tower
   - Queen heals your tower with the lowest HP and is consumed
   - A troop that destroys a tower gets a second attack in the same turn
   - The game ends when a player's King Tower is destroyed

## User Account System

TCR now includes a simple user account system:

1. **Registration**: New users can create an account with a username and password.
2. **Authentication**: Users must provide valid credentials to log in.
3. **Session Management**: The server prevents multiple logins with the same account.

User data is stored in JSON format in the `data/users/` directory on the server.

## Network Protocol

TCR uses a simple network protocol based on JSON messages with length-prefixed framing:
1. Each message is encoded as JSON
2. A 4-byte header containing the length of the JSON message is prepended
3. Messages include a type and a payload specific to that type

For details on the protocol and message formats, see `doc/ApplicationPDUDescription.md`.

## Current Limitations

- **Phase 1**: Troop exhaustion - Troops are consumed after deployment and not replenished. Players start with 3 regular troops plus special troops (Queen).
- **Phase 2**: Limited command support - The networking layer currently only supports basic connection and login. Game actions will be implemented in Phase 3.

## Coming Soon

The next phases will implement:
- Integration of networked gameplay for Simple TCR (Phase 3)
- Enhanced TCR with real-time gameplay (Phase 4)
- Mana system (Phase 4)
- Critical hits (Phase 4)
- Player progression with experience and leveling (Phases 4-5)
- Troop replenishment mechanism (Phase 3-4)