package engine

import (
	"fmt"
	"math"
	"time"
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
)

// CommandProcessor handles command processing for units and buildings
type CommandProcessor struct {
	world *World
}

// NewCommandProcessor creates a new command processor
func NewCommandProcessor(world *World) *CommandProcessor {
	return &CommandProcessor{
		world: world,
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
	// Handle grid-aware movement with collision detection

	// Convert target to grid coordinates if not already done
	var targetGrid GridPosition
	if command.GridTarget != nil {
		targetGrid = *command.GridTarget
	} else if command.Target != nil {
		targetGrid = WorldToGrid(*command.Target, cp.world.tileSize)
		command.GridTarget = &targetGrid
	} else {
		unit.CurrentCommand = nil
		unit.State = UnitStateIdle
		return
	}

	// Validate target position and find walkable alternative if needed
	if !cp.world.IsPositionWalkable(targetGrid.Grid) {
		// Find nearest valid position using UnitManager
		nearestWalkable := cp.world.ObjectManager.UnitManager.FindNearestFreePosition(targetGrid.Grid)
		targetGrid.Grid = nearestWalkable
		targetGrid.Offset = Vector2{X: 0.5, Y: 0.5} // Center of tile
		command.GridTarget = &targetGrid
	}

	// Convert grid target back to world coordinates
	worldTarget := GridToWorld(targetGrid, cp.world.tileSize)
	command.Target = &worldTarget

	// Check if we've reached the target (using grid-based tolerance)
	currentGrid := unit.GetGridPosition()
	if currentGrid.Grid.X == targetGrid.Grid.X && currentGrid.Grid.Y == targetGrid.Grid.Y {
		// Reached target grid tile, now check sub-tile precision
		distance := cp.calculateDistance(unit.Position, worldTarget)
		if distance < 0.3 { // Sub-tile tolerance
			// Update both coordinate systems
			unit.SetGridPosition(targetGrid, cp.world.tileSize)
			unit.CurrentCommand = nil
			unit.State = UnitStateIdle
			unit.Target = nil

			// Free old grid position and mark new one as occupied
			cp.world.SetOccupied(currentGrid.Grid, false)
			cp.world.SetOccupied(targetGrid.Grid, true)
			return
		}
	}

	// Continue movement - check for collisions along the path
	nextPos := cp.calculateNextPosition(unit, worldTarget, deltaTime)
	nextGrid := WorldToGrid(nextPos, cp.world.tileSize)

	// Check if next position is walkable
	if cp.world.IsPositionWalkable(nextGrid.Grid) {
		// Update positions if path is clear
		oldGridPos := unit.GetGridPosition().Grid
		unit.UpdatePositions(nextPos, cp.world.tileSize)

		// Update occupancy grid if unit moved to different tile
		if oldGridPos.X != nextGrid.Grid.X || oldGridPos.Y != nextGrid.Grid.Y {
			cp.world.SetOccupied(oldGridPos, false)
			cp.world.SetOccupied(nextGrid.Grid, true)
		}
	} else {
		// Path blocked, recalculate route to target
		alternativeTarget := cp.world.ObjectManager.UnitManager.FindNearestFreePosition(targetGrid.Grid)
		if alternativeTarget.X != targetGrid.Grid.X || alternativeTarget.Y != targetGrid.Grid.Y {
			// Update target to alternative position
			targetGrid.Grid = alternativeTarget
			targetGrid.Offset = Vector2{X: 0.5, Y: 0.5} // Center of tile
			command.GridTarget = &targetGrid
			worldTarget = GridToWorld(targetGrid, cp.world.tileSize)
			command.Target = &worldTarget
		} else {
			// No alternative found, stop movement
			unit.CurrentCommand = nil
			unit.State = UnitStateIdle
			unit.Target = nil
		}
	}

	// Set movement target for unit.updateMovement()
	unit.Target = command.Target
}

func (cp *CommandProcessor) processAttackCommand(unit *GameUnit, command *UnitCommand, deltaTime time.Duration) {
	if command.TargetUnit == nil || !command.TargetUnit.IsAlive() {
		unit.CurrentCommand = nil
		unit.State = UnitStateIdle
		unit.AttackTarget = nil
		return
	}

	// Check if target is in range
	distance := cp.calculateDistance(unit.Position, command.TargetUnit.Position)
	if distance > float32(unit.AttackRange) {
		// Move closer to target using grid-aware pathfinding
		targetPos := command.TargetUnit.Position
		targetGrid := WorldToGrid(targetPos, cp.world.tileSize)

		// Find a position near the target that's walkable
		if !cp.world.IsPositionWalkable(targetGrid.Grid) {
			targetGrid.Grid = cp.world.ObjectManager.UnitManager.FindNearestFreePosition(targetGrid.Grid)
		}

		// Move to an adjacent position to attack from
		neighbors := GetCardinalNeighbors(targetGrid.Grid)
		for _, neighbor := range neighbors {
			if cp.world.IsPositionWalkable(neighbor) {
				// Found a good attack position
				targetGrid.Grid = neighbor
				targetGrid.Offset = Vector2{X: 0.5, Y: 0.5}
				break
			}
		}

		// Convert back to world coordinates and set as movement target
		worldTarget := GridToWorld(targetGrid, cp.world.tileSize)
		unit.State = UnitStateMoving
		unit.Target = &worldTarget
	} else {
		// Attack target
		unit.State = UnitStateAttacking
		unit.AttackTarget = command.TargetUnit
		unit.Target = nil
	}
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
			// Mark build location as occupied
			cp.world.SetOccupied(buildGrid.Grid, true)
			cp.world.SetWalkable(buildGrid.Grid, false)

			// Create building (simplified - would check resources, etc.)
			buildingType := "basic_building"
			if buildingTypeParam, ok := command.Parameters["building_type"]; ok {
				if bt, ok := buildingTypeParam.(string); ok {
					buildingType = bt
				}
			}

			// This would integrate with the object manager to create the building
			_ = buildingType
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

	cost := make(map[string]int)
	if costParam, ok := command.Parameters["cost"]; ok {
		if c, ok := costParam.(map[string]int); ok {
			cost = c
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