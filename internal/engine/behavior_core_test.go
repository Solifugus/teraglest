package engine

import (
	"testing"
	"time"
)

// TestBasicBehaviorTree tests core behavior tree functionality without complex dependencies
func TestBasicBehaviorTree(t *testing.T) {
	// Create simple test context
	unit := &GameUnit{
		ID:       1,
		PlayerID: 0,
		Position: Vector3{X: 0, Y: 0, Z: 0},
		Health:   100,
		MaxHealth: 100,
		State:    UnitStateIdle,
	}

	world := &World{
		Width:  64,
		Height: 64,
		resources: make(map[int]*ResourceNode),
	}

	context := NewBehaviorContext(unit, world, time.Millisecond*16)

	// Test sequence node
	sequence := NewSequenceNode("TestSequence")
	success1 := &MockAction{name: "Success1", result: StatusSuccess}
	success2 := &MockAction{name: "Success2", result: StatusSuccess}

	sequence.AddChild(success1)
	sequence.AddChild(success2)

	status := sequence.Execute(context)
	if status != StatusSuccess {
		t.Errorf("Expected sequence to succeed, got %v", status)
	}

	// Test selector node
	selector := NewSelectorNode("TestSelector")
	failure := &MockAction{name: "Failure", result: StatusFailure}
	success := &MockAction{name: "Success", result: StatusSuccess}

	selector.AddChild(failure)
	selector.AddChild(success)

	status = selector.Execute(context)
	if status != StatusSuccess {
		t.Errorf("Expected selector to succeed, got %v", status)
	}

	// Test inverter
	inverter := NewInverterNode("TestInverter")
	inverter.AddChild(success)

	status = inverter.Execute(context)
	if status != StatusFailure {
		t.Errorf("Expected inverter to invert success to failure, got %v", status)
	}
}

// TestBehaviorTreeLibraryBasic tests the template library basic functionality
func TestBehaviorTreeLibraryBasic(t *testing.T) {
	library := NewBehaviorTreeLibrary()

	// Check that default templates are loaded
	templates := library.GetAllTemplateNames()
	if len(templates) == 0 {
		t.Error("Expected default templates to be loaded")
	}

	// Test worker template exists
	workerTemplates := library.GetTemplatesForUnitType("worker")
	if len(workerTemplates) == 0 {
		t.Error("Expected worker templates to exist")
	}

	// Test general template exists for any unit type
	generalTemplates := library.GetTemplatesForUnitType("unknown_unit")
	if len(generalTemplates) == 0 {
		t.Error("Expected general template to exist for any unit type")
	}

	// Test creating a tree from template
	tree, err := library.CreateBehaviorTree("worker_ai")
	if err != nil {
		t.Errorf("Failed to create worker AI tree: %v", err)
	}
	if tree == nil {
		t.Error("Expected non-nil tree")
	}
}

// TestBehaviorTreeManagerBasic tests basic manager functionality
func TestBehaviorTreeManagerBasic(t *testing.T) {
	world := &World{
		Width:     64,
		Height:    64,
		resources: make(map[int]*ResourceNode),
	}
	world.ObjectManager = &ObjectManager{
		world: world,
		UnitManager: &UnitManager{
			world: world,
			units: make(map[int]*GameUnit),
		},
	}

	unit := &GameUnit{
		ID:       1,
		PlayerID: 0,
		Health:   100,
		MaxHealth: 100,
		State:    UnitStateIdle,
	}
	world.ObjectManager.UnitManager.units[unit.ID] = unit

	manager := NewBehaviorTreeManager(world)

	// Create simple tree
	root := &MockAction{name: "Root", result: StatusSuccess}
	tree := NewBehaviorTree(root)

	// Test setting behavior tree
	err := manager.SetBehaviorTree(unit.ID, tree)
	if err != nil {
		t.Errorf("Failed to set behavior tree: %v", err)
	}

	// Test tree exists
	if !manager.HasBehaviorTree(unit.ID) {
		t.Error("Expected unit to have behavior tree")
	}

	// Test update (tree should complete and be removed)
	manager.Update(time.Millisecond * 16)

	// Check tree was removed after completion
	if manager.HasBehaviorTree(unit.ID) {
		t.Error("Expected completed tree to be removed")
	}
}

// TestBasicConditions tests basic condition nodes
func TestBasicConditions(t *testing.T) {
	unit := &GameUnit{
		ID:       1,
		Health:   30,
		MaxHealth: 100,
		CarriedResources: map[string]int{"wood": 75},
	}

	world := &World{}
	context := NewBehaviorContext(unit, world, time.Millisecond*16)

	// Test health condition
	healthCondition := NewIsHealthLowCondition("HealthCheck", 0.5)
	status := healthCondition.Execute(context)
	if status != StatusSuccess {
		t.Error("Expected health condition to succeed with 30/100 health and threshold 0.5")
	}

	// Test carrying resources condition
	carryingCondition := NewIsCarryingResourcesCondition("CarryingCheck", 50)
	status = carryingCondition.Execute(context)
	if status != StatusSuccess {
		t.Error("Expected carrying condition to succeed with 75 resources and threshold 50")
	}

	// Test unit idle condition
	idleCondition := NewIsUnitIdleCondition("IdleCheck")
	status = idleCondition.Execute(context)
	if status != StatusSuccess {
		t.Error("Expected idle condition to succeed with idle unit")
	}
}

// TestBasicActions tests basic action nodes
func TestBasicActions(t *testing.T) {
	unit := &GameUnit{
		ID:       1,
		Position: Vector3{X: 0, Y: 0, Z: 0},
	}

	world := &World{}
	context := NewBehaviorContext(unit, world, time.Millisecond*16)

	// Test blackboard action
	setAction := NewSetBlackboardValueAction("SetValue", "test_key", "test_value")
	status := setAction.Execute(context)
	if status != StatusSuccess {
		t.Error("Expected set blackboard action to succeed")
	}

	// Verify value was set
	value, exists := context.Blackboard.GetString("test_key")
	if !exists || value != "test_value" {
		t.Errorf("Expected blackboard value 'test_value', got %v (exists: %v)", value, exists)
	}

	// Test wait action
	waitAction := NewWaitAction("WaitTest", time.Millisecond*1)

	// First execution should start wait
	status = waitAction.Execute(context)
	if status != StatusRunning {
		t.Error("Expected wait action to return Running on first execution")
	}

	// Wait for duration
	time.Sleep(time.Millisecond * 2)

	// Should complete now
	status = waitAction.Execute(context)
	if status != StatusSuccess {
		t.Error("Expected wait action to succeed after duration")
	}
}

// Mock action for testing
type MockAction struct {
	name   string
	result NodeStatus
	execCount int
}

func (m *MockAction) Execute(context *BehaviorContext) NodeStatus {
	m.execCount++
	return m.result
}

func (m *MockAction) Reset() {
	m.execCount = 0
}

func (m *MockAction) GetName() string {
	return m.name
}

func (m *MockAction) GetChildren() []BehaviorNode {
	return []BehaviorNode{}
}

func (m *MockAction) AddChild(child BehaviorNode) error {
	return nil // Leaf node, no children allowed
}