package engine

import (
	"testing"
	"time"

	"teraglest/internal/data"
)

// TestProductionSystemBasics tests basic production system functionality
func TestProductionSystemBasics(t *testing.T) {
	// Create minimal world for testing
	world := createTestWorldForProduction(t)
	productionSys := world.productionSys

	if productionSys == nil {
		t.Fatal("ProductionSystem should not be nil")
	}

	// Test technology tree initialization
	techTree := productionSys.GetTechnologyTree()
	if techTree == nil {
		t.Fatal("TechnologyTree should not be nil")
	}

	// Test population manager initialization
	populationMgr := productionSys.GetPopulationManager()
	if populationMgr == nil {
		t.Fatal("PopulationManager should not be nil")
	}
}

// TestBuildingConstruction tests worker units building structures
func TestBuildingConstruction(t *testing.T) {
	world := createTestWorldForProduction(t)
	productionSys := world.productionSys

	// Create a worker unit
	workerPos := Vector3{X: 10, Y: 0, Z: 10}
	worker, err := world.ObjectManager.CreateUnit(1, "worker", workerPos, nil)
	if err != nil {
		t.Fatalf("Failed to create worker: %v", err)
	}

	// Create a building under construction
	buildingPos := Vector3{X: 15, Y: 0, Z: 15}
	building, err := world.ObjectManager.CreateBuilding(1, "barracks", buildingPos, nil)
	if err != nil {
		t.Fatalf("Failed to create building: %v", err)
	}

	// Set building as under construction
	building.IsBuilt = false
	building.BuildProgress = 0.5 // 50% complete

	// Assign worker to building
	worker.State = UnitStateBuilding
	worker.BuildTarget = building

	// Simulate construction progress
	deltaTime := 1 * time.Second
	productionSys.ProcessWorkerConstruction(deltaTime)

	// Verify construction progressed
	if building.BuildProgress <= 0.5 {
		t.Errorf("Expected building progress to increase from 0.5, got %f", building.BuildProgress)
	}

	// Continue until construction complete
	for i := 0; i < 20 && !building.IsBuilt; i++ {
		productionSys.ProcessWorkerConstruction(deltaTime)
	}

	if !building.IsBuilt {
		t.Error("Building should be complete after sufficient construction time")
	}

	if worker.State != UnitStateIdle {
		t.Errorf("Worker should be idle after construction complete, got %s", worker.State.String())
	}

	if worker.BuildTarget != nil {
		t.Error("Worker should have no build target after construction complete")
	}
}

// TestUnitProduction tests building unit production
func TestUnitProduction(t *testing.T) {
	world := createTestWorldForProduction(t)
	productionSys := world.productionSys

	// Add resources to player
	world.AddResources(1, map[string]int{"wood": 200, "gold": 100}, "test")

	// Create a completed building
	buildingPos := Vector3{X: 20, Y: 0, Z: 20}
	building, err := world.ObjectManager.CreateBuilding(1, "barracks", buildingPos, nil)
	if err != nil {
		t.Fatalf("Failed to create building: %v", err)
	}
	building.IsBuilt = true

	// Issue unit production command
	unitType := "swordman"
	cost := map[string]int{"wood": 50, "gold": 25}
	duration := 5 * time.Second

	err = productionSys.IssueProductionCommand(building.ID, unitType, cost, duration)
	if err != nil {
		t.Fatalf("Failed to issue production command: %v", err)
	}

	// Verify production was queued
	queue, current, err := productionSys.GetProductionQueue(building.ID)
	if err != nil {
		t.Fatalf("Failed to get production queue: %v", err)
	}

	if current == nil && len(queue) == 0 {
		t.Fatal("Production should be started or queued")
	}

	// Simulate production progress
	deltaTime := 1 * time.Second
	initialUnitCount := len(world.ObjectManager.GetUnitsForPlayer(1))

	// Process production until complete
	for i := 0; i < 10; i++ {
		productionSys.ProcessBuildingProduction(deltaTime)
	}

	// Verify unit was created
	finalUnitCount := len(world.ObjectManager.GetUnitsForPlayer(1))
	if finalUnitCount <= initialUnitCount {
		t.Errorf("Expected unit count to increase from %d, got %d", initialUnitCount, finalUnitCount)
	}
}

// TestPopulationLimits tests population management
func TestPopulationLimits(t *testing.T) {
	world := createTestWorldForProduction(t)
	populationMgr := world.productionSys.GetPopulationManager()

	// Test initial population capacity
	canCreate, reason := populationMgr.CanCreateUnit(1, "worker")
	if !canCreate {
		t.Errorf("Should be able to create initial units: %s", reason)
	}

	// Test maximum units calculation
	maxWorkers := populationMgr.GetMaxUnitsCanCreate(1, "worker")
	if maxWorkers <= 0 {
		t.Error("Should be able to create some workers initially")
	}

	// Add housing building to increase capacity
	housePos := Vector3{X: 25, Y: 0, Z: 25}
	house, err := world.ObjectManager.CreateBuilding(1, "house", housePos, nil)
	if err != nil {
		t.Fatalf("Failed to create house: %v", err)
	}
	house.IsBuilt = true

	// Verify population capacity increased
	newStatus := populationMgr.GetPopulationStatus(1)
	if newStatus.MaxPopulation <= 10 {
		t.Errorf("Expected population capacity to increase with housing, got %d", newStatus.MaxPopulation)
	}
}

// TestTechnologyTree tests research and technology dependencies
func TestTechnologyTree(t *testing.T) {
	world := createTestWorldForProduction(t)
	techTree := world.productionSys.GetTechnologyTree()

	// Initialize player
	techTree.InitializePlayer(1)

	// Test basic technology research
	techName := "iron_weapons"
	canResearch, reason := techTree.CanResearchTechnology(1, techName)
	if !canResearch {
		t.Errorf("Should be able to research basic technology %s: %s", techName, reason)
	}

	// Start research
	err := techTree.StartResearch(1, techName, 1)
	if err != nil {
		t.Errorf("Failed to start research: %v", err)
	}

	// Process research
	deltaTime := 1 * time.Second
	for i := 0; i < 50; i++ { // 50 seconds should complete most research
		techTree.ProcessResearch(deltaTime)
	}

	// Verify technology was researched
	if !techTree.HasTechnology(1, techName) {
		t.Errorf("Technology %s should be researched after processing time", techName)
	}

	// Test technology dependency
	dependentTech := "advanced_construction"
	canResearchDependent, _ := techTree.CanResearchTechnology(1, dependentTech)
	if canResearchDependent {
		// This should work if construction_efficiency is a prerequisite and available
		err := techTree.StartResearch(1, dependentTech, 1)
		if err == nil {
			// Process research
			for i := 0; i < 70; i++ { // Longer research time
				techTree.ProcessResearch(deltaTime)
			}

			if !techTree.HasTechnology(1, dependentTech) {
				t.Errorf("Dependent technology %s should be researched", dependentTech)
			}
		}
	}
}

// TestProductionQueueManagement tests production queue operations
func TestProductionQueueManagement(t *testing.T) {
	world := createTestWorldForProduction(t)
	productionSys := world.productionSys

	// Add resources
	world.AddResources(1, map[string]int{"wood": 500, "gold": 300}, "test")

	// Create building
	buildingPos := Vector3{X: 30, Y: 0, Z: 30}
	building, err := world.ObjectManager.CreateBuilding(1, "barracks", buildingPos, nil)
	if err != nil {
		t.Fatalf("Failed to create building: %v", err)
	}
	building.IsBuilt = true

	// Queue multiple units
	unitTypes := []string{"swordman", "archer", "worker"}
	cost := map[string]int{"wood": 40, "gold": 20}
	duration := 2 * time.Second

	for _, unitType := range unitTypes {
		err := productionSys.IssueProductionCommand(building.ID, unitType, cost, duration)
		if err != nil {
			t.Errorf("Failed to queue %s: %v", unitType, err)
		}
	}

	// Verify queue length
	queue, current, err := productionSys.GetProductionQueue(building.ID)
	if err != nil {
		t.Fatalf("Failed to get production queue: %v", err)
	}

	totalInProduction := len(queue)
	if current != nil {
		totalInProduction++
	}

	if totalInProduction != len(unitTypes) {
		t.Errorf("Expected %d items in production, got %d", len(unitTypes), totalInProduction)
	}

	// Test production cancellation
	err = productionSys.CancelProduction(building.ID)
	if err != nil {
		t.Errorf("Failed to cancel production: %v", err)
	}

	// Verify current production was cancelled
	_, currentAfterCancel, err := productionSys.GetProductionQueue(building.ID)
	if err != nil {
		t.Errorf("Failed to get queue after cancellation: %v", err)
	}

	// Should either be nil or the next item from queue
	if currentAfterCancel != nil && currentAfterCancel == current {
		t.Error("Current production should have changed after cancellation")
	}
}

// Helper function to create a test world
func createTestWorldForProduction(t *testing.T) *World {
	settings := GameSettings{
		PlayerFactions: map[int]string{1: "romans"},
		MaxPlayers:     4,
		ResourceMultiplier: 1.0,
	}

	// Create world without initializing players (to avoid AssetManager issues in tests)
	world, err := NewWorld(settings, &data.TechTree{}, &data.AssetManager{})
	if err != nil {
		t.Fatalf("Failed to create test world: %v", err)
	}

	// Manually create a test player without faction loading
	player := &Player{
		ID:          1,
		Name:        "TestPlayer",
		FactionName: "romans",
		IsAI:        false,
		Resources:   make(map[string]int),
		FactionData: nil, // Skip faction loading for tests
	}

	// Initialize with basic resources
	player.Resources["wood"] = 1000
	player.Resources["gold"] = 1000
	player.Resources["stone"] = 500
	player.Resources["energy"] = 500

	world.players[1] = player
	world.initialized = true

	return world
}