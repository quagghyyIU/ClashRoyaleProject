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
- **Queen**: Special unit that heals the friendly tower with the lowest HP

## Configuration

Game data is stored in JSON configuration files under the `configs/` directory:
- `troops.json`: Contains troop specifications
- `towers.json`: Contains tower specifications

## Development Status

Currently in Phase 0 of development, which includes:
- Project structure setup
- Core data structures
- Configuration files
- Basic loading utilities