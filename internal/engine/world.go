package engine

import (
	"fmt"
	"sync"
	"time"

	"teraglest/internal/data"
)

// World represents the complete game world state
type World struct {
	mutex        sync.RWMutex                    // Thread-safe access to world state
	settings     GameSettings                    // Game configuration
	techTree     *data.TechTree                  // Tech tree data
	assetMgr     *data.AssetManager              // Asset management

	// World state
	players      map[int]*Player                 // All players in the game (human + AI)
	ObjectManager *ObjectManager                 // Centralized object management
	resources    map[int]*ResourceNode           // Resource nodes on the map

	// World management
	nextEntityID int                             // Next available entity ID
	gameTime     time.Duration                   // Total game time elapsed
	initialized  bool                            // Whether world has been initialized

	// Spatial organization
	Width        int                             // Map width in tiles
	Height       int                             // Map height in tiles
	tileSize     float32                         // Size of each map tile

	// Grid system for positioning and collision detection
	occupancyGrid [][]bool                      // Track which tiles have units/buildings
	heightMap     [][]float32                   // Basic terrain heights
	walkableGrid  [][]bool                      // Which tiles are passable

	// Game mechanics
	resourceGenerationRate map[string]float32    // Resource generation rates
	unitCap              int                     // Maximum units per player
	buildingCap          int                     // Maximum buildings per player
}

// Player represents a player (human or AI) in the game
type Player struct {
	ID           int                             // Unique player ID
	Name         string                          // Player name
	FactionName  string                          // Faction being played
	IsAI         bool                            // Whether this is an AI player
	IsActive     bool                            // Whether player is still active

	// Player state
	Resources    map[string]int                  // Current resource amounts

	// Player statistics
	UnitsCreated    int                          // Total units created
	UnitsLost       int                          // Total units lost
	BuildingsBuilt  int                          // Total buildings constructed
	ResourcesGathered map[string]int             // Total resources gathered

	// Faction data
	FactionData  *data.FactionDefinition         // Loaded faction definition
}


// ResourceNode represents a resource source on the map
type ResourceNode struct {
	ID           int                             // Unique resource node ID
	ResourceType string                          // Type of resource (gold, wood, etc.)
	Position     Vector3                         // World position
	Amount       int                             // Current resource amount
	MaxAmount    int                             // Maximum resource amount
	IsDepletable bool                            // Whether resource can be depleted
}


// NewWorld creates a new game world instance
func NewWorld(settings GameSettings, techTree *data.TechTree, assetMgr *data.AssetManager) (*World, error) {
	world := &World{
		settings:      settings,
		techTree:      techTree,
		assetMgr:      assetMgr,
		players:       make(map[int]*Player),
		resources:     make(map[int]*ResourceNode),
		nextEntityID:  1,
		Width:         64,  // Default map size
		Height:        64,
		tileSize:      1.0,
		resourceGenerationRate: make(map[string]float32),
		unitCap:       200, // Default unit cap per player
		buildingCap:   50,  // Default building cap per player
	}

	// Initialize default resource generation rates
	world.resourceGenerationRate["gold"] = 1.0
	world.resourceGenerationRate["wood"] = 1.0
	world.resourceGenerationRate["stone"] = 1.0
	world.resourceGenerationRate["energy"] = 2.0

	// Initialize ObjectManager
	world.ObjectManager = NewObjectManager(world)

	// Initialize grid system
	if err := world.initializeGrid(); err != nil {
		return nil, fmt.Errorf("failed to initialize grid system: %w", err)
	}

	return world, nil
}

// Initialize sets up the world state and creates initial players/units
func (w *World) Initialize() error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if w.initialized {
		return fmt.Errorf("world is already initialized")
	}

	// Create human players
	for playerID, factionName := range w.settings.PlayerFactions {
		if err := w.createPlayer(playerID, factionName, false); err != nil {
			return fmt.Errorf("failed to create human player %d: %w", playerID, err)
		}
	}

	// Create AI players
	for playerID, factionName := range w.settings.AIFactions {
		if err := w.createPlayer(playerID, factionName, true); err != nil {
			return fmt.Errorf("failed to create AI player %d: %w", playerID, err)
		}
	}

	// Initialize starting units and resources for each player
	for _, player := range w.players {
		if err := w.initializePlayerStartingState(player); err != nil {
			return fmt.Errorf("failed to initialize starting state for player %d: %w", player.ID, err)
		}
	}

	// Generate resource nodes on the map (simplified for now)
	w.generateResourceNodes()

	w.initialized = true
	return nil
}

// Update advances the world state by the given delta time
func (w *World) Update(deltaTime time.Duration) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if !w.initialized {
		return
	}

	// Update game time
	w.gameTime += deltaTime

	// Update all game objects through the ObjectManager
	w.ObjectManager.Update(deltaTime)

	// Update players (resource generation, etc.)
	for _, player := range w.players {
		w.updatePlayer(player, deltaTime)
	}

	// Process game mechanics (simplified for now)
	w.processGameMechanics(deltaTime)
}

// GetPlayerCount returns the number of active players
func (w *World) GetPlayerCount() int {
	w.mutex.RLock()
	defer w.mutex.RUnlock()

	count := 0
	for _, player := range w.players {
		if player.IsActive {
			count++
		}
	}
	return count
}

// GetTotalUnitCount returns the total number of units in the world
func (w *World) GetTotalUnitCount() int {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	stats := w.ObjectManager.GetStats()
	return stats.TotalUnits
}

// GetPlayer returns a player by ID (thread-safe)
func (w *World) GetPlayer(playerID int) *Player {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	return w.players[playerID]
}

// GetAllPlayers returns a copy of all players (thread-safe)
func (w *World) GetAllPlayers() map[int]*Player {
	w.mutex.RLock()
	defer w.mutex.RUnlock()

	players := make(map[int]*Player)
	for id, player := range w.players {
		// Create a copy to avoid concurrent access issues
		playerCopy := *player
		players[id] = &playerCopy
	}
	return players
}

// GetWorldStats returns current world statistics
func (w *World) GetWorldStats() WorldStats {
	w.mutex.RLock()
	defer w.mutex.RUnlock()

	objectStats := w.ObjectManager.GetStats()
	stats := WorldStats{
		TotalPlayers:     len(w.players),
		ActivePlayers:    w.GetPlayerCount(),
		TotalUnits:       objectStats.TotalUnits,
		TotalBuildings:   objectStats.TotalBuildings,
		TotalResources:   len(w.resources),
		GameTime:         w.gameTime,
		MapSize:          fmt.Sprintf("%dx%d", w.Width, w.Height),
	}

	// Calculate resource totals
	stats.ResourceDistribution = make(map[string]int)
	for _, player := range w.players {
		for resType, amount := range player.Resources {
			stats.ResourceDistribution[resType] += amount
		}
	}

	return stats
}

// WorldStats contains statistics about the world state
type WorldStats struct {
	TotalPlayers         int
	ActivePlayers        int
	TotalUnits           int
	TotalBuildings       int
	TotalResources       int
	GameTime             time.Duration
	MapSize              string
	ResourceDistribution map[string]int
}

// Internal methods

// createPlayer creates a new player with the specified faction
func (w *World) createPlayer(playerID int, factionName string, isAI bool) error {
	// Load faction data
	factions, err := w.assetMgr.LoadFactions()
	if err != nil {
		return fmt.Errorf("failed to load factions: %w", err)
	}

	var factionData *data.FactionDefinition
	for _, faction := range factions {
		if faction.Name == factionName {
			factionData = &faction
			break
		}
	}

	if factionData == nil {
		return fmt.Errorf("faction '%s' not found", factionName)
	}

	// Create player
	player := &Player{
		ID:          playerID,
		Name:        fmt.Sprintf("Player %d", playerID),
		FactionName: factionName,
		IsAI:        isAI,
		IsActive:    true,
		Resources:   make(map[string]int),
		ResourcesGathered: make(map[string]int),
		FactionData: factionData,
	}

	// Initialize starting resources
	for _, startingRes := range factionData.Faction.StartingResources {
		player.Resources[startingRes.Name] = startingRes.Amount
		player.ResourcesGathered[startingRes.Name] = 0
	}

	w.players[playerID] = player
	return nil
}

// initializePlayerStartingState creates starting units for a player
func (w *World) initializePlayerStartingState(player *Player) error {
	// Create starting units
	for _, startingUnit := range player.FactionData.Faction.StartingUnits {
		for i := 0; i < startingUnit.Amount; i++ {
			// Load unit definition
			unitDef, err := w.assetMgr.LoadUnit(player.FactionName, startingUnit.Name)
			if err != nil {
				// Skip this unit if it can't be loaded, but continue with others
				continue
			}

			// Create unit using ObjectManager
			position := Vector3{X: float64(player.ID * 10), Y: 0, Z: float64(i * 2)}
			_, err = w.ObjectManager.CreateUnit(player.ID, startingUnit.Name, position, unitDef)
			if err != nil {
				// Skip this unit if it can't be created, but continue with others
				continue
			}
			player.UnitsCreated++
		}
	}

	return nil
}

// generateResourceNodes creates resource nodes on the map
func (w *World) generateResourceNodes() {
	// Simple resource node generation (placeholder)
	resourceTypes := []string{"gold", "wood", "stone"}

	for i := 0; i < 20; i++ { // Create 20 resource nodes
		nodeID := w.nextEntityID
		w.nextEntityID++

		resourceType := resourceTypes[i%len(resourceTypes)]

		node := &ResourceNode{
			ID:           nodeID,
			ResourceType: resourceType,
			Position:     Vector3{X: float64(i*3 + 5), Y: 0, Z: float64(i*2 + 5)},
			Amount:       1000,  // Starting amount
			MaxAmount:    1000,
			IsDepletable: true,
		}

		w.resources[nodeID] = node
	}
}


// updatePlayer updates player-specific state
func (w *World) updatePlayer(player *Player, deltaTime time.Duration) {
	// Resource generation from buildings
	playerBuildings := w.ObjectManager.GetBuildingsForPlayer(player.ID)
	for _, building := range playerBuildings {
		if building.IsBuilt {
			// Simple resource generation (placeholder)
			// Real implementation would check building type and generate appropriate resources
			for resourceType, rate := range w.resourceGenerationRate {
				if rate > 0 {
					generated := int(rate * float32(deltaTime.Seconds()) * w.settings.ResourceMultiplier)
					if generated > 0 {
						player.Resources[resourceType] += generated
						player.ResourcesGathered[resourceType] += generated
					}
				}
			}
		}
	}
}

// processGameMechanics handles global game mechanics
func (w *World) processGameMechanics(deltaTime time.Duration) {
	// Check win conditions, handle global events, etc.
	// Placeholder for future implementation
}

// Grid System Methods

// initializeGrid initializes the grid system arrays
func (w *World) initializeGrid() error {
	if w.Width <= 0 || w.Height <= 0 {
		return fmt.Errorf("invalid map dimensions: width=%d, height=%d", w.Width, w.Height)
	}

	// Initialize occupancy grid (tracks units/buildings)
	w.occupancyGrid = make([][]bool, w.Height)
	for i := range w.occupancyGrid {
		w.occupancyGrid[i] = make([]bool, w.Width)
	}

	// Initialize height map (terrain heights)
	w.heightMap = make([][]float32, w.Height)
	for i := range w.heightMap {
		w.heightMap[i] = make([]float32, w.Width)
		// Initialize with default ground level (0.0)
		for j := range w.heightMap[i] {
			w.heightMap[i][j] = 0.0
		}
	}

	// Initialize walkable grid (pathfinding)
	w.walkableGrid = make([][]bool, w.Height)
	for i := range w.walkableGrid {
		w.walkableGrid[i] = make([]bool, w.Width)
		// Initialize all tiles as walkable by default
		for j := range w.walkableGrid[i] {
			w.walkableGrid[i][j] = true
		}
	}

	return nil
}

// IsPositionWalkable checks if a grid position is walkable (not occupied and not blocked)
func (w *World) IsPositionWalkable(gridPos Vector2i) bool {
	w.mutex.RLock()
	defer w.mutex.RUnlock()

	// Check bounds
	if !w.isValidGridPosition(gridPos) {
		return false
	}

	// Check if tile is walkable and not occupied
	return w.walkableGrid[gridPos.Y][gridPos.X] && !w.occupancyGrid[gridPos.Y][gridPos.X]
}

// SetOccupied sets the occupancy status of a grid tile
func (w *World) SetOccupied(gridPos Vector2i, occupied bool) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if !w.isValidGridPosition(gridPos) {
		return
	}

	w.occupancyGrid[gridPos.Y][gridPos.X] = occupied
}

// GetHeight returns the terrain height at a grid position
func (w *World) GetHeight(gridPos Vector2i) float32 {
	w.mutex.RLock()
	defer w.mutex.RUnlock()

	if !w.isValidGridPosition(gridPos) {
		return 0.0 // Default ground level
	}

	return w.heightMap[gridPos.Y][gridPos.X]
}

// SetHeight sets the terrain height at a grid position
func (w *World) SetHeight(gridPos Vector2i, height float32) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if !w.isValidGridPosition(gridPos) {
		return
	}

	w.heightMap[gridPos.Y][gridPos.X] = height
}

// SetWalkable sets whether a grid position is walkable
func (w *World) SetWalkable(gridPos Vector2i, walkable bool) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if !w.isValidGridPosition(gridPos) {
		return
	}

	w.walkableGrid[gridPos.Y][gridPos.X] = walkable
}

// GetUnitsInTile returns all units at a specific grid position
func (w *World) GetUnitsInTile(gridPos Vector2i) []*GameUnit {
	w.mutex.RLock()
	defer w.mutex.RUnlock()

	if !w.isValidGridPosition(gridPos) {
		return nil
	}

	return w.ObjectManager.GetUnitsAtPosition(gridPos)
}

// GetNearestWalkablePosition finds the nearest walkable position to a target
func (w *World) GetNearestWalkablePosition(targetPos Vector2i) Vector2i {
	w.mutex.RLock()
	defer w.mutex.RUnlock()

	// If the target position is already walkable, return it
	if w.isValidGridPosition(targetPos) && w.walkableGrid[targetPos.Y][targetPos.X] && !w.occupancyGrid[targetPos.Y][targetPos.X] {
		return targetPos
	}

	// Search in expanding rings around the target position
	for radius := 1; radius <= 10; radius++ {
		for dx := -radius; dx <= radius; dx++ {
			for dy := -radius; dy <= radius; dy++ {
				// Skip positions that aren't on the edge of the current radius
				if dx*dx+dy*dy > radius*radius || dx*dx+dy*dy < (radius-1)*(radius-1) {
					continue
				}

				testPos := Vector2i{X: targetPos.X + dx, Y: targetPos.Y + dy}
				if w.isValidGridPosition(testPos) && w.walkableGrid[testPos.Y][testPos.X] && !w.occupancyGrid[testPos.Y][testPos.X] {
					return testPos
				}
			}
		}
	}

	// Fallback to original position if no walkable position found
	return targetPos
}

// isValidGridPosition checks if a grid position is within world bounds (internal helper)
func (w *World) isValidGridPosition(pos Vector2i) bool {
	return pos.X >= 0 && pos.Y >= 0 && pos.X < w.Width && pos.Y < w.Height
}

