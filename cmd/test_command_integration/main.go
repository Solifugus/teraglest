package main

import (
	"fmt"
	"path/filepath"
	"time"

	"teraglest/internal/data"
	"teraglest/internal/engine"
)

func main() {
	fmt.Println("=== Phase 2.5 Command System Integration Test ===")

	// Initialize minimal game setup
	techTreePath := filepath.Join("megaglest-source", "data", "glest_game", "techs", "megapack", "megapack.xml")
	_, err := data.LoadTechTree(techTreePath)
	if err != nil {
		fmt.Printf("Failed to load tech tree: %v\n", err)
		return
	}

	assetManager := data.NewAssetManager(filepath.Join("megaglest-source", "data", "glest_game", "techs", "megapack"))

	gameSettings := engine.GameSettings{
		TechTreePath:       techTreePath,
		MaxPlayers:         4,
		GameSpeed:          1.0,
		ResourceMultiplier: 1.0,
		PlayerFactions: map[int]string{
			1: "magic",
			2: "tech",
		},
	}

	// Create game
	game, err := engine.NewGame(gameSettings, assetManager)
	if err != nil {
		fmt.Printf("Failed to create game: %v\n", err)
		return
	}

	fmt.Println("✅ Game created successfully")

	// Start the game (this should start the background update loop)
	err = game.Start()
	if err != nil {
		fmt.Printf("Failed to start game: %v\n", err)
		return
	}

	fmt.Println("✅ Game started - background update loop running")

	// Wait a moment for the game to initialize
	time.Sleep(500 * time.Millisecond)

	// Get world and create a test unit
	world := game.GetWorld()
	players := world.GetAllPlayers()
	if len(players) == 0 {
		fmt.Println("❌ No players found")
		return
	}

	player := players[0]
	fmt.Printf("✅ Found player %d\n", player.ID)

	// Create a test unit
	unitPos := engine.Vector3{X: 10, Y: 0, Z: 10}
	unit, err := world.ObjectManager.CreateUnit(player.ID, "initiate", unitPos, nil)
	if err != nil {
		fmt.Printf("Failed to create unit: %v\n", err)
		return
	}

	fmt.Printf("✅ Created test unit %d at position (%.1f, %.1f, %.1f)\n",
		unit.GetID(), unitPos.X, unitPos.Y, unitPos.Z)

	// Issue a move command
	moveTarget := engine.Vector3{X: 20, Y: 0, Z: 20}
	moveCmd := engine.CreateMoveCommand(moveTarget, false)

	commandProcessor := engine.NewCommandProcessor(world)
	err = commandProcessor.IssueCommand(unit.GetID(), moveCmd)
	if err != nil {
		fmt.Printf("Failed to issue move command: %v\n", err)
		return
	}

	fmt.Printf("✅ Issued move command to position (%.1f, %.1f, %.1f)\n",
		moveTarget.X, moveTarget.Y, moveTarget.Z)
	fmt.Printf("   Unit has %d commands in queue\n", len(unit.CommandQueue))
	fmt.Printf("   Current command: %v\n", unit.CurrentCommand != nil)

	// Wait for command processing (the background game loop should process this)
	fmt.Println("\n🔄 Waiting 3 seconds for command processing...")

	initialPos := unit.Position
	for i := 0; i < 30; i++ {
		time.Sleep(100 * time.Millisecond)

		if i%10 == 0 { // Print every second
			fmt.Printf("   [%.1fs] Unit position: (%.2f, %.2f, %.2f), State: %s, Commands: %d\n",
				float64(i)*0.1, unit.Position.X, unit.Position.Y, unit.Position.Z,
				unit.State.String(), len(unit.CommandQueue))
		}
	}

	finalPos := unit.Position
	moved := (finalPos.X != initialPos.X) || (finalPos.Z != initialPos.Z)

	fmt.Printf("\n📊 Results:\n")
	fmt.Printf("   Initial position: (%.2f, %.2f, %.2f)\n", initialPos.X, initialPos.Y, initialPos.Z)
	fmt.Printf("   Final position:   (%.2f, %.2f, %.2f)\n", finalPos.X, finalPos.Y, finalPos.Z)
	fmt.Printf("   Unit moved: %t\n", moved)
	fmt.Printf("   Current state: %s\n", unit.State.String())
	fmt.Printf("   Remaining commands: %d\n", len(unit.CommandQueue))

	if moved {
		fmt.Println("✅ SUCCESS: Command system integration working - unit moved!")
	} else {
		fmt.Println("❌ ISSUE: Unit did not move - commands may not be processing")
	}

	// Stop the game
	game.Stop()
	fmt.Println("✅ Game stopped")
}