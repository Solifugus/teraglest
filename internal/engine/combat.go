package engine

import (
	"fmt"
	"math"
	"time"
)

// CombatSystem handles all combat-related calculations and mechanics
type CombatSystem struct {
	world *World
}

// NewCombatSystem creates a new combat system instance
func NewCombatSystem(world *World) *CombatSystem {
	return &CombatSystem{
		world: world,
	}
}

// CombatResult represents the result of a combat calculation
type CombatResult struct {
	Damage        int     // Final damage dealt
	BaseDamage    int     // Base damage before modifiers
	Multiplier    float64 // Attack vs armor type multiplier
	AttackType    string  // Type of attack used
	ArmorType     string  // Type of armor defending
	WasKilled     bool    // Whether target was killed
	CanAttack     bool    // Whether attack can proceed (range, etc.)
	ErrorMessage  string  // Error if attack cannot proceed
}

// CalculateDamage calculates damage dealt when an attacker hits a target
func (cs *CombatSystem) CalculateDamage(attacker, target *GameUnit) CombatResult {
	result := CombatResult{
		CanAttack: false,
	}

	// Basic validation
	if attacker == nil || target == nil {
		result.ErrorMessage = "attacker or target is nil"
		return result
	}

	if !attacker.IsAlive() {
		result.ErrorMessage = "attacker is dead"
		return result
	}

	if !target.IsAlive() {
		result.ErrorMessage = "target is dead"
		return result
	}

	// Check if attacker can attack (has attack capability)
	if attacker.AttackDamage <= 0 {
		result.ErrorMessage = "attacker has no attack damage"
		return result
	}

	// Range checking
	distance := cs.world.CalculateDistance(attacker.Position, target.Position)
	if distance > float64(attacker.AttackRange) {
		result.ErrorMessage = fmt.Sprintf("target out of range: %.1f > %.1f", distance, attacker.AttackRange)
		return result
	}

	// Get attack and armor types from unit definitions
	attackType := cs.getAttackType(attacker)
	armorType := cs.getArmorType(target)

	result.AttackType = attackType
	result.ArmorType = armorType
	result.BaseDamage = attacker.AttackDamage

	// Get damage multiplier from tech tree
	multiplier := cs.world.techTree.GetDamageMultiplier(attackType, armorType)
	result.Multiplier = multiplier

	// Calculate final damage
	finalDamage := float64(attacker.AttackDamage) * multiplier

	// Apply target armor reduction
	armoredDamage := finalDamage - float64(target.Armor)
	if armoredDamage < 1 {
		armoredDamage = 1 // Minimum 1 damage
	}

	result.Damage = int(math.Round(armoredDamage))
	result.CanAttack = true

	// Check if attack would kill target
	result.WasKilled = target.Health <= result.Damage

	return result
}

// ApplyDamage applies damage to a target unit and handles death
func (cs *CombatSystem) ApplyDamage(target *GameUnit, damage int) bool {
	if target == nil || !target.IsAlive() {
		return false
	}

	target.mutex.Lock()
	defer target.mutex.Unlock()

	// Apply damage
	target.Health -= damage
	if target.Health <= 0 {
		target.Health = 0
		target.State = UnitStateDead

		// Handle unit death
		cs.handleUnitDeath(target)
		return true // Unit was killed
	}

	return false // Unit survived
}

// CanAttack checks if an attacker can attack a target (range, line of sight, etc.)
func (cs *CombatSystem) CanAttack(attacker, target *GameUnit) (bool, string) {
	if attacker == nil || target == nil {
		return false, "attacker or target is nil"
	}

	if !attacker.IsAlive() {
		return false, "attacker is dead"
	}

	if !target.IsAlive() {
		return false, "target is dead"
	}

	// Same player check
	if attacker.PlayerID == target.PlayerID {
		return false, "cannot attack same player units"
	}

	// Enhanced range check with attack type consideration
	if !cs.isInAttackRange(attacker, target) {
		distance := cs.world.CalculateDistance(attacker.Position, target.Position)
		return false, fmt.Sprintf("target out of range: %.1f > %.1f", distance, attacker.AttackRange)
	}

	// Attack cooldown check
	if !cs.canAttackNow(attacker) {
		return false, "attack on cooldown"
	}

	// Basic line of sight check (simplified)
	if !cs.hasLineOfSight(attacker, target) {
		return false, "no line of sight to target"
	}

	return true, ""
}

// ExecuteAttack performs a complete attack sequence
func (cs *CombatSystem) ExecuteAttack(attacker, target *GameUnit) CombatResult {
	// Calculate damage
	result := cs.CalculateDamage(attacker, target)

	if !result.CanAttack {
		return result
	}

	// Apply damage to target
	killed := cs.ApplyDamage(target, result.Damage)
	result.WasKilled = killed

	// Update attacker's last attack time
	attacker.LastAttack = time.Now()

	// Create combat event for logging/statistics
	cs.logCombatEvent(attacker, target, result)

	return result
}

// getAttackType gets the attack type for a unit from its definition
func (cs *CombatSystem) getAttackType(unit *GameUnit) string {
	// Default attack type if not found in unit definition
	defaultAttackType := "normal"

	if unit.UnitDef == nil || unit.UnitDef.Unit.Skills == nil {
		return defaultAttackType
	}

	// Look for attack skills with attack type
	for _, skill := range unit.UnitDef.Unit.Skills {
		if skill.AttackType != nil {
			return skill.AttackType.Value
		}
	}

	return defaultAttackType
}

// getArmorType gets the armor type for a unit from its definition
func (cs *CombatSystem) getArmorType(unit *GameUnit) string {
	// Default armor type if not found
	defaultArmorType := "organic"

	if unit.UnitDef == nil {
		return defaultArmorType
	}

	return unit.UnitDef.Unit.Parameters.ArmorType.Value
}

// canAttackNow checks if a unit can attack based on attack speed/cooldown
func (cs *CombatSystem) canAttackNow(unit *GameUnit) bool {
	if unit.AttackSpeed <= 0 {
		return true // No cooldown restriction
	}

	// Calculate time since last attack
	timeSinceLastAttack := time.Since(unit.LastAttack)

	// Convert attack speed to cooldown duration
	// AttackSpeed is attacks per second, so cooldown = 1/AttackSpeed seconds
	cooldownDuration := time.Duration(float64(time.Second) / float64(unit.AttackSpeed))

	return timeSinceLastAttack >= cooldownDuration
}

// hasLineOfSight performs line of sight checking using Bresenham's line algorithm
func (cs *CombatSystem) hasLineOfSight(attacker, target *GameUnit) bool {
	attackerGrid := attacker.GetGridPosition()
	targetGrid := target.GetGridPosition()

	// Use Bresenham's line algorithm to check each grid cell along the line
	return cs.checkLineOfSight(attackerGrid.Grid, targetGrid.Grid)
}

// checkLineOfSight uses Bresenham's line algorithm to check for obstacles
func (cs *CombatSystem) checkLineOfSight(from, to Vector2i) bool {
	x0, y0 := from.X, from.Y
	x1, y1 := to.X, to.Y

	dx := absInt(x1 - x0)
	dy := absInt(y1 - y0)

	x, y := x0, y0

	n := 1 + dx + dy
	x_inc := 1
	if x1 < x0 {
		x_inc = -1
	}
	y_inc := 1
	if y1 < y0 {
		y_inc = -1
	}

	error := dx - dy
	dx *= 2
	dy *= 2

	for i := 0; i < n; i++ {
		// Check if current position blocks line of sight
		pos := Vector2i{X: x, Y: y}

		// Skip the starting and ending positions
		if pos != from && pos != to {
			if !cs.world.IsPositionWalkable(pos) {
				return false // Blocked by obstacle
			}

			// Also check for tall buildings/units that might block sight
			if cs.hasObstacleAtPosition(pos) {
				return false
			}
		}

		if error > 0 {
			x += x_inc
			error -= dy
		} else {
			y += y_inc
			error += dx
		}
	}

	return true // Line of sight is clear
}

// hasObstacleAtPosition checks if there's a sight-blocking obstacle at a position
func (cs *CombatSystem) hasObstacleAtPosition(pos Vector2i) bool {
	// Check for buildings that might block line of sight
	buildings := cs.world.ObjectManager.GetBuildingsForPlayer(-1) // All players
	for _, building := range buildings {
		if building.IsBuilt && building.Health > 0 {
			buildingGrid := WorldToGrid(building.Position, cs.world.GetTileSize())
			if buildingGrid.Grid.X == pos.X && buildingGrid.Grid.Y == pos.Y {
				// Large buildings might block sight
				return cs.buildingBlocksSight(building)
			}
		}
	}

	// Could also check for terrain features, trees, etc.
	return false
}

// buildingBlocksSight determines if a building blocks line of sight
func (cs *CombatSystem) buildingBlocksSight(building *GameBuilding) bool {
	// Most buildings block sight, but some exceptions could exist
	// For example, walls might block sight while towers might not
	switch building.BuildingType {
	case "tower", "watchtower":
		return false // Towers don't block sight
	default:
		return true // Most buildings block sight
	}
}

// isInAttackRange checks if target is within attack range with attack type considerations
func (cs *CombatSystem) isInAttackRange(attacker, target *GameUnit) bool {
	distance := cs.world.CalculateDistance(attacker.Position, target.Position)

	// Basic range check
	if distance > float64(attacker.AttackRange) {
		return false
	}

	// For melee attacks, require adjacency (within 1.5 tiles)
	attackType := cs.getAttackType(attacker)
	if cs.isMeleeAttack(attackType) {
		// Melee attacks need to be very close
		return distance <= 1.5 * float64(cs.world.GetTileSize())
	}

	// Ranged attacks use full range
	return true
}

// isMeleeAttack determines if an attack type is melee
func (cs *CombatSystem) isMeleeAttack(attackType string) bool {
	meleeTypes := map[string]bool{
		"slash":     true,
		"pierce":    true,
		"impact":    true,
		"blade":     true,
		"sword":     true,
		"axe":       true,
		"hammer":    true,
		"claw":      true,
		"bite":      true,
		"normal":    true, // Default melee
	}

	return meleeTypes[attackType]
}

// GetOptimalAttackPosition finds the best position for a unit to attack a target
func (cs *CombatSystem) GetOptimalAttackPosition(attacker, target *GameUnit) (Vector3, bool) {
	attackType := cs.getAttackType(attacker)

	if cs.isMeleeAttack(attackType) {
		return cs.findMeleeAttackPosition(attacker, target)
	}

	return cs.findRangedAttackPosition(attacker, target)
}

// findMeleeAttackPosition finds the best adjacent position for melee attack
func (cs *CombatSystem) findMeleeAttackPosition(attacker, target *GameUnit) (Vector3, bool) {
	targetGrid := target.GetGridPosition()

	// Check all adjacent positions
	neighbors := GetCardinalNeighbors(targetGrid.Grid)

	bestPos := Vector3{}
	bestDistance := float64(999999)
	found := false

	for _, neighbor := range neighbors {
		if cs.world.IsPositionWalkable(neighbor) {
			worldPos := GridToWorld(GridPosition{
				Grid:   neighbor,
				Offset: Vector2{X: 0.5, Y: 0.5},
			}, cs.world.GetTileSize())

			// Calculate distance from attacker's current position
			distance := cs.world.CalculateDistance(attacker.Position, worldPos)

			if distance < bestDistance {
				bestDistance = distance
				bestPos = worldPos
				found = true
			}
		}
	}

	return bestPos, found
}

// findRangedAttackPosition finds a good position for ranged attack with line of sight
func (cs *CombatSystem) findRangedAttackPosition(attacker, target *GameUnit) (Vector3, bool) {
	// For ranged attacks, current position is usually fine if in range and has line of sight
	if cs.isInAttackRange(attacker, target) && cs.hasLineOfSight(attacker, target) {
		return attacker.Position, true
	}

	// TODO: Could implement more sophisticated positioning for ranged units
	// - Find elevated positions
	// - Avoid clustering with other ranged units
	// - Consider cover and retreat paths

	return attacker.Position, false
}

// absInt returns the absolute value of an integer
func absInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// handleUnitDeath handles comprehensive cleanup when a unit dies
func (cs *CombatSystem) handleUnitDeath(unit *GameUnit) {
	// Clear current command and queue
	unit.CurrentCommand = nil
	unit.CommandQueue = []UnitCommand{}

	// Clear all targets and references
	unit.Target = nil
	unit.AttackTarget = nil
	unit.GatherTarget = nil
	unit.BuildTarget = nil

	// Clear carried resources (drop them)
	cs.handleResourceDrop(unit)

	// Cancel any buildings this unit was constructing
	cs.handleConstructionCancellation(unit)

	// Free up the grid position
	gridPos := unit.GetGridPosition()
	cs.world.SetOccupied(gridPos.Grid, false)

	// Update player statistics
	player := cs.world.GetPlayer(unit.PlayerID)
	if player != nil {
		player.UnitsLost++
	}

	// Clear any commands targeting this dead unit from other units
	cs.clearTargetingCommands(unit)

	// Notify ObjectManager for proper cleanup
	cs.world.ObjectManager.RemoveUnit(unit.ID)

	// Create death event
	cs.createDeathEvent(unit)
}

// handleResourceDrop handles dropping carried resources when a unit dies
func (cs *CombatSystem) handleResourceDrop(unit *GameUnit) {
	if len(unit.CarriedResources) == 0 {
		return
	}

	// Add carried resources back to player pool (simplified)
	// In a more complex implementation, resources might be dropped on the ground
	player := cs.world.GetPlayer(unit.PlayerID)
	if player != nil {
		for resourceType, amount := range unit.CarriedResources {
			player.Resources[resourceType] += amount
		}
	}

	// Clear carried resources
	unit.CarriedResources = make(map[string]int)
}

// handleConstructionCancellation cancels any building construction this unit was doing
func (cs *CombatSystem) handleConstructionCancellation(unit *GameUnit) {
	if unit.BuildTarget == nil {
		return
	}

	building := unit.BuildTarget

	// If building is not yet complete, mark it for cancellation
	if !building.IsBuilt {
		building.Health = 0 // Mark as destroyed

		// Free up the building's grid position
		buildingGrid := WorldToGrid(building.Position, cs.world.GetTileSize())
		cs.world.SetOccupied(buildingGrid.Grid, false)
		cs.world.SetWalkable(buildingGrid.Grid, true)

		// Remove building from ObjectManager
		cs.world.ObjectManager.RemoveBuilding(building.ID)
	}

	unit.BuildTarget = nil
}

// clearTargetingCommands removes commands from other units that target this dead unit
func (cs *CombatSystem) clearTargetingCommands(deadUnit *GameUnit) {
	// Check all players' units for commands targeting the dead unit
	allPlayers := cs.world.GetAllPlayers()
	for _, player := range allPlayers {
		units := cs.world.ObjectManager.GetUnitsForPlayer(player.ID)
		for _, unit := range units {
			if unit.ID == deadUnit.ID {
				continue // Skip the dead unit itself
			}

			// Clear current command if it targets the dead unit
			if unit.CurrentCommand != nil && cs.commandTargetsUnit(unit.CurrentCommand, deadUnit) {
				unit.CurrentCommand = nil
				unit.State = UnitStateIdle
				unit.AttackTarget = nil
			}

			// Remove commands from queue that target the dead unit
			newQueue := []UnitCommand{}
			for _, cmd := range unit.CommandQueue {
				if !cs.commandTargetsUnit(&cmd, deadUnit) {
					newQueue = append(newQueue, cmd)
				}
			}
			unit.CommandQueue = newQueue
		}
	}
}

// commandTargetsUnit checks if a command targets a specific unit
func (cs *CombatSystem) commandTargetsUnit(command *UnitCommand, targetUnit *GameUnit) bool {
	if command == nil || targetUnit == nil {
		return false
	}

	// Check direct unit target
	if command.TargetUnit != nil && command.TargetUnit.ID == targetUnit.ID {
		return true
	}

	// Could also check for other targeting mechanisms
	return false
}

// createDeathEvent creates a death event for logging and statistics
func (cs *CombatSystem) createDeathEvent(unit *GameUnit) {
	deathEvent := UnitDeathEvent{
		UnitID:    unit.ID,
		PlayerID:  unit.PlayerID,
		UnitType:  unit.UnitType,
		Position:  unit.Position,
		Timestamp: time.Now(),
	}

	// Send to event system
	cs.sendDeathEvent(deathEvent)
}

// sendDeathEvent sends death events to the game's event system
func (cs *CombatSystem) sendDeathEvent(event UnitDeathEvent) {
	// TODO: Connect to actual game event system
	// fmt.Printf("[DEATH EVENT] Unit %d (%s) died at position (%.1f, %.1f)\n",
	//     event.UnitID, event.UnitType, event.Position.X, event.Position.Z)
}

// RegenerateHealth handles passive health regeneration for units
func (cs *CombatSystem) RegenerateHealth(unit *GameUnit, deltaTime time.Duration) {
	if unit == nil || !unit.IsAlive() {
		return
	}

	// Only regenerate if not at full health
	if unit.Health >= unit.MaxHealth {
		return
	}

	// Get regeneration rate from unit definition or use default
	regenRate := cs.getHealthRegenRate(unit)
	if regenRate <= 0 {
		return
	}

	// Calculate health to regenerate this frame
	regenAmount := int(float64(regenRate) * deltaTime.Seconds())
	if regenAmount < 1 && deltaTime >= time.Second {
		regenAmount = 1 // Minimum 1 health per second if there's any regen
	}

	if regenAmount > 0 {
		unit.mutex.Lock()
		unit.Health += regenAmount
		if unit.Health > unit.MaxHealth {
			unit.Health = unit.MaxHealth
		}
		unit.mutex.Unlock()
	}
}

// getHealthRegenRate gets the health regeneration rate for a unit
func (cs *CombatSystem) getHealthRegenRate(unit *GameUnit) float64 {
	// Default regeneration rate (health per second)
	defaultRegen := 0.1 // Very slow regeneration

	// Could be enhanced to read from unit definition
	// For now, return default for all units
	return defaultRegen
}

// UnitDeathEvent represents a unit death for event logging
type UnitDeathEvent struct {
	UnitID    int       // ID of the unit that died
	PlayerID  int       // Player who owned the unit
	UnitType  string    // Type of unit that died
	Position  Vector3   // Where the unit died
	Timestamp time.Time // When the unit died
}

// logCombatEvent logs combat events for statistics and debugging
func (cs *CombatSystem) logCombatEvent(attacker, target *GameUnit, result CombatResult) {
	// Create combat event
	combatEvent := CombatEvent{
		AttackerID:   attacker.ID,
		TargetID:     target.ID,
		AttackerPlayerID: attacker.PlayerID,
		TargetPlayerID:   target.PlayerID,
		Damage:       result.Damage,
		AttackType:   result.AttackType,
		ArmorType:    result.ArmorType,
		Multiplier:   result.Multiplier,
		WasKilled:    result.WasKilled,
		Timestamp:    time.Now(),
	}

	// Send to game event system
	cs.sendCombatEvent(combatEvent)
}

// sendCombatEvent sends combat events to the game's event system
func (cs *CombatSystem) sendCombatEvent(event CombatEvent) {
	// TODO: This would be connected to the actual game instance's event system
	// For now, we'll just log it (could be expanded to send to game.eventQueue)

	// Simple logging for development - in production this would send to game.eventQueue
	// fmt.Printf("[COMBAT EVENT] Unit %d attacked Unit %d for %d damage\n",
	//     event.AttackerID, event.TargetID, event.Damage)
}

// CombatEvent represents a combat action for event logging
type CombatEvent struct {
	AttackerID       int       // ID of attacking unit
	TargetID         int       // ID of target unit
	AttackerPlayerID int       // Player ID of attacker
	TargetPlayerID   int       // Player ID of target
	Damage           int       // Damage dealt
	AttackType       string    // Type of attack
	ArmorType        string    // Type of armor
	Multiplier       float64   // Damage multiplier applied
	WasKilled        bool      // Whether target was killed
	Timestamp        time.Time // When combat occurred
}

// GetCombatStats returns combat statistics for a player
func (cs *CombatSystem) GetCombatStats(playerID int) CombatStats {
	// This would track various combat statistics
	// For now, return basic stats from player data
	player := cs.world.GetPlayer(playerID)
	if player == nil {
		return CombatStats{}
	}

	return CombatStats{
		PlayerID:     playerID,
		UnitsLost:    player.UnitsLost,
		UnitsCreated: player.UnitsCreated,
		// TODO: Add more detailed combat statistics
	}
}

// CombatStats represents combat statistics for a player
type CombatStats struct {
	PlayerID        int // Player ID
	UnitsLost       int // Total units lost
	UnitsCreated    int // Total units created
	DamageDealt     int // Total damage dealt
	DamageReceived  int // Total damage received
	UnitsKilled     int // Enemy units killed
}