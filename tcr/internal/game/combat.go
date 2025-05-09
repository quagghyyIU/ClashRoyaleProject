package game

// CalculateDamage calculates damage dealt by an attacker to a defender
// Formula: DMG = ATK_A - DEF_B (if ≥ 0)
func CalculateDamage(attackerEffectiveATK int, defenderEffectiveDEF int) int {
	damage := attackerEffectiveATK - defenderEffectiveDEF
	if damage < 0 {
		return 0
	}
	return damage
}
