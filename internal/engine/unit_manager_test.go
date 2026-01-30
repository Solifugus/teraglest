package engine

import (
	"testing"
	"time"

	"teraglest/internal/data"
)

func createTestWorldForUnits() *World {
	// Create a minimal world for testing
	settings := GameSettings{
		PlayerFactions: map[int]string{1: "test_faction"},
	}

	techTree := &data.TechTree{}
	assetMgr := &data.AssetManager{}

	world, _ := NewWorld(settings, techTree, assetMgr)
	return world
}

func createTestUnitDefinition() *data.UnitDefinition {
	return &data.UnitDefinition{
		Name: "Test Unit",
	}
}

func TestUnitManagerCreation(t *testing.T) {
	world := createTestWorldForUnits()
	unitManager := NewUnitManager(world)

	if unitManager == nil {
		t.Fatal("UnitManager creation failed")
	}

	if unitManager.world != world {
		t.Error("UnitManager world reference not set correctly")
	}

	if unitManager.nextID != 1 {
		t.Errorf("Expected initial nextID to be 1, got %d", unitManager.nextID)
	}

	if len(unitManager.units) != 0 {
		t.Error("Expected empty units map initially")
	}

	if len(unitManager.unitsByPlayer) != 0 {
		t.Error("Expected empty unitsByPlayer map initially")
	}
}

func TestUnitManagerCreateUnit(t *testing.T) {
	world := createTestWorldForUnits()
	unitManager := NewUnitManager(world)
	unitDef := createTestUnitDefinition()

	position := Vector3{X: 5, Y: 0, Z: 10}
	playerID := 1
	unitType := "worker"

	// Create unit
	unit, err := unitManager.CreateUnit(playerID, unitType, position, unitDef)

	if err != nil {
		t.Fatalf("Unit creation failed: %v", err)
	}

	if unit == nil {
		t.Fatal("Created unit is nil")
	}

	// Test unit properties
	if unit.GetID() != 1 {
		t.Errorf("Expected unit ID 1, got %d", unit.GetID())
	}

	if unit.GetPlayerID() != playerID {
		t.Errorf("Expected player ID %d, got %d", playerID, unit.GetPlayerID())
	}

	if unit.UnitType != unitType {
		t.Errorf("Expected unit type %s, got %s", unitType, unit.UnitType)
	}

	// Test that unit is stored in manager
	retrievedUnit := unitManager.GetUnit(1)
	if retrievedUnit != unit {
		t.Error("Unit not stored correctly in manager")
	}

	// Test grid position initialization
	expectedGrid := WorldToGrid(position, world.tileSize)
	if unit.GridPos.Grid != expectedGrid.Grid {
		t.Error("Grid position not initialized correctly")
	}

	// Test that position has units in world
	unitsInTile := world.GetUnitsInTile(unit.GridPos.Grid)
	if len(unitsInTile) == 0 {
		t.Error("Unit position should have units in world grid")
	}

	// Test stats
	stats := unitManager.GetStats()
	if stats.TotalUnits != 1 {
		t.Errorf("Expected 1 total unit, got %d", stats.TotalUnits)
	}

	if stats.UnitsPerPlayer[playerID] != 1 {
		t.Errorf("Expected 1 unit for player %d, got %d", playerID, stats.UnitsPerPlayer[playerID])
	}
}

func TestUnitManagerMultipleUnits(t *testing.T) {
	world := createTestWorldForUnits()
	unitManager := NewUnitManager(world)
	unitDef := createTestUnitDefinition()

	// Create multiple units for different players
	positions := []Vector3{
		{X: 0, Y: 0, Z: 0},
		{X: 5, Y: 0, Z: 5},
		{X: 10, Y: 0, Z: 10},
	}

	players := []int{1, 1, 2}

	var createdUnits []*GameUnit

	for i, pos := range positions {
		unit, err := unitManager.CreateUnit(players[i], "worker", pos, unitDef)
		if err != nil {
			t.Fatalf("Failed to create unit %d: %v", i, err)
		}
		createdUnits = append(createdUnits, unit)
	}

	// Test total unit count
	stats := unitManager.GetStats()
	if stats.TotalUnits != 3 {
		t.Errorf("Expected 3 total units, got %d", stats.TotalUnits)
	}

	// Test units per player
	if stats.UnitsPerPlayer[1] != 2 {
		t.Errorf("Expected 2 units for player 1, got %d", stats.UnitsPerPlayer[1])
	}

	if stats.UnitsPerPlayer[2] != 1 {
		t.Errorf("Expected 1 unit for player 2, got %d", stats.UnitsPerPlayer[2])
	}

	// Test GetUnitsForPlayer
	player1Units := unitManager.GetUnitsForPlayer(1)
	if len(player1Units) != 2 {
		t.Errorf("Expected 2 units for player 1, got %d", len(player1Units))
	}

	player2Units := unitManager.GetUnitsForPlayer(2)
	if len(player2Units) != 1 {
		t.Errorf("Expected 1 unit for player 2, got %d", len(player2Units))
	}

	// Test that retrieved units match created units
	for _, unit := range createdUnits {
		if unit.GetPlayerID() == 1 {
			if _, exists := player1Units[unit.GetID()]; !exists {
				t.Errorf("Unit %d not found in player 1 units", unit.GetID())
			}
		}
		if unit.GetPlayerID() == 2 {
			if _, exists := player2Units[unit.GetID()]; !exists {
				t.Errorf("Unit %d not found in player 2 units", unit.GetID())
			}
		}
	}
}

func TestUnitManagerSpatialQueries(t *testing.T) {
	world := createTestWorldForUnits()
	unitManager := NewUnitManager(world)
	unitDef := createTestUnitDefinition()

	// Create units at specific positions
	positions := []struct {
		pos      Vector3
		expected Vector2i
	}{
		{Vector3{X: 0, Y: 0, Z: 0}, Vector2i{X: 0, Y: 0}},
		{Vector3{X: 1, Y: 0, Z: 1}, Vector2i{X: 1, Y: 1}},
		{Vector3{X: 2, Y: 0, Z: 2}, Vector2i{X: 2, Y: 2}},
		{Vector3{X: 1, Y: 0, Z: 1}, Vector2i{X: 1, Y: 1}}, // Second unit at same grid position
	}

	var units []*GameUnit
	for i, pos := range positions {
		unit, err := unitManager.CreateUnit(1, "worker", pos.pos, unitDef)
		if err != nil {
			t.Fatalf("Failed to create unit %d: %v", i, err)
		}
		units = append(units, unit)
	}

	t.Run("GetUnitsAtPosition", func(t *testing.T) {
		// Test getting units at specific positions
		unitsAt00 := unitManager.GetUnitsAtPosition(Vector2i{X: 0, Y: 0})
		if len(unitsAt00) != 1 {
			t.Errorf("Expected 1 unit at (0,0), got %d", len(unitsAt00))
		}

		unitsAt11 := unitManager.GetUnitsAtPosition(Vector2i{X: 1, Y: 1})
		if len(unitsAt11) != 2 {
			t.Errorf("Expected 2 units at (1,1), got %d", len(unitsAt11))
		}

		unitsAt99 := unitManager.GetUnitsAtPosition(Vector2i{X: 9, Y: 9})
		if len(unitsAt99) != 0 {
			t.Errorf("Expected 0 units at (9,9), got %d", len(unitsAt99))
		}
	})

	t.Run("GetUnitsInArea", func(t *testing.T) {
		// Test area queries
		unitsInArea := unitManager.GetUnitsInArea(Vector2i{X: 0, Y: 0}, Vector2i{X: 2, Y: 2})
		if len(unitsInArea) != 4 {
			t.Errorf("Expected 4 units in area (0,0)-(2,2), got %d", len(unitsInArea))
		}

		smallArea := unitManager.GetUnitsInArea(Vector2i{X: 1, Y: 1}, Vector2i{X: 1, Y: 1})
		if len(smallArea) != 2 {
			t.Errorf("Expected 2 units in area (1,1)-(1,1), got %d", len(smallArea))
		}
	})

	t.Run("IsPositionOccupied", func(t *testing.T) {
		// Test occupancy checks
		if !unitManager.IsPositionOccupied(Vector2i{X: 0, Y: 0}) {
			t.Error("Position (0,0) should be occupied")
		}

		if !unitManager.IsPositionOccupied(Vector2i{X: 1, Y: 1}) {
			t.Error("Position (1,1) should be occupied")
		}

		if unitManager.IsPositionOccupied(Vector2i{X: 5, Y: 5}) {
			t.Error("Position (5,5) should not be occupied")
		}
	})

	t.Run("FindNearestFreePosition", func(t *testing.T) {
		// Test finding free positions
		freePos := unitManager.FindNearestFreePosition(Vector2i{X: 0, Y: 0})

		// Should find a nearby free position
		if unitManager.IsPositionOccupied(freePos) {
			t.Error("FindNearestFreePosition returned an occupied position")
		}

		// Distance should be reasonable (adjacent or nearby)
		distance := CalculateGridDistance(Vector2i{X: 0, Y: 0}, freePos)
		if distance > 3 {
			t.Errorf("Free position too far away: distance %d", distance)
		}

		// Test with already free position
		freePos2 := unitManager.FindNearestFreePosition(Vector2i{X: 10, Y: 10})
		if freePos2.X != 10 || freePos2.Y != 10 {
			t.Error("Should return the same position if already free")
		}
	})

	t.Run("GetNearestUnit", func(t *testing.T) {
		// Test finding nearest units
		nearestToOrigin := unitManager.GetNearestUnit(Vector2i{X: 0, Y: 0}, 5, -1)
		if nearestToOrigin != units[0] {
			t.Error("Should find the unit at origin")
		}

		// Test with exclusion
		nearestExcludingPlayer1 := unitManager.GetNearestUnit(Vector2i{X: 0, Y: 0}, 5, 1)
		if nearestExcludingPlayer1 != nil {
			t.Error("Should not find any units when excluding player 1")
		}

		// Test with limited radius
		tooFar := unitManager.GetNearestUnit(Vector2i{X: 10, Y: 10}, 1, -1)
		if tooFar != nil {
			t.Error("Should not find units outside radius")
		}
	})
}

func TestUnitManagerRemoveUnit(t *testing.T) {
	world := createTestWorldForUnits()
	unitManager := NewUnitManager(world)
	unitDef := createTestUnitDefinition()

	// Create a unit
	unit, err := unitManager.CreateUnit(1, "worker", Vector3{X: 5, Y: 0, Z: 5}, unitDef)
	if err != nil {
		t.Fatalf("Failed to create unit: %v", err)
	}

	unitID := unit.GetID()
	gridPos := unit.GetGridPosition().Grid

	// Verify unit exists
	if unitManager.GetUnit(unitID) == nil {
		t.Error("Unit should exist before removal")
	}

	if !unitManager.IsPositionOccupied(gridPos) {
		t.Error("Position should be occupied before removal")
	}

	// Remove unit
	err = unitManager.RemoveUnit(unitID)
	if err != nil {
		t.Fatalf("Failed to remove unit: %v", err)
	}

	// Verify unit is removed
	if unitManager.GetUnit(unitID) != nil {
		t.Error("Unit should not exist after removal")
	}

	if unitManager.IsPositionOccupied(gridPos) {
		t.Error("Position should not be occupied after removal")
	}

	// Verify stats updated
	stats := unitManager.GetStats()
	if stats.TotalUnits != 0 {
		t.Errorf("Expected 0 total units after removal, got %d", stats.TotalUnits)
	}

	// Test removing non-existent unit
	err = unitManager.RemoveUnit(999)
	if err == nil {
		t.Error("Should return error when removing non-existent unit")
	}
}

func TestUnitManagerUpdate(t *testing.T) {
	world := createTestWorldForUnits()
	unitManager := NewUnitManager(world)
	unitDef := createTestUnitDefinition()

	// Create a unit
	unit, err := unitManager.CreateUnit(1, "worker", Vector3{X: 0, Y: 0, Z: 0}, unitDef)
	if err != nil {
		t.Fatalf("Failed to create unit: %v", err)
	}

	oldUpdateTime := unit.LastUpdate

	// Update unit manager
	deltaTime := 100 * time.Millisecond
	unitManager.Update(deltaTime)

	// Verify unit was updated
	if !unit.LastUpdate.After(oldUpdateTime) {
		t.Error("Unit should have been updated")
	}

	// Test with dead unit (should be removed)
	unit.SetHealth(0)
	initialCount := len(unitManager.units)

	unitManager.Update(deltaTime)

	// Dead unit should be removed
	if len(unitManager.units) >= initialCount {
		t.Error("Dead unit should have been removed")
	}
}

func TestUnitManagerThreadSafety(t *testing.T) {
	world := createTestWorldForUnits()
	unitManager := NewUnitManager(world)
	unitDef := createTestUnitDefinition()

	// Test concurrent unit creation and removal
	done := make(chan bool, 2)

	// Goroutine 1: Create units
	go func() {
		for i := 0; i < 10; i++ {
			_, _ = unitManager.CreateUnit(1, "worker", Vector3{X: float64(i), Y: 0, Z: 0}, unitDef)
		}
		done <- true
	}()

	// Goroutine 2: Read unit stats
	go func() {
		for i := 0; i < 10; i++ {
			_ = unitManager.GetStats()
			time.Sleep(1 * time.Millisecond)
		}
		done <- true
	}()

	// Wait for both goroutines
	<-done
	<-done

	// Should not panic or corrupt data
	stats := unitManager.GetStats()
	if stats.TotalUnits < 0 || stats.TotalUnits > 10 {
		t.Errorf("Invalid unit count after concurrent operations: %d", stats.TotalUnits)
	}
}

func TestUnitManagerPhase23Requirements(t *testing.T) {
	world := createTestWorldForUnits()
	unitManager := NewUnitManager(world)
	unitDef := createTestUnitDefinition()

	t.Run("UnitManager for spawning, tracking, removing units", func(t *testing.T) {
		// ✅ UnitManager for spawning, tracking, removing units

		// Test spawning
		unit, err := unitManager.CreateUnit(1, "worker", Vector3{X: 0, Y: 0, Z: 0}, unitDef)
		if err != nil || unit == nil {
			t.Error("UnitManager should be able to spawn units")
		}

		// Test tracking
		trackedUnit := unitManager.GetUnit(unit.GetID())
		if trackedUnit != unit {
			t.Error("UnitManager should be able to track units")
		}

		playerUnits := unitManager.GetUnitsForPlayer(1)
		if len(playerUnits) != 1 {
			t.Error("UnitManager should track units by player")
		}

		// Test removing
		err = unitManager.RemoveUnit(unit.GetID())
		if err != nil {
			t.Error("UnitManager should be able to remove units")
		}

		if unitManager.GetUnit(unit.GetID()) != nil {
			t.Error("Unit should be removed from tracking")
		}
	})

	t.Run("Grid-based spatial queries", func(t *testing.T) {
		// Create test units
		positions := []Vector3{
			{X: 0, Y: 0, Z: 0}, // Grid (0,0)
			{X: 1, Y: 0, Z: 1}, // Grid (1,1)
			{X: 2, Y: 0, Z: 1}, // Grid (2,1)
		}

		for i, pos := range positions {
			_, err := unitManager.CreateUnit(1, "worker", pos, unitDef)
			if err != nil {
				t.Fatalf("Failed to create unit %d", i)
			}
		}

		// Test spatial queries
		unitsAt00 := unitManager.GetUnitsAtPosition(Vector2i{X: 0, Y: 0})
		if len(unitsAt00) != 1 {
			t.Error("Spatial query at (0,0) should return 1 unit")
		}

		unitsInArea := unitManager.GetUnitsInArea(Vector2i{X: 0, Y: 0}, Vector2i{X: 2, Y: 1})
		if len(unitsInArea) != 3 {
			t.Error("Area query should return all 3 units")
		}

		if !unitManager.IsPositionOccupied(Vector2i{X: 1, Y: 1}) {
			t.Error("Occupancy check should detect unit at (1,1)")
		}

		freePos := unitManager.FindNearestFreePosition(Vector2i{X: 0, Y: 0})
		if unitManager.IsPositionOccupied(freePos) {
			t.Error("FindNearestFreePosition should return unoccupied position")
		}
	})

	t.Run("Basic collision detection", func(t *testing.T) {
		// Create unit at specific position
		unit, err := unitManager.CreateUnit(1, "worker", Vector3{X: 5, Y: 0, Z: 5}, unitDef)
		if err != nil {
			t.Fatal("Failed to create unit")
		}

		gridPos := unit.GetGridPosition().Grid

		// Position should be marked as occupied
		if !unitManager.IsPositionOccupied(gridPos) {
			t.Error("Unit position should be marked as occupied (collision detection)")
		}

		// World grid should also reflect occupancy
		unitsInTile := world.GetUnitsInTile(gridPos)
		if len(unitsInTile) == 0 {
			t.Error("World grid should show position as occupied")
		}

		// Remove unit and verify position is freed
		unitManager.RemoveUnit(unit.GetID())
		if unitManager.IsPositionOccupied(gridPos) {
			t.Error("Position should be freed after unit removal")
		}
	})
}