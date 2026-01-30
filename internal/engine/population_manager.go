package engine

import (
	"fmt"
)

// PopulationManager handles population and housing management for players
type PopulationManager struct {
	world *World
}

// PopulationStatus represents the current population state for a player
type PopulationStatus struct {
	CurrentPopulation int                // Current population count
	MaxPopulation     int                // Maximum population (housing capacity)
	HousingBuildings  []*GameBuilding    // Buildings that provide housing
	PopulationUnits   []*GameUnit        // Units that consume population
}

// UnitPopulationCost represents population costs for different unit types
var UnitPopulationCost = map[string]int{
	// Default population costs - can be overridden from XML
	"worker":       1,
	"archer":       1,
	"swordsman":    2,
	"cavalry":      3,
	"siege":        5,
	"hero":         10,
}

// BuildingHousingCapacity represents housing capacity for different building types
var BuildingHousingCapacity = map[string]int{
	// Default housing values - can be overridden from XML
	"house":        5,
	"barracks":     10,
	"castle":       20,
	"town_hall":    15,
}

// NewPopulationManager creates a new population manager
func NewPopulationManager(world *World) *PopulationManager {
	return &PopulationManager{
		world: world,
	}
}

// GetPopulationStatus returns current population status for a player
func (pm *PopulationManager) GetPopulationStatus(playerID int) PopulationStatus {
	buildingMap := pm.world.ObjectManager.GetBuildingsForPlayer(playerID)
	unitMap := pm.world.ObjectManager.UnitManager.GetUnitsForPlayer(playerID)

	// Convert maps to slices
	buildings := make([]*GameBuilding, 0, len(buildingMap))
	for _, building := range buildingMap {
		buildings = append(buildings, building)
	}

	units := make([]*GameUnit, 0, len(unitMap))
	for _, unit := range unitMap {
		units = append(units, unit)
	}

	maxPop := pm.calculateHousingCapacity(buildings)
	currentPop := pm.calculateCurrentPopulation(units)

	return PopulationStatus{
		CurrentPopulation: currentPop,
		MaxPopulation:     maxPop,
		HousingBuildings:  pm.filterHousingBuildings(buildings),
		PopulationUnits:   pm.filterPopulationUnits(units),
	}
}

// CanCreateUnit checks if a player can create a unit without exceeding population limits
func (pm *PopulationManager) CanCreateUnit(playerID int, unitType string) (bool, string) {
	status := pm.GetPopulationStatus(playerID)
	unitPopCost := pm.getUnitPopulationCost(playerID, unitType)

	if status.CurrentPopulation+unitPopCost > status.MaxPopulation {
		return false, fmt.Sprintf("population limit exceeded: %d/%d (need %d for %s)",
			status.CurrentPopulation, status.MaxPopulation, unitPopCost, unitType)
	}

	return true, ""
}

// CanCreateMultipleUnits checks if multiple units can be created
func (pm *PopulationManager) CanCreateMultipleUnits(playerID int, unitType string, count int) (bool, string) {
	status := pm.GetPopulationStatus(playerID)
	unitPopCost := pm.getUnitPopulationCost(playerID, unitType)
	totalPopCost := unitPopCost * count

	if status.CurrentPopulation+totalPopCost > status.MaxPopulation {
		maxPossible := (status.MaxPopulation - status.CurrentPopulation) / unitPopCost
		return false, fmt.Sprintf("population limit: can only create %d %s units (requested %d)",
			maxPossible, unitType, count)
	}

	return true, ""
}

// calculateHousingCapacity calculates total housing capacity from buildings
func (pm *PopulationManager) calculateHousingCapacity(buildings []*GameBuilding) int {
	capacity := 0

	for _, building := range buildings {
		if building.IsBuilt && building.Health > 0 {
			housingValue := pm.getBuildingHousingValue(building)
			capacity += housingValue
		}
	}

	// Minimum capacity of 10 (starting population allowance)
	if capacity < 10 {
		capacity = 10
	}

	return capacity
}

// calculateCurrentPopulation calculates current population from units
func (pm *PopulationManager) calculateCurrentPopulation(units []*GameUnit) int {
	population := 0

	for _, unit := range units {
		if unit.Health > 0 && unit.State != UnitStateDead {
			popCost := pm.getUnitPopulationCostFromUnit(unit)
			population += popCost
		}
	}

	return population
}

// getBuildingHousingValue gets housing value for a specific building
func (pm *PopulationManager) getBuildingHousingValue(building *GameBuilding) int {
	// First try to get from XML resource requirements (negative housing value)
	housingFromXML := pm.getBuildingHousingFromXML(building)
	if housingFromXML != 0 {
		return housingFromXML
	}

	// Fallback to default building capacity
	if capacity, exists := BuildingHousingCapacity[building.BuildingType]; exists {
		return capacity
	}

	// Default: no housing provided
	return 0
}

// getBuildingHousingFromXML extracts housing value from building's XML resource requirements
func (pm *PopulationManager) getBuildingHousingFromXML(building *GameBuilding) int {
	// Get building definition from AssetManager
	player := pm.world.GetPlayer(building.PlayerID)
	if player == nil || player.FactionData == nil {
		return 0
	}

	buildingDef, err := pm.world.assetMgr.LoadUnit(player.FactionName, building.BuildingType)
	if err != nil {
		return 0
	}

	// Look for housing in resource requirements (negative value = provides housing)
	for _, req := range buildingDef.Unit.Parameters.ResourceRequirements {
		if req.Name == "housing" && req.Amount < 0 {
			return -req.Amount // Convert negative requirement to positive housing capacity
		}
	}

	return 0
}

// getUnitPopulationCost gets population cost for a unit type from a specific player
func (pm *PopulationManager) getUnitPopulationCost(playerID int, unitType string) int {
	// First try to get from XML resource requirements
	popFromXML := pm.getUnitPopulationCostFromXML(playerID, unitType)
	if popFromXML > 0 {
		return popFromXML
	}

	// Fallback to default unit costs
	if cost, exists := UnitPopulationCost[unitType]; exists {
		return cost
	}

	// Default: 1 population per unit
	return 1
}

// getUnitPopulationCostFromXML extracts population cost from unit's XML resource requirements
func (pm *PopulationManager) getUnitPopulationCostFromXML(playerID int, unitType string) int {
	// Get unit definition from AssetManager
	player := pm.world.GetPlayer(playerID)
	if player == nil || player.FactionData == nil {
		return 0
	}

	unitDef, err := pm.world.assetMgr.LoadUnit(player.FactionName, unitType)
	if err != nil {
		return 0
	}

	// Look for housing in resource requirements (positive value = consumes population)
	for _, req := range unitDef.Unit.Parameters.ResourceRequirements {
		if req.Name == "housing" && req.Amount > 0 {
			return req.Amount
		}
	}

	return 0
}

// getUnitPopulationCostFromUnit gets population cost for an existing unit
func (pm *PopulationManager) getUnitPopulationCostFromUnit(unit *GameUnit) int {
	return pm.getUnitPopulationCost(unit.PlayerID, unit.UnitType)
}

// filterHousingBuildings returns only buildings that provide housing
func (pm *PopulationManager) filterHousingBuildings(buildings []*GameBuilding) []*GameBuilding {
	housing := make([]*GameBuilding, 0)

	for _, building := range buildings {
		if pm.getBuildingHousingValue(building) > 0 {
			housing = append(housing, building)
		}
	}

	return housing
}

// filterPopulationUnits returns only units that consume population
func (pm *PopulationManager) filterPopulationUnits(units []*GameUnit) []*GameUnit {
	population := make([]*GameUnit, 0)

	for _, unit := range units {
		if pm.getUnitPopulationCostFromUnit(unit) > 0 {
			population = append(population, unit)
		}
	}

	return population
}

// GetMaxUnitsCanCreate calculates how many units of a type can be created within population limits
func (pm *PopulationManager) GetMaxUnitsCanCreate(playerID int, unitType string) int {
	status := pm.GetPopulationStatus(playerID)
	unitPopCost := pm.getUnitPopulationCost(playerID, unitType)

	if unitPopCost <= 0 {
		return 999 // No population cost, can create many
	}

	availablePopulation := status.MaxPopulation - status.CurrentPopulation
	maxUnits := availablePopulation / unitPopCost

	if maxUnits < 0 {
		maxUnits = 0
	}

	return maxUnits
}

// ValidatePopulation validates population state for debugging
func (pm *PopulationManager) ValidatePopulation(playerID int) []string {
	issues := make([]string, 0)
	status := pm.GetPopulationStatus(playerID)

	// Check for population limit violations
	if status.CurrentPopulation > status.MaxPopulation {
		issues = append(issues, fmt.Sprintf("Population over limit: %d/%d",
			status.CurrentPopulation, status.MaxPopulation))
	}

	// Check for negative population
	if status.CurrentPopulation < 0 {
		issues = append(issues, fmt.Sprintf("Negative population: %d", status.CurrentPopulation))
	}

	// Check for negative housing
	if status.MaxPopulation < 0 {
		issues = append(issues, fmt.Sprintf("Negative housing capacity: %d", status.MaxPopulation))
	}

	return issues
}