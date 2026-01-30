package engine

import (
	"fmt"
	"math"
	"testing"
)

func TestVector2i(t *testing.T) {
	t.Run("String representation", func(t *testing.T) {
		v := Vector2i{X: 5, Y: 10}
		expected := "(5, 10)"
		if v.String() != expected {
			t.Errorf("Expected %s, got %s", expected, v.String())
		}
	})

	t.Run("Add operation", func(t *testing.T) {
		v1 := Vector2i{X: 2, Y: 3}
		v2 := Vector2i{X: 4, Y: 5}
		result := v1.Add(v2)
		expected := Vector2i{X: 6, Y: 8}
		if result != expected {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})

	t.Run("Sub operation", func(t *testing.T) {
		v1 := Vector2i{X: 10, Y: 15}
		v2 := Vector2i{X: 3, Y: 7}
		result := v1.Sub(v2)
		expected := Vector2i{X: 7, Y: 8}
		if result != expected {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})

	t.Run("Manhattan distance", func(t *testing.T) {
		v1 := Vector2i{X: 0, Y: 0}
		v2 := Vector2i{X: 3, Y: 4}
		result := v1.Distance(v2)
		expected := 7 // |3-0| + |4-0| = 7
		if result != expected {
			t.Errorf("Expected %d, got %d", expected, result)
		}

		// Test negative coordinates
		v3 := Vector2i{X: -2, Y: -3}
		v4 := Vector2i{X: 1, Y: 2}
		result = v3.Distance(v4)
		expected = 8 // |1-(-2)| + |2-(-3)| = 3 + 5 = 8
		if result != expected {
			t.Errorf("Expected %d, got %d", expected, result)
		}
	})

	t.Run("Euclidean distance", func(t *testing.T) {
		v1 := Vector2i{X: 0, Y: 0}
		v2 := Vector2i{X: 3, Y: 4}
		result := v1.EuclideanDistance(v2)
		expected := 5.0 // sqrt(3^2 + 4^2) = 5
		if math.Abs(result-expected) > 0.0001 {
			t.Errorf("Expected %f, got %f", expected, result)
		}
	})
}

func TestVector2(t *testing.T) {
	t.Run("String representation", func(t *testing.T) {
		v := Vector2{X: 5.123, Y: 10.456}
		expected := "(5.123, 10.456)"
		if v.String() != expected {
			t.Errorf("Expected %s, got %s", expected, v.String())
		}
	})
}

func TestGridPosition(t *testing.T) {
	t.Run("String representation", func(t *testing.T) {
		gp := GridPosition{
			Grid:   Vector2i{X: 5, Y: 10},
			Offset: Vector2{X: 0.25, Y: 0.75},
		}
		expected := "Grid(5, 10)+(0.250, 0.750)"
		if gp.String() != expected {
			t.Errorf("Expected %s, got %s", expected, gp.String())
		}
	})

	t.Run("IsValid - valid position", func(t *testing.T) {
		gp := GridPosition{
			Grid:   Vector2i{X: 5, Y: 8},
			Offset: Vector2{X: 0.5, Y: 0.3},
		}
		if !gp.IsValid(10, 15) {
			t.Error("Expected position to be valid")
		}
	})

	t.Run("IsValid - grid out of bounds", func(t *testing.T) {
		testCases := []struct {
			name     string
			position GridPosition
			width    int
			height   int
		}{
			{
				name:     "negative X",
				position: GridPosition{Grid: Vector2i{X: -1, Y: 5}, Offset: Vector2{X: 0.5, Y: 0.5}},
				width:    10, height: 10,
			},
			{
				name:     "negative Y",
				position: GridPosition{Grid: Vector2i{X: 5, Y: -1}, Offset: Vector2{X: 0.5, Y: 0.5}},
				width:    10, height: 10,
			},
			{
				name:     "X >= width",
				position: GridPosition{Grid: Vector2i{X: 10, Y: 5}, Offset: Vector2{X: 0.5, Y: 0.5}},
				width:    10, height: 10,
			},
			{
				name:     "Y >= height",
				position: GridPosition{Grid: Vector2i{X: 5, Y: 10}, Offset: Vector2{X: 0.5, Y: 0.5}},
				width:    10, height: 10,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				if tc.position.IsValid(tc.width, tc.height) {
					t.Error("Expected position to be invalid")
				}
			})
		}
	})

	t.Run("IsValid - offset out of bounds", func(t *testing.T) {
		testCases := []struct {
			name   string
			offset Vector2
		}{
			{"negative X offset", Vector2{X: -0.1, Y: 0.5}},
			{"negative Y offset", Vector2{X: 0.5, Y: -0.1}},
			{"X offset > 1", Vector2{X: 1.1, Y: 0.5}},
			{"Y offset > 1", Vector2{X: 0.5, Y: 1.1}},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				gp := GridPosition{
					Grid:   Vector2i{X: 5, Y: 5},
					Offset: tc.offset,
				}
				if gp.IsValid(10, 10) {
					t.Error("Expected position to be invalid due to offset")
				}
			})
		}
	})
}

func TestWorldToGrid(t *testing.T) {
	tileSize := float32(2.0)

	testCases := []struct {
		name           string
		worldPos       Vector3
		expectedGrid   Vector2i
		expectedOffset Vector2
	}{
		{
			name:           "origin",
			worldPos:       Vector3{X: 0, Y: 0, Z: 0},
			expectedGrid:   Vector2i{X: 0, Y: 0},
			expectedOffset: Vector2{X: 0, Y: 0},
		},
		{
			name:           "center of tile 1,1",
			worldPos:       Vector3{X: 3, Y: 0, Z: 3},
			expectedGrid:   Vector2i{X: 1, Y: 1},
			expectedOffset: Vector2{X: 0.5, Y: 0.5},
		},
		{
			name:           "quarter into tile 2,3",
			worldPos:       Vector3{X: 4.5, Y: 0, Z: 6.5},
			expectedGrid:   Vector2i{X: 2, Y: 3},
			expectedOffset: Vector2{X: 0.25, Y: 0.25},
		},
		{
			name:           "edge of tile boundary",
			worldPos:       Vector3{X: 4, Y: 0, Z: 6},
			expectedGrid:   Vector2i{X: 2, Y: 3},
			expectedOffset: Vector2{X: 0, Y: 0},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := WorldToGrid(tc.worldPos, tileSize)

			if result.Grid != tc.expectedGrid {
				t.Errorf("Expected grid %v, got %v", tc.expectedGrid, result.Grid)
			}

			if math.Abs(result.Offset.X-tc.expectedOffset.X) > 0.0001 {
				t.Errorf("Expected offset X %f, got %f", tc.expectedOffset.X, result.Offset.X)
			}

			if math.Abs(result.Offset.Y-tc.expectedOffset.Y) > 0.0001 {
				t.Errorf("Expected offset Y %f, got %f", tc.expectedOffset.Y, result.Offset.Y)
			}
		})
	}
}

func TestGridToWorld(t *testing.T) {
	tileSize := float32(2.0)

	testCases := []struct {
		name        string
		gridPos     GridPosition
		expectedPos Vector3
	}{
		{
			name:        "origin",
			gridPos:     GridPosition{Grid: Vector2i{X: 0, Y: 0}, Offset: Vector2{X: 0, Y: 0}},
			expectedPos: Vector3{X: 0, Y: 0, Z: 0},
		},
		{
			name:        "center of tile 1,1",
			gridPos:     GridPosition{Grid: Vector2i{X: 1, Y: 1}, Offset: Vector2{X: 0.5, Y: 0.5}},
			expectedPos: Vector3{X: 3, Y: 0, Z: 3},
		},
		{
			name:        "quarter into tile 2,3",
			gridPos:     GridPosition{Grid: Vector2i{X: 2, Y: 3}, Offset: Vector2{X: 0.25, Y: 0.25}},
			expectedPos: Vector3{X: 4.5, Y: 0, Z: 6.5},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := GridToWorld(tc.gridPos, tileSize)

			if math.Abs(result.X-tc.expectedPos.X) > 0.0001 {
				t.Errorf("Expected X %f, got %f", tc.expectedPos.X, result.X)
			}

			if math.Abs(result.Y-tc.expectedPos.Y) > 0.0001 {
				t.Errorf("Expected Y %f, got %f", tc.expectedPos.Y, result.Y)
			}

			if math.Abs(result.Z-tc.expectedPos.Z) > 0.0001 {
				t.Errorf("Expected Z %f, got %f", tc.expectedPos.Z, result.Z)
			}
		})
	}
}

func TestRoundTripConversion(t *testing.T) {
	// Test that converting world -> grid -> world preserves position (within tolerance)
	tileSize := float32(1.5)

	testPositions := []Vector3{
		{X: 0, Y: 0, Z: 0},
		{X: 5.7, Y: 1.2, Z: 8.3},
		{X: 12.45, Y: 0, Z: 3.78},
		{X: 0.1, Y: 0, Z: 0.1},
	}

	for i, pos := range testPositions {
		t.Run(fmt.Sprintf("position_%d", i), func(t *testing.T) {
			// Convert to grid and back
			gridPos := WorldToGrid(pos, tileSize)
			backToWorld := GridToWorld(gridPos, tileSize)

			// Check X coordinate (tolerance due to floating point precision)
			if math.Abs(pos.X-backToWorld.X) > 0.0001 {
				t.Errorf("X coordinate changed: %f -> %f", pos.X, backToWorld.X)
			}

			// Check Z coordinate (Y is always set to 0 in conversion)
			if math.Abs(pos.Z-backToWorld.Z) > 0.0001 {
				t.Errorf("Z coordinate changed: %f -> %f", pos.Z, backToWorld.Z)
			}

			// Y should always be 0 after conversion
			if backToWorld.Y != 0 {
				t.Errorf("Y should be 0 after grid conversion, got %f", backToWorld.Y)
			}
		})
	}
}

func TestIsValidGridPosition(t *testing.T) {
	worldWidth, worldHeight := 10, 8

	testCases := []struct {
		name     string
		pos      Vector2i
		expected bool
	}{
		{"valid center", Vector2i{X: 5, Y: 4}, true},
		{"valid origin", Vector2i{X: 0, Y: 0}, true},
		{"valid edge", Vector2i{X: 9, Y: 7}, true},
		{"invalid negative X", Vector2i{X: -1, Y: 4}, false},
		{"invalid negative Y", Vector2i{X: 5, Y: -1}, false},
		{"invalid X >= width", Vector2i{X: 10, Y: 4}, false},
		{"invalid Y >= height", Vector2i{X: 5, Y: 8}, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := IsValidGridPosition(tc.pos, worldWidth, worldHeight)
			if result != tc.expected {
				t.Errorf("Expected %v, got %v for position %v", tc.expected, result, tc.pos)
			}
		})
	}
}

func TestClampToWorldBounds(t *testing.T) {
	worldWidth, worldHeight := 10, 8

	testCases := []struct {
		name     string
		pos      Vector2i
		expected Vector2i
	}{
		{"already valid", Vector2i{X: 5, Y: 4}, Vector2i{X: 5, Y: 4}},
		{"clamp negative X", Vector2i{X: -3, Y: 4}, Vector2i{X: 0, Y: 4}},
		{"clamp negative Y", Vector2i{X: 5, Y: -2}, Vector2i{X: 5, Y: 0}},
		{"clamp large X", Vector2i{X: 15, Y: 4}, Vector2i{X: 9, Y: 4}},
		{"clamp large Y", Vector2i{X: 5, Y: 12}, Vector2i{X: 5, Y: 7}},
		{"clamp both negative", Vector2i{X: -1, Y: -1}, Vector2i{X: 0, Y: 0}},
		{"clamp both large", Vector2i{X: 20, Y: 20}, Vector2i{X: 9, Y: 7}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ClampToWorldBounds(tc.pos, worldWidth, worldHeight)
			if result != tc.expected {
				t.Errorf("Expected %v, got %v for position %v", tc.expected, result, tc.pos)
			}
		})
	}
}

func TestGetNeighbors(t *testing.T) {
	pos := Vector2i{X: 5, Y: 5}
	neighbors := GetNeighbors(pos)

	expected := []Vector2i{
		{X: 4, Y: 4}, // Top-left
		{X: 5, Y: 4}, // Top
		{X: 6, Y: 4}, // Top-right
		{X: 4, Y: 5}, // Left
		{X: 6, Y: 5}, // Right
		{X: 4, Y: 6}, // Bottom-left
		{X: 5, Y: 6}, // Bottom
		{X: 6, Y: 6}, // Bottom-right
	}

	if len(neighbors) != 8 {
		t.Errorf("Expected 8 neighbors, got %d", len(neighbors))
	}

	for i, expectedNeighbor := range expected {
		if neighbors[i] != expectedNeighbor {
			t.Errorf("Neighbor %d: expected %v, got %v", i, expectedNeighbor, neighbors[i])
		}
	}
}

func TestGetCardinalNeighbors(t *testing.T) {
	pos := Vector2i{X: 5, Y: 5}
	neighbors := GetCardinalNeighbors(pos)

	expected := []Vector2i{
		{X: 5, Y: 4}, // North
		{X: 6, Y: 5}, // East
		{X: 5, Y: 6}, // South
		{X: 4, Y: 5}, // West
	}

	if len(neighbors) != 4 {
		t.Errorf("Expected 4 neighbors, got %d", len(neighbors))
	}

	for i, expectedNeighbor := range expected {
		if neighbors[i] != expectedNeighbor {
			t.Errorf("Cardinal neighbor %d: expected %v, got %v", i, expectedNeighbor, neighbors[i])
		}
	}
}

func TestCalculateGridDistanceFloat(t *testing.T) {
	pos1 := GridPosition{Grid: Vector2i{X: 0, Y: 0}, Offset: Vector2{X: 0, Y: 0}}
	pos2 := GridPosition{Grid: Vector2i{X: 3, Y: 4}, Offset: Vector2{X: 0, Y: 0}}

	result := CalculateGridDistanceFloat(pos1, pos2)
	expected := 5.0 // sqrt(3^2 + 4^2) = 5

	if math.Abs(result-expected) > 0.0001 {
		t.Errorf("Expected %f, got %f", expected, result)
	}
}

func TestGetGridCenter(t *testing.T) {
	gridPos := Vector2i{X: 3, Y: 4}
	tileSize := float32(2.0)

	result := GetGridCenter(gridPos, tileSize)
	expected := Vector3{X: 7, Y: 0, Z: 9} // (3+0.5)*2 = 7, (4+0.5)*2 = 9

	if math.Abs(result.X-expected.X) > 0.0001 {
		t.Errorf("Expected X %f, got %f", expected.X, result.X)
	}

	if result.Y != expected.Y {
		t.Errorf("Expected Y %f, got %f", expected.Y, result.Y)
	}

	if math.Abs(result.Z-expected.Z) > 0.0001 {
		t.Errorf("Expected Z %f, got %f", expected.Z, result.Z)
	}
}

func TestSnapToGrid(t *testing.T) {
	worldPos := Vector3{X: 5.7, Y: 1.2, Z: 8.3}
	tileSize := float32(2.0)

	result := SnapToGrid(worldPos, tileSize)

	// Should snap to center of grid tile (2,4) -> world pos (5,0,9)
	expected := Vector3{X: 5, Y: 0, Z: 9}

	if math.Abs(result.X-expected.X) > 0.0001 {
		t.Errorf("Expected X %f, got %f", expected.X, result.X)
	}

	if result.Y != expected.Y {
		t.Errorf("Expected Y %f, got %f", expected.Y, result.Y)
	}

	if math.Abs(result.Z-expected.Z) > 0.0001 {
		t.Errorf("Expected Z %f, got %f", expected.Z, result.Z)
	}
}