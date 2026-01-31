package engine

import (
	"math"
	"sort"
	"time"
)

// EconomicManager handles AI economic decisions and resource management
type EconomicManager struct {
	playerID       int           // Player ID this manager controls
	world          *World        // Game world reference
	strategicAI    *StrategicAI  // Parent strategic AI
	lastEvaluation time.Time     // Last economic evaluation time

	// Economic state tracking
	resourcePriorities map[string]float64 // Priority for each resource type
	workerAllocation   map[string]int     // Workers assigned to each resource
	productionQueue    []ProductionOrder  // Planned production orders
	economicTargets    []EconomicTarget   // Economic expansion targets
}

// ProductionOrder represents a planned economic production
type ProductionOrder struct {
	Type         string                 // Type of unit/building to produce
	Priority     float64               // Priority score
	Building     string                // Building that should produce it
	Parameters   map[string]interface{} // Additional parameters
	Deadline     time.Time             // When this should be completed by
}

// EconomicTarget represents an economic expansion opportunity
type EconomicTarget struct {
	Type        string    // "resource_node", "expansion_site", "trade_route"
	Location    Vector3   // World position
	Value       float64   // Economic value score
	Risk        float64   // Risk factor
	Distance    float64   // Distance from nearest controlled area
	Priority    float64   // Combined priority score
}

// NewEconomicManager creates a new economic manager
func NewEconomicManager(playerID int, world *World, strategicAI *StrategicAI) *EconomicManager {
	em := &EconomicManager{
		playerID:           playerID,
		world:             world,
		strategicAI:       strategicAI,
		resourcePriorities: make(map[string]float64),
		workerAllocation:   make(map[string]int),
		productionQueue:    make([]ProductionOrder, 0),
		economicTargets:    make([]EconomicTarget, 0),
	}

	// Initialize default resource priorities
	em.resourcePriorities["gold"] = 0.8
	em.resourcePriorities["wood"] = 0.7
	em.resourcePriorities["stone"] = 0.6
	em.resourcePriorities["energy"] = 0.5

	return em
}

// Update performs economic management decisions
func (em *EconomicManager) Update(deltaTime time.Duration) {
	// Evaluate economic situation every 3 seconds
	if time.Since(em.lastEvaluation) >= 3*time.Second {
		em.evaluateEconomicSituation()
		em.updateResourcePriorities()
		em.planProduction()
		em.manageWorkerAllocation()
		em.lastEvaluation = time.Now()
	}

	em.executeProductionOrders()
}

// ExecuteEconomicFocus implements economic focus decisions from strategic AI
func (em *EconomicManager) ExecuteEconomicFocus(parameters map[string]interface{}) {
	focus, exists := parameters["focus"]
	if !exists {
		return
	}

	switch focus {
	case "workers":
		em.prioritizeWorkerProduction()
	case "resources":
		em.prioritizeResourceBuildings()
	case "infrastructure":
		em.prioritizeInfrastructure()
	}
}

// evaluateEconomicSituation analyzes current economic state
func (em *EconomicManager) evaluateEconomicSituation() {
	player := em.world.GetPlayer(em.playerID)
	if player == nil {
		return
	}

	// Evaluate resource stockpiles
	for resType := range em.resourcePriorities {
		amount, exists := player.Resources[resType]
		if !exists {
			amount = 0
		}

		// Calculate resource urgency based on current amount and consumption
		consumption := em.estimateResourceConsumption(resType)
		stockpileDays := float64(amount) / math.Max(consumption, 1.0)

		// Adjust priority based on stockpile duration
		if stockpileDays < 2.0 {
			em.resourcePriorities[resType] = 1.0 // Critical
		} else if stockpileDays < 5.0 {
			em.resourcePriorities[resType] = 0.8 // High
		} else if stockpileDays < 10.0 {
			em.resourcePriorities[resType] = 0.5 // Medium
		} else {
			em.resourcePriorities[resType] = 0.2 // Low
		}
	}

	// Find economic expansion opportunities
	em.identifyEconomicTargets()
}

// updateResourcePriorities adjusts resource priorities based on strategic phase
func (em *EconomicManager) updateResourcePriorities() {
	state := em.strategicAI.GetStrategyState()
	personality := em.strategicAI.GetPersonality()

	// Adjust priorities based on game phase
	switch state.Phase {
	case PhaseEarlyGame:
		// Early game: prioritize basic resources
		em.resourcePriorities["gold"] = 0.9
		em.resourcePriorities["wood"] = 0.8
		em.resourcePriorities["stone"] = 0.4
		em.resourcePriorities["energy"] = 0.3

	case PhaseMidGame:
		// Mid game: balanced resource needs
		em.resourcePriorities["gold"] = 0.8
		em.resourcePriorities["wood"] = 0.7
		em.resourcePriorities["stone"] = 0.7
		em.resourcePriorities["energy"] = 0.6

	case PhaseLateGame:
		// Late game: advanced resources more important
		em.resourcePriorities["gold"] = 0.7
		em.resourcePriorities["wood"] = 0.5
		em.resourcePriorities["stone"] = 0.8
		em.resourcePriorities["energy"] = 0.9

	case PhaseEndGame:
		// End game: focus on military resources
		em.resourcePriorities["gold"] = 0.9
		em.resourcePriorities["energy"] = 0.9
		em.resourcePriorities["stone"] = 0.7
		em.resourcePriorities["wood"] = 0.4
	}

	// Adjust based on personality
	if personality.EconomicFocus > 0.7 {
		// Boost all resource priorities for economic-focused AI
		for resType := range em.resourcePriorities {
			em.resourcePriorities[resType] = math.Min(1.0, em.resourcePriorities[resType]*1.2)
		}
	}
}

// planProduction creates production orders for economic units and buildings
func (em *EconomicManager) planProduction() {
	// Clear old production orders
	em.productionQueue = em.productionQueue[:0]

	// Assess worker needs
	em.planWorkerProduction()

	// Assess building needs
	em.planBuildingProduction()

	// Sort by priority
	sort.Slice(em.productionQueue, func(i, j int) bool {
		return em.productionQueue[i].Priority > em.productionQueue[j].Priority
	})
}

// planWorkerProduction determines worker production needs
func (em *EconomicManager) planWorkerProduction() {
	currentWorkers := em.countWorkers()
	optimalWorkers := em.calculateOptimalWorkerCount()

	if currentWorkers < optimalWorkers {
		deficit := optimalWorkers - currentWorkers
		for i := 0; i < deficit && i < 3; i++ { // Limit to 3 workers per planning cycle
			priority := 0.8 - float64(i)*0.1 // Decreasing priority

			em.productionQueue = append(em.productionQueue, ProductionOrder{
				Type:     "worker",
				Priority: priority,
				Building: "main_building", // Where workers are produced
				Parameters: map[string]interface{}{
					"resource_focus": em.findHighestPriorityResource(),
				},
				Deadline: time.Now().Add(time.Duration(30+i*15) * time.Second),
			})
		}
	}
}

// planBuildingProduction determines building construction needs
func (em *EconomicManager) planBuildingProduction() {
	state := em.strategicAI.GetStrategyState()

	// Resource buildings based on priorities
	for resType, priority := range em.resourcePriorities {
		if priority > 0.6 {
			buildingType := em.getResourceBuildingType(resType)
			currentCount := em.countResourceBuildings(buildingType)
			optimalCount := em.calculateOptimalBuildingCount(buildingType)

			if currentCount < optimalCount {
				em.productionQueue = append(em.productionQueue, ProductionOrder{
					Type:     buildingType,
					Priority: priority * 0.7,
					Building: "worker", // Workers construct buildings
					Parameters: map[string]interface{}{
						"resource_type": resType,
						"location_type": "resource_area",
					},
					Deadline: time.Now().Add(2 * time.Minute),
				})
			}
		}
	}

	// Housing buildings if population is constrained
	if float64(state.Population)/float64(state.MaxPopulation) > 0.8 {
		em.productionQueue = append(em.productionQueue, ProductionOrder{
			Type:     "house",
			Priority: 0.85,
			Building: "worker",
			Parameters: map[string]interface{}{
				"location_type": "base_area",
			},
			Deadline: time.Now().Add(90 * time.Second),
		})
	}
}

// manageWorkerAllocation assigns workers to optimal tasks
func (em *EconomicManager) manageWorkerAllocation() {
	workers := em.getAvailableWorkers()

	// Clear current allocations
	for resType := range em.workerAllocation {
		em.workerAllocation[resType] = 0
	}

	// Sort resources by priority
	var resources []string
	for resType := range em.resourcePriorities {
		resources = append(resources, resType)
	}
	sort.Slice(resources, func(i, j int) bool {
		return em.resourcePriorities[resources[i]] > em.resourcePriorities[resources[j]]
	})

	// Allocate workers to highest priority resources
	workerIndex := 0
	for _, resType := range resources {
		if workerIndex >= len(workers) {
			break
		}

		// Determine optimal workers for this resource
		optimalWorkers := em.calculateOptimalResourceWorkers(resType)
		assigned := math.Min(float64(optimalWorkers), float64(len(workers)-workerIndex))

		em.workerAllocation[resType] = int(assigned)

		// Actually assign workers (would interact with unit AI system)
		for i := 0; i < int(assigned); i++ {
			em.assignWorkerToResource(workers[workerIndex+i], resType)
		}

		workerIndex += int(assigned)
	}
}

// executeProductionOrders carries out planned production
func (em *EconomicManager) executeProductionOrders() {
	for i, order := range em.productionQueue {
		if i >= 2 { // Limit concurrent orders
			break
		}

		if time.Now().After(order.Deadline) {
			// Order is overdue, increase priority
			order.Priority *= 1.2
		}

		em.executeProductionOrder(order)
	}
}

// executeProductionOrder executes a specific production order
func (em *EconomicManager) executeProductionOrder(order ProductionOrder) {
	// Find appropriate producer
	producer := em.findProducer(order.Building, order.Type)
	if producer == nil {
		return
	}

	// Issue production command
	switch order.Type {
	case "worker":
		em.issueWorkerProductionCommand(producer, order)
	default:
		if em.isBuildingType(order.Type) {
			em.issueBuildingConstructionCommand(producer, order)
		}
	}
}

// prioritizeWorkerProduction increases worker production
func (em *EconomicManager) prioritizeWorkerProduction() {
	// Add urgent worker production orders
	for i := 0; i < 2; i++ {
		em.productionQueue = append(em.productionQueue, ProductionOrder{
			Type:     "worker",
			Priority: 0.95,
			Building: "main_building",
			Parameters: map[string]interface{}{
				"urgent": true,
			},
			Deadline: time.Now().Add(20 * time.Second),
		})
	}
}

// prioritizeResourceBuildings increases resource building construction
func (em *EconomicManager) prioritizeResourceBuildings() {
	// Add resource building orders for highest priority resources
	for resType, priority := range em.resourcePriorities {
		if priority > 0.7 {
			buildingType := em.getResourceBuildingType(resType)
			em.productionQueue = append(em.productionQueue, ProductionOrder{
				Type:     buildingType,
				Priority: 0.9,
				Building: "worker",
				Parameters: map[string]interface{}{
					"resource_type": resType,
					"urgent":        true,
				},
				Deadline: time.Now().Add(45 * time.Second),
			})
		}
	}
}

// prioritizeInfrastructure increases infrastructure development
func (em *EconomicManager) prioritizeInfrastructure() {
	// Add infrastructure building orders
	infrastructureTypes := []string{"house", "storage", "market"}

	for _, buildingType := range infrastructureTypes {
		em.productionQueue = append(em.productionQueue, ProductionOrder{
			Type:     buildingType,
			Priority: 0.8,
			Building: "worker",
			Parameters: map[string]interface{}{
				"infrastructure": true,
			},
			Deadline: time.Now().Add(60 * time.Second),
		})
	}
}

// identifyEconomicTargets finds expansion opportunities
func (em *EconomicManager) identifyEconomicTargets() {
	em.economicTargets = em.economicTargets[:0]

	// Find resource node opportunities
	for _, resource := range em.world.resources {
		if em.isResourceAccessible(resource) {
			target := EconomicTarget{
				Type:     "resource_node",
				Location: resource.Position,
				Value:    em.calculateResourceValue(resource),
				Risk:     em.calculateResourceRisk(resource),
				Distance: em.calculateDistanceFromBase(resource.Position),
			}
			target.Priority = target.Value * (1.0 - target.Risk) / (1.0 + target.Distance/10.0)
			em.economicTargets = append(em.economicTargets, target)
		}
	}

	// Sort by priority
	sort.Slice(em.economicTargets, func(i, j int) bool {
		return em.economicTargets[i].Priority > em.economicTargets[j].Priority
	})
}

// Helper methods for economic calculations

func (em *EconomicManager) estimateResourceConsumption(resourceType string) float64 {
	// Calculate estimated resource consumption per day
	// Based on current units, buildings, and production
	return 50.0 // Placeholder
}

func (em *EconomicManager) countWorkers() int {
	units := em.world.ObjectManager.GetUnitsForPlayer(em.playerID)
	count := 0
	for _, unit := range units {
		if unit.UnitType == "worker" && unit.IsAlive() {
			count++
		}
	}
	return count
}

func (em *EconomicManager) calculateOptimalWorkerCount() int {
	// Calculate optimal number of workers based on current situation
	state := em.strategicAI.GetStrategyState()
	baseWorkers := 8

	// Scale with game phase
	switch state.Phase {
	case PhaseEarlyGame:
		return baseWorkers
	case PhaseMidGame:
		return baseWorkers + 4
	case PhaseLateGame:
		return baseWorkers + 6
	case PhaseEndGame:
		return baseWorkers + 2
	}

	return baseWorkers
}

func (em *EconomicManager) findHighestPriorityResource() string {
	highestPriority := 0.0
	var highestResource string

	for resType, priority := range em.resourcePriorities {
		if priority > highestPriority {
			highestPriority = priority
			highestResource = resType
		}
	}

	return highestResource
}

func (em *EconomicManager) getResourceBuildingType(resourceType string) string {
	switch resourceType {
	case "gold":
		return "gold_mine"
	case "wood":
		return "lumber_mill"
	case "stone":
		return "stone_quarry"
	case "energy":
		return "energy_plant"
	default:
		return "resource_building"
	}
}

func (em *EconomicManager) countResourceBuildings(buildingType string) int {
	buildings := em.world.ObjectManager.GetBuildingsForPlayer(em.playerID)
	count := 0
	for _, building := range buildings {
		if building.BuildingType == buildingType && building.IsBuilt {
			count++
		}
	}
	return count
}

func (em *EconomicManager) calculateOptimalBuildingCount(buildingType string) int {
	// Calculate optimal number of a specific building type
	switch buildingType {
	case "gold_mine", "lumber_mill":
		return 3 // High priority resources
	case "stone_quarry", "energy_plant":
		return 2 // Medium priority resources
	default:
		return 1
	}
}

func (em *EconomicManager) getAvailableWorkers() []*GameUnit {
	units := em.world.ObjectManager.GetUnitsForPlayer(em.playerID)
	var workers []*GameUnit

	for _, unit := range units {
		if unit.UnitType == "worker" && unit.IsAlive() && unit.State == UnitStateIdle {
			workers = append(workers, unit)
		}
	}

	return workers
}

func (em *EconomicManager) calculateOptimalResourceWorkers(resourceType string) int {
	priority := em.resourcePriorities[resourceType]
	return int(priority * 4) // Max 4 workers per resource type
}

func (em *EconomicManager) assignWorkerToResource(worker *GameUnit, resourceType string) {
	// Find nearest resource node of this type and assign worker
	// This would interact with the behavior tree system
	// Implementation would set worker's blackboard to target specific resource
}

func (em *EconomicManager) findProducer(producerType, productType string) interface{} {
	// Find building or unit that can produce the requested item
	// Returns GameBuilding or GameUnit that can produce
	return nil // Placeholder
}

func (em *EconomicManager) issueWorkerProductionCommand(producer interface{}, order ProductionOrder) {
	// Issue command to produce a worker unit
	// Implementation would use command system
}

func (em *EconomicManager) issueBuildingConstructionCommand(producer interface{}, order ProductionOrder) {
	// Issue command to construct a building
	// Implementation would use command system
}

func (em *EconomicManager) isBuildingType(itemType string) bool {
	buildingTypes := []string{"house", "gold_mine", "lumber_mill", "stone_quarry", "energy_plant", "storage", "market"}
	for _, buildingType := range buildingTypes {
		if itemType == buildingType {
			return true
		}
	}
	return false
}

func (em *EconomicManager) isResourceAccessible(resource *ResourceNode) bool {
	// Check if resource is reachable and not controlled by enemy
	return resource.Amount > 0 // Simplified check
}

func (em *EconomicManager) calculateResourceValue(resource *ResourceNode) float64 {
	// Calculate economic value of a resource node
	baseValue := float64(resource.Amount) / 1000.0
	priority := em.resourcePriorities[resource.ResourceType]
	return baseValue * priority
}

func (em *EconomicManager) calculateResourceRisk(resource *ResourceNode) float64 {
	// Calculate risk of accessing this resource (enemy proximity, etc.)
	return 0.2 // Placeholder
}

func (em *EconomicManager) calculateDistanceFromBase(position Vector3) float64 {
	// Calculate distance from main base to position
	// Would find main building and calculate distance
	return 15.0 // Placeholder
}

// MilitaryManager handles AI military decisions and army management
type MilitaryManager struct {
	playerID       int           // Player ID this manager controls
	world          *World        // Game world reference
	strategicAI    *StrategicAI  // Parent strategic AI
	lastEvaluation time.Time     // Last military evaluation time

	// Military state tracking
	armyComposition    map[string]int     // Current army composition
	recruitmentPlan    []RecruitmentOrder // Planned unit recruitment
	militaryTargets    []MilitaryTarget   // Military targets and threats
	battleGroups      []BattleGroup      // Organized military units
	defensivePositions []Vector3          // Key defensive positions
}

// RecruitmentOrder represents planned military unit production
type RecruitmentOrder struct {
	UnitType    string    // Type of unit to recruit
	Quantity    int       // Number of units
	Priority    float64   // Priority score
	Purpose     string    // "offense", "defense", "patrol"
	Deadline    time.Time // When recruitment should complete
}

// MilitaryTarget represents a military objective or threat
type MilitaryTarget struct {
	Type        string    // "enemy_base", "enemy_army", "strategic_point"
	Location    Vector3   // World position
	ThreatLevel float64   // Threat assessment (0.0-1.0)
	Opportunity float64   // Attack opportunity (0.0-1.0)
	Priority    float64   // Combined priority score
	LastSeen    time.Time // When target was last observed
}

// BattleGroup represents an organized group of military units
type BattleGroup struct {
	ID          int         // Unique group ID
	Units       []*GameUnit // Units in this group
	Role        string      // "attack", "defend", "patrol", "scout"
	Target      *Vector3    // Current target location
	Formation   string      // Formation type
	Status      string      // "ready", "moving", "engaged", "regrouping"
}

// NewMilitaryManager creates a new military manager
func NewMilitaryManager(playerID int, world *World, strategicAI *StrategicAI) *MilitaryManager {
	return &MilitaryManager{
		playerID:           playerID,
		world:             world,
		strategicAI:       strategicAI,
		armyComposition:    make(map[string]int),
		recruitmentPlan:    make([]RecruitmentOrder, 0),
		militaryTargets:    make([]MilitaryTarget, 0),
		battleGroups:      make([]BattleGroup, 0),
		defensivePositions: make([]Vector3, 0),
	}
}

// Update performs military management decisions
func (mm *MilitaryManager) Update(deltaTime time.Duration) {
	// Evaluate military situation every 4 seconds
	if time.Since(mm.lastEvaluation) >= 4*time.Second {
		mm.evaluateMilitarySituation()
		mm.updateArmyComposition()
		mm.planRecruitment()
		mm.manageBattleGroups()
		mm.lastEvaluation = time.Now()
	}

	mm.executeMilitaryOrders()
}

// ExecuteMilitaryBuildup implements military focus decisions from strategic AI
func (mm *MilitaryManager) ExecuteMilitaryBuildup(parameters map[string]interface{}) {
	focus, exists := parameters["focus"]
	if !exists {
		return
	}

	switch focus {
	case "army":
		mm.prioritizeArmyExpansion()
	case "defense":
		mm.prioritizeDefensiveUnits()
	case "offense":
		mm.prioritizeOffensiveUnits()
	}
}

// ExecuteAttackStrategy implements attack decisions from strategic AI
func (mm *MilitaryManager) ExecuteAttackStrategy(parameters map[string]interface{}) {
	strategy, exists := parameters["strategy"]
	if !exists {
		return
	}

	switch strategy {
	case "offensive":
		mm.planOffensiveOperation()
	case "raid":
		mm.planRaidOperation()
	case "siege":
		mm.planSiegeOperation()
	}
}

// ExecuteDefensiveStrategy implements defensive decisions from strategic AI
func (mm *MilitaryManager) ExecuteDefensiveStrategy(parameters map[string]interface{}) {
	strategy, exists := parameters["strategy"]
	if !exists {
		return
	}

	switch strategy {
	case "defensive":
		mm.strengthenDefenses()
	case "fortify":
		mm.fortifyPositions()
	case "fallback":
		mm.executeFallbackPlan()
	}
}

// evaluateMilitarySituation analyzes current military status
func (mm *MilitaryManager) evaluateMilitarySituation() {
	// Update army composition
	mm.updateArmyComposition()

	// Identify military targets and threats
	mm.identifyMilitaryTargets()

	// Assess defensive positions
	mm.assessDefensivePositions()

	// Organize units into battle groups
	mm.organizeBattleGroups()
}

// updateArmyComposition counts current military units
func (mm *MilitaryManager) updateArmyComposition() {
	// Clear current composition
	for unitType := range mm.armyComposition {
		mm.armyComposition[unitType] = 0
	}

	// Count military units
	units := mm.world.ObjectManager.GetUnitsForPlayer(mm.playerID)
	for _, unit := range units {
		if unit.IsAlive() && mm.isMilitaryUnit(unit) {
			mm.armyComposition[unit.UnitType]++
		}
	}
}

// planRecruitment determines military recruitment needs
func (mm *MilitaryManager) planRecruitment() {
	mm.recruitmentPlan = mm.recruitmentPlan[:0]

	state := mm.strategicAI.GetStrategyState()
	// TODO: Use personality to influence recruitment decisions

	// Calculate optimal army composition
	optimalComposition := mm.calculateOptimalArmyComposition()

	// Plan recruitment to reach optimal composition
	for unitType, optimal := range optimalComposition {
		current := mm.armyComposition[unitType]
		if current < optimal {
			deficit := optimal - current
			priority := mm.calculateRecruitmentPriority(unitType, deficit)

			mm.recruitmentPlan = append(mm.recruitmentPlan, RecruitmentOrder{
				UnitType: unitType,
				Quantity: deficit,
				Priority: priority,
				Purpose:  mm.determineUnitPurpose(unitType),
				Deadline: time.Now().Add(time.Duration(deficit*45) * time.Second),
			})
		}
	}

	// Additional recruitment based on threat level
	if state.ThreatLevel > 0.6 {
		mm.addEmergencyRecruitment()
	}

	// Sort by priority
	sort.Slice(mm.recruitmentPlan, func(i, j int) bool {
		return mm.recruitmentPlan[i].Priority > mm.recruitmentPlan[j].Priority
	})
}

// manageBattleGroups organizes and commands military units
func (mm *MilitaryManager) manageBattleGroups() {
	// Update existing battle groups
	mm.updateBattleGroups()

	// Create new battle groups as needed
	mm.formNewBattleGroups()

	// Assign missions to battle groups
	mm.assignBattleGroupMissions()
}

// executeMilitaryOrders carries out military plans
func (mm *MilitaryManager) executeMilitaryOrders() {
	// Execute recruitment orders
	mm.executeRecruitmentOrders()

	// Execute battle group orders
	mm.executeBattleGroupOrders()
}

// Helper methods for military management

func (mm *MilitaryManager) isMilitaryUnit(unit *GameUnit) bool {
	militaryTypes := []string{"soldier", "warrior", "knight", "archer", "guard", "cavalry"}
	for _, militaryType := range militaryTypes {
		if unit.UnitType == militaryType {
			return true
		}
	}
	return false
}

func (mm *MilitaryManager) calculateOptimalArmyComposition() map[string]int {
	state := mm.strategicAI.GetStrategyState()
	composition := make(map[string]int)

	baseArmy := 8 // Base army size

	// Scale with military strength and game phase
	switch state.Phase {
	case PhaseEarlyGame:
		composition["soldier"] = baseArmy / 2
		composition["archer"] = baseArmy / 4
	case PhaseMidGame:
		composition["soldier"] = baseArmy
		composition["archer"] = baseArmy / 2
		composition["knight"] = baseArmy / 4
	case PhaseLateGame:
		composition["soldier"] = baseArmy
		composition["archer"] = baseArmy / 2
		composition["knight"] = baseArmy / 2
		composition["cavalry"] = baseArmy / 4
	case PhaseEndGame:
		composition["knight"] = baseArmy
		composition["cavalry"] = baseArmy / 2
		composition["archer"] = baseArmy / 2
	}

	return composition
}

func (mm *MilitaryManager) calculateRecruitmentPriority(unitType string, deficit int) float64 {
	basePriority := 0.6
	personality := mm.strategicAI.GetPersonality()
	state := mm.strategicAI.GetStrategyState()

	// Adjust for military focus
	basePriority *= personality.MilitaryFocus

	// Adjust for threat level
	basePriority += state.ThreatLevel * 0.3

	// Adjust for unit type importance
	switch unitType {
	case "soldier", "warrior":
		basePriority += 0.2 // Core units
	case "archer":
		basePriority += 0.1 // Support units
	case "knight", "cavalry":
		basePriority += 0.15 // Elite units
	}

	// Adjust for deficit severity
	basePriority += math.Min(float64(deficit)/5.0, 0.2)

	return math.Min(basePriority, 1.0)
}

func (mm *MilitaryManager) determineUnitPurpose(unitType string) string {
	personality := mm.strategicAI.GetPersonality()

	if personality.DefensivePosture > 0.7 {
		return "defense"
	} else if personality.AggressionLevel > 0.7 {
		return "offense"
	} else {
		return "patrol"
	}
}

// Placeholder implementations for complex military operations

func (mm *MilitaryManager) identifyMilitaryTargets() {
	// Identify enemy units, buildings, and strategic positions
	mm.militaryTargets = mm.militaryTargets[:0]
	// Implementation would scan for enemies and assess threats
}

func (mm *MilitaryManager) assessDefensivePositions() {
	// Evaluate key defensive positions around base
	mm.defensivePositions = mm.defensivePositions[:0]
	// Implementation would identify choke points and strategic locations
}

func (mm *MilitaryManager) organizeBattleGroups() {
	// Organize military units into coherent battle groups
	// Implementation would group nearby units with complementary roles
}

func (mm *MilitaryManager) updateBattleGroups() {
	// Update status of existing battle groups
	// Check if units are still alive, missions complete, etc.
}

func (mm *MilitaryManager) formNewBattleGroups() {
	// Create new battle groups from available units
	// Implementation would group unassigned military units
}

func (mm *MilitaryManager) assignBattleGroupMissions() {
	// Assign missions to battle groups based on strategic priorities
	// Implementation would assign attack, defense, or patrol missions
}

func (mm *MilitaryManager) executeRecruitmentOrders() {
	// Execute planned recruitment orders
	// Implementation would issue unit production commands
}

func (mm *MilitaryManager) executeBattleGroupOrders() {
	// Execute battle group missions
	// Implementation would issue movement and attack commands to groups
}

func (mm *MilitaryManager) prioritizeArmyExpansion() {
	// Add high-priority recruitment orders for army expansion
}

func (mm *MilitaryManager) prioritizeDefensiveUnits() {
	// Focus recruitment on defensive unit types
}

func (mm *MilitaryManager) prioritizeOffensiveUnits() {
	// Focus recruitment on offensive unit types
}

func (mm *MilitaryManager) planOffensiveOperation() {
	// Plan and execute offensive military operation
}

func (mm *MilitaryManager) planRaidOperation() {
	// Plan quick raid on enemy resources or positions
}

func (mm *MilitaryManager) planSiegeOperation() {
	// Plan sustained siege of enemy fortification
}

func (mm *MilitaryManager) strengthenDefenses() {
	// Reinforce defensive positions
}

func (mm *MilitaryManager) fortifyPositions() {
	// Build defensive structures and position units
}

func (mm *MilitaryManager) executeFallbackPlan() {
	// Execute strategic withdrawal and defensive consolidation
}

func (mm *MilitaryManager) addEmergencyRecruitment() {
	// Add emergency recruitment orders due to high threat
	mm.recruitmentPlan = append(mm.recruitmentPlan, RecruitmentOrder{
		UnitType: "soldier",
		Quantity: 3,
		Priority: 0.95,
		Purpose:  "emergency_defense",
		Deadline: time.Now().Add(30 * time.Second),
	})
}