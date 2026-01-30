package engine

import (
	"fmt"
	"math"
)

// Vector2i represents integer-based 2D coordinates for grid positioning
type Vector2i struct {
	X, Y int
}

// String returns the string representation of Vector2i
func (v Vector2i) String() string {
	return fmt.Sprintf("(%d, %d)", v.X, v.Y)
}

// Add returns the sum of two Vector2i coordinates
func (v Vector2i) Add(other Vector2i) Vector2i {
	return Vector2i{X: v.X + other.X, Y: v.Y + other.Y}
}

// Sub returns the difference between two Vector2i coordinates
func (v Vector2i) Sub(other Vector2i) Vector2i {
	return Vector2i{X: v.X - other.X, Y: v.Y - other.Y}
}

// Distance returns the Manhattan distance between two grid positions
func (v Vector2i) Distance(other Vector2i) int {
	dx := v.X - other.X
	dy := v.Y - other.Y
	if dx < 0 {
		dx = -dx
	}
	if dy < 0 {
		dy = -dy
	}
	return dx + dy
}

// EuclideanDistance returns the Euclidean distance between two grid positions
func (v Vector2i) EuclideanDistance(other Vector2i) float64 {
	dx := float64(v.X - other.X)
	dy := float64(v.Y - other.Y)
	return math.Sqrt(dx*dx + dy*dy)
}

// Vector2 represents floating-point 2D coordinates for sub-tile positioning
type Vector2 struct {
	X, Y float64
}

// String returns the string representation of Vector2
func (v Vector2) String() string {
	return fmt.Sprintf("(%.3f, %.3f)", v.X, v.Y)
}

// GridPosition represents a position in the grid-based coordinate system
// It combines grid tile coordinates with sub-tile positioning for smooth movement
type GridPosition struct {
	Grid   Vector2i `json:"grid"`   // Grid tile coordinates (integer)
	Offset Vector2  `json:"offset"` // Sub-tile offset within the grid cell (0.0-1.0)
}

// String returns the string representation of GridPosition
func (gp GridPosition) String() string {
	return fmt.Sprintf("Grid%s+%s", gp.Grid.String(), gp.Offset.String())
}

// IsValid checks if the grid position is within bounds and has valid offsets
func (gp GridPosition) IsValid(worldWidth, worldHeight int) bool {
	// Check grid bounds
	if gp.Grid.X < 0 || gp.Grid.Y < 0 || gp.Grid.X >= worldWidth || gp.Grid.Y >= worldHeight {
		return false
	}

	// Check offset bounds (should be 0.0-1.0)
	if gp.Offset.X < 0.0 || gp.Offset.X > 1.0 || gp.Offset.Y < 0.0 || gp.Offset.Y > 1.0 {
		return false
	}

	return true
}

// WorldToGrid converts a world position (Vector3) to grid coordinates
func WorldToGrid(worldPos Vector3, tileSize float32) GridPosition {
	// Calculate grid tile coordinates (floor division)
	gridX := int(math.Floor(worldPos.X / float64(tileSize)))
	gridY := int(math.Floor(worldPos.Z / float64(tileSize))) // Using Z for 2D grid

	// Calculate sub-tile offset (0.0-1.0)
	offsetX := (worldPos.X / float64(tileSize)) - float64(gridX)
	offsetY := (worldPos.Z / float64(tileSize)) - float64(gridY)

	// Clamp offsets to valid range
	if offsetX < 0.0 {
		offsetX = 0.0
	}
	if offsetX > 1.0 {
		offsetX = 1.0
	}
	if offsetY < 0.0 {
		offsetY = 0.0
	}
	if offsetY > 1.0 {
		offsetY = 1.0
	}

	return GridPosition{
		Grid:   Vector2i{X: gridX, Y: gridY},
		Offset: Vector2{X: offsetX, Y: offsetY},
	}
}

// GridToWorld converts grid coordinates to world position (Vector3)
func GridToWorld(gridPos GridPosition, tileSize float32) Vector3 {
	// Calculate world position from grid coordinates and offset
	worldX := (float64(gridPos.Grid.X) + gridPos.Offset.X) * float64(tileSize)
	worldZ := (float64(gridPos.Grid.Y) + gridPos.Offset.Y) * float64(tileSize)

	// Y coordinate remains unchanged (height will be handled by terrain system later)
	return Vector3{
		X: worldX,
		Y: 0.0, // Default to ground level
		Z: worldZ,
	}
}

// IsValidGridPosition checks if a grid position is within world bounds
func IsValidGridPosition(pos Vector2i, worldWidth, worldHeight int) bool {
	return pos.X >= 0 && pos.Y >= 0 && pos.X < worldWidth && pos.Y < worldHeight
}

// ClampToWorldBounds clamps a grid position to world boundaries
func ClampToWorldBounds(pos Vector2i, worldWidth, worldHeight int) Vector2i {
	clampedX := pos.X
	clampedY := pos.Y

	if clampedX < 0 {
		clampedX = 0
	}
	if clampedX >= worldWidth {
		clampedX = worldWidth - 1
	}
	if clampedY < 0 {
		clampedY = 0
	}
	if clampedY >= worldHeight {
		clampedY = worldHeight - 1
	}

	return Vector2i{X: clampedX, Y: clampedY}
}

// GetNeighbors returns the 8 neighboring grid positions (including diagonals)
func GetNeighbors(pos Vector2i) []Vector2i {
	return []Vector2i{
		{X: pos.X - 1, Y: pos.Y - 1}, // Top-left
		{X: pos.X, Y: pos.Y - 1},     // Top
		{X: pos.X + 1, Y: pos.Y - 1}, // Top-right
		{X: pos.X - 1, Y: pos.Y},     // Left
		{X: pos.X + 1, Y: pos.Y},     // Right
		{X: pos.X - 1, Y: pos.Y + 1}, // Bottom-left
		{X: pos.X, Y: pos.Y + 1},     // Bottom
		{X: pos.X + 1, Y: pos.Y + 1}, // Bottom-right
	}
}

// GetCardinalNeighbors returns the 4 cardinal neighboring grid positions (no diagonals)
func GetCardinalNeighbors(pos Vector2i) []Vector2i {
	return []Vector2i{
		{X: pos.X, Y: pos.Y - 1}, // North
		{X: pos.X + 1, Y: pos.Y}, // East
		{X: pos.X, Y: pos.Y + 1}, // South
		{X: pos.X - 1, Y: pos.Y}, // West
	}
}

// CalculateGridDistance calculates the grid distance between two positions
func CalculateGridDistance(pos1, pos2 Vector2i) int {
	return pos1.Distance(pos2)
}

// CalculateGridDistanceFloat calculates the precise distance between two grid positions
func CalculateGridDistanceFloat(pos1, pos2 GridPosition) float64 {
	// Convert to world coordinates for precise calculation
	world1 := GridToWorld(pos1, 1.0) // Use unit tile size for calculation
	world2 := GridToWorld(pos2, 1.0)

	dx := world1.X - world2.X
	dz := world1.Z - world2.Z

	return math.Sqrt(dx*dx + dz*dz)
}

// GetGridCenter returns the center position of a grid tile in world coordinates
func GetGridCenter(gridPos Vector2i, tileSize float32) Vector3 {
	centerGrid := GridPosition{
		Grid:   gridPos,
		Offset: Vector2{X: 0.5, Y: 0.5}, // Center of tile
	}
	return GridToWorld(centerGrid, tileSize)
}

// SnapToGrid snaps a world position to the nearest grid center
func SnapToGrid(worldPos Vector3, tileSize float32) Vector3 {
	gridPos := WorldToGrid(worldPos, tileSize)
	return GetGridCenter(gridPos.Grid, tileSize)
}