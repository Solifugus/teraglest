package engine

import (
	"fmt"
	"sync"
	"time"

	"teraglest/internal/data"
)

// UnitManager handles unit creation, tracking, and spatial queries
type UnitManager struct {
	units         map[int]*GameUnit       // All units indexed by ID
	unitsByPlayer map[int]map[int]*GameUnit // Units indexed by player ID, then unit ID
	world         *World                   // Reference to world for grid operations
	nextID        int                      // Next available unit ID
	mutex         sync.RWMutex             // Thread-safe access
}

// NewUnitManager creates a new unit manager
func NewUnitManager(world *World) *UnitManager {
	return &UnitManager{
		units:         make(map[int]*GameUnit),
		unitsByPlayer: make(map[int]map[int]*GameUnit),
		world:         world,
		nextID:        1,
	}
}

// CreateUnit creates a new game unit
func (um *UnitManager) CreateUnit(playerID int, unitType string, position Vector3, unitDef *data.UnitDefinition) (*GameUnit, error) {
	um.mutex.Lock()
	defer um.mutex.Unlock()

	unitID := um.nextID
	um.nextID++

	unit := &GameUnit{
		ID:           unitID,
		PlayerID:     playerID,
		UnitType:     unitType,
		Name:         unitDef.Name,
		Position:     position,
		GridPos:      WorldToGrid(position, um.world.tileSize), // Initialize grid position
		Health:       unitDef.Unit.Parameters.MaxHP.Value,
		MaxHealth:    unitDef.Unit.Parameters.MaxHP.Value,
		Armor:        unitDef.Unit.Parameters.Armor.Value,
		Energy:       100, // Default energy
		MaxEnergy:    100,
		State:        UnitStateIdle,
		CreationTime: time.Now(),
		LastUpdate:   time.Now(),
		CommandQueue: make([]UnitCommand, 0),
		Speed:        2.0, // Default movement speed
		CarriedResources: make(map[string]int),
		GatherRate:   map[string]float32{"wood": 10.0, "stone": 8.0, "gold": 12.0},
		UnitDef:      unitDef,
	}

	// Set combat stats based on unit definition
	if len(unitDef.Unit.Parameters.ResourceRequirements) > 0 {
		// Infer combat stats from cost and armor
		unit.AttackDamage = 10 + unit.Armor/2 // Simple damage calculation
		unit.AttackRange = 1.0 + float32(unit.Armor)/10.0 // Range based on armor
		unit.AttackSpeed = 1.0 // Attacks per second
	}

	// Store unit
	um.units[unitID] = unit

	// Index by player
	if um.unitsByPlayer[playerID] == nil {
		um.unitsByPlayer[playerID] = make(map[int]*GameUnit)
	}
	um.unitsByPlayer[playerID][unitID] = unit

	// Mark grid position as occupied
	um.world.SetOccupied(unit.GridPos.Grid, true)

	return unit, nil
}

// GetUnit returns a unit by ID (thread-safe)
func (um *UnitManager) GetUnit(unitID int) *GameUnit {
	um.mutex.RLock()
	defer um.mutex.RUnlock()
	return um.units[unitID]
}

// GetUnitsForPlayer returns all units for a specific player (thread-safe)
func (um *UnitManager) GetUnitsForPlayer(playerID int) map[int]*GameUnit {
	um.mutex.RLock()
	defer um.mutex.RUnlock()

	playerUnits := make(map[int]*GameUnit)
	if units, exists := um.unitsByPlayer[playerID]; exists {
		// Create a copy to prevent external modifications
		for id, unit := range units {
			playerUnits[id] = unit
		}
	}
	return playerUnits
}

// RemoveUnit removes a unit from the game (thread-safe)
func (um *UnitManager) RemoveUnit(unitID int) error {
	um.mutex.Lock()
	defer um.mutex.Unlock()

	unit, exists := um.units[unitID]
	if !exists {
		return fmt.Errorf("unit with ID %d not found", unitID)
	}

	// Free grid position
	um.world.SetOccupied(unit.GridPos.Grid, false)

	// Remove from global index
	delete(um.units, unitID)

	// Remove from player index
	if playerUnits, exists := um.unitsByPlayer[unit.PlayerID]; exists {
		delete(playerUnits, unitID)
		// Clean up empty player map
		if len(playerUnits) == 0 {
			delete(um.unitsByPlayer, unit.PlayerID)
		}
	}

	return nil
}

// GetUnitsAtPosition returns all units at a specific grid position
func (um *UnitManager) GetUnitsAtPosition(gridPos Vector2i) []*GameUnit {
	um.mutex.RLock()
	defer um.mutex.RUnlock()

	var unitsAtPosition []*GameUnit
	for _, unit := range um.units {
		if unit.GridPos.Grid.X == gridPos.X && unit.GridPos.Grid.Y == gridPos.Y {
			unitsAtPosition = append(unitsAtPosition, unit)
		}
	}
	return unitsAtPosition
}

// GetUnitsInTile returns all units at a specific grid tile
func (um *UnitManager) GetUnitsInTile(gridPos Vector2i) []*GameUnit {
	return um.GetUnitsAtPosition(gridPos)
}

// GetUnitsInArea returns all units within a rectangular area
func (um *UnitManager) GetUnitsInArea(topLeft, bottomRight Vector2i) []*GameUnit {
	um.mutex.RLock()
	defer um.mutex.RUnlock()

	var unitsInArea []*GameUnit
	for _, unit := range um.units {
		unitPos := unit.GridPos.Grid
		if unitPos.X >= topLeft.X && unitPos.X <= bottomRight.X &&
		   unitPos.Y >= topLeft.Y && unitPos.Y <= bottomRight.Y {
			unitsInArea = append(unitsInArea, unit)
		}
	}
	return unitsInArea
}

// IsPositionOccupied checks if a grid position is occupied by any unit
func (um *UnitManager) IsPositionOccupied(gridPos Vector2i) bool {
	return len(um.GetUnitsAtPosition(gridPos)) > 0
}

// FindNearestFreePosition finds the nearest unoccupied position to a target
func (um *UnitManager) FindNearestFreePosition(targetPos Vector2i) Vector2i {
	// Check if target position is already free
	if !um.IsPositionOccupied(targetPos) && um.world.IsPositionWalkable(targetPos) {
		return targetPos
	}

	// Use world's method which checks both walkability and occupancy
	return um.world.GetNearestWalkablePosition(targetPos)
}

// GetNearestUnit finds the nearest unit to a given position within a radius
func (um *UnitManager) GetNearestUnit(position Vector2i, radius int, excludePlayerID int) *GameUnit {
	um.mutex.RLock()
	defer um.mutex.RUnlock()

	var nearestUnit *GameUnit
	nearestDistance := float64(radius * radius + 1) // Start with beyond max radius

	for _, unit := range um.units {
		// Skip units from the same player if exclusion is specified
		if excludePlayerID >= 0 && unit.PlayerID == excludePlayerID {
			continue
		}

		// Calculate distance
		unitPos := unit.GridPos.Grid
		dx := float64(position.X - unitPos.X)
		dy := float64(position.Y - unitPos.Y)
		distance := dx*dx + dy*dy

		if distance < nearestDistance && distance <= float64(radius*radius) {
			nearestDistance = distance
			nearestUnit = unit
		}
	}

	return nearestUnit
}

// Update updates all units
func (um *UnitManager) Update(deltaTime time.Duration) {
	um.mutex.RLock()
	units := make([]*GameUnit, 0, len(um.units))
	for _, unit := range um.units {
		units = append(units, unit)
	}
	um.mutex.RUnlock()

	// Update units without holding the main lock
	for _, unit := range units {
		if unit.IsAlive() {
			oldGridPos := unit.GridPos.Grid

			// Update the unit
			unit.Update(deltaTime)

			// Check if unit moved to a new grid position
			newGridPos := unit.GridPos.Grid
			if oldGridPos.X != newGridPos.X || oldGridPos.Y != newGridPos.Y {
				um.updateUnitGridPosition(unit, oldGridPos, newGridPos)
			}
		} else {
			// Remove dead units
			um.RemoveUnit(unit.GetID())
		}
	}
}

// updateUnitGridPosition updates occupancy grid when a unit moves
func (um *UnitManager) updateUnitGridPosition(unit *GameUnit, oldPos, newPos Vector2i) {
	// Free old position if no other units are there
	if len(um.GetUnitsAtPosition(oldPos)) == 0 {
		um.world.SetOccupied(oldPos, false)
	}

	// Occupy new position
	um.world.SetOccupied(newPos, true)
}

// GetStats returns statistics about the units
func (um *UnitManager) GetStats() UnitManagerStats {
	um.mutex.RLock()
	defer um.mutex.RUnlock()

	stats := UnitManagerStats{
		TotalUnits:    len(um.units),
		UnitsPerPlayer: make(map[int]int),
		UnitsPerState:  make(map[UnitState]int),
	}

	for _, unit := range um.units {
		// Count units per player
		stats.UnitsPerPlayer[unit.PlayerID]++

		// Count units per state
		stats.UnitsPerState[unit.State]++
	}

	return stats
}

// UnitManagerStats contains statistics about the unit manager
type UnitManagerStats struct {
	TotalUnits     int                 `json:"total_units"`
	UnitsPerPlayer map[int]int         `json:"units_per_player"`
	UnitsPerState  map[UnitState]int   `json:"units_per_state"`
}