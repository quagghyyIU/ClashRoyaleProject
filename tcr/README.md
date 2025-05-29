# Text-Based Clash Royale (TCR)

A text-based implementation of a simplified Clash Royale-like game using Go, TCP networking, and JSON for data exchange.

## Project Overview

TCR is a strategic card game where players deploy troops to attack the opponent's towers. The game currently emphasizes turn-based mechanics with a mana system, critical hits, and player progression.

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
(See `CONFIG_GUIDE.md` for more details, including where core constants like mana rates are defined.)

## Development Status

Currently, the project has completed most **Phase 3** features and is incorporating elements of **Phase 4**.
Key implemented features include:
- Robust project structure
- Core data structures and configuration file loading
- Complete game logic implementation for turn-based play
- Offline testing functionality
- TCP client-server communication
- User registration, login, and basic session management
- Networked gameplay with commands: deploy, skip, status, help, quit
- Mana system (initial/regen/max)
- Basic critical hit system
- Player experience (EXP) and leveling system
- Troop replenishment after deployment

## Simple TCR Game Rules (Current Implementation)

1. Players take turns deploying troops to attack opponent towers.
2. Mana is required to deploy most troops.
3. The Guard Tower 1 must be destroyed before Guard Tower 2 or King Tower can be targeted.
4. When a troop destroys a tower, the player gets an immediate second attack opportunity in the same turn.
5. The Queen troop can be deployed to heal the friendly tower with the lowest HP percentage (consumes troop, costs mana like other special abilities if applicable).
6. Players can `skip` their turn to gain a 1.5x mana regeneration bonus for that turn.
7. The game ends when a player's King Tower is destroyed.

## How to Run

### Server
1. Navigate to the `tcr` directory: `cd tcr`
2. Build the server: `go build ./cmd/server`
3. Run the server: `./server` (defaults to online mode on port :8080)
   - For offline testing: `./server -mode offline`

### Client (for Online Mode)
1. Navigate to the `tcr` directory: `cd tcr` (in a separate terminal)
2. Build the client: `go build ./cmd/client`
3. Run the client: `./client` (defaults to connect to `localhost:8080`)
   - Run two client instances for a networked game.

### Automated Build & Run (Windows Batch File)
A `run_game.bat` script is available in the project root. Double-click it to automatically build the server and client, then launch the server and two client windows.

### Gameplay (Online Mode)
1. Choose to register a new account or login with an existing account on each client.
2. Once two players are connected and logged in, a game will automatically start.
3. Available commands during the game:
   - `d <troop_name>` - Deploy a troop (auto-targeting is enabled)
     - Example: `d Knight`
     - Example: `d Queen` (heals your lowest HP tower)
   - `skip` - Skip your turn and gain bonus mana (1.5x normal regeneration)
   - `status` - Display the current game status
   - `help` - Display available commands
   - `quit` - Forfeit the game and exit

## User Account System

TCR includes a user account system:
1. **Registration**: New users can create an account with a username and password.
2. **Authentication**: Users must provide valid credentials to log in.
3. **Session Management**: The server prevents multiple logins with the same account.
User data is stored in JSON format in the `data/users/` directory on the server.

## Recent Gameplay Enhancements

- **Mana Economy Update**: The game now features a more generous mana system to improve gameplay flow:
  - Initial Mana: 15
  - Mana Regeneration Rate (per turn): 5
  - Max Mana: 20
- **Skip Turn**: Players can now use the `skip` command to pass their turn and receive a 1.5x mana regeneration bonus for that turn.
- **Critical Hits**: Troops now have a chance to deal critical damage.
- **EXP & Leveling**: Players gain EXP for actions like destroying towers and winning games, allowing them to level up.

## Network Protocol

TCR uses a simple network protocol based on JSON messages with length-prefixed framing.
For details on the protocol and message formats, see `doc/ApplicationPDUDescription.md`.

## Current Focus & Coming Soon

Focus is on refining existing systems and moving towards **Enhanced TCR** features:
- Further balancing of troop stats, mana costs, and game pacing.
- Potential for more diverse troop special abilities.
- Exploring real-time gameplay mechanics (core of Phase 4).
- UI/UX improvements for the text-based interface.