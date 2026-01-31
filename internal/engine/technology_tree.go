package engine

import (
	"fmt"
	"sync"
	"time"
)

// TechnologyTree manages research and technology dependencies
type TechnologyTree struct {
	// Player technology tracking
	playerTechnologies  map[int]map[string]*Technology // player_id -> tech_name -> technology
	researchQueue       map[int][]*ResearchItem        // player_id -> research queue
	currentResearch     map[int]*ResearchItem          // player_id -> current research

	// Technology definitions
	technologies        map[string]*TechnologyDefinition // tech_name -> definition
	dependencies        map[string][]string              // tech_name -> required_technologies

	mutex               sync.RWMutex
}

// TechnologyDefinition defines a technology that can be researched
type TechnologyDefinition struct {
	Name         string              `json:"name"`
	DisplayName  string              `json:"display_name"`
	Description  string              `json:"description"`
	Requirements []string            `json:"requirements"`    // Required technologies
	Cost         map[string]int      `json:"cost"`           // Resource cost
	Duration     time.Duration       `json:"duration"`       // Research time
	Effects      []TechnologyEffect  `json:"effects"`        // What this tech provides
	Category     string              `json:"category"`       // military, economic, etc.
	Era          int                 `json:"era"`            // Technology tier/era
}

// TechnologyEffect describes what a technology enables or improves
type TechnologyEffect struct {
	Type         string      `json:"type"`         // "enable_unit", "improve_stat", "unlock_building", etc.
	Target       string      `json:"target"`       // Unit/building type affected
	Value        float64     `json:"value"`        // Improvement value
	Description  string      `json:"description"`  // Human-readable description
}

// Technology represents a researched technology for a player
type Technology struct {
	Name         string          `json:"name"`
	PlayerID     int             `json:"player_id"`
	ResearchedAt time.Time       `json:"researched_at"`
	Definition   *TechnologyDefinition `json:"-"`
}

// ResearchItem represents an item being researched
type ResearchItem struct {
	TechName     string          `json:"tech_name"`
	PlayerID     int             `json:"player_id"`
	Progress     float32         `json:"progress"`     // 0.0 to 1.0
	StartTime    time.Time       `json:"start_time"`
	Duration     time.Duration   `json:"duration"`
	Cost         map[string]int  `json:"cost"`
	BuildingID   int             `json:"building_id"`  // Building conducting research
}

// NewTechnologyTree creates a new technology tree
func NewTechnologyTree() *TechnologyTree {
	tt := &TechnologyTree{
		playerTechnologies: make(map[int]map[string]*Technology),
		researchQueue:      make(map[int][]*ResearchItem),
		currentResearch:    make(map[int]*ResearchItem),
		technologies:       make(map[string]*TechnologyDefinition),
		dependencies:       make(map[string][]string),
	}

	// Initialize default technologies
	tt.initializeDefaultTechnologies()

	return tt
}

// initializeDefaultTechnologies sets up basic technology definitions
func (tt *TechnologyTree) initializeDefaultTechnologies() {
	// Basic technologies for both factions
	tt.technologies["construction_efficiency"] = &TechnologyDefinition{
		Name:         "construction_efficiency",
		DisplayName:  "Construction Efficiency",
		Description:  "Workers build 25% faster",
		Requirements: []string{},
		Cost:         map[string]int{"wood": 100, "gold": 50},
		Duration:     30 * time.Second,
		Effects: []TechnologyEffect{
			{Type: "improve_stat", Target: "construction_speed", Value: 1.25, Description: "+25% construction speed"},
		},
		Category: "economic",
		Era:      1,
	}

	tt.technologies["advanced_construction"] = &TechnologyDefinition{
		Name:         "advanced_construction",
		DisplayName:  "Advanced Construction",
		Description:  "Workers build 50% faster and unlock advanced buildings",
		Requirements: []string{"construction_efficiency"},
		Cost:         map[string]int{"wood": 200, "stone": 100, "gold": 100},
		Duration:     60 * time.Second,
		Effects: []TechnologyEffect{
			{Type: "improve_stat", Target: "construction_speed", Value: 1.5, Description: "+50% construction speed"},
			{Type: "unlock_building", Target: "fortress", Value: 0, Description: "Unlocks Fortress"},
		},
		Category: "economic",
		Era:      2,
	}

	// Military technologies
	tt.technologies["iron_weapons"] = &TechnologyDefinition{
		Name:         "iron_weapons",
		DisplayName:  "Iron Weapons",
		Description:  "Unlocks iron weapons and +2 attack for melee units",
		Requirements: []string{},
		Cost:         map[string]int{"wood": 75, "stone": 75, "gold": 100},
		Duration:     45 * time.Second,
		Effects: []TechnologyEffect{
			{Type: "improve_stat", Target: "melee_attack", Value: 2, Description: "+2 attack for melee units"},
			{Type: "enable_unit", Target: "swordman", Value: 0, Description: "Unlocks Swordsman"},
		},
		Category: "military",
		Era:      1,
	}

	tt.technologies["archery"] = &TechnologyDefinition{
		Name:         "archery",
		DisplayName:  "Archery",
		Description:  "Unlocks ranged units and improves range",
		Requirements: []string{},
		Cost:         map[string]int{"wood": 100, "gold": 75},
		Duration:     40 * time.Second,
		Effects: []TechnologyEffect{
			{Type: "enable_unit", Target: "archer", Value: 0, Description: "Unlocks Archer"},
			{Type: "improve_stat", Target: "ranged_range", Value: 1, Description: "+1 attack range"},
		},
		Category: "military",
		Era:      1,
	}

	// Magic faction technologies
	tt.technologies["summoning"] = &TechnologyDefinition{
		Name:         "summoning",
		DisplayName:  "Summoning Arts",
		Description:  "Unlocks basic summoned creatures",
		Requirements: []string{},
		Cost:         map[string]int{"energy": 100, "gold": 100},
		Duration:     50 * time.Second,
		Effects: []TechnologyEffect{
			{Type: "enable_unit", Target: "daemon", Value: 0, Description: "Unlocks Daemon"},
			{Type: "unlock_building", Target: "summoner_guild", Value: 0, Description: "Unlocks Summoner Guild"},
		},
		Category: "magic",
		Era:      1,
	}

	tt.technologies["dragon_mastery"] = &TechnologyDefinition{
		Name:         "dragon_mastery",
		DisplayName:  "Dragon Mastery",
		Description:  "Unlocks dragon units",
		Requirements: []string{"summoning", "archmage_training"},
		Cost:         map[string]int{"energy": 300, "gold": 500},
		Duration:     120 * time.Second,
		Effects: []TechnologyEffect{
			{Type: "enable_unit", Target: "dragon", Value: 0, Description: "Unlocks Dragon"},
		},
		Category: "magic",
		Era:      3,
	}

	// Set up dependencies
	tt.dependencies["advanced_construction"] = []string{"construction_efficiency"}
	tt.dependencies["dragon_mastery"] = []string{"summoning", "archmage_training"}
}

// InitializePlayer initializes technology tracking for a player
func (tt *TechnologyTree) InitializePlayer(playerID int) {
	tt.mutex.Lock()
	defer tt.mutex.Unlock()

	tt.playerTechnologies[playerID] = make(map[string]*Technology)
	tt.researchQueue[playerID] = make([]*ResearchItem, 0)
	tt.currentResearch[playerID] = nil
}

// HasTechnology checks if a player has researched a technology
func (tt *TechnologyTree) HasTechnology(playerID int, techName string) bool {
	tt.mutex.RLock()
	defer tt.mutex.RUnlock()

	playerTechs, exists := tt.playerTechnologies[playerID]
	if !exists {
		return false
	}

	_, hasTech := playerTechs[techName]
	return hasTech
}

// HasRequiredTech checks if a player has the required technology for a unit/building
func (tt *TechnologyTree) HasRequiredTech(playerID int, unitOrBuildingType string) bool {
	// Get required technologies for this unit/building type
	requiredTechs := tt.getRequiredTechnologiesFor(unitOrBuildingType)

	// Check if player has all required technologies
	for _, techName := range requiredTechs {
		if !tt.HasTechnology(playerID, techName) {
			return false
		}
	}

	return true
}

// getRequiredTechnologiesFor returns what technologies are required for a unit/building
func (tt *TechnologyTree) getRequiredTechnologiesFor(unitType string) []string {
	// Define technology requirements for units and buildings
	requirements := map[string][]string{
		// Units requiring technology
		"swordman":      {"iron_weapons"},
		"archer":        {"archery"},
		"daemon":        {"summoning"},
		"dragon":        {"dragon_mastery"},
		"catapult":      {"advanced_construction"},

		// Buildings requiring technology
		"fortress":      {"advanced_construction"},
		"summoner_guild": {"summoning"},
		"dragon_lair":   {"dragon_mastery"},
	}

	if techs, exists := requirements[unitType]; exists {
		return techs
	}

	return []string{} // No technology requirements
}

// CanResearchTechnology checks if a player can research a technology
func (tt *TechnologyTree) CanResearchTechnology(playerID int, techName string) (bool, string) {
	tt.mutex.RLock()
	defer tt.mutex.RUnlock()

	// Check if technology exists
	techDef, exists := tt.technologies[techName]
	if !exists {
		return false, fmt.Sprintf("technology %s does not exist", techName)
	}

	// Check if already researched
	if tt.HasTechnology(playerID, techName) {
		return false, fmt.Sprintf("technology %s already researched", techName)
	}

	// Check if already researching this technology
	if currentResearch, exists := tt.currentResearch[playerID]; exists && currentResearch != nil {
		if currentResearch.TechName == techName {
			return false, fmt.Sprintf("technology %s already being researched", techName)
		}
	}

	// Check if prerequisites are met
	for _, reqTech := range techDef.Requirements {
		if !tt.HasTechnology(playerID, reqTech) {
			return false, fmt.Sprintf("missing required technology: %s", reqTech)
		}
	}

	return true, ""
}

// StartResearch begins researching a technology
func (tt *TechnologyTree) StartResearch(playerID int, techName string, buildingID int) error {
	// Validate research
	canResearch, reason := tt.CanResearchTechnology(playerID, techName)
	if !canResearch {
		return fmt.Errorf("cannot research %s: %s", techName, reason)
	}

	tt.mutex.Lock()
	defer tt.mutex.Unlock()

	// Get technology definition
	techDef := tt.technologies[techName]

	// Create research item
	research := &ResearchItem{
		TechName:   techName,
		PlayerID:   playerID,
		Progress:   0.0,
		StartTime:  time.Now(),
		Duration:   techDef.Duration,
		Cost:       techDef.Cost,
		BuildingID: buildingID,
	}

	// Check if something is already being researched
	if current := tt.currentResearch[playerID]; current != nil {
		// Add to queue
		tt.researchQueue[playerID] = append(tt.researchQueue[playerID], research)
	} else {
		// Start immediately
		tt.currentResearch[playerID] = research
	}

	return nil
}

// ProcessResearch updates research progress for all players
func (tt *TechnologyTree) ProcessResearch(deltaTime time.Duration) {
	tt.mutex.Lock()
	defer tt.mutex.Unlock()

	for playerID, research := range tt.currentResearch {
		if research == nil {
			continue
		}

		// Update research progress
		elapsed := time.Since(research.StartTime)
		research.Progress = float32(elapsed.Seconds()) / float32(research.Duration.Seconds())

		// Check if research is complete
		if research.Progress >= 1.0 {
			tt.completeResearch(playerID, research)

			// Start next research if queued
			if len(tt.researchQueue[playerID]) > 0 {
				tt.currentResearch[playerID] = tt.researchQueue[playerID][0]
				tt.researchQueue[playerID] = tt.researchQueue[playerID][1:]
				tt.currentResearch[playerID].StartTime = time.Now()
			} else {
				tt.currentResearch[playerID] = nil
			}
		}
	}
}

// completeResearch finishes a technology research
func (tt *TechnologyTree) completeResearch(playerID int, research *ResearchItem) {
	// Ensure player technologies map exists
	if tt.playerTechnologies[playerID] == nil {
		tt.playerTechnologies[playerID] = make(map[string]*Technology)
	}

	// Add technology to player
	tech := &Technology{
		Name:         research.TechName,
		PlayerID:     playerID,
		ResearchedAt: time.Now(),
		Definition:   tt.technologies[research.TechName],
	}

	tt.playerTechnologies[playerID][research.TechName] = tech
}

// GetResearchProgress returns current research progress for a player
func (tt *TechnologyTree) GetResearchProgress(playerID int) (*ResearchItem, []*ResearchItem) {
	tt.mutex.RLock()
	defer tt.mutex.RUnlock()

	current := tt.currentResearch[playerID]

	// Copy queue to avoid race conditions
	queue := make([]*ResearchItem, 0, len(tt.researchQueue[playerID]))
	for _, item := range tt.researchQueue[playerID] {
		queue = append(queue, item)
	}

	return current, queue
}

// GetPlayerTechnologies returns all technologies researched by a player
func (tt *TechnologyTree) GetPlayerTechnologies(playerID int) map[string]*Technology {
	tt.mutex.RLock()
	defer tt.mutex.RUnlock()

	playerTechs := tt.playerTechnologies[playerID]
	if playerTechs == nil {
		return make(map[string]*Technology)
	}

	// Copy to avoid race conditions
	result := make(map[string]*Technology)
	for name, tech := range playerTechs {
		result[name] = tech
	}

	return result
}

// GetAvailableTechnologies returns technologies that can be researched
func (tt *TechnologyTree) GetAvailableTechnologies(playerID int) []*TechnologyDefinition {
	available := make([]*TechnologyDefinition, 0)

	for _, techDef := range tt.technologies {
		canResearch, _ := tt.CanResearchTechnology(playerID, techDef.Name)
		if canResearch {
			available = append(available, techDef)
		}
	}

	return available
}

// CancelResearch cancels current research and refunds partial resources
func (tt *TechnologyTree) CancelResearch(playerID int) (*ResearchItem, map[string]int) {
	tt.mutex.Lock()
	defer tt.mutex.Unlock()

	research := tt.currentResearch[playerID]
	if research == nil {
		return nil, nil
	}

	// Calculate refund based on progress (lose some resources for partial research)
	refundRatio := 1.0 - (float64(research.Progress) * 0.5) // Lose 50% of progress
	refundResources := make(map[string]int)

	for resource, amount := range research.Cost {
		refundAmount := int(float64(amount) * refundRatio)
		if refundAmount > 0 {
			refundResources[resource] = refundAmount
		}
	}

	// Clear current research
	tt.currentResearch[playerID] = nil

	// Start next research if queued
	if len(tt.researchQueue[playerID]) > 0 {
		tt.currentResearch[playerID] = tt.researchQueue[playerID][0]
		tt.researchQueue[playerID] = tt.researchQueue[playerID][1:]
		tt.currentResearch[playerID].StartTime = time.Now()
	}

	return research, refundResources
}

// GetTechnologyDefinition returns the definition for a technology
func (tt *TechnologyTree) GetTechnologyDefinition(techName string) *TechnologyDefinition {
	tt.mutex.RLock()
	defer tt.mutex.RUnlock()

	return tt.technologies[techName]
}

// Update performs technology tree updates
func (tt *TechnologyTree) Update(deltaTime time.Duration) {
	tt.ProcessResearch(deltaTime)
}