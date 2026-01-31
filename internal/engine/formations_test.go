package engine

import (
	"testing"
	"time"
)

// createTestUnits creates test units for formation testing
func createTestUnits(count int, playerID int) []*GameUnit {
	units := make([]*GameUnit, count)
	for i := 0; i < count; i++ {
		units[i] = &GameUnit{
			ID:           i + 1,
			PlayerID:     playerID,
			UnitType:     "soldier",
			Position:     Vector3{X: float64(i), Y: 0, Z: 0},
			Health:       100,
			MaxHealth:    100,
			State:        UnitStateIdle,
			CreationTime: time.Now(),
			LastUpdate:   time.Now(),
		}
	}
	return units
}

// TestFormationCreation tests basic formation creation
func TestFormationCreation(t *testing.T) {
	units := createTestUnits(5, 0)
	group := NewUnitGroup(1, 0, units, FormationLine)

	if group == nil {
		t.Fatal("Failed to create unit group")
	}

	if group.GetUnitCount() != 5 {
		t.Errorf("Expected 5 units, got %d", group.GetUnitCount())
	}

	if group.Formation != FormationLine {
		t.Errorf("Expected FormationLine, got %v", group.Formation)
	}

	if group.Leader == nil {
		t.Error("Expected leader to be set")
	}
}

// TestLineFormation tests line formation positioning
func TestLineFormation(t *testing.T) {
	units := createTestUnits(3, 0)
	group := NewUnitGroup(1, 0, units, FormationLine)

	// Check that positions were generated
	if len(group.Positions) != 3 {
		t.Errorf("Expected 3 formation positions, got %d", len(group.Positions))
	}

	// Check that positions are in a line (same Z, different X)
	for _, unit := range units {
		pos, exists := group.Positions[unit.ID]
		if !exists {
			t.Errorf("No position generated for unit %d", unit.ID)
			continue
		}

		if pos.RelativePos.Z != 0 {
			t.Errorf("Line formation should have Z=0, got %f", pos.RelativePos.Z)
		}
	}
}

// TestColumnFormation tests column formation positioning
func TestColumnFormation(t *testing.T) {
	units := createTestUnits(4, 0)
	group := NewUnitGroup(1, 0, units, FormationColumn)

	// Check positions are in a column (same X, different Z)
	for _, unit := range units {
		pos, exists := group.Positions[unit.ID]
		if !exists {
			t.Errorf("No position generated for unit %d", unit.ID)
			continue
		}

		if pos.RelativePos.X != 0 {
			t.Errorf("Column formation should have X=0, got %f", pos.RelativePos.X)
		}
	}
}

// TestWedgeFormation tests wedge formation positioning
func TestWedgeFormation(t *testing.T) {
	units := createTestUnits(5, 0)
	group := NewUnitGroup(1, 0, units, FormationWedge)

	// Leader should be at front center (0,0,0)
	if group.Leader == nil {
		t.Fatal("Expected leader to be set")
	}

	leaderPos, exists := group.Positions[group.Leader.ID]
	if !exists {
		t.Error("Leader should have a position")
	} else {
		if leaderPos.RelativePos.X != 0 || leaderPos.RelativePos.Z != 0 {
			t.Errorf("Leader should be at center front (0,0,0), got (%f,%f,%f)",
				leaderPos.RelativePos.X, leaderPos.RelativePos.Y, leaderPos.RelativePos.Z)
		}
	}
}

// TestCircleFormation tests circular formation positioning
func TestCircleFormation(t *testing.T) {
	units := createTestUnits(6, 0)
	group := NewUnitGroup(1, 0, units, FormationCircle)

	// All units should be roughly the same distance from center
	var distances []float32
	for _, unit := range units {
		pos, exists := group.Positions[unit.ID]
		if !exists {
			t.Errorf("No position generated for unit %d", unit.ID)
			continue
		}

		distance := float32(pos.RelativePos.X*pos.RelativePos.X + pos.RelativePos.Z*pos.RelativePos.Z)
		distances = append(distances, distance)
	}

	if len(distances) < 2 {
		t.Fatal("Not enough positions to test circular formation")
	}

	// Check that all distances are roughly equal (within tolerance)
	tolerance := float32(0.1)
	firstDistance := distances[0]
	for i, distance := range distances {
		if distance < firstDistance-tolerance || distance > firstDistance+tolerance {
			t.Errorf("Unit %d distance %f not equal to expected %f (tolerance %f)",
				i, distance, firstDistance, tolerance)
		}
	}
}

// TestFormationChange tests changing formation type
func TestFormationChange(t *testing.T) {
	units := createTestUnits(4, 0)
	group := NewUnitGroup(1, 0, units, FormationLine)

	// Verify initial formation
	if group.Formation != FormationLine {
		t.Errorf("Expected FormationLine, got %v", group.Formation)
	}

	// Change to column formation
	group.SetFormation(FormationColumn)

	if group.Formation != FormationColumn {
		t.Errorf("Expected FormationColumn after change, got %v", group.Formation)
	}

	// Verify positions were regenerated for new formation
	for _, unit := range units {
		pos, exists := group.Positions[unit.ID]
		if !exists {
			t.Errorf("No position after formation change for unit %d", unit.ID)
			continue
		}

		// In column formation, X should be 0
		if pos.RelativePos.X != 0 {
			t.Errorf("Column formation should have X=0, got %f", pos.RelativePos.X)
		}
	}
}

// TestGroupMovement tests group movement functionality
func TestGroupMovement(t *testing.T) {
	units := createTestUnits(3, 0)
	group := NewUnitGroup(1, 0, units, FormationLine)

	// Set initial target
	target := Vector3{X: 10, Y: 0, Z: 10}
	group.MoveToPosition(target)

	if !group.IsMoving {
		t.Error("Group should be marked as moving after move command")
	}

	if group.TargetPos.X != 10 || group.TargetPos.Z != 10 {
		t.Errorf("Target position not set correctly, got (%f, %f)", group.TargetPos.X, group.TargetPos.Z)
	}

	// Formation should be marked as not formed during movement
	if group.IsFormed {
		t.Error("Group should not be in formation during movement")
	}
}

// TestUnitAddRemove tests adding and removing units from groups
func TestUnitAddRemove(t *testing.T) {
	units := createTestUnits(3, 0)
	group := NewUnitGroup(1, 0, units, FormationLine)

	initialCount := group.GetUnitCount()
	if initialCount != 3 {
		t.Errorf("Expected 3 units initially, got %d", initialCount)
	}

	// Add a new unit
	newUnit := &GameUnit{
		ID:       99,
		PlayerID: 0,
		UnitType: "soldier",
		Position: Vector3{X: 0, Y: 0, Z: 0},
		Health:   100,
		State:    UnitStateIdle,
	}

	group.AddUnit(newUnit)

	if group.GetUnitCount() != 4 {
		t.Errorf("Expected 4 units after adding, got %d", group.GetUnitCount())
	}

	// Check that position was generated for new unit
	_, exists := group.Positions[newUnit.ID]
	if !exists {
		t.Error("Position should be generated for newly added unit")
	}

	// Remove a unit
	group.RemoveUnit(units[0].ID)

	if group.GetUnitCount() != 3 {
		t.Errorf("Expected 3 units after removing, got %d", group.GetUnitCount())
	}

	// Check that position was removed
	_, exists = group.Positions[units[0].ID]
	if exists {
		t.Error("Position should be removed for removed unit")
	}
}

// TestLeaderAssignment tests leader assignment and reassignment
func TestLeaderAssignment(t *testing.T) {
	units := createTestUnits(3, 0)
	group := NewUnitGroup(1, 0, units, FormationLine)

	// Initial leader should be first unit
	if group.Leader == nil {
		t.Fatal("Leader should be assigned initially")
	}

	originalLeaderID := group.Leader.ID

	// Remove the leader
	group.RemoveUnit(originalLeaderID)

	// New leader should be assigned
	if group.Leader == nil {
		t.Error("New leader should be assigned after removing original leader")
	}

	if group.Leader.ID == originalLeaderID {
		t.Error("Leader should be different after removal")
	}
}

// TestFormationParameters tests formation parameter functionality
func TestFormationParameters(t *testing.T) {
	units := createTestUnits(3, 0)
	group := NewUnitGroup(1, 0, units, FormationLine)

	// Test default parameters
	if group.Parameters.UnitSpacing <= 0 {
		t.Error("Default unit spacing should be positive")
	}

	if group.Parameters.MaxSpeed <= 0 {
		t.Error("Default max speed should be positive")
	}

	// Test parameter modification
	group.Parameters.UnitSpacing = 5.0
	group.generateFormationPositions()

	// Verify spacing was applied (line formation with wider spacing)
	positions := make([]Vector3, 0, len(units))
	for _, unit := range units {
		if pos, exists := group.Positions[unit.ID]; exists {
			positions = append(positions, pos.RelativePos)
		}
	}

	if len(positions) >= 2 {
		// Check distance between first two units
		distance := positions[1].X - positions[0].X
		if distance < 4.0 || distance > 6.0 { // Should be close to 5.0
			t.Errorf("Unit spacing not applied correctly, got distance %f", distance)
		}
	}
}

// TestGroupUpdate tests group update functionality
func TestGroupUpdate(t *testing.T) {
	units := createTestUnits(3, 0)
	group := NewUnitGroup(1, 0, units, FormationLine)

	// Update group
	deltaTime := time.Millisecond * 16
	group.Update(deltaTime)

	// Check that last update was recorded
	if time.Since(group.LastUpdate) > time.Second {
		t.Error("LastUpdate should be recent after Update call")
	}
}

// TestEmptyGroup tests empty group handling
func TestEmptyGroup(t *testing.T) {
	// Create group with no units
	group := NewUnitGroup(1, 0, []*GameUnit{}, FormationLine)

	if !group.IsEmpty() {
		t.Error("Group with no units should be empty")
	}

	if group.GetUnitCount() != 0 {
		t.Errorf("Empty group should have 0 units, got %d", group.GetUnitCount())
	}

	// Adding units should make it non-empty
	unit := createTestUnits(1, 0)[0]
	group.AddUnit(unit)

	if group.IsEmpty() {
		t.Error("Group should not be empty after adding unit")
	}

	if group.GetUnitCount() != 1 {
		t.Errorf("Group should have 1 unit after adding, got %d", group.GetUnitCount())
	}
}

// TestFormationWorldPositions tests world position calculation
func TestFormationWorldPositions(t *testing.T) {
	units := createTestUnits(2, 0)
	group := NewUnitGroup(1, 0, units, FormationLine)

	// Set group position
	group.CenterPos = Vector3{X: 10, Y: 0, Z: 10}
	group.TargetPos = Vector3{X: 10, Y: 0, Z: 10}

	// Get world positions
	for _, unit := range units {
		worldPos, exists := group.GetFormationPosition(unit.ID)
		if !exists {
			t.Errorf("No world position for unit %d", unit.ID)
			continue
		}

		// World position should be offset from center
		if worldPos.X < 8 || worldPos.X > 12 {
			t.Errorf("World position X should be near center (10), got %f", worldPos.X)
		}

		if worldPos.Z < 8 || worldPos.Z > 12 {
			t.Errorf("World position Z should be near center (10), got %f", worldPos.Z)
		}
	}
}