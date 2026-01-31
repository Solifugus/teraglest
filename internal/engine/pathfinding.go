package engine

import (
	"container/heap"
	"fmt"
	"math"
)

// PathNode represents a node in the A* pathfinding algorithm
type PathNode struct {
	X, Y     int     // Grid coordinates
	GCost    float32 // Distance from starting node
	HCost    float32 // Heuristic distance to target
	FCost    float32 // GCost + HCost (total cost)
	Parent   *PathNode
	HeapIndex int // Index in the priority queue
}

// PathNodeHeap implements a priority queue for A* pathfinding
type PathNodeHeap []*PathNode

func (h PathNodeHeap) Len() int           { return len(h) }
func (h PathNodeHeap) Less(i, j int) bool { return h[i].FCost < h[j].FCost }
func (h PathNodeHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].HeapIndex = i
	h[j].HeapIndex = j
}

func (h *PathNodeHeap) Push(x interface{}) {
	n := len(*h)
	node := x.(*PathNode)
	node.HeapIndex = n
	*h = append(*h, node)
}

func (h *PathNodeHeap) Pop() interface{} {
	old := *h
	n := len(old)
	node := old[n-1]
	node.HeapIndex = -1
	*h = old[0 : n-1]
	return node
}

// PathRequest represents a request for pathfinding
type PathRequest struct {
	Start      GridPosition
	Target     GridPosition
	UnitSize   int     // Size of the unit (for collision detection)
	MaxRange   float32 // Maximum search range (0 = unlimited)
	AllowPartial bool  // Allow partial paths when target unreachable
}

// PathResult contains the result of pathfinding
type PathResult struct {
	Success   bool
	Path      []Vector3     // World space path coordinates
	GridPath  []GridPosition // Grid space path coordinates
	Distance  float32        // Total path distance
	Partial   bool           // True if this is a partial path
}

// Pathfinder handles A* pathfinding for units
type Pathfinder struct {
	world       *World
	nodePool    []*PathNode    // Pre-allocated node pool for performance
	openSet     PathNodeHeap   // Priority queue for open nodes
	closedSet   map[int]*PathNode // Closed nodes (using packed coordinates as key)
}

// NewPathfinder creates a new pathfinding system
func NewPathfinder(world *World) *Pathfinder {
	return &Pathfinder{
		world:     world,
		nodePool:  make([]*PathNode, 0, 1000), // Pre-allocate for performance
		closedSet: make(map[int]*PathNode, 1000),
	}
}

// FindPath computes an optimal path using A* algorithm
func (pf *Pathfinder) FindPath(request PathRequest) PathResult {
	// Reset pathfinder state
	pf.reset()

	// Validate request
	if !pf.isValidPosition(request.Start) || !pf.isValidPosition(request.Target) {
		return PathResult{Success: false}
	}

	// If already at target, return empty path
	if request.Start.Grid.X == request.Target.Grid.X && request.Start.Grid.Y == request.Target.Grid.Y {
		return PathResult{
			Success: true,
			Path:    []Vector3{pf.gridToWorld(request.Target)},
			GridPath: []GridPosition{request.Target},
		}
	}

	// Initialize starting node
	startNode := pf.getNode(request.Start.Grid.X, request.Start.Grid.Y)
	startNode.GCost = 0
	startNode.HCost = pf.heuristic(request.Start, request.Target)
	startNode.FCost = startNode.GCost + startNode.HCost
	startNode.Parent = nil

	heap.Push(&pf.openSet, startNode)

	// A* main loop
	maxIterations := 10000 // Prevent infinite loops
	iterations := 0

	for len(pf.openSet) > 0 && iterations < maxIterations {
		iterations++

		// Get node with lowest F cost
		currentNode := heap.Pop(&pf.openSet).(*PathNode)
		pf.closedSet[pf.packCoordinates(currentNode.X, currentNode.Y)] = currentNode

		// Check if we reached the target
		if currentNode.X == request.Target.Grid.X && currentNode.Y == request.Target.Grid.Y {
			return pf.reconstructPath(currentNode, request)
		}

		// Check range limit
		if request.MaxRange > 0 && currentNode.GCost > request.MaxRange {
			continue
		}

		// Explore neighbors
		pf.exploreNeighbors(currentNode, request)
	}

	// No path found - try to return partial path if allowed
	if request.AllowPartial {
		return pf.findPartialPath(request)
	}

	return PathResult{Success: false}
}

// exploreNeighbors examines all valid neighboring nodes
func (pf *Pathfinder) exploreNeighbors(currentNode *PathNode, request PathRequest) {
	// 8-directional movement (including diagonals)
	directions := []struct{ dx, dy int }{
		{-1, -1}, {0, -1}, {1, -1}, // Top row
		{-1, 0}, {1, 0},            // Middle row (skip center)
		{-1, 1}, {0, 1}, {1, 1},    // Bottom row
	}

	for _, dir := range directions {
		neighborX := currentNode.X + dir.dx
		neighborY := currentNode.Y + dir.dy

		// Check bounds
		if !pf.isValidPosition(GridPosition{Grid: Vector2i{X: neighborX, Y: neighborY}}) {
			continue
		}

		// Check if node is in closed set
		packedCoord := pf.packCoordinates(neighborX, neighborY)
		if _, closed := pf.closedSet[packedCoord]; closed {
			continue
		}

		// Check if position is walkable for unit
		if !pf.isWalkable(neighborX, neighborY, request.UnitSize) {
			continue
		}

		// Calculate movement cost
		isDiagonal := dir.dx != 0 && dir.dy != 0
		movementCost := float32(1.0)
		if isDiagonal {
			movementCost = float32(math.Sqrt2) // ~1.414 for diagonal movement
		}

		// Apply terrain cost modifiers
		terrainCost := pf.getTerrainCost(neighborX, neighborY)
		movementCost *= terrainCost

		newGCost := currentNode.GCost + movementCost

		// Get or create neighbor node
		neighborNode := pf.getNode(neighborX, neighborY)

		// Check if this path to neighbor is better
		if neighborNode.HeapIndex == -1 { // Not in open set
			neighborNode.GCost = newGCost
			neighborNode.HCost = pf.heuristic(
				GridPosition{Grid: Vector2i{X: neighborX, Y: neighborY}},
				request.Target,
			)
			neighborNode.FCost = neighborNode.GCost + neighborNode.HCost
			neighborNode.Parent = currentNode
			heap.Push(&pf.openSet, neighborNode)
		} else if newGCost < neighborNode.GCost {
			// Better path found, update node
			neighborNode.GCost = newGCost
			neighborNode.FCost = neighborNode.GCost + neighborNode.HCost
			neighborNode.Parent = currentNode
			heap.Fix(&pf.openSet, neighborNode.HeapIndex)
		}
	}
}

// reconstructPath builds the final path from target back to start
func (pf *Pathfinder) reconstructPath(targetNode *PathNode, request PathRequest) PathResult {
	var gridPath []GridPosition
	var worldPath []Vector3
	totalDistance := float32(0)

	// Build path by following parent pointers
	current := targetNode
	for current != nil {
		gridPath = append(gridPath, GridPosition{Grid: Vector2i{X: current.X, Y: current.Y}})
		worldPath = append(worldPath, pf.gridToWorld(GridPosition{Grid: Vector2i{X: current.X, Y: current.Y}}))

		if current.Parent != nil {
			// Calculate distance between consecutive nodes
			dx := float32(current.X - current.Parent.X)
			dy := float32(current.Y - current.Parent.Y)
			totalDistance += float32(math.Sqrt(float64(dx*dx + dy*dy)))
		}

		current = current.Parent
	}

	// Reverse path (we built it backwards)
	for i := 0; i < len(gridPath)/2; i++ {
		j := len(gridPath) - 1 - i
		gridPath[i], gridPath[j] = gridPath[j], gridPath[i]
		worldPath[i], worldPath[j] = worldPath[j], worldPath[i]
	}

	return PathResult{
		Success:  true,
		Path:     worldPath,
		GridPath: gridPath,
		Distance: totalDistance,
	}
}

// findPartialPath attempts to find a path to the closest reachable position
func (pf *Pathfinder) findPartialPath(request PathRequest) PathResult {
	var bestNode *PathNode
	bestDistance := float32(math.MaxFloat32)

	// Find the closed node closest to target
	for _, node := range pf.closedSet {
		distance := pf.heuristic(GridPosition{Grid: Vector2i{X: node.X, Y: node.Y}}, request.Target)
		if distance < bestDistance {
			bestDistance = distance
			bestNode = node
		}
	}

	if bestNode != nil {
		result := pf.reconstructPath(bestNode, request)
		result.Partial = true
		return result
	}

	return PathResult{Success: false}
}

// heuristic calculates the heuristic distance between two positions (Manhattan + diagonal)
func (pf *Pathfinder) heuristic(a, b GridPosition) float32 {
	dx := float32(absPath(a.Grid.X - b.Grid.X))
	dy := float32(absPath(a.Grid.Y - b.Grid.Y))

	// Octile distance (combination of Manhattan and Euclidean)
	// This is optimal for 8-directional movement
	diagonal := min(dx, dy)
	straight := max(dx, dy) - diagonal
	return diagonal*float32(math.Sqrt2) + straight
}

// isValidPosition checks if a grid position is within world bounds
func (pf *Pathfinder) isValidPosition(pos GridPosition) bool {
	if pf.world == nil || pf.world.TerrainMap == nil {
		return false
	}

	return pos.Grid.X >= 0 && pos.Grid.X < pf.world.TerrainMap.Width &&
		   pos.Grid.Y >= 0 && pos.Grid.Y < pf.world.TerrainMap.Height
}

// isWalkable checks if a position is walkable for a unit of given size
func (pf *Pathfinder) isWalkable(x, y, unitSize int) bool {
	// Check all grid cells that the unit would occupy
	for dx := 0; dx < unitSize; dx++ {
		for dy := 0; dy < unitSize; dy++ {
			checkX := x + dx
			checkY := y + dy

			// Check bounds
			if !pf.isValidPosition(GridPosition{Grid: Vector2i{X: checkX, Y: checkY}}) {
				return false
			}

			// Check terrain walkability
			if !pf.world.IsWalkable(GridPosition{Grid: Vector2i{X: checkX, Y: checkY}}) {
				return false
			}

			// Check for unit/building occupation
			if pf.world.IsOccupied(GridPosition{Grid: Vector2i{X: checkX, Y: checkY}}) {
				return false
			}
		}
	}

	return true
}

// getTerrainCost returns the movement cost for a terrain type
func (pf *Pathfinder) getTerrainCost(x, y int) float32 {
	if pf.world == nil || pf.world.TerrainMap == nil {
		return 1.0
	}

	// Get terrain type at position
	if x < 0 || x >= pf.world.TerrainMap.Width || y < 0 || y >= pf.world.TerrainMap.Height {
		return 10.0 // High cost for out-of-bounds
	}

	// Apply terrain-based movement costs
	terrainType := pf.world.TerrainMap.GetTerrain(x, y)
	switch terrainType {
	case 0: // Grass - normal movement
		return 1.0
	case 1: // Stone - slower movement
		return 1.5
	case 2: // Water - much slower or impassable depending on unit
		return 3.0
	case 3: // Sand - slightly slower
		return 1.2
	default:
		return 1.0
	}
}

// gridToWorld converts grid coordinates to world coordinates
func (pf *Pathfinder) gridToWorld(gridPos GridPosition) Vector3 {
	// Convert grid position to world position
	// Assuming grid spacing of 1.0 world units
	return Vector3{
		X: float64(gridPos.Grid.X) + gridPos.Offset.X,
		Y: 0.0, // Terrain height would be calculated here
		Z: float64(gridPos.Grid.Y) + gridPos.Offset.Y,
	}
}

// getNode gets or creates a node from the pool
func (pf *Pathfinder) getNode(x, y int) *PathNode {
	// Try to reuse a node from the pool
	for _, node := range pf.nodePool {
		if node.X == x && node.Y == y {
			// Reset node state
			node.GCost = 0
			node.HCost = 0
			node.FCost = 0
			node.Parent = nil
			node.HeapIndex = -1
			return node
		}
	}

	// Create new node
	node := &PathNode{
		X: x,
		Y: y,
		HeapIndex: -1,
	}
	pf.nodePool = append(pf.nodePool, node)
	return node
}

// packCoordinates packs x,y coordinates into a single integer for map keys
func (pf *Pathfinder) packCoordinates(x, y int) int {
	return x<<16 | (y & 0xFFFF)
}

// reset clears pathfinder state for a new search
func (pf *Pathfinder) reset() {
	pf.openSet = pf.openSet[:0] // Clear slice but keep capacity

	// Clear closed set
	for k := range pf.closedSet {
		delete(pf.closedSet, k)
	}

	// Reset all nodes in pool
	for _, node := range pf.nodePool {
		node.HeapIndex = -1
		node.GCost = 0
		node.HCost = 0
		node.FCost = 0
		node.Parent = nil
	}
}

// Utility functions
func absPath(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func min(a, b float32) float32 {
	if a < b {
		return a
	}
	return b
}

func max(a, b float32) float32 {
	if a > b {
		return a
	}
	return b
}

// PathfindingManager manages pathfinding for all units
type PathfindingManager struct {
	pathfinder *Pathfinder
	world      *World
}

// NewPathfindingManager creates a new pathfinding manager
func NewPathfindingManager(world *World) *PathfindingManager {
	return &PathfindingManager{
		pathfinder: NewPathfinder(world),
		world:      world,
	}
}

// RequestPath requests a path for a unit
func (pm *PathfindingManager) RequestPath(unit *GameUnit, target Vector3) (*PathResult, error) {
	if unit == nil {
		return nil, fmt.Errorf("unit is nil")
	}

	// Convert world positions to grid positions
	startGrid := pm.world.WorldToGrid(unit.Position)
	targetGrid := pm.world.WorldToGrid(target)

	// Create path request
	request := PathRequest{
		Start:        startGrid,
		Target:       targetGrid,
		UnitSize:     1, // Default unit size, could be read from unit properties
		MaxRange:     0, // No range limit
		AllowPartial: true, // Allow partial paths
	}

	// Find path
	result := pm.pathfinder.FindPath(request)
	return &result, nil
}

// RequestPathWithRange requests a path with a maximum range limit
func (pm *PathfindingManager) RequestPathWithRange(unit *GameUnit, target Vector3, maxRange float32) (*PathResult, error) {
	if unit == nil {
		return nil, fmt.Errorf("unit is nil")
	}

	startGrid := pm.world.WorldToGrid(unit.Position)
	targetGrid := pm.world.WorldToGrid(target)

	request := PathRequest{
		Start:        startGrid,
		Target:       targetGrid,
		UnitSize:     1,
		MaxRange:     maxRange,
		AllowPartial: true,
	}

	result := pm.pathfinder.FindPath(request)
	return &result, nil
}