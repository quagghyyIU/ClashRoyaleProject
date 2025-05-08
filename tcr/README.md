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

Currently in Phase 1 of development, which includes:
- Project structure setup
- Core data structures and configuration files
- Basic loading utilities
- Complete game logic implementation for Simple TCR
- Offline testing functionality

## Simple TCR Game Rules

In Simple TCR mode:
1. Players take turns deploying troops to attack opponent towers
2. The Guard Tower 1 must be destroyed before Guard Tower 2 or King Tower can be targeted
3. When a troop destroys a tower, the player gets an immediate second attack
4. The Queen troop can be deployed to heal the friendly tower with the lowest HP percentage
5. The game ends when a player's King Tower is destroyed

## How to Run (Phase 1)

To test the Simple TCR game logic in offline mode:

1. Build and run the server:
   ```bash
   cd tcr
   go build ./cmd/server
   ./server
   ```

2. Follow the on-screen prompts to play the game:
   - The game starts with Player A's turn
   - Use the `d <troop_name> <tower_number>` command to attack
   - Tower numbers: 1=Guard1, 2=Guard2, 3=King
   - Examples: `d Pawn 1` to attack Guard Tower 1, `d Knight 3` to attack King Tower (if valid)
   - To deploy the Queen and use her healing ability, use: `d Queen 1` (the target number doesn't matter)

## Current Limitations in Phase 1

- **Troop Exhaustion**: Troops are consumed after deployment and not replenished. Players start with 3 regular troops plus special troops (Queen). If a player runs out of troops, they won't be able to attack. Troop replenishment will be addressed in future phases.
- **Simple Command Interface**: The current implementation uses a basic console interface. A more sophisticated UI will be implemented in later phases.

## Coming Soon

The next phases will implement:
- Network communication between client and server
- Enhanced TCR with real-time gameplay
- Mana system
- Critical hits
- Player progression with experience and leveling
- Troop replenishment mechanism