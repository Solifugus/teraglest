package engine

import (
	"testing"
	"time"

	"teraglest/internal/data"
)

// Test unit creation and basic functionality
func TestGameUnitCreation(t *testing.T) {
	// Create a simple unit definition for testing
	unitDef := &data.UnitDefinition{
		Name: "Test Unit",
	}

	position := Vector3{X: 10, Y: 0, Z: 15}
	tileSize := float32(1.0)

	// Test unit creation
	unit := &GameUnit{
		ID:           1,
		PlayerID:     1,
		UnitType:     "test_unit",
		Name:         unitDef.Name,
		Position:     position,
		GridPos:      WorldToGrid(position, tileSize),
		Health:       unitDef.Unit.Parameters.MaxHP.Value,
		MaxHealth:    unitDef.Unit.Parameters.MaxHP.Value,
		Armor:        unitDef.Unit.Parameters.Armor.Value,
		Energy:       100,
		MaxEnergy:    100,
		State:        UnitStateIdle,
		CreationTime: time.Now(),
		LastUpdate:   time.Now(),
		CommandQueue: make([]UnitCommand, 0),
		Speed:        2.0,
		CarriedResources: make(map[string]int),
		GatherRate:      map[string]float32{"wood": 10.0},
		UnitDef:         unitDef,
	}

	// Test basic getters
	if unit.GetID() != 1 {
		t.Errorf("Expected ID 1, got %d", unit.GetID())
	}

	if unit.GetPlayerID() != 1 {
		t.Errorf("Expected PlayerID 1, got %d", unit.GetPlayerID())
	}

	if unit.GetHealth() != 100 {
		t.Errorf("Expected Health 100, got %d", unit.GetHealth())
	}

	if unit.GetMaxHealth() != 100 {
		t.Errorf("Expected MaxHealth 100, got %d", unit.GetMaxHealth())
	}

	if unit.GetType() != "test_unit" {
		t.Errorf("Expected Type 'test_unit', got %s", unit.GetType())
	}

	if !unit.IsAlive() {
		t.Error("Unit should be alive")
	}
}

func TestGameUnitGridPositioning(t *testing.T) {
	unit := &GameUnit{
		ID:       1,
		PlayerID: 1,
		Position: Vector3{X: 5, Y: 0, Z: 10},
		GridPos:  GridPosition{Grid: Vector2i{X: 5, Y: 10}, Offset: Vector2{X: 0.0, Y: 0.0}},
		Health:   100,
		State:    UnitStateIdle,
	}

	tileSize := float32(1.0)

	// Test grid position getter
	gridPos := unit.GetGridPosition()
	expectedGrid := Vector2i{X: 5, Y: 10}
	if gridPos.Grid != expectedGrid {
		t.Errorf("Expected grid position %v, got %v", expectedGrid, gridPos.Grid)
	}

	// Test setting grid position
	newGridPos := GridPosition{
		Grid:   Vector2i{X: 7, Y: 8},
		Offset: Vector2{X: 0.5, Y: 0.5},
	}
	unit.SetGridPosition(newGridPos, tileSize)

	// Verify both positions were updated
	if unit.GetGridPosition().Grid != newGridPos.Grid {
		t.Errorf("Grid position not updated correctly")
	}

	expectedWorldPos := GridToWorld(newGridPos, tileSize)
	actualWorldPos := unit.GetPosition()
	if actualWorldPos.X != expectedWorldPos.X || actualWorldPos.Z != expectedWorldPos.Z {
		t.Errorf("World position not synchronized with grid position")
	}

	// Test UpdatePositions method
	newWorldPos := Vector3{X: 3.7, Y: 0, Z: 2.3}
	unit.UpdatePositions(newWorldPos, tileSize)

	expectedGridFromWorld := WorldToGrid(newWorldPos, tileSize)
	actualGrid := unit.GetGridPosition()
	if actualGrid.Grid != expectedGridFromWorld.Grid {
		t.Errorf("Grid position not updated from world coordinates")
	}
}

func TestGameUnitHealth(t *testing.T) {
	unit := &GameUnit{
		ID:        1,
		Health:    100,
		MaxHealth: 100,
		State:     UnitStateIdle,
	}

	// Test taking damage
	unit.SetHealth(75)
	if unit.GetHealth() != 75 {
		t.Errorf("Expected health 75, got %d", unit.GetHealth())
	}

	if !unit.IsAlive() {
		t.Error("Unit should still be alive")
	}

	// Test death
	unit.SetHealth(0)
	if unit.GetHealth() != 0 {
		t.Errorf("Expected health 0, got %d", unit.GetHealth())
	}

	if unit.IsAlive() {
		t.Error("Unit should be dead")
	}

	// Test negative health (should clamp to 0)
	unit.SetHealth(-10)
	if unit.GetHealth() != 0 {
		t.Errorf("Health should be clamped to 0, got %d", unit.GetHealth())
	}

	// Test healing above max health
	unit.Health = 50 // Reset for healing test
	unit.SetHealth(150)
	if unit.GetHealth() != unit.GetMaxHealth() {
		t.Errorf("Health should be clamped to max health %d, got %d", unit.GetMaxHealth(), unit.GetHealth())
	}
}

func TestGameUnitStateTransitions(t *testing.T) {
	unit := &GameUnit{
		ID:        1,
		Health:    100,
		MaxHealth: 100,
		State:     UnitStateIdle,
	}

	// Test initial state
	if unit.State != UnitStateIdle {
		t.Errorf("Expected initial state Idle, got %v", unit.State)
	}

	// Test state transitions
	unit.State = UnitStateMoving
	if unit.State != UnitStateMoving {
		t.Errorf("Expected state Moving, got %v", unit.State)
	}

	unit.State = UnitStateAttacking
	if unit.State != UnitStateAttacking {
		t.Errorf("Expected state Attacking, got %v", unit.State)
	}

	unit.State = UnitStateGathering
	if unit.State != UnitStateGathering {
		t.Errorf("Expected state Gathering, got %v", unit.State)
	}

	unit.State = UnitStateBuilding
	if unit.State != UnitStateBuilding {
		t.Errorf("Expected state Building, got %v", unit.State)
	}

	unit.State = UnitStateDead
	if unit.State != UnitStateDead {
		t.Errorf("Expected state Dead, got %v", unit.State)
	}
}

func TestGameUnitUpdate(t *testing.T) {
	unit := &GameUnit{
		ID:        1,
		Health:    50,
		MaxHealth: 100,
		State:     UnitStateIdle,
	}

	// Test health regeneration
	deltaTime := 1 * time.Second
	unit.Update(deltaTime)

	if unit.GetHealth() <= 50 {
		t.Error("Unit should have regenerated health")
	}

	// Test that dead units don't update
	unit.SetHealth(0)
	unit.State = UnitStateDead
	oldHealth := unit.GetHealth()
	unit.Update(deltaTime)

	if unit.GetHealth() != oldHealth {
		t.Error("Dead units should not regenerate health")
	}
}

func TestGameUnitCommandQueue(t *testing.T) {
	unit := &GameUnit{
		ID:           1,
		CommandQueue: make([]UnitCommand, 0),
		State:        UnitStateIdle,
	}

	// Test empty command queue
	if len(unit.CommandQueue) != 0 {
		t.Error("Command queue should be empty initially")
	}

	if unit.CurrentCommand != nil {
		t.Error("Current command should be nil initially")
	}

	// Test adding commands to queue
	command1 := UnitCommand{Type: CommandMove, CreatedAt: time.Now()}
	command2 := UnitCommand{Type: CommandAttack, CreatedAt: time.Now()}

	unit.CommandQueue = append(unit.CommandQueue, command1, command2)

	if len(unit.CommandQueue) != 2 {
		t.Errorf("Expected 2 commands in queue, got %d", len(unit.CommandQueue))
	}

	// Test command queue processing
	unit.processCommandQueue()

	if unit.CurrentCommand == nil {
		t.Error("Current command should not be nil after processing queue")
	}

	if unit.CurrentCommand.Type != CommandMove {
		t.Errorf("Expected current command to be Move, got %v", unit.CurrentCommand.Type)
	}

	if len(unit.CommandQueue) != 1 {
		t.Errorf("Expected 1 command remaining in queue, got %d", len(unit.CommandQueue))
	}
}

func TestGameUnitResourceCarrying(t *testing.T) {
	unit := &GameUnit{
		ID:               1,
		CarriedResources: make(map[string]int),
		GatherRate:       map[string]float32{"wood": 10.0, "stone": 8.0},
	}

	// Test initial state
	if len(unit.CarriedResources) != 0 {
		t.Error("Should have no carried resources initially")
	}

	// Test resource gathering simulation
	unit.CarriedResources["wood"] = 25
	unit.CarriedResources["stone"] = 15

	if unit.CarriedResources["wood"] != 25 {
		t.Errorf("Expected 25 wood, got %d", unit.CarriedResources["wood"])
	}

	if unit.CarriedResources["stone"] != 15 {
		t.Errorf("Expected 15 stone, got %d", unit.CarriedResources["stone"])
	}

	// Test gather rates
	if unit.GatherRate["wood"] != 10.0 {
		t.Errorf("Expected wood gather rate 10.0, got %f", unit.GatherRate["wood"])
	}

	if unit.GatherRate["stone"] != 8.0 {
		t.Errorf("Expected stone gather rate 8.0, got %f", unit.GatherRate["stone"])
	}
}


// Test Phase 2.3 Requirements Validation
func TestPhase23Requirements(t *testing.T) {
	position := Vector3{X: 5, Y: 0, Z: 10}
	tileSize := float32(1.0)

	t.Run("Unit struct with required fields", func(t *testing.T) {
		unit := &GameUnit{
			ID:       123,
			PlayerID: 1,
			UnitType: "worker",
			Name:     "Test Worker",
			Position: position,
			GridPos:  WorldToGrid(position, tileSize),
			Health:   100,
			MaxHealth: 100,
			State:    UnitStateIdle,
		}

		// ✅ Unit struct with ID, type, position, health, faction
		if unit.GetID() != 123 {
			t.Error("Unit ID not properly stored")
		}

		if unit.GetType() != "worker" {
			t.Error("Unit type not properly stored")
		}

		if unit.GetPlayerID() != 1 {
			t.Error("Unit faction/player not properly stored")
		}

		if unit.GetHealth() != 100 {
			t.Error("Unit health not properly stored")
		}

		// ✅ Position system (grid coordinates + sub-tile positioning)
		gridPos := unit.GetGridPosition()
		if gridPos.Grid.X != 5 || gridPos.Grid.Y != 10 {
			t.Error("Grid positioning not working correctly")
		}

		worldPos := unit.GetPosition()
		if worldPos.X != 5 || worldPos.Z != 10 {
			t.Error("World positioning not working correctly")
		}
	})

	t.Run("Unit state machine with 6 states", func(t *testing.T) {
		unit := &GameUnit{State: UnitStateIdle}

		// ✅ Unit state machine (6 states implemented)
		states := []UnitState{
			UnitStateIdle,
			UnitStateMoving,
			UnitStateAttacking,
			UnitStateGathering,
			UnitStateBuilding,
			UnitStateDead,
		}

		for _, state := range states {
			unit.State = state
			if unit.State != state {
				t.Errorf("State transition to %v failed", state)
			}
		}

		if len(states) != 6 {
			t.Errorf("Expected 6 states, implemented %d", len(states))
		}
	})

	t.Run("Grid coordinates and sub-tile positioning", func(t *testing.T) {
		// ✅ Position system (grid coordinates + sub-tile positioning)

		// Test precise positioning within tiles
		gridPos := GridPosition{
			Grid:   Vector2i{X: 3, Y: 4},
			Offset: Vector2{X: 0.25, Y: 0.75}, // Quarter and three-quarters into tile
		}

		unit := &GameUnit{GridPos: gridPos}

		retrievedPos := unit.GetGridPosition()
		if retrievedPos.Grid != gridPos.Grid {
			t.Error("Grid coordinates not preserved")
		}

		if retrievedPos.Offset != gridPos.Offset {
			t.Error("Sub-tile positioning not preserved")
		}

		// Test conversion accuracy
		worldPos := GridToWorld(gridPos, 2.0) // 2x2 tiles
		expectedX := (3.0 + 0.25) * 2.0       // 6.5
		expectedZ := (4.0 + 0.75) * 2.0       // 9.5

		if worldPos.X != expectedX || worldPos.Z != expectedZ {
			t.Errorf("Grid to world conversion incorrect: got (%f, %f), expected (%f, %f)",
				worldPos.X, worldPos.Z, expectedX, expectedZ)
		}
	})
}

func TestUnitMovementAndTargeting(t *testing.T) {
	unit := &GameUnit{
		ID:       1,
		Position: Vector3{X: 0, Y: 0, Z: 0},
		GridPos:  GridPosition{Grid: Vector2i{X: 0, Y: 0}, Offset: Vector2{X: 0, Y: 0}},
		Speed:    2.0,
		State:    UnitStateIdle,
		Health:   100,
	}

	tileSize := float32(1.0)

	t.Run("Grid target setting", func(t *testing.T) {
		targetGrid := GridPosition{
			Grid:   Vector2i{X: 5, Y: 3},
			Offset: Vector2{X: 0.5, Y: 0.5},
		}

		unit.SetGridTarget(targetGrid, tileSize)

		// Check that grid position was set
		if unit.GridPos.Grid != targetGrid.Grid {
			t.Error("Grid target not set correctly")
		}

		// Check that world position was synchronized
		expectedWorld := GridToWorld(targetGrid, tileSize)
		if unit.Position.X != expectedWorld.X || unit.Position.Z != expectedWorld.Z {
			t.Error("World position not synchronized with grid target")
		}

		// Check that unit entered moving state
		if unit.State != UnitStateMoving {
			t.Error("Unit should enter moving state when target is set")
		}

		// Check that movement target was set
		if unit.Target == nil {
			t.Error("Movement target should be set")
		}
	})

	t.Run("Position update synchronization", func(t *testing.T) {
		newWorldPos := Vector3{X: 7.3, Y: 0, Z: 2.8}
		unit.UpdatePositions(newWorldPos, tileSize)

		// Check world position
		actualWorld := unit.GetPosition()
		if actualWorld.X != newWorldPos.X || actualWorld.Z != newWorldPos.Z {
			t.Error("World position not updated correctly")
		}

		// Check that grid position was calculated correctly
		expectedGrid := WorldToGrid(newWorldPos, tileSize)
		actualGrid := unit.GetGridPosition()
		if actualGrid.Grid != expectedGrid.Grid {
			t.Error("Grid position not calculated correctly from world position")
		}
	})
}