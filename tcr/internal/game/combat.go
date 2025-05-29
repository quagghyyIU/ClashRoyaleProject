package game

import (
	"math/rand"
	"tcr/internal/shared"
	"time"
)

// CalculateDamage calculates damage dealt by an attacker to a defender
// Formula: DMG = ATK_A - DEF_B (if ≥ 0)
func CalculateDamage(attackerEffectiveATK int, defenderEffectiveDEF int) int {
	damage := attackerEffectiveATK - defenderEffectiveDEF
	if damage < 0 {
		return 0
	}
	return damage
}

// CalculateDamageEnhanced calculates damage dealt by an attacker to a defender,
// incorporating critical hit logic based on the attacker's crit chance.
// Formula: DMG = (ATK_A or ATK_A * CritDamageMultiplier if CRIT) - DEF_B (if ≥ 0)
// It returns the calculated damage and a boolean indicating if a critical hit occurred.
func CalculateDamageEnhanced(attackerEffectiveATK int, defenderEffectiveDEF int, attackerCritChancePercent float64) (damage int, didCrit bool) {
	// Seed random number generator only once if not already done elsewhere globally
	// For simplicity in this function, we can seed it. In a larger app, seed once at startup.
	// Consider moving seed to main or init if not already there.
	rand.Seed(time.Now().UnixNano()) // Note: For true randomness and performance, seed once globally.

	rawAttack := float64(attackerEffectiveATK)
	didCrit = false

	if attackerCritChancePercent > 0 && rand.Float64()*100 < attackerCritChancePercent {
		didCrit = true
		rawAttack *= shared.CritDamageMultiplier
	}

	calculatedDamage := int(rawAttack) - defenderEffectiveDEF
	if calculatedDamage < 0 {
		return 0, didCrit
	}
	return calculatedDamage, didCrit
}
