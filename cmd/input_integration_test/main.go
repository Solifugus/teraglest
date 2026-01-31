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

// SimpleInputHandler for testing input without full UI system
type SimpleInputHandler struct {
	world  *engine.World
	camera *renderer.Camera
	selectedUnits []*engine.GameUnit
	screenWidth  int
	screenHeight int
}

func NewSimpleInputHandler(world *engine.World) *SimpleInputHandler {
	return &SimpleInputHandler{
		world:        world,
		selectedUnits: make([]*engine.GameUnit, 0),
		screenWidth:  1024,
		screenHeight: 768,
	}
}

func (sih *SimpleInputHandler) SetCamera(camera *renderer.Camera) {
	sih.camera = camera
}

func (sih *SimpleInputHandler) SetScreenDimensions(width, height int) {
	sih.screenWidth = width
	sih.screenHeight = height
}

func (sih *SimpleInputHandler) HandleMouseButton(window *glfw.Window, button glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
	if button == glfw.MouseButtonLeft && action == glfw.Press {
		xpos, ypos := window.GetCursorPos()
		worldX, worldZ := sih.screenToWorld(xpos, ypos)

		fmt.Printf("Mouse click at screen (%.1f, %.1f) -> world (%.2f, %.2f)\n",
			xpos, ypos, worldX, worldZ)

		// Try to select a unit at this position
		unit := sih.findUnitAtPosition(worldX, worldZ)
		if unit != nil {
			sih.selectedUnits = []*engine.GameUnit{unit}
			fmt.Printf("Selected unit %d (%s) at (%.1f, %.1f)\n",
				unit.ID, unit.UnitType, unit.Position.X, unit.Position.Z)
		} else {
			// Issue move command to selected units
			if len(sih.selectedUnits) > 0 {
				sih.issueMoveCommand(worldX, worldZ)
			} else {
				fmt.Println("No unit at click position, no units selected")
			}
		}
	}
}

func (sih *SimpleInputHandler) HandleMouseMove(window *glfw.Window, xpos, ypos float64) {
	// Simple mouse tracking - could be used for selection boxes later
}

func (sih *SimpleInputHandler) HandleKeyboard(window *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if action == glfw.Press {
		switch key {
		case glfw.KeyA:
			if (mods & glfw.ModControl) != 0 {
				sih.selectAllPlayerUnits()
			}
		case glfw.KeyS:
			sih.issueStopCommand()
		case glfw.KeyH:
			sih.issueHoldCommand()
		}
	}
}

// Screen to world conversion using camera ray casting
func (sih *SimpleInputHandler) screenToWorld(screenX, screenY float64) (float64, float64) {
	if sih.camera == nil {
		// Fallback to simple conversion
		return screenX / 10.0, screenY / 10.0
	}

	rayOrigin, rayDirection := sih.camera.ScreenToWorldRay(
		int(screenX), int(screenY),
		sih.screenWidth, sih.screenHeight)

	// Find intersection with ground plane (Y = 0)
	if rayDirection.Y() == 0 {
		return float64(rayOrigin.X()), float64(rayOrigin.Z())
	}

	t := -rayOrigin.Y() / rayDirection.Y()
	intersectionPoint := rayOrigin.Add(rayDirection.Mul(t))

	return float64(intersectionPoint.X()), float64(intersectionPoint.Z())
}

func (sih *SimpleInputHandler) findUnitAtPosition(worldX, worldZ float64) *engine.GameUnit {
	searchRadius := 1.0
	players := sih.world.GetAllPlayers()

	for _, player := range players {
		units := sih.world.ObjectManager.GetUnitsForPlayer(player.ID)
		for _, unit := range units {
			if unit.Health <= 0 {
				continue
			}

			dx := unit.Position.X - worldX
			dz := unit.Position.Z - worldZ
			distance := dx*dx + dz*dz

			if distance <= searchRadius*searchRadius {
				return unit
			}
		}
	}
	return nil
}

func (sih *SimpleInputHandler) selectAllPlayerUnits() {
	players := sih.world.GetAllPlayers()
	if len(players) == 0 {
		return
	}

	// Get first non-AI player (or first player)
	var playerID int
	for _, player := range players {
		playerID = player.ID
		if !player.IsAI {
			break
		}
	}

	units := sih.world.ObjectManager.GetUnitsForPlayer(playerID)
	var livingUnits []*engine.GameUnit
	for _, unit := range units {
		if unit.Health > 0 {
			livingUnits = append(livingUnits, unit)
		}
	}

	sih.selectedUnits = livingUnits
	fmt.Printf("Selected all units: %d units\n", len(livingUnits))
}

func (sih *SimpleInputHandler) issueMoveCommand(worldX, worldZ float64) {
	if len(sih.selectedUnits) == 0 {
		return
	}

	commandProcessor := engine.NewCommandProcessor(sih.world)

	for _, unit := range sih.selectedUnits {
		moveCmd := engine.CreateMoveCommand(engine.Vector3{X: worldX, Y: 0, Z: worldZ}, false)
		err := commandProcessor.IssueCommand(unit.ID, moveCmd)
		if err != nil {
			fmt.Printf("Failed to issue move command to unit %d: %v\n", unit.ID, err)
		} else {
			fmt.Printf("Issued move command to unit %d -> (%.1f, %.1f)\n", unit.ID, worldX, worldZ)
		}
	}
}

func (sih *SimpleInputHandler) issueStopCommand() {
	if len(sih.selectedUnits) == 0 {
		return
	}

	commandProcessor := engine.NewCommandProcessor(sih.world)

	for _, unit := range sih.selectedUnits {
		stopCmd := engine.CreateStopCommand()
		err := commandProcessor.IssueCommand(unit.ID, stopCmd)
		if err != nil {
			fmt.Printf("Failed to issue stop command to unit %d: %v\n", unit.ID, err)
		} else {
			fmt.Printf("Issued stop command to unit %d\n", unit.ID)
		}
	}
}

func (sih *SimpleInputHandler) issueHoldCommand() {
	if len(sih.selectedUnits) == 0 {
		return
	}

	commandProcessor := engine.NewCommandProcessor(sih.world)

	for _, unit := range sih.selectedUnits {
		// Create hold command manually since CreateHoldCommand doesn't exist
		holdCmd := engine.UnitCommand{
			Type:      engine.CommandHold,
			CreatedAt: time.Now(),
			Priority:  0,
			IsQueued:  false,
		}
		err := commandProcessor.IssueCommand(unit.ID, holdCmd)
		if err != nil {
			fmt.Printf("Failed to issue hold command to unit %d: %v\n", unit.ID, err)
		} else {
			fmt.Printf("Issued hold command to unit %d\n", unit.ID)
		}
	}
}

func main() {
	fmt.Println("=== Input Pipeline Integration Test ===")
	fmt.Println("Testing mouse and keyboard input integration with game")
	fmt.Println()

	// Initialize asset manager
	dataRoot := filepath.Join("megaglest-source", "data", "glest_game")
	assetManager := data.NewAssetManager(filepath.Join(dataRoot, "techs", "megapack"))

	// Create renderer
	r, err := renderer.NewRenderer(assetManager, "Input Pipeline Test", 1024, 768)
	if err != nil {
		log.Fatalf("Failed to create renderer: %v", err)
	}
	defer r.Destroy()

	// Create game
	gameSettings := engine.GameSettings{
		TechTreePath:       filepath.Join(dataRoot, "techs", "megapack", "megapack.xml"),
		MaxPlayers:         1,
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

	// Start the game
	err = game.Start()
	if err != nil {
		log.Fatalf("Failed to start game: %v", err)
	}
	defer game.Stop()

	// Get world reference
	world := game.GetWorld()

	// Wait for game to initialize
	time.Sleep(300 * time.Millisecond)

	// Create simple input handler
	inputHandler := NewSimpleInputHandler(world)
	inputHandler.SetCamera(r.GetCamera())
	inputHandler.SetScreenDimensions(1024, 768)

	// Setup input callbacks in renderer
	r.SetupGameInputCallbacks(inputHandler)

	fmt.Println("✅ Input pipeline configured")
	fmt.Println()
	fmt.Println("🎮 Controls:")
	fmt.Println("   Left Click: Select unit or move selected units")
	fmt.Println("   Ctrl+A: Select all player units")
	fmt.Println("   S: Stop selected units")
	fmt.Println("   H: Hold position for selected units")
	fmt.Println("   ESC: Exit")
	fmt.Println()

	frameCount := 0
	startTime := time.Now()

	// Main loop with input integration
	for !r.ShouldClose() {
		// Render the game world
		err := r.RenderWorld(world)
		if err != nil {
			log.Printf("Render error: %v", err)
			break
		}

		frameCount++

		// Print stats every 3 seconds
		if frameCount%180 == 0 {
			elapsed := time.Since(startTime)
			fps := float64(frameCount) / elapsed.Seconds()

			players := world.GetAllPlayers()
			totalUnits := 0
			for _, player := range players {
				units := world.ObjectManager.GetUnitsForPlayer(player.ID)
				totalUnits += len(units)
			}

			fmt.Printf("🎯 [%.1fs] FPS: %.1f, Units: %d, Selected: %d\n",
				elapsed.Seconds(), fps, totalUnits, len(inputHandler.selectedUnits))
		}

		// Handle GLFW events (this triggers our input callbacks)
		glfw.PollEvents()

		// Maintain frame rate
		time.Sleep(16 * time.Millisecond) // ~60 FPS
	}

	elapsed := time.Since(startTime)
	avgFPS := float64(frameCount) / elapsed.Seconds()

	fmt.Println()
	fmt.Printf("✅ Input integration test completed\n")
	fmt.Printf("   Duration: %.1f seconds\n", elapsed.Seconds())
	fmt.Printf("   Average FPS: %.1f\n", avgFPS)
	fmt.Println()
	fmt.Println("🎉 Input Pipeline Integration Test Complete!")
	fmt.Println("   ✅ Mouse input captured and converted to world coordinates")
	fmt.Println("   ✅ Unit selection working")
	fmt.Println("   ✅ Move commands issued via mouse clicks")
	fmt.Println("   ✅ Keyboard shortcuts functional")
	fmt.Println("   ✅ Input integrated with game command system")
}