package engine

import (
	"fmt"
	"math"
	"time"

	"teraglest/internal/data"
)

// UnitCommand represents a command that can be given to a unit
type UnitCommand struct {
	Type        CommandType     `json:"type"`
	Target      *Vector3        `json:"target"`           // World coordinates target
	GridTarget  *GridPosition   `json:"grid_target"`      // Grid coordinates target
	TargetUnit  *GameUnit       `json:"target_unit"`
	TargetBuilding *GameBuilding `json:"target_building"`
	TargetResource *ResourceNode `json:"target_resource"`
	Parameters  map[string]interface{} `json:"parameters"`
	CreatedAt   time.Time       `json:"created_at"`
	StartedAt   time.Time       `json:"started_at"`
	Priority    int             `json:"priority"`
	IsQueued    bool            `json:"is_queued"`
}

// CommandType represents different types of commands
type CommandType int

const (
	CommandMove      CommandType = iota // Move to a position
	CommandAttack                       // Attack a target
	CommandGather                       // Gather resources
	CommandBuild                        // Construct a building
	CommandRepair                       // Repair a building/unit
	CommandStop                         // Stop current action
	CommandHold                         // Hold position
	CommandPatrol                       // Patrol between points
	CommandFollow                       // Follow another unit
	CommandGuard                        // Guard a target
	CommandProduce                      // Produce a unit (building command)
	CommandUpgrade                      // Upgrade building/technology
	CommandFormation                    // Formation-related commands
	CommandGroupMove                    // Move entire group
	CommandGroupAttack                  // Group attack command
)

// CommandProcessor handles command processing for units and buildings
type CommandProcessor struct {
	world           *World
	combatSystem    *AdvancedCombatSystem
	statusEffectMgr *StatusEffectManager
	visualSystem    *CombatVisualSystem
}

// NewCommandProcessor creates a new command processor
func NewCommandProcessor(world *World) *CommandProcessor {
	statusMgr := NewStatusEffectManager()
	statusMgr.SetWorld(world) // Set world reference for unit lookup
	combatSys := NewAdvancedCombatSystem(world)
	visualSys := NewCombatVisualSystem(world)

	return &CommandProcessor{
		world:           world,
		combatSystem:    combatSys,
		statusEffectMgr: statusMgr,
		visualSystem:    visualSys,
	}
}

// IssueCommand issues a command to a unit
func (cp *CommandProcessor) IssueCommand(unitID int, command UnitCommand) error {
	unit := cp.world.ObjectManager.GetUnit(unitID)
	if unit == nil {
		return fmt.Errorf("unit %d not found", unitID)
	}

	command.CreatedAt = time.Now()

	// Validate command based on unit capabilities
	if err := cp.validateCommand(unit, command); err != nil {
		return fmt.Errorf("invalid command: %w", err)
	}

	unit.mutex.Lock()
	defer unit.mutex.Unlock()

	// Handle immediate vs queued commands
	if command.IsQueued && unit.CurrentCommand != nil {
		// Add to queue
		unit.CommandQueue = append(unit.CommandQueue, command)
	} else {
		// Replace current command
		unit.CurrentCommand = &command
		unit.CommandQueue = []UnitCommand{} // Clear queue if not queuing
		cp.startCommand(unit, &command)
	}

	return nil
}

// IssueBuildingCommand issues a command to a building
func (cp *CommandProcessor) IssueBuildingCommand(buildingID int, command UnitCommand) error {
	building := cp.world.ObjectManager.GetBuilding(buildingID)
	if building == nil {
		return fmt.Errorf("building %d not found", buildingID)
	}

	command.CreatedAt = time.Now()

	building.mutex.Lock()
	defer building.mutex.Unlock()

	switch command.Type {
	case CommandProduce:
		return cp.startProduction(building, command)
	case CommandUpgrade:
		return cp.startUpgrade(building, command)
	default:
		return fmt.Errorf("unsupported building command: %v", command.Type)
	}
}

// CancelCommand cancels a unit's current command
func (cp *CommandProcessor) CancelCommand(unitID int) error {
	unit := cp.world.ObjectManager.GetUnit(unitID)
	if unit == nil {
		return fmt.Errorf("unit %d not found", unitID)
	}

	unit.mutex.Lock()
	defer unit.mutex.Unlock()

	// Stop current command
	unit.CurrentCommand = nil
	unit.State = UnitStateIdle
	unit.Target = nil
	unit.AttackTarget = nil
	unit.GatherTarget = nil
	unit.BuildTarget = nil

	return nil
}

// ClearCommandQueue clears a unit's command queue
func (cp *CommandProcessor) ClearCommandQueue(unitID int) error {
	unit := cp.world.ObjectManager.GetUnit(unitID)
	if unit == nil {
		return fmt.Errorf("unit %d not found", unitID)
	}

	unit.mutex.Lock()
	defer unit.mutex.Unlock()

	unit.CommandQueue = []UnitCommand{}
	return nil
}

// ProcessCommand processes the current command for a unit (called from unit.Update)
func (cp *CommandProcessor) ProcessCommand(unit *GameUnit, command *UnitCommand, deltaTime time.Duration) {
	if command == nil {
		return
	}

	// Mark as started if not already
	if command.StartedAt.IsZero() {
		command.StartedAt = time.Now()
		cp.startCommand(unit, command)
	}

	// Process command based on type
	switch command.Type {
	case CommandMove:
		cp.processMoveCommand(unit, command, deltaTime)
	case CommandAttack:
		cp.processAttackCommand(unit, command, deltaTime)
	case CommandGather:
		cp.processGatherCommand(unit, command, deltaTime)
	case CommandBuild:
		cp.processBuildCommand(unit, command, deltaTime)
	case CommandRepair:
		cp.processRepairCommand(unit, command, deltaTime)
	case CommandStop:
		cp.processStopCommand(unit, command)
	case CommandHold:
		cp.processHoldCommand(unit, command)
	case CommandPatrol:
		cp.processPatrolCommand(unit, command, deltaTime)
	case CommandFollow:
		cp.processFollowCommand(unit, command, deltaTime)
	case CommandGuard:
		cp.processGuardCommand(unit, command, deltaTime)
	}
}

// Update processes all active unit commands and building production commands
func (cp *CommandProcessor) Update(deltaTime time.Duration) {
	// Process all active unit commands for all players
	allPlayers := cp.world.GetAllPlayers()
	for _, player := range allPlayers {
		// Get units for this player
		playerUnits := cp.world.ObjectManager.GetUnitsForPlayer(player.ID)
		for _, unit := range playerUnits {
			// Process health regeneration for living units
			if unit.IsAlive() {
				cp.combatSystem.RegenerateHealth(unit, deltaTime)
			}

			// Process active commands
			if unit.CurrentCommand != nil {
				cp.ProcessCommand(unit, unit.CurrentCommand, deltaTime)
			}
			// Process command queue progression
			unit.processCommandQueue()
		}

		// Process building production/upgrade commands for this player
		playerBuildings := cp.world.ObjectManager.GetBuildingsForPlayer(player.ID)
		for _, building := range playerBuildings {
			cp.processBuildingCommands(building, deltaTime)
		}
	}
}

// processBuildingCommands handles building production and upgrade processing
func (cp *CommandProcessor) processBuildingCommands(building *GameBuilding, deltaTime time.Duration) {
	building.mutex.Lock()
	defer building.mutex.Unlock()

	// Process ongoing production
	if building.CurrentProduction != nil {
		cp.processUnitProduction(building, deltaTime)
	}

	// Process ongoing upgrades
	if building.CurrentUpgrade != nil {
		cp.processUpgradeProgress(building, deltaTime)
	}
}

// processUnitProduction handles unit production progress for buildings
func (cp *CommandProcessor) processUnitProduction(building *GameBuilding, deltaTime time.Duration) {
	if building.CurrentProduction == nil {
		return
	}

	production := building.CurrentProduction
	production.Progress += float32(deltaTime.Seconds()) / float32(production.Duration.Seconds())

	if production.Progress >= 1.0 {
		// Production complete - spawn unit
		spawnPos := cp.findUnitSpawnPosition(building)
		_, err := cp.world.ObjectManager.CreateUnit(
			building.PlayerID,
			production.ItemName,
			spawnPos,
			nil) // faction data will be looked up

		if err == nil {
			// Success - clear production
			building.CurrentProduction = nil

			// Process next item in queue
			if len(building.ProductionQueue) > 0 {
				building.CurrentProduction = &building.ProductionQueue[0]
				building.ProductionQueue = building.ProductionQueue[1:]
			}
		}
		// On failure, production continues (retry next frame)
	}
}

// processUpgradeProgress handles building upgrade progress
func (cp *CommandProcessor) processUpgradeProgress(building *GameBuilding, deltaTime time.Duration) {
	if building.CurrentUpgrade == nil {
		return
	}

	upgrade := building.CurrentUpgrade
	upgrade.Progress += float32(deltaTime.Seconds()) / float32(upgrade.Duration.Seconds())

	if upgrade.Progress >= 1.0 {
		// Upgrade complete
		building.UpgradeLevel++
		building.CurrentUpgrade = nil

		// Apply upgrade effects (enhanced resource generation, etc.)
		cp.applyUpgradeEffects(building, upgrade.UpgradeType)
	}
}

// findUnitSpawnPosition finds a suitable position near a building to spawn a unit
func (cp *CommandProcessor) findUnitSpawnPosition(building *GameBuilding) Vector3 {
	// Start from building position and search for free space nearby
	basePos := building.Position

	// Try positions in a spiral pattern around the building
	for radius := 2.0; radius <= 10.0; radius += 1.0 {
		for angle := 0.0; angle < 360.0; angle += 45.0 {
			// Convert angle to radians
			angleRad := angle * math.Pi / 180.0

			testPos := Vector3{
				X: basePos.X + radius*math.Cos(angleRad),
				Y: basePos.Y,
				Z: basePos.Z + radius*math.Sin(angleRad),
			}

			// Convert to grid position for walkability check
			gridPos := Vector2i{
				X: int(testPos.X / float64(cp.world.GetTileSize())),
				Y: int(testPos.Z / float64(cp.world.GetTileSize())),
			}

			// Check if position is walkable and not occupied
			if cp.world.IsPositionWalkable(gridPos) {
				return testPos
			}
		}
	}

	// Fallback to building position if no free space found
	return basePos
}

// applyUpgradeEffects applies the effects of a completed upgrade to a building
func (cp *CommandProcessor) applyUpgradeEffects(building *GameBuilding, upgradeType string) {
	// Apply upgrade-specific effects
	switch upgradeType {
	case "resource_efficiency":
		// Increase resource generation by 25%
		for resType, rate := range building.ResourceGeneration {
			building.ResourceGeneration[resType] = rate * 1.25
		}
	case "production_speed":
		// Increase production rate
		building.ProductionRate *= 1.3
	case "durability":
		// Increase max health
		building.MaxHealth = int(float32(building.MaxHealth) * 1.2)
		building.Health = building.MaxHealth // Heal to full
	}
}

// Validation methods

func (cp *CommandProcessor) validateCommand(unit *GameUnit, command UnitCommand) error {
	if !unit.IsAlive() {
		return fmt.Errorf("unit is dead")
	}

	switch command.Type {
	case CommandMove:
		if command.Target == nil {
			return fmt.Errorf("move command requires target position")
		}
	case CommandAttack:
		if command.TargetUnit == nil {
			return fmt.Errorf("attack command requires target unit")
		}
		if !command.TargetUnit.IsAlive() {
			return fmt.Errorf("cannot attack dead unit")
		}
	case CommandGather:
		if command.TargetResource == nil {
			return fmt.Errorf("gather command requires target resource")
		}
		if command.TargetResource.Amount <= 0 {
			return fmt.Errorf("resource node is depleted")
		}
	case CommandBuild:
		if command.Target == nil {
			return fmt.Errorf("build command requires target position")
		}
		// Resource validation for building construction
		if buildingType, ok := command.Parameters["building_type"].(string); ok {
			cost := cp.getBuildingCost(buildingType, unit.PlayerID)
			if cost != nil {
				validator := NewResourceValidator(cp.world)
				result := validator.ValidateResources(ResourceCheck{
					PlayerID: unit.PlayerID,
					Required: cost,
					Purpose:  fmt.Sprintf("building construction (%s)", buildingType),
				})
				if !result.Valid {
					return fmt.Errorf("insufficient resources for building: %s", result.Error)
				}
			}

			// Note: Buildings don't usually consume population, but validation available if needed
		}
	case CommandProduce:
		// Resource validation for unit production
		if unitType, ok := command.Parameters["unit_type"].(string); ok {
			cost := cp.getUnitCost(unitType, unit.PlayerID)
			if cost != nil {
				validator := NewResourceValidator(cp.world)
				result := validator.ValidateResources(ResourceCheck{
					PlayerID: unit.PlayerID,
					Required: cost,
					Purpose:  fmt.Sprintf("unit production (%s)", unitType),
				})
				if !result.Valid {
					return fmt.Errorf("insufficient resources for unit: %s", result.Error)
				}
			}

			// Population validation for unit production
			popManager := NewPopulationManager(cp.world)
			canCreate, reason := popManager.CanCreateUnit(unit.PlayerID, unitType)
			if !canCreate {
				return fmt.Errorf("cannot create unit: %s", reason)
			}
		}
	case CommandUpgrade:
		// Resource validation for upgrades
		if upgradeType, ok := command.Parameters["upgrade_type"].(string); ok {
			cost := cp.getUpgradeCost(upgradeType, unit.PlayerID)
			if cost != nil {
				validator := NewResourceValidator(cp.world)
				result := validator.ValidateResources(ResourceCheck{
					PlayerID: unit.PlayerID,
					Required: cost,
					Purpose:  fmt.Sprintf("upgrade (%s)", upgradeType),
				})
				if !result.Valid {
					return fmt.Errorf("insufficient resources for upgrade: %s", result.Error)
				}
			}
		}
	case CommandFollow:
		if command.TargetUnit == nil {
			return fmt.Errorf("follow command requires target unit")
		}
	}

	return nil
}

// Command execution methods

func (cp *CommandProcessor) startCommand(unit *GameUnit, command *UnitCommand) {
	switch command.Type {
	case CommandMove:
		unit.State = UnitStateMoving
		// Initialize grid target if only world target was provided
		if command.GridTarget == nil && command.Target != nil {
			gridTarget := WorldToGrid(*command.Target, cp.world.tileSize)
			command.GridTarget = &gridTarget
		}
		unit.Target = command.Target
	case CommandAttack:
		unit.State = UnitStateAttacking
		unit.AttackTarget = command.TargetUnit
	case CommandGather:
		unit.State = UnitStateGathering
		unit.GatherTarget = command.TargetResource
	case CommandBuild:
		unit.State = UnitStateBuilding
		// Initialize grid target for building position validation
		if command.GridTarget == nil && command.Target != nil {
			gridTarget := WorldToGrid(*command.Target, cp.world.tileSize)
			command.GridTarget = &gridTarget
		}
		unit.Target = command.Target
	case CommandStop:
		unit.State = UnitStateIdle
	case CommandHold:
		unit.State = UnitStateIdle
	}
}

func (cp *CommandProcessor) processMoveCommand(unit *GameUnit, command *UnitCommand, deltaTime time.Duration) {
	// Handle A* pathfinding-based movement

	// Validate command target
	if command.Target == nil {
		unit.CurrentCommand = nil
		unit.State = UnitStateIdle
		return
	}

	// Initialize pathfinding if unit doesn't have a computed path
	if unit.Path == nil || len(unit.Path) == 0 || unit.PathIndex >= len(unit.Path) {
		// Request new path from pathfinding system
		pathResult, err := cp.world.pathfindingMgr.RequestPath(unit, *command.Target)
		if err != nil || !pathResult.Success {
			// Pathfinding failed, try to find nearest walkable position
			targetGrid := cp.world.WorldToGrid(*command.Target)
			nearestWalkable := cp.world.ObjectManager.UnitManager.FindNearestFreePosition(targetGrid.Grid)

			nearestWorldPos := cp.world.GridToWorld(GridPosition{Grid: nearestWalkable})
			fallbackResult, fallbackErr := cp.world.pathfindingMgr.RequestPath(unit, nearestWorldPos)

			if fallbackErr != nil || !fallbackResult.Success {
				// Complete pathfinding failure, cancel command
				unit.CurrentCommand = nil
				unit.State = UnitStateIdle
				unit.Target = nil
				return
			}

			// Use fallback path
			pathResult = fallbackResult
		}

		// Store computed path in unit
		unit.Path = pathResult.Path
		unit.PathIndex = 0

		// If path is partial, update command target to achievable position
		if pathResult.Partial && len(pathResult.Path) > 0 {
			finalTarget := pathResult.Path[len(pathResult.Path)-1]
			command.Target = &finalTarget
		}
	}

	// Check if we've reached the final target
	if unit.PathIndex >= len(unit.Path) {
		// Path completed
		unit.CurrentCommand = nil
		unit.State = UnitStateIdle
		unit.Target = nil
		unit.Path = nil
		unit.PathIndex = 0
		return
	}

	// Get current waypoint target
	currentWaypoint := unit.Path[unit.PathIndex]

	// Check if we've reached the current waypoint
	distance := cp.calculateDistance(unit.Position, currentWaypoint)
	if distance < 0.5 { // Waypoint tolerance
		// Advance to next waypoint
		unit.PathIndex++

		// Update occupancy grid for current position
		currentGrid := unit.GetGridPosition()
		newGrid := cp.world.WorldToGrid(currentWaypoint)

		if currentGrid.Grid.X != newGrid.Grid.X || currentGrid.Grid.Y != newGrid.Grid.Y {
			cp.world.SetOccupied(currentGrid.Grid, false)
			cp.world.SetOccupied(newGrid.Grid, true)
			unit.UpdatePositions(currentWaypoint, cp.world.tileSize)
		}

		// Check if this was the final waypoint
		if unit.PathIndex >= len(unit.Path) {
			// Reached destination
			unit.CurrentCommand = nil
			unit.State = UnitStateIdle
			unit.Target = nil
			unit.Path = nil
			unit.PathIndex = 0
			return
		}

		// Update to next waypoint
		currentWaypoint = unit.Path[unit.PathIndex]
	}

	// Move toward current waypoint
	nextPos := cp.calculateNextPosition(unit, currentWaypoint, deltaTime)
	nextGrid := cp.world.WorldToGrid(nextPos)

	// Check if next position is still walkable (dynamic obstacles)
	if cp.world.IsWalkable(nextGrid) && !cp.world.IsOccupied(nextGrid) {
		// Path is clear, continue movement
		oldGridPos := unit.GetGridPosition()
		unit.UpdatePositions(nextPos, cp.world.tileSize)

		// Update occupancy grid if unit moved to different tile
		newGridPos := cp.world.WorldToGrid(nextPos)
		if oldGridPos.Grid.X != newGridPos.Grid.X || oldGridPos.Grid.Y != newGridPos.Grid.Y {
			cp.world.SetOccupied(oldGridPos.Grid, false)
			cp.world.SetOccupied(newGridPos.Grid, true)
		}
	} else {
		// Path blocked by dynamic obstacle, recalculate path
		pathResult, err := cp.world.pathfindingMgr.RequestPath(unit, *command.Target)
		if err != nil || !pathResult.Success {
			// Cannot find alternative path, stop
			unit.CurrentCommand = nil
			unit.State = UnitStateIdle
			unit.Target = nil
			unit.Path = nil
			unit.PathIndex = 0
		} else {
			// Update with new path
			unit.Path = pathResult.Path
			unit.PathIndex = 0
		}
	}

	// Set movement target for unit.updateMovement()
	unit.Target = &currentWaypoint
}

func (cp *CommandProcessor) processAttackCommand(unit *GameUnit, command *UnitCommand, deltaTime time.Duration) {
	target := command.TargetUnit

	// Validate target
	if target == nil || !target.IsAlive() {
		cp.cancelAttackCommand(unit, "target is dead or invalid")
		return
	}

	// Check if unit can still attack this target
	canAttack, reason := cp.combatSystem.CanAttack(unit, target)
	if !canAttack {
		if reason == "target out of range" {
			// Try to move closer to attack
			cp.moveToAttackPosition(unit, target)
			return
		} else {
			// Cannot attack for other reasons (no line of sight, cooldown, etc.)
			cp.cancelAttackCommand(unit, reason)
			return
		}
	}

	// Unit is in position and can attack - execute the attack
	cp.executeAttack(unit, target)
}

// moveToAttackPosition moves a unit to the optimal position for attacking a target
func (cp *CommandProcessor) moveToAttackPosition(unit *GameUnit, target *GameUnit) {
	// Get optimal attack position from combat system
	attackPos, found := cp.combatSystem.GetOptimalAttackPosition(unit, target)

	if !found {
		cp.cancelAttackCommand(unit, "no valid attack position found")
		return
	}

	// Set unit to move to attack position
	unit.State = UnitStateMoving
	unit.Target = &attackPos
	unit.AttackTarget = target

	// Keep the attack command active - we'll continue attacking once in position
}

// executeAttack performs the actual attack execution
func (cp *CommandProcessor) executeAttack(unit *GameUnit, target *GameUnit) {
	unit.State = UnitStateAttacking
	unit.AttackTarget = target
	unit.Target = nil

	// Execute advanced combat with AOE, formation bonuses, and special effects
	advancedResult := cp.combatSystem.ExecuteAdvancedAttack(unit, target)

	// Create visual effects for the attack
	allVictims := make([]SplashVictim, 0)
	if advancedResult.PrimaryTarget != nil {
		allVictims = append(allVictims, SplashVictim{
			Unit:     advancedResult.PrimaryTarget,
			Damage:   advancedResult.PrimaryDamage,
			Distance: 0.0,
		})
	}
	allVictims = append(allVictims, advancedResult.SplashTargets...)

	if len(allVictims) > 0 {
		// Get the damage type for visual effects
		damageType := cp.getUnitAdvancedDamageType(unit)

		if len(advancedResult.SplashTargets) > 0 {
			// AOE attack - create splash damage visual effect
			cp.visualSystem.CreateSplashDamageEffect(
				target.Position,
				damageType.SplashRadius,
				damageType.Name,
				allVictims,
			)
		} else {
			// Single target attack - create appropriate effect
			if cp.isRangedAttack(damageType) {
				cp.visualSystem.CreateRangedAttackEffect(
					unit.Position,
					target.Position,
					damageType.Name,
					"default_projectile",
				)
			} else {
				cp.visualSystem.CreateMeleeHitEffect(
					unit.Position,
					target.Position,
					damageType.Name,
					advancedResult.PrimaryDamage,
					false, // isCritical - placeholder
				)
			}
		}
	}

	// Apply status effects to all victims
	for _, victim := range allVictims {
		if victim.Unit != nil {
			// Apply any special effects from the attack
			damageType := cp.getUnitAdvancedDamageType(unit)
			for _, effect := range damageType.SpecialEffects {
				// Apply the effect directly based on type
				cp.applySpecialCombatEffect(victim.Unit, effect, unit)
			}
		}
	}

	// Check if primary target was killed
	primaryTarget := advancedResult.PrimaryTarget
	if primaryTarget != nil && !primaryTarget.IsAlive() {
		// Primary target died, attack command is complete
		unit.CurrentCommand = nil
		unit.State = UnitStateIdle
		unit.AttackTarget = nil
		return
	}

	// Target is still alive, continue attacking if no other commands are queued
}

// getUnitAdvancedDamageType determines the advanced damage type for a unit
func (cp *CommandProcessor) getUnitAdvancedDamageType(unit *GameUnit) AdvancedDamageType {
	// Map unit types to advanced damage types
	switch unit.UnitType {
	case "catapult", "siege_engine", "ballista":
		return AdvancedDamageTypes["catapult"]
	case "mage", "wizard", "sorcerer":
		return AdvancedDamageTypes["fireball"]
	case "archer", "bowman", "crossbow":
		return AdvancedDamageTypes["arrow"]
	case "warrior", "swordsman", "knight":
		return AdvancedDamageTypes["sword"]
	case "lightning_mage", "storm_caller":
		return AdvancedDamageTypes["lightning"]
	default:
		// Default to sword (melee) damage
		return AdvancedDamageTypes["sword"]
	}
}

// isRangedAttack checks if a damage type represents a ranged attack
func (cp *CommandProcessor) isRangedAttack(damageType AdvancedDamageType) bool {
	switch damageType.Name {
	case "arrow", "catapult", "fireball", "lightning":
		return true
	case "sword", "burn", "explosion":
		return false
	default:
		return false
	}
}

// applySpecialCombatEffect applies special combat effects to units
func (cp *CommandProcessor) applySpecialCombatEffect(target *GameUnit, effect CombatEffect, source *GameUnit) {
	switch effect.Type {
	case EffectPoison:
		cp.statusEffectMgr.ApplyStatusEffect(target, "poison", source)
	case EffectBurn:
		cp.statusEffectMgr.ApplyStatusEffect(target, "burn", source)
	case EffectStun:
		cp.statusEffectMgr.ApplyStatusEffect(target, "stun", source)
	case EffectSlow:
		cp.statusEffectMgr.ApplyStatusEffect(target, "slow", source)
		// Add more effect types as needed
	}
}

// cancelAttackCommand cancels the current attack command and resets unit state
func (cp *CommandProcessor) cancelAttackCommand(unit *GameUnit, reason string) {
	unit.CurrentCommand = nil
	unit.State = UnitStateIdle
	unit.AttackTarget = nil
	unit.Target = nil

	// Log cancellation reason for debugging
	// fmt.Printf("Attack command cancelled for unit %d: %s\n", unit.ID, reason)
}

func (cp *CommandProcessor) processGatherCommand(unit *GameUnit, command *UnitCommand, deltaTime time.Duration) {
	if command.TargetResource == nil || command.TargetResource.Amount <= 0 {
		unit.CurrentCommand = nil
		unit.State = UnitStateIdle
		unit.GatherTarget = nil
		return
	}

	// Check if carrying capacity is full
	totalCarried := 0
	for _, amount := range unit.CarriedResources {
		totalCarried += amount
	}
	if totalCarried >= 100 { // Max carrying capacity
		// Find nearest drop-off point (simplified - would be player's buildings)
		unit.CurrentCommand = nil
		unit.State = UnitStateIdle
		// Would add auto-return-to-base logic here
		return
	}

	// Check if we're close enough to gather (using grid-based positioning)
	resourceGrid := WorldToGrid(command.TargetResource.Position, cp.world.tileSize)
	unitGrid := unit.GetGridPosition()

	// Calculate grid distance for gathering range
	gridDistance := CalculateGridDistance(unitGrid.Grid, resourceGrid.Grid)
	if gridDistance <= 1 { // Adjacent tiles can gather
		// Start gathering
		unit.State = UnitStateGathering
		unit.GatherTarget = command.TargetResource
		unit.Target = nil
	} else {
		// Move closer to resource using grid-aware pathfinding
		// Find a position adjacent to the resource
		neighbors := GetCardinalNeighbors(resourceGrid.Grid)
		var targetGrid GridPosition

		// Find the nearest walkable position adjacent to the resource
		bestDistance := float64(999)
		targetFound := false

		for _, neighbor := range neighbors {
			if cp.world.IsPositionWalkable(neighbor) {
				distance := CalculateGridDistanceFloat(unitGrid, GridPosition{
					Grid:   neighbor,
					Offset: Vector2{X: 0.5, Y: 0.5},
				})
				if distance < bestDistance {
					bestDistance = distance
					targetGrid = GridPosition{
						Grid:   neighbor,
						Offset: Vector2{X: 0.5, Y: 0.5},
					}
					targetFound = true
				}
			}
		}

		if targetFound {
			// Move to the adjacent position
			worldTarget := GridToWorld(targetGrid, cp.world.tileSize)
			unit.State = UnitStateMoving
			unit.Target = &worldTarget
		} else {
			// No accessible position found
			unit.CurrentCommand = nil
			unit.State = UnitStateIdle
			unit.GatherTarget = nil
		}
	}
}

func (cp *CommandProcessor) processBuildCommand(unit *GameUnit, command *UnitCommand, deltaTime time.Duration) {
	if command.Target == nil {
		unit.CurrentCommand = nil
		unit.State = UnitStateIdle
		return
	}

	// Validate build location using grid coordinates
	var buildGrid GridPosition
	if command.GridTarget != nil {
		buildGrid = *command.GridTarget
	} else {
		buildGrid = WorldToGrid(*command.Target, cp.world.tileSize)
		command.GridTarget = &buildGrid
	}

	// Check if build location is valid (walkable and unoccupied)
	if !cp.world.IsPositionWalkable(buildGrid.Grid) {
		// Find nearest valid build location
		alternativePos := cp.world.ObjectManager.UnitManager.FindNearestFreePosition(buildGrid.Grid)
		buildGrid.Grid = alternativePos
		buildGrid.Offset = Vector2{X: 0.5, Y: 0.5}
		command.GridTarget = &buildGrid

		// Update world target
		worldTarget := GridToWorld(buildGrid, cp.world.tileSize)
		command.Target = &worldTarget
	}

	// Check if we're adjacent to the build location (builders need to be next to build site)
	unitGrid := unit.GetGridPosition()
	buildDistance := CalculateGridDistance(unitGrid.Grid, buildGrid.Grid)

	if buildDistance <= 1 {
		// Start building
		unit.State = UnitStateBuilding
		if unit.BuildTarget == nil {
			// Get building type from command parameters
			buildingType := "basic_building"
			if buildingTypeParam, ok := command.Parameters["building_type"]; ok {
				if bt, ok := buildingTypeParam.(string); ok {
					buildingType = bt
				}
			}

			// Check and deduct resources for building construction
			cost := cp.getBuildingCost(buildingType, unit.PlayerID)
			if cost != nil {
				err := cp.world.DeductResources(unit.PlayerID, cost, "building_construction")
				if err != nil {
					// Insufficient resources, cancel command
					unit.CurrentCommand = nil
					unit.State = UnitStateIdle
					return
				}
			}

			// Mark build location as occupied
			cp.world.SetOccupied(buildGrid.Grid, true)
			cp.world.SetWalkable(buildGrid.Grid, false)

			// Load building definition for creation
			player := cp.world.GetPlayer(unit.PlayerID)
			var buildingDef *data.UnitDefinition
			if player != nil && player.FactionData != nil {
				buildingDef, _ = cp.world.assetMgr.LoadUnit(player.FactionName, buildingType)
			}

			// Create actual building using ObjectManager
			worldPos := GridToWorld(buildGrid, cp.world.tileSize)
			building, err := cp.world.ObjectManager.CreateBuilding(unit.PlayerID, buildingType, worldPos, buildingDef)
			if err != nil {
				// Refund resources on creation failure
				if cost != nil {
					cp.world.AddResources(unit.PlayerID, cost, "construction_refund")
				}
				// Restore walkability
				cp.world.SetOccupied(buildGrid.Grid, false)
				cp.world.SetWalkable(buildGrid.Grid, true)
				unit.CurrentCommand = nil
				unit.State = UnitStateIdle
				return
			}

			// Set build target and track construction progress
			unit.BuildTarget = building
		}
	} else {
		// Move to a position adjacent to the build site
		neighbors := GetCardinalNeighbors(buildGrid.Grid)
		var targetGrid GridPosition
		targetFound := false

		// Find the nearest walkable position adjacent to the build site
		bestDistance := float64(999)
		for _, neighbor := range neighbors {
			if cp.world.IsPositionWalkable(neighbor) {
				distance := CalculateGridDistanceFloat(unitGrid, GridPosition{
					Grid:   neighbor,
					Offset: Vector2{X: 0.5, Y: 0.5},
				})
				if distance < bestDistance {
					bestDistance = distance
					targetGrid = GridPosition{
						Grid:   neighbor,
						Offset: Vector2{X: 0.5, Y: 0.5},
					}
					targetFound = true
				}
			}
		}

		if targetFound {
			// Move to the adjacent position
			worldTarget := GridToWorld(targetGrid, cp.world.tileSize)
			unit.State = UnitStateMoving
			unit.Target = &worldTarget
		} else {
			// No accessible position found, cancel build
			unit.CurrentCommand = nil
			unit.State = UnitStateIdle
		}
	}
}

func (cp *CommandProcessor) processRepairCommand(unit *GameUnit, command *UnitCommand, deltaTime time.Duration) {
	// Simplified repair logic
	if command.TargetBuilding == nil {
		unit.CurrentCommand = nil
		unit.State = UnitStateIdle
		return
	}

	// Check if building needs repair
	if command.TargetBuilding.Health >= command.TargetBuilding.MaxHealth {
		unit.CurrentCommand = nil
		unit.State = UnitStateIdle
		return
	}

	// Check if we're close enough
	distance := cp.calculateDistance(unit.Position, command.TargetBuilding.Position)
	if distance > 3.0 {
		// Move closer
		unit.State = UnitStateMoving
		unit.Target = &command.TargetBuilding.Position
	} else {
		// Repair
		repairRate := 10.0 * float32(deltaTime.Seconds()) // 10 HP per second
		newHealth := command.TargetBuilding.Health + int(repairRate)
		if newHealth > command.TargetBuilding.MaxHealth {
			newHealth = command.TargetBuilding.MaxHealth
		}
		command.TargetBuilding.SetHealth(newHealth)
	}
}

func (cp *CommandProcessor) processStopCommand(unit *GameUnit, command *UnitCommand) {
	unit.CurrentCommand = nil
	unit.State = UnitStateIdle
	unit.Target = nil
	unit.AttackTarget = nil
	unit.GatherTarget = nil
	unit.BuildTarget = nil
}

func (cp *CommandProcessor) processHoldCommand(unit *GameUnit, command *UnitCommand) {
	unit.State = UnitStateIdle
	unit.Target = nil
	// Unit will defend position but not chase enemies
}

func (cp *CommandProcessor) processPatrolCommand(unit *GameUnit, command *UnitCommand, deltaTime time.Duration) {
	// Simplified patrol logic - would patrol between current position and target
	if command.Target == nil {
		unit.CurrentCommand = nil
		unit.State = UnitStateIdle
		return
	}

	distance := cp.calculateDistance(unit.Position, *command.Target)
	if distance < 0.5 {
		// Reached patrol point, reverse direction
		originalPos := unit.Position
		unit.Position = *command.Target
		command.Target = &originalPos
	}

	unit.State = UnitStateMoving
	unit.Target = command.Target
}

func (cp *CommandProcessor) processFollowCommand(unit *GameUnit, command *UnitCommand, deltaTime time.Duration) {
	if command.TargetUnit == nil || !command.TargetUnit.IsAlive() {
		unit.CurrentCommand = nil
		unit.State = UnitStateIdle
		return
	}

	// Follow at a distance
	distance := cp.calculateDistance(unit.Position, command.TargetUnit.Position)
	if distance > 3.0 { // Follow distance
		unit.State = UnitStateMoving
		unit.Target = &command.TargetUnit.Position
	} else {
		unit.State = UnitStateIdle
		unit.Target = nil
	}
}

func (cp *CommandProcessor) processGuardCommand(unit *GameUnit, command *UnitCommand, deltaTime time.Duration) {
	// Guard a position or unit
	if command.TargetUnit != nil {
		// Guard unit
		distance := cp.calculateDistance(unit.Position, command.TargetUnit.Position)
		if distance > 5.0 {
			// Move closer to guarded unit
			unit.State = UnitStateMoving
			unit.Target = &command.TargetUnit.Position
		} else {
			// Look for enemies near the guarded unit
			unit.State = UnitStateIdle
			// Would scan for nearby enemies and attack them
		}
	} else if command.Target != nil {
		// Guard position
		distance := cp.calculateDistance(unit.Position, *command.Target)
		if distance > 2.0 {
			// Return to guard position
			unit.State = UnitStateMoving
			unit.Target = command.Target
		} else {
			unit.State = UnitStateIdle
			// Look for enemies in the area
		}
	}
}

// Building command methods

func (cp *CommandProcessor) startProduction(building *GameBuilding, command UnitCommand) error {
	if !building.IsBuilt {
		return fmt.Errorf("building is not complete")
	}

	// Get production parameters
	unitType, ok := command.Parameters["unit_type"].(string)
	if !ok {
		return fmt.Errorf("production command requires unit_type parameter")
	}

	duration := 30 * time.Second // Default production time
	if durationParam, ok := command.Parameters["duration"]; ok {
		if d, ok := durationParam.(time.Duration); ok {
			duration = d
		}
	}

	// Get cost from AssetManager or command parameters
	cost := cp.getUnitCost(unitType, building.PlayerID)
	if cost == nil {
		// Fallback to cost from command parameters
		cost = make(map[string]int)
		if costParam, ok := command.Parameters["cost"]; ok {
			if c, ok := costParam.(map[string]int); ok {
				cost = c
			}
		}
	}

	// Deduct resources before starting production
	if len(cost) > 0 {
		err := cp.world.DeductResources(building.PlayerID, cost, "unit_production")
		if err != nil {
			return fmt.Errorf("insufficient resources for %s production: %w", unitType, err)
		}
	}

	// Create production item
	productionItem := ProductionItem{
		ItemType:  "unit",
		ItemName:  unitType,
		Progress:  0.0,
		Duration:  duration,
		Cost:      cost,
		StartTime: time.Now(),
	}

	// Add to production queue
	building.ProductionQueue = append(building.ProductionQueue, productionItem)

	return nil
}

func (cp *CommandProcessor) startUpgrade(building *GameBuilding, command UnitCommand) error {
	if !building.IsBuilt {
		return fmt.Errorf("building is not complete")
	}

	if building.UpgradeLevel >= building.MaxUpgradeLevel {
		return fmt.Errorf("building is already at maximum upgrade level")
	}

	if building.CurrentUpgrade != nil {
		return fmt.Errorf("building is already upgrading")
	}

	// Get upgrade parameters
	upgradeType, ok := command.Parameters["upgrade_type"].(string)
	if !ok {
		upgradeType = "level_upgrade"
	}

	duration := 60 * time.Second // Default upgrade time
	if durationParam, ok := command.Parameters["duration"]; ok {
		if d, ok := durationParam.(time.Duration); ok {
			duration = d
		}
	}

	cost := make(map[string]int)
	if costParam, ok := command.Parameters["cost"]; ok {
		if c, ok := costParam.(map[string]int); ok {
			cost = c
		}
	}

	// Create upgrade item
	upgradeItem := UpgradeItem{
		UpgradeType: upgradeType,
		UpgradeName: fmt.Sprintf("Level %d Upgrade", building.UpgradeLevel+1),
		Progress:    0.0,
		Duration:    duration,
		Cost:        cost,
		StartTime:   time.Now(),
	}

	building.CurrentUpgrade = &upgradeItem

	return nil
}

// Helper methods

func (cp *CommandProcessor) calculateDistance(pos1, pos2 Vector3) float32 {
	dx := pos1.X - pos2.X
	dy := pos1.Y - pos2.Y
	dz := pos1.Z - pos2.Z
	return float32(math.Sqrt(float64(dx*dx + dy*dy + dz*dz)))
}

// calculateNextPosition calculates the next position for a unit moving toward target
func (cp *CommandProcessor) calculateNextPosition(unit *GameUnit, target Vector3, deltaTime time.Duration) Vector3 {
	currentPos := unit.GetPosition()

	// Calculate direction vector
	dx := target.X - currentPos.X
	dy := target.Y - currentPos.Y
	dz := target.Z - currentPos.Z

	// Calculate distance to target
	distance := math.Sqrt(dx*dx + dy*dy + dz*dz)

	if distance < 0.01 {
		// Already at target
		return currentPos
	}

	// Calculate movement distance based on unit speed
	moveDistance := float64(unit.Speed) * deltaTime.Seconds()

	// Don't overshoot the target
	if moveDistance > distance {
		return target
	}

	// Calculate normalized direction and apply movement
	factor := moveDistance / distance
	return Vector3{
		X: currentPos.X + dx*factor,
		Y: currentPos.Y + dy*factor,
		Z: currentPos.Z + dz*factor,
	}
}

// Command creation helpers for easier command building

// CreateMoveCommand creates a move command
func CreateMoveCommand(target Vector3, queued bool) UnitCommand {
	return UnitCommand{
		Type:       CommandMove,
		Target:     &target,
		Parameters: make(map[string]interface{}),
		IsQueued:   queued,
	}
}

// CreateGridMoveCommand creates a move command with grid coordinates
func CreateGridMoveCommand(gridTarget GridPosition, tileSize float32, queued bool) UnitCommand {
	worldTarget := GridToWorld(gridTarget, tileSize)
	return UnitCommand{
		Type:       CommandMove,
		Target:     &worldTarget,
		GridTarget: &gridTarget,
		Parameters: make(map[string]interface{}),
		IsQueued:   queued,
	}
}

// CreateAttackCommand creates an attack command
func CreateAttackCommand(target *GameUnit, queued bool) UnitCommand {
	return UnitCommand{
		Type:       CommandAttack,
		TargetUnit: target,
		Parameters: make(map[string]interface{}),
		IsQueued:   queued,
	}
}

// CreateGatherCommand creates a gather command
func CreateGatherCommand(target *ResourceNode, queued bool) UnitCommand {
	return UnitCommand{
		Type:           CommandGather,
		TargetResource: target,
		Parameters:     make(map[string]interface{}),
		IsQueued:       queued,
	}
}

// CreateBuildCommand creates a build command
func CreateBuildCommand(position Vector3, buildingType string, queued bool) UnitCommand {
	params := make(map[string]interface{})
	params["building_type"] = buildingType

	return UnitCommand{
		Type:       CommandBuild,
		Target:     &position,
		Parameters: params,
		IsQueued:   queued,
	}
}

// CreateGridBuildCommand creates a build command with grid coordinates
func CreateGridBuildCommand(gridPosition GridPosition, buildingType string, tileSize float32, queued bool) UnitCommand {
	params := make(map[string]interface{})
	params["building_type"] = buildingType

	worldPosition := GridToWorld(gridPosition, tileSize)
	return UnitCommand{
		Type:       CommandBuild,
		Target:     &worldPosition,
		GridTarget: &gridPosition,
		Parameters: params,
		IsQueued:   queued,
	}
}

// CreateProduceCommand creates a production command for buildings
func CreateProduceCommand(unitType string, cost map[string]int) UnitCommand {
	params := make(map[string]interface{})
	params["unit_type"] = unitType
	params["cost"] = cost
	params["duration"] = 30 * time.Second

	return UnitCommand{
		Type:       CommandProduce,
		Parameters: params,
	}
}

// CreateUpgradeCommand creates an upgrade command for buildings
func CreateUpgradeCommand(upgradeType string, cost map[string]int) UnitCommand {
	params := make(map[string]interface{})
	params["upgrade_type"] = upgradeType
	params["cost"] = cost
	params["duration"] = 60 * time.Second

	return UnitCommand{
		Type:       CommandUpgrade,
		Parameters: params,
	}
}

// String methods for debugging

func (ct CommandType) String() string {
	switch ct {
	case CommandMove:
		return "Move"
	case CommandAttack:
		return "Attack"
	case CommandGather:
		return "Gather"
	case CommandBuild:
		return "Build"
	case CommandRepair:
		return "Repair"
	case CommandStop:
		return "Stop"
	case CommandHold:
		return "Hold"
	case CommandPatrol:
		return "Patrol"
	case CommandFollow:
		return "Follow"
	case CommandGuard:
		return "Guard"
	case CommandProduce:
		return "Produce"
	case CommandUpgrade:
		return "Upgrade"
	default:
		return "Unknown"
	}
}

// Resource Cost Helper Methods

// getUnitCost extracts resource cost for creating a unit from AssetManager
func (cp *CommandProcessor) getUnitCost(unitType string, playerID int) map[string]int {
	player := cp.world.GetPlayer(playerID)
	if player == nil || player.FactionData == nil {
		return nil
	}

	// Load unit from AssetManager
	unit, err := cp.world.assetMgr.LoadUnit(player.FactionName, unitType)
	if err != nil {
		return nil
	}

	// Extract resource requirements
	costs := make(map[string]int)
	for _, req := range unit.Unit.Parameters.ResourceRequirements {
		costs[req.Name] = req.Amount
	}

	return costs
}

// getBuildingCost extracts resource cost for constructing a building from AssetManager
func (cp *CommandProcessor) getBuildingCost(buildingType string, playerID int) map[string]int {
	player := cp.world.GetPlayer(playerID)
	if player == nil || player.FactionData == nil {
		return nil
	}

	// Load building from AssetManager (buildings are also loaded as units in MegaGlest)
	building, err := cp.world.assetMgr.LoadUnit(player.FactionName, buildingType)
	if err != nil {
		return nil
	}

	// Extract resource requirements
	costs := make(map[string]int)
	for _, req := range building.Unit.Parameters.ResourceRequirements {
		costs[req.Name] = req.Amount
	}

	return costs
}

// getUpgradeCost extracts resource cost for an upgrade (placeholder implementation)
func (cp *CommandProcessor) getUpgradeCost(upgradeType string, playerID int) map[string]int {
	// TODO: Implement upgrade cost extraction from AssetManager
	// For now, return basic costs for common upgrades
	upgradeCosts := map[string]map[string]int{
		"weapon_upgrade": {"gold": 150, "wood": 100},
		"armor_upgrade":  {"gold": 200, "stone": 150},
		"speed_upgrade":  {"gold": 100, "wood": 50},
	}

	if cost, exists := upgradeCosts[upgradeType]; exists {
		return cost
	}

	return nil
}


// CreateStopCommand creates a stop command
func CreateStopCommand() UnitCommand {
	return UnitCommand{
		Type:      CommandStop,
		CreatedAt: time.Now(),
	}
}

// CreatePatrolCommand creates a patrol command
func CreatePatrolCommand(target Vector3, queued bool) UnitCommand {
	return UnitCommand{
		Type:      CommandPatrol,
		Target:    &target,
		IsQueued:  queued,
		CreatedAt: time.Now(),
	}
}

// Priority constants for commands
const (
	PriorityLow      = 1
	PriorityNormal   = 2
	PriorityHigh     = 3
	PriorityCritical = 4
)

// SortCommandsByPriority sorts commands by priority (highest first)
func SortCommandsByPriority(commands []UnitCommand) {
	for i := 0; i < len(commands)-1; i++ {
		for j := i + 1; j < len(commands); j++ {
			if commands[j].Priority > commands[i].Priority {
				commands[i], commands[j] = commands[j], commands[i]
			}
		}
	}
}

// Production System Integration Methods

// IssueUnitProductionCommand issues a unit production command to a building
func (cp *CommandProcessor) IssueUnitProductionCommand(buildingID int, unitType string) error {
	building := cp.world.ObjectManager.GetBuilding(buildingID)
	if building == nil {
		return fmt.Errorf("building %d not found", buildingID)
	}

	// Get unit cost from asset manager
	cost := cp.getUnitCost(unitType, building.PlayerID)
	if cost == nil {
		// Fallback to default costs
		cost = map[string]int{"wood": 50, "gold": 25}
	}

	// Get production duration (can be configured per unit type)
	duration := 30 * time.Second
	switch unitType {
	case "worker", "peasant", "initiate":
		duration = 20 * time.Second
	case "swordman", "archer", "daemon":
		duration = 30 * time.Second
	case "horseman", "catapult", "archmage":
		duration = 45 * time.Second
	case "dragon", "behemoth":
		duration = 90 * time.Second
	}

	// Use production system to handle the command
	return cp.world.productionSys.IssueProductionCommand(buildingID, unitType, cost, duration)
}

// StartResearchCommand initiates research at a building
func (cp *CommandProcessor) StartResearchCommand(buildingID int, technologyName string) error {
	building := cp.world.ObjectManager.GetBuilding(buildingID)
	if building == nil {
		return fmt.Errorf("building %d not found", buildingID)
	}

	// Check if building can conduct research
	if !cp.canBuildingResearch(building) {
		return fmt.Errorf("building %s cannot conduct research", building.BuildingType)
	}

	// Get technology tree from production system
	techTree := cp.world.productionSys.GetTechnologyTree()

	// Validate and deduct resources for research
	techDef := techTree.GetTechnologyDefinition(technologyName)
	if techDef == nil {
		return fmt.Errorf("technology %s not found", technologyName)
	}

	if len(techDef.Cost) > 0 {
		err := cp.world.DeductResources(building.PlayerID, techDef.Cost, "research")
		if err != nil {
			return fmt.Errorf("insufficient resources for research %s: %w", technologyName, err)
		}
	}

	// Start research
	return techTree.StartResearch(building.PlayerID, technologyName, buildingID)
}

// canBuildingResearch checks if a building type can conduct research
func (cp *CommandProcessor) canBuildingResearch(building *GameBuilding) bool {
	researchBuildings := map[string]bool{
		"mage_tower":     true,
		"summoner_guild": true,
		"library":        true,
		"blacksmith":     true,
		"castle":         true,
		"town_hall":      true,
	}

	return researchBuildings[building.BuildingType]
}

// GetProductionSystemStats returns production system statistics for debugging
func (cp *CommandProcessor) GetProductionSystemStats() map[string]interface{} {
	stats := make(map[string]interface{})

	// Get technology tree stats for actual players
	techTree := cp.world.productionSys.GetTechnologyTree()
	for playerID := range cp.world.players {
		playerTechs := techTree.GetPlayerTechnologies(playerID)
		stats[fmt.Sprintf("player_%d_technologies", playerID)] = len(playerTechs)

		current, queue := techTree.GetResearchProgress(playerID)
		stats[fmt.Sprintf("player_%d_current_research", playerID)] = current
		stats[fmt.Sprintf("player_%d_research_queue_length", playerID)] = len(queue)
	}

	// Get population stats for actual players
	populationMgr := cp.world.productionSys.GetPopulationManager()
	for playerID := range cp.world.players {
		popStats := populationMgr.GetPopulationStatus(playerID)
		stats[fmt.Sprintf("player_%d_population", playerID)] = fmt.Sprintf("%d/%d",
			popStats.CurrentPopulation, popStats.MaxPopulation)
	}

	return stats
}
