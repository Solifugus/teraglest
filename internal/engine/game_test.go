package engine

import (
	"testing"
	"time"

	"teraglest/internal/data"
)

func TestNewGame(t *testing.T) {
	// Setup test environment
	techTreeRoot := "../../megaglest-source/data/glest_game/techs/megapack"
	assetMgr := data.NewAssetManager(techTreeRoot)

	settings := GameSettings{
		TechTreePath: techTreeRoot,
		PlayerFactions: map[int]string{
			1: "magic",
		},
		AIFactions: map[int]string{
			2: "tech",
		},
		GameSpeed:          1.0,
		ResourceMultiplier: 1.0,
		MaxPlayers:         8,
		EnableFogOfWar:     true,
		AllowCheats:        false,
	}

	// Test game creation
	game, err := NewGame(settings, assetMgr)
	if err != nil {
		t.Fatalf("Failed to create game: %v", err)
	}

	if game == nil {
		t.Error("NewGame returned nil")
	}

	// Test initial state
	if game.GetState() != GameStateLoading {
		t.Errorf("Expected initial state to be Loading, got %v", game.GetState())
	}

	// Test settings
	gameSettings := game.GetSettings()
	if gameSettings.TechTreePath != techTreeRoot {
		t.Errorf("Tech tree path not set correctly")
	}

	// Test tech tree was loaded
	techTree := game.GetTechTree()
	if techTree == nil {
		t.Error("Tech tree was not loaded")
	}

	// Test world was created
	world := game.GetWorld()
	if world == nil {
		t.Error("World was not created")
	}
}

func TestGameStateTransitions(t *testing.T) {
	game := createTestGame(t)

	// Test Start
	err := game.Start()
	if err != nil {
		t.Fatalf("Failed to start game: %v", err)
	}

	if game.GetState() != GameStatePlaying {
		t.Errorf("Expected state to be Playing after start, got %v", game.GetState())
	}

	// Test Pause
	err = game.Pause()
	if err != nil {
		t.Fatalf("Failed to pause game: %v", err)
	}

	if game.GetState() != GameStatePaused {
		t.Errorf("Expected state to be Paused after pause, got %v", game.GetState())
	}

	// Test Resume
	err = game.Resume()
	if err != nil {
		t.Fatalf("Failed to resume game: %v", err)
	}

	if game.GetState() != GameStatePlaying {
		t.Errorf("Expected state to be Playing after resume, got %v", game.GetState())
	}

	// Test Stop
	err = game.Stop()
	if err != nil {
		t.Fatalf("Failed to stop game: %v", err)
	}

	if game.GetState() != GameStateEnded {
		t.Errorf("Expected state to be Ended after stop, got %v", game.GetState())
	}
}

func TestGameStateValidation(t *testing.T) {
	game := createTestGame(t)

	// Test invalid state transitions
	err := game.Pause() // Can't pause when not playing
	if err == nil {
		t.Error("Expected error when pausing game that's not playing")
	}

	err = game.Resume() // Can't resume when not paused
	if err == nil {
		t.Error("Expected error when resuming game that's not paused")
	}

	err = game.Stop() // Can't stop when not running
	if err == nil {
		t.Error("Expected error when stopping game that's not running")
	}
}

func TestGameEvents(t *testing.T) {
	game := createTestGame(t)

	// Start game to generate events
	err := game.Start()
	if err != nil {
		t.Fatalf("Failed to start game: %v", err)
	}

	// Wait a bit for events to be generated
	time.Sleep(10 * time.Millisecond)

	// Get events
	events := game.GetEvents()

	// Should have at least one event (game start)
	if len(events) == 0 {
		t.Error("Expected at least one event after starting game")
	}

	// Check for game start event
	foundGameStart := false
	for _, event := range events {
		if event.Type == EventTypeGameStart {
			foundGameStart = true
			break
		}
	}

	if !foundGameStart {
		t.Error("Expected to find GameStart event")
	}

	// Test pause event
	err = game.Pause()
	if err != nil {
		t.Fatalf("Failed to pause game: %v", err)
	}

	events = game.GetEvents()
	foundGamePause := false
	for _, event := range events {
		if event.Type == EventTypeGamePause {
			foundGamePause = true
			break
		}
	}

	if !foundGamePause {
		t.Error("Expected to find GamePause event")
	}

	// Cleanup
	game.Stop()
}

func TestGameStats(t *testing.T) {
	game := createTestGame(t)

	// Get initial stats
	stats := game.GetStats()

	// Test that start time is set
	if stats.StartTime.IsZero() {
		t.Error("Start time should be set")
	}

	// Test frame count starts at 0
	if stats.FrameCount != 0 {
		t.Errorf("Expected initial frame count to be 0, got %d", stats.FrameCount)
	}

	// Start game and let it run briefly
	err := game.Start()
	if err != nil {
		t.Fatalf("Failed to start game: %v", err)
	}

	// Wait for a few frames
	time.Sleep(100 * time.Millisecond)

	// Get updated stats
	stats = game.GetStats()

	// Check that frame count increased
	if stats.FrameCount == 0 {
		t.Error("Frame count should have increased after running")
	}

	// Check that current game time is positive
	if stats.CurrentGameTime <= 0 {
		t.Error("Current game time should be positive")
	}

	// Check that players are counted
	if stats.PlayersActive != 2 { // We have 2 players in test
		t.Errorf("Expected 2 active players, got %d", stats.PlayersActive)
	}

	// Cleanup
	game.Stop()
}

func TestGameSettingsValidation(t *testing.T) {
	techTreeRoot := "../../megaglest-source/data/glest_game/techs/megapack"
	assetMgr := data.NewAssetManager(techTreeRoot)

	// Test empty tech tree path
	settings := GameSettings{}
	_, err := NewGame(settings, assetMgr)
	if err == nil {
		t.Error("Expected error with empty tech tree path")
	}

	// Test no players
	settings = GameSettings{
		TechTreePath:   techTreeRoot,
		PlayerFactions: map[int]string{},
		AIFactions:     map[int]string{},
	}
	_, err = NewGame(settings, assetMgr)
	if err == nil {
		t.Error("Expected error with no players")
	}

	// Test too many players
	settings = GameSettings{
		TechTreePath: techTreeRoot,
		PlayerFactions: map[int]string{
			1: "magic", 2: "tech", 3: "magic", 4: "tech",
			5: "magic", 6: "tech", 7: "magic", 8: "tech",
			9: "magic", // One too many
		},
		MaxPlayers: 8,
	}
	_, err = NewGame(settings, assetMgr)
	if err == nil {
		t.Error("Expected error with too many players")
	}
}

func TestGameStateString(t *testing.T) {
	tests := []struct {
		state    GameState
		expected string
	}{
		{GameStateLoading, "Loading"},
		{GameStatePlaying, "Playing"},
		{GameStatePaused, "Paused"},
		{GameStateEnded, "Ended"},
	}

	for _, test := range tests {
		result := test.state.String()
		if result != test.expected {
			t.Errorf("Expected %s, got %s", test.expected, result)
		}
	}
}

func TestGameEventTypeString(t *testing.T) {
	tests := []struct {
		eventType GameEventType
		expected  string
	}{
		{EventTypeGameStart, "GameStart"},
		{EventTypeGamePause, "GamePause"},
		{EventTypeGameResume, "GameResume"},
		{EventTypeGameEnd, "GameEnd"},
		{EventTypeUnitCreated, "UnitCreated"},
		{EventTypeUnitDestroyed, "UnitDestroyed"},
		{EventTypeResourceGained, "ResourceGained"},
		{EventTypeResourceSpent, "ResourceSpent"},
	}

	for _, test := range tests {
		result := test.eventType.String()
		if result != test.expected {
			t.Errorf("Expected %s, got %s", test.expected, result)
		}
	}
}

func TestGameConcurrency(t *testing.T) {
	game := createTestGame(t)

	// Start the game
	err := game.Start()
	if err != nil {
		t.Fatalf("Failed to start game: %v", err)
	}

	// Test concurrent access to game state
	done := make(chan bool, 3)

	// Goroutine 1: Read state repeatedly
	go func() {
		for i := 0; i < 10; i++ {
			_ = game.GetState()
			_ = game.GetStats()
			_ = game.GetWorld()
			time.Sleep(1 * time.Millisecond)
		}
		done <- true
	}()

	// Goroutine 2: Read events
	go func() {
		for i := 0; i < 10; i++ {
			_ = game.GetEvents()
			time.Sleep(1 * time.Millisecond)
		}
		done <- true
	}()

	// Goroutine 3: State transitions
	go func() {
		time.Sleep(20 * time.Millisecond)
		game.Pause()
		time.Sleep(10 * time.Millisecond)
		game.Resume()
		time.Sleep(10 * time.Millisecond)
		done <- true
	}()

	// Wait for all goroutines to complete
	for i := 0; i < 3; i++ {
		<-done
	}

	// Cleanup
	game.Stop()
}

func TestGamePerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	game := createTestGame(t)

	// Start the game
	err := game.Start()
	if err != nil {
		t.Fatalf("Failed to start game: %v", err)
	}

	// Let it run for a bit
	time.Sleep(500 * time.Millisecond)

	// Check performance metrics
	stats := game.GetStats()

	// Should have processed many frames
	if stats.FrameCount < 10 {
		t.Errorf("Expected at least 10 frames, got %d", stats.FrameCount)
	}

	// Average frame time should be reasonable (less than 100ms)
	if stats.AverageFrameTime > 100*time.Millisecond {
		t.Errorf("Average frame time too high: %v", stats.AverageFrameTime)
	}

	// Cleanup
	game.Stop()
}

// Helper function to create a test game
func createTestGame(t *testing.T) *Game {
	techTreeRoot := "../../megaglest-source/data/glest_game/techs/megapack"
	assetMgr := data.NewAssetManager(techTreeRoot)

	settings := GameSettings{
		TechTreePath: techTreeRoot,
		PlayerFactions: map[int]string{
			1: "magic",
		},
		AIFactions: map[int]string{
			2: "tech",
		},
		GameSpeed:          1.0,
		ResourceMultiplier: 1.0,
		MaxPlayers:         8,
		EnableFogOfWar:     true,
		AllowCheats:        false,
	}

	game, err := NewGame(settings, assetMgr)
	if err != nil {
		t.Fatalf("Failed to create test game: %v", err)
	}

	return game
}