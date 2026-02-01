package engine

import (
	"fmt"
	"math"
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
	commandProcessor *CommandProcessor           // Command system integration
	pathfindingMgr *PathfindingManager           // A* pathfinding system
	behaviorTreeMgr *BehaviorTreeManager         // Unit AI behavior tree system
	strategicAIMgr *StrategicAIManager           // Strategic AI management system
	groupMgr     *GroupManager                   // Unit formation and group management
	productionSys *ProductionSystem              // Building and unit production system
	resources    map[int]*ResourceNode           // Resource nodes on the map

	// World management
	nextEntityID int                             // Next available entity ID
	gameTime     time.Duration                   // Total game time elapsed
	initialized  bool                            // Whether world has been initialized

	// Spatial organization
	Width        int                             // Map width in tiles
	Height       int                             // Map height in tiles
	tileSize     float32                         // Size of each map tile
	Map          *Map                            // Loaded map data (if created from map)
	TerrainMap   *TerrainMap                     // Terrain data for pathfinding

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
	ResourcesSpent    map[string]int             // Total resources spent

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

// TerrainMap represents terrain data for pathfinding and rendering
type TerrainMap struct {
	Width       int        // Map width in tiles
	Height      int        // Map height in tiles
	TerrainData [][]int    // 2D array of terrain type IDs
}

// GetTerrain returns the terrain type at the specified coordinates
func (tm *TerrainMap) GetTerrain(x, y int) int {
	if x < 0 || x >= tm.Width || y < 0 || y >= tm.Height {
		return -1 // Invalid terrain
	}
	return tm.TerrainData[y][x]
}

// NewTerrainMap creates a new terrain map with default grass terrain
func NewTerrainMap(width, height int) *TerrainMap {
	// Initialize with default grass terrain (type 0)
	terrainData := make([][]int, height)
	for y := 0; y < height; y++ {
		terrainData[y] = make([]int, width)
		for x := 0; x < width; x++ {
			terrainData[y][x] = 0 // Grass terrain
		}
	}

	return &TerrainMap{
		Width:       width,
		Height:      height,
		TerrainData: terrainData,
	}
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

	// Initialize CommandProcessor
	world.commandProcessor = NewCommandProcessor(world)

	// Initialize PathfindingManager
	world.pathfindingMgr = NewPathfindingManager(world)

	// Initialize BehaviorTreeManager
	world.behaviorTreeMgr = NewBehaviorTreeManager(world)

	// Initialize StrategicAIManager
	world.strategicAIMgr = NewStrategicAIManager(world)

	// Initialize GroupManager
	world.groupMgr = NewGroupManager(world)

	// Initialize ProductionSystem
	world.productionSys = NewProductionSystem(world)

	// Initialize grid system
	if err := world.initializeGrid(); err != nil {
		return nil, fmt.Errorf("failed to initialize grid system: %w", err)
	}

	return world, nil
}

// NewWorldFromMap creates a new game world instance from a map file
func NewWorldFromMap(settings GameSettings, techTree *data.TechTree, assetMgr *data.AssetManager, mapName string) (*World, error) {
	// Create MapManager for loading map data
	dataRoot := "/home/solifugus/development/teraglest/megaglest-source/data/glest_game" // TODO: make configurable
	mapManager := NewMapManager(assetMgr, dataRoot)

	// Load map data
	mapData, err := mapManager.LoadMap(mapName)
	if err != nil {
		return nil, fmt.Errorf("failed to load map %s: %w", mapName, err)
	}

	// Validate map
	issues := mapManager.ValidateMap(mapData)
	if len(issues) > 0 {
		return nil, fmt.Errorf("map validation failed: %v", issues)
	}

	// Create world with map dimensions
	world := &World{
		settings:      settings,
		techTree:      techTree,
		assetMgr:      assetMgr,
		players:       make(map[int]*Player),
		resources:     make(map[int]*ResourceNode),
		nextEntityID:  1,
		Width:         mapData.Width,     // From map file
		Height:        mapData.Height,    // From map file
		tileSize:      1.0,               // Standard tile size
		Map:           mapData,           // Store map reference
		resourceGenerationRate: make(map[string]float32),
		unitCap:       200,               // Default unit cap per player
		buildingCap:   50,                // Default building cap per player
	}

	// Initialize default resource generation rates
	world.resourceGenerationRate["gold"] = 1.0
	world.resourceGenerationRate["wood"] = 1.0
	world.resourceGenerationRate["stone"] = 1.0
	world.resourceGenerationRate["energy"] = 2.0

	// Initialize ObjectManager
	world.ObjectManager = NewObjectManager(world)

	// Initialize CommandProcessor
	world.commandProcessor = NewCommandProcessor(world)

	// Initialize PathfindingManager
	world.pathfindingMgr = NewPathfindingManager(world)

	// Initialize BehaviorTreeManager
	world.behaviorTreeMgr = NewBehaviorTreeManager(world)

	// Initialize StrategicAIManager
	world.strategicAIMgr = NewStrategicAIManager(world)

	// Initialize GroupManager
	world.groupMgr = NewGroupManager(world)

	// Initialize ProductionSystem
	world.productionSys = NewProductionSystem(world)

	// Initialize grid system from map data
	if err := world.initializeFromMap(mapData); err != nil {
		return nil, fmt.Errorf("failed to initialize world from map: %w", err)
	}

	return world, nil
}

// Initialize sets up the world state and creates initial players/units
func (w *World) Initialize() error {
	// Check if already initialized (with lock)
	w.mutex.Lock()
	if w.initialized {
		w.mutex.Unlock()
		return fmt.Errorf("world is already initialized")
	}
	w.mutex.Unlock()

	// Create human players (no lock needed for this)
	for playerID, factionName := range w.settings.PlayerFactions {
		if err := w.createPlayer(playerID, factionName, false); err != nil {
			return fmt.Errorf("failed to create human player %d: %w", playerID, err)
		}
	}

	// Create AI players (no lock needed for this)
	for playerID, factionName := range w.settings.AIFactions {
		if err := w.createPlayer(playerID, factionName, true); err != nil {
			return fmt.Errorf("failed to create AI player %d: %w", playerID, err)
		}
	}

	// Initialize starting units and resources for each player (no world lock needed)
	for _, player := range w.players {
		if err := w.initializePlayerStartingState(player); err != nil {
			return fmt.Errorf("failed to initialize starting state for player %d: %w", player.ID, err)
		}
	}

	// Generate resource nodes on the map (simplified for now)
	w.generateResourceNodes()

	// Set initialized flag (with lock)
	w.mutex.Lock()
	w.initialized = true
	w.mutex.Unlock()

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

	// Process commands after object updates (pass players to avoid nested locking)
	w.commandProcessor.UpdateWithPlayers(deltaTime, w.players)

	// Update production system (building construction and unit production)
	if w.productionSys != nil {
		w.productionSys.Update(deltaTime)
	}

	// Update behavior trees for unit AI
	w.behaviorTreeMgr.Update(deltaTime)

	// Update strategic AI for AI players
	if w.strategicAIMgr != nil {
		w.strategicAIMgr.Update(deltaTime)
	}

	// Update unit formations and groups
	if w.groupMgr != nil {
		w.groupMgr.Update(deltaTime)
	}

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

// GetGameTime returns the total elapsed game time
func (w *World) GetGameTime() time.Duration {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	return w.gameTime
}

// AddPlayer adds a new player to the world
func (w *World) AddPlayer(playerID int, name string, factionName string, isAI bool) error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if _, exists := w.players[playerID]; exists {
		return fmt.Errorf("player with ID %d already exists", playerID)
	}

	player := &Player{
		ID:           playerID,
		Name:         name,
		FactionName:  factionName,
		IsAI:         isAI,
		IsActive:     true,
		Resources:    make(map[string]int),
		ResourcesGathered: make(map[string]int),
		ResourcesSpent:    make(map[string]int),
	}

	// Initialize starting resources
	player.Resources["gold"] = 1000
	player.Resources["wood"] = 1000
	player.Resources["stone"] = 500
	player.Resources["energy"] = 300

	w.players[playerID] = player
	return nil
}

// InitializeAIPlayer creates AI behavior for a player
func (w *World) InitializeAIPlayer(playerID int, personality string, difficulty string) error {
	if w.strategicAIMgr == nil {
		return fmt.Errorf("strategic AI manager not initialized")
	}

	player := w.GetPlayer(playerID)
	if player == nil {
		return fmt.Errorf("player with ID %d not found", playerID)
	}

	if !player.IsAI {
		return fmt.Errorf("player %d is not an AI player", playerID)
	}

	// Convert string parameters to types
	var aiPersonality AIPersonality
	switch personality {
	case "conservative":
		aiPersonality = ConservativePersonality
	case "aggressive":
		aiPersonality = AggressivePersonality
	case "balanced":
		aiPersonality = BalancedPersonality
	case "technological":
		aiPersonality = TechnologicalPersonality
	case "expansionist":
		aiPersonality = ExpansionistPersonality
	default:
		aiPersonality = BalancedPersonality
	}

	var aiDifficulty AIDifficulty
	switch difficulty {
	case "easy":
		aiDifficulty = DifficultyEasy
	case "normal":
		aiDifficulty = DifficultyNormal
	case "hard":
		aiDifficulty = DifficultyHard
	case "expert":
		aiDifficulty = DifficultyExpert
	default:
		aiDifficulty = DifficultyNormal
	}

	return w.strategicAIMgr.InitializeAIPlayer(playerID, aiPersonality, aiDifficulty)
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
		ResourcesSpent: make(map[string]int),
		FactionData: factionData,
	}

	// Initialize starting resources
	for _, startingRes := range factionData.Faction.StartingResources {
		player.Resources[startingRes.Name] = startingRes.Amount
		player.ResourcesGathered[startingRes.Name] = 0
		player.ResourcesSpent[startingRes.Name] = 0
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


// updatePlayer updates player-specific state with enhanced resource management
func (w *World) updatePlayer(player *Player, deltaTime time.Duration) {
	// Enhanced resource generation from buildings
	playerBuildings := w.ObjectManager.GetBuildingsForPlayer(player.ID)
	generatedResources := make(map[string]int)

	for _, building := range playerBuildings {
		if building.IsBuilt && building.Health > 0 {
			// Process building-specific resource generation
			for resourceType, rate := range building.ResourceGeneration {
				generated := w.calculateResourceGeneration(rate, deltaTime, building)
				if generated > 0 {
					generatedResources[resourceType] += generated
				}
			}

			// Update last generation time
			building.LastResourceGen = time.Now()
		}
	}

	// Apply generated resources using the new AddResources method
	if len(generatedResources) > 0 {
		// Don't use AddResources here to avoid mutex lock since we're already in updatePlayer
		for resourceType, amount := range generatedResources {
			player.Resources[resourceType] += amount
			player.ResourcesGathered[resourceType] += amount
		}

		// Log generation event
		w.logResourceTransaction(player.ID, generatedResources, "building_generation", "addition")
	}

	// Process resource dropoffs from gathering units
	w.processResourceDropoffs(player)
}

// processGameMechanics handles global game mechanics
func (w *World) processGameMechanics(deltaTime time.Duration) {
	// Check win conditions, handle global events, etc.
	// Placeholder for future implementation
}

// Resource Management Methods

// ResourceStatus represents the current resource state for a player
type ResourceStatus struct {
	Resources        map[string]int     // Current resource amounts
	ResourceRates    map[string]float32 // Current generation rates per second
	ResourcesGathered map[string]int    // Total resources gathered (statistics)
	ResourcesSpent    map[string]int    // Total resources spent (statistics)
}

// DeductResources safely deducts resources from a player with validation
func (w *World) DeductResources(playerID int, cost map[string]int, purpose string) error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	player := w.players[playerID]
	if player == nil {
		return fmt.Errorf("player %d not found", playerID)
	}

	// Validate before deduction using ResourceValidator
	validator := NewResourceValidator(w)
	result := validator.ValidateResources(ResourceCheck{
		PlayerID: playerID,
		Required: cost,
		Purpose:  purpose,
	})

	if !result.Valid {
		return fmt.Errorf("resource deduction failed: %s", result.Error)
	}

	// Perform deduction
	for resourceType, amount := range cost {
		if amount > 0 { // Only deduct positive amounts
			player.Resources[resourceType] -= amount

			// Track spending statistics
			if player.ResourcesSpent == nil {
				player.ResourcesSpent = make(map[string]int)
			}
			player.ResourcesSpent[resourceType] += amount
		}
	}

	// Log resource deduction event
	w.logResourceTransaction(playerID, cost, purpose, "deduction")

	return nil
}

// AddResources adds resources to a player's pool
func (w *World) AddResources(playerID int, resources map[string]int, source string) error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	player := w.players[playerID]
	if player == nil {
		return fmt.Errorf("player %d not found", playerID)
	}

	// Add resources
	for resourceType, amount := range resources {
		if amount > 0 { // Only add positive amounts
			player.Resources[resourceType] += amount
			player.ResourcesGathered[resourceType] += amount
		}
	}

	// Log resource addition event
	w.logResourceTransaction(playerID, resources, source, "addition")

	return nil
}

// GetResourceStatus returns current resource status for a player
func (w *World) GetResourceStatus(playerID int) ResourceStatus {
	w.mutex.RLock()
	defer w.mutex.RUnlock()

	player := w.players[playerID]
	if player == nil {
		return ResourceStatus{}
	}

	// Calculate current resource generation rates
	resourceRates := w.calculateResourceRates(playerID)

	status := ResourceStatus{
		Resources:        make(map[string]int),
		ResourceRates:    resourceRates,
		ResourcesGathered: make(map[string]int),
		ResourcesSpent:    make(map[string]int),
	}

	// Copy current resources
	for k, v := range player.Resources {
		status.Resources[k] = v
	}

	// Copy gathered statistics
	for k, v := range player.ResourcesGathered {
		status.ResourcesGathered[k] = v
	}

	// Copy spent statistics
	if player.ResourcesSpent != nil {
		for k, v := range player.ResourcesSpent {
			status.ResourcesSpent[k] = v
		}
	}

	return status
}

// calculateResourceRates calculates current resource generation rates per second
func (w *World) calculateResourceRates(playerID int) map[string]float32 {
	rates := make(map[string]float32)

	// Get rates from buildings
	playerBuildings := w.ObjectManager.GetBuildingsForPlayer(playerID)
	for _, building := range playerBuildings {
		if building.IsBuilt && building.Health > 0 {
			for resType, rate := range building.ResourceGeneration {
				// Apply upgrade multipliers and game settings
				upgradeMultiplier := 1.0 + (float32(building.UpgradeLevel-1) * 0.2) // 20% per upgrade
				gameMultiplier := w.settings.ResourceMultiplier
				effectiveRate := rate * upgradeMultiplier * gameMultiplier
				rates[resType] += effectiveRate
			}
		}
	}

	// Add gathering rates from active gathering units
	units := w.ObjectManager.UnitManager.GetUnitsForPlayer(playerID)
	for _, unit := range units {
		if unit.State == UnitStateGathering && unit.GatherTarget != nil {
			resType := unit.GatherTarget.ResourceType
			if gatherRate, ok := unit.GatherRate[resType]; ok {
				rates[resType] += gatherRate
			}
		}
	}

	return rates
}

// calculateResourceGeneration calculates resource generation for a building
func (w *World) calculateResourceGeneration(baseRate float32, deltaTime time.Duration, building *GameBuilding) int {
	// Factor in building upgrade level and game settings
	upgradeMultiplier := 1.0 + (float32(building.UpgradeLevel-1) * 0.2) // 20% per upgrade
	gameMultiplier := w.settings.ResourceMultiplier

	effectiveRate := baseRate * upgradeMultiplier * gameMultiplier
	generated := int(effectiveRate * float32(deltaTime.Seconds()))

	return generated
}

// processResourceDropoffs handles units returning resources to collection points
func (w *World) processResourceDropoffs(player *Player) {
	// Get all units for this player
	units := w.ObjectManager.UnitManager.GetUnitsForPlayer(player.ID)

	for _, unit := range units {
		// Check if unit has carried resources and is at dropoff point
		if len(unit.CarriedResources) > 0 && w.isAtDropoffPoint(unit) {
			// Add carried resources to player pool
			for resourceType, amount := range unit.CarriedResources {
				if amount > 0 {
					player.Resources[resourceType] += amount
					player.ResourcesGathered[resourceType] += amount
				}
			}

			// Clear carried resources
			unit.CarriedResources = make(map[string]int)

			// Log dropoff event
			w.logResourceTransaction(player.ID, unit.CarriedResources, "resource_dropoff", "addition")
		}
	}
}

// isAtDropoffPoint checks if a unit is close enough to a dropoff point
func (w *World) isAtDropoffPoint(unit *GameUnit) bool {
	// Simple implementation: check if unit is near any building that can accept resources
	playerBuildings := w.ObjectManager.GetBuildingsForPlayer(unit.PlayerID)

	for _, building := range playerBuildings {
		if building.IsBuilt && building.Health > 0 {
			// Check distance to building (simplified: within 2 units)
			distance := w.CalculateDistance(unit.Position, building.Position)
			if distance <= 2.0 {
				return true
			}
		}
	}

	return false
}

// logResourceTransaction logs resource transactions for events and statistics
func (w *World) logResourceTransaction(playerID int, resources map[string]int, source, transactionType string) {
	// Create resource transaction events for each resource type
	for resourceType, amount := range resources {
		if amount > 0 {
			// Create ResourceEvent data
			resourceEvent := ResourceEvent{
				PlayerID:        playerID,
				ResourceType:    resourceType,
				Amount:          amount,
				Source:          source,
				Timestamp:       time.Now(),
				TransactionType: transactionType,
			}

			// Determine event type based on transaction type
			var eventType GameEventType
			var message string

			if transactionType == "addition" {
				eventType = EventTypeResourceGained
				message = fmt.Sprintf("Player %d gained %d %s from %s", playerID, amount, resourceType, source)
			} else if transactionType == "deduction" {
				eventType = EventTypeResourceSpent
				message = fmt.Sprintf("Player %d spent %d %s for %s", playerID, amount, resourceType, source)
			} else {
				continue // Skip unknown transaction types
			}

			// Create game event
			gameEvent := GameEvent{
				Type:      eventType,
				Timestamp: time.Now(),
				PlayerID:  playerID,
				Data:      resourceEvent,
				Message:   message,
			}

			// Send event to game event system if available
			w.sendResourceEvent(gameEvent)
		}
	}
}

// sendResourceEvent sends a resource event to the game's event system
func (w *World) sendResourceEvent(event GameEvent) {
	// TODO: This would be connected to the actual game instance's event system
	// For now, we'll just log it (could be expanded to send to game.eventQueue)

	// Simple logging for development - in production this would send to game.eventQueue
	// fmt.Printf("[RESOURCE EVENT] %s\n", event.Message)
}

// CalculateDistance calculates the Euclidean distance between two 3D points
func (w *World) CalculateDistance(pos1, pos2 Vector3) float64 {
	dx := pos1.X - pos2.X
	dy := pos1.Y - pos2.Y
	dz := pos1.Z - pos2.Z
	return math.Sqrt(dx*dx + dy*dy + dz*dz)
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

	// Initialize terrain map for pathfinding
	w.TerrainMap = NewTerrainMap(w.Width, w.Height)

	return nil
}

// initializeFromMap initializes the world grid system from loaded map data
func (w *World) initializeFromMap(mapData *Map) error {
	// First initialize the basic grid arrays
	if err := w.initializeGrid(); err != nil {
		return err
	}

	// Populate terrain data from map
	for y := 0; y < mapData.Height; y++ {
		for x := 0; x < mapData.Width; x++ {
			// Copy height data from map
			w.heightMap[y][x] = mapData.HeightMap[y][x]

			// Calculate walkability based on terrain objects and surfaces
			w.walkableGrid[y][x] = w.calculateWalkability(mapData, x, y)
		}
	}

	// Initialize player starting positions
	if err := w.initializeStartPositions(mapData.StartPositions); err != nil {
		return fmt.Errorf("failed to initialize start positions: %w", err)
	}

	// Place resource nodes from map data
	if err := w.placeResourceNodesFromMap(mapData); err != nil {
		return fmt.Errorf("failed to place resource nodes: %w", err)
	}

	return nil
}

// calculateWalkability determines if a tile is walkable based on terrain data
func (w *World) calculateWalkability(mapData *Map, x, y int) bool {
	// Check terrain object walkability
	objectIndex := mapData.ObjectMap[y][x]
	if objectIndex > 0 && mapData.Tileset != nil {
		if obj := mapData.Tileset.GetObject(int(objectIndex)); obj != nil {
			return obj.Walkable
		}
	}

	// Check surface walkability (water, cliffs, etc.)
	surfaceIndex := mapData.SurfaceMap[y][x]
	height := mapData.HeightMap[y][x]

	return w.isSurfaceWalkable(int(surfaceIndex), height, mapData)
}

// isSurfaceWalkable determines if a surface type at a given height is walkable
func (w *World) isSurfaceWalkable(surfaceIndex int, height float32, mapData *Map) bool {
	// Check water level
	if height <= mapData.WaterLevel {
		return false // Water tiles are not walkable by default
	}

	// Check for steep cliffs (large height differences)
	if mapData.Version == MapVersionMGM && height >= mapData.CliffLevel {
		return false // Cliff areas are not walkable
	}

	// All other surfaces are walkable by default
	// This could be enhanced with surface-specific walkability rules
	return true
}

// initializeStartPositions sets up player starting positions
func (w *World) initializeStartPositions(startPositions []Vector2i) error {
	// Store start positions for player initialization
	// This could be enhanced to validate positions and create starting units

	for i, pos := range startPositions {
		// Ensure the start position is valid and walkable
		if !w.Map.IsValidPosition(pos.X, pos.Y) {
			return fmt.Errorf("start position %d is out of bounds: (%d, %d)", i+1, pos.X, pos.Y)
		}

		// Mark the area around start positions as occupied temporarily
		// This prevents resource nodes from spawning too close to start positions
		for dy := -2; dy <= 2; dy++ {
			for dx := -2; dx <= 2; dx++ {
				nx, ny := pos.X+dx, pos.Y+dy
				if w.Map.IsValidPosition(nx, ny) {
					// Temporary marking - will be cleared when units are spawned
					w.occupancyGrid[ny][nx] = true
				}
			}
		}
	}

	return nil
}

// placeResourceNodesFromMap creates resource nodes based on map data
func (w *World) placeResourceNodesFromMap(mapData *Map) error {
	resourceNodeCount := 0

	// Scan the map for areas suitable for resource placement
	// This is a simplified implementation - real maps might have explicit resource placement data
	for y := 0; y < mapData.Height; y += 8 { // Sample every 8 tiles
		for x := 0; x < mapData.Width; x += 8 {
			// Check if this area is suitable for resources
			if w.isResourceAreaSuitable(x, y, mapData) {
				// Place a resource node
				resourceType := w.determineResourceType(x, y, mapData)
				if resourceType != "" {
					err := w.placeResourceNode(x, y, resourceType)
					if err != nil {
						continue // Skip this position if placement fails
					}
					resourceNodeCount++
				}
			}
		}
	}

	fmt.Printf("Placed %d resource nodes from map data\n", resourceNodeCount)
	return nil
}

// isResourceAreaSuitable checks if an area is suitable for resource placement
func (w *World) isResourceAreaSuitable(x, y int, mapData *Map) bool {
	// Check if the area is walkable
	if !w.walkableGrid[y][x] {
		return false
	}

	// Check if area is not too close to start positions (already marked as occupied)
	if w.occupancyGrid[y][x] {
		return false
	}

	// Check terrain height (avoid placing resources in water or on cliffs)
	height := mapData.HeightMap[y][x]
	if height <= mapData.WaterLevel || height >= mapData.WaterLevel+10 {
		return false
	}

	return true
}

// determineResourceType determines what type of resource to place based on terrain
func (w *World) determineResourceType(x, y int, mapData *Map) string {
	// Simple heuristic based on position and terrain
	// This could be enhanced with more sophisticated logic

	height := mapData.HeightMap[y][x]
	surfaceIndex := mapData.SurfaceMap[y][x]

	// Higher elevations -> stone
	if height > mapData.WaterLevel+6 {
		return "stone"
	}

	// Forest areas (object index suggests trees) -> wood
	objectIndex := mapData.ObjectMap[y][x]
	if objectIndex > 0 && mapData.Tileset != nil {
		if obj := mapData.Tileset.GetObject(int(objectIndex)); obj != nil && !obj.Walkable {
			return "wood" // Trees are not walkable, so areas with trees get wood
		}
	}

	// Grass surfaces -> gold
	if MapSurfaceType(surfaceIndex) == SurfaceGrass {
		return "gold"
	}

	return "" // No resource for this area
}

// placeResourceNode creates a resource node at the specified position
func (w *World) placeResourceNode(x, y int, resourceType string) error {
	position := Vector3{
		X: float64(x) * float64(w.tileSize),
		Y: float64(w.heightMap[y][x]),
		Z: float64(y) * float64(w.tileSize),
	}

	// Determine resource amount based on type and map size
	var amount int
	switch resourceType {
	case "gold":
		amount = 1000 + (w.Width*w.Height)/100 // Scale with map size
	case "wood":
		amount = 800 + (w.Width*w.Height)/120
	case "stone":
		amount = 1200 + (w.Width*w.Height)/80
	default:
		amount = 500
	}

	// Create the resource node
	resourceNode := &ResourceNode{
		ID:           w.nextEntityID,
		ResourceType: resourceType,
		Position:     position,
		Amount:       amount,
		MaxAmount:    amount,
		IsDepletable: true,
	}

	w.nextEntityID++
	w.resources[resourceNode.ID] = resourceNode

	// Mark the tile as occupied
	w.occupancyGrid[y][x] = true

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

	// DEBUG: Check if occupancyGrid is nil
	if w.occupancyGrid == nil {
		panic(fmt.Sprintf("CRITICAL: occupancyGrid is nil! World: %dx%d, GridPos: (%d,%d)",
			w.Width, w.Height, gridPos.X, gridPos.Y))
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

// GetTileSize returns the size of each map tile
func (w *World) GetTileSize() float32 {
	return w.tileSize
}

// WorldToGrid converts world coordinates to grid coordinates using this world's tile size
func (w *World) WorldToGrid(worldPos Vector3) GridPosition {
	return WorldToGrid(worldPos, w.tileSize)
}

// GridToWorld converts grid coordinates to world coordinates using this world's tile size
func (w *World) GridToWorld(gridPos GridPosition) Vector3 {
	return GridToWorld(gridPos, w.tileSize)
}

// IsWalkable checks if a grid position is walkable (pathfinding compatible method)
func (w *World) IsWalkable(gridPos GridPosition) bool {
	return w.IsPositionWalkable(gridPos.Grid)
}

// IsOccupied checks if a grid position is occupied (pathfinding compatible method)
func (w *World) IsOccupied(gridPos GridPosition) bool {
	return w.ObjectManager.UnitManager.IsPositionOccupied(gridPos.Grid)
}

// SetOccupiedGrid sets the occupancy state for a grid position (pathfinding compatible method)
func (w *World) SetOccupiedGrid(gridPos GridPosition, occupied bool) {
	w.SetOccupied(gridPos.Grid, occupied)
}

// GetAllResourceNodes returns all resource nodes in the world
func (w *World) GetAllResourceNodes() []*ResourceNode {
	w.mutex.RLock()
	defer w.mutex.RUnlock()

	nodes := make([]*ResourceNode, 0, len(w.resources))
	for _, node := range w.resources {
		nodes = append(nodes, node)
	}
	return nodes
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

// Accessor methods for UI integration

// GetCommandProcessor returns the command processor for issuing commands
func (w *World) GetCommandProcessor() interface{} {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	return w.commandProcessor
}

// GetProductionSystem returns the production system for managing production
func (w *World) GetProductionSystem() *ProductionSystem {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	return w.productionSys
}

// GetPlayers returns all players in the world
func (w *World) GetPlayers() map[int]*Player {
	w.mutex.RLock()
	defer w.mutex.RUnlock()

	// Return copy to avoid race conditions
	players := make(map[int]*Player)
	for id, player := range w.players {
		players[id] = player
	}
	return players
}

// GetResources returns all resource nodes in the world
func (w *World) GetResources() map[int]*ResourceNode {
	w.mutex.RLock()
	defer w.mutex.RUnlock()

	// Return copy to avoid race conditions
	resources := make(map[int]*ResourceNode)
	for id, resource := range w.resources {
		resources[id] = resource
	}
	return resources
}


// GetNextEntityID returns the next available entity ID
func (w *World) GetNextEntityID() int {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	id := w.nextEntityID
	w.nextEntityID++
	return id
}

// GetResourcesMutable returns a mutable reference to resources (for test setup)
func (w *World) GetResourcesMutable() map[int]*ResourceNode {
	return w.resources
}

