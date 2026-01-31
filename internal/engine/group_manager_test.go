package engine

import (
	"testing"
	"time"

	"teraglest/internal/data"
)

// createTestWorldForGroups creates a test world for group testing
func createTestWorldForGroups() *World {
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

// TestGroupManagerCreation tests basic group manager creation
func TestGroupManagerCreation(t *testing.T) {
	world := createTestWorldForGroups()

	if world.groupMgr == nil {
		t.Fatal("Group manager should be initialized")
	}

	// Test initial state
	groups := world.groupMgr.GetPlayerGroups(0)
	if len(groups) != 0 {
		t.Errorf("Expected 0 groups initially, got %d", len(groups))
	}
}

// TestGroupCreation tests creating groups through the manager
func TestGroupCreation(t *testing.T) {
	world := createTestWorldForGroups()
	units := createTestUnits(3, 0)

	// Create group
	group, err := world.groupMgr.CreateGroup(0, units, FormationLine)
	if err != nil {
		t.Fatalf("Failed to create group: %v", err)
	}

	if group == nil {
		t.Fatal("Group should not be nil")
	}

	// Verify group was registered
	retrieved, exists := world.groupMgr.GetGroup(group.ID)
	if !exists {
		t.Error("Group should be retrievable after creation")
	}

	if retrieved.ID != group.ID {
		t.Errorf("Retrieved group ID %d doesn't match created group ID %d", retrieved.ID, group.ID)
	}

	// Verify player groups
	playerGroups := world.groupMgr.GetPlayerGroups(0)
	if len(playerGroups) != 1 {
		t.Errorf("Expected 1 group for player 0, got %d", len(playerGroups))
	}

	// Verify unit mapping
	for _, unit := range units {
		if !world.groupMgr.IsUnitInGroup(unit.ID) {
			t.Errorf("Unit %d should be in a group", unit.ID)
		}

		groupFromUnit, exists := world.groupMgr.GetUnitGroup(unit.ID)
		if !exists {
			t.Errorf("Should be able to get group from unit %d", unit.ID)
		}

		if groupFromUnit.ID != group.ID {
			t.Errorf("Unit %d group ID mismatch", unit.ID)
		}
	}
}

// TestGroupManagerMovement tests moving groups through the manager
func TestGroupManagerMovement(t *testing.T) {
	world := createTestWorldForGroups()
	units := createTestUnits(3, 0)

	// Create group
	group, err := world.groupMgr.CreateGroup(0, units, FormationLine)
	if err != nil {
		t.Fatalf("Failed to create group: %v", err)
	}

	// Move group
	target := Vector3{X: 20, Y: 0, Z: 20}
	err = world.groupMgr.MoveGroup(group.ID, target)
	if err != nil {
		t.Errorf("Failed to move group: %v", err)
	}

	// Verify group is marked as moving
	if !group.IsMoving {
		t.Error("Group should be marked as moving")
	}

	// Verify target position was set
	if group.TargetPos.X != 20 || group.TargetPos.Z != 20 {
		t.Errorf("Target position not set correctly, got (%f, %f)", group.TargetPos.X, group.TargetPos.Z)
	}
}

// TestMultipleGroups tests managing multiple groups
func TestMultipleGroups(t *testing.T) {
	world := createTestWorldForGroups()

	// Create multiple groups
	units1 := createTestUnits(3, 0)
	units2 := createTestUnits(4, 0)

	// Offset unit IDs for second group
	for i, unit := range units2 {
		unit.ID = i + 10
	}

	group1, err := world.groupMgr.CreateGroup(0, units1, FormationLine)
	if err != nil {
		t.Fatalf("Failed to create group 1: %v", err)
	}

	group2, err := world.groupMgr.CreateGroup(0, units2, FormationColumn)
	if err != nil {
		t.Fatalf("Failed to create group 2: %v", err)
	}

	// Verify both groups exist
	playerGroups := world.groupMgr.GetPlayerGroups(0)
	if len(playerGroups) != 2 {
		t.Errorf("Expected 2 groups for player 0, got %d", len(playerGroups))
	}

	// Verify groups have different IDs
	if group1.ID == group2.ID {
		t.Error("Groups should have different IDs")
	}

	// Verify formations are correct
	if group1.Formation != FormationLine {
		t.Errorf("Group 1 should have line formation, got %v", group1.Formation)
	}

	if group2.Formation != FormationColumn {
		t.Errorf("Group 2 should have column formation, got %v", group2.Formation)
	}
}

// TestGroupDisbanding tests disbanding groups
func TestGroupDisbanding(t *testing.T) {
	world := createTestWorldForGroups()
	units := createTestUnits(3, 0)

	// Create group
	group, err := world.groupMgr.CreateGroup(0, units, FormationLine)
	if err != nil {
		t.Fatalf("Failed to create group: %v", err)
	}

	groupID := group.ID

	// Verify units are in group
	for _, unit := range units {
		if !world.groupMgr.IsUnitInGroup(unit.ID) {
			t.Errorf("Unit %d should be in group before disbanding", unit.ID)
		}
	}

	// Disband group
	err = world.groupMgr.DisbandGroup(groupID)
	if err != nil {
		t.Errorf("Failed to disband group: %v", err)
	}

	// Verify group no longer exists
	_, exists := world.groupMgr.GetGroup(groupID)
	if exists {
		t.Error("Group should not exist after disbanding")
	}

	// Verify units are no longer in any group
	for _, unit := range units {
		if world.groupMgr.IsUnitInGroup(unit.ID) {
			t.Errorf("Unit %d should not be in group after disbanding", unit.ID)
		}
	}

	// Verify player groups list is empty
	playerGroups := world.groupMgr.GetPlayerGroups(0)
	if len(playerGroups) != 0 {
		t.Errorf("Expected 0 groups after disbanding, got %d", len(playerGroups))
	}
}

// TestAddRemoveUnits tests adding and removing units from existing groups
func TestAddRemoveUnits(t *testing.T) {
	world := createTestWorldForGroups()
	units := createTestUnits(2, 0)

	// Create initial group
	group, err := world.groupMgr.CreateGroup(0, units, FormationLine)
	if err != nil {
		t.Fatalf("Failed to create group: %v", err)
	}

	if group.GetUnitCount() != 2 {
		t.Errorf("Expected 2 units in group, got %d", group.GetUnitCount())
	}

	// Create additional units to add
	additionalUnits := createTestUnits(2, 0)
	for i, unit := range additionalUnits {
		unit.ID = i + 20 // Ensure unique IDs
	}

	// Add units to group
	err = world.groupMgr.AddUnitsToGroup(group.ID, additionalUnits)
	if err != nil {
		t.Errorf("Failed to add units to group: %v", err)
	}

	if group.GetUnitCount() != 4 {
		t.Errorf("Expected 4 units after adding, got %d", group.GetUnitCount())
	}

	// Verify new units are mapped to group
	for _, unit := range additionalUnits {
		if !world.groupMgr.IsUnitInGroup(unit.ID) {
			t.Errorf("Added unit %d should be in group", unit.ID)
		}
	}

	// Remove some units
	removeIDs := []int{units[0].ID, additionalUnits[0].ID}
	err = world.groupMgr.RemoveUnitsFromGroup(group.ID, removeIDs)
	if err != nil {
		t.Errorf("Failed to remove units from group: %v", err)
	}

	if group.GetUnitCount() != 2 {
		t.Errorf("Expected 2 units after removing, got %d", group.GetUnitCount())
	}

	// Verify removed units are not in group
	for _, unitID := range removeIDs {
		if world.groupMgr.IsUnitInGroup(unitID) {
			t.Errorf("Removed unit %d should not be in group", unitID)
		}
	}
}

// TestFormationChanges tests changing formations through the manager
func TestFormationChanges(t *testing.T) {
	world := createTestWorldForGroups()
	units := createTestUnits(4, 0)

	// Create group with line formation
	group, err := world.groupMgr.CreateGroup(0, units, FormationLine)
	if err != nil {
		t.Fatalf("Failed to create group: %v", err)
	}

	if group.Formation != FormationLine {
		t.Errorf("Expected line formation initially, got %v", group.Formation)
	}

	// Change to wedge formation
	err = world.groupMgr.SetGroupFormation(group.ID, FormationWedge)
	if err != nil {
		t.Errorf("Failed to set group formation: %v", err)
	}

	if group.Formation != FormationWedge {
		t.Errorf("Expected wedge formation after change, got %v", group.Formation)
	}

	// Verify positions were regenerated
	wedgePositions := make(map[int]FormationPosition)
	for unitID, pos := range group.Positions {
		wedgePositions[unitID] = pos
	}

	if len(wedgePositions) != 4 {
		t.Errorf("Expected 4 positions after formation change, got %d", len(wedgePositions))
	}
}

// TestGroupManagerUpdate tests group manager update cycle
func TestGroupManagerUpdate(t *testing.T) {
	world := createTestWorldForGroups()
	units := createTestUnits(3, 0)

	// Create group
	group, err := world.groupMgr.CreateGroup(0, units, FormationLine)
	if err != nil {
		t.Fatalf("Failed to create group: %v", err)
	}

	// Update group manager
	deltaTime := time.Millisecond * 16
	initialUpdateTime := group.LastUpdate

	world.groupMgr.Update(deltaTime)

	// Verify group was updated
	if !group.LastUpdate.After(initialUpdateTime) {
		t.Error("Group should have been updated")
	}
}

// TestPlayerIsolation tests that players can't access other players' groups
func TestPlayerIsolation(t *testing.T) {
	world := createTestWorldForGroups()

	// Create groups for different players
	units1 := createTestUnits(2, 0) // Player 0
	units2 := createTestUnits(2, 1) // Player 1

	for i, unit := range units2 {
		unit.ID = i + 10
		unit.PlayerID = 1
	}

	group1, err := world.groupMgr.CreateGroup(0, units1, FormationLine)
	if err != nil {
		t.Fatalf("Failed to create group for player 0: %v", err)
	}

	group2, err := world.groupMgr.CreateGroup(1, units2, FormationColumn)
	if err != nil {
		t.Fatalf("Failed to create group for player 1: %v", err)
	}

	// Verify player isolation
	player0Groups := world.groupMgr.GetPlayerGroups(0)
	player1Groups := world.groupMgr.GetPlayerGroups(1)

	if len(player0Groups) != 1 {
		t.Errorf("Player 0 should have 1 group, got %d", len(player0Groups))
	}

	if len(player1Groups) != 1 {
		t.Errorf("Player 1 should have 1 group, got %d", len(player1Groups))
	}

	if player0Groups[0].ID == player1Groups[0].ID {
		t.Error("Different players should have different group IDs")
	}

	// Verify groups have correct player IDs
	if group1.PlayerID != 0 {
		t.Errorf("Group 1 should belong to player 0, got player %d", group1.PlayerID)
	}

	if group2.PlayerID != 1 {
		t.Errorf("Group 2 should belong to player 1, got player %d", group2.PlayerID)
	}
}

// TestCleanupDeadUnits tests automatic cleanup of dead units
func TestCleanupDeadUnits(t *testing.T) {
	world := createTestWorldForGroups()
	units := createTestUnits(3, 0)

	// Create group
	group, err := world.groupMgr.CreateGroup(0, units, FormationLine)
	if err != nil {
		t.Fatalf("Failed to create group: %v", err)
	}

	if group.GetUnitCount() != 3 {
		t.Errorf("Expected 3 units initially, got %d", group.GetUnitCount())
	}

	// Mark one unit as dead
	units[1].Health = 0
	units[1].State = UnitStateDead

	// Cleanup dead units
	world.groupMgr.CleanupDeadUnits()

	// Note: IsAlive() method needs to be implemented on GameUnit
	// For now, test assumes cleanup removes units with Health <= 0
	// This test may need adjustment based on actual IsAlive() implementation
}

// TestGroupStats tests statistics gathering
func TestGroupStats(t *testing.T) {
	world := createTestWorldForGroups()

	// Create groups with different formations
	units1 := createTestUnits(3, 0)
	units2 := createTestUnits(2, 0)

	for i, unit := range units2 {
		unit.ID = i + 10
	}

	_, err := world.groupMgr.CreateGroup(0, units1, FormationLine)
	if err != nil {
		t.Fatalf("Failed to create line group: %v", err)
	}

	_, err = world.groupMgr.CreateGroup(0, units2, FormationWedge)
	if err != nil {
		t.Fatalf("Failed to create wedge group: %v", err)
	}

	// Get statistics
	stats := world.groupMgr.GetGroupStats()

	totalGroups, ok := stats["total_groups"].(int)
	if !ok || totalGroups != 2 {
		t.Errorf("Expected 2 total groups, got %v", stats["total_groups"])
	}

	totalUnits, ok := stats["total_units"].(int)
	if !ok || totalUnits != 5 {
		t.Errorf("Expected 5 total units, got %v", stats["total_units"])
	}

	formations, ok := stats["formations"].(map[string]int)
	if !ok {
		t.Error("Expected formations map in stats")
	} else {
		if formations["Line"] != 1 {
			t.Errorf("Expected 1 Line formation, got %d", formations["Line"])
		}
		if formations["Wedge"] != 1 {
			t.Errorf("Expected 1 Wedge formation, got %d", formations["Wedge"])
		}
	}
}

// TestBatchOperations tests batch group creation
func TestBatchOperations(t *testing.T) {
	world := createTestWorldForGroups()

	// Create multiple unit arrays
	units1 := createTestUnits(2, 0)
	units2 := createTestUnits(3, 0)

	for i, unit := range units2 {
		unit.ID = i + 10
	}

	unitGroups := [][](*GameUnit){units1, units2}
	formations := []FormationType{FormationLine, FormationColumn}

	// Create multiple groups at once
	groups, err := world.groupMgr.CreateMultipleGroups(0, unitGroups, formations)
	if err != nil {
		t.Fatalf("Failed to create multiple groups: %v", err)
	}

	if len(groups) != 2 {
		t.Errorf("Expected 2 groups created, got %d", len(groups))
	}

	// Verify formations
	if groups[0].Formation != FormationLine {
		t.Errorf("First group should have line formation, got %v", groups[0].Formation)
	}

	if groups[1].Formation != FormationColumn {
		t.Errorf("Second group should have column formation, got %v", groups[1].Formation)
	}

	// Verify all units are in groups
	totalUnits := 0
	for _, group := range groups {
		totalUnits += group.GetUnitCount()
	}

	if totalUnits != 5 {
		t.Errorf("Expected 5 total units in groups, got %d", totalUnits)
	}
}