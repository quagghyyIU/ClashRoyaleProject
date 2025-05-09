package game

import (
	"strconv"
	"tcr/internal/models"
	"tcr/internal/shared"
)

// IsValidTarget checks if the target tower is a valid target based on the game rules
// For Simple TCR, G1 (GuardTower1) must be destroyed before G2 or King can be targeted
func IsValidTarget(attackingPlayer *Player, targetTowerID string, gameState *GameState) bool {
	// Get opponent (tower owner)
	opponentPlayer := gameState.GetOpponentPlayer()

	// Can't target your own towers
	if targetTowerID == attackingPlayer.KingTower.ID ||
		targetTowerID == attackingPlayer.GuardTower1.ID ||
		targetTowerID == attackingPlayer.GuardTower2.ID {
		return false
	}

	// First check if the target tower exists and is not already destroyed
	var targetTower *TowerInstance
	if targetTowerID == opponentPlayer.KingTower.ID {
		targetTower = opponentPlayer.KingTower
	} else if targetTowerID == opponentPlayer.GuardTower1.ID {
		targetTower = opponentPlayer.GuardTower1
	} else if targetTowerID == opponentPlayer.GuardTower2.ID {
		targetTower = opponentPlayer.GuardTower2
	} else {
		// Target tower ID doesn't match any known tower
		return false
	}

	if targetTower.Destroyed {
		// Can't target already destroyed towers
		return false
	}

	// Simple TCR Rule: GuardTower1 must be destroyed before targeting GuardTower2 or King
	if targetTowerID == opponentPlayer.GuardTower2.ID || targetTowerID == opponentPlayer.KingTower.ID {
		// Check if GuardTower1 is destroyed
		if !opponentPlayer.GuardTower1.Destroyed {
			return false
		}
	}

	return true
}

// CanDeployTroop checks if a player can deploy a specific troop
func CanDeployTroop(player *Player, troopName string, gameState *GameState) bool {
	// Check if it's the player's turn (for Simple TCR)
	if gameState.CurrentTurn != player.Username {
		return false
	}

	// Check if the troop is in the player's hand
	for _, troop := range player.Troops {
		if troop.Spec.Name == troopName {
			return true
		}
	}

	return false
}

// FindTroopInHand finds a troop in the player's hand by name
// Returns the troop and its index, or nil and -1 if not found
func FindTroopInHand(player *Player, troopName string) (*TroopInstance, int) {
	for i, troop := range player.Troops {
		if troop.Spec.Name == troopName {
			return troop, i
		}
	}
	return nil, -1
}

// ApplySpecialAbility handles special abilities of troops
// For now, this only implements Queen's heal ability
func ApplySpecialAbility(actingPlayer *Player, troopSpec *models.TroopSpec) string {
	if troopSpec.SpecialAbility == shared.HealLowestHPTowerAbility {
		return applyQueenHeal(actingPlayer)
	}
	return "No special ability applied."
}

// applyQueenHeal implements the Queen's heal ability
// It finds the friendly tower with the lowest HP percentage and heals it
func applyQueenHeal(player *Player) string {
	// Find the tower with the lowest HP percentage
	var lowestHPTower *TowerInstance
	lowestHPPercentage := 1.0 // Start with 100%

	// Check KingTower
	kingHPPerc := float64(player.KingTower.CurrentHP) / float64(player.KingTower.Spec.BaseHP)
	if kingHPPerc < lowestHPPercentage {
		lowestHPTower = player.KingTower
		lowestHPPercentage = kingHPPerc
	}

	// Check GuardTower1 if not destroyed
	if !player.GuardTower1.Destroyed {
		g1HPPerc := float64(player.GuardTower1.CurrentHP) / float64(player.GuardTower1.Spec.BaseHP)
		if g1HPPerc < lowestHPPercentage {
			lowestHPTower = player.GuardTower1
			lowestHPPercentage = g1HPPerc
		}
	}

	// Check GuardTower2 if not destroyed
	if !player.GuardTower2.Destroyed {
		g2HPPerc := float64(player.GuardTower2.CurrentHP) / float64(player.GuardTower2.Spec.BaseHP)
		if g2HPPerc < lowestHPPercentage {
			lowestHPTower = player.GuardTower2
			lowestHPPercentage = g2HPPerc
		}
	}

	// Apply healing
	if lowestHPTower != nil {
		oldHP := lowestHPTower.CurrentHP
		originalMaxHP := lowestHPTower.Spec.BaseHP

		// Calculate how much to heal (not exceeding max HP)
		healAmount := shared.QueenHealAmount
		if oldHP+healAmount > originalMaxHP {
			healAmount = originalMaxHP - oldHP
		}

		lowestHPTower.CurrentHP += healAmount

		return "Queen healed " + lowestHPTower.ID + " for " +
			strconv.Itoa(healAmount) + " HP (" + strconv.Itoa(oldHP) + " → " +
			strconv.Itoa(lowestHPTower.CurrentHP) + ")"
	}

	return "Queen couldn't find a tower to heal."
}
