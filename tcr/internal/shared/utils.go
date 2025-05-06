package shared

import (
	"math/rand"
	"time"
)

var (
	// Initialize random seed
	rng = rand.New(rand.NewSource(time.Now().UnixNano()))
)

// GetRandomInt returns a random integer between min and max (inclusive)
func GetRandomInt(min, max int) int {
	return rng.Intn(max-min+1) + min
}

// RollForCritical determines if a critical hit occurs based on the given percent chance
func RollForCritical(critChancePercent int) bool {
	return GetRandomInt(1, 100) <= critChancePercent
}

// CalculateStatWithLevelBonus applies level bonus to a base stat
func CalculateStatWithLevelBonus(baseStat, level int) int {
	levelMultiplier := 1.0 + float64(level-1)*0.1 // 10% per level above 1
	return int(float64(baseStat) * levelMultiplier)
}

// CalculateRequiredEXP calculates the EXP required to reach the next level
func CalculateRequiredEXP(currentLevel int) int {
	exp := BaseEXPForLevelUp
	for i := 1; i < currentLevel; i++ {
		exp = int(float64(exp) * LevelUpEXPMultiplier)
	}
	return exp
}
