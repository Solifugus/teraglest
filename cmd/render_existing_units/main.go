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
	fmt.Println("=== TeraGlest Render Existing Units Demo ===")
	fmt.Println("Rendering the 9 units created during world initialization")
	fmt.Println()

	// Initialize asset manager
	dataRoot := "/home/solifugus/development/teraglest/megaglest-source/data/glest_game"
	assetManager := data.NewAssetManager(filepath.Join(dataRoot, "techs", "megapack"))
	fmt.Printf("✅ AssetManager initialized with data root: %s\n", dataRoot)

	// Create renderer
	r, err := renderer.NewRenderer(assetManager, "TeraGlest Render Units Demo", 1024, 768)
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

	// Start the game (this creates the 9 units we saw in debug output)
	err = game.Start()
	if err != nil {
		log.Fatalf("Failed to start game: %v", err)
	}
	fmt.Println("✅ Game started - 9 units should be created automatically")

	// Wait for initialization to complete
	time.Sleep(300 * time.Millisecond)

	// Get world reference
	world := game.GetWorld()
	if world == nil {
		log.Fatalf("Failed to get game world")
	}

	fmt.Printf("✅ Game world available: %dx%d with %d players\n",
		world.Width, world.Height, len(world.GetAllPlayers()))

	// Check how many units were actually created
	players := world.GetAllPlayers()
	totalUnits := 0
	for _, player := range players {
		units := world.ObjectManager.GetUnitsForPlayer(player.ID)
		buildings := world.ObjectManager.GetBuildingsForPlayer(player.ID)
		playerUnits := len(units)
		playerBuildings := len(buildings)
		totalUnits += playerUnits + playerBuildings
		fmt.Printf("✅ Player %d: %d units, %d buildings\n", player.ID, playerUnits, playerBuildings)
	}
	fmt.Printf("✅ Total game objects to render: %d\n", totalUnits)

	// Position camera to look at the units (they're around grid position 10,0)
	// The renderer should have a default camera position that can see the world
	fmt.Println("✅ Using default camera position to view rendered units")

	// Set up input callbacks
	context := r.GetContext()
	window := context.GetWindow()

	window.SetKeyCallback(func(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
		if action == glfw.Press || action == glfw.Repeat {
			switch key {
			case glfw.KeySpace:
				fmt.Printf("📊 Frame %d: FPS=%.1f, Units=%d\n", r.GetFrameCount(), r.GetFPS(), totalUnits)
			case glfw.KeyR:
				fmt.Printf("🔄 Renderer stats: Frame %d, FPS=%.1f\n",
					r.GetFrameCount(), r.GetFPS())
			case glfw.KeyW:
				// Move camera forward (closer to models)
				camera := r.GetCamera()
				pos := camera.Position
				camera.SetPosition(pos[0], pos[1], pos[2]-1.0)
				fmt.Printf("📹 Camera moved to: %.1f, %.1f, %.1f\n", pos[0], pos[1], pos[2]-1.0)
			case glfw.KeyS:
				// Move camera back
				camera := r.GetCamera()
				pos := camera.Position
				camera.SetPosition(pos[0], pos[1], pos[2]+1.0)
				fmt.Printf("📹 Camera moved to: %.1f, %.1f, %.1f\n", pos[0], pos[1], pos[2]+1.0)
			case glfw.KeyA:
				// Move camera left
				camera := r.GetCamera()
				pos := camera.Position
				camera.SetPosition(pos[0]-1.0, pos[1], pos[2])
				fmt.Printf("📹 Camera moved to: %.1f, %.1f, %.1f\n", pos[0]-1.0, pos[1], pos[2])
			case glfw.KeyD:
				// Move camera right
				camera := r.GetCamera()
				pos := camera.Position
				camera.SetPosition(pos[0]+1.0, pos[1], pos[2])
				fmt.Printf("📹 Camera moved to: %.1f, %.1f, %.1f\n", pos[0]+1.0, pos[1], pos[2])
			case glfw.KeyQ:
				// Move camera up
				camera := r.GetCamera()
				pos := camera.Position
				camera.SetPosition(pos[0], pos[1]+1.0, pos[2])
				fmt.Printf("📹 Camera moved to: %.1f, %.1f, %.1f\n", pos[0], pos[1]+1.0, pos[2])
			case glfw.KeyE:
				// Move camera down
				camera := r.GetCamera()
				pos := camera.Position
				camera.SetPosition(pos[0], pos[1]-1.0, pos[2])
				fmt.Printf("📹 Camera moved to: %.1f, %.1f, %.1f\n", pos[0], pos[1]-1.0, pos[2])
			}
		}
	})

	fmt.Println()
	fmt.Println("=== CONTROLS ===")
	fmt.Println("ESC:    Exit demo")
	fmt.Println("SPACE:  Print stats")
	fmt.Println("R:      Print renderer status")
	fmt.Println()
	fmt.Println("=== CAMERA MOVEMENT ===")
	fmt.Println("W/S:    Move forward/back")
	fmt.Println("A/D:    Move left/right")
	fmt.Println("Q/E:    Move up/down")
	fmt.Println()

	fmt.Printf("🎮 Starting render loop to display %d game objects...\n", totalUnits)

	frameCount := 0
	for !r.ShouldClose() {
		// Render the entire game world (including all created units)
		err := r.RenderWorld(world)
		if err != nil {
			log.Printf("Render error: %v", err)
			break
		}

		frameCount++

		// Print confirmation every 2 seconds that we're actively rendering
		if frameCount%120 == 0 { // Every 2 seconds at 60 FPS
			fmt.Printf("🎨 Rendering frame %d with %d objects (FPS: %.1f)\n",
				frameCount, totalUnits, r.GetFPS())
		}
	}

	// Stop the game
	game.Stop()

	fmt.Printf("\n✅ Demo completed after %d frames\n", frameCount)
	fmt.Println("🎉 Successfully rendered existing units!")
}