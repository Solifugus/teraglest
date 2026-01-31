package main

import (
	"log"
	"time"

	"teraglest/internal/data"
	"teraglest/internal/engine"
	"teraglest/internal/graphics/renderer"
	"teraglest/internal/ui"

	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/inkyblackness/imgui-go/v4"
)

func main() {
	// Create asset manager
	assetManager := data.NewAssetManager("")

	// Create game settings
	settings := engine.GameSettings{
		PlayerFactions:     map[int]string{1: "romans"},
		MaxPlayers:         2,
		ResourceMultiplier: 1.0,
	}

	// Create world
	world, err := engine.NewWorld(settings, &data.TechTree{}, assetManager)
	if err != nil {
		log.Fatalf("Failed to create world: %v", err)
	}

	// Initialize world
	err = world.Initialize()
	if err != nil {
		log.Fatalf("Failed to initialize world: %v", err)
	}

	// Create test units and buildings for demonstration
	createTestGameObjects(world)

	// Create 3D renderer
	gameRenderer, err := renderer.NewRenderer(assetManager, "TeraGlest UI Test", 1024, 768)
	if err != nil {
		log.Fatalf("Failed to create renderer: %v", err)
	}
	defer gameRenderer.Destroy()

	// Create UI manager
	uiManager, err := ui.NewUIManager(world, gameRenderer)
	if err != nil {
		log.Fatalf("Failed to create UI manager: %v", err)
	}
	defer uiManager.Cleanup()

	// Create input handler
	inputHandler := ui.NewInputHandler(world, uiManager)

	// Get GLFW window for input setup
	window := gameRenderer.GetContext().GetWindow()

	// Set up input callbacks
	window.SetMouseButtonCallback(inputHandler.HandleMouseButton)
	window.SetCursorPosCallback(inputHandler.HandleMouseMove)
	window.SetKeyCallback(inputHandler.HandleKeyboard)

	// Game loop
	lastTime := time.Now()

	log.Println("Starting UI test...")
	log.Println("Controls:")
	log.Println("  - Left click: Select units/buildings")
	log.Println("  - Drag: Area select")
	log.Println("  - Right click: Issue commands")
	log.Println("  - F1: Toggle debug info")
	log.Println("  - M: Toggle minimap")
	log.Println("  - ESC: Exit")

	for !gameRenderer.ShouldClose() {
		// Calculate delta time
		currentTime := time.Now()
		deltaTime := currentTime.Sub(lastTime)
		lastTime = currentTime

		// Update world
		world.Update(deltaTime)

		// Process UI input
		uiManager.ProcessInput(window)

		// Update UI
		uiManager.Update(deltaTime)

		// Clear screen and render 3D world
		err = gameRenderer.RenderFrame()
		if err != nil {
			log.Printf("Render error: %v", err)
		}

		// Render UI overlay
		uiManager.Render()

		// Swap buffers and poll events
		window.SwapBuffers()
		glfw.PollEvents()

		// Limit frame rate
		time.Sleep(16 * time.Millisecond) // ~60 FPS
	}

	log.Println("UI test completed")
}

// createTestGameObjects creates some test units and buildings for the UI demo
func createTestGameObjects(world *engine.World) {
	// Add resources to player
	world.AddResources(1, map[string]int{
		"wood":   1000,
		"gold":   500,
		"stone":  300,
		"energy": 200,
		"food":   150,
	}, "test_setup")

	// Create test units
	unitPositions := []engine.Vector3{
		{X: 10, Y: 0, Z: 10},
		{X: 12, Y: 0, Z: 10},
		{X: 14, Y: 0, Z: 10},
		{X: 10, Y: 0, Z: 12},
		{X: 12, Y: 0, Z: 12},
	}

	unitTypes := []string{"worker", "swordman", "archer", "worker", "swordman"}

	for i, pos := range unitPositions {
		unit, err := world.ObjectManager.CreateUnit(1, unitTypes[i], pos, nil)
		if err != nil {
			log.Printf("Failed to create unit %s: %v", unitTypes[i], err)
			continue
		}
		log.Printf("Created %s at (%.1f, %.1f, %.1f)", unitTypes[i], pos.X, pos.Y, pos.Z)
		_ = unit
	}

	// Create test buildings
	buildingPositions := []engine.Vector3{
		{X: 20, Y: 0, Z: 20},
		{X: 25, Y: 0, Z: 20},
		{X: 30, Y: 0, Z: 25},
	}

	buildingTypes := []string{"barracks", "house", "mage_tower"}

	for i, pos := range buildingPositions {
		building, err := world.ObjectManager.CreateBuilding(1, buildingTypes[i], pos, nil)
		if err != nil {
			log.Printf("Failed to create building %s: %v", buildingTypes[i], err)
			continue
		}

		// Make buildings completed for testing
		building.IsBuilt = true
		building.BuildProgress = 1.0

		log.Printf("Created %s at (%.1f, %.1f, %.1f)", buildingTypes[i], pos.X, pos.Y, pos.Z)
	}

	// Create some resource nodes
	resourcePositions := []engine.Vector3{
		{X: 5, Y: 0, Z: 25},
		{X: 35, Y: 0, Z: 15},
		{X: 15, Y: 0, Z: 35},
	}

	for i, pos := range resourcePositions {
		resourceNode := &engine.ResourceNode{
			ID:       world.GetNextEntityID(),
			Type:     "wood", // All wood for simplicity
			Amount:   1000,
			MaxAmount: 1000,
			Position: pos,
		}

		// Add to world resources (this is normally done during map loading)
		world.GetResourcesMutable()[resourceNode.ID] = resourceNode
		log.Printf("Created resource node at (%.1f, %.1f, %.1f)", pos.X, pos.Y, pos.Z)
	}
}