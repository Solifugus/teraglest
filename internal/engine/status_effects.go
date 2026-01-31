package engine

import (
	"sync"
	"time"
)

// StatusEffectManager manages temporary status effects on units
type StatusEffectManager struct {
	unitEffects map[int][]ActiveStatusEffect // Unit ID to active effects
	mutex       sync.RWMutex                   // Thread safety
	world       *World                        // World reference for unit lookup
}

// NewStatusEffectManager creates a new status effect manager
func NewStatusEffectManager() *StatusEffectManager {
	return &StatusEffectManager{
		unitEffects: make(map[int][]ActiveStatusEffect),
	}
}

// SetWorld sets the world reference for unit lookup
func (sem *StatusEffectManager) SetWorld(world *World) {
	sem.world = world
}

// ActiveStatusEffect represents an active effect on a unit
type ActiveStatusEffect struct {
	Effect      StatusEffect  `json:"effect"`
	StartTime   time.Time     `json:"start_time"`
	LastTick    time.Time     `json:"last_tick"`
	TicksLeft   int          `json:"ticks_left"`
	Source      *GameUnit    `json:"source"`      // Unit that caused this effect
	StackCount  int          `json:"stack_count"` // For stackable effects
}

// StatusEffect represents a temporary effect that can be applied to units
type StatusEffect struct {
	ID           string        `json:"id"`           // Unique effect identifier
	Name         string        `json:"name"`         // Display name
	Type         EffectType    `json:"type"`         // Type of effect
	Duration     time.Duration `json:"duration"`     // Total duration
	TickInterval time.Duration `json:"tick_interval"` // How often effect ticks
	Magnitude    float64       `json:"magnitude"`    // Effect strength
	IsStackable  bool         `json:"is_stackable"` // Can multiple instances exist
	MaxStacks    int          `json:"max_stacks"`   // Maximum stack count
	IsBuff       bool         `json:"is_buff"`      // Beneficial vs harmful
	IsDispellable bool        `json:"is_dispellable"` // Can be removed by dispel
	IconPath     string       `json:"icon_path"`    // UI icon path
}

// Predefined status effects
var StatusEffects = map[string]StatusEffect{
	"poison": {
		ID:           "poison",
		Name:         "Poison",
		Type:         EffectPoison,
		Duration:     time.Second * 10,
		TickInterval: time.Second * 2,
		Magnitude:    5.0, // 5 damage per tick
		IsStackable:  true,
		MaxStacks:    3,
		IsBuff:       false,
		IsDispellable: true,
		IconPath:     "effects/poison.png",
	},
	"burn": {
		ID:           "burn",
		Name:         "Burning",
		Type:         EffectBurn,
		Duration:     time.Second * 8,
		TickInterval: time.Second * 1,
		Magnitude:    3.0, // 3 damage per tick
		IsStackable:  false,
		MaxStacks:    1,
		IsBuff:       false,
		IsDispellable: true,
		IconPath:     "effects/burn.png",
	},
	"stun": {
		ID:           "stun",
		Name:         "Stunned",
		Type:         EffectStun,
		Duration:     time.Second * 3,
		TickInterval: time.Second, // Check every second
		Magnitude:    1.0,
		IsStackable:  false,
		MaxStacks:    1,
		IsBuff:       false,
		IsDispellable: true,
		IconPath:     "effects/stun.png",
	},
	"slow": {
		ID:           "slow",
		Name:         "Slowed",
		Type:         EffectSlow,
		Duration:     time.Second * 6,
		TickInterval: time.Second,
		Magnitude:    0.3, // 30% speed reduction
		IsStackable:  false,
		MaxStacks:    1,
		IsBuff:       false,
		IsDispellable: true,
		IconPath:     "effects/slow.png",
	},
	"rage": {
		ID:           "rage",
		Name:         "Rage",
		Type:         EffectRage,
		Duration:     time.Second * 15,
		TickInterval: time.Second,
		Magnitude:    0.25, // 25% damage bonus
		IsStackable:  true,
		MaxStacks:    2,
		IsBuff:       true,
		IsDispellable: false, // Cannot be dispelled
		IconPath:     "effects/rage.png",
	},
	"armor_buff": {
		ID:           "armor_buff",
		Name:         "Iron Skin",
		Type:         EffectArmor,
		Duration:     time.Second * 20,
		TickInterval: time.Second,
		Magnitude:    5.0, // +5 armor
		IsStackable:  false,
		MaxStacks:    1,
		IsBuff:       true,
		IsDispellable: true,
		IconPath:     "effects/armor.png",
	},
	"fear": {
		ID:           "fear",
		Name:         "Feared",
		Type:         EffectFear,
		Duration:     time.Second * 4,
		TickInterval: time.Second,
		Magnitude:    1.0,
		IsStackable:  false,
		MaxStacks:    1,
		IsBuff:       false,
		IsDispellable: true,
		IconPath:     "effects/fear.png",
	},
}

// ApplyStatusEffect applies a status effect to a unit
func (sem *StatusEffectManager) ApplyStatusEffect(unit *GameUnit, effectID string, source *GameUnit) bool {
	sem.mutex.Lock()
	defer sem.mutex.Unlock()

	effect, exists := StatusEffects[effectID]
	if !exists {
		return false
	}

	// Check if effect can be applied
	if !sem.canApplyEffect(unit, effect) {
		return false
	}

	// Handle stacking
	if effect.IsStackable {
		sem.addStackableEffect(unit, effect, source)
	} else {
		sem.addOrRefreshEffect(unit, effect, source)
	}

	// Apply immediate effect if any
	sem.applyImmediateEffect(unit, effect)

	return true
}

// canApplyEffect checks if an effect can be applied to a unit
func (sem *StatusEffectManager) canApplyEffect(unit *GameUnit, effect StatusEffect) bool {
	// Dead units can't receive effects
	if !unit.IsAlive() {
		return false
	}

	// Check immunity (could be enhanced with unit properties)
	if sem.isImmuneToEffect(unit, effect) {
		return false
	}

	// Check if stackable effect is at max stacks
	if effect.IsStackable {
		currentStacks := sem.getEffectStackCount(unit.ID, effect.ID)
		return currentStacks < effect.MaxStacks
	}

	return true
}

// addStackableEffect adds a new stack of a stackable effect
func (sem *StatusEffectManager) addStackableEffect(unit *GameUnit, effect StatusEffect, source *GameUnit) {
	unitEffects := sem.unitEffects[unit.ID]

	// Find existing effect to stack
	for i, activeEffect := range unitEffects {
		if activeEffect.Effect.ID == effect.ID {
			// Increase stack count
			sem.unitEffects[unit.ID][i].StackCount++
			sem.unitEffects[unit.ID][i].StartTime = time.Now() // Refresh duration
			return
		}
	}

	// Add new effect
	newEffect := ActiveStatusEffect{
		Effect:     effect,
		StartTime:  time.Now(),
		LastTick:   time.Now(),
		TicksLeft:  int(effect.Duration / effect.TickInterval),
		Source:     source,
		StackCount: 1,
	}

	sem.unitEffects[unit.ID] = append(unitEffects, newEffect)
}

// addOrRefreshEffect adds a new effect or refreshes existing non-stackable effect
func (sem *StatusEffectManager) addOrRefreshEffect(unit *GameUnit, effect StatusEffect, source *GameUnit) {
	unitEffects := sem.unitEffects[unit.ID]

	// Find existing effect to refresh
	for i, activeEffect := range unitEffects {
		if activeEffect.Effect.ID == effect.ID {
			// Refresh existing effect
			sem.unitEffects[unit.ID][i].StartTime = time.Now()
			sem.unitEffects[unit.ID][i].LastTick = time.Now()
			sem.unitEffects[unit.ID][i].TicksLeft = int(effect.Duration / effect.TickInterval)
			return
		}
	}

	// Add new effect
	newEffect := ActiveStatusEffect{
		Effect:     effect,
		StartTime:  time.Now(),
		LastTick:   time.Now(),
		TicksLeft:  int(effect.Duration / effect.TickInterval),
		Source:     source,
		StackCount: 1,
	}

	sem.unitEffects[unit.ID] = append(unitEffects, newEffect)
}

// Update processes all active status effects
func (sem *StatusEffectManager) Update(deltaTime time.Duration) {
	sem.mutex.Lock()
	defer sem.mutex.Unlock()

	for unitID, effects := range sem.unitEffects {
		remainingEffects := make([]ActiveStatusEffect, 0, len(effects))

		for _, effect := range effects {
			// Check if effect should tick
			if time.Since(effect.LastTick) >= effect.Effect.TickInterval {
				// Get the unit
				unit := sem.getUnitByID(unitID)
				if unit != nil && unit.IsAlive() {
					// Process effect tick
					sem.processEffectTick(unit, &effect)

					effect.LastTick = time.Now()
					effect.TicksLeft--
				}
			}

			// Check if effect is still active
			if effect.TicksLeft > 0 && time.Since(effect.StartTime) < effect.Effect.Duration {
				remainingEffects = append(remainingEffects, effect)
			} else {
				// Effect expired, remove its influence
				unit := sem.getUnitByID(unitID)
				if unit != nil {
					sem.removeEffectInfluence(unit, effect)
				}
			}
		}

		if len(remainingEffects) > 0 {
			sem.unitEffects[unitID] = remainingEffects
		} else {
			delete(sem.unitEffects, unitID)
		}
	}
}

// processEffectTick processes a single tick of an effect
func (sem *StatusEffectManager) processEffectTick(unit *GameUnit, effect *ActiveStatusEffect) {
	switch effect.Effect.Type {
	case EffectPoison, EffectBurn:
		// Apply damage over time
		damage := int(effect.Effect.Magnitude * float64(effect.StackCount))
		if damage > 0 {
			unit.mutex.Lock()
			unit.Health -= damage
			if unit.Health <= 0 {
				unit.Health = 0
				unit.State = UnitStateDead
			}
			unit.mutex.Unlock()
		}

	case EffectStun:
		// Keep unit stunned (cannot act)
		if unit.State != UnitStateDead {
			unit.State = UnitStateIdle // Simplified - would need proper action blocking
			unit.CurrentCommand = nil
		}

	case EffectFear:
		// Make unit run away (simplified implementation)
		if unit.State != UnitStateDead && unit.CurrentCommand == nil {
			// Would implement fear behavior here
			unit.State = UnitStateIdle
		}
	}
}

// applyImmediateEffect applies any immediate effects when the status is first applied
func (sem *StatusEffectManager) applyImmediateEffect(unit *GameUnit, effect StatusEffect) {
	switch effect.Type {
	case EffectStun:
		// Immediately interrupt current action
		unit.CurrentCommand = nil
		unit.State = UnitStateIdle

	case EffectSlow:
		// Immediately reduce speed (would need proper modifier system)
		// For now, this is conceptual as we'd need a modifier system

	case EffectRage, EffectArmor:
		// Buff effects would immediately modify unit stats
		// Implementation would depend on how unit modifiers are handled
	}
}

// removeEffectInfluence removes the influence of an effect when it expires
func (sem *StatusEffectManager) removeEffectInfluence(unit *GameUnit, effect ActiveStatusEffect) {
	switch effect.Effect.Type {
	case EffectSlow:
		// Restore normal speed
		// Would need modifier system to properly implement

	case EffectRage, EffectArmor:
		// Remove stat bonuses
		// Would need modifier system to properly implement
	}
}

// RemoveEffect removes a specific effect from a unit
func (sem *StatusEffectManager) RemoveEffect(unitID int, effectID string) bool {
	sem.mutex.Lock()
	defer sem.mutex.Unlock()

	effects, exists := sem.unitEffects[unitID]
	if !exists {
		return false
	}

	for i, effect := range effects {
		if effect.Effect.ID == effectID {
			// Remove effect influence
			unit := sem.getUnitByID(unitID)
			if unit != nil {
				sem.removeEffectInfluence(unit, effect)
			}

			// Remove from list
			sem.unitEffects[unitID] = append(effects[:i], effects[i+1:]...)

			// Clean up empty lists
			if len(sem.unitEffects[unitID]) == 0 {
				delete(sem.unitEffects, unitID)
			}

			return true
		}
	}

	return false
}

// DispelEffects removes dispellable effects from a unit
func (sem *StatusEffectManager) DispelEffects(unitID int, dispelBeneficial, dispelHarmful bool) int {
	sem.mutex.Lock()
	defer sem.mutex.Unlock()

	effects, exists := sem.unitEffects[unitID]
	if !exists {
		return 0
	}

	remaining := make([]ActiveStatusEffect, 0, len(effects))
	dispelled := 0

	for _, effect := range effects {
		shouldRemove := false

		if effect.Effect.IsDispellable {
			if dispelBeneficial && effect.Effect.IsBuff {
				shouldRemove = true
			} else if dispelHarmful && !effect.Effect.IsBuff {
				shouldRemove = true
			}
		}

		if shouldRemove {
			// Remove effect influence
			unit := sem.getUnitByID(unitID)
			if unit != nil {
				sem.removeEffectInfluence(unit, effect)
			}
			dispelled++
		} else {
			remaining = append(remaining, effect)
		}
	}

	if len(remaining) > 0 {
		sem.unitEffects[unitID] = remaining
	} else {
		delete(sem.unitEffects, unitID)
	}

	return dispelled
}

// GetUnitEffects returns all active effects on a unit
func (sem *StatusEffectManager) GetUnitEffects(unitID int) []ActiveStatusEffect {
	sem.mutex.RLock()
	defer sem.mutex.RUnlock()

	effects, exists := sem.unitEffects[unitID]
	if !exists {
		return []ActiveStatusEffect{}
	}

	// Return a copy to avoid race conditions
	result := make([]ActiveStatusEffect, len(effects))
	copy(result, effects)
	return result
}

// HasEffect checks if a unit has a specific effect
func (sem *StatusEffectManager) HasEffect(unitID int, effectID string) bool {
	sem.mutex.RLock()
	defer sem.mutex.RUnlock()

	effects, exists := sem.unitEffects[unitID]
	if !exists {
		return false
	}

	for _, effect := range effects {
		if effect.Effect.ID == effectID {
			return true
		}
	}

	return false
}

// getEffectStackCount returns the number of stacks for a stackable effect
func (sem *StatusEffectManager) getEffectStackCount(unitID int, effectID string) int {
	effects, exists := sem.unitEffects[unitID]
	if !exists {
		return 0
	}

	for _, effect := range effects {
		if effect.Effect.ID == effectID {
			return effect.StackCount
		}
	}

	return 0
}

// isImmuneToEffect checks if a unit is immune to a specific effect
func (sem *StatusEffectManager) isImmuneToEffect(unit *GameUnit, effect StatusEffect) bool {
	// TODO: Implement immunity system
	// Could check unit type, equipment, other effects, etc.
	// For now, no immunities
	return false
}

// getUnitByID retrieves a unit by ID (helper function)
func (sem *StatusEffectManager) getUnitByID(unitID int) *GameUnit {
	if sem.world == nil || sem.world.ObjectManager == nil {
		return nil
	}
	return sem.world.ObjectManager.GetUnit(unitID)
}

// CleanupDeadUnit removes all effects from a dead unit
func (sem *StatusEffectManager) CleanupDeadUnit(unitID int) {
	sem.mutex.Lock()
	defer sem.mutex.Unlock()

	delete(sem.unitEffects, unitID)
}

// GetEffectStats returns statistics about active effects
func (sem *StatusEffectManager) GetEffectStats() map[string]interface{} {
	sem.mutex.RLock()
	defer sem.mutex.RUnlock()

	stats := map[string]interface{}{
		"total_affected_units": len(sem.unitEffects),
		"total_effects":        0,
		"effects_by_type":      make(map[string]int),
	}

	totalEffects := 0
	effectCounts := make(map[string]int)

	for _, effects := range sem.unitEffects {
		totalEffects += len(effects)
		for _, effect := range effects {
			effectCounts[effect.Effect.ID]++
		}
	}

	stats["total_effects"] = totalEffects
	stats["effects_by_type"] = effectCounts

	return stats
}