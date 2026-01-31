package engine

import (
	"context"
	"fmt"
	"sync"
	"time"

	"teraglest/internal/data"
)

// GameState represents the current state of the game
type GameState int

const (
	GameStateLoading GameState = iota // Game is loading assets and initializing
	GameStatePlaying                  // Game is actively being played
	GameStatePaused                   // Game is paused
	GameStateEnded                    // Game has ended (victory/defeat/quit)
)

// GameSettings contains configurable game parameters
type GameSettings struct {
	TechTreePath     string            // Path to tech tree data
	MapPath          string            // Path to map file (optional for now)
	PlayerFactions   map[int]string    // Player ID to faction name mapping
	AIFactions       map[int]string    // AI player ID to faction name mapping
	GameSpeed        float32           // Game speed multiplier (1.0 = normal)
	ResourceMultiplier float32         // Resource generation multiplier
	MaxPlayers       int               // Maximum number of players
	GameTimeLimit    time.Duration     // Game time limit (0 = no limit)
	EnableFogOfWar   bool              // Whether fog of war is enabled
	AllowCheats      bool              // Whether cheat codes are allowed
}

// GameStats tracks game performance and statistics
type GameStats struct {
	StartTime        time.Time         // When the game started
	CurrentGameTime  time.Duration     // Current in-game time
	FrameCount       uint64            // Total frames processed
	AverageFrameTime time.Duration     // Average time per frame
	PlayersActive    int               // Number of active players
	UnitsTotal       int               // Total number of units in game
	ResourcesTotal   map[string]int64  // Total resources across all players
	LastUpdateTime   time.Time         // When stats were last updated
}

// Game represents the main game controller and state manager
type Game struct {
	mutex       sync.RWMutex          // Thread-safe access to game state
	state       GameState             // Current game state
	settings    GameSettings          // Game configuration
	stats       GameStats             // Game performance statistics
	world       *World                // Game world state
	assetMgr    *data.AssetManager    // Asset management system
	techTree    *data.TechTree        // Loaded tech tree data

	// Lifecycle management
	ctx         context.Context       // Game context for cancellation
	cancel      context.CancelFunc    // Function to cancel game operations
	updateTicker *time.Ticker         // Game update timer
	isRunning   bool                  // Whether game loop is running

	// Game loop timing
	targetFPS   int                   // Target frames per second
	frameTime   time.Duration         // Target time per frame
	lastUpdate  time.Time             // Last update timestamp

	// Event system (basic for now)
	eventQueue  chan GameEvent        // Game event queue
	maxEvents   int                   // Maximum events in queue
}

// GameEvent represents an event that occurs during gameplay
type GameEvent struct {
	Type        GameEventType         // Type of event
	Timestamp   time.Time             // When the event occurred
	PlayerID    int                   // Player associated with event (-1 for system)
	Data        interface{}           // Event-specific data
	Message     string                // Human-readable message
}

// ResourceEvent represents a resource transaction event
type ResourceEvent struct {
	PlayerID     int                   // Player involved in transaction
	ResourceType string                // Type of resource (gold, wood, etc.)
	Amount       int                   // Amount of resource
	Source       string                // Source of transaction (building_generation, unit_cost, etc.)
	Timestamp    time.Time             // When the transaction occurred
	TransactionType string             // "addition" or "deduction"
}

// GameEventType represents different types of game events
type GameEventType int

const (
	EventTypeGameStart    GameEventType = iota // Game started
	EventTypeGamePause                         // Game paused
	EventTypeGameResume                        // Game resumed
	EventTypeGameEnd                           // Game ended
	EventTypeUnitCreated                       // Unit was created
	EventTypeUnitDestroyed                     // Unit was destroyed
	EventTypeResourceGained                    // Player gained resources
	EventTypeResourceSpent                     // Player spent resources
	EventTypeResourceDepleted                  // Resource node depleted
	EventTypePopulationLimit                   // Population limit reached
	EventTypeTechResearched                    // Technology was researched
	EventTypeBuildingCompleted                 // Building construction completed
	EventTypePlayerDefeated                    // Player was defeated
	EventTypePlayerVictory                     // Player achieved victory
)

// NewGame creates a new game instance with the specified settings
func NewGame(settings GameSettings, assetMgr *data.AssetManager) (*Game, error) {
	// Validate settings
	if err := validateGameSettings(settings); err != nil {
		return nil, fmt.Errorf("invalid game settings: %w", err)
	}

	// Create game context
	ctx, cancel := context.WithCancel(context.Background())

	// Initialize game instance
	game := &Game{
		state:       GameStateLoading,
		settings:    settings,
		assetMgr:    assetMgr,
		ctx:         ctx,
		cancel:      cancel,
		targetFPS:   60,
		frameTime:   time.Second / 60,
		eventQueue:  make(chan GameEvent, 1000), // Buffer for 1000 events
		maxEvents:   1000,
		lastUpdate:  time.Now(),
	}

	// Initialize game statistics
	game.stats = GameStats{
		StartTime:       time.Now(),
		LastUpdateTime:  time.Now(),
		ResourcesTotal:  make(map[string]int64),
	}

	// Load tech tree data
	if err := game.loadTechTree(); err != nil {
		return nil, fmt.Errorf("failed to load tech tree: %w", err)
	}

	// Initialize world
	world, err := NewWorld(settings, game.techTree, assetMgr)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize world: %w", err)
	}
	game.world = world

	return game, nil
}

// Start begins the game and starts the game loop
func (g *Game) Start() error {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	if g.isRunning {
		return fmt.Errorf("game is already running")
	}

	if g.state != GameStateLoading {
		return fmt.Errorf("game must be in loading state to start")
	}

	// Initialize world state
	if err := g.world.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize world: %w", err)
	}

	// Transition to playing state
	g.setState(GameStatePlaying)
	g.isRunning = true

	// Start game loop
	g.updateTicker = time.NewTicker(g.frameTime)
	go g.gameLoop()

	// Send game start event
	g.sendEvent(GameEvent{
		Type:      EventTypeGameStart,
		Timestamp: time.Now(),
		PlayerID:  -1,
		Message:   "Game started",
	})

	return nil
}

// Pause pauses the game if it's currently playing
func (g *Game) Pause() error {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	if g.state != GameStatePlaying {
		return fmt.Errorf("can only pause when game is playing")
	}

	g.setState(GameStatePaused)

	g.sendEvent(GameEvent{
		Type:      EventTypeGamePause,
		Timestamp: time.Now(),
		PlayerID:  -1,
		Message:   "Game paused",
	})

	return nil
}

// Resume resumes the game if it's currently paused
func (g *Game) Resume() error {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	if g.state != GameStatePaused {
		return fmt.Errorf("can only resume when game is paused")
	}

	g.setState(GameStatePlaying)
	g.lastUpdate = time.Now() // Reset timing to prevent large delta

	g.sendEvent(GameEvent{
		Type:      EventTypeGameResume,
		Timestamp: time.Now(),
		PlayerID:  -1,
		Message:   "Game resumed",
	})

	return nil
}

// Stop stops the game and cleans up resources
func (g *Game) Stop() error {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	if !g.isRunning {
		return fmt.Errorf("game is not running")
	}

	// Stop game loop
	g.isRunning = false
	if g.updateTicker != nil {
		g.updateTicker.Stop()
	}

	// Cancel context to signal shutdown
	g.cancel()

	// Set ended state
	g.setState(GameStateEnded)

	g.sendEvent(GameEvent{
		Type:      EventTypeGameEnd,
		Timestamp: time.Now(),
		PlayerID:  -1,
		Message:   "Game ended",
	})

	return nil
}

// GetState returns the current game state (thread-safe)
func (g *Game) GetState() GameState {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	return g.state
}

// GetSettings returns a copy of the game settings (thread-safe)
func (g *Game) GetSettings() GameSettings {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	return g.settings
}

// GetStats returns a copy of the current game statistics (thread-safe)
func (g *Game) GetStats() GameStats {
	g.mutex.RLock()
	defer g.mutex.RUnlock()

	// Update runtime statistics
	stats := g.stats
	stats.CurrentGameTime = time.Since(g.stats.StartTime)
	if g.world != nil {
		stats.PlayersActive = g.world.GetPlayerCount()
		stats.UnitsTotal = g.world.GetTotalUnitCount()
	}

	return stats
}

// GetWorld returns the game world (thread-safe access should use world methods)
func (g *Game) GetWorld() *World {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	return g.world
}

// GetTechTree returns the loaded tech tree data
func (g *Game) GetTechTree() *data.TechTree {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	return g.techTree
}

// GetEvents returns available events from the event queue (non-blocking)
func (g *Game) GetEvents() []GameEvent {
	events := make([]GameEvent, 0)

	// Drain available events from queue
	for {
		select {
		case event := <-g.eventQueue:
			events = append(events, event)
		default:
			return events
		}
	}
}

// Internal methods

// gameLoop runs the main game update loop
func (g *Game) gameLoop() {
	for g.isRunning {
		select {
		case <-g.ctx.Done():
			return
		case <-g.updateTicker.C:
			g.update()
		}
	}
}

// update performs one game update cycle
func (g *Game) update() {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	// Skip updates if not playing
	if g.state != GameStatePlaying {
		return
	}

	// Calculate delta time
	now := time.Now()
	deltaTime := now.Sub(g.lastUpdate)
	g.lastUpdate = now

	// Update frame statistics
	g.stats.FrameCount++
	if g.stats.FrameCount > 0 {
		totalTime := now.Sub(g.stats.StartTime)
		g.stats.AverageFrameTime = totalTime / time.Duration(g.stats.FrameCount)
	}
	g.stats.LastUpdateTime = now

	// Update world state
	if g.world != nil {
		g.world.Update(deltaTime)
	}

	// Process any pending events or game logic here
	// (placeholder for future expansion)
}

// setState changes the game state and handles transitions
func (g *Game) setState(newState GameState) {
	oldState := g.state
	g.state = newState

	// Handle state transition logic
	switch newState {
	case GameStatePlaying:
		// Game is now active
	case GameStatePaused:
		// Game is paused, stop certain systems
	case GameStateEnded:
		// Game has ended, begin cleanup
	}

	_ = oldState // Placeholder for future state transition logic
}

// sendEvent adds an event to the event queue
func (g *Game) sendEvent(event GameEvent) {
	// Non-blocking send to avoid deadlocks
	select {
	case g.eventQueue <- event:
		// Event queued successfully
	default:
		// Queue is full, could log this situation
	}
}

// sendResourceGainedEvent sends a resource gained event
func (g *Game) sendResourceGainedEvent(playerID int, resourceType string, amount int, source string) {
	resourceEvent := ResourceEvent{
		PlayerID:        playerID,
		ResourceType:    resourceType,
		Amount:          amount,
		Source:          source,
		Timestamp:       time.Now(),
		TransactionType: "addition",
	}

	event := GameEvent{
		Type:      EventTypeResourceGained,
		Timestamp: time.Now(),
		PlayerID:  playerID,
		Data:      resourceEvent,
		Message:   fmt.Sprintf("Player %d gained %d %s from %s", playerID, amount, resourceType, source),
	}

	g.sendEvent(event)
}

// sendResourceSpentEvent sends a resource spent event
func (g *Game) sendResourceSpentEvent(playerID int, resourceType string, amount int, purpose string) {
	resourceEvent := ResourceEvent{
		PlayerID:        playerID,
		ResourceType:    resourceType,
		Amount:          amount,
		Source:          purpose,
		Timestamp:       time.Now(),
		TransactionType: "deduction",
	}

	event := GameEvent{
		Type:      EventTypeResourceSpent,
		Timestamp: time.Now(),
		PlayerID:  playerID,
		Data:      resourceEvent,
		Message:   fmt.Sprintf("Player %d spent %d %s for %s", playerID, amount, resourceType, purpose),
	}

	g.sendEvent(event)
}

// sendPopulationLimitEvent sends a population limit reached event
func (g *Game) sendPopulationLimitEvent(playerID int, currentPop, maxPop int) {
	event := GameEvent{
		Type:      EventTypePopulationLimit,
		Timestamp: time.Now(),
		PlayerID:  playerID,
		Data:      map[string]int{"currentPop": currentPop, "maxPop": maxPop},
		Message:   fmt.Sprintf("Player %d has reached population limit: %d/%d", playerID, currentPop, maxPop),
	}

	g.sendEvent(event)
}

// sendResourceDepletedEvent sends a resource depleted event
func (g *Game) sendResourceDepletedEvent(resourceNodeID int, resourceType string, position Vector3) {
	event := GameEvent{
		Type:      EventTypeResourceDepleted,
		Timestamp: time.Now(),
		PlayerID:  -1, // System event
		Data:      map[string]interface{}{"nodeID": resourceNodeID, "position": position},
		Message:   fmt.Sprintf("%s resource node depleted at position (%.1f, %.1f, %.1f)", resourceType, position.X, position.Y, position.Z),
	}

	g.sendEvent(event)
}

// loadTechTree loads the tech tree data for the game
func (g *Game) loadTechTree() error {
	techTree, err := g.assetMgr.LoadTechTree()
	if err != nil {
		return fmt.Errorf("failed to load tech tree: %w", err)
	}
	g.techTree = techTree
	return nil
}

// validateGameSettings validates the provided game settings
func validateGameSettings(settings GameSettings) error {
	if settings.TechTreePath == "" {
		return fmt.Errorf("tech tree path cannot be empty")
	}

	if settings.GameSpeed <= 0 {
		settings.GameSpeed = 1.0
	}

	if settings.ResourceMultiplier <= 0 {
		settings.ResourceMultiplier = 1.0
	}

	if settings.MaxPlayers <= 0 {
		settings.MaxPlayers = 8
	}

	totalPlayers := len(settings.PlayerFactions) + len(settings.AIFactions)
	if totalPlayers > settings.MaxPlayers {
		return fmt.Errorf("total players (%d) exceeds maximum (%d)", totalPlayers, settings.MaxPlayers)
	}

	if totalPlayers == 0 {
		return fmt.Errorf("at least one player must be configured")
	}

	return nil
}

// String methods for enums
func (gs GameState) String() string {
	switch gs {
	case GameStateLoading:
		return "Loading"
	case GameStatePlaying:
		return "Playing"
	case GameStatePaused:
		return "Paused"
	case GameStateEnded:
		return "Ended"
	default:
		return "Unknown"
	}
}

func (et GameEventType) String() string {
	switch et {
	case EventTypeGameStart:
		return "GameStart"
	case EventTypeGamePause:
		return "GamePause"
	case EventTypeGameResume:
		return "GameResume"
	case EventTypeGameEnd:
		return "GameEnd"
	case EventTypeUnitCreated:
		return "UnitCreated"
	case EventTypeUnitDestroyed:
		return "UnitDestroyed"
	case EventTypeResourceGained:
		return "ResourceGained"
	case EventTypeResourceSpent:
		return "ResourceSpent"
	case EventTypeTechResearched:
		return "TechResearched"
	case EventTypeBuildingCompleted:
		return "BuildingCompleted"
	case EventTypePlayerDefeated:
		return "PlayerDefeated"
	case EventTypePlayerVictory:
		return "PlayerVictory"
	default:
		return "Unknown"
	}
}