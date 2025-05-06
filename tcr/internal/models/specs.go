package models

// TroopSpec defines the specifications for a troop type
type TroopSpec struct {
	// Name is the unique identifier for the troop
	Name string `json:"Name"`

	// BaseHP is the base hit points of the troop (at level 1)
	BaseHP int `json:"BaseHP"`

	// BaseATK is the base attack value of the troop (at level 1)
	BaseATK int `json:"BaseATK"`

	// BaseDEF is the base defense value of the troop (at level 1)
	BaseDEF int `json:"BaseDEF"`

	// ManaCost is the mana required to deploy this troop (used in Enhanced TCR)
	ManaCost int `json:"ManaCost"`

	// DestroyEXP is the experience points reward for destroying this troop
	DestroyEXP int `json:"DestroyEXP"`

	// SpecialAbility defines any special abilities the troop has (e.g., "HEAL_LOWEST_HP_TOWER_300" for Queen)
	SpecialAbility string `json:"SpecialAbility"`

	// IsSpecialOnly indicates if the troop only performs special abilities and doesn't engage in normal combat
	// For example, Queen is marked as true since she only heals and doesn't attack/defend
	IsSpecialOnly bool `json:"IsSpecialOnly"`
}

// TowerSpec defines the specifications for a tower type
type TowerSpec struct {
	// Name is the display name of the tower (e.g., "King Tower", "Guard Tower")
	Name string `json:"Name"`

	// Type is the identifier for the tower type (e.g., "KING", "GUARD1", "GUARD2")
	Type string `json:"Type"`

	// BaseHP is the base hit points of the tower (at level 1)
	BaseHP int `json:"BaseHP"`

	// BaseATK is the base attack value of the tower (at level 1)
	BaseATK int `json:"BaseATK"`

	// BaseDEF is the base defense value of the tower (at level 1)
	BaseDEF int `json:"BaseDEF"`

	// CritChancePercent is the chance of critical hit for this tower
	// Note: Initially unused for tower's own attacks as per assumption
	CritChancePercent int `json:"CritChancePercent"`

	// DestroyEXP is the experience points reward for destroying this tower
	DestroyEXP int `json:"DestroyEXP"`
}
