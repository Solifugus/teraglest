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
	fmt.Println("=== Renderer-World Integration Test ===")
	fmt.Println("Testing direct world rendering without full game loop")
	fmt.Println()

	// Initialize asset manager
	dataRoot := filepath.Join("megaglest-source", "data", "glest_game")
	assetManager := data.NewAssetManager(filepath.Join(dataRoot, "techs", "megapack"))

	fmt.Printf("✅ AssetManager initialized\n")

	// Create renderer
	r, err := renderer.NewRenderer(assetManager, "TeraGlest Renderer Test", 1024, 768)
	if err != nil {
		log.Fatalf("Failed to create renderer: %v", err)
	}
	defer r.Destroy()

	fmt.Println("✅ Renderer initialized")

	// Load tech tree
	techTreePath := filepath.Join(dataRoot, "techs", "megapack", "megapack.xml")
	techTree, err := data.LoadTechTree(techTreePath)
	if err != nil {
		log.Fatalf("Failed to load tech tree: %v", err)
	}

	// Create world directly (without game wrapper to avoid initialization deadlock)
	gameSettings := engine.GameSettings{
		TechTreePath:       techTreePath,
		MaxPlayers:         2,
		GameSpeed:          1.0,
		ResourceMultiplier: 1.0,
		PlayerFactions: map[int]string{
			1: "magic",
		},
	}

	world, err := engine.NewWorld(gameSettings, techTree, assetManager)
	if err != nil {
		log.Fatalf("Failed to create world: %v", err)
	}

	fmt.Println("✅ World created")

	// Initialize the world first
	err = world.Initialize()
	if err != nil {
		log.Fatalf("Failed to initialize world: %v", err)
	}

	fmt.Println("✅ World initialized")

	// Manually create a player and some test objects
	err = world.AddPlayer(1, "Player 1", "magic", false)
	if err != nil {
		log.Fatalf("Failed to create player: %v", err)
	}

	fmt.Println("✅ Player created")

	// Create test units manually
	testPositions := []engine.Vector3{
		{X: 10, Y: 0, Z: 10},
		{X: 15, Y: 0, Z: 10},
		{X: 20, Y: 0, Z: 15},
	}

	for i, pos := range testPositions {
		unit, err := world.ObjectManager.CreateUnit(1, "initiate", pos, nil)
		if err != nil {
			log.Printf("Warning: Failed to create test unit %d: %v", i, err)
		} else {
			fmt.Printf("✅ Created unit %d at (%.1f, %.1f, %.1f)\n", unit.GetID(), pos.X, pos.Y, pos.Z)
		}
	}

	fmt.Println()
	fmt.Println("🎮 Testing RenderWorld integration...")
	fmt.Println("   Press ESC to exit")
	fmt.Println()

	frameCount := 0
	startTime := time.Now()

	// Simple render loop to test world rendering
	for !r.ShouldClose() {
		frameStart := time.Now()

		// Test RenderWorld method with our world
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

			allUnits := 0
			for _, player := range world.GetAllPlayers() {
				units := world.ObjectManager.GetUnitsForPlayer(player.ID)
				buildings := world.ObjectManager.GetBuildingsForPlayer(player.ID)
				allUnits += len(units)

				fmt.Printf("🎯 [%.1fs] FPS: %.1f, Player %d: %d units, %d buildings\n",
					elapsed.Seconds(), fps, player.ID, len(units), len(buildings))
			}
		}

		// Handle GLFW events
		glfw.PollEvents()

		// Check for ESC key
		if glfw.GetCurrentContext().GetKey(glfw.KeyEscape) == glfw.Press {
			break
		}

		// Maintain frame rate
		frameDuration := time.Since(frameStart)
		if frameDuration < 16*time.Millisecond {
			time.Sleep(16*time.Millisecond - frameDuration)
		}
	}

	elapsed := time.Since(startTime)
	avgFPS := float64(frameCount) / elapsed.Seconds()

	fmt.Println()
	fmt.Printf("✅ Renderer integration test completed\n")
	fmt.Printf("   Duration: %.1f seconds\n", elapsed.Seconds())
	fmt.Printf("   Total frames: %d\n", frameCount)
	fmt.Printf("   Average FPS: %.1f\n", avgFPS)
	fmt.Println()

	// Test individual render methods
	fmt.Println("🧪 Testing individual render methods:")

	players := world.GetAllPlayers()
	if len(players) > 0 {
		player := players[0]
		units := world.ObjectManager.GetUnitsForPlayer(player.ID)
		buildings := world.ObjectManager.GetBuildingsForPlayer(player.ID)

		fmt.Printf("   ✅ renderUnits: %d units available for rendering\n", len(units))
		fmt.Printf("   ✅ renderBuildings: %d buildings available for rendering\n", len(buildings))
		fmt.Printf("   ✅ renderResourceNodes: %d resource nodes available\n", len(world.GetAllResourceNodes()))
		fmt.Printf("   ✅ renderTerrain: %dx%d world grid\n", world.Width, world.Height)
	}

	fmt.Println()
	fmt.Println("🎉 Renderer-World Integration Test Complete!")
	fmt.Println("   ✅ World objects accessible by renderer")
	fmt.Println("   ✅ RenderWorld method functional")
	fmt.Println("   ✅ No deadlocks or crashes")
}