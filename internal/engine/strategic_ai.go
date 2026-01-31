package engine

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"time"
)

// AIPersonality defines different AI behavior characteristics
type AIPersonality struct {
	Name                 string  // AI personality name
	Description          string  // Personality description
	AggressionLevel      float64 // 0.0-1.0, how aggressive the AI is
	EconomicFocus        float64 // 0.0-1.0, emphasis on economic development
	MilitaryFocus        float64 // 0.0-1.0, emphasis on military buildup
	ExpansionTendency    float64 // 0.0-1.0, likelihood to expand territory
	TechnologyPriority   float64 // 0.0-1.0, emphasis on tech advancement
	DefensivePosture     float64 // 0.0-1.0, defensive vs offensive stance
	RiskTolerance        float64 // 0.0-1.0, willingness to take risks
	AdaptabilityLevel    float64 // 0.0-1.0, ability to adapt strategies
}

// Predefined AI personalities
var (
	ConservativePersonality = AIPersonality{
		Name:                 "Conservative",
		Description:          "Focuses on economy and defense, slow to attack",
		AggressionLevel:      0.2,
		EconomicFocus:        0.8,
		MilitaryFocus:        0.4,
		ExpansionTendency:    0.3,
		TechnologyPriority:   0.7,
		DefensivePosture:     0.8,
		RiskTolerance:        0.2,
		AdaptabilityLevel:    0.5,
	}

	AggressivePersonality = AIPersonality{
		Name:                 "Aggressive",
		Description:          "Prioritizes military and early attacks",
		AggressionLevel:      0.9,
		EconomicFocus:        0.3,
		MilitaryFocus:        0.9,
		ExpansionTendency:    0.7,
		TechnologyPriority:   0.4,
		DefensivePosture:     0.2,
		RiskTolerance:        0.8,
		AdaptabilityLevel:    0.6,
	}

	BalancedPersonality = AIPersonality{
		Name:                 "Balanced",
		Description:          "Even focus on all aspects",
		AggressionLevel:      0.5,
		EconomicFocus:        0.6,
		MilitaryFocus:        0.6,
		ExpansionTendency:    0.5,
		TechnologyPriority:   0.6,
		DefensivePosture:     0.5,
		RiskTolerance:        0.5,
		AdaptabilityLevel:    0.7,
	}

	TechnologicalPersonality = AIPersonality{
		Name:                 "Technological",
		Description:          "Emphasizes research and advanced units",
		AggressionLevel:      0.3,
		EconomicFocus:        0.7,
		MilitaryFocus:        0.5,
		ExpansionTendency:    0.4,
		TechnologyPriority:   0.9,
		DefensivePosture:     0.6,
		RiskTolerance:        0.4,
		AdaptabilityLevel:    0.8,
	}

	ExpansionistPersonality = AIPersonality{
		Name:                 "Expansionist",
		Description:          "Focuses on rapid expansion and territory control",
		AggressionLevel:      0.6,
		EconomicFocus:        0.5,
		MilitaryFocus:        0.6,
		ExpansionTendency:    0.9,
		TechnologyPriority:   0.5,
		DefensivePosture:     0.4,
		RiskTolerance:        0.7,
		AdaptabilityLevel:    0.6,
	}
)

// StrategyState represents the current strategic situation
type StrategyState struct {
	Phase             StrategyPhase       // Current phase of the game
	EconomicStrength  float64            // Economic power assessment (0.0-1.0)
	MilitaryStrength  float64            // Military power assessment (0.0-1.0)
	ThreatLevel       float64            // Perceived threat level (0.0-1.0)
	ResourceSecurity  map[string]float64 // Security of each resource type (0.0-1.0)
	TerritoryControl  float64            // Amount of map controlled (0.0-1.0)
	TechLevel         float64            // Technology advancement (0.0-1.0)
	Population        int                // Current population
	MaxPopulation     int                // Population limit
	LastUpdate        time.Time          // When state was last updated
}

// StrategyPhase represents the current phase of the game
type StrategyPhase int

const (
	PhaseEarlyGame   StrategyPhase = iota // Early development phase
	PhaseMidGame                          // Expansion and conflict phase
	PhaseLateGame                         // Technology and decisive battles
	PhaseEndGame                          // Final push or cleanup
)

// String returns the string representation of StrategyPhase
func (sp StrategyPhase) String() string {
	switch sp {
	case PhaseEarlyGame:
		return "Early Game"
	case PhaseMidGame:
		return "Mid Game"
	case PhaseLateGame:
		return "Late Game"
	case PhaseEndGame:
		return "End Game"
	default:
		return "Unknown"
	}
}

// StrategicDecision represents a high-level strategic decision
type StrategicDecision struct {
	Type        DecisionType               // Type of decision
	Priority    float64                    // Priority score (0.0-1.0)
	Parameters  map[string]interface{}     // Decision parameters
	Rationale   string                     // Explanation for the decision
	ExpectedOutcome string                 // Expected result
	Confidence  float64                    // Confidence in decision (0.0-1.0)
	Timestamp   time.Time                  // When decision was made
}

// DecisionType represents different types of strategic decisions
type DecisionType int

const (
	DecisionExpand          DecisionType = iota // Expand to new territory
	DecisionBuildEconomy                        // Focus on economic development
	DecisionBuildMilitary                       // Build up military forces
	DecisionAttack                              // Launch attack on enemy
	DecisionDefend                              // Strengthen defenses
	DecisionResearch                            // Focus on technology
	DecisionRetreat                             // Fall back and consolidate
	DecisionAlliance                            // Seek diplomatic relations
	DecisionScout                               // Gather intelligence
	DecisionUpgrade                             // Upgrade existing units/buildings
)

// String returns the string representation of DecisionType
func (dt DecisionType) String() string {
	switch dt {
	case DecisionExpand:
		return "Expand"
	case DecisionBuildEconomy:
		return "Build Economy"
	case DecisionBuildMilitary:
		return "Build Military"
	case DecisionAttack:
		return "Attack"
	case DecisionDefend:
		return "Defend"
	case DecisionResearch:
		return "Research"
	case DecisionRetreat:
		return "Retreat"
	case DecisionAlliance:
		return "Alliance"
	case DecisionScout:
		return "Scout"
	case DecisionUpgrade:
		return "Upgrade"
	default:
		return "Unknown"
	}
}

// StrategicAI represents the main AI strategic decision-making system
type StrategicAI struct {
	playerID         int                    // Player ID this AI controls
	world           *World                 // Game world reference
	personality     AIPersonality          // AI personality profile
	difficulty      AIDifficulty           // AI difficulty level
	state           StrategyState          // Current strategic state
	decisions       []StrategicDecision    // Recent decisions made
	economicMgr     *EconomicManager       // Economic decision manager
	militaryMgr     *MilitaryManager       // Military strategy manager
	lastUpdateTime  time.Time              // Last AI update time
	updateInterval  time.Duration          // How often to make decisions
	random          *rand.Rand             // Random number generator for decisions
}

// AIDifficulty represents different AI skill levels
type AIDifficulty int

const (
	DifficultyEasy   AIDifficulty = iota // Easy AI - makes suboptimal decisions
	DifficultyNormal                     // Normal AI - balanced play
	DifficultyHard                       // Hard AI - optimized decisions
	DifficultyExpert                     // Expert AI - advanced strategies
)

// String returns the string representation of AIDifficulty
func (diff AIDifficulty) String() string {
	switch diff {
	case DifficultyEasy:
		return "Easy"
	case DifficultyNormal:
		return "Normal"
	case DifficultyHard:
		return "Hard"
	case DifficultyExpert:
		return "Expert"
	default:
		return "Unknown"
	}
}

// NewStrategicAI creates a new strategic AI instance
func NewStrategicAI(playerID int, world *World, personality AIPersonality, difficulty AIDifficulty) *StrategicAI {
	ai := &StrategicAI{
		playerID:       playerID,
		world:          world,
		personality:    personality,
		difficulty:     difficulty,
		decisions:      make([]StrategicDecision, 0),
		updateInterval: 5 * time.Second, // Update every 5 seconds
		random:         rand.New(rand.NewSource(time.Now().UnixNano() + int64(playerID))),
	}

	// Initialize sub-managers
	ai.economicMgr = NewEconomicManager(playerID, world, ai)
	ai.militaryMgr = NewMilitaryManager(playerID, world, ai)

	// Initialize strategy state
	ai.state = StrategyState{
		Phase:            PhaseEarlyGame,
		EconomicStrength: 0.1,
		MilitaryStrength: 0.1,
		ThreatLevel:      0.0,
		ResourceSecurity: make(map[string]float64),
		TerritoryControl: 0.1,
		TechLevel:        0.0,
		LastUpdate:       time.Now(),
	}

	// Initialize resource security
	resourceTypes := []string{"gold", "wood", "stone", "energy"}
	for _, resType := range resourceTypes {
		ai.state.ResourceSecurity[resType] = 0.5 // Start with medium security
	}

	return ai
}

// Update performs strategic AI decision-making
func (ai *StrategicAI) Update(deltaTime time.Duration) {
	// Check if it's time for a strategic update
	if time.Since(ai.lastUpdateTime) < ai.updateInterval {
		return
	}

	// Update strategic state assessment
	ai.updateStrategyState()

	// Generate potential decisions
	decisions := ai.generateDecisions()

	// Evaluate and select best decision
	if len(decisions) > 0 {
		bestDecision := ai.selectBestDecision(decisions)
		ai.executeDecision(bestDecision)
		ai.decisions = append(ai.decisions, bestDecision)

		// Keep only recent decisions (last 10)
		if len(ai.decisions) > 10 {
			ai.decisions = ai.decisions[1:]
		}
	}

	// Update sub-managers
	ai.economicMgr.Update(deltaTime)
	ai.militaryMgr.Update(deltaTime)

	ai.lastUpdateTime = time.Now()
}

// updateStrategyState analyzes current game situation and updates strategy state
func (ai *StrategicAI) updateStrategyState() {
	player := ai.world.GetPlayer(ai.playerID)
	if player == nil {
		return
	}

	// Update game phase based on time and development
	gameTime := ai.world.GetGameTime()
	if gameTime > 20*time.Minute {
		ai.state.Phase = PhaseEndGame
	} else if gameTime > 12*time.Minute {
		ai.state.Phase = PhaseLateGame
	} else if gameTime > 5*time.Minute {
		ai.state.Phase = PhaseMidGame
	} else {
		ai.state.Phase = PhaseEarlyGame
	}

	// Assess economic strength
	ai.state.EconomicStrength = ai.assessEconomicStrength()

	// Assess military strength
	ai.state.MilitaryStrength = ai.assessMilitaryStrength()

	// Assess threat level
	ai.state.ThreatLevel = ai.assessThreatLevel()

	// Assess resource security
	ai.assessResourceSecurity()

	// Assess territory control
	ai.state.TerritoryControl = ai.assessTerritoryControl()

	// Update population info
	ai.state.Population = ai.getPopulationCount()
	ai.state.MaxPopulation = ai.getMaxPopulation()

	ai.state.LastUpdate = time.Now()
}

// assessEconomicStrength evaluates the player's economic power
func (ai *StrategicAI) assessEconomicStrength() float64 {
	player := ai.world.GetPlayer(ai.playerID)
	if player == nil {
		return 0.0
	}

	// Calculate resource wealth
	totalResources := 0
	for _, amount := range player.Resources {
		totalResources += amount
	}

	// Number of workers and economic buildings
	workers := ai.countUnitsByType("worker")
	economicBuildings := ai.countEconomicBuildings()

	// Resource generation rate
	resourceIncome := ai.calculateResourceIncome()

	// Normalize to 0.0-1.0 scale (rough estimates)
	resourceScore := math.Min(float64(totalResources)/5000.0, 1.0)
	workerScore := math.Min(float64(workers)/20.0, 1.0)
	buildingScore := math.Min(float64(economicBuildings)/10.0, 1.0)
	incomeScore := math.Min(resourceIncome/100.0, 1.0)

	// Weighted combination
	return 0.3*resourceScore + 0.3*workerScore + 0.2*buildingScore + 0.2*incomeScore
}

// assessMilitaryStrength evaluates the player's military power
func (ai *StrategicAI) assessMilitaryStrength() float64 {
	// Count military units
	soldiers := ai.countMilitaryUnits()

	// Count military buildings
	militaryBuildings := ai.countMilitaryBuildings()

	// Assess unit quality and upgrades
	unitQuality := ai.assessUnitQuality()

	// Normalize to 0.0-1.0 scale
	soldierScore := math.Min(float64(soldiers)/30.0, 1.0)
	buildingScore := math.Min(float64(militaryBuildings)/5.0, 1.0)

	// Weighted combination
	return 0.5*soldierScore + 0.3*buildingScore + 0.2*unitQuality
}

// assessThreatLevel evaluates threats from enemies
func (ai *StrategicAI) assessThreatLevel() float64 {
	// Find enemy military units near our territory
	nearbyEnemies := ai.countNearbyEnemyUnits()

	// Assess enemy economic and military strength
	enemyStrength := ai.assessEnemyStrength()

	// Recent attacks or aggressive actions
	recentThreat := ai.assessRecentThreats()

	// Normalize and combine
	proximityThreat := math.Min(float64(nearbyEnemies)/20.0, 1.0)
	strengthThreat := enemyStrength

	return math.Max(proximityThreat, math.Max(strengthThreat, recentThreat))
}

// assessResourceSecurity evaluates security of resource access
func (ai *StrategicAI) assessResourceSecurity() {
	resourceTypes := []string{"gold", "wood", "stone", "energy"}

	for _, resType := range resourceTypes {
		// Count controlled resource nodes
		controlledNodes := ai.countControlledResourceNodes(resType)

		// Distance to resource sources
		avgDistance := ai.averageDistanceToResources(resType)

		// Threat to resource gathering
		resourceThreat := ai.assessResourceThreat(resType)

		// Calculate security score
		nodeScore := math.Min(float64(controlledNodes)/5.0, 1.0)
		distanceScore := math.Max(0.0, 1.0-avgDistance/20.0)
		threatScore := 1.0 - resourceThreat

		ai.state.ResourceSecurity[resType] = 0.4*nodeScore + 0.3*distanceScore + 0.3*threatScore
	}
}

// assessTerritoryControl evaluates amount of map controlled
func (ai *StrategicAI) assessTerritoryControl() float64 {
	// Count buildings and units spread across map
	buildings := ai.countPlayerBuildings()
	units := ai.countPlayerUnits()

	// Assess map presence and control points
	mapCoverage := ai.calculateMapCoverage()

	buildingScore := math.Min(float64(buildings)/15.0, 1.0)
	unitScore := math.Min(float64(units)/50.0, 1.0)

	return 0.4*buildingScore + 0.3*unitScore + 0.3*mapCoverage
}

// generateDecisions creates potential strategic decisions
func (ai *StrategicAI) generateDecisions() []StrategicDecision {
	var decisions []StrategicDecision

	// Generate decisions based on current state and personality
	decisions = append(decisions, ai.generateEconomicDecisions()...)
	decisions = append(decisions, ai.generateMilitaryDecisions()...)
	decisions = append(decisions, ai.generateExpansionDecisions()...)
	decisions = append(decisions, ai.generateTechnologyDecisions()...)
	decisions = append(decisions, ai.generateDefensiveDecisions()...)

	return decisions
}

// generateEconomicDecisions creates economy-focused decisions
func (ai *StrategicAI) generateEconomicDecisions() []StrategicDecision {
	var decisions []StrategicDecision

	// Build more workers if economy is weak
	if ai.state.EconomicStrength < 0.5 {
		priority := (1.0 - ai.state.EconomicStrength) * ai.personality.EconomicFocus
		decisions = append(decisions, StrategicDecision{
			Type:            DecisionBuildEconomy,
			Priority:        priority,
			Parameters:      map[string]interface{}{"focus": "workers"},
			Rationale:       "Economic strength is low, need more workers",
			ExpectedOutcome: "Increased resource generation",
			Confidence:      0.8,
			Timestamp:       time.Now(),
		})
	}

	// Expand resource gathering if resource security is low
	for resType, security := range ai.state.ResourceSecurity {
		if security < 0.4 {
			priority := (1.0 - security) * ai.personality.EconomicFocus * 0.7
			decisions = append(decisions, StrategicDecision{
				Type:       DecisionExpand,
				Priority:   priority,
				Parameters: map[string]interface{}{"target": "resources", "resource_type": resType},
				Rationale:  fmt.Sprintf("Low %s security, need better resource access", resType),
				ExpectedOutcome: "Improved resource security",
				Confidence: 0.6,
				Timestamp:  time.Now(),
			})
		}
	}

	return decisions
}

// generateMilitaryDecisions creates military-focused decisions
func (ai *StrategicAI) generateMilitaryDecisions() []StrategicDecision {
	var decisions []StrategicDecision

	// Build military if threat level is high or military focus is high
	if ai.state.ThreatLevel > 0.3 || (ai.personality.MilitaryFocus > 0.6 && ai.state.MilitaryStrength < 0.6) {
		priority := math.Max(ai.state.ThreatLevel, ai.personality.MilitaryFocus) * 0.8
		decisions = append(decisions, StrategicDecision{
			Type:            DecisionBuildMilitary,
			Priority:        priority,
			Parameters:      map[string]interface{}{"focus": "army"},
			Rationale:       "Need stronger military for defense/offense",
			ExpectedOutcome: "Increased military strength",
			Confidence:      0.7,
			Timestamp:       time.Now(),
		})
	}

	// Attack if aggressive and military is strong
	if ai.personality.AggressionLevel > 0.6 && ai.state.MilitaryStrength > 0.5 && ai.state.ThreatLevel < 0.3 {
		priority := ai.personality.AggressionLevel * ai.state.MilitaryStrength
		decisions = append(decisions, StrategicDecision{
			Type:            DecisionAttack,
			Priority:        priority,
			Parameters:      map[string]interface{}{"strategy": "offensive"},
			Rationale:       "Strong military and aggressive personality",
			ExpectedOutcome: "Territorial gains and enemy weakening",
			Confidence:      0.6,
			Timestamp:       time.Now(),
		})
	}

	return decisions
}

// generateExpansionDecisions creates expansion-focused decisions
func (ai *StrategicAI) generateExpansionDecisions() []StrategicDecision {
	var decisions []StrategicDecision

	// Expand if expansion personality and low territory control
	if ai.personality.ExpansionTendency > 0.5 && ai.state.TerritoryControl < 0.4 {
		priority := ai.personality.ExpansionTendency * (1.0 - ai.state.TerritoryControl)
		decisions = append(decisions, StrategicDecision{
			Type:            DecisionExpand,
			Priority:        priority,
			Parameters:      map[string]interface{}{"target": "territory"},
			Rationale:       "Low territory control, opportunity for expansion",
			ExpectedOutcome: "Increased map control and resources",
			Confidence:      0.7,
			Timestamp:       time.Now(),
		})
	}

	return decisions
}

// generateTechnologyDecisions creates research-focused decisions
func (ai *StrategicAI) generateTechnologyDecisions() []StrategicDecision {
	var decisions []StrategicDecision

	// Focus on research if technology personality and sufficient economy
	if ai.personality.TechnologyPriority > 0.6 && ai.state.EconomicStrength > 0.4 {
		priority := ai.personality.TechnologyPriority * ai.state.EconomicStrength
		decisions = append(decisions, StrategicDecision{
			Type:            DecisionResearch,
			Priority:        priority,
			Parameters:      map[string]interface{}{"focus": "upgrades"},
			Rationale:       "Technology focus with sufficient economy",
			ExpectedOutcome: "Improved unit capabilities",
			Confidence:      0.8,
			Timestamp:       time.Now(),
		})
	}

	return decisions
}

// generateDefensiveDecisions creates defense-focused decisions
func (ai *StrategicAI) generateDefensiveDecisions() []StrategicDecision {
	var decisions []StrategicDecision

	// Defend if high threat and defensive personality
	if ai.state.ThreatLevel > 0.5 || (ai.personality.DefensivePosture > 0.6 && ai.state.MilitaryStrength < 0.5) {
		priority := math.Max(ai.state.ThreatLevel, ai.personality.DefensivePosture) * 0.9
		decisions = append(decisions, StrategicDecision{
			Type:            DecisionDefend,
			Priority:        priority,
			Parameters:      map[string]interface{}{"strategy": "defensive"},
			Rationale:       "High threat level requires defensive measures",
			ExpectedOutcome: "Improved base security",
			Confidence:      0.8,
			Timestamp:       time.Now(),
		})
	}

	return decisions
}

// selectBestDecision chooses the optimal decision from candidates
func (ai *StrategicAI) selectBestDecision(decisions []StrategicDecision) StrategicDecision {
	if len(decisions) == 0 {
		// Default decision if no candidates
		return StrategicDecision{
			Type:            DecisionScout,
			Priority:        0.3,
			Parameters:      map[string]interface{}{},
			Rationale:       "No clear strategic priority, scouting for information",
			ExpectedOutcome: "Better situational awareness",
			Confidence:      0.5,
			Timestamp:       time.Now(),
		}
	}

	// Sort decisions by priority
	sort.Slice(decisions, func(i, j int) bool {
		return decisions[i].Priority > decisions[j].Priority
	})

	// Apply difficulty modifiers
	bestDecision := decisions[0]
	bestDecision.Priority = ai.applyDifficultyModifier(bestDecision.Priority)

	// Add some randomization based on AI personality
	randomFactor := ai.random.Float64() * (1.0 - ai.personality.AdaptabilityLevel) * 0.2
	if randomFactor > 0.1 && len(decisions) > 1 {
		// Sometimes pick second-best decision for unpredictability
		bestDecision = decisions[1]
	}

	return bestDecision
}

// executeDecision implements the chosen strategic decision
func (ai *StrategicAI) executeDecision(decision StrategicDecision) {
	// Log decision for debugging
	ai.logDecision(decision)

	switch decision.Type {
	case DecisionBuildEconomy:
		ai.economicMgr.ExecuteEconomicFocus(decision.Parameters)
	case DecisionBuildMilitary:
		ai.militaryMgr.ExecuteMilitaryBuildup(decision.Parameters)
	case DecisionAttack:
		ai.militaryMgr.ExecuteAttackStrategy(decision.Parameters)
	case DecisionDefend:
		ai.militaryMgr.ExecuteDefensiveStrategy(decision.Parameters)
	case DecisionExpand:
		ai.executeExpansionStrategy(decision.Parameters)
	case DecisionResearch:
		ai.executeResearchStrategy(decision.Parameters)
	case DecisionScout:
		ai.executeScoutingStrategy(decision.Parameters)
	}
}

// logDecision logs strategic decisions for analysis and debugging
func (ai *StrategicAI) logDecision(decision StrategicDecision) {
	// In a real implementation, this might log to file or telemetry
	// For now, it's a placeholder for decision tracking
}

// executeExpansionStrategy implements expansion decisions
func (ai *StrategicAI) executeExpansionStrategy(params map[string]interface{}) {
	// Find good expansion locations
	expansionSites := ai.findExpansionSites()

	if len(expansionSites) > 0 {
		// Order construction of expansion base
		ai.orderExpansionBase(expansionSites[0])
	}

	// Also order worker production to support expansion
	ai.orderWorkerProduction()
}

// executeResearchStrategy implements research decisions
func (ai *StrategicAI) executeResearchStrategy(params map[string]interface{}) {
	// Identify priority upgrades
	upgrades := ai.identifyPriorityUpgrades()

	if len(upgrades) > 0 {
		// Order the most important upgrade
		ai.orderUpgrade(upgrades[0])
	}
}

// executeScoutingStrategy implements scouting decisions
func (ai *StrategicAI) executeScoutingStrategy(params map[string]interface{}) {
	// Find unexplored areas
	scoutTargets := ai.findScoutTargets()

	if len(scoutTargets) > 0 {
		// Send scouts to explore
		ai.orderScouting(scoutTargets)
	}
}

// applyDifficultyModifier adjusts decision quality based on AI difficulty
func (ai *StrategicAI) applyDifficultyModifier(priority float64) float64 {
	switch ai.difficulty {
	case DifficultyEasy:
		return priority * 0.7 // Make suboptimal decisions
	case DifficultyNormal:
		return priority * 0.85
	case DifficultyHard:
		return priority * 1.0
	case DifficultyExpert:
		return priority * 1.15 // Enhanced decision-making
	default:
		return priority
	}
}

// Helper methods for state assessment

func (ai *StrategicAI) countUnitsByType(unitType string) int {
	units := ai.world.ObjectManager.GetUnitsForPlayer(ai.playerID)
	count := 0
	for _, unit := range units {
		if unit.UnitType == unitType && unit.IsAlive() {
			count++
		}
	}
	return count
}

func (ai *StrategicAI) countMilitaryUnits() int {
	units := ai.world.ObjectManager.GetUnitsForPlayer(ai.playerID)
	count := 0
	militaryTypes := []string{"soldier", "warrior", "knight", "archer", "guard"}

	for _, unit := range units {
		if unit.IsAlive() {
			for _, militaryType := range militaryTypes {
				if unit.UnitType == militaryType {
					count++
					break
				}
			}
		}
	}
	return count
}

func (ai *StrategicAI) countEconomicBuildings() int {
	// This would count resource-generating buildings
	// Implementation depends on building system
	return 3 // Placeholder
}

func (ai *StrategicAI) countMilitaryBuildings() int {
	// This would count military production buildings
	// Implementation depends on building system
	return 1 // Placeholder
}

func (ai *StrategicAI) calculateResourceIncome() float64 {
	// Calculate per-second resource generation
	// Implementation depends on economic system
	return 10.0 // Placeholder
}

func (ai *StrategicAI) assessUnitQuality() float64 {
	// Assess average unit upgrade level and effectiveness
	return 0.5 // Placeholder
}

func (ai *StrategicAI) countNearbyEnemyUnits() int {
	// Count enemy units within threatening range
	return 0 // Placeholder
}

func (ai *StrategicAI) assessEnemyStrength() float64 {
	// Assess relative strength of all enemies
	return 0.3 // Placeholder
}

func (ai *StrategicAI) assessRecentThreats() float64 {
	// Check for recent attacks or aggressive actions
	return 0.0 // Placeholder
}

func (ai *StrategicAI) countControlledResourceNodes(resourceType string) int {
	// Count resource nodes under our control
	return 2 // Placeholder
}

func (ai *StrategicAI) averageDistanceToResources(resourceType string) float64 {
	// Average distance from base to resource nodes
	return 10.0 // Placeholder
}

func (ai *StrategicAI) assessResourceThreat(resourceType string) float64 {
	// Assess threat to resource gathering operations
	return 0.2 // Placeholder
}

func (ai *StrategicAI) countPlayerBuildings() int {
	buildings := ai.world.ObjectManager.GetBuildingsForPlayer(ai.playerID)
	return len(buildings)
}

func (ai *StrategicAI) countPlayerUnits() int {
	units := ai.world.ObjectManager.GetUnitsForPlayer(ai.playerID)
	aliveCount := 0
	for _, unit := range units {
		if unit.IsAlive() {
			aliveCount++
		}
	}
	return aliveCount
}

func (ai *StrategicAI) calculateMapCoverage() float64 {
	// Calculate what percentage of map we have presence on
	return 0.3 // Placeholder
}

func (ai *StrategicAI) getPopulationCount() int {
	return ai.countPlayerUnits()
}

func (ai *StrategicAI) getMaxPopulation() int {
	// Calculate current population limit
	return 50 // Placeholder
}

// Stub implementations for expansion/research/scouting

func (ai *StrategicAI) findExpansionSites() []Vector3 {
	// Find good locations for expansion bases
	return []Vector3{{X: 20, Y: 0, Z: 20}} // Placeholder
}

func (ai *StrategicAI) orderExpansionBase(location Vector3) {
	// Find a worker to send for construction
	workers := ai.findAvailableWorkers()
	if len(workers) == 0 {
		return
	}

	worker := workers[0]

	// Create build command for expansion building (e.g., town center or barracks)
	expansionBuildingType := "castle" // Default expansion building
	if ai.personality.EconomicFocus > 0.6 {
		expansionBuildingType = "town_center"
	}

	command := UnitCommand{
		Type:     CommandBuild,
		Target:   &location,
		Parameters: map[string]interface{}{
			"building_type": expansionBuildingType,
		},
	}

	// Issue the build command
	ai.world.commandProcessor.IssueCommand(worker.ID, command)
}

// orderWorkerProduction orders production of additional workers
func (ai *StrategicAI) orderWorkerProduction() {
	// Find production buildings capable of creating workers
	buildings := ai.world.ObjectManager.GetBuildingsForPlayer(ai.playerID)

	for _, building := range buildings {
		if ai.canProduceWorkers(building) && building.CurrentProduction == nil {
			// Order worker production
			command := UnitCommand{
				Type: CommandProduce,
				Parameters: map[string]interface{}{
					"unit_type": "worker",
				},
			}

			// Issue command to building (buildings can receive commands too)
			ai.world.commandProcessor.IssueCommand(building.ID, command)
			break // Only order one worker at a time
		}
	}
}

// canProduceWorkers checks if a building can produce workers
func (ai *StrategicAI) canProduceWorkers(building *GameBuilding) bool {
	workerProducingBuildings := []string{"castle", "town_center", "barracks"}
	for _, buildingType := range workerProducingBuildings {
		if building.BuildingType == buildingType {
			return true
		}
	}
	return false
}

// findAvailableWorkers finds idle workers for task assignment
func (ai *StrategicAI) findAvailableWorkers() []*GameUnit {
	units := ai.world.ObjectManager.GetUnitsForPlayer(ai.playerID)
	var workers []*GameUnit

	for _, unit := range units {
		if unit.UnitType == "worker" && unit.State == UnitStateIdle && unit.IsAlive() {
			workers = append(workers, unit)
		}
	}

	return workers
}

func (ai *StrategicAI) identifyPriorityUpgrades() []string {
	// Identify most important upgrades to research
	return []string{"weapon_upgrade"} // Placeholder
}

func (ai *StrategicAI) orderUpgrade(upgradeType string) {
	// Order a specific upgrade
	// Implementation would interact with research system
}

func (ai *StrategicAI) findScoutTargets() []Vector3 {
	// Find unexplored areas to scout
	return []Vector3{{X: 30, Y: 0, Z: 30}} // Placeholder
}

func (ai *StrategicAI) orderScouting(targets []Vector3) {
	// Find available scout units (fast, light units)
	scouts := ai.findAvailableScouts()

	// Assign each scout to a target
	for i, scout := range scouts {
		if i >= len(targets) {
			break // More scouts than targets
		}

		target := targets[i]
		command := UnitCommand{
			Type:   CommandMove,
			Target: &target,
			Parameters: map[string]interface{}{
				"mission_type": "scout",
			},
		}

		ai.world.commandProcessor.IssueCommand(scout.ID, command)
	}
}

// findAvailableScouts finds units suitable for scouting
func (ai *StrategicAI) findAvailableScouts() []*GameUnit {
	units := ai.world.ObjectManager.GetUnitsForPlayer(ai.playerID)
	var scouts []*GameUnit

	scoutTypes := []string{"worker", "scout", "horseman", "archer"} // Fast units suitable for scouting

	for _, unit := range units {
		if unit.State == UnitStateIdle && unit.IsAlive() {
			for _, scoutType := range scoutTypes {
				if unit.UnitType == scoutType {
					scouts = append(scouts, unit)
					break
				}
			}
		}
	}

	return scouts
}

// GetPersonality returns the AI's personality profile
func (ai *StrategicAI) GetPersonality() AIPersonality {
	return ai.personality
}

// GetStrategyState returns the current strategic state
func (ai *StrategicAI) GetStrategyState() StrategyState {
	return ai.state
}

// GetRecentDecisions returns recent strategic decisions made
func (ai *StrategicAI) GetRecentDecisions() []StrategicDecision {
	return ai.decisions
}

// SetDifficulty changes the AI difficulty level
func (ai *StrategicAI) SetDifficulty(difficulty AIDifficulty) {
	ai.difficulty = difficulty
}

// SetPersonality changes the AI personality
func (ai *StrategicAI) SetPersonality(personality AIPersonality) {
	ai.personality = personality
}

// StrategicAIManager coordinates all AI players in the game
type StrategicAIManager struct {
	world       *World                    // Game world reference
	aiPlayers   map[int]*StrategicAI     // AI instance for each AI player
	updateTimer time.Duration            // Time since last update
	updateRate  time.Duration            // How often to update AI (reduces CPU load)
}

// NewStrategicAIManager creates a new strategic AI manager
func NewStrategicAIManager(world *World) *StrategicAIManager {
	return &StrategicAIManager{
		world:       world,
		aiPlayers:   make(map[int]*StrategicAI),
		updateRate:  time.Millisecond * 500, // Update AI every 500ms
	}
}

// InitializeAIPlayer creates and initializes AI for a player
func (mgr *StrategicAIManager) InitializeAIPlayer(playerID int, personality AIPersonality, difficulty AIDifficulty) error {
	if mgr.world == nil {
		return fmt.Errorf("world reference is nil")
	}

	// Check if player exists
	player := mgr.world.GetPlayer(playerID)
	if player == nil {
		return fmt.Errorf("player with ID %d not found", playerID)
	}

	if !player.IsAI {
		return fmt.Errorf("player %d is not marked as AI player", playerID)
	}

	// Create strategic AI instance
	ai := NewStrategicAI(playerID, mgr.world, personality, difficulty)
	mgr.aiPlayers[playerID] = ai

	return nil
}

// Update updates all AI players
func (mgr *StrategicAIManager) Update(deltaTime time.Duration) {
	mgr.updateTimer += deltaTime

	// Only update AI at reduced frequency to save CPU
	if mgr.updateTimer < mgr.updateRate {
		return
	}

	// Reset timer
	mgr.updateTimer = time.Duration(0)

	// Update each AI player
	for playerID, ai := range mgr.aiPlayers {
		// Check if player is still active
		player := mgr.world.GetPlayer(playerID)
		if player == nil || !player.IsActive {
			continue
		}

		// Update this AI player
		ai.Update(mgr.updateRate)
	}
}

// RemoveAIPlayer removes AI control for a player (e.g., when player is defeated)
func (mgr *StrategicAIManager) RemoveAIPlayer(playerID int) {
	delete(mgr.aiPlayers, playerID)
}

// GetAIPlayer returns the AI instance for a player
func (mgr *StrategicAIManager) GetAIPlayer(playerID int) *StrategicAI {
	return mgr.aiPlayers[playerID]
}

// GetAIPlayerCount returns number of active AI players
func (mgr *StrategicAIManager) GetAIPlayerCount() int {
	return len(mgr.aiPlayers)
}

// SetUpdateRate changes how frequently AI players are updated
func (mgr *StrategicAIManager) SetUpdateRate(rate time.Duration) {
	mgr.updateRate = rate
}