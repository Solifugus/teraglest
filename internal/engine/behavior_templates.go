package engine

import (
	"fmt"
	"time"
)

// BehaviorTreeTemplate represents a template for creating behavior trees
type BehaviorTreeTemplate struct {
	Name        string                                    // Template name
	Description string                                    // Template description
	Builder     func() BehaviorNode                      // Function to build the behavior tree
	UnitTypes   []string                                 // Unit types this template applies to
}

// BehaviorTreeLibrary manages predefined behavior tree templates
type BehaviorTreeLibrary struct {
	templates map[string]*BehaviorTreeTemplate // Template registry
}

// NewBehaviorTreeLibrary creates a new behavior tree library with default templates
func NewBehaviorTreeLibrary() *BehaviorTreeLibrary {
	library := &BehaviorTreeLibrary{
		templates: make(map[string]*BehaviorTreeTemplate),
	}

	// Register default templates
	library.registerDefaultTemplates()
	return library
}

// RegisterTemplate adds a new behavior tree template
func (btl *BehaviorTreeLibrary) RegisterTemplate(template *BehaviorTreeTemplate) {
	btl.templates[template.Name] = template
}

// GetTemplate retrieves a template by name
func (btl *BehaviorTreeLibrary) GetTemplate(name string) (*BehaviorTreeTemplate, bool) {
	template, exists := btl.templates[name]
	return template, exists
}

// GetTemplatesForUnitType returns all templates applicable to a unit type
func (btl *BehaviorTreeLibrary) GetTemplatesForUnitType(unitType string) []*BehaviorTreeTemplate {
	var matches []*BehaviorTreeTemplate

	for _, template := range btl.templates {
		for _, applicableType := range template.UnitTypes {
			if applicableType == unitType || applicableType == "*" {
				matches = append(matches, template)
				break
			}
		}
	}

	return matches
}

// CreateBehaviorTree creates a new behavior tree from a template
func (btl *BehaviorTreeLibrary) CreateBehaviorTree(templateName string) (*BehaviorTree, error) {
	template, exists := btl.templates[templateName]
	if !exists {
		return nil, fmt.Errorf("template %s not found", templateName)
	}

	root := template.Builder()
	return NewBehaviorTree(root), nil
}

// GetAllTemplateNames returns names of all registered templates
func (btl *BehaviorTreeLibrary) GetAllTemplateNames() []string {
	var names []string
	for name := range btl.templates {
		names = append(names, name)
	}
	return names
}

// registerDefaultTemplates registers the built-in behavior tree templates
func (btl *BehaviorTreeLibrary) registerDefaultTemplates() {
	// Worker/Villager AI
	btl.RegisterTemplate(&BehaviorTreeTemplate{
		Name:        "worker_ai",
		Description: "Basic worker AI that gathers resources when idle",
		UnitTypes:   []string{"worker", "villager", "peasant"},
		Builder:     buildWorkerAI,
	})

	// Soldier AI
	btl.RegisterTemplate(&BehaviorTreeTemplate{
		Name:        "soldier_ai",
		Description: "Basic soldier AI that attacks enemies and defends",
		UnitTypes:   []string{"soldier", "warrior", "knight", "archer"},
		Builder:     buildSoldierAI,
	})

	// Scout AI
	btl.RegisterTemplate(&BehaviorTreeTemplate{
		Name:        "scout_ai",
		Description: "Scout AI that explores the map and reports enemies",
		UnitTypes:   []string{"scout", "explorer"},
		Builder:     buildScoutAI,
	})

	// Builder AI
	btl.RegisterTemplate(&BehaviorTreeTemplate{
		Name:        "builder_ai",
		Description: "Builder AI that constructs buildings when ordered",
		UnitTypes:   []string{"builder", "engineer"},
		Builder:     buildBuilderAI,
	})

	// Guard AI
	btl.RegisterTemplate(&BehaviorTreeTemplate{
		Name:        "guard_ai",
		Description: "Guard AI that patrols and defends a specific area",
		UnitTypes:   []string{"guard", "sentry"},
		Builder:     buildGuardAI,
	})

	// General Purpose AI
	btl.RegisterTemplate(&BehaviorTreeTemplate{
		Name:        "general_ai",
		Description: "General purpose AI suitable for most unit types",
		UnitTypes:   []string{"*"},
		Builder:     buildGeneralAI,
	})
}

// Worker AI: Gathers resources when idle, returns to base when carrying resources
func buildWorkerAI() BehaviorNode {
	// Main sequence: Check for work, execute work
	mainSequence := NewSequenceNode("WorkerMainSequence")

	// Check if carrying resources and need to return to base
	carryingCheck := NewIsCarryingResourcesCondition("IsCarryingResources", 50)
	returnToBase := NewMoveToPositionAction("ReturnToBase", "home_position", 2.0)

	carryingSelector := NewSelectorNode("CarryingResourcesSelector")
	carryingSequence := NewSequenceNode("ReturnResourcesSequence")
	carryingSequence.AddChild(carryingCheck)
	carryingSequence.AddChild(returnToBase)
	carryingSelector.AddChild(carryingSequence)

	// Look for resources to gather
	findResource := NewIsResourceInRangeCondition("FindNearbyResource", 15.0, "target_resource")
	moveToResource := NewMoveToPositionAction("MoveToResource", "resource_position", 1.5)
	gatherResource := NewGatherResourceAction("GatherResource", "target_resource")

	gatherSequence := NewSequenceNode("GatherResourceSequence")
	gatherSequence.AddChild(findResource)
	gatherSequence.AddChild(moveToResource)
	gatherSequence.AddChild(gatherResource)
	carryingSelector.AddChild(gatherSequence)

	// Idle behavior - wait
	idleWait := NewWaitAction("IdleWait", 2*time.Second)
	carryingSelector.AddChild(idleWait)

	mainSequence.AddChild(carryingSelector)

	return mainSequence
}

// Soldier AI: Attacks enemies when found, otherwise patrols or waits
func buildSoldierAI() BehaviorNode {
	// Main selector: Combat or patrol
	mainSelector := NewSelectorNode("SoldierMainSelector")

	// Combat sequence: Find enemy, attack enemy
	combatSequence := NewSequenceNode("CombatSequence")
	findEnemy := NewIsEnemyInRangeCondition("FindEnemy", 10.0, "target_enemy")
	attackEnemy := NewAttackTargetAction("AttackEnemy", "target_enemy")

	combatSequence.AddChild(findEnemy)
	combatSequence.AddChild(attackEnemy)
	mainSelector.AddChild(combatSequence)

	// Patrol sequence: Move to patrol point, wait, move to next point
	patrolSequence := NewSequenceNode("PatrolSequence")
	moveToPatrol := NewMoveToPositionAction("MoveToPatrol", "patrol_point", 2.0)
	patrolWait := NewWaitAction("PatrolWait", 3*time.Second)

	patrolSequence.AddChild(moveToPatrol)
	patrolSequence.AddChild(patrolWait)
	mainSelector.AddChild(patrolSequence)

	return mainSelector
}

// Scout AI: Explores map, reports enemies, avoids combat
func buildScoutAI() BehaviorNode {
	// Main selector: Report enemy or explore
	mainSelector := NewSelectorNode("ScoutMainSelector")

	// Enemy reporting sequence
	reportSequence := NewSequenceNode("ReportEnemySequence")
	findEnemy := NewIsEnemyInRangeCondition("DetectEnemy", 12.0, "detected_enemy")

	// Move away from enemy (scout should avoid combat)
	retreatAction := NewMoveToPositionAction("Retreat", "retreat_position", 3.0)
	reportSequence.AddChild(findEnemy)
	reportSequence.AddChild(retreatAction)
	mainSelector.AddChild(reportSequence)

	// Exploration sequence
	exploreSequence := NewSequenceNode("ExploreSequence")
	moveToExplore := NewMoveToPositionAction("MoveToExplore", "explore_target", 2.0)
	exploreWait := NewWaitAction("ExploreWait", 1*time.Second)

	exploreSequence.AddChild(moveToExplore)
	exploreSequence.AddChild(exploreWait)
	mainSelector.AddChild(exploreSequence)

	return mainSelector
}

// Builder AI: Constructs buildings when ordered, otherwise waits
func buildBuilderAI() BehaviorNode {
	// Main selector: Build or wait
	mainSelector := NewSelectorNode("BuilderMainSelector")

	// Build sequence: Check for build order, move to location, build
	buildSequence := NewSequenceNode("BuildSequence")
	checkBuildOrder := NewIsBlackboardKeySetCondition("HasBuildOrder", "build_position")
	moveToBuildSite := NewMoveToPositionAction("MoveToBuildSite", "build_position", 1.0)
	buildStructure := NewBuildStructureAction("BuildStructure", "build_position", "barracks")

	buildSequence.AddChild(checkBuildOrder)
	buildSequence.AddChild(moveToBuildSite)
	buildSequence.AddChild(buildStructure)
	mainSelector.AddChild(buildSequence)

	// Idle wait
	idleWait := NewWaitAction("BuilderIdle", 5*time.Second)
	mainSelector.AddChild(idleWait)

	return mainSelector
}

// Guard AI: Defends area, attacks intruders
func buildGuardAI() BehaviorNode {
	// Main selector: Defend or patrol
	mainSelector := NewSelectorNode("GuardMainSelector")

	// Defense sequence: Enemy in guard zone -> attack
	defendSequence := NewSequenceNode("DefendSequence")
	detectIntruder := NewIsEnemyInRangeCondition("DetectIntruder", 8.0, "intruder")
	attackIntruder := NewAttackTargetAction("AttackIntruder", "intruder")

	defendSequence.AddChild(detectIntruder)
	defendSequence.AddChild(attackIntruder)
	mainSelector.AddChild(defendSequence)

	// Patrol sequence: Move around guard area
	patrolSequence := NewSequenceNode("GuardPatrolSequence")
	returnToPost := NewMoveToPositionAction("ReturnToPost", "guard_post", 1.0)
	guardWait := NewWaitAction("GuardWait", 4*time.Second)

	patrolSequence.AddChild(returnToPost)
	patrolSequence.AddChild(guardWait)
	mainSelector.AddChild(patrolSequence)

	return mainSelector
}

// General AI: Flexible AI suitable for most unit types
func buildGeneralAI() BehaviorNode {
	// Main selector: Combat, work, or idle
	mainSelector := NewSelectorNode("GeneralMainSelector")

	// Health check - retreat if low health
	healthSelector := NewSelectorNode("HealthSelector")
	lowHealthCheck := NewIsHealthLowCondition("IsHealthLow", 0.3)
	retreatAction := NewMoveToPositionAction("Retreat", "safe_position", 3.0)

	healthSequence := NewSequenceNode("RetreatSequence")
	healthSequence.AddChild(lowHealthCheck)
	healthSequence.AddChild(retreatAction)
	healthSelector.AddChild(healthSequence)

	// Combat sequence
	combatSequence := NewSequenceNode("GeneralCombatSequence")
	findEnemy := NewIsEnemyInRangeCondition("FindNearbyEnemy", 8.0, "combat_target")
	attackTarget := NewAttackTargetAction("AttackTarget", "combat_target")

	combatSequence.AddChild(findEnemy)
	combatSequence.AddChild(attackTarget)
	healthSelector.AddChild(combatSequence)

	// Work sequence - gather resources if carrying capacity allows
	workSequence := NewSequenceNode("WorkSequence")
	notCarryingMuch := NewIsCarryingResourcesCondition("NotCarryingMuch", 10)
	notCarryingInverter := NewInverterNode("NotCarryingInverter")
	notCarryingInverter.AddChild(notCarryingMuch)

	findResource := NewIsResourceInRangeCondition("FindResource", 12.0, "work_resource")
	gatherResource := NewGatherResourceAction("GatherResource", "work_resource")

	workSequence.AddChild(notCarryingInverter)
	workSequence.AddChild(findResource)
	workSequence.AddChild(gatherResource)
	healthSelector.AddChild(workSequence)

	// Idle behavior
	idleWait := NewWaitAction("GeneralIdle", 3*time.Second)
	healthSelector.AddChild(idleWait)

	mainSelector.AddChild(healthSelector)

	return mainSelector
}

// BehaviorTreeFactory creates and configures behavior trees for units
type BehaviorTreeFactory struct {
	library *BehaviorTreeLibrary
	world   *World
}

// NewBehaviorTreeFactory creates a new behavior tree factory
func NewBehaviorTreeFactory(world *World) *BehaviorTreeFactory {
	return &BehaviorTreeFactory{
		library: NewBehaviorTreeLibrary(),
		world:   world,
	}
}

// CreateTreeForUnit creates an appropriate behavior tree for a unit
func (btf *BehaviorTreeFactory) CreateTreeForUnit(unit *GameUnit) (*BehaviorTree, error) {
	// Get unit type from unit definition or use default
	unitType := "general" // Default unit type
	if unit.UnitType != "" {
		unitType = unit.UnitType
	}

	// Find appropriate templates for this unit type
	templates := btf.library.GetTemplatesForUnitType(unitType)
	if len(templates) == 0 {
		// Fallback to general AI
		templates = btf.library.GetTemplatesForUnitType("*")
	}

	if len(templates) == 0 {
		return nil, fmt.Errorf("no suitable behavior tree template found for unit type: %s", unitType)
	}

	// Use the first matching template
	template := templates[0]
	tree, err := btf.library.CreateBehaviorTree(template.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to create behavior tree: %w", err)
	}

	// Configure the tree with unit-specific data
	btf.configureTreeForUnit(tree, unit)

	return tree, nil
}

// CreateTreeByName creates a behavior tree from a specific template
func (btf *BehaviorTreeFactory) CreateTreeByName(templateName string, unit *GameUnit) (*BehaviorTree, error) {
	tree, err := btf.library.CreateBehaviorTree(templateName)
	if err != nil {
		return nil, err
	}

	btf.configureTreeForUnit(tree, unit)
	return tree, nil
}

// configureTreeForUnit sets up unit-specific data in the behavior tree
func (btf *BehaviorTreeFactory) configureTreeForUnit(tree *BehaviorTree, unit *GameUnit) {
	// This would be called after tree.Start() is called and context is available
	// For now, we'll store the configuration to apply later
}

// GetAvailableTemplates returns all available template names
func (btf *BehaviorTreeFactory) GetAvailableTemplates() []string {
	return btf.library.GetAllTemplateNames()
}

// GetLibrary returns the behavior tree library
func (btf *BehaviorTreeFactory) GetLibrary() *BehaviorTreeLibrary {
	return btf.library
}

// SetupUnitBehavior is a convenience function to create and assign a behavior tree to a unit
func (btf *BehaviorTreeFactory) SetupUnitBehavior(unit *GameUnit, btManager *BehaviorTreeManager) error {
	tree, err := btf.CreateTreeForUnit(unit)
	if err != nil {
		return fmt.Errorf("failed to create behavior tree for unit %d: %w", unit.ID, err)
	}

	return btManager.SetBehaviorTree(unit.ID, tree)
}

// SetupUnitBehaviorByTemplate assigns a specific behavior template to a unit
func (btf *BehaviorTreeFactory) SetupUnitBehaviorByTemplate(unit *GameUnit, templateName string, btManager *BehaviorTreeManager) error {
	tree, err := btf.CreateTreeByName(templateName, unit)
	if err != nil {
		return fmt.Errorf("failed to create behavior tree %s for unit %d: %w", templateName, unit.ID, err)
	}

	return btManager.SetBehaviorTree(unit.ID, tree)
}