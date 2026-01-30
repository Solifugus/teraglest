package engine

import (
	"testing"
	"time"

	"teraglest/internal/data"
)

// createTestWorld creates a world for testing resource management
func createTestWorld() *World {
	// Create basic game settings
	settings := GameSettings{
		TechTreePath: "../../megaglest-source/data/glest_game/techs/megapack",
		PlayerFactions: map[int]string{
			1: "magic", // Use magic faction for testing
		},
		AIFactions:         map[int]string{},
		GameSpeed:          1.0,
		ResourceMultiplier: 1.0,
		MaxPlayers:         2,
	}

	// Create asset manager (mock for testing)
	assetMgr := data.NewAssetManager("../../megaglest-source/data/glest_game/techs/megapack")

	// Load tech tree
	techTree, err := assetMgr.LoadTechTree()
	if err != nil {
		// Use minimal tech tree for testing
		techTree = &data.TechTree{}
	}

	// Create world
	world, err := NewWorld(settings, techTree, assetMgr)
	if err != nil {
		panic("Failed to create test world: " + err.Error())
	}

	// Initialize world
	err = world.Initialize()
	if err != nil {
		panic("Failed to initialize test world: " + err.Error())
	}

	return world
}

// createTestPlayer creates a test player with basic resources
func createTestPlayer(world *World, playerID int) {
	err := world.createPlayer(playerID, "magic", false)
	if err != nil {
		panic("Failed to create test player: " + err.Error())
	}

	// Set initial resources for testing
	player := world.GetPlayer(playerID)
	if player != nil {
		player.Resources["gold"] = 1000
		player.Resources["wood"] = 500
		player.Resources["stone"] = 300
		player.Resources["energy"] = 100
		player.ResourcesGathered["gold"] = 0
		player.ResourcesGathered["wood"] = 0
		player.ResourcesGathered["stone"] = 0
		player.ResourcesGathered["energy"] = 0
		player.ResourcesSpent["gold"] = 0
		player.ResourcesSpent["wood"] = 0
		player.ResourcesSpent["stone"] = 0
		player.ResourcesSpent["energy"] = 0
	}
}

func TestResourceValidator(t *testing.T) {
	world := createTestWorld()
	createTestPlayer(world, 1)

	validator := NewResourceValidator(world)

	t.Run("ValidateResources_SufficientResources", func(t *testing.T) {
		result := validator.ValidateResources(ResourceCheck{
			PlayerID: 1,
			Required: map[string]int{"gold": 100, "wood": 50},
			Purpose:  "unit_creation",
		})

		if !result.Valid {
			t.Errorf("Expected validation to pass, got: %s", result.Error)
		}
	})

	t.Run("ValidateResources_InsufficientResources", func(t *testing.T) {
		result := validator.ValidateResources(ResourceCheck{
			PlayerID: 1,
			Required: map[string]int{"gold": 2000}, // More than available
			Purpose:  "building_construction",
		})

		if result.Valid {
			t.Error("Expected validation to fail for insufficient resources")
		}

		if len(result.Missing) == 0 {
			t.Error("Expected missing resources to be reported")
		}

		if result.Missing["gold"] != 1000 { // 2000 - 1000 = 1000 missing
			t.Errorf("Expected missing gold to be 1000, got %d", result.Missing["gold"])
		}
	})

	t.Run("CanAfford_True", func(t *testing.T) {
		canAfford := validator.CanAfford(1, map[string]int{"gold": 100})
		if !canAfford {
			t.Error("Expected player to be able to afford cost")
		}
	})

	t.Run("CanAfford_False", func(t *testing.T) {
		canAfford := validator.CanAfford(1, map[string]int{"gold": 2000})
		if canAfford {
			t.Error("Expected player to not be able to afford cost")
		}
	})

	t.Run("ValidateResources_PlayerNotFound", func(t *testing.T) {
		result := validator.ValidateResources(ResourceCheck{
			PlayerID: 999, // Non-existent player
			Required: map[string]int{"gold": 100},
			Purpose:  "test",
		})

		if result.Valid {
			t.Error("Expected validation to fail for non-existent player")
		}
	})
}

func TestResourceDeduction(t *testing.T) {
	world := createTestWorld()
	createTestPlayer(world, 1)

	t.Run("DeductResources_Success", func(t *testing.T) {
		initialGold := world.GetPlayer(1).Resources["gold"]

		err := world.DeductResources(1, map[string]int{"gold": 100}, "test_deduction")
		if err != nil {
			t.Errorf("Expected successful resource deduction, got error: %v", err)
		}

		player := world.GetPlayer(1)
		newGold := player.Resources["gold"]
		spentGold := player.ResourcesSpent["gold"]

		if newGold != initialGold-100 {
			t.Errorf("Expected gold to be %d, got %d", initialGold-100, newGold)
		}

		if spentGold != 100 {
			t.Errorf("Expected spent gold to be 100, got %d", spentGold)
		}
	})

	t.Run("DeductResources_InsufficientFunds", func(t *testing.T) {
		err := world.DeductResources(1, map[string]int{"gold": 10000}, "test_deduction")
		if err == nil {
			t.Error("Expected error for insufficient resources")
		}
	})

	t.Run("AddResources_Success", func(t *testing.T) {
		initialGold := world.GetPlayer(1).Resources["gold"]
		initialGathered := world.GetPlayer(1).ResourcesGathered["gold"]

		err := world.AddResources(1, map[string]int{"gold": 200}, "test_addition")
		if err != nil {
			t.Errorf("Expected successful resource addition, got error: %v", err)
		}

		player := world.GetPlayer(1)
		newGold := player.Resources["gold"]
		newGathered := player.ResourcesGathered["gold"]

		if newGold != initialGold+200 {
			t.Errorf("Expected gold to be %d, got %d", initialGold+200, newGold)
		}

		if newGathered != initialGathered+200 {
			t.Errorf("Expected gathered gold to be %d, got %d", initialGathered+200, newGathered)
		}
	})
}

func TestPopulationManager(t *testing.T) {
	world := createTestWorld()
	createTestPlayer(world, 1)

	popManager := NewPopulationManager(world)

	t.Run("GetPopulationStatus_EmptyWorld", func(t *testing.T) {
		status := popManager.GetPopulationStatus(1)

		if status.CurrentPopulation != 0 {
			t.Errorf("Expected current population to be 0, got %d", status.CurrentPopulation)
		}

		// Should have minimum capacity of 10
		if status.MaxPopulation < 10 {
			t.Errorf("Expected max population to be at least 10, got %d", status.MaxPopulation)
		}
	})

	t.Run("CanCreateUnit_WithinLimits", func(t *testing.T) {
		canCreate, reason := popManager.CanCreateUnit(1, "worker")
		if !canCreate {
			t.Errorf("Expected to be able to create worker unit, reason: %s", reason)
		}
	})

	t.Run("GetMaxUnitsCanCreate", func(t *testing.T) {
		maxUnits := popManager.GetMaxUnitsCanCreate(1, "worker")
		if maxUnits <= 0 {
			t.Errorf("Expected to be able to create at least 1 worker unit, got %d", maxUnits)
		}
	})

	t.Run("GetUnitPopulationCost_DefaultCost", func(t *testing.T) {
		cost := popManager.getUnitPopulationCost(1, "worker")
		if cost != 1 {
			t.Errorf("Expected worker population cost to be 1, got %d", cost)
		}
	})

	t.Run("ValidatePopulation_NoIssues", func(t *testing.T) {
		issues := popManager.ValidatePopulation(1)
		if len(issues) != 0 {
			t.Errorf("Expected no population issues, got: %v", issues)
		}
	})
}

func TestResourceGeneration(t *testing.T) {
	world := createTestWorld()
	createTestPlayer(world, 1)

	t.Run("CalculateResourceRates_NoBuildings", func(t *testing.T) {
		rates := world.calculateResourceRates(1)
		if len(rates) != 0 {
			t.Errorf("Expected no resource rates with no buildings, got: %v", rates)
		}
	})

	t.Run("CalculateResourceGeneration", func(t *testing.T) {
		// Create a mock building for testing
		building := &GameBuilding{
			ID:           1,
			PlayerID:     1,
			BuildingType: "test_building",
			IsBuilt:      true,
			Health:       100,
			UpgradeLevel: 1,
		}

		baseRate := float32(2.0)
		deltaTime := time.Second

		generated := world.calculateResourceGeneration(baseRate, deltaTime, building)
		expected := int(baseRate * 1.0 * world.settings.ResourceMultiplier) // Should be 2

		if generated != expected {
			t.Errorf("Expected %d resources generated, got %d", expected, generated)
		}
	})
}

func TestResourceStatus(t *testing.T) {
	world := createTestWorld()
	createTestPlayer(world, 1)

	t.Run("GetResourceStatus_Complete", func(t *testing.T) {
		status := world.GetResourceStatus(1)

		// Check that all resource maps are initialized
		if status.Resources == nil {
			t.Error("Expected resources map to be initialized")
		}

		if status.ResourceRates == nil {
			t.Error("Expected resource rates map to be initialized")
		}

		if status.ResourcesGathered == nil {
			t.Error("Expected gathered resources map to be initialized")
		}

		if status.ResourcesSpent == nil {
			t.Error("Expected spent resources map to be initialized")
		}

		// Check specific resource values
		if status.Resources["gold"] != 1000 {
			t.Errorf("Expected gold to be 1000, got %d", status.Resources["gold"])
		}
	})
}

func TestPlayerUpdateWithResources(t *testing.T) {
	world := createTestWorld()
	createTestPlayer(world, 1)

	t.Run("UpdatePlayer_ProcessResourceDropoffs", func(t *testing.T) {
		player := world.GetPlayer(1)
		initialGold := player.Resources["gold"]

		// Call updatePlayer to test resource processing
		deltaTime := time.Second
		world.updatePlayer(player, deltaTime)

		// Since we don't have buildings generating resources in test,
		// resources should remain the same
		if player.Resources["gold"] != initialGold {
			t.Errorf("Expected gold to remain %d, got %d", initialGold, player.Resources["gold"])
		}
	})
}

func TestResourceEventSystem(t *testing.T) {
	world := createTestWorld()
	createTestPlayer(world, 1)

	t.Run("LogResourceTransaction", func(t *testing.T) {
		// This test verifies that the logging function doesn't crash
		// In a full implementation, you would test actual event emission
		resources := map[string]int{"gold": 100}
		world.logResourceTransaction(1, resources, "test_source", "addition")

		// Test shouldn't crash - in full implementation you'd verify events
	})
}

func TestIntegration_FullResourceWorkflow(t *testing.T) {
	world := createTestWorld()
	createTestPlayer(world, 1)

	// Test complete workflow: validation -> deduction -> status check
	t.Run("CompleteWorkflow", func(t *testing.T) {
		player := world.GetPlayer(1)
		initialGold := player.Resources["gold"]

		// 1. Validate we can afford something
		validator := NewResourceValidator(world)
		cost := map[string]int{"gold": 150, "wood": 100}

		result := validator.ValidateResources(ResourceCheck{
			PlayerID: 1,
			Required: cost,
			Purpose:  "test_workflow",
		})

		if !result.Valid {
			t.Fatalf("Expected to be able to afford cost, got: %s", result.Error)
		}

		// 2. Deduct resources
		err := world.DeductResources(1, cost, "test_workflow")
		if err != nil {
			t.Fatalf("Expected successful deduction, got: %v", err)
		}

		// 3. Check final status
		status := world.GetResourceStatus(1)

		expectedGold := initialGold - 150
		if status.Resources["gold"] != expectedGold {
			t.Errorf("Expected final gold to be %d, got %d", expectedGold, status.Resources["gold"])
		}

		expectedWood := 500 - 100 // Initial wood minus cost
		if status.Resources["wood"] != expectedWood {
			t.Errorf("Expected final wood to be %d, got %d", expectedWood, status.Resources["wood"])
		}

		// 4. Verify spending tracking
		if status.ResourcesSpent["gold"] != 150 {
			t.Errorf("Expected spent gold to be 150, got %d", status.ResourcesSpent["gold"])
		}
	})
}