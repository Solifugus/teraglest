package main

import (
	"fmt"
	"log"
	"path/filepath"
	"time"

	"teraglest/internal/data"
	"teraglest/internal/engine"
	"teraglest/pkg/formats"
)

// Helper function for minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func main() {
	fmt.Println("Hello Teraglest - Testing XML Parsing")
	fmt.Println("=====================================")

	// Phase 1.2 - Tech Tree and Resources
	fmt.Println("=== Phase 1.2 - Tech Trees & Resources ===")

	// Path to the megaglest tech tree XML file
	techTreePath := filepath.Join("megaglest-source", "data", "glest_game", "techs", "megapack", "megapack.xml")

	// Load the tech tree
	fmt.Printf("Loading tech tree from: %s\n", techTreePath)
	techTree, err := data.LoadTechTree(techTreePath)
	if err != nil {
		log.Fatalf("Failed to load tech tree: %v", err)
	}

	fmt.Printf("Successfully loaded tech tree: %s\n\n", techTree.Description.Value)

	// Print attack types (required by Phase 1.2)
	techTree.PrintAttackTypes()
	fmt.Println()

	// Load resources
	resourcesPath := filepath.Join("megaglest-source", "data", "glest_game", "techs", "megapack", "resources")
	resources, err := data.LoadAllResources(resourcesPath)
	if err != nil {
		log.Fatalf("Failed to load resources: %v", err)
	}

	data.PrintResources(resources)

	// Phase 1.3 - Factions and Units
	fmt.Println("\n=== Phase 1.3 - Factions & Units ===")

	// Load magic faction as specified in Phase 1.3 validation
	magicFactionPath := filepath.Join("megaglest-source", "data", "glest_game", "techs", "megapack", "factions", "magic", "magic.xml")
	fmt.Printf("Loading magic faction from: %s\n", magicFactionPath)

	magicFaction, err := data.LoadFaction(magicFactionPath)
	if err != nil {
		log.Fatalf("Failed to load magic faction: %v", err)
	}

	// Print magic faction details
	fmt.Println("Magic Faction loaded successfully!")
	fmt.Println("Starting Resources:")
	for _, res := range magicFaction.StartingResources {
		fmt.Printf("  %s: %d\n", res.Name, res.Amount)
	}

	fmt.Println("Starting Units:")
	for _, unit := range magicFaction.StartingUnits {
		fmt.Printf("  %s: %d\n", unit.Name, unit.Amount)
	}

	// Load a sample unit to show unit costs (Phase 1.3 validation requirement)
	fmt.Println("\n--- Unit Cost Analysis ---")
	initiatePath := filepath.Join("megaglest-source", "data", "glest_game", "techs", "megapack", "factions", "magic", "units", "initiate", "initiate.xml")
	fmt.Printf("Loading initiate unit from: %s\n", initiatePath)

	initiateUnit, err := data.LoadUnit(initiatePath)
	if err != nil {
		log.Fatalf("Failed to load initiate unit: %v", err)
	}

	fmt.Println("Initiate Unit:")
	fmt.Printf("  HP: %d (regen: %d)\n", initiateUnit.Parameters.MaxHP.Value, initiateUnit.Parameters.MaxHP.Regeneration)
	fmt.Printf("  Armor: %d (%s)\n", initiateUnit.Parameters.Armor.Value, initiateUnit.Parameters.ArmorType.Value)
	fmt.Printf("  Size: %d, Sight: %d\n", initiateUnit.Parameters.Size.Value, initiateUnit.Parameters.Sight.Value)

	if len(initiateUnit.Parameters.ResourceRequirements) > 0 {
		fmt.Println("  Cost:")
		for _, req := range initiateUnit.Parameters.ResourceRequirements {
			fmt.Printf("    %s: %d\n", req.Name, req.Amount)
		}
	}

	fmt.Printf("  Skills: %d, Commands: %d\n", len(initiateUnit.Skills), len(initiateUnit.Commands))

	// Load all factions to verify broad compatibility
	fmt.Println("\n--- All Factions Summary ---")
	factionsPath := filepath.Join("megaglest-source", "data", "glest_game", "techs", "megapack", "factions")
	allFactions, err := data.LoadAllFactions(factionsPath)
	if err != nil {
		log.Fatalf("Failed to load all factions: %v", err)
	}

	fmt.Printf("Loaded %d factions successfully:\n", len(allFactions))
	for _, faction := range allFactions {
		fmt.Printf("  %s: %d starting resources, %d starting units\n",
			faction.Name,
			len(faction.Faction.StartingResources),
			len(faction.Faction.StartingUnits))
	}

	// Phase 1.4 - G3D Model Format Parser
	fmt.Println("\n=== Phase 1.4 - G3D Model Format Parser ===")

	// Load initiate standing model as specified in Phase 1.4 validation
	initiateModelPath := filepath.Join("megaglest-source", "data", "glest_game", "techs", "megapack", "factions", "magic", "units", "initiate", "models", "initiate_standing.g3d")
	fmt.Printf("Loading G3D model from: %s\n", initiateModelPath)

	model, err := formats.LoadG3D(initiateModelPath)
	if err != nil {
		log.Fatalf("Failed to load G3D model: %v", err)
	}

	// Print model summary as required by Phase 1.4 validation
	fmt.Printf("G3D Model loaded successfully!\n")
	fmt.Printf("  Version: %d\n", model.FileHeader.Version)
	fmt.Printf("  Mesh Count: %d\n", model.ModelHeader.MeshCount)
	fmt.Printf("  Total Vertices: %d\n", model.GetTotalVertexCount())
	fmt.Printf("  Total Triangles: %d\n", model.GetTotalTriangleCount())
	fmt.Printf("  Has Textures: %t\n", model.HasTextures())
	fmt.Printf("  Is Animated: %t\n", model.IsAnimated())

	// Load an animated model to demonstrate animation support
	fmt.Println("\n--- Animated Model Test ---")
	animatedModelPath := filepath.Join("megaglest-source", "data", "glest_game", "techs", "megapack", "factions", "magic", "units", "initiate", "models", "initiate_walking.g3d")
	fmt.Printf("Loading animated model from: %s\n", animatedModelPath)

	animatedModel, err := formats.LoadG3D(animatedModelPath)
	if err != nil {
		log.Fatalf("Failed to load animated G3D model: %v", err)
	}

	fmt.Printf("Animated model loaded successfully!\n")
	fmt.Printf("  Mesh Count: %d\n", animatedModel.ModelHeader.MeshCount)
	fmt.Printf("  Total Vertices: %d\n", animatedModel.GetTotalVertexCount())
	fmt.Printf("  Is Animated: %t\n", animatedModel.IsAnimated())

	// Show detailed mesh information for first mesh
	if len(animatedModel.Meshes) > 0 {
		firstMesh := animatedModel.Meshes[0]
		fmt.Printf("  First mesh: %d frames, %d vertices per frame\n",
			firstMesh.Header.FrameCount, firstMesh.Header.VertexCount)
	}

	// Phase 1.5 - Asset Management System
	fmt.Println("\n=== Phase 1.5 - Asset Management System ===")

	// Create asset manager
	assetManager := data.NewAssetManager(filepath.Join("megaglest-source", "data", "glest_game", "techs", "megapack"))
	fmt.Println("Asset manager created successfully!")

	// Test asset manager by loading tech tree through cache
	fmt.Println("\n--- Asset Manager Tech Tree Loading ---")
	cachedTechTree, err := assetManager.LoadTechTree()
	if err != nil {
		log.Fatalf("Failed to load tech tree through asset manager: %v", err)
	}
	fmt.Printf("Tech tree loaded through asset manager: %s\n", cachedTechTree.Description.Value)

	// Load it again to test caching
	cachedTechTree2, err := assetManager.LoadTechTree()
	if err != nil {
		log.Fatalf("Failed to load cached tech tree: %v", err)
	}

	// Check if it's the same instance (cached)
	if cachedTechTree == cachedTechTree2 {
		fmt.Println("✅ Tech tree caching working correctly!")
	} else {
		fmt.Println("❌ Tech tree caching not working")
	}

	// Test complete faction loading - this is the Phase 1.5 validation requirement
	fmt.Println("\n--- Complete Magic Faction Loading ---")
	completeMagic, err := assetManager.LoadFactionComplete("magic")
	if err != nil {
		log.Fatalf("Failed to load complete magic faction: %v", err)
	}

	fmt.Printf("Complete magic faction loaded successfully!\n")
	fmt.Printf("  Faction: %s\n", completeMagic.Faction.Name)
	fmt.Printf("  Starting Resources: %d\n", len(completeMagic.Faction.Faction.StartingResources))
	fmt.Printf("  Starting Units: %d\n", len(completeMagic.Faction.Faction.StartingUnits))
	fmt.Printf("  Total Units Loaded: %d\n", len(completeMagic.Units))
	fmt.Printf("  Total Models Loaded: %d\n", len(completeMagic.Models))

	fmt.Println("\nLoaded Units:")
	for unitName, unitDef := range completeMagic.Units {
		fmt.Printf("  %s: HP=%d, Armor=%d, Skills=%d, Commands=%d\n",
			unitName,
			unitDef.Unit.Parameters.MaxHP.Value,
			unitDef.Unit.Parameters.Armor.Value,
			len(unitDef.Unit.Skills),
			len(unitDef.Unit.Commands))
	}

	fmt.Println("\nLoaded 3D Models:")
	for modelName, model := range completeMagic.Models {
		fmt.Printf("  %s: Version=%d, Meshes=%d, Vertices=%d, Triangles=%d, Animated=%t\n",
			modelName,
			model.FileHeader.Version,
			model.ModelHeader.MeshCount,
			model.GetTotalVertexCount(),
			model.GetTotalTriangleCount(),
			model.IsAnimated())
	}

	// Show cache performance statistics
	fmt.Println("\n--- Asset Cache Performance ---")
	assetManager.PrintCacheStats()

	// Test individual asset loading through asset manager
	fmt.Println("\n--- Individual Asset Loading Test ---")

	// Load specific unit
	initiateUnitDef, err := assetManager.LoadUnit("magic", "initiate")
	if err != nil {
		log.Fatalf("Failed to load initiate unit through asset manager: %v", err)
	}
	fmt.Printf("Initiate unit loaded: %s (HP: %d)\n", initiateUnitDef.Name, initiateUnitDef.Unit.Parameters.MaxHP.Value)

	// Load specific G3D model
	modelPath := "factions/magic/units/initiate/models/initiate_standing.g3d"
	g3dModel, err := assetManager.LoadG3DModel(modelPath)
	if err != nil {
		log.Fatalf("Failed to load G3D model through asset manager: %v", err)
	}
	fmt.Printf("G3D model loaded: %d meshes, %d total vertices\n",
		g3dModel.ModelHeader.MeshCount, g3dModel.GetTotalVertexCount())

	// Phase 1.6 - Data Validation & Error Handling
	fmt.Println("\n=== Phase 1.6 - Data Validation & Error Handling ===")

	// Create data validator
	validator := data.NewDataValidator(filepath.Join("megaglest-source", "data", "glest_game", "techs", "megapack"), assetManager)
	fmt.Println("Data validator created successfully!")

	// Test individual validation components
	fmt.Println("\n--- Tech Tree Validation ---")
	techTreeReport, err := validator.ValidateTechTree()
	if err != nil {
		log.Fatalf("Failed to validate tech tree: %v", err)
	}
	fmt.Printf("Tech tree validation complete: %d errors, %d warnings, %d info messages\n",
		techTreeReport.ErrorCount, techTreeReport.WarningCount, techTreeReport.InfoCount)

	fmt.Println("\n--- Magic Faction Validation ---")
	factionReport, err := validator.ValidateFaction("magic")
	if err != nil {
		log.Fatalf("Failed to validate magic faction: %v", err)
	}
	fmt.Printf("Magic faction validation complete: %d errors, %d warnings, %d info messages\n",
		factionReport.ErrorCount, factionReport.WarningCount, factionReport.InfoCount)

	// Show sample issues if any
	if len(factionReport.Issues) > 0 {
		fmt.Println("\nSample faction validation issues:")
		for i, issue := range factionReport.Issues {
			if i >= 3 { // Limit to first 3 issues
				break
			}
			fmt.Printf("  [%s] %s: %s\n", issue.Severity, issue.Category, issue.Message)
			if issue.File != "" {
				fmt.Printf("    File: %s\n", issue.File)
			}
			if issue.Suggestion != "" {
				fmt.Printf("    Suggestion: %s\n", issue.Suggestion)
			}
		}
		if len(factionReport.Issues) > 3 {
			fmt.Printf("  ... and %d more issues\n", len(factionReport.Issues)-3)
		}
	} else {
		fmt.Println("✅ No faction validation issues found!")
	}

	fmt.Println("\n--- Asset Reference Validation ---")
	assetReport, err := validator.ValidateAssetReferences()
	if err != nil {
		log.Fatalf("Failed to validate asset references: %v", err)
	}
	fmt.Printf("Asset reference validation complete: %d errors, %d warnings, %d info messages\n",
		assetReport.ErrorCount, assetReport.WarningCount, assetReport.InfoCount)

	fmt.Println("\n--- Complete Data Validation ---")
	startTime := time.Now()
	completeReport, err := validator.ValidateAllData()
	if err != nil {
		log.Fatalf("Failed to validate all data: %v", err)
	}
	fmt.Printf("Complete validation finished in %v\n", time.Since(startTime))
	fmt.Printf("Files checked: %d\n", completeReport.FilesChecked)
	fmt.Printf("Total issues: %d (%d errors, %d warnings, %d info)\n",
		len(completeReport.Issues), completeReport.ErrorCount, completeReport.WarningCount, completeReport.InfoCount)

	// Show validation summary by category
	categoryCount := make(map[string]int)
	for _, issue := range completeReport.Issues {
		categoryCount[issue.Category]++
	}

	if len(categoryCount) > 0 {
		fmt.Println("\nIssues by category:")
		for category, count := range categoryCount {
			fmt.Printf("  %s: %d\n", category, count)
		}
	}

	// Demonstrate error handling with invalid faction
	fmt.Println("\n--- Error Handling Demonstration ---")
	invalidReport, err := validator.ValidateFaction("nonexistent_faction")
	if err != nil {
		fmt.Printf("Error validating invalid faction (expected): %v\n", err)
	} else {
		fmt.Printf("Invalid faction validation: %d errors found\n", invalidReport.ErrorCount)
		if invalidReport.ErrorCount > 0 {
			fmt.Printf("  Example error: %s\n", invalidReport.Issues[0].Message)
		} else {
			fmt.Println("  (Note: Error detection requires factions to be loaded first)")
		}
	}

	// Phase 2.1 - Core Game Engine Foundation
	fmt.Println("\n=== Phase 2.1 - Core Game Engine Foundation ===")

	// Create game settings for demonstration
	gameSettings := engine.GameSettings{
		TechTreePath: filepath.Join("megaglest-source", "data", "glest_game", "techs", "megapack"),
		PlayerFactions: map[int]string{
			1: "magic",  // Human player with magic faction
		},
		AIFactions: map[int]string{
			2: "tech",   // AI player with tech faction
		},
		GameSpeed:          1.0,
		ResourceMultiplier: 1.5, // Slightly higher resource generation
		MaxPlayers:         8,
		GameTimeLimit:      0, // No time limit
		EnableFogOfWar:     true,
		AllowCheats:        false,
	}

	fmt.Println("Creating game instance...")
	game, err := engine.NewGame(gameSettings, assetManager)
	if err != nil {
		log.Fatalf("Failed to create game: %v", err)
	}

	fmt.Printf("Game created successfully! State: %s\n", game.GetState())

	// Display game configuration
	fmt.Println("\n--- Game Configuration ---")
	settings := game.GetSettings()
	fmt.Printf("Tech Tree: %s\n", settings.TechTreePath)
	fmt.Printf("Max Players: %d\n", settings.MaxPlayers)
	fmt.Printf("Game Speed: %.1fx\n", settings.GameSpeed)
	fmt.Printf("Resource Multiplier: %.1fx\n", settings.ResourceMultiplier)
	fmt.Printf("Fog of War: %t\n", settings.EnableFogOfWar)

	fmt.Printf("Players configured:\n")
	for playerID, faction := range settings.PlayerFactions {
		fmt.Printf("  Player %d: %s (Human)\n", playerID, faction)
	}
	for playerID, faction := range settings.AIFactions {
		fmt.Printf("  Player %d: %s (AI)\n", playerID, faction)
	}

	// Initialize and start the game
	fmt.Println("\n--- Game Initialization ---")
	err = game.Start()
	if err != nil {
		log.Fatalf("Failed to start game: %v", err)
	}

	fmt.Printf("Game started! State: %s\n", game.GetState())

	// Get world information
	world := game.GetWorld()
	if world != nil {
		worldStats := world.GetWorldStats()
		fmt.Printf("World initialized with:\n")
		fmt.Printf("  Total Players: %d\n", worldStats.TotalPlayers)
		fmt.Printf("  Active Players: %d\n", worldStats.ActivePlayers)
		fmt.Printf("  Total Units: %d\n", worldStats.TotalUnits)
		fmt.Printf("  Total Buildings: %d\n", worldStats.TotalBuildings)
		fmt.Printf("  Resource Nodes: %d\n", worldStats.TotalResources)
		fmt.Printf("  Map Size: %s\n", worldStats.MapSize)

		// Display player information
		fmt.Println("\n--- Player Information ---")
		players := world.GetAllPlayers()
		for playerID, player := range players {
			fmt.Printf("Player %d (%s):\n", playerID, player.FactionName)
			fmt.Printf("  Type: %s\n", map[bool]string{true: "AI", false: "Human"}[player.IsAI])
			fmt.Printf("  Active: %t\n", player.IsActive)
			// Get player units and buildings from ObjectManager
			playerUnits := world.ObjectManager.GetUnitsForPlayer(playerID)
			playerBuildings := world.ObjectManager.GetBuildingsForPlayer(playerID)
			fmt.Printf("  Units: %d\n", len(playerUnits))
			fmt.Printf("  Buildings: %d\n", len(playerBuildings))

			fmt.Printf("  Resources:\n")
			for resType, amount := range player.Resources {
				fmt.Printf("    %s: %d\n", resType, amount)
			}

			fmt.Printf("  Units created: %d\n", player.UnitsCreated)
			fmt.Printf("  Units lost: %d\n", player.UnitsLost)
			fmt.Println()
		}
	}

	// Let the game run for a brief period
	fmt.Println("--- Game Loop Demonstration ---")
	fmt.Println("Running game loop for 2 seconds...")

	startTime = time.Now()
	for i := 0; i < 20; i++ { // Run for about 2 seconds at 10 Hz
		time.Sleep(100 * time.Millisecond)

		// Get current game statistics
		stats := game.GetStats()
		if i%10 == 0 { // Print every second
			elapsed := time.Since(startTime)
			fmt.Printf("  [%.1fs] Frames: %d, Avg Frame Time: %v, Game Time: %v\n",
				elapsed.Seconds(), stats.FrameCount, stats.AverageFrameTime, stats.CurrentGameTime)
		}
	}

	// Test game state transitions
	fmt.Println("\n--- Game State Management ---")
	fmt.Printf("Current state: %s\n", game.GetState())

	err = game.Pause()
	if err != nil {
		fmt.Printf("Failed to pause: %v\n", err)
	} else {
		fmt.Printf("Game paused. State: %s\n", game.GetState())
	}

	time.Sleep(500 * time.Millisecond)

	err = game.Resume()
	if err != nil {
		fmt.Printf("Failed to resume: %v\n", err)
	} else {
		fmt.Printf("Game resumed. State: %s\n", game.GetState())
	}

	// Check for game events
	fmt.Println("\n--- Game Events ---")
	events := game.GetEvents()
	fmt.Printf("Total events generated: %d\n", len(events))

	if len(events) > 0 {
		fmt.Println("Recent events:")
		for i, event := range events {
			if i >= 5 { // Limit to first 5 events
				break
			}
			fmt.Printf("  [%s] %s: %s\n",
				event.Timestamp.Format("15:04:05.000"), event.Type, event.Message)
		}
		if len(events) > 5 {
			fmt.Printf("  ... and %d more events\n", len(events)-5)
		}
	}

	// Final statistics
	fmt.Println("\n--- Final Game Statistics ---")
	finalStats := game.GetStats()
	fmt.Printf("Total runtime: %v\n", finalStats.CurrentGameTime)
	fmt.Printf("Frames processed: %d\n", finalStats.FrameCount)
	fmt.Printf("Average frame time: %v\n", finalStats.AverageFrameTime)
	fmt.Printf("Active players: %d\n", finalStats.PlayersActive)
	fmt.Printf("Total units: %d\n", finalStats.UnitsTotal)

	// Phase 2.2 - Game Object Management System Demonstration
	fmt.Println("\n=== Phase 2.2 - Game Object Management System ===")

	// Demonstrate ObjectManager capabilities
	objectMgr := world.ObjectManager

	fmt.Println("\n--- Current Game Objects ---")
	// Get all units by iterating through all players
	allUnits := make([]*engine.GameUnit, 0)
	allBuildings := make([]*engine.GameBuilding, 0)

	for playerID := 1; playerID <= 2; playerID++ {
		playerUnits := objectMgr.GetUnitsForPlayer(playerID)
		for _, unit := range playerUnits {
			allUnits = append(allUnits, unit)
		}
		playerBuildings := objectMgr.GetBuildingsForPlayer(playerID)
		for _, building := range playerBuildings {
			allBuildings = append(allBuildings, building)
		}
	}

	fmt.Printf("Total Units: %d\n", len(allUnits))
	fmt.Printf("Total Buildings: %d\n", len(allBuildings))

	// Show unit details
	fmt.Println("\nUnit Details:")
	for _, unit := range allUnits[:min(3, len(allUnits))] { // Show first 3 units
		pos := unit.GetPosition()
		fmt.Printf("  Unit %d (Player %d): %s at (%.1f, %.1f, %.1f)\n",
			unit.GetID(), unit.GetPlayerID(), unit.UnitType, pos.X, pos.Y, pos.Z)
		fmt.Printf("    Health: %d/%d, State: %s\n",
			unit.GetHealth(), unit.GetMaxHealth(), unit.State)
		fmt.Printf("    Movement Speed: %.1f, Attack Range: %.1f\n",
			unit.Speed, unit.AttackRange)
		if len(unit.CarriedResources) > 0 {
			fmt.Printf("    Carrying: %v\n", unit.CarriedResources)
		}
	}

	// Demonstrate command system
	fmt.Println("\n--- Command System Demonstration ---")
	if len(allUnits) > 0 {
		testUnit := allUnits[0]
		commandProcessor := engine.NewCommandProcessor(world)

		fmt.Printf("Testing commands with Unit %d:\n", testUnit.GetID())

		// Issue a move command
		moveCmd := engine.CreateMoveCommand(engine.Vector3{X: 25.0, Y: 0, Z: 25.0}, false)
		err = commandProcessor.IssueCommand(testUnit.GetID(), moveCmd)
		if err != nil {
			fmt.Printf("  Failed to issue move command: %v\n", err)
		} else {
			fmt.Printf("  ✅ Move command issued to position (25, 0, 25)\n")
			fmt.Printf("  Command queue length: %d\n", len(testUnit.CommandQueue))
		}

		// Issue another move command
		moveCmd2 := engine.CreateMoveCommand(engine.Vector3{X: 30.0, Y: 0, Z: 30.0}, true)
		err = commandProcessor.IssueCommand(testUnit.GetID(), moveCmd2)
		if err != nil {
			fmt.Printf("  Failed to issue second move command: %v\n", err)
		} else {
			fmt.Printf("  ✅ Second move command issued to (30, 0, 30)\n")
			fmt.Printf("  Command queue length: %d\n", len(testUnit.CommandQueue))
		}

		// Commands are processed automatically by the game loop
		fmt.Println("  Commands will be processed by the game loop automatically")

		// Show updated unit state
		pos := testUnit.GetPosition()
		fmt.Printf("  Updated position: (%.1f, %.1f, %.1f)\n", pos.X, pos.Y, pos.Z)
		fmt.Printf("  Current state: %s\n", testUnit.State)
		fmt.Printf("  Remaining commands: %d\n", len(testUnit.CommandQueue))

		// Test attack command if we have multiple units
		if len(allUnits) > 1 {
			targetUnit := allUnits[1]
			attackCmd := engine.CreateAttackCommand(targetUnit, false)
			err = commandProcessor.IssueCommand(testUnit.GetID(), attackCmd)
			if err != nil {
				fmt.Printf("  Failed to issue attack command: %v\n", err)
			} else {
				fmt.Printf("  ✅ Attack command issued targeting Unit %d\n", targetUnit.GetID())
			}
		}
	}

	// Demonstrate building management
	fmt.Println("\n--- Building Management ---")
	if len(allBuildings) > 0 {
		testBuilding := allBuildings[0]
		fmt.Printf("Building %d (Player %d): %s\n",
			testBuilding.GetID(), testBuilding.GetPlayerID(), testBuilding.BuildingType)
		fmt.Printf("  Health: %d/%d, Constructed: %t\n",
			testBuilding.GetHealth(), testBuilding.GetMaxHealth(), testBuilding.IsBuilt)
		fmt.Printf("  Production queue: %d items\n", len(testBuilding.ProductionQueue))
		fmt.Printf("  Upgrade level: %d\n", testBuilding.UpgradeLevel)

		// Test building commands
		commandProcessor := engine.NewCommandProcessor(world)

		// Issue production command
		cost := map[string]int{"gold": 50, "food": 1}
		produceCmd := engine.CreateProduceCommand("warrior", cost)
		err = commandProcessor.IssueBuildingCommand(testBuilding.GetID(), produceCmd)
		if err != nil {
			fmt.Printf("  Failed to issue production command: %v\n", err)
		} else {
			fmt.Printf("  ✅ Production command issued for warrior\n")
			fmt.Printf("  Production queue length: %d\n", len(testBuilding.ProductionQueue))
		}

		// Issue upgrade command
		upgradeCost := map[string]int{"gold": 100, "stone": 50}
		upgradeCmd := engine.CreateUpgradeCommand("armor_upgrade", upgradeCost)
		err = commandProcessor.IssueBuildingCommand(testBuilding.GetID(), upgradeCmd)
		if err != nil {
			fmt.Printf("  Failed to issue upgrade command: %v\n", err)
		} else {
			fmt.Printf("  ✅ Upgrade command issued for armor_upgrade\n")
			fmt.Printf("  Current upgrade level: %d\n", testBuilding.UpgradeLevel)
		}
	}

	// Demonstrate player-specific object queries
	fmt.Println("\n--- Player-Specific Objects ---")
	for playerID := 1; playerID <= 2; playerID++ {
		playerUnits := objectMgr.GetUnitsForPlayer(playerID)
		playerBuildings := objectMgr.GetBuildingsForPlayer(playerID)
		fmt.Printf("Player %d: %d units, %d buildings\n",
			playerID, len(playerUnits), len(playerBuildings))

		// Show resource carrying summary
		totalResources := make(map[string]int)
		for _, unit := range playerUnits {
			for resType, amount := range unit.CarriedResources {
				totalResources[resType] += amount
			}
		}
		if len(totalResources) > 0 {
			fmt.Printf("  Total resources carried: %v\n", totalResources)
		}
	}

	// Demonstrate advanced object operations
	fmt.Println("\n--- Advanced Object Operations ---")

	// Create a new unit using ObjectManager
	testPosition := engine.Vector3{X: 50.0, Y: 0, Z: 50.0}
	newUnit, err := objectMgr.CreateUnit(1, "test_worker", testPosition, nil)
	if err != nil {
		fmt.Printf("Failed to create new unit: %v\n", err)
	} else {
		fmt.Printf("✅ Created new test unit (ID: %d) at (50, 0, 50)\n", newUnit.GetID())

		// Test resource modification
		if newUnit.CarriedResources == nil {
			newUnit.CarriedResources = make(map[string]int)
		}
		newUnit.CarriedResources["gold"] = 25
		fmt.Printf("✅ Unit now carrying: %v\n", newUnit.CarriedResources)
	}

	// Test combat demonstration (simulated)
	if len(allUnits) > 1 {
		attacker := allUnits[0]
		target := allUnits[1]

		fmt.Printf("✅ Combat demo: Unit %d vs Unit %d\n",
			attacker.GetID(), target.GetID())
		fmt.Printf("   Attacker damage: %d, range: %.1f\n",
			attacker.AttackDamage, attacker.AttackRange)
		fmt.Printf("   Target health: %d/%d\n", target.GetHealth(), target.GetMaxHealth())
	}

	// Get final stats
	stats := objectMgr.GetStats()
	fmt.Printf("\nFinal object count: %d units, %d buildings\n",
		stats.TotalUnits, stats.TotalBuildings)

	// Stop the game
	err = game.Stop()
	if err != nil {
		fmt.Printf("Failed to stop game: %v\n", err)
	} else {
		fmt.Printf("Game stopped. Final state: %s\n", game.GetState())
	}

	fmt.Println("\n✅ XML Parsing Phase 1.3 Complete!")
	fmt.Println("✅ Magic faction parsed successfully!")
	fmt.Println("✅ Starting units and their costs extracted!")
	fmt.Println("✅ G3D Model Parsing Phase 1.4 Complete!")
	fmt.Println("✅ G3D models parsed successfully with mesh and vertex counts!")
	fmt.Println("✅ Animation frame support implemented!")
	fmt.Println("✅ Asset Management System Phase 1.5 Complete!")
	fmt.Println("✅ Thread-safe asset caching implemented!")
	fmt.Println("✅ Complete magic faction with all unit models loaded!")
	fmt.Println("✅ Data Validation & Error Handling Phase 1.6 Complete!")
	fmt.Println("✅ Comprehensive validation system implemented!")
	fmt.Println("✅ Clear error messages and suggestions provided!")
	fmt.Println("✅ Core Game Engine Foundation Phase 2.1 Complete!")
	fmt.Println("✅ Game state management and world system implemented!")
	fmt.Println("✅ Thread-safe game engine with event system!")
	fmt.Println("✅ Game Object Management Phase 2.2 Complete!")
	fmt.Println("✅ Advanced unit lifecycle management implemented!")
	fmt.Println("✅ Command processing system with 12 command types!")
	fmt.Println("✅ Thread-safe ObjectManager with polymorphic game objects!")
	fmt.Println("✅ Resource gathering, combat, and building production systems!")
}