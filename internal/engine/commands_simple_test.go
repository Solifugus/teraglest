package engine

import (
	"testing"
)

// TestCommandCreation tests basic command creation
func TestCommandCreation(t *testing.T) {
	// Test move command creation
	moveCmd := CreateMoveCommand(Vector3{X: 10, Y: 0, Z: 10}, false)
	if moveCmd.Type != CommandMove {
		t.Errorf("Expected CommandMove, got %v", moveCmd.Type)
	}
	if moveCmd.Target.X != 10 {
		t.Errorf("Expected target X=10, got %f", moveCmd.Target.X)
	}

	// Test attack command creation with mock unit
	mockTarget := &GameUnit{ID: 1, UnitType: "archer"}
	attackCmd := CreateAttackCommand(mockTarget, false)
	if attackCmd.Type != CommandAttack {
		t.Errorf("Expected CommandAttack, got %v", attackCmd.Type)
	}
	if attackCmd.TargetUnit != mockTarget {
		t.Error("Expected target unit to match")
	}

	// Test build command creation
	buildCmd := CreateBuildCommand(Vector3{X: 5, Y: 0, Z: 5}, "barracks", false)
	if buildCmd.Type != CommandBuild {
		t.Errorf("Expected CommandBuild, got %v", buildCmd.Type)
	}

	// Test resource commands
	mockResource := &ResourceNode{ID: 1, ResourceType: "wood"}
	gatherCmd := CreateGatherCommand(mockResource, false)
	if gatherCmd.Type != CommandGather {
		t.Errorf("Expected CommandGather, got %v", gatherCmd.Type)
	}

	// Test production commands
	cost := map[string]int{"gold": 100}
	produceCmd := CreateProduceCommand("warrior", cost)
	if produceCmd.Type != CommandProduce {
		t.Errorf("Expected CommandProduce, got %v", produceCmd.Type)
	}

	// Test utility commands
	stopCmd := CreateStopCommand()
	if stopCmd.Type != CommandStop {
		t.Errorf("Expected CommandStop, got %v", stopCmd.Type)
	}

	patrolCmd := CreatePatrolCommand(Vector3{X: 15, Y: 0, Z: 15}, false)
	if patrolCmd.Type != CommandPatrol {
		t.Errorf("Expected CommandPatrol, got %v", patrolCmd.Type)
	}
}

// TestCommandPriority tests command priority system
func TestCommandPriority(t *testing.T) {
	// Test priority constants
	if PriorityLow >= PriorityNormal {
		t.Error("PriorityLow should be less than PriorityNormal")
	}
	if PriorityNormal >= PriorityHigh {
		t.Error("PriorityNormal should be less than PriorityHigh")
	}
	if PriorityHigh >= PriorityCritical {
		t.Error("PriorityHigh should be less than PriorityCritical")
	}

	// Test sorting
	commands := []UnitCommand{
		{Type: CommandMove, Priority: PriorityLow},
		{Type: CommandAttack, Priority: PriorityCritical},
		{Type: CommandStop, Priority: PriorityNormal},
		{Type: CommandBuild, Priority: PriorityHigh},
	}

	SortCommandsByPriority(commands)

	// Should be sorted: Critical, High, Normal, Low
	expectedPriorities := []int{PriorityCritical, PriorityHigh, PriorityNormal, PriorityLow}
	for i, cmd := range commands {
		if cmd.Priority != expectedPriorities[i] {
			t.Errorf("Command %d: expected priority %d, got %d", i, expectedPriorities[i], cmd.Priority)
		}
	}
}

// TestCommandProcessorBasic tests basic CommandProcessor functionality
func TestCommandProcessorBasic(t *testing.T) {
	// Create minimal world for testing
	world := &World{}
	processor := NewCommandProcessor(world)

	if processor == nil {
		t.Error("CommandProcessor should not be nil")
	}

	if processor.world != world {
		t.Error("CommandProcessor should reference the world")
	}

	// Test that processor has required components
	if processor.combatSystem == nil {
		t.Error("CommandProcessor should have combat system")
	}

	if processor.statusEffectMgr == nil {
		t.Error("CommandProcessor should have status effect manager")
	}

	if processor.visualSystem == nil {
		t.Error("CommandProcessor should have visual system")
	}
}

// TestAdvancedCombatIntegration tests the integration with advanced combat
func TestAdvancedCombatIntegration(t *testing.T) {
	world := &World{}
	processor := NewCommandProcessor(world)

	// Test that advanced combat system is available
	if processor.combatSystem == nil {
		t.Fatal("Advanced combat system should be available")
	}

	// Test status effect manager is set up
	if processor.statusEffectMgr == nil {
		t.Fatal("Status effect manager should be available")
	}

	// Test unit advanced damage type mapping
	testUnit := &GameUnit{UnitType: "catapult"}
	damageType := processor.getUnitAdvancedDamageType(testUnit)

	if damageType.Name != "catapult" {
		t.Errorf("Expected catapult damage type, got %s", damageType.Name)
	}

	// Test ranged attack detection
	if !processor.isRangedAttack(damageType) {
		t.Error("Catapult should be detected as ranged attack")
	}

	// Test melee unit
	meleeUnit := &GameUnit{UnitType: "warrior"}
	meleeDamageType := processor.getUnitAdvancedDamageType(meleeUnit)

	if processor.isRangedAttack(meleeDamageType) {
		t.Error("Warrior should not be detected as ranged attack")
	}
}

// TestCommandStrings tests command type string representation
func TestCommandStrings(t *testing.T) {
	tests := []struct {
		cmdType  CommandType
		expected string
	}{
		{CommandMove, "Move"},
		{CommandAttack, "Attack"},
		{CommandGather, "Gather"},
		{CommandBuild, "Build"},
		{CommandRepair, "Repair"},
		{CommandStop, "Stop"},
		{CommandHold, "Hold"},
		{CommandPatrol, "Patrol"},
		{CommandFollow, "Follow"},
		{CommandGuard, "Guard"},
		{CommandProduce, "Produce"},
		{CommandUpgrade, "Upgrade"},
	}

	for _, test := range tests {
		result := test.cmdType.String()
		if result != test.expected {
			t.Errorf("CommandType %d: expected %s, got %s", test.cmdType, test.expected, result)
		}
	}
}