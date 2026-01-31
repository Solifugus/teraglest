package engine

import (
	"math"
	"testing"
	"time"

	"teraglest/internal/data"
)

// TestPathfindingBasic tests basic A* pathfinding functionality
func TestPathfindingBasic(t *testing.T) {
	// Create test world
	world, err := createTestWorldForPathfinding(t)
	if err != nil {
		t.Fatalf("Failed to create test world: %v", err)
	}

	pathfinder := NewPathfinder(world)

	// Test simple straight-line path
	request := PathRequest{
		Start:        GridPosition{Grid: Vector2i{X: 0, Y: 0}, Offset: Vector2{X: 0.5, Y: 0.5}},
		Target:       GridPosition{Grid: Vector2i{X: 5, Y: 0}, Offset: Vector2{X: 0.5, Y: 0.5}},
		UnitSize:     1,
		MaxRange:     0,
		AllowPartial: false,
	}

	result := pathfinder.FindPath(request)

	if !result.Success {
		t.Fatal("Expected pathfinding to succeed for simple straight path")
	}

	if len(result.Path) != 6 { // 0,1,2,3,4,5
		t.Errorf("Expected path length 6, got %d", len(result.Path))
	}

	// Verify path goes through expected coordinates
	expectedX := []float32{0, 1, 2, 3, 4, 5}
	for i, waypoint := range result.Path {
		if int(waypoint.X) != int(expectedX[i]) || int(waypoint.Z) != 0 {
			t.Errorf("Waypoint %d: expected (%v,0), got (%v,%v)",
				i, expectedX[i], waypoint.X, waypoint.Z)
		}
	}
}

// TestPathfindingWithObstacles tests pathfinding around obstacles
func TestPathfindingWithObstacles(t *testing.T) {
	world, err := createTestWorldForPathfinding(t)
	if err != nil {
		t.Fatalf("Failed to create test world: %v", err)
	}

	// Create obstacles (walls) in the middle
	for y := 1; y <= 3; y++ {
		world.SetOccupied(Vector2i{X: 2, Y: y}, true)
		world.SetWalkable(Vector2i{X: 2, Y: y}, false)
	}

	pathfinder := NewPathfinder(world)

	// Test pathfinding around obstacle
	request := PathRequest{
		Start:        GridPosition{Grid: Vector2i{X: 0, Y: 2}, Offset: Vector2{X: 0.5, Y: 0.5}},
		Target:       GridPosition{Grid: Vector2i{X: 4, Y: 2}, Offset: Vector2{X: 0.5, Y: 0.5}},
		UnitSize:     1,
		MaxRange:     0,
		AllowPartial: false,
	}

	result := pathfinder.FindPath(request)

	if !result.Success {
		t.Fatal("Expected pathfinding to succeed around obstacle")
	}

	// Path should go around the obstacle, so should be longer than straight line
	if len(result.Path) < 5 {
		t.Errorf("Expected path to go around obstacle (length >= 5), got %d", len(result.Path))
	}

	// Verify path doesn't go through obstacles
	for i, waypoint := range result.Path {
		gridX := int(waypoint.X)
		gridY := int(waypoint.Z)
		if gridX == 2 && gridY >= 1 && gridY <= 3 {
			t.Errorf("Path waypoint %d goes through obstacle at (%d,%d)", i, gridX, gridY)
		}
	}
}

// TestPathfindingNoPath tests when no path exists
func TestPathfindingNoPath(t *testing.T) {
	world, err := createTestWorldForPathfinding(t)
	if err != nil {
		t.Fatalf("Failed to create test world: %v", err)
	}

	// Create complete wall blocking path
	for x := 0; x < world.Width; x++ {
		world.SetOccupied(Vector2i{X: x, Y: 2}, true)
		world.SetWalkable(Vector2i{X: x, Y: 2}, false)
	}

	pathfinder := NewPathfinder(world)

	// Try to path across the wall
	request := PathRequest{
		Start:        GridPosition{Grid: Vector2i{X: 0, Y: 0}, Offset: Vector2{X: 0.5, Y: 0.5}},
		Target:       GridPosition{Grid: Vector2i{X: 0, Y: 4}, Offset: Vector2{X: 0.5, Y: 0.5}},
		UnitSize:     1,
		MaxRange:     0,
		AllowPartial: false,
	}

	result := pathfinder.FindPath(request)

	if result.Success {
		t.Error("Expected pathfinding to fail when no path exists")
	}
}

// TestPathfindingPartialPath tests partial path functionality
func TestPathfindingPartialPath(t *testing.T) {
	world, err := createTestWorldForPathfinding(t)
	if err != nil {
		t.Fatalf("Failed to create test world: %v", err)
	}

	// Create wall that blocks complete path but allows partial progress
	for x := 3; x < world.Width; x++ {
		world.SetOccupied(Vector2i{X: x, Y: 2}, true)
		world.SetWalkable(Vector2i{X: x, Y: 2}, false)
	}

	pathfinder := NewPathfinder(world)

	// Try to path beyond the wall with partial path allowed
	request := PathRequest{
		Start:        GridPosition{Grid: Vector2i{X: 0, Y: 2}, Offset: Vector2{X: 0.5, Y: 0.5}},
		Target:       GridPosition{Grid: Vector2i{X: 7, Y: 2}, Offset: Vector2{X: 0.5, Y: 0.5}},
		UnitSize:     1,
		MaxRange:     0,
		AllowPartial: true,
	}

	result := pathfinder.FindPath(request)

	if !result.Success {
		t.Fatal("Expected pathfinding to succeed with partial path")
	}

	if !result.Partial {
		t.Error("Expected result to be marked as partial")
	}

	// Should reach as close as possible (position 2,2)
	finalPos := result.Path[len(result.Path)-1]
	if int(finalPos.X) >= 3 {
		t.Errorf("Partial path should stop before obstacle, final X: %v", finalPos.X)
	}
}

// TestPathfindingWithRange tests range-limited pathfinding
func TestPathfindingWithRange(t *testing.T) {
	world, err := createTestWorldForPathfinding(t)
	if err != nil {
		t.Fatalf("Failed to create test world: %v", err)
	}

	pathfinder := NewPathfinder(world)

	// Test with limited range
	request := PathRequest{
		Start:        GridPosition{Grid: Vector2i{X: 0, Y: 0}, Offset: Vector2{X: 0.5, Y: 0.5}},
		Target:       GridPosition{Grid: Vector2i{X: 10, Y: 0}, Offset: Vector2{X: 0.5, Y: 0.5}},
		UnitSize:     1,
		MaxRange:     3.0, // Limit search to 3 tiles
		AllowPartial: true,
	}

	result := pathfinder.FindPath(request)

	if !result.Success {
		t.Fatal("Expected pathfinding to succeed with range limit")
	}

	// Should not reach the full target due to range limit
	finalPos := result.Path[len(result.Path)-1]
	if int(finalPos.X) > 3 {
		t.Errorf("Range-limited path exceeded range, final X: %v", finalPos.X)
	}
}

// TestPathfindingManager tests the higher-level PathfindingManager
func TestPathfindingManager(t *testing.T) {
	world, err := createTestWorldForPathfinding(t)
	if err != nil {
		t.Fatalf("Failed to create test world: %v", err)
	}

	pathfindingMgr := NewPathfindingManager(world)

	// Create a test unit
	unit := &GameUnit{
		ID:       1,
		PlayerID: 0,
		Position: Vector3{X: 0, Y: 0, Z: 0},
		Speed:    2.0,
	}

	// Request path for unit
	targetPos := Vector3{X: 5, Y: 0, Z: 3}
	result, err := pathfindingMgr.RequestPath(unit, targetPos)

	if err != nil {
		t.Fatalf("PathfindingManager failed: %v", err)
	}

	if !result.Success {
		t.Fatal("Expected PathfindingManager to succeed")
	}

	if len(result.Path) == 0 {
		t.Error("Expected non-empty path from PathfindingManager")
	}

	// Verify path starts near unit position and ends near target
	startPos := result.Path[0]
	endPos := result.Path[len(result.Path)-1]

	if int(startPos.X) != 0 || int(startPos.Z) != 0 {
		t.Errorf("Path should start at unit position, got (%v,%v)", startPos.X, startPos.Z)
	}

	if int(endPos.X) != 5 || int(endPos.Z) != 3 {
		t.Errorf("Path should end at target position, got (%v,%v)", endPos.X, endPos.Z)
	}
}

// TestDiagonalMovement tests pathfinding with diagonal movement
func TestDiagonalMovement(t *testing.T) {
	world, err := createTestWorldForPathfinding(t)
	if err != nil {
		t.Fatalf("Failed to create test world: %v", err)
	}

	pathfinder := NewPathfinder(world)

	// Test diagonal path
	request := PathRequest{
		Start:        GridPosition{Grid: Vector2i{X: 0, Y: 0}, Offset: Vector2{X: 0.5, Y: 0.5}},
		Target:       GridPosition{Grid: Vector2i{X: 3, Y: 3}, Offset: Vector2{X: 0.5, Y: 0.5}},
		UnitSize:     1,
		MaxRange:     0,
		AllowPartial: false,
	}

	result := pathfinder.FindPath(request)

	if !result.Success {
		t.Fatal("Expected diagonal pathfinding to succeed")
	}

	// Should prefer diagonal movement for efficiency
	if len(result.Path) > 4 { // Optimal diagonal path should be 4 steps
		t.Errorf("Diagonal path seems inefficient, length: %d", len(result.Path))
	}
}

// TestTerrainCosts tests pathfinding with different terrain movement costs
func TestTerrainCosts(t *testing.T) {
	world, err := createTestWorldForPathfinding(t)
	if err != nil {
		t.Fatalf("Failed to create test world: %v", err)
	}

	// Set some tiles to expensive terrain (water = type 2)
	for x := 1; x <= 3; x++ {
		world.TerrainMap.TerrainData[1][x] = 2 // Water terrain
	}

	pathfinder := NewPathfinder(world)

	// Test path that could go through expensive terrain or around it
	request := PathRequest{
		Start:        GridPosition{Grid: Vector2i{X: 0, Y: 1}, Offset: Vector2{X: 0.5, Y: 0.5}},
		Target:       GridPosition{Grid: Vector2i{X: 4, Y: 1}, Offset: Vector2{X: 0.5, Y: 0.5}},
		UnitSize:     1,
		MaxRange:     0,
		AllowPartial: false,
	}

	result := pathfinder.FindPath(request)

	if !result.Success {
		t.Fatal("Expected pathfinding with terrain costs to succeed")
	}

	// Path might avoid expensive terrain by going around
	// This is hard to verify exactly, but the path should exist
	if len(result.Path) == 0 {
		t.Error("Expected non-empty path with terrain costs")
	}
}

// Helper function to create a test world for pathfinding
func createTestWorldForPathfinding(t *testing.T) (*World, error) {
	// Create minimal world for testing
	settings := GameSettings{
		PlayerFactions: map[int]string{0: "tech"},
		AIFactions:     map[int]string{},
	}

	// Create a simple tech tree
	techTree := &data.TechTree{
		Description: data.TechTreeDescription{Value: "test"},
	}

	// Create asset manager
	assetMgr := data.NewAssetManager("")

	// Create world
	world := &World{
		settings:      settings,
		techTree:      techTree,
		assetMgr:      assetMgr,
		players:       make(map[int]*Player),
		resources:     make(map[int]*ResourceNode),
		nextEntityID:  1,
		Width:         10,
		Height:        10,
		tileSize:      1.0,
		resourceGenerationRate: make(map[string]float32),
		unitCap:       200,
		buildingCap:   50,
	}

	// Initialize ObjectManager
	world.ObjectManager = NewObjectManager(world)

	// Initialize CommandProcessor
	world.commandProcessor = NewCommandProcessor(world)

	// Initialize PathfindingManager
	world.pathfindingMgr = NewPathfindingManager(world)

	// Initialize grid system
	if err := world.initializeGrid(); err != nil {
		return nil, err
	}

	return world, nil
}

// Benchmark pathfinding performance
func BenchmarkPathfinding(b *testing.B) {
	world, err := createTestWorldForPathfinding(nil)
	if err != nil {
		b.Fatalf("Failed to create test world: %v", err)
	}

	pathfinder := NewPathfinder(world)

	request := PathRequest{
		Start:        GridPosition{Grid: Vector2i{X: 0, Y: 0}, Offset: Vector2{X: 0.5, Y: 0.5}},
		Target:       GridPosition{Grid: Vector2i{X: 9, Y: 9}, Offset: Vector2{X: 0.5, Y: 0.5}},
		UnitSize:     1,
		MaxRange:     0,
		AllowPartial: false,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := pathfinder.FindPath(request)
		if !result.Success {
			b.Fatal("Pathfinding failed in benchmark")
		}
	}
}

// TestIntegrationWithMovementCommands tests pathfinding integration with movement commands
func TestIntegrationWithMovementCommands(t *testing.T) {
	world, err := createTestWorldForPathfinding(t)
	if err != nil {
		t.Fatalf("Failed to create test world: %v", err)
	}

	// Create a test player and unit
	player := &Player{
		ID:          0,
		Name:        "TestPlayer",
		FactionName: "tech",
		Resources:   make(map[string]int),
		IsActive:    true,
	}
	world.players[0] = player

	// Create test unit
	startPos := Vector3{X: 0, Y: 0, Z: 0}
	unit, err := world.ObjectManager.CreateUnit(0, "test_unit", startPos, nil)
	if err != nil {
		t.Fatalf("Failed to create test unit: %v", err)
	}

	// Issue move command
	targetPos := Vector3{X: 5, Y: 0, Z: 3}
	command := CreateMoveCommand(targetPos, false)

	err = world.commandProcessor.IssueCommand(unit.ID, command)
	if err != nil {
		t.Fatalf("Failed to issue move command: %v", err)
	}

	// Process command for a few frames to see if pathfinding works
	deltaTime := 16 * time.Millisecond // ~60 FPS

	for i := 0; i < 10; i++ {
		world.commandProcessor.ProcessCommand(unit, unit.CurrentCommand, deltaTime)

		// Check if unit has received a path
		if unit.Path != nil && len(unit.Path) > 0 {
			// Path was successfully computed
			break
		}

		if i == 9 {
			t.Error("Unit should have received a path from A* pathfinding")
		}
	}

	// Verify the unit has a valid path
	if unit.Path == nil || len(unit.Path) == 0 {
		t.Error("Expected unit to have computed path")
	}

	// Verify path starts near current position and ends near target
	if len(unit.Path) > 0 {
		startPath := unit.Path[0]
		endPath := unit.Path[len(unit.Path)-1]

		startDistance := calculateDistance3D(startPath, unit.Position)
		if startDistance > 2.0 { // Allow some tolerance
			t.Errorf("Path start too far from unit position: distance %v", startDistance)
		}

		endDistance := calculateDistance3D(endPath, targetPos)
		if endDistance > 2.0 { // Allow some tolerance
			t.Errorf("Path end too far from target: distance %v", endDistance)
		}
	}
}

// Helper function for 3D distance calculation
func calculateDistance3D(a, b Vector3) float32 {
	dx := a.X - b.X
	dy := a.Y - b.Y
	dz := a.Z - b.Z
	return float32(math.Sqrt(float64(dx*dx + dy*dy + dz*dz)))
}