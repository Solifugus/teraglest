package engine

import (
	"math"
	"time"
)

// ActionNode represents a leaf node that performs an action
type ActionNode struct {
	BaseNode
}

// ConditionNode represents a leaf node that evaluates a condition
type ConditionNode struct {
	BaseNode
}

// Action Nodes

// MoveToPositionAction moves the unit to a specified position
type MoveToPositionAction struct {
	ActionNode
	targetKey     string  // Blackboard key for target position
	tolerance     float64 // Distance tolerance for reaching target
	commandIssued bool    // Whether move command has been issued
}

// NewMoveToPositionAction creates a new move action
func NewMoveToPositionAction(name, targetKey string, tolerance float64) *MoveToPositionAction {
	return &MoveToPositionAction{
		ActionNode: ActionNode{
			BaseNode: BaseNode{name: name},
		},
		targetKey:     targetKey,
		tolerance:     tolerance,
		commandIssued: false,
	}
}

// Execute performs the move action
func (action *MoveToPositionAction) Execute(context *BehaviorContext) NodeStatus {
	// Get target position from blackboard
	targetPos, exists := context.Blackboard.GetVector3(action.targetKey)
	if !exists {
		return StatusFailure
	}

	unit := context.Unit
	currentPos := unit.Position

	// Check if we're already at the target
	distance := calculateDistance(currentPos, targetPos)
	if distance <= action.tolerance {
		return StatusSuccess
	}

	// Issue move command if not already issued
	if !action.commandIssued {
		moveCmd := CreateMoveCommand(targetPos, false)
		err := context.World.commandProcessor.IssueCommand(unit.ID, moveCmd)
		if err != nil {
			return StatusFailure
		}
		action.commandIssued = true
	}

	// Check if unit is currently moving toward target
	if unit.CurrentCommand != nil && unit.CurrentCommand.Type == CommandMove {
		return StatusRunning
	}

	// Command completed, check if we reached target
	finalDistance := calculateDistance(unit.Position, targetPos)
	if finalDistance <= action.tolerance {
		return StatusSuccess
	}

	return StatusFailure
}

// Reset resets the move action
func (action *MoveToPositionAction) Reset() {
	action.commandIssued = false
}

// AttackTargetAction attacks a specified target
type AttackTargetAction struct {
	ActionNode
	targetKey     string // Blackboard key for target unit
	commandIssued bool   // Whether attack command has been issued
}

// NewAttackTargetAction creates a new attack action
func NewAttackTargetAction(name, targetKey string) *AttackTargetAction {
	return &AttackTargetAction{
		ActionNode: ActionNode{
			BaseNode: BaseNode{name: name},
		},
		targetKey:     targetKey,
		commandIssued: false,
	}
}

// Execute performs the attack action
func (action *AttackTargetAction) Execute(context *BehaviorContext) NodeStatus {
	// Get target unit from blackboard
	targetUnit, exists := context.Blackboard.GetUnit(action.targetKey)
	if !exists || targetUnit == nil {
		return StatusFailure
	}

	// Check if target is still alive
	if !targetUnit.IsAlive() {
		return StatusSuccess // Target eliminated
	}

	unit := context.Unit

	// Issue attack command if not already issued
	if !action.commandIssued {
		attackCmd := CreateAttackCommand(targetUnit, false)
		err := context.World.commandProcessor.IssueCommand(unit.ID, attackCmd)
		if err != nil {
			return StatusFailure
		}
		action.commandIssued = true
	}

	// Check if unit is currently attacking
	if unit.CurrentCommand != nil && unit.CurrentCommand.Type == CommandAttack {
		return StatusRunning
	}

	// Check if target was eliminated
	if !targetUnit.IsAlive() {
		return StatusSuccess
	}

	return StatusRunning
}

// Reset resets the attack action
func (action *AttackTargetAction) Reset() {
	action.commandIssued = false
}

// GatherResourceAction gathers from a specified resource node
type GatherResourceAction struct {
	ActionNode
	resourceKey   string // Blackboard key for target resource
	commandIssued bool   // Whether gather command has been issued
}

// NewGatherResourceAction creates a new gather action
func NewGatherResourceAction(name, resourceKey string) *GatherResourceAction {
	return &GatherResourceAction{
		ActionNode: ActionNode{
			BaseNode: BaseNode{name: name},
		},
		resourceKey:   resourceKey,
		commandIssued: false,
	}
}

// Execute performs the gather action
func (action *GatherResourceAction) Execute(context *BehaviorContext) NodeStatus {
	// Get target resource from blackboard
	resource, exists := context.Blackboard.Get(action.resourceKey)
	if !exists {
		return StatusFailure
	}

	resourceNode, ok := resource.(*ResourceNode)
	if !ok || resourceNode == nil {
		return StatusFailure
	}

	// Check if resource is depleted
	if resourceNode.Amount <= 0 {
		return StatusSuccess // Resource exhausted
	}

	unit := context.Unit

	// Check if unit is at carrying capacity
	totalCarried := 0
	for _, amount := range unit.CarriedResources {
		totalCarried += amount
	}
	if totalCarried >= 100 { // Max carrying capacity
		return StatusSuccess // Unit full
	}

	// Issue gather command if not already issued
	if !action.commandIssued {
		gatherCmd := CreateGatherCommand(resourceNode, false)
		err := context.World.commandProcessor.IssueCommand(unit.ID, gatherCmd)
		if err != nil {
			return StatusFailure
		}
		action.commandIssued = true
	}

	// Check if unit is currently gathering
	if unit.CurrentCommand != nil && unit.CurrentCommand.Type == CommandGather {
		return StatusRunning
	}

	// Check completion conditions
	if resourceNode.Amount <= 0 || totalCarried >= 100 {
		return StatusSuccess
	}

	return StatusRunning
}

// Reset resets the gather action
func (action *GatherResourceAction) Reset() {
	action.commandIssued = false
}

// BuildStructureAction builds a structure at a specified location
type BuildStructureAction struct {
	ActionNode
	positionKey   string // Blackboard key for build position
	buildingType  string // Type of building to construct
	commandIssued bool   // Whether build command has been issued
}

// NewBuildStructureAction creates a new build action
func NewBuildStructureAction(name, positionKey, buildingType string) *BuildStructureAction {
	return &BuildStructureAction{
		ActionNode: ActionNode{
			BaseNode: BaseNode{name: name},
		},
		positionKey:   positionKey,
		buildingType:  buildingType,
		commandIssued: false,
	}
}

// Execute performs the build action
func (action *BuildStructureAction) Execute(context *BehaviorContext) NodeStatus {
	// Get build position from blackboard
	buildPos, exists := context.Blackboard.GetVector3(action.positionKey)
	if !exists {
		return StatusFailure
	}

	unit := context.Unit

	// Issue build command if not already issued
	if !action.commandIssued {
		buildCmd := CreateBuildCommand(buildPos, action.buildingType, false)
		err := context.World.commandProcessor.IssueCommand(unit.ID, buildCmd)
		if err != nil {
			return StatusFailure
		}
		action.commandIssued = true
	}

	// Check if unit is currently building
	if unit.CurrentCommand != nil && unit.CurrentCommand.Type == CommandBuild {
		return StatusRunning
	}

	// Check if building was completed
	if unit.BuildTarget != nil && unit.BuildTarget.IsBuilt {
		return StatusSuccess
	}

	return StatusRunning
}

// Reset resets the build action
func (action *BuildStructureAction) Reset() {
	action.commandIssued = false
}

// WaitAction waits for a specified duration
type WaitAction struct {
	ActionNode
	duration    time.Duration // How long to wait
	startTime   time.Time     // When wait started
	hasStarted  bool          // Whether wait has started
}

// NewWaitAction creates a new wait action
func NewWaitAction(name string, duration time.Duration) *WaitAction {
	return &WaitAction{
		ActionNode: ActionNode{
			BaseNode: BaseNode{name: name},
		},
		duration:   duration,
		hasStarted: false,
	}
}

// Execute performs the wait action
func (action *WaitAction) Execute(context *BehaviorContext) NodeStatus {
	if !action.hasStarted {
		action.startTime = time.Now()
		action.hasStarted = true
	}

	elapsed := time.Since(action.startTime)
	if elapsed >= action.duration {
		return StatusSuccess
	}

	return StatusRunning
}

// Reset resets the wait action
func (action *WaitAction) Reset() {
	action.hasStarted = false
}

// SetBlackboardValueAction sets a value in the blackboard
type SetBlackboardValueAction struct {
	ActionNode
	key   string      // Blackboard key to set
	value interface{} // Value to store
}

// NewSetBlackboardValueAction creates a new set blackboard value action
func NewSetBlackboardValueAction(name, key string, value interface{}) *SetBlackboardValueAction {
	return &SetBlackboardValueAction{
		ActionNode: ActionNode{
			BaseNode: BaseNode{name: name},
		},
		key:   key,
		value: value,
	}
}

// Execute sets the blackboard value
func (action *SetBlackboardValueAction) Execute(context *BehaviorContext) NodeStatus {
	context.Blackboard.Set(action.key, action.value)
	return StatusSuccess
}

// Condition Nodes

// IsHealthLowCondition checks if unit health is below a threshold
type IsHealthLowCondition struct {
	ConditionNode
	threshold float64 // Health percentage threshold (0.0 to 1.0)
}

// NewIsHealthLowCondition creates a new health check condition
func NewIsHealthLowCondition(name string, threshold float64) *IsHealthLowCondition {
	return &IsHealthLowCondition{
		ConditionNode: ConditionNode{
			BaseNode: BaseNode{name: name},
		},
		threshold: threshold,
	}
}

// Execute checks if health is low
func (condition *IsHealthLowCondition) Execute(context *BehaviorContext) NodeStatus {
	unit := context.Unit
	healthPercentage := float64(unit.Health) / float64(unit.MaxHealth)

	if healthPercentage <= condition.threshold {
		return StatusSuccess
	}
	return StatusFailure
}

// IsEnemyInRangeCondition checks if an enemy is within attack range
type IsEnemyInRangeCondition struct {
	ConditionNode
	range_    float64 // Search range
	targetKey string  // Blackboard key to store found enemy
}

// NewIsEnemyInRangeCondition creates a new enemy detection condition
func NewIsEnemyInRangeCondition(name string, searchRange float64, targetKey string) *IsEnemyInRangeCondition {
	return &IsEnemyInRangeCondition{
		ConditionNode: ConditionNode{
			BaseNode: BaseNode{name: name},
		},
		range_:    searchRange,
		targetKey: targetKey,
	}
}

// Execute checks for enemies in range
func (condition *IsEnemyInRangeCondition) Execute(context *BehaviorContext) NodeStatus {
	unit := context.Unit
	world := context.World

	// Find enemy units within range
	enemyUnits := world.ObjectManager.GetUnitsForPlayer(-1) // Get all units
	closestEnemy := (*GameUnit)(nil)
	closestDistance := condition.range_ + 1

	for _, enemy := range enemyUnits {
		// Skip friendly and dead units
		if enemy.PlayerID == unit.PlayerID || !enemy.IsAlive() {
			continue
		}

		distance := calculateDistance(unit.Position, enemy.Position)
		if distance <= condition.range_ && distance < closestDistance {
			closestEnemy = enemy
			closestDistance = distance
		}
	}

	if closestEnemy != nil {
		// Store enemy in blackboard
		context.Blackboard.Set(condition.targetKey, closestEnemy)
		return StatusSuccess
	}

	return StatusFailure
}

// IsResourceInRangeCondition checks if a resource is within gathering range
type IsResourceInRangeCondition struct {
	ConditionNode
	range_      float64 // Search range
	resourceKey string  // Blackboard key to store found resource
}

// NewIsResourceInRangeCondition creates a new resource detection condition
func NewIsResourceInRangeCondition(name string, searchRange float64, resourceKey string) *IsResourceInRangeCondition {
	return &IsResourceInRangeCondition{
		ConditionNode: ConditionNode{
			BaseNode: BaseNode{name: name},
		},
		range_:      searchRange,
		resourceKey: resourceKey,
	}
}

// Execute checks for resources in range
func (condition *IsResourceInRangeCondition) Execute(context *BehaviorContext) NodeStatus {
	unit := context.Unit
	world := context.World

	// Find resource nodes within range
	closestResource := (*ResourceNode)(nil)
	closestDistance := condition.range_ + 1

	for _, resource := range world.resources {
		// Skip depleted resources
		if resource.Amount <= 0 {
			continue
		}

		distance := calculateDistance(unit.Position, resource.Position)
		if distance <= condition.range_ && distance < closestDistance {
			closestResource = resource
			closestDistance = distance
		}
	}

	if closestResource != nil {
		// Store resource in blackboard
		context.Blackboard.Set(condition.resourceKey, closestResource)
		return StatusSuccess
	}

	return StatusFailure
}

// IsCarryingResourcesCondition checks if unit is carrying resources
type IsCarryingResourcesCondition struct {
	ConditionNode
	minAmount int // Minimum resource amount to consider "carrying"
}

// NewIsCarryingResourcesCondition creates a new resource carrying condition
func NewIsCarryingResourcesCondition(name string, minAmount int) *IsCarryingResourcesCondition {
	return &IsCarryingResourcesCondition{
		ConditionNode: ConditionNode{
			BaseNode: BaseNode{name: name},
		},
		minAmount: minAmount,
	}
}

// Execute checks if unit is carrying resources
func (condition *IsCarryingResourcesCondition) Execute(context *BehaviorContext) NodeStatus {
	unit := context.Unit

	totalCarried := 0
	for _, amount := range unit.CarriedResources {
		totalCarried += amount
	}

	if totalCarried >= condition.minAmount {
		return StatusSuccess
	}
	return StatusFailure
}

// IsBlackboardKeySetCondition checks if a blackboard key exists
type IsBlackboardKeySetCondition struct {
	ConditionNode
	key string // Blackboard key to check
}

// NewIsBlackboardKeySetCondition creates a new blackboard key check condition
func NewIsBlackboardKeySetCondition(name, key string) *IsBlackboardKeySetCondition {
	return &IsBlackboardKeySetCondition{
		ConditionNode: ConditionNode{
			BaseNode: BaseNode{name: name},
		},
		key: key,
	}
}

// Execute checks if blackboard key exists
func (condition *IsBlackboardKeySetCondition) Execute(context *BehaviorContext) NodeStatus {
	if context.Blackboard.Has(condition.key) {
		return StatusSuccess
	}
	return StatusFailure
}

// IsUnitIdleCondition checks if unit has no active commands
type IsUnitIdleCondition struct {
	ConditionNode
}

// NewIsUnitIdleCondition creates a new idle check condition
func NewIsUnitIdleCondition(name string) *IsUnitIdleCondition {
	return &IsUnitIdleCondition{
		ConditionNode: ConditionNode{
			BaseNode: BaseNode{name: name},
		},
	}
}

// Execute checks if unit is idle
func (condition *IsUnitIdleCondition) Execute(context *BehaviorContext) NodeStatus {
	unit := context.Unit

	if unit.CurrentCommand == nil && unit.State == UnitStateIdle {
		return StatusSuccess
	}
	return StatusFailure
}

// Helper function for distance calculation
func calculateDistance(pos1, pos2 Vector3) float64 {
	dx := pos1.X - pos2.X
	dy := pos1.Y - pos2.Y
	dz := pos1.Z - pos2.Z
	return math.Sqrt(dx*dx + dy*dy + dz*dz)
}