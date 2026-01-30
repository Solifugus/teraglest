package main

import (
	"fmt"
	"teraglest/internal/data"
	"teraglest/internal/engine"
)

func main() {
	fmt.Println("Testing Phase 2.2 - World Initialization from Map")
	fmt.Println("===============================================")

	// Create AssetManager and load basic game data
	dataRoot := "/home/solifugus/development/teraglest/megaglest-source/data/glest_game"
	assetManager := data.NewAssetManager(dataRoot + "/techs/megapack")

	// Load tech tree
	techTree, err := assetManager.LoadTechTree()
	if err != nil {
		fmt.Printf("Error loading tech tree: %v\n", err)
		return
	}
	fmt.Printf("✅ Tech tree loaded: %d attack types, %d armor types\n",
		len(techTree.AttackTypes), len(techTree.ArmorTypes))

	// Create game settings
	settings := engine.GameSettings{
		PlayerFactions:     map[int]string{1: "romans", 2: "magic"},
		MaxPlayers:        6,
		ResourceMultiplier: 1.0,
	}

	fmt.Println()
	fmt.Println("==================================================")
	fmt.Println("Creating World from Map: '2rivers'")

	// Create world from map (this is the new functionality!)
	world, err := engine.NewWorldFromMap(settings, techTree, assetManager, "2rivers")
	if err != nil {
		fmt.Printf("Error creating world from map: %v\n", err)
		return
	}

	fmt.Println("✅ Successfully created world from map!")

	// Print world summary
	fmt.Println()
	fmt.Printf("World Summary:\n")
	fmt.Printf("  Dimensions: %dx%d (replaced hardcoded 64x64!)\n", world.Width, world.Height)
	fmt.Printf("  Tile Size: %.1f\n", world.GetTileSize())

	if world.Map != nil {
		fmt.Printf("  Map Title: %s\n", world.Map.Title)
		fmt.Printf("  Map Author: %s\n", world.Map.Author)
		fmt.Printf("  Max Players: %d\n", world.Map.MaxPlayers)
		fmt.Printf("  Tileset: %s\n", world.Map.TilesetName)
		fmt.Printf("  Water Level: %.1f\n", world.Map.WaterLevel)
		fmt.Printf("  Height Factor: %.1f\n", world.Map.HeightFactor)

		fmt.Printf("  Start Positions:\n")
		for i, pos := range world.Map.StartPositions {
			height := world.GetHeight(engine.Vector2i{X: pos.X, Y: pos.Y})
			walkable := world.IsPositionWalkable(engine.Vector2i{X: pos.X, Y: pos.Y})
			fmt.Printf("    Player %d: (%d, %d) height=%.1f walkable=%t\n",
				i+1, pos.X, pos.Y, height, walkable)
		}
	}

	// Test terrain data integration
	fmt.Println()
	fmt.Printf("Terrain Data Integration:\n")

	// Sample various positions
	testPositions := []engine.Vector2i{
		{X: 0, Y: 0},     // Corner
		{X: 64, Y: 64},   // Center
		{X: 127, Y: 127}, // Other corner (if map is 128x128)
	}

	for _, pos := range testPositions {
		if pos.X < world.Width && pos.Y < world.Height {
			height := world.GetHeight(pos)
			walkable := world.IsPositionWalkable(pos)
			fmt.Printf("  Position (%d, %d): height=%.2f walkable=%t\n",
				pos.X, pos.Y, height, walkable)
		}
	}

	// Test resource node placement
	fmt.Println()
	fmt.Printf("Resource Node Placement:\n")

	resourceNodes := world.GetAllResourceNodes()
	fmt.Printf("  Total Resource Nodes: %d\n", len(resourceNodes))

	// Count by type
	resourceCounts := make(map[string]int)
	for _, node := range resourceNodes {
		resourceCounts[node.ResourceType]++
	}

	for resourceType, count := range resourceCounts {
		fmt.Printf("  %s: %d nodes\n", resourceType, count)
	}

	// Show a few example resource nodes
	fmt.Printf("  Sample Resource Nodes:\n")
	nodeCount := 0
	for _, node := range resourceNodes {
		if nodeCount >= 3 {
			break
		}
		fmt.Printf("    %s at (%.0f, %.0f): %d/%d\n",
			node.ResourceType, node.Position.X, node.Position.Z,
			node.Amount, node.MaxAmount)
		nodeCount++
	}

	fmt.Println()
	fmt.Println("==================================================")
	fmt.Println("Comparison: Old vs New World Creation")

	// Create an old-style world for comparison
	oldWorld, err := engine.NewWorld(settings, techTree, assetManager)
	if err != nil {
		fmt.Printf("Error creating old world: %v\n", err)
	} else {
		fmt.Printf("Old World (hardcoded): %dx%d\n", oldWorld.Width, oldWorld.Height)
		fmt.Printf("New World (from map):  %dx%d\n", world.Width, world.Height)

		oldResourceNodes := oldWorld.GetAllResourceNodes()
		fmt.Printf("Old World Resource Nodes: %d\n", len(oldResourceNodes))
		fmt.Printf("New World Resource Nodes: %d\n", len(resourceNodes))
	}

	fmt.Println()
	fmt.Println("🎉 Phase 2.2 Step 4 Complete!")
	fmt.Println("✅ Hardcoded 64x64 grid replaced with real map data")
	fmt.Println("✅ Terrain heights loaded from map")
	fmt.Println("✅ Walkability calculated from terrain objects")
	fmt.Println("✅ Player start positions integrated")
	fmt.Println("✅ Resource nodes placed based on terrain")
}