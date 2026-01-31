package engine

import (
	"fmt"
	"math"
	"sync"
	"time"

	"teraglest/internal/data"
)

// ProductionSystem manages unit production and technology research for buildings
type ProductionSystem struct {
	world           *World
	technologyTree  *TechnologyTree
	populationMgr   *PopulationManager
	mutex           sync.RWMutex
}

// NewProductionSystem creates a new production system
func NewProductionSystem(world *World) *ProductionSystem {
	return &ProductionSystem{
		world:         world,
		technologyTree: NewTechnologyTree(),
		populationMgr: NewPopulationManager(world),
	}
}

// ProcessBuildingProduction processes production for all buildings
func (ps *ProductionSystem) ProcessBuildingProduction(deltaTime time.Duration) {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()

	// Get all buildings from all players (iterate through actual players)
	for playerID := range ps.world.players {
		buildings := ps.world.ObjectManager.GetBuildingsForPlayer(playerID)
		for _, building := range buildings {
			if building.IsBuilt {
				ps.processBuildingProductionQueue(building, deltaTime)
				ps.processBuildingUpgrades(building, deltaTime)
			}
		}
	}
}

// processBuildingProductionQueue handles unit production for a single building
func (ps *ProductionSystem) processBuildingProductionQueue(building *GameBuilding, deltaTime time.Duration) {
	building.mutex.Lock()
	defer building.mutex.Unlock()

	// Start next production if nothing is currently being produced
	if building.CurrentProduction == nil && len(building.ProductionQueue) > 0 {
		building.CurrentProduction = &building.ProductionQueue[0]
		building.ProductionQueue = building.ProductionQueue[1:]
		building.CurrentProduction.StartTime = time.Now()
	}

	// Process current production
	if building.CurrentProduction != nil {
		production := building.CurrentProduction

		// Calculate production progress
		elapsed := time.Since(production.StartTime)
		totalDuration := production.Duration
		production.Progress = float32(elapsed.Seconds()) / float32(totalDuration.Seconds())

		// Check if production is complete
		if production.Progress >= 1.0 {
			ps.completeProduction(building, production)
			building.CurrentProduction = nil
		}
	}
}

// processBuildingUpgrades handles building upgrade processing
func (ps *ProductionSystem) processBuildingUpgrades(building *GameBuilding, deltaTime time.Duration) {
	building.mutex.Lock()
	defer building.mutex.Unlock()

	if building.CurrentUpgrade != nil {
		upgrade := building.CurrentUpgrade

		// Calculate upgrade progress
		elapsed := time.Since(upgrade.StartTime)
		upgrade.Progress = float32(elapsed.Seconds()) / float32(upgrade.Duration.Seconds())

		// Check if upgrade is complete
		if upgrade.Progress >= 1.0 {
			ps.applyUpgrade(building, upgrade)
			building.CurrentUpgrade = nil
		}
	}
}

// completeProduction handles the completion of a production item
func (ps *ProductionSystem) completeProduction(building *GameBuilding, production *ProductionItem) {
	switch production.ItemType {
	case "unit":
		ps.spawnUnit(building, production)
	case "upgrade":
		// Convert ProductionItem to UpgradeItem for upgrade processing
		upgrade := &UpgradeItem{
			UpgradeType: production.ItemName,
			UpgradeName: production.ItemName,
			Progress:    production.Progress,
			Duration:    production.Duration,
			Cost:        production.Cost,
			StartTime:   production.StartTime,
		}
		ps.applyUpgrade(building, upgrade)
	case "research":
		ps.completeResearch(building, production)
	}
}

// spawnUnit creates a new unit when production completes
func (ps *ProductionSystem) spawnUnit(building *GameBuilding, production *ProductionItem) {
	// Check population limits using the existing PopulationManager interface
	canCreate, reason := ps.populationMgr.CanCreateUnit(building.PlayerID, production.ItemName)
	if !canCreate {
		// Refund resources if population limit reached
		if len(production.Cost) > 0 {
			ps.world.AddResources(building.PlayerID, production.Cost, "production_refund_population_limit")
		}
		fmt.Printf("Unit production failed for player %d: %s\n", building.PlayerID, reason)
		return
	}

	// Find spawn position near building
	spawnPos := ps.findUnitSpawnPosition(building)

	// Load unit definition
	player := ps.world.GetPlayer(building.PlayerID)
	var unitDef *data.UnitDefinition
	if player != nil && player.FactionData != nil {
		unitDef, _ = ps.world.assetMgr.LoadUnit(player.FactionName, production.ItemName)
	}

	// Create the unit
	unit, err := ps.world.ObjectManager.CreateUnit(
		building.PlayerID,
		production.ItemName,
		spawnPos,
		unitDef,
	)

	if err != nil {
		// Refund resources on spawn failure
		if len(production.Cost) > 0 {
			ps.world.AddResources(building.PlayerID, production.Cost, "production_refund_spawn_failure")
		}
		return
	}

	// Unit creation successful - population tracking is handled by existing systems
	// The PopulationManager will query units when needed rather than tracking directly

	// Emit production complete event
	ps.emitProductionEvent(building, production, unit.ID)
}

// applyUpgrade applies the effects of a completed upgrade to a building
func (ps *ProductionSystem) applyUpgrade(building *GameBuilding, upgrade *UpgradeItem) {
	// Apply upgrade effects based on upgrade type
	switch upgrade.UpgradeType {
	case "production_speed":
		building.ProductionRate *= 1.25 // 25% faster production
	case "durability":
		building.MaxHealth = int(float32(building.MaxHealth) * 1.2) // 20% more health
		building.Health = building.MaxHealth // Heal to full
	case "resource_efficiency":
		// Increase resource generation by 25%
		for resType, rate := range building.ResourceGeneration {
			building.ResourceGeneration[resType] = rate * 1.25
		}
	}

	// Emit upgrade complete event
	ps.emitUpgradeCompleteEvent(building, upgrade)
}

// completeResearch handles research completion (placeholder for future research system integration)
func (ps *ProductionSystem) completeResearch(building *GameBuilding, research *ProductionItem) {
	// Research completion is handled by TechnologyTree.completeResearch
	// This is a placeholder for any building-specific research effects
	ps.emitResearchCompleteEvent(building, research)
}

// findUnitSpawnPosition finds a valid spawn position near a building
func (ps *ProductionSystem) findUnitSpawnPosition(building *GameBuilding) Vector3 {
	basePos := building.Position
	tileSize := float64(ps.world.GetTileSize())

	// Try positions in expanding radius around building
	for radius := 1.0; radius <= 5.0; radius += 1.0 {
		for angle := 0.0; angle < 360.0; angle += 45.0 {
			// Calculate test position
			angleRad := angle * 3.14159 / 180.0
			testPos := Vector3{
				X: basePos.X + radius*math.Cos(angleRad),
				Y: basePos.Y,
				Z: basePos.Z + radius*math.Sin(angleRad),
			}

			// Convert to grid position for walkability check
			gridPos := Vector2i{
				X: int(testPos.X / tileSize),
				Y: int(testPos.Z / tileSize),
			}

			// Check if position is walkable and not occupied
			if ps.world.IsPositionWalkable(gridPos) {
				return testPos
			}
		}
	}

	// Fallback to building position if no free space found
	return basePos
}

// IssueProductionCommand adds a unit to the building's production queue
func (ps *ProductionSystem) IssueProductionCommand(buildingID int, unitType string, cost map[string]int, duration time.Duration) error {
	building := ps.world.ObjectManager.GetBuilding(buildingID)
	if building == nil {
		return fmt.Errorf("building %d not found", buildingID)
	}

	if !building.IsBuilt {
		return fmt.Errorf("building is not complete")
	}

	// Check technology requirements
	if !ps.technologyTree.HasRequiredTech(building.PlayerID, unitType) {
		return fmt.Errorf("missing required technology for %s", unitType)
	}

	// Check population capacity using existing PopulationManager interface
	canCreate, reason := ps.populationMgr.CanCreateUnit(building.PlayerID, unitType)
	if !canCreate {
		return fmt.Errorf("population limit reached for %s: %s", unitType, reason)
	}

	// Validate and deduct resources
	if len(cost) > 0 {
		err := ps.world.DeductResources(building.PlayerID, cost, "unit_production")
		if err != nil {
			return fmt.Errorf("insufficient resources for %s production: %w", unitType, err)
		}
	}

	// Create production item
	productionItem := ProductionItem{
		ItemType:  "unit",
		ItemName:  unitType,
		Progress:  0.0,
		Duration:  duration,
		Cost:      cost,
		StartTime: time.Time{}, // Will be set when production starts
	}

	// Add to production queue
	building.mutex.Lock()
	building.ProductionQueue = append(building.ProductionQueue, productionItem)
	building.mutex.Unlock()

	return nil
}

// ProcessWorkerConstruction handles worker units that are building structures
func (ps *ProductionSystem) ProcessWorkerConstruction(deltaTime time.Duration) {
	// Get all worker units from all players (iterate through actual players)
	for playerID := range ps.world.players {
		units := ps.world.ObjectManager.GetUnitsForPlayer(playerID)
		for _, unit := range units {
			if unit.State == UnitStateBuilding && unit.BuildTarget != nil {
				ps.processUnitConstruction(unit, deltaTime)
			}
		}
	}
}

// processUnitConstruction handles a single worker unit building progress
func (ps *ProductionSystem) processUnitConstruction(unit *GameUnit, deltaTime time.Duration) {
	building := unit.BuildTarget
	if building == nil || building.IsBuilt {
		// Construction already complete
		unit.BuildTarget = nil
		unit.State = UnitStateIdle
		unit.CurrentCommand = nil
		return
	}

	// Calculate construction rate (0.1 = 10 seconds to complete building)
	constructionRate := 0.1 // Base construction rate per second

	// Apply worker efficiency bonuses
	workerEfficiency := ps.getWorkerEfficiency(unit)
	constructionRate *= float64(workerEfficiency)

	// Update building progress
	building.mutex.Lock()
	building.BuildProgress += float32(constructionRate * deltaTime.Seconds())

	if building.BuildProgress >= 1.0 {
		building.BuildProgress = 1.0
		building.IsBuilt = true
		building.CompletionTime = time.Now()

		// Construction complete - worker becomes idle
		unit.BuildTarget = nil
		unit.State = UnitStateIdle
		unit.CurrentCommand = nil

		// Emit construction complete event
		ps.emitConstructionCompleteEvent(building, unit.ID)
	}
	building.mutex.Unlock()
}

// getWorkerEfficiency calculates construction efficiency based on worker skills/upgrades
func (ps *ProductionSystem) getWorkerEfficiency(unit *GameUnit) float32 {
	baseEfficiency := float32(1.0)

	// Apply technology bonuses
	if ps.technologyTree.HasTechnology(unit.PlayerID, "construction_efficiency") {
		baseEfficiency *= 1.25
	}

	if ps.technologyTree.HasTechnology(unit.PlayerID, "advanced_construction") {
		baseEfficiency *= 1.5
	}

	// Worker type bonuses
	switch unit.UnitType {
	case "worker", "peasant":
		baseEfficiency *= 1.0
	case "engineer", "master_builder":
		baseEfficiency *= 1.3
	case "initiate": // Magic faction worker
		baseEfficiency *= 0.9 // Slower at construction
	}

	return baseEfficiency
}

// emitProductionEvent emits an event when unit production completes
func (ps *ProductionSystem) emitProductionEvent(building *GameBuilding, production *ProductionItem, unitID int) {
	// Event system integration - placeholder for future event system
	event := ProductionCompleteEvent{
		BuildingID: building.ID,
		PlayerID:   building.PlayerID,
		UnitType:   production.ItemName,
		UnitID:     unitID,
		Timestamp:  time.Now(),
	}
	_ = event // Use event when event system is implemented
}

// emitConstructionCompleteEvent emits an event when building construction completes
func (ps *ProductionSystem) emitConstructionCompleteEvent(building *GameBuilding, workerID int) {
	event := ConstructionCompleteEvent{
		BuildingID: building.ID,
		PlayerID:   building.PlayerID,
		WorkerID:   workerID,
		Timestamp:  time.Now(),
	}
	_ = event // Use event when event system is implemented
}

// emitUpgradeCompleteEvent emits an event when building upgrade completes
func (ps *ProductionSystem) emitUpgradeCompleteEvent(building *GameBuilding, upgrade *UpgradeItem) {
	event := UpgradeCompleteEvent{
		BuildingID:   building.ID,
		PlayerID:     building.PlayerID,
		UpgradeType:  upgrade.UpgradeType,
		UpgradeName:  upgrade.UpgradeName,
		Timestamp:    time.Now(),
	}
	_ = event // Use event when event system is implemented
}

// emitResearchCompleteEvent emits an event when research completes
func (ps *ProductionSystem) emitResearchCompleteEvent(building *GameBuilding, research *ProductionItem) {
	event := ResearchCompleteEvent{
		BuildingID:   building.ID,
		PlayerID:     building.PlayerID,
		ResearchName: research.ItemName,
		Timestamp:    time.Now(),
	}
	_ = event // Use event when event system is implemented
}

// GetProductionQueue returns the production queue for a building
func (ps *ProductionSystem) GetProductionQueue(buildingID int) ([]ProductionItem, *ProductionItem, error) {
	building := ps.world.ObjectManager.GetBuilding(buildingID)
	if building == nil {
		return nil, nil, fmt.Errorf("building %d not found", buildingID)
	}

	building.mutex.RLock()
	defer building.mutex.RUnlock()

	// Copy queue to avoid race conditions
	queue := make([]ProductionItem, len(building.ProductionQueue))
	copy(queue, building.ProductionQueue)

	var current *ProductionItem
	if building.CurrentProduction != nil {
		current = &ProductionItem{
			ItemType:  building.CurrentProduction.ItemType,
			ItemName:  building.CurrentProduction.ItemName,
			Progress:  building.CurrentProduction.Progress,
			Duration:  building.CurrentProduction.Duration,
			Cost:      building.CurrentProduction.Cost,
			StartTime: building.CurrentProduction.StartTime,
		}
	}

	return queue, current, nil
}

// CancelProduction cancels the current production and refunds resources
func (ps *ProductionSystem) CancelProduction(buildingID int) error {
	building := ps.world.ObjectManager.GetBuilding(buildingID)
	if building == nil {
		return fmt.Errorf("building %d not found", buildingID)
	}

	building.mutex.Lock()
	defer building.mutex.Unlock()

	if building.CurrentProduction != nil {
		// Refund partial resources based on progress
		refundRatio := 1.0 - float64(building.CurrentProduction.Progress)
		if refundRatio > 0 && len(building.CurrentProduction.Cost) > 0 {
			refundResources := make(map[string]int)
			for resource, amount := range building.CurrentProduction.Cost {
				refundAmount := int(float64(amount) * refundRatio)
				if refundAmount > 0 {
					refundResources[resource] = refundAmount
				}
			}
			if len(refundResources) > 0 {
				ps.world.AddResources(building.PlayerID, refundResources, "production_cancellation")
			}
		}

		building.CurrentProduction = nil
	}

	return nil
}

// Event types for future event system integration
type ProductionCompleteEvent struct {
	BuildingID int
	PlayerID   int
	UnitType   string
	UnitID     int
	Timestamp  time.Time
}

type ConstructionCompleteEvent struct {
	BuildingID int
	PlayerID   int
	WorkerID   int
	Timestamp  time.Time
}

type UpgradeCompleteEvent struct {
	BuildingID   int
	PlayerID     int
	UpgradeType  string
	UpgradeName  string
	Timestamp    time.Time
}

type ResearchCompleteEvent struct {
	BuildingID   int
	PlayerID     int
	ResearchName string
	Timestamp    time.Time
}

// Update integrates production system updates into the main game loop
func (ps *ProductionSystem) Update(deltaTime time.Duration) {
	ps.ProcessBuildingProduction(deltaTime)
	ps.ProcessWorkerConstruction(deltaTime)
}

// GetTechnologyTree returns the technology tree for external access
func (ps *ProductionSystem) GetTechnologyTree() *TechnologyTree {
	return ps.technologyTree
}

// GetPopulationManager returns the population manager for external access
func (ps *ProductionSystem) GetPopulationManager() *PopulationManager {
	return ps.populationMgr
}