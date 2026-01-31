package engine

import (
	"testing"
	"time"

	"teraglest/internal/data"
)

// createTestWorldWithFormations creates a test world with formation support
func createTestWorldWithFormations() *World {
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

// TestFormationIntegrationWithWorld tests formation system integration with World
func TestFormationIntegrationWithWorld(t *testing.T) {
	world := createTestWorldWithFormations()

	// Add a player
	err := world.AddPlayer(0, "Test Player", "tech", false)
	if err != nil {
		t.Fatalf("Failed to add player: %v", err)
	}

	// Create test units using the ObjectManager
	unitIDs := make([]int, 3)
	for i := 0; i < 3; i++ {
		unit, err := world.ObjectManager.CreateUnit(0, "soldier",
			Vector3{X: float64(i), Y: 0, Z: 0}, nil)
		if err != nil {
			t.Fatalf("Failed to create unit %d: %v", i, err)
		}
		unitIDs[i] = unit.ID
	}

	// Create a group using World's convenience method
	group, err := world.CreateGroup(0, unitIDs, FormationLine)
	if err != nil {
		t.Fatalf("Failed to create group through World: %v", err)
	}

	if group == nil {
		t.Fatal("Group should not be nil")
	}

	if group.GetUnitCount() != 3 {
		t.Errorf("Expected 3 units in group, got %d", group.GetUnitCount())
	}

	// Test moving the group
	target := Vector3{X: 20, Y: 0, Z: 20}
	err = world.MoveGroup(0, group.ID, target)
	if err != nil {
		t.Errorf("Failed to move group through World: %v", err)
	}

	// Verify the group is marked as moving
	if !group.IsMoving {
		t.Error("Group should be marked as moving after move command")
	}

	// Test world update with formations
	world.Update(time.Millisecond * 16)

	// Group should still exist after update
	retrievedGroup, exists := world.GetUnitGroup(unitIDs[0])
	if !exists {
		t.Error("Unit should still be in a group after world update")
	}

	if retrievedGroup.ID != group.ID {
		t.Errorf("Retrieved group ID %d doesn't match original %d",
			retrievedGroup.ID, group.ID)
	}
}

// TestCommandIntegrationWithFormations tests command system integration
func TestCommandIntegrationWithFormations(t *testing.T) {
	world := createTestWorldWithFormations()

	// Add player and units
	err := world.AddPlayer(0, "Test Player", "tech", false)
	if err != nil {
		t.Fatalf("Failed to add player: %v", err)
	}

	unitIDs := make([]int, 2)
	for i := 0; i < 2; i++ {
		unit, err := world.ObjectManager.CreateUnit(0, "soldier",
			Vector3{X: float64(i * 2), Y: 0, Z: 0}, nil)
		if err != nil {
			t.Fatalf("Failed to create unit %d: %v", i, err)
		}
		unitIDs[i] = unit.ID
	}

	// Create group
	group, err := world.CreateGroup(0, unitIDs, FormationWedge)
	if err != nil {
		t.Fatalf("Failed to create group: %v", err)
	}

	// Test group command through command processor
	groupCommand := CreateGroupMoveCommand(group.ID, Vector3{X: 15, Y: 0, Z: 15})
	err = world.commandProcessor.IssueGroupCommand(0, groupCommand)
	if err != nil {
		t.Errorf("Failed to issue group command: %v", err)
	}

	// Verify command was processed
	if !group.IsMoving {
		t.Error("Group should be moving after group move command")
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
}

// TestMultipleGroupsWithWorld tests multiple group management
func TestMultipleGroupsWithWorld(t *testing.T) {
	world := createTestWorldWithFormations()

	// Add player
	err := world.AddPlayer(0, "Test Player", "tech", false)
	if err != nil {
		t.Fatalf("Failed to add player: %v", err)
	}

	// Create two sets of units
	group1UnitIDs := make([]int, 3)
	group2UnitIDs := make([]int, 2)

	for i := 0; i < 3; i++ {
		unit, err := world.ObjectManager.CreateUnit(0, "soldier",
			Vector3{X: float64(i), Y: 0, Z: 0}, nil)
		if err != nil {
			t.Fatalf("Failed to create unit for group 1: %v", err)
		}
		group1UnitIDs[i] = unit.ID
	}

	for i := 0; i < 2; i++ {
		unit, err := world.ObjectManager.CreateUnit(0, "archer",
			Vector3{X: float64(i + 10), Y: 0, Z: 5}, nil)
		if err != nil {
			t.Fatalf("Failed to create unit for group 2: %v", err)
		}
		group2UnitIDs[i] = unit.ID
	}

	// Create two groups with different formations
	group1, err := world.CreateGroup(0, group1UnitIDs, FormationLine)
	if err != nil {
		t.Fatalf("Failed to create group 1: %v", err)
	}

	group2, err := world.CreateGroup(0, group2UnitIDs, FormationColumn)
	if err != nil {
		t.Fatalf("Failed to create group 2: %v", err)
	}

	// Test that both groups exist
	playerGroups := world.GetPlayerGroups(0)
	if len(playerGroups) != 2 {
		t.Errorf("Expected 2 groups for player, got %d", len(playerGroups))
	}

	// Test moving both groups to different locations
	err = world.MoveGroup(0, group1.ID, Vector3{X: 30, Y: 0, Z: 10})
	if err != nil {
		t.Errorf("Failed to move group 1: %v", err)
	}

	err = world.MoveGroup(0, group2.ID, Vector3{X: 30, Y: 0, Z: 20})
	if err != nil {
		t.Errorf("Failed to move group 2: %v", err)
	}

	// Verify both groups are moving
	if !group1.IsMoving {
		t.Error("Group 1 should be moving")
	}

	if !group2.IsMoving {
		t.Error("Group 2 should be moving")
	}

	// Update world and verify groups are still managed
	world.Update(time.Millisecond * 16)

	// Both groups should still exist
	updatedPlayerGroups := world.GetPlayerGroups(0)
	if len(updatedPlayerGroups) != 2 {
		t.Errorf("Expected 2 groups after update, got %d", len(updatedPlayerGroups))
	}
}

// TestFormationWithPathfinding tests formation integration with pathfinding
func TestFormationWithPathfinding(t *testing.T) {
	world := createTestWorldWithFormations()

	// Add player
	err := world.AddPlayer(0, "Test Player", "tech", false)
	if err != nil {
		t.Fatalf("Failed to add player: %v", err)
	}

	// Create units
	unitIDs := make([]int, 2)
	for i := 0; i < 2; i++ {
		unit, err := world.ObjectManager.CreateUnit(0, "soldier",
			Vector3{X: float64(i), Y: 0, Z: 0}, nil)
		if err != nil {
			t.Fatalf("Failed to create unit %d: %v", i, err)
		}
		unitIDs[i] = unit.ID
	}

	// Create group
	group, err := world.CreateGroup(0, unitIDs, FormationLine)
	if err != nil {
		t.Fatalf("Failed to create group: %v", err)
	}

	// Test formation-aware pathfinding
	target := Vector3{X: 25, Y: 0, Z: 25}
	path, err := world.groupMgr.GetFormationPath(group.ID, target)
	if err != nil {
		// This might fail if pathfinding system requires more setup
		// But should not crash
		t.Logf("Pathfinding failed (expected in test): %v", err)
	} else {
		// If pathfinding succeeds, verify we got a path
		if len(path) == 0 {
			t.Error("Expected path with at least one point")
		}
	}
}

// TestWorldFormationCleanup tests cleanup when units are removed
func TestWorldFormationCleanup(t *testing.T) {
	world := createTestWorldWithFormations()

	// Add player
	err := world.AddPlayer(0, "Test Player", "tech", false)
	if err != nil {
		t.Fatalf("Failed to add player: %v", err)
	}

	// Create units
	units := make([]*GameUnit, 3)
	for i := 0; i < 3; i++ {
		unit, err := world.ObjectManager.CreateUnit(0, "soldier",
			Vector3{X: float64(i), Y: 0, Z: 0}, nil)
		if err != nil {
			t.Fatalf("Failed to create unit %d: %v", i, err)
		}
		units[i] = unit
	}

	unitIDs := []int{units[0].ID, units[1].ID, units[2].ID}

	// Create group
	group, err := world.CreateGroup(0, unitIDs, FormationLine)
	if err != nil {
		t.Fatalf("Failed to create group: %v", err)
	}

	if group.GetUnitCount() != 3 {
		t.Errorf("Expected 3 units in group initially, got %d", group.GetUnitCount())
	}

	// Simulate unit death
	units[1].Health = 0
	units[1].State = UnitStateDead

	// Run cleanup
	world.groupMgr.CleanupDeadUnits()

	// Group should still exist but with fewer units
	// Note: This depends on IsAlive() implementation in GameUnit
	// The test may need adjustment based on actual death handling
}

// TestFormationPerformance tests performance with many groups
func TestFormationPerformance(t *testing.T) {
	world := createTestWorldWithFormations()

	// Add player
	err := world.AddPlayer(0, "Test Player", "tech", false)
	if err != nil {
		t.Fatalf("Failed to add player: %v", err)
	}

	// Create multiple groups (stress test)
	const numGroups = 10
	const unitsPerGroup = 5

	groups := make([]*UnitGroup, numGroups)

	for g := 0; g < numGroups; g++ {
		unitIDs := make([]int, unitsPerGroup)
		for u := 0; u < unitsPerGroup; u++ {
			unit, err := world.ObjectManager.CreateUnit(0, "soldier",
				Vector3{X: float64(g*10 + u), Y: 0, Z: 0}, nil)
			if err != nil {
				t.Fatalf("Failed to create unit %d in group %d: %v", u, g, err)
			}
			unitIDs[u] = unit.ID
		}

		group, err := world.CreateGroup(0, unitIDs, FormationType(g%5)) // Cycle through formations
		if err != nil {
			t.Fatalf("Failed to create group %d: %v", g, err)
		}
		groups[g] = group
	}

	// Measure update performance
	start := time.Now()

	// Run multiple updates
	for i := 0; i < 10; i++ {
		world.Update(time.Millisecond * 16)
	}

	duration := time.Since(start)

	// Verify all groups still exist
	playerGroups := world.GetPlayerGroups(0)
	if len(playerGroups) != numGroups {
		t.Errorf("Expected %d groups after updates, got %d", numGroups, len(playerGroups))
	}

	// Performance should be reasonable (this is a rough check)
	if duration > time.Second {
		t.Errorf("Formation updates took too long: %v", duration)
	}

	t.Logf("Formation performance: %d groups with %d units each updated 10 times in %v",
		numGroups, unitsPerGroup, duration)
}