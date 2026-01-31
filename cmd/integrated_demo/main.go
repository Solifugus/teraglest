package main

import (
	"fmt"
	"log"
	"path/filepath"
	"time"

	"teraglest/internal/data"
	"teraglest/internal/engine"
	"teraglest/internal/graphics/renderer"

	"github.com/go-gl/glfw/v3.3/glfw"
)

func main() {
	fmt.Println("=== TeraGlest Integrated Game + Renderer Demo ===")
	fmt.Println("Game engine with visual rendering integration test")
	fmt.Println()

	// Initialize asset manager
	dataRoot := filepath.Join("megaglest-source", "data", "glest_game")
	assetManager := data.NewAssetManager(filepath.Join(dataRoot, "techs", "megapack"))

	fmt.Printf("✅ AssetManager initialized with data root: %s\n", dataRoot)

	// Create renderer
	r, err := renderer.NewRenderer(assetManager, "TeraGlest Integrated Demo", 1024, 768)
	if err != nil {
		log.Fatalf("Failed to create renderer: %v", err)
	}
	defer r.Destroy()

	fmt.Println("✅ OpenGL renderer initialized")

	// Create game with minimal settings
	gameSettings := engine.GameSettings{
		TechTreePath:       filepath.Join(dataRoot, "techs", "megapack", "megapack.xml"),
		MaxPlayers:         2,
		GameSpeed:          1.0,
		ResourceMultiplier: 1.0,
		PlayerFactions: map[int]string{
			1: "magic",
		},
	}

	game, err := engine.NewGame(gameSettings, assetManager)
	if err != nil {
		log.Fatalf("Failed to create game: %v", err)
	}

	fmt.Println("✅ Game engine initialized")

	// Start the game (this launches the background update loop)
	err = game.Start()
	if err != nil {
		log.Fatalf("Failed to start game: %v", err)
	}

	fmt.Println("✅ Game started - background loop running")
	fmt.Println()

	// Wait for game to initialize
	time.Sleep(200 * time.Millisecond)

	// Get world reference for rendering
	world := game.GetWorld()
	if world == nil {
		log.Fatalf("Failed to get game world")
	}

	fmt.Printf("✅ Game world available: %dx%d with %d players\n",
		world.Width, world.Height, len(world.GetAllPlayers()))

	// Create some test units to see in the renderer
	players := world.GetAllPlayers()
	if len(players) > 0 {
		player := players[0]

		// Create a few test units
		positions := []engine.Vector3{
			{X: 15, Y: 0, Z: 15},
			{X: 20, Y: 0, Z: 15},
			{X: 25, Y: 0, Z: 20},
		}

		for i, pos := range positions {
			unit, err := world.ObjectManager.CreateUnit(player.ID, "initiate", pos, nil)
			if err != nil {
				log.Printf("Warning: Failed to create test unit %d: %v", i, err)
			} else {
				fmt.Printf("✅ Created test unit %d at position (%.1f, %.1f, %.1f)\n",
					unit.GetID(), pos.X, pos.Y, pos.Z)
			}
		}
	}

	fmt.Println()
	fmt.Println("🎮 Starting integrated game loop...")
	fmt.Println("   Game Update + Renderer Integration")
	fmt.Println()

	frameCount := 0
	startTime := time.Now()

	// Integrated game loop: combines game logic with rendering
	for !r.ShouldClose() {
		frameStart := time.Now()

		// The game.Start() already launched a background loop for Game.update()
		// Here we just need to render the current game state

		// Render the entire game world
		err := r.RenderWorld(world)
		if err != nil {
			log.Printf("Render error: %v", err)
			break
		}

		frameCount++

		// Print stats every 2 seconds
		if frameCount%120 == 0 {
			elapsed := time.Since(startTime)
			fps := float64(frameCount) / elapsed.Seconds()

			// Get current game stats
			gameStats := game.GetStats()

			fmt.Printf("🎯 [%.1fs] FPS: %.1f, Game Frames: %d, Game Time: %v\n",
				elapsed.Seconds(), fps, gameStats.FrameCount, gameStats.CurrentGameTime)

			// Show game state info
			allUnits := 0
			for _, player := range world.GetAllPlayers() {
				units := world.ObjectManager.GetUnitsForPlayer(player.ID)
				buildings := world.ObjectManager.GetBuildingsForPlayer(player.ID)
				allUnits += len(units)

				fmt.Printf("   Player %d: %d units, %d buildings\n", player.ID, len(units), len(buildings))
			}
		}

		// Handle GLFW events
		glfw.PollEvents()

		// Maintain reasonable frame rate
		frameDuration := time.Since(frameStart)
		if frameDuration < 16*time.Millisecond { // ~60 FPS cap
			time.Sleep(16*time.Millisecond - frameDuration)
		}
	}

	// Stop the game
	game.Stop()

	elapsed := time.Since(startTime)
	avgFPS := float64(frameCount) / elapsed.Seconds()

	fmt.Println()
	fmt.Printf("✅ Demo completed after %.1f seconds\n", elapsed.Seconds())
	fmt.Printf("   Total frames rendered: %d\n", frameCount)
	fmt.Printf("   Average FPS: %.1f\n", avgFPS)
	fmt.Println()
	fmt.Println("🎉 Integrated Game + Renderer Demo Complete!")
	fmt.Println("   ✅ Game logic running in background")
	fmt.Println("   ✅ Visual rendering of game state")
	fmt.Println("   ✅ Real-time game world display")
}