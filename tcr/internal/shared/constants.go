package shared

// Game constants
const (
	// Base requirements
	BaseEXPForLevelUp    = 100
	LevelUpEXPMultiplier = 1.1 // 10% increase per level

	// Enhanced TCR constants
	InitialMana         = 15
	ManaRegenRate       = 5
	GameDurationSeconds = 180 // 3 minutes
	MaxMana             = 20  // Maximum mana a player can hold

	// Combat constants
	CritDamageMultiplier   = 1.2 // 20% bonus damage on critical hit
	DefaultTroopCritChance = 20  // 20% chance for troops in Enhanced mode

	// EXP rewards for match results
	WinEXPReward  = 30
	DrawEXPReward = 10

	// Special ability constants
	QueenHealAmount = 300
)

// Tower types
const (
	KingTowerType   = "KING"
	GuardTower1Type = "GUARD1"
	GuardTower2Type = "GUARD2"
)

// Special ability identifiers
const (
	HealLowestHPTowerAbility = "HEAL_LOWEST_HP_TOWER_300"
)
