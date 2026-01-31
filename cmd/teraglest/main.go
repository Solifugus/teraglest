package main

import (
	"fmt"
	"log"
	"path/filepath"
	"runtime"
	"time"

	"teraglest/internal/audio"
	"teraglest/internal/data"
	"teraglest/internal/engine"
	"teraglest/internal/graphics/renderer"
	"teraglest/internal/ui"

	"github.com/go-gl/glfw/v3.3/glfw"
)

// GameConfig holds configuration for the main game
type GameConfig struct {
	WindowTitle    string
	WindowWidth    int
	WindowHeight   int
	DataRoot       string
	AudioEnabled   bool
	VsyncEnabled   bool
	TargetFPS      int
}

// DefaultGameConfig returns a default configuration
func DefaultGameConfig() GameConfig {
	return GameConfig{
		WindowTitle:    "TeraGlest - Real-Time Strategy Game",
		WindowWidth:    1024,
		WindowHeight:   768,
		DataRoot:       filepath.Join("megaglest-source", "data", "glest_game"),
		AudioEnabled:   true,
		VsyncEnabled:   true,
		TargetFPS:      60,
	}
}

// TeraGlest represents the main game application
type TeraGlest struct {
	config       GameConfig
	assetManager *data.AssetManager
	renderer     *renderer.Renderer
	game         *engine.Game
	world        *engine.World
	inputHandler *ui.InputHandler
	uiManager    *ui.SimpleUIManager
	audioManager *audio.AudioManager

	// Performance tracking
	frameCount   int64
	lastFPSCheck time.Time
	currentFPS   float64

	// Timing
	lastFrameTime time.Time
	frameTime     time.Duration

	// Game state
	running      bool
	paused       bool
}

// NewTeraGlest creates a new game instance
func NewTeraGlest(config GameConfig) (*TeraGlest, error) {
	tg := &TeraGlest{
		config:        config,
		lastFPSCheck:  time.Now(),
		lastFrameTime: time.Now(),
		running:       true,
		paused:        false,
	}

	// Initialize GLFW (done before other systems)
	if err := tg.initializeGLFW(); err != nil {
		return nil, fmt.Errorf("failed to initialize GLFW: %v", err)
	}

	// Initialize asset manager
	if err := tg.initializeAssetManager(); err != nil {
		return nil, fmt.Errorf("failed to initialize asset manager: %v", err)
	}

	// Initialize renderer
	if err := tg.initializeRenderer(); err != nil {
		return nil, fmt.Errorf("failed to initialize renderer: %v", err)
	}

	// Initialize audio system
	if config.AudioEnabled {
		if err := tg.initializeAudio(); err != nil {
			log.Printf("Warning: Audio initialization failed: %v", err)
			// Continue without audio
		}
	}

	// Initialize game engine
	if err := tg.initializeGame(); err != nil {
		return nil, fmt.Errorf("failed to initialize game: %v", err)
	}

	// Initialize UI and input systems
	if err := tg.initializeUI(); err != nil {
		return nil, fmt.Errorf("failed to initialize UI: %v", err)
	}

	log.Printf("TeraGlest initialized successfully")
	log.Printf("  Window: %dx%d", config.WindowWidth, config.WindowHeight)
	log.Printf("  Audio: %v", config.AudioEnabled)
	log.Printf("  Target FPS: %d", config.TargetFPS)

	return tg, nil
}

// initializeGLFW initializes the GLFW library
func (tg *TeraGlest) initializeGLFW() error {
	runtime.LockOSThread() // Required for OpenGL context

	err := glfw.Init()
	if err != nil {
		return fmt.Errorf("failed to initialize GLFW: %v", err)
	}

	return nil
}

// initializeAssetManager initializes the asset management system
func (tg *TeraGlest) initializeAssetManager() error {
	techPath := filepath.Join(tg.config.DataRoot, "techs", "megapack")
	tg.assetManager = data.NewAssetManager(techPath)

	log.Printf("Asset manager initialized with path: %s", techPath)
	return nil
}

// initializeRenderer initializes the rendering system
func (tg *TeraGlest) initializeRenderer() error {
	var err error
	tg.renderer, err = renderer.NewRenderer(
		tg.assetManager,
		tg.config.WindowTitle,
		tg.config.WindowWidth,
		tg.config.WindowHeight,
	)
	if err != nil {
		return err
	}

	// Configure renderer settings
	if tg.config.VsyncEnabled {
		glfw.SwapInterval(1) // Enable VSync
	} else {
		glfw.SwapInterval(0) // Disable VSync
	}

	log.Printf("Renderer initialized: %dx%d", tg.config.WindowWidth, tg.config.WindowHeight)
	return nil
}

// initializeAudio initializes the audio system
func (tg *TeraGlest) initializeAudio() error {
	// Create mock backend for now (can be replaced with real audio backend)
	backend := &audio.MockAudioBackend{}

	var err error
	tg.audioManager, err = audio.NewAudioManager(backend)
	if err != nil {
		return err
	}

	log.Printf("Audio system initialized with mock backend")
	return nil
}

// initializeGame initializes the game engine and world
func (tg *TeraGlest) initializeGame() error {
	// Create game settings
	gameSettings := engine.GameSettings{
		TechTreePath:       filepath.Join(tg.config.DataRoot, "techs", "megapack", "megapack.xml"),
		MaxPlayers:         1, // Start with single player
		GameSpeed:          1.0,
		ResourceMultiplier: 1.0,
		PlayerFactions: map[int]string{
			1: "magic", // Default to magic faction
		},
	}

	// Create game instance
	var err error
	tg.game, err = engine.NewGame(gameSettings, tg.assetManager)
	if err != nil {
		return fmt.Errorf("failed to create game: %v", err)
	}

	// Start the game
	err = tg.game.Start()
	if err != nil {
		return fmt.Errorf("failed to start game: %v", err)
	}

	// Get world reference
	tg.world = tg.game.GetWorld()
	if tg.world == nil {
		return fmt.Errorf("game world is nil after start")
	}

	log.Printf("Game initialized: World %dx%d", tg.world.Width, tg.world.Height)
	return nil
}

// initializeUI initializes the UI and input systems
func (tg *TeraGlest) initializeUI() error {
	// Create simple UI manager (without ImGui dependencies)
	tg.uiManager = ui.NewSimpleUIManager(tg.world)

	// Create input handler
	tg.inputHandler = ui.NewInputHandler(tg.world, tg.uiManager)
	tg.inputHandler.SetCamera(tg.renderer.GetCamera())
	tg.inputHandler.SetScreenDimensions(tg.config.WindowWidth, tg.config.WindowHeight)

	// Setup input callbacks in renderer
	tg.renderer.SetupGameInputCallbacks(tg.inputHandler)

	log.Printf("UI and input systems initialized")
	return nil
}

// main entry point
func main() {
	// Print startup information
	fmt.Println("TeraGlest - Real-Time Strategy Game")
	fmt.Printf("Version: Development Build\n")
	fmt.Printf("Go Runtime: %s\n", runtime.Version())
	fmt.Println()

	// Create game configuration
	config := DefaultGameConfig()

	// TODO: Parse command line arguments to override config
	// For now, use defaults

	// Create and run game
	game, err := NewTeraGlest(config)
	if err != nil {
		log.Fatalf("Failed to create game: %v", err)
	}

	// Run the game
	err = game.Run()
	if err != nil {
		log.Fatalf("Game error: %v", err)
	}

	fmt.Println("TeraGlest exited successfully")
}

// Run starts the main game loop
func (tg *TeraGlest) Run() error {
	defer tg.Cleanup()

	log.Printf("Starting TeraGlest main game loop...")

	// Display initial status
	tg.printGameStatus()

	// Calculate frame duration for target FPS
	targetFrameTime := time.Duration(1000/tg.config.TargetFPS) * time.Millisecond

	// Main game loop
	for tg.running && !tg.renderer.ShouldClose() {
		frameStart := time.Now()

		// Update frame timing
		tg.frameTime = frameStart.Sub(tg.lastFrameTime)
		tg.lastFrameTime = frameStart

		// Process window events (input)
		glfw.PollEvents()

		// Update game logic (if not paused)
		if !tg.paused {
			tg.updateGame(tg.frameTime)
		}

		// Render frame
		tg.render()

		// Update performance metrics
		tg.updatePerformanceMetrics()

		// Frame rate limiting
		frameEnd := time.Now()
		frameDuration := frameEnd.Sub(frameStart)

		if frameDuration < targetFrameTime {
			sleepTime := targetFrameTime - frameDuration
			time.Sleep(sleepTime)
		}
	}

	log.Printf("Main game loop ended")
	return nil
}

// updateGame updates all game systems
func (tg *TeraGlest) updateGame(deltaTime time.Duration) {
	// Note: Game engine runs its own internal loop, we don't update it directly
	// The game automatically updates itself when started

	// Update UI manager
	tg.uiManager.Update(deltaTime)

	// Update audio system
	if tg.audioManager != nil {
		// Get camera position for spatial audio
		camera := tg.renderer.GetCamera()
		position := audio.Vector3{
			X: camera.Position.X(),
			Y: camera.Position.Y(),
			Z: camera.Position.Z(),
		}
		tg.audioManager.SetListenerPosition(position)

		// Note: AudioManager uses internal update loop
	}

	// Process any game events for audio
	tg.processAudioEvents()
}

// render renders the current frame
func (tg *TeraGlest) render() {
	// Render the world
	err := tg.renderer.RenderWorld(tg.world)
	if err != nil {
		log.Printf("Render error: %v", err)
	}

	// Render UI elements
	tg.renderUI()
}

// renderUI renders UI overlays
func (tg *TeraGlest) renderUI() {
	// TODO: Implement selection box rendering
	// selectionBox := tg.inputHandler.GetSelectionBox()
	// if selectionBox.Active {
	//     tg.renderer.RenderSelectionBox(selectionBox)
	// }

	// Render UI manager components
	tg.uiManager.Render()

	// Render UI elements (health bars, resource counts, etc.)
	tg.renderGameUI()
}

// renderGameUI renders game-specific UI elements
func (tg *TeraGlest) renderGameUI() {
	// Get selected units for UI display
	selectedUnits := tg.uiManager.GetSelectedUnits()

	// TODO: Implement UI rendering for:
	// - Resource counters
	// - Selected unit information
	// - Minimap
	// - Command buttons

	// For now, just track selection count in console
	if len(selectedUnits) > 0 && tg.frameCount%180 == 0 { // Every 3 seconds at 60fps
		log.Printf("Selected units: %d", len(selectedUnits))
	}
}

// processAudioEvents processes game events for audio feedback
func (tg *TeraGlest) processAudioEvents() {
	if tg.audioManager == nil {
		return
	}

	// Get recently selected units for audio feedback
	selectedUnits := tg.uiManager.GetSelectedUnits()
	if len(selectedUnits) > 0 {
		// Play selection sound (every 60 frames = 1 second at 60fps)
		if tg.frameCount%60 == 0 {
			for _, unit := range selectedUnits {
				position := audio.Vector3{
					X: float32(unit.Position.X),
					Y: float32(unit.Position.Y),
					Z: float32(unit.Position.Z),
				}
				event := audio.AudioEvent{
					Type:     audio.AudioEventUIClick, // Use existing UI event type for now
					Position: &position,
					Volume:   1.0,
					Pitch:    1.0,
					Loop:     false,
					Metadata: map[string]interface{}{"unit_type": unit.UnitType},
					Timestamp: time.Now(),
				}
				tg.audioManager.TriggerEvent(audio.AudioEventUIClick, event)
			}
		}
	}

	// Process building completion events, combat events, etc.
	// TODO: Implement game event system integration
}

// updatePerformanceMetrics updates FPS and performance tracking
func (tg *TeraGlest) updatePerformanceMetrics() {
	tg.frameCount++

	// Update FPS every second
	now := time.Now()
	if now.Sub(tg.lastFPSCheck) >= time.Second {
		elapsed := now.Sub(tg.lastFPSCheck).Seconds()
		framesSinceLastCheck := tg.frameCount
		tg.currentFPS = float64(framesSinceLastCheck) / elapsed

		tg.lastFPSCheck = now
		tg.frameCount = 0

		// Log performance every 5 seconds
		if int(elapsed*5)%5 == 0 {
			tg.logPerformanceStats()
		}
	}
}

// logPerformanceStats logs current performance statistics
func (tg *TeraGlest) logPerformanceStats() {
	// Get unit and building counts
	allPlayers := tg.world.GetAllPlayers()
	totalUnits := 0
	totalBuildings := 0

	for _, player := range allPlayers {
		units := tg.world.ObjectManager.GetUnitsForPlayer(player.ID)
		buildings := tg.world.ObjectManager.GetBuildingsForPlayer(player.ID)
		totalUnits += len(units)
		totalBuildings += len(buildings)
	}

	selectedCount := len(tg.uiManager.GetSelectedUnits())

	log.Printf("Performance: %.1f FPS | Units: %d | Buildings: %d | Selected: %d | Frame time: %.2fms",
		tg.currentFPS,
		totalUnits,
		totalBuildings,
		selectedCount,
		float64(tg.frameTime.Nanoseconds())/1000000.0,
	)
}

// printGameStatus prints initial game status information
func (tg *TeraGlest) printGameStatus() {
	fmt.Println()
	fmt.Println("=== TeraGlest Game Started ===")
	fmt.Printf("Window: %dx%d\n", tg.config.WindowWidth, tg.config.WindowHeight)
	fmt.Printf("World: %dx%d\n", tg.world.Width, tg.world.Height)

	allPlayers := tg.world.GetAllPlayers()
	fmt.Printf("Players: %d\n", len(allPlayers))

	for _, player := range allPlayers {
		units := tg.world.ObjectManager.GetUnitsForPlayer(player.ID)
		buildings := tg.world.ObjectManager.GetBuildingsForPlayer(player.ID)
		fmt.Printf("  Player %d: %d units, %d buildings\n",
			player.ID, len(units), len(buildings))
	}

	fmt.Println()
	fmt.Println("🎮 Controls:")
	fmt.Println("  Left Click: Select unit or move selected units")
	fmt.Println("  Right Click: Move/Attack/Gather command")
	fmt.Println("  Drag: Box selection")
	fmt.Println("  Ctrl+A: Select all units")
	fmt.Println("  S: Stop selected units")
	fmt.Println("  H: Hold position")
	fmt.Println("  P: Pause/Resume game")
	fmt.Println("  ESC: Exit game")
	fmt.Println("=== Game Running ===")
	fmt.Println()
}

// Cleanup cleans up all game resources
func (tg *TeraGlest) Cleanup() {
	log.Printf("Cleaning up TeraGlest...")

	if tg.game != nil {
		tg.game.Stop()
	}

	if tg.audioManager != nil {
		tg.audioManager.Shutdown()
	}

	if tg.renderer != nil {
		tg.renderer.Destroy()
	}

	glfw.Terminate()
	log.Printf("TeraGlest cleanup complete")
}