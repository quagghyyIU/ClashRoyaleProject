# Configuration Files Guide

This document describes the structure and purpose of the configuration files used in TCR.

## troops.json

The `troops.json` file contains an array of troop specifications with the following structure:

```json
[
  {
    "Name": "TroopName",        // Unique identifier for the troop
    "BaseHP": 100,              // Base hit points at level 1
    "BaseATK": 100,             // Base attack value at level 1
    "BaseDEF": 50,              // Base defense value at level 1
    "ManaCost": 3,              // Mana cost to deploy (Enhanced TCR)
    "DestroyEXP": 10,           // EXP reward for destroying this troop
    "SpecialAbility": "",       // Special ability identifier (if any)
    "IsSpecialOnly": false      // Whether troop only performs special abilities
  }
]
```

### Special Cases

#### Queen Troop
The Queen is a special troop that only performs a healing ability and is consumed after use:

```json
{
  "Name": "Queen",
  "BaseHP": 0,                  // Queen doesn't engage in combat
  "BaseATK": 0,                 // Queen doesn't attack
  "BaseDEF": 0,                 // Queen has no defense
  "ManaCost": 5,                // Relatively high mana cost
  "DestroyEXP": 30,             // High EXP reward
  "SpecialAbility": "HEAL_LOWEST_HP_TOWER_300",  // Heals tower by 300 HP
  "IsSpecialOnly": true         // Queen only performs healing
}
```

## towers.json

The `towers.json` file contains an array of tower specifications with the following structure:

```json
[
  {
    "Name": "Tower Name",       // Display name of the tower
    "Type": "TOWER_TYPE",       // Identifier (KING, GUARD1, GUARD2)
    "BaseHP": 1000,             // Base hit points at level 1
    "BaseATK": 100,             // Base attack value at level 1
    "BaseDEF": 50,              // Base defense value at level 1
    "CritChancePercent": 5,     // Critical hit chance (%)
    "DestroyEXP": 100           // EXP reward for destroying this tower
  }
]
```

### Notes on Tower Types

- **King Tower** (Type: "KING"): The main tower. If destroyed, the game ends.
- **Guard Tower 1** (Type: "GUARD1"): Must be destroyed before attacking Guard Tower 2 or King Tower.
- **Guard Tower 2** (Type: "GUARD2"): Can only be attacked after Guard Tower 1 is destroyed.