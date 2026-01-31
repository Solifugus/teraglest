package engine

import (
	"math"
	"time"
)

// AdvancedCombatSystem extends the basic combat system with AOE, formations, and complex damage types
type AdvancedCombatSystem struct {
	*CombatSystem // Embed basic combat system
	world         *World
}

// NewAdvancedCombatSystem creates a new advanced combat system
func NewAdvancedCombatSystem(world *World) *AdvancedCombatSystem {
	return &AdvancedCombatSystem{
		CombatSystem: NewCombatSystem(world),
		world:        world,
	}
}

// AdvancedDamageType represents complex damage types with special properties
type AdvancedDamageType struct {
	Name           string             `json:"name"`
	Category       DamageCategory     `json:"category"`
	Penetration    float64            `json:"penetration"`     // Armor penetration (0.0-1.0)
	SplashRadius   float64            `json:"splash_radius"`   // AOE radius in tiles
	SplashFalloff  SplashFalloffType  `json:"splash_falloff"`  // How splash damage decreases
	DamageOverTime DOTEffect          `json:"damage_over_time"` // DOT effect
	SpecialEffects []CombatEffect     `json:"special_effects"` // Additional effects
}

// DamageCategory represents broad categories of damage
type DamageCategory int

const (
	DamagePhysical DamageCategory = iota // Physical damage (pierce, slash, crush)
	DamageMagical                        // Magical damage (fire, ice, lightning)
	DamageRanged                         // Ranged physical damage (arrows, bullets)
	DamageSiege                          // Anti-building damage (catapults)
	DamageTrue                           // True damage (bypasses armor)
)

// SplashFalloffType defines how splash damage decreases with distance
type SplashFalloffType int

const (
	SplashLinear     SplashFalloffType = iota // Linear decrease
	SplashQuadratic                           // Quadratic decrease
	SplashConstant                            // Constant damage in radius
	SplashStep                                // Step function (full/half/none)
)

// DOTEffect represents damage over time effects
type DOTEffect struct {
	Enabled       bool          `json:"enabled"`
	DamagePerTick int           `json:"damage_per_tick"`
	TickInterval  time.Duration `json:"tick_interval"`
	Duration      time.Duration `json:"duration"`
	EffectType    string        `json:"effect_type"` // "poison", "burn", "bleed", etc.
}

// CombatEffect represents special combat effects
type CombatEffect struct {
	Type       EffectType    `json:"type"`
	Duration   time.Duration `json:"duration"`
	Magnitude  float64       `json:"magnitude"`
	Parameters map[string]interface{} `json:"parameters"`
}

// EffectType represents types of combat effects
type EffectType int

const (
	EffectStun      EffectType = iota // Unit cannot act
	EffectSlow                        // Reduced movement speed
	EffectFear                        // Unit runs away
	EffectPoison                      // Damage over time
	EffectBurn                        // Fire damage over time
	EffectFreeze                      // Cannot move or attack
	EffectBlind                       // Reduced attack accuracy
	EffectWeaken                      // Reduced damage output
	EffectArmor                       // Increased damage resistance
	EffectRage                        // Increased damage output
)

// FormationCombatBonus represents combat bonuses from formations
type FormationCombatBonus struct {
	DamageBonus      float64 `json:"damage_bonus"`      // Damage multiplier bonus
	AccuracyBonus    float64 `json:"accuracy_bonus"`    // Hit chance bonus
	DefenseBonus     float64 `json:"defense_bonus"`     // Damage reduction bonus
	RangeBonus       float64 `json:"range_bonus"`       // Attack range bonus
	MoraleBonus      float64 `json:"morale_bonus"`      // Resistance to fear effects
	FlankeingBonus   float64 `json:"flankeing_bonus"`   // Bonus when attacking from sides/rear
}

// SplashDamageResult contains results from splash damage calculation
type SplashDamageResult struct {
	PrimaryTarget    *GameUnit       `json:"primary_target"`
	PrimaryDamage    int             `json:"primary_damage"`
	SplashTargets    []SplashVictim  `json:"splash_targets"`
	TotalTargets     int             `json:"total_targets"`
	FormationBonus   float64         `json:"formation_bonus"`
}

// SplashVictim represents a unit hit by splash damage
type SplashVictim struct {
	Unit     *GameUnit `json:"unit"`
	Damage   int       `json:"damage"`
	Distance float64   `json:"distance"`
}

// Predefined advanced damage types
var AdvancedDamageTypes = map[string]AdvancedDamageType{
	"sword": {
		Name:          "sword",
		Category:      DamagePhysical,
		Penetration:   0.1,
		SplashRadius:  0.0,
		SplashFalloff: SplashConstant,
	},
	"arrow": {
		Name:          "arrow",
		Category:      DamageRanged,
		Penetration:   0.2,
		SplashRadius:  0.0,
		SplashFalloff: SplashConstant,
	},
	"catapult": {
		Name:          "catapult",
		Category:      DamageSiege,
		Penetration:   0.8,
		SplashRadius:  2.5,
		SplashFalloff: SplashQuadratic,
	},
	"fireball": {
		Name:          "fireball",
		Category:      DamageMagical,
		Penetration:   0.3,
		SplashRadius:  1.5,
		SplashFalloff: SplashLinear,
		DamageOverTime: DOTEffect{
			Enabled:       true,
			DamagePerTick: 5,
			TickInterval:  time.Second,
			Duration:      time.Second * 5,
			EffectType:    "burn",
		},
	},
	"lightning": {
		Name:          "lightning",
		Category:      DamageMagical,
		Penetration:   0.9,
		SplashRadius:  0.8,
		SplashFalloff: SplashStep,
		SpecialEffects: []CombatEffect{
			{
				Type:      EffectStun,
				Duration:  time.Millisecond * 500,
				Magnitude: 1.0,
			},
		},
	},
	"explosion": {
		Name:          "explosion",
		Category:      DamageSiege,
		Penetration:   0.6,
		SplashRadius:  3.0,
		SplashFalloff: SplashQuadratic,
		SpecialEffects: []CombatEffect{
			{
				Type:      EffectStun,
				Duration:  time.Second * 2,
				Magnitude: 0.8,
			},
		},
	},
}

// Formation combat bonuses by formation type
var FormationCombatBonuses = map[FormationType]FormationCombatBonus{
	FormationLine: {
		DamageBonus:    0.1,  // 10% damage bonus
		AccuracyBonus:  0.05, // 5% accuracy bonus
		DefenseBonus:   0.0,
		RangeBonus:     0.1,  // 10% range bonus
		MoraleBonus:    0.15,
		FlankeingBonus: 0.0,
	},
	FormationWedge: {
		DamageBonus:    0.2,  // 20% damage bonus (aggressive formation)
		AccuracyBonus:  0.1,  // 10% accuracy bonus
		DefenseBonus:   -0.1, // -10% defense (exposed flanks)
		RangeBonus:     0.0,
		MoraleBonus:    0.2,  // High morale in attack formation
		FlankeingBonus: 0.3,  // 30% bonus when flanking
	},
	FormationBox: {
		DamageBonus:    -0.05, // -5% damage (defensive stance)
		AccuracyBonus:  0.0,
		DefenseBonus:   0.25,  // 25% defense bonus
		RangeBonus:     -0.1,  // -10% range (cramped)
		MoraleBonus:    0.1,
		FlankeingBonus: 0.0,
	},
	FormationCircle: {
		DamageBonus:    0.0,
		AccuracyBonus:  0.0,
		DefenseBonus:   0.2,   // 20% defense bonus
		RangeBonus:     0.0,
		MoraleBonus:    0.25,  // High morale (surrounded formation)
		FlankeingBonus: -0.2,  // Cannot be flanked easily
	},
	FormationColumn: {
		DamageBonus:    -0.1, // -10% damage (marching formation)
		AccuracyBonus:  -0.05,
		DefenseBonus:   -0.1, // -10% defense (exposed)
		RangeBonus:     0.0,
		MoraleBonus:    0.05,
		FlankeingBonus: 0.0,
	},
	FormationScatter: {
		DamageBonus:    0.0,
		AccuracyBonus:  -0.1, // -10% accuracy (spread out)
		DefenseBonus:   0.3,  // 30% defense against AOE
		RangeBonus:     0.05,
		MoraleBonus:    -0.1, // Lower morale when spread
		FlankeingBonus: 0.1,
	},
}

// ExecuteAdvancedAttack performs an attack with advanced damage calculation
func (acs *AdvancedCombatSystem) ExecuteAdvancedAttack(attacker, target *GameUnit) SplashDamageResult {
	result := SplashDamageResult{
		PrimaryTarget: target,
		SplashTargets: make([]SplashVictim, 0),
	}

	// Get advanced damage type
	attackType := acs.getAttackType(attacker)
	advancedDamage, exists := AdvancedDamageTypes[attackType]
	if !exists {
		// Fall back to basic combat for unknown types
		basicResult := acs.ExecuteAttack(attacker, target)
		result.PrimaryDamage = basicResult.Damage
		result.TotalTargets = 1
		return result
	}

	// Calculate formation bonus
	formationBonus := acs.calculateFormationBonus(attacker, target)
	result.FormationBonus = formationBonus

	// Calculate primary damage with formation bonus
	baseDamage := acs.calculateAdvancedDamage(attacker, target, advancedDamage, formationBonus)
	result.PrimaryDamage = baseDamage

	// Apply primary damage
	acs.ApplyDamage(target, baseDamage)

	// Apply special effects to primary target
	acs.applySpecialEffects(target, advancedDamage.SpecialEffects)

	// Apply damage over time effects
	if advancedDamage.DamageOverTime.Enabled {
		acs.applyDamageOverTime(target, advancedDamage.DamageOverTime)
	}

	// Handle splash damage if applicable
	if advancedDamage.SplashRadius > 0 {
		splashTargets := acs.calculateSplashDamage(attacker, target, advancedDamage, formationBonus)
		result.SplashTargets = splashTargets
		result.TotalTargets = 1 + len(splashTargets)
	} else {
		result.TotalTargets = 1
	}

	// Update attacker's last attack time
	attacker.LastAttack = time.Now()

	// Log advanced combat event
	acs.logAdvancedCombatEvent(attacker, result, advancedDamage)

	return result
}

// calculateAdvancedDamage computes damage with advanced damage type modifiers
func (acs *AdvancedCombatSystem) calculateAdvancedDamage(attacker, target *GameUnit, damageType AdvancedDamageType, formationBonus float64) int {
	// Start with basic damage calculation
	basicResult := acs.CalculateDamage(attacker, target)
	if !basicResult.CanAttack {
		return 0
	}

	baseDamage := float64(basicResult.BaseDamage)

	// Apply formation bonus
	if formationBonus != 0 {
		baseDamage *= (1.0 + formationBonus)
	}

	// Apply damage type multipliers
	finalDamage := baseDamage * basicResult.Multiplier

	// Apply armor penetration
	if damageType.Penetration > 0 {
		effectiveArmor := float64(target.Armor) * (1.0 - damageType.Penetration)
		finalDamage = finalDamage - effectiveArmor
	} else {
		finalDamage = finalDamage - float64(target.Armor)
	}

	// Apply category-specific bonuses/penalties
	finalDamage = acs.applyCategoryModifiers(finalDamage, damageType.Category, target)

	// Minimum damage
	if finalDamage < 1 {
		finalDamage = 1
	}

	return int(math.Round(finalDamage))
}

// applyCategoryModifiers applies damage category-specific modifiers
func (acs *AdvancedCombatSystem) applyCategoryModifiers(damage float64, category DamageCategory, target *GameUnit) float64 {
	// Apply category-specific bonuses based on target type
	switch category {
	case DamageSiege:
		// Siege weapons do extra damage to buildings
		if acs.isBuilding(target) {
			damage *= 2.0
		}
	case DamageMagical:
		// Magic damage might ignore some armor types or be more effective against certain units
		armorType := acs.getArmorType(target)
		if armorType == "light" {
			damage *= 1.2 // Magic is effective against light armor
		}
	case DamageTrue:
		// True damage bypasses all armor (already handled by penetration = 1.0)
		break
	}

	return damage
}

// calculateFormationBonus calculates combat bonuses from formations
func (acs *AdvancedCombatSystem) calculateFormationBonus(attacker, target *GameUnit) float64 {
	bonus := 0.0

	// Check if attacker is in a formation
	if acs.world.groupMgr != nil {
		attackerGroup, exists := acs.world.groupMgr.GetUnitGroup(attacker.ID)
		if exists && attackerGroup.IsFormed {
			formationBonus := FormationCombatBonuses[attackerGroup.Formation]

			// Base damage bonus from formation
			bonus += formationBonus.DamageBonus

			// Check for flanking bonus
			if acs.isFlanking(attacker, target, attackerGroup) {
				bonus += formationBonus.FlankeingBonus
			}

			// Check if there are nearby friendly units for coordination bonus
			nearbyAllies := acs.countNearbyAllies(attacker, 3.0) // Within 3 tiles
			if nearbyAllies >= 2 {
				bonus += 0.05 * float64(nearbyAllies-1) // 5% per additional nearby ally
			}
		}
	}

	return bonus
}

// calculateSplashDamage calculates and applies splash damage to nearby units
func (acs *AdvancedCombatSystem) calculateSplashDamage(attacker, primaryTarget *GameUnit, damageType AdvancedDamageType, formationBonus float64) []SplashVictim {
	victims := make([]SplashVictim, 0)

	// Find all units within splash radius
	nearbyUnits := acs.findUnitsInRadius(primaryTarget.Position, damageType.SplashRadius)

	for _, unit := range nearbyUnits {
		// Skip primary target (already handled)
		if unit.ID == primaryTarget.ID {
			continue
		}

		// Skip friendly units (no friendly fire for now)
		if unit.PlayerID == attacker.PlayerID {
			continue
		}

		// Calculate distance from splash center
		distance := acs.world.CalculateDistance(primaryTarget.Position, unit.Position)

		// Calculate splash damage based on falloff type
		splashDamage := acs.calculateSplashDamageAmount(attacker, unit, damageType, distance, formationBonus)

		if splashDamage > 0 {
			// Apply splash damage
			acs.ApplyDamage(unit, splashDamage)

			// Apply reduced special effects
			reducedEffects := acs.reduceEffectsForSplash(damageType.SpecialEffects, distance, damageType.SplashRadius)
			acs.applySpecialEffects(unit, reducedEffects)

			victims = append(victims, SplashVictim{
				Unit:     unit,
				Damage:   splashDamage,
				Distance: distance,
			})
		}
	}

	return victims
}

// calculateSplashDamageAmount calculates splash damage amount based on distance and falloff
func (acs *AdvancedCombatSystem) calculateSplashDamageAmount(attacker, target *GameUnit, damageType AdvancedDamageType, distance, formationBonus float64) int {
	// Calculate base damage for this target
	baseDamage := acs.calculateAdvancedDamage(attacker, target, damageType, formationBonus)

	// Apply splash falloff
	multiplier := acs.calculateSplashMultiplier(distance, damageType.SplashRadius, damageType.SplashFalloff)

	splashDamage := float64(baseDamage) * multiplier

	// Minimum splash damage
	if splashDamage < 1 && multiplier > 0 {
		splashDamage = 1
	}

	return int(math.Round(splashDamage))
}

// calculateSplashMultiplier calculates damage multiplier based on distance and falloff type
func (acs *AdvancedCombatSystem) calculateSplashMultiplier(distance, radius float64, falloff SplashFalloffType) float64 {
	if distance > radius {
		return 0.0
	}

	ratio := distance / radius

	switch falloff {
	case SplashConstant:
		return 1.0
	case SplashLinear:
		return 1.0 - ratio
	case SplashQuadratic:
		return 1.0 - (ratio * ratio)
	case SplashStep:
		if ratio <= 0.33 {
			return 1.0
		} else if ratio <= 0.66 {
			return 0.5
		} else {
			return 0.25
		}
	default:
		return 1.0 - ratio
	}
}

// findUnitsInRadius finds all units within a given radius of a position
func (acs *AdvancedCombatSystem) findUnitsInRadius(center Vector3, radius float64) []*GameUnit {
	allUnits := make([]*GameUnit, 0)

	// Get all players and their units
	players := acs.world.GetAllPlayers()
	for _, player := range players {
		units := acs.world.ObjectManager.GetUnitsForPlayer(player.ID)
		for _, unit := range units {
			if unit.IsAlive() {
				distance := acs.world.CalculateDistance(center, unit.Position)
				if distance <= radius {
					allUnits = append(allUnits, unit)
				}
			}
		}
	}

	return allUnits
}

// applySpecialEffects applies special combat effects to a unit
func (acs *AdvancedCombatSystem) applySpecialEffects(target *GameUnit, effects []CombatEffect) {
	for _, effect := range effects {
		acs.applyEffect(target, effect)
	}
}

// applyEffect applies a single effect to a unit
func (acs *AdvancedCombatSystem) applyEffect(target *GameUnit, effect CombatEffect) {
	// For now, implement basic effects
	// A full implementation would have an effect manager system
	switch effect.Type {
	case EffectStun:
		// Temporarily disable unit actions
		target.State = UnitStateIdle // Simplified - would need proper status effect system
	case EffectSlow:
		// Reduce movement speed (would need temporary effect system)
		// target.Speed *= (1.0 - float32(effect.Magnitude))
	case EffectPoison:
		// Apply poison DOT effect
		acs.applyDamageOverTime(target, DOTEffect{
			Enabled:       true,
			DamagePerTick: int(effect.Magnitude),
			TickInterval:  time.Second,
			Duration:      effect.Duration,
			EffectType:    "poison",
		})
	}
}

// applyDamageOverTime applies a damage over time effect
func (acs *AdvancedCombatSystem) applyDamageOverTime(target *GameUnit, dot DOTEffect) {
	if !dot.Enabled {
		return
	}

	// TODO: Implement proper DOT system with effect tracking
	// For now, this is a placeholder that would integrate with a status effect system
	// The DOT would be tracked and processed during world updates
}

// reduceEffectsForSplash reduces effect intensity for splash damage
func (acs *AdvancedCombatSystem) reduceEffectsForSplash(effects []CombatEffect, distance, radius float64) []CombatEffect {
	reduced := make([]CombatEffect, len(effects))

	multiplier := acs.calculateSplashMultiplier(distance, radius, SplashLinear)

	for i, effect := range effects {
		reduced[i] = effect
		reduced[i].Magnitude *= multiplier
		reduced[i].Duration = time.Duration(float64(effect.Duration) * multiplier)
	}

	return reduced
}

// isFlanking checks if the attacker is flanking the target
func (acs *AdvancedCombatSystem) isFlanking(attacker, target *GameUnit, attackerGroup *UnitGroup) bool {
	// Check if there are friendly units on the opposite side of the target
	targetPos := target.Position
	attackerPos := attacker.Position

	// Calculate angle from target to attacker
	attackAngle := math.Atan2(attackerPos.Z-targetPos.Z, attackerPos.X-targetPos.X)

	// Look for friendly units on the opposite side (180° ± 45°)
	oppositeAngleMin := attackAngle + math.Pi - math.Pi/4
	oppositeAngleMax := attackAngle + math.Pi + math.Pi/4

	nearbyAllies := acs.findUnitsInRadius(targetPos, 4.0) // Check within 4 tiles
	for _, ally := range nearbyAllies {
		if ally.PlayerID == attacker.PlayerID && ally.ID != attacker.ID && ally.IsAlive() {
			allyAngle := math.Atan2(ally.Position.Z-targetPos.Z, ally.Position.X-targetPos.X)

			// Normalize angles
			if allyAngle < oppositeAngleMin && oppositeAngleMin > math.Pi {
				allyAngle += 2 * math.Pi
			}

			if allyAngle >= oppositeAngleMin && allyAngle <= oppositeAngleMax {
				return true // Flanking detected
			}
		}
	}

	return false
}

// countNearbyAllies counts friendly units within a given radius
func (acs *AdvancedCombatSystem) countNearbyAllies(unit *GameUnit, radius float64) int {
	count := 0
	nearbyUnits := acs.findUnitsInRadius(unit.Position, radius)

	for _, nearby := range nearbyUnits {
		if nearby.PlayerID == unit.PlayerID && nearby.ID != unit.ID && nearby.IsAlive() {
			count++
		}
	}

	return count
}

// isBuilding checks if a unit is actually a building (placeholder)
func (acs *AdvancedCombatSystem) isBuilding(unit *GameUnit) bool {
	// This would check unit type or other properties to determine if it's a building
	// For now, simple check based on unit type name
	buildingTypes := map[string]bool{
		"castle":      true,
		"barracks":    true,
		"farm":        true,
		"tower":       true,
		"wall":        true,
		"town_center": true,
	}

	return buildingTypes[unit.UnitType]
}

// logAdvancedCombatEvent logs advanced combat events
func (acs *AdvancedCombatSystem) logAdvancedCombatEvent(attacker *GameUnit, result SplashDamageResult, damageType AdvancedDamageType) {
	// Create advanced combat event
	event := AdvancedCombatEvent{
		AttackerID:     attacker.ID,
		PrimaryTargetID: result.PrimaryTarget.ID,
		AttackType:     damageType.Name,
		PrimaryDamage:  result.PrimaryDamage,
		SplashTargets:  len(result.SplashTargets),
		TotalDamage:    result.PrimaryDamage + acs.sumSplashDamage(result.SplashTargets),
		FormationBonus: result.FormationBonus,
		Timestamp:      time.Now(),
	}

	// Send to event system (would be expanded in full implementation)
	acs.sendAdvancedCombatEvent(event)
}

// sumSplashDamage calculates total splash damage dealt
func (acs *AdvancedCombatSystem) sumSplashDamage(victims []SplashVictim) int {
	total := 0
	for _, victim := range victims {
		total += victim.Damage
	}
	return total
}

// sendAdvancedCombatEvent sends advanced combat events to the game's event system
func (acs *AdvancedCombatSystem) sendAdvancedCombatEvent(event AdvancedCombatEvent) {
	// TODO: Connect to actual game event system
	// For development: could log or send to game.eventQueue
}

// AdvancedCombatEvent represents an advanced combat event
type AdvancedCombatEvent struct {
	AttackerID      int       `json:"attacker_id"`
	PrimaryTargetID int       `json:"primary_target_id"`
	AttackType      string    `json:"attack_type"`
	PrimaryDamage   int       `json:"primary_damage"`
	SplashTargets   int       `json:"splash_targets"`
	TotalDamage     int       `json:"total_damage"`
	FormationBonus  float64   `json:"formation_bonus"`
	Timestamp       time.Time `json:"timestamp"`
}