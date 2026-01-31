package engine

import (
	"testing"
	"time"

	"teraglest/internal/data"
)

// createSimpleTestWorld creates a minimal test world
func createSimpleTestWorld() *World {
	techTree := &data.TechTree{}
	assetMgr := &data.AssetManager{}
	settings := GameSettings{
		MaxPlayers: 4,
		GameSpeed:  1.0,
	}

	world, err := NewWorld(settings, techTree, assetMgr)
	if err != nil {
		panic("Failed to create test world")
	}

	err = world.Initialize()
	if err != nil {
		panic("Failed to initialize test world")
	}

	return world
}

// TestFormationSystemIntegration tests basic formation system integration
func TestFormationSystemIntegration(t *testing.T) {
	world := createSimpleTestWorld()

	// Verify all formation components are initialized
	if world.groupMgr == nil {
		t.Fatal("GroupManager should be initialized")
	}

	if world.commandProcessor == nil {
		t.Fatal("CommandProcessor should be initialized")
	}

	// Test that world updates don't crash with formation system
	world.Update(time.Millisecond * 16)

	// Test formation manager stats (should be empty initially)
	stats := world.groupMgr.GetGroupStats()
	totalGroups, ok := stats["total_groups"].(int)
	if !ok || totalGroups != 0 {
		t.Errorf("Expected 0 groups initially, got %v", stats["total_groups"])
	}
}

// TestFormationSystemWithMockUnits tests formations with manually created units
func TestFormationSystemWithMockUnits(t *testing.T) {
	world := createSimpleTestWorld()

	// Create mock units directly (avoid ObjectManager complexity)
	units := make([]*GameUnit, 3)
	for i := 0; i < 3; i++ {
		units[i] = &GameUnit{
			ID:           i + 1,
			PlayerID:     0,
			UnitType:     "soldier",
			Position:     Vector3{X: float64(i * 2), Y: 0, Z: 0},
			Health:       100,
			MaxHealth:    100,
			State:        UnitStateIdle,
			CreationTime: time.Now(),
			LastUpdate:   time.Now(),
		}
	}

	// Create group using GroupManager directly
	group, err := world.groupMgr.CreateGroup(0, units, FormationLine)
	if err != nil {
		t.Fatalf("Failed to create group: %v", err)
	}

	if group == nil {
		t.Fatal("Group should not be nil")
	}

	if group.GetUnitCount() != 3 {
		t.Errorf("Expected 3 units in group, got %d", group.GetUnitCount())
	}

	// Test formation positioning
	for _, unit := range units {
		pos, exists := group.GetFormationPosition(unit.ID)
		if !exists {
			t.Errorf("No formation position for unit %d", unit.ID)
		}

		// Position should be different from unit's current position
		// (formation positions are relative to group center)
		if pos.X == unit.Position.X && pos.Z == unit.Position.Z {
			t.Logf("Unit %d formation pos: (%f, %f), current pos: (%f, %f)",
				unit.ID, pos.X, pos.Z, unit.Position.X, unit.Position.Z)
		}
	}

	// Test group movement
	target := Vector3{X: 20, Y: 0, Z: 20}
	err = world.groupMgr.MoveGroup(group.ID, target)
	if err != nil {
		t.Errorf("Failed to move group: %v", err)
	}

	if !group.IsMoving {
		t.Error("Group should be marked as moving")
	}

	// Test world update with formation
	world.Update(time.Millisecond * 16)

	// Group should still exist after update
	retrievedGroup, exists := world.groupMgr.GetGroup(group.ID)
	if !exists {
		t.Error("Group should still exist after world update")
	}

	if retrievedGroup.ID != group.ID {
		t.Errorf("Retrieved group ID %d doesn't match original %d",
			retrievedGroup.ID, group.ID)
	}
}

// TestGroupCommands tests group command processing
func TestGroupCommands(t *testing.T) {
	world := createSimpleTestWorld()

	// Create mock units
	units := make([]*GameUnit, 2)
	for i := 0; i < 2; i++ {
		units[i] = &GameUnit{
			ID:       i + 1,
			PlayerID: 0,
			UnitType: "soldier",
			Position: Vector3{X: float64(i), Y: 0, Z: 0},
			Health:   100,
			State:    UnitStateIdle,
		}
	}

	// For this test, we'll create the group manually first since the command
	// processor needs actual units in the ObjectManager
	group, err := world.groupMgr.CreateGroup(0, units, FormationWedge)
	if err != nil {
		t.Fatalf("Failed to create group manually: %v", err)
	}

	// Test formation change command
	formationCommand := CreateSetFormationCommand(group.ID, FormationCircle)
	err = world.commandProcessor.IssueGroupCommand(0, formationCommand)
	if err != nil {
		t.Errorf("Failed to issue formation command: %v", err)
	}

	// Verify formation changed
	if group.Formation != FormationCircle {
		t.Errorf("Expected FormationCircle, got %v", group.Formation)
	}

	// Test group move command
	moveCommand := CreateGroupMoveCommand(group.ID, Vector3{X: 10, Y: 0, Z: 10})
	err = world.commandProcessor.IssueGroupCommand(0, moveCommand)
	if err != nil {
		t.Errorf("Failed to issue group move command: %v", err)
	}

	// Verify group is marked as moving
	if !group.IsMoving {
		t.Error("Group should be moving after move command")
	}

	// Test disband command
	disbandCommand := CreateDisbandGroupCommand(group.ID)
	err = world.commandProcessor.IssueGroupCommand(0, disbandCommand)
	if err != nil {
		t.Errorf("Failed to issue disband command: %v", err)
	}

	// Verify group no longer exists
	_, exists := world.groupMgr.GetGroup(group.ID)
	if exists {
		t.Error("Group should not exist after disband command")
	}
}

// TestFormationTypes tests all formation types work correctly
func TestFormationTypes(t *testing.T) {
	world := createSimpleTestWorld()

	formations := []FormationType{
		FormationLine,
		FormationColumn,
		FormationWedge,
		FormationBox,
		FormationCircle,
		FormationScatter,
	}

	for _, formation := range formations {
		t.Run(formation.String(), func(t *testing.T) {
			// Create units for this formation test
			units := make([]*GameUnit, 4)
			for i := 0; i < 4; i++ {
				units[i] = &GameUnit{
					ID:       i + 100 + int(formation)*10, // Unique IDs
					PlayerID: 0,
					UnitType: "soldier",
					Position: Vector3{X: float64(i), Y: 0, Z: 0},
					Health:   100,
					State:    UnitStateIdle,
				}
			}

			// Create group with this formation
			group, err := world.groupMgr.CreateGroup(0, units, formation)
			if err != nil {
				t.Fatalf("Failed to create group with %s formation: %v", formation.String(), err)
			}

			if group.Formation != formation {
				t.Errorf("Expected %s formation, got %v", formation.String(), group.Formation)
			}

			// Verify positions were generated
			if len(group.Positions) != 4 {
				t.Errorf("Expected 4 formation positions for %s, got %d",
					formation.String(), len(group.Positions))
			}

			// All units should have positions
			for _, unit := range units {
				if _, exists := group.Positions[unit.ID]; !exists {
					t.Errorf("No position generated for unit %d in %s formation",
						unit.ID, formation.String())
				}
			}

			// Clean up - disband the group
			world.groupMgr.DisbandGroup(group.ID)
		})
	}
}

// TestFormationParameterModification tests parameter modification
func TestFormationParameterModification(t *testing.T) {
	world := createSimpleTestWorld()

	units := make([]*GameUnit, 3)
	for i := 0; i < 3; i++ {
		units[i] = &GameUnit{
			ID:       i + 1,
			PlayerID: 0,
			UnitType: "soldier",
			Position: Vector3{X: float64(i), Y: 0, Z: 0},
			Health:   100,
			State:    UnitStateIdle,
		}
	}

	group, err := world.groupMgr.CreateGroup(0, units, FormationLine)
	if err != nil {
		t.Fatalf("Failed to create group: %v", err)
	}

	// Test default parameters
	if group.Parameters.UnitSpacing <= 0 {
		t.Error("Default unit spacing should be positive")
	}

	// Modify parameters
	group.Parameters.UnitSpacing = 5.0
	group.Parameters.MaxSpeed = 12.0

	// Regenerate formation with new parameters
	group.SetFormation(FormationLine)

	// Parameters should be applied (this is tested more thoroughly in formation tests)
	if group.Parameters.UnitSpacing != 5.0 {
		t.Errorf("Expected unit spacing 5.0, got %f", group.Parameters.UnitSpacing)
	}

	if group.Parameters.MaxSpeed != 12.0 {
		t.Errorf("Expected max speed 12.0, got %f", group.Parameters.MaxSpeed)
	}
}

// TestFormationWorldUpdates tests formation behavior during world updates
func TestFormationWorldUpdates(t *testing.T) {
	world := createSimpleTestWorld()

	units := make([]*GameUnit, 2)
	for i := 0; i < 2; i++ {
		units[i] = &GameUnit{
			ID:       i + 1,
			PlayerID: 0,
			UnitType: "soldier",
			Position: Vector3{X: float64(i), Y: 0, Z: 0},
			Health:   100,
			State:    UnitStateIdle,
		}
	}

	group, err := world.groupMgr.CreateGroup(0, units, FormationLine)
	if err != nil {
		t.Fatalf("Failed to create group: %v", err)
	}

	initialUpdateTime := group.LastUpdate

	// Multiple world updates should update the group
	for i := 0; i < 5; i++ {
		world.Update(time.Millisecond * 16)
	}

	// Group should have been updated
	if !group.LastUpdate.After(initialUpdateTime) {
		t.Error("Group should have been updated during world updates")
	}

	// Group should still exist
	if group.IsEmpty() {
		t.Error("Group should not be empty after updates")
	}

	// Stats should reflect the group
	stats := world.groupMgr.GetGroupStats()
	totalGroups, ok := stats["total_groups"].(int)
	if !ok || totalGroups != 1 {
		t.Errorf("Expected 1 group in stats, got %v", stats["total_groups"])
	}
}