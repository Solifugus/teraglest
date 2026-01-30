package engine

import (
	"fmt"
	"sync"
	"time"

	"teraglest/internal/data"
)

// Vector3 represents a 3D position or direction
type Vector3 struct {
	X, Y, Z float64
}

// UnitState represents the current state/behavior of a unit
// GameObject represents a base interface for all game objects
type GameObject interface {
	GetID() int
	GetPlayerID() int
	GetPosition() Vector3
	SetPosition(Vector3)
	GetHealth() int
	SetHealth(int)
	GetMaxHealth() int
	IsAlive() bool
	Update(deltaTime time.Duration)
	GetType() string
}


// GameBuilding represents an enhanced building with production and upgrade systems
type GameBuilding struct {
	// Base properties
	ID           int                 `json:"id"`
	PlayerID     int                 `json:"player_id"`
	BuildingType string              `json:"building_type"`
	Name         string              `json:"name"`

	// State management
	Position     Vector3             `json:"position"`
	Rotation     float32             `json:"rotation"`
	Health       int                 `json:"health"`
	MaxHealth    int                 `json:"max_health"`
	Armor        int                 `json:"armor"`

	// Construction lifecycle
	IsBuilt         bool              `json:"is_built"`
	BuildProgress   float32           `json:"build_progress"`
	ConstructionTime time.Duration    `json:"construction_time"`
	CreationTime    time.Time         `json:"creation_time"`
	CompletionTime  time.Time         `json:"completion_time"`

	// Production system
	ProductionQueue []ProductionItem  `json:"production_queue"`
	CurrentProduction *ProductionItem `json:"current_production"`
	ProductionRate   float32          `json:"production_rate"`

	// Resource generation
	ResourceGeneration map[string]float32 `json:"resource_generation"`
	LastResourceGen    time.Time          `json:"last_resource_gen"`

	// Upgrade system
	UpgradeLevel    int                   `json:"upgrade_level"`
	MaxUpgradeLevel int                   `json:"max_upgrade_level"`
	UpgradeProgress float32               `json:"upgrade_progress"`
	CurrentUpgrade  *UpgradeItem          `json:"current_upgrade"`

	// Building definition data
	UnitDef      *data.UnitDefinition     `json:"-"`

	// Threading
	mutex        sync.RWMutex             `json:"-"`
}

// ProductionItem represents an item being produced by a building
type ProductionItem struct {
	ItemType     string                   `json:"item_type"`
	ItemName     string                   `json:"item_name"`
	Progress     float32                  `json:"progress"`
	Duration     time.Duration            `json:"duration"`
	Cost         map[string]int           `json:"cost"`
	StartTime    time.Time                `json:"start_time"`
}

// UpgradeItem represents a building upgrade being processed
type UpgradeItem struct {
	UpgradeType  string                   `json:"upgrade_type"`
	UpgradeName  string                   `json:"upgrade_name"`
	Progress     float32                  `json:"progress"`
	Duration     time.Duration            `json:"duration"`
	Cost         map[string]int           `json:"cost"`
	StartTime    time.Time                `json:"start_time"`
}

// ObjectManager manages all game objects (units, buildings, etc.)
type ObjectManager struct {
	// Object managers
	UnitManager  *UnitManager            `json:"-"`

	// Object storage
	buildings    map[int]*GameBuilding   `json:"buildings"`
	resources    map[int]*ResourceNode   `json:"resources"`

	// ID management
	nextID       int                     `json:"next_id"`

	// Spatial organization
	buildingsByPlayer map[int]map[int]*GameBuilding `json:"buildings_by_player"`

	// Update management
	lastUpdate   time.Time               `json:"last_update"`

	// Threading
	mutex        sync.RWMutex            `json:"-"`

	// Dependencies
	world        *World                  `json:"-"`
}

// NewObjectManager creates a new object manager
func NewObjectManager(world *World) *ObjectManager {
	om := &ObjectManager{
		buildings:         make(map[int]*GameBuilding),
		resources:         make(map[int]*ResourceNode),
		nextID:            1,
		buildingsByPlayer: make(map[int]map[int]*GameBuilding),
		lastUpdate:        time.Now(),
		world:             world,
	}

	// Initialize unit manager
	om.UnitManager = NewUnitManager(world)

	return om
}

// CreateUnit creates a new game unit

// CreateBuilding creates a new game building
func (om *ObjectManager) CreateBuilding(playerID int, buildingType string, position Vector3, unitDef *data.UnitDefinition) (*GameBuilding, error) {
	om.mutex.Lock()
	defer om.mutex.Unlock()

	buildingID := om.nextID
	om.nextID++

	building := &GameBuilding{
		ID:              buildingID,
		PlayerID:        playerID,
		BuildingType:    buildingType,
		Name:            unitDef.Name,
		Position:        position,
		Health:          unitDef.Unit.Parameters.MaxHP.Value,
		MaxHealth:       unitDef.Unit.Parameters.MaxHP.Value,
		Armor:           unitDef.Unit.Parameters.Armor.Value,
		IsBuilt:         false,
		BuildProgress:   0.0,
		ConstructionTime: 30 * time.Second, // Default construction time
		CreationTime:    time.Now(),
		ProductionQueue: make([]ProductionItem, 0),
		ProductionRate:  1.0,
		ResourceGeneration: make(map[string]float32),
		LastResourceGen: time.Now(),
		UpgradeLevel:    1,
		MaxUpgradeLevel: 3,
		UnitDef:         unitDef,
	}

	// Set default resource generation for certain building types
	switch buildingType {
	case "mage_tower", "energy_source":
		building.ResourceGeneration["energy"] = 2.0
	case "farm":
		building.ResourceGeneration["food"] = 1.0
	}

	// Store building
	om.buildings[buildingID] = building

	// Index by player
	if om.buildingsByPlayer[playerID] == nil {
		om.buildingsByPlayer[playerID] = make(map[int]*GameBuilding)
	}
	om.buildingsByPlayer[playerID][buildingID] = building

	return building, nil
}

// GetUnit retrieves a unit by ID (delegates to UnitManager)
func (om *ObjectManager) GetUnit(unitID int) *GameUnit {
	return om.UnitManager.GetUnit(unitID)
}

// GetBuilding retrieves a building by ID
func (om *ObjectManager) GetBuilding(buildingID int) *GameBuilding {
	om.mutex.RLock()
	defer om.mutex.RUnlock()
	return om.buildings[buildingID]
}

// GetUnitsForPlayer returns all units owned by a player (delegates to UnitManager)
func (om *ObjectManager) GetUnitsForPlayer(playerID int) map[int]*GameUnit {
	return om.UnitManager.GetUnitsForPlayer(playerID)
}

// GetBuildingsForPlayer returns all buildings owned by a player
func (om *ObjectManager) GetBuildingsForPlayer(playerID int) map[int]*GameBuilding {
	om.mutex.RLock()
	defer om.mutex.RUnlock()

	result := make(map[int]*GameBuilding)
	if playerBuildings, exists := om.buildingsByPlayer[playerID]; exists {
		for id, building := range playerBuildings {
			result[id] = building
		}
	}
	return result
}

// CreateUnit creates a new game unit (delegates to UnitManager)
func (om *ObjectManager) CreateUnit(playerID int, unitType string, position Vector3, unitDef *data.UnitDefinition) (*GameUnit, error) {
	return om.UnitManager.CreateUnit(playerID, unitType, position, unitDef)
}

// RemoveUnit removes a unit from the game (delegates to UnitManager)
func (om *ObjectManager) RemoveUnit(unitID int) error {
	return om.UnitManager.RemoveUnit(unitID)
}

// GetUnitsAtPosition returns all units at a specific grid position (delegates to UnitManager)
func (om *ObjectManager) GetUnitsAtPosition(gridPos Vector2i) []*GameUnit {
	return om.UnitManager.GetUnitsAtPosition(gridPos)
}

// RemoveBuilding removes a building from the game
func (om *ObjectManager) RemoveBuilding(buildingID int) error {
	om.mutex.Lock()
	defer om.mutex.Unlock()

	building, exists := om.buildings[buildingID]
	if !exists {
		return fmt.Errorf("building %d not found", buildingID)
	}

	// Remove from player index
	if playerBuildings, exists := om.buildingsByPlayer[building.PlayerID]; exists {
		delete(playerBuildings, buildingID)
	}

	// Remove from main storage
	delete(om.buildings, buildingID)
	return nil
}

// Update updates all game objects
func (om *ObjectManager) Update(deltaTime time.Duration) {
	// Update units through UnitManager
	om.UnitManager.Update(deltaTime)

	// Update buildings
	om.mutex.RLock()
	buildings := make([]*GameBuilding, 0, len(om.buildings))
	for _, building := range om.buildings {
		buildings = append(buildings, building)
	}
	om.mutex.RUnlock()

	for _, building := range buildings {
		building.Update(deltaTime)
	}

	// Process resource generation
	om.processResourceGeneration(deltaTime)
}

// GetStats returns object manager statistics
func (om *ObjectManager) GetStats() ObjectManagerStats {
	om.mutex.RLock()
	defer om.mutex.RUnlock()

	// Get unit stats from UnitManager
	unitStats := om.UnitManager.GetStats()

	stats := ObjectManagerStats{
		TotalUnits:     unitStats.TotalUnits,
		TotalBuildings: len(om.buildings),
		TotalResources: len(om.resources),
		UnitsPerPlayer: unitStats.UnitsPerPlayer,
		BuildingsPerPlayer: make(map[int]int),
	}

	for playerID, playerBuildings := range om.buildingsByPlayer {
		stats.BuildingsPerPlayer[playerID] = len(playerBuildings)
	}

	return stats
}

// ObjectManagerStats contains object manager statistics
type ObjectManagerStats struct {
	TotalUnits         int
	TotalBuildings     int
	TotalResources     int
	UnitsPerPlayer     map[int]int
	BuildingsPerPlayer map[int]int
}



// GameBuilding methods implementing GameObject interface
func (b *GameBuilding) GetID() int {
	b.mutex.RLock()
	defer b.mutex.RUnlock()
	return b.ID
}

func (b *GameBuilding) GetPlayerID() int {
	b.mutex.RLock()
	defer b.mutex.RUnlock()
	return b.PlayerID
}

func (b *GameBuilding) GetPosition() Vector3 {
	b.mutex.RLock()
	defer b.mutex.RUnlock()
	return b.Position
}

func (b *GameBuilding) SetPosition(pos Vector3) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	b.Position = pos
}

func (b *GameBuilding) GetHealth() int {
	b.mutex.RLock()
	defer b.mutex.RUnlock()
	return b.Health
}

func (b *GameBuilding) SetHealth(health int) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	b.Health = health
}

func (b *GameBuilding) GetMaxHealth() int {
	b.mutex.RLock()
	defer b.mutex.RUnlock()
	return b.MaxHealth
}

func (b *GameBuilding) IsAlive() bool {
	b.mutex.RLock()
	defer b.mutex.RUnlock()
	return b.Health > 0
}

func (b *GameBuilding) GetType() string {
	b.mutex.RLock()
	defer b.mutex.RUnlock()
	return b.BuildingType
}

// Building-specific update logic
func (b *GameBuilding) Update(deltaTime time.Duration) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	// Update construction progress
	if !b.IsBuilt {
		b.updateConstruction(deltaTime)
	} else {
		// Update production
		b.updateProduction(deltaTime)

		// Update upgrades
		b.updateUpgrades(deltaTime)
	}
}

// Internal helper methods

// processResourceGeneration handles resource generation from buildings
func (om *ObjectManager) processResourceGeneration(deltaTime time.Duration) {
	om.mutex.RLock()
	defer om.mutex.RUnlock()

	// This would integrate with the world's player resource system
	// For now, it's a placeholder that demonstrates the concept
	for _, building := range om.buildings {
		if building.IsBuilt {
			for resourceType, rate := range building.ResourceGeneration {
				generated := rate * float32(deltaTime.Seconds())
				if generated > 0 {
					// Would update player resources here
					_ = resourceType
					_ = generated
				}
			}
		}
	}
}


// Helper methods for unit behavior (simplified implementations)

// Helper methods for building behavior
func (b *GameBuilding) updateConstruction(deltaTime time.Duration) {
	constructionRate := 1.0 / b.ConstructionTime.Seconds()
	b.BuildProgress += float32(constructionRate * deltaTime.Seconds())

	if b.BuildProgress >= 1.0 {
		b.BuildProgress = 1.0
		b.IsBuilt = true
		b.CompletionTime = time.Now()
	}
}

func (b *GameBuilding) updateProduction(deltaTime time.Duration) {
	if b.CurrentProduction == nil && len(b.ProductionQueue) > 0 {
		// Start next production item
		b.CurrentProduction = &b.ProductionQueue[0]
		b.ProductionQueue = b.ProductionQueue[1:]
	}

	if b.CurrentProduction != nil {
		progressRate := 1.0 / b.CurrentProduction.Duration.Seconds()
		b.CurrentProduction.Progress += float32(progressRate * deltaTime.Seconds())

		if b.CurrentProduction.Progress >= 1.0 {
			// Production complete
			// Would create new unit or trigger completion event here
			b.CurrentProduction = nil
		}
	}
}

func (b *GameBuilding) updateUpgrades(deltaTime time.Duration) {
	if b.CurrentUpgrade != nil {
		progressRate := 1.0 / b.CurrentUpgrade.Duration.Seconds()
		b.CurrentUpgrade.Progress += float32(progressRate * deltaTime.Seconds())

		if b.CurrentUpgrade.Progress >= 1.0 {
			// Upgrade complete
			b.UpgradeLevel++
			b.CurrentUpgrade = nil
		}
	}
}