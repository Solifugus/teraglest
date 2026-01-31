package engine

import (
	"testing"
	"time"

	"teraglest/internal/data"
)

// TestBlackboard tests the blackboard functionality
func TestBlackboard(t *testing.T) {
	bb := NewBlackboard()

	// Test basic set/get
	bb.Set("test_key", "test_value")
	value, exists := bb.Get("test_key")
	if !exists {
		t.Error("Expected key to exist in blackboard")
	}
	if value != "test_value" {
		t.Errorf("Expected 'test_value', got %v", value)
	}

	// Test type-specific getters
	bb.Set("string_key", "hello")
	bb.Set("int_key", 42)
	bb.Set("float_key", 3.14)
	bb.Set("bool_key", true)

	str, ok := bb.GetString("string_key")
	if !ok || str != "hello" {
		t.Errorf("GetString failed: got %v, %v", str, ok)
	}

	intVal, ok := bb.GetInt("int_key")
	if !ok || intVal != 42 {
		t.Errorf("GetInt failed: got %v, %v", intVal, ok)
	}

	floatVal, ok := bb.GetFloat("float_key")
	if !ok || floatVal != 3.14 {
		t.Errorf("GetFloat failed: got %v, %v", floatVal, ok)
	}

	boolVal, ok := bb.GetBool("bool_key")
	if !ok || !boolVal {
		t.Errorf("GetBool failed: got %v, %v", boolVal, ok)
	}

	// Test has and remove
	if !bb.Has("test_key") {
		t.Error("Has should return true for existing key")
	}

	bb.Remove("test_key")
	if bb.Has("test_key") {
		t.Error("Has should return false after removal")
	}

	// Test clear
	bb.Clear()
	if bb.Has("string_key") {
		t.Error("Blackboard should be empty after clear")
	}
}

// TestSequenceNode tests sequence node behavior
func TestSequenceNode(t *testing.T) {
	world, unit := createTestUnitAndWorld(t)
	context := NewBehaviorContext(unit, world, time.Millisecond*16)

	sequence := NewSequenceNode("TestSequence")

	// Add some test actions
	success1 := &MockSuccessAction{BaseNode: BaseNode{name: "Success1"}}
	success2 := &MockSuccessAction{BaseNode: BaseNode{name: "Success2"}}
	failure1 := &MockFailureAction{BaseNode: BaseNode{name: "Failure1"}}

	// Test successful sequence
	sequence.AddChild(success1)
	sequence.AddChild(success2)

	status := sequence.Execute(context)
	if status != StatusSuccess {
		t.Errorf("Expected StatusSuccess, got %v", status)
	}

	// Reset and test sequence with failure
	sequence.Reset()
	sequence.AddChild(failure1)

	status = sequence.Execute(context)
	if status != StatusFailure {
		t.Errorf("Expected StatusFailure, got %v", status)
	}
}

// TestSelectorNode tests selector node behavior
func TestSelectorNode(t *testing.T) {
	world, unit := createTestUnitAndWorld(t)
	context := NewBehaviorContext(unit, world, time.Millisecond*16)

	selector := NewSelectorNode("TestSelector")

	failure1 := &MockFailureAction{BaseNode: BaseNode{name: "Failure1"}}
	failure2 := &MockFailureAction{BaseNode: BaseNode{name: "Failure2"}}
	success1 := &MockSuccessAction{BaseNode: BaseNode{name: "Success1"}}

	// Test selector with eventual success
	selector.AddChild(failure1)
	selector.AddChild(failure2)
	selector.AddChild(success1)

	status := selector.Execute(context)
	if status != StatusSuccess {
		t.Errorf("Expected StatusSuccess, got %v", status)
	}

	// Test selector with all failures
	selector2 := NewSelectorNode("TestSelector2")
	selector2.AddChild(failure1)
	selector2.AddChild(failure2)

	status = selector2.Execute(context)
	if status != StatusFailure {
		t.Errorf("Expected StatusFailure, got %v", status)
	}
}

// TestInverterNode tests inverter decorator behavior
func TestInverterNode(t *testing.T) {
	world, unit := createTestUnitAndWorld(t)
	context := NewBehaviorContext(unit, world, time.Millisecond*16)

	inverter := NewInverterNode("TestInverter")

	// Test inverting success to failure
	success := &MockSuccessAction{BaseNode: BaseNode{name: "Success"}}
	inverter.AddChild(success)

	status := inverter.Execute(context)
	if status != StatusFailure {
		t.Errorf("Expected StatusFailure (inverted success), got %v", status)
	}

	// Test inverting failure to success
	inverter2 := NewInverterNode("TestInverter2")
	failure := &MockFailureAction{BaseNode: BaseNode{name: "Failure"}}
	inverter2.AddChild(failure)

	status = inverter2.Execute(context)
	if status != StatusSuccess {
		t.Errorf("Expected StatusSuccess (inverted failure), got %v", status)
	}
}

// TestRepeaterNode tests repeater decorator behavior
func TestRepeaterNode(t *testing.T) {
	world, unit := createTestUnitAndWorld(t)
	context := NewBehaviorContext(unit, world, time.Millisecond*16)

	// Test repeater with limited repetitions
	repeater := NewRepeaterNode("TestRepeater", 3)
	success := &MockSuccessAction{BaseNode: BaseNode{name: "Success"}}
	repeater.AddChild(success)

	// Should run 3 times then succeed
	var status NodeStatus
	for i := 0; i < 3; i++ {
		status = repeater.Execute(context)
	}

	if status != StatusSuccess {
		t.Errorf("Expected StatusSuccess after 3 repetitions, got %v", status)
	}

	// Test repeater with failure
	repeater2 := NewRepeaterNode("TestRepeater2", 5)
	failure := &MockFailureAction{BaseNode: BaseNode{name: "Failure"}}
	repeater2.AddChild(failure)

	status = repeater2.Execute(context)
	if status != StatusFailure {
		t.Errorf("Expected StatusFailure on first failure, got %v", status)
	}
}

// TestMoveToPositionAction tests the move action
func TestMoveToPositionAction(t *testing.T) {
	world, unit := createTestUnitAndWorld(t)
	context := NewBehaviorContext(unit, world, time.Millisecond*16)

	// Set target position in blackboard
	targetPos := Vector3{X: 10, Y: 0, Z: 10}
	context.Blackboard.Set("target_pos", targetPos)

	moveAction := NewMoveToPositionAction("MoveToTarget", "target_pos", 1.0)

	// First execution should issue move command
	status := moveAction.Execute(context)
	if status != StatusRunning {
		t.Errorf("Expected StatusRunning for first execution, got %v", status)
	}

	// Check if unit has received move command
	if unit.CurrentCommand == nil || unit.CurrentCommand.Type != CommandMove {
		t.Error("Expected unit to have move command after action execution")
	}

	// Simulate unit reaching target
	unit.Position = targetPos
	unit.CurrentCommand = nil

	status = moveAction.Execute(context)
	if status != StatusSuccess {
		t.Errorf("Expected StatusSuccess when at target, got %v", status)
	}
}

// TestIsHealthLowCondition tests the health condition
func TestIsHealthLowCondition(t *testing.T) {
	world, unit := createTestUnitAndWorld(t)
	context := NewBehaviorContext(unit, world, time.Millisecond*16)

	condition := NewIsHealthLowCondition("HealthCheck", 0.5)

	// Test with high health
	unit.Health = 80
	unit.MaxHealth = 100

	status := condition.Execute(context)
	if status != StatusFailure {
		t.Errorf("Expected StatusFailure with high health, got %v", status)
	}

	// Test with low health
	unit.Health = 30
	status = condition.Execute(context)
	if status != StatusSuccess {
		t.Errorf("Expected StatusSuccess with low health, got %v", status)
	}
}

// TestBehaviorTree tests the complete behavior tree functionality
func TestBehaviorTree(t *testing.T) {
	world, unit := createTestUnitAndWorld(t)

	// Create a simple behavior tree
	root := NewSequenceNode("RootSequence")
	success1 := &MockSuccessAction{BaseNode: BaseNode{name: "Success1"}}
	success2 := &MockSuccessAction{BaseNode: BaseNode{name: "Success2"}}

	root.AddChild(success1)
	root.AddChild(success2)

	tree := NewBehaviorTree(root)

	// Test tree lifecycle
	if tree.IsActive() {
		t.Error("Tree should not be active before start")
	}

	tree.Start(unit, world)
	if !tree.IsActive() {
		t.Error("Tree should be active after start")
	}

	// Update tree
	status := tree.Update(time.Millisecond * 16)
	if status != StatusSuccess {
		t.Errorf("Expected StatusSuccess, got %v", status)
	}

	// Tree should stop after completion
	if tree.IsActive() {
		t.Error("Tree should not be active after successful completion")
	}
}

// TestBehaviorTreeManager tests the behavior tree manager
func TestBehaviorTreeManager(t *testing.T) {
	world, unit := createTestUnitAndWorld(t)

	btManager := NewBehaviorTreeManager(world)

	// Create and assign behavior tree
	root := &MockSuccessAction{BaseNode: BaseNode{name: "RootAction"}}
	tree := NewBehaviorTree(root)

	err := btManager.SetBehaviorTree(unit.ID, tree)
	if err != nil {
		t.Errorf("Failed to set behavior tree: %v", err)
	}

	// Check tree exists
	if !btManager.HasBehaviorTree(unit.ID) {
		t.Error("Expected unit to have behavior tree")
	}

	// Update manager
	btManager.Update(time.Millisecond * 16)

	// Tree should complete and be removed
	if btManager.HasBehaviorTree(unit.ID) {
		t.Error("Expected completed tree to be removed")
	}
}

// TestBehaviorTreeLibrary tests the template library
func TestBehaviorTreeLibrary(t *testing.T) {
	library := NewBehaviorTreeLibrary()

	// Test getting default templates
	templates := library.GetAllTemplateNames()
	if len(templates) == 0 {
		t.Error("Expected default templates to be registered")
	}

	// Test getting template for unit type
	workerTemplates := library.GetTemplatesForUnitType("worker")
	if len(workerTemplates) == 0 {
		t.Error("Expected to find templates for worker unit type")
	}

	// Test creating tree from template
	tree, err := library.CreateBehaviorTree("worker_ai")
	if err != nil {
		t.Errorf("Failed to create tree from template: %v", err)
	}
	if tree == nil {
		t.Error("Expected non-nil behavior tree")
	}
}

// TestBehaviorTreeFactory tests the behavior tree factory
func TestBehaviorTreeFactory(t *testing.T) {
	world, unit := createTestUnitAndWorld(t)
	btManager := NewBehaviorTreeManager(world)

	factory := NewBehaviorTreeFactory(world)

	// Test creating tree for unit
	unit.UnitType = "worker"
	err := factory.SetupUnitBehavior(unit, btManager)
	if err != nil {
		t.Errorf("Failed to setup unit behavior: %v", err)
	}

	// Check that unit has behavior tree
	if !btManager.HasBehaviorTree(unit.ID) {
		t.Error("Expected unit to have behavior tree after setup")
	}

	// Test creating tree by specific template
	unit2 := &GameUnit{
		ID:       2,
		PlayerID: 0,
		UnitType: "soldier",
		Position: Vector3{X: 5, Y: 0, Z: 5},
		Health:   100,
		MaxHealth: 100,
		State:    UnitStateIdle,
	}

	err = factory.SetupUnitBehaviorByTemplate(unit2, "soldier_ai", btManager)
	if err != nil {
		t.Errorf("Failed to setup unit behavior by template: %v", err)
	}

	if !btManager.HasBehaviorTree(unit2.ID) {
		t.Error("Expected unit2 to have behavior tree after template setup")
	}
}

// TestWaitAction tests the wait action
func TestWaitAction(t *testing.T) {
	world, unit := createTestUnitAndWorld(t)
	context := NewBehaviorContext(unit, world, time.Millisecond*16)

	waitAction := NewWaitAction("TestWait", time.Millisecond*50)

	// First execution should start waiting
	status := waitAction.Execute(context)
	if status != StatusRunning {
		t.Errorf("Expected StatusRunning for first execution, got %v", status)
	}

	// Should still be running before duration
	status = waitAction.Execute(context)
	if status != StatusRunning {
		t.Errorf("Expected StatusRunning before duration, got %v", status)
	}

	// Wait for duration to pass
	time.Sleep(time.Millisecond * 60)

	status = waitAction.Execute(context)
	if status != StatusSuccess {
		t.Errorf("Expected StatusSuccess after duration, got %v", status)
	}
}

// Mock action nodes for testing

type MockSuccessAction struct {
	BaseNode
}

func (action *MockSuccessAction) Execute(context *BehaviorContext) NodeStatus {
	return StatusSuccess
}

type MockFailureAction struct {
	BaseNode
}

func (action *MockFailureAction) Execute(context *BehaviorContext) NodeStatus {
	return StatusFailure
}

type MockRunningAction struct {
	BaseNode
	executionCount int
	maxExecutions  int
}

func (action *MockRunningAction) Execute(context *BehaviorContext) NodeStatus {
	action.executionCount++
	if action.executionCount >= action.maxExecutions {
		return StatusSuccess
	}
	return StatusRunning
}

func (action *MockRunningAction) Reset() {
	action.executionCount = 0
}

// Helper function to create test unit and world
func createTestUnitAndWorld(t *testing.T) (*World, *GameUnit) {
	// Create minimal world for testing
	settings := GameSettings{
		PlayerFactions: map[int]string{0: "tech"},
		AIFactions:     map[int]string{},
	}

	techTree := &data.TechTree{}
	assetMgr := data.NewAssetManager("")

	world := &World{
		settings:      settings,
		techTree:      techTree,
		assetMgr:      assetMgr,
		players:       make(map[int]*Player),
		resources:     make(map[int]*ResourceNode),
		nextEntityID:  1,
		Width:         64,
		Height:        64,
		tileSize:      1.0,
		resourceGenerationRate: make(map[string]float32),
		unitCap:       200,
		buildingCap:   50,
	}

	// Initialize required systems
	world.ObjectManager = NewObjectManager(world)
	world.commandProcessor = NewCommandProcessor(world)
	world.pathfindingMgr = NewPathfindingManager(world)
	world.behaviorTreeMgr = NewBehaviorTreeManager(world)

	// Initialize grid system
	err := world.initializeGrid()
	if err != nil {
		t.Fatalf("Failed to initialize grid system: %v", err)
	}

	// Create test player
	player := &Player{
		ID:          0,
		Name:        "TestPlayer",
		FactionName: "tech",
		Resources:   make(map[string]int),
		IsActive:    true,
	}
	world.players[0] = player

	// Create test unit
	unit := &GameUnit{
		ID:       1,
		PlayerID: 0,
		UnitType: "worker",
		Position: Vector3{X: 0, Y: 0, Z: 0},
		Health:   100,
		MaxHealth: 100,
		State:    UnitStateIdle,
		Speed:    2.0,
		CarriedResources: make(map[string]int),
	}

	return world, unit
}

// TestIntegrationWithCommandSystem tests behavior trees working with commands
func TestIntegrationWithCommandSystem(t *testing.T) {
	world, unit := createTestUnitAndWorld(t)
	btManager := world.behaviorTreeMgr

	// Add unit to object manager through the proper interface
	_, err := world.ObjectManager.CreateUnit(unit.PlayerID, unit.UnitType, unit.Position, nil)
	if err != nil {
		t.Fatalf("Failed to create unit: %v", err)
	}

	// Create a simple behavior tree that issues a move command
	root := NewSequenceNode("MoveSequence")

	// Set target in blackboard manually for test
	setTargetAction := NewSetBlackboardValueAction("SetTarget", "move_target", Vector3{X: 10, Y: 0, Z: 10})
	moveAction := NewMoveToPositionAction("MoveToTarget", "move_target", 1.0)

	root.AddChild(setTargetAction)
	root.AddChild(moveAction)

	tree := NewBehaviorTree(root)
	btManager.SetBehaviorTree(unit.ID, tree)

	// Update the behavior tree manager
	deltaTime := time.Millisecond * 16
	btManager.Update(deltaTime)

	// Check if unit received move command
	if unit.CurrentCommand == nil {
		t.Error("Expected unit to have received a move command from behavior tree")
	}

	if unit.CurrentCommand.Type != CommandMove {
		t.Errorf("Expected move command, got %v", unit.CurrentCommand.Type)
	}

	// Verify target position
	if unit.CurrentCommand.Target == nil {
		t.Error("Expected move command to have target position")
	} else {
		expected := Vector3{X: 10, Y: 0, Z: 10}
		if *unit.CurrentCommand.Target != expected {
			t.Errorf("Expected target %v, got %v", expected, *unit.CurrentCommand.Target)
		}
	}
}

// Benchmark behavior tree performance
func BenchmarkBehaviorTreeExecution(b *testing.B) {
	world, unit := createTestUnitAndWorld(nil)
	context := NewBehaviorContext(unit, world, time.Millisecond*16)

	// Create complex behavior tree
	root := NewSelectorNode("Root")

	// Combat branch
	combat := NewSequenceNode("Combat")
	findEnemy := NewIsEnemyInRangeCondition("FindEnemy", 10.0, "enemy")
	attack := NewAttackTargetAction("Attack", "enemy")
	combat.AddChild(findEnemy)
	combat.AddChild(attack)

	// Work branch
	work := NewSequenceNode("Work")
	findResource := NewIsResourceInRangeCondition("FindResource", 15.0, "resource")
	gather := NewGatherResourceAction("Gather", "resource")
	work.AddChild(findResource)
	work.AddChild(gather)

	// Idle branch
	idle := NewWaitAction("Idle", time.Second)

	root.AddChild(combat)
	root.AddChild(work)
	root.AddChild(idle)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		root.Execute(context)
		root.Reset()
	}
}