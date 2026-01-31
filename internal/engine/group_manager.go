package engine

import (
	"fmt"
	"sync"
	"time"
)

// GroupManager manages all unit groups and formations for a world
type GroupManager struct {
	world       *World                    // Reference to game world
	groups      map[int]*UnitGroup        // All active groups by ID
	playerGroups map[int]map[int]*UnitGroup // Groups by player ID
	unitGroups  map[int]*UnitGroup        // Unit ID to group mapping
	nextGroupID int                       // Next available group ID
	mutex       sync.RWMutex              // Thread safety
}

// NewGroupManager creates a new group manager
func NewGroupManager(world *World) *GroupManager {
	return &GroupManager{
		world:        world,
		groups:       make(map[int]*UnitGroup),
		playerGroups: make(map[int]map[int]*UnitGroup),
		unitGroups:   make(map[int]*UnitGroup),
		nextGroupID:  1,
	}
}

// CreateGroup creates a new unit group with the specified units
func (gm *GroupManager) CreateGroup(playerID int, units []*GameUnit, formationType FormationType) (*UnitGroup, error) {
	gm.mutex.Lock()
	defer gm.mutex.Unlock()

	if len(units) == 0 {
		return nil, fmt.Errorf("cannot create group with no units")
	}

	// Remove units from existing groups first
	for _, unit := range units {
		gm.removeUnitFromAnyGroup(unit.ID)
	}

	// Create new group
	groupID := gm.nextGroupID
	gm.nextGroupID++

	group := NewUnitGroup(groupID, playerID, units, formationType)
	gm.groups[groupID] = group

	// Add to player groups
	if gm.playerGroups[playerID] == nil {
		gm.playerGroups[playerID] = make(map[int]*UnitGroup)
	}
	gm.playerGroups[playerID][groupID] = group

	// Map units to group
	for _, unit := range units {
		gm.unitGroups[unit.ID] = group
	}

	return group, nil
}

// AddUnitsToGroup adds units to an existing group
func (gm *GroupManager) AddUnitsToGroup(groupID int, units []*GameUnit) error {
	gm.mutex.Lock()
	defer gm.mutex.Unlock()

	group, exists := gm.groups[groupID]
	if !exists {
		return fmt.Errorf("group %d not found", groupID)
	}

	// Remove units from other groups first
	for _, unit := range units {
		gm.removeUnitFromAnyGroup(unit.ID)
	}

	// Add to target group
	for _, unit := range units {
		group.AddUnit(unit)
		gm.unitGroups[unit.ID] = group
	}

	return nil
}

// RemoveUnitsFromGroup removes specific units from a group
func (gm *GroupManager) RemoveUnitsFromGroup(groupID int, unitIDs []int) error {
	gm.mutex.Lock()
	defer gm.mutex.Unlock()

	group, exists := gm.groups[groupID]
	if !exists {
		return fmt.Errorf("group %d not found", groupID)
	}

	for _, unitID := range unitIDs {
		group.RemoveUnit(unitID)
		delete(gm.unitGroups, unitID)
	}

	// If group is empty, remove it
	if group.IsEmpty() {
		gm.removeGroup(groupID)
	}

	return nil
}

// DisbandGroup disbands a group and removes all units from formation
func (gm *GroupManager) DisbandGroup(groupID int) error {
	gm.mutex.Lock()
	defer gm.mutex.Unlock()

	group, exists := gm.groups[groupID]
	if !exists {
		return fmt.Errorf("group %d not found", groupID)
	}

	// Remove all unit mappings
	for unitID := range group.Units {
		delete(gm.unitGroups, unitID)
	}

	// Remove group
	gm.removeGroup(groupID)

	return nil
}

// GetGroup returns a group by ID
func (gm *GroupManager) GetGroup(groupID int) (*UnitGroup, bool) {
	gm.mutex.RLock()
	defer gm.mutex.RUnlock()

	group, exists := gm.groups[groupID]
	return group, exists
}

// GetUnitGroup returns the group that contains the specified unit
func (gm *GroupManager) GetUnitGroup(unitID int) (*UnitGroup, bool) {
	gm.mutex.RLock()
	defer gm.mutex.RUnlock()

	group, exists := gm.unitGroups[unitID]
	return group, exists
}

// GetPlayerGroups returns all groups belonging to a player
func (gm *GroupManager) GetPlayerGroups(playerID int) []*UnitGroup {
	gm.mutex.RLock()
	defer gm.mutex.RUnlock()

	playerGroupMap, exists := gm.playerGroups[playerID]
	if !exists {
		return []*UnitGroup{}
	}

	groups := make([]*UnitGroup, 0, len(playerGroupMap))
	for _, group := range playerGroupMap {
		groups = append(groups, group)
	}

	return groups
}

// IsUnitInGroup checks if a unit is part of any group
func (gm *GroupManager) IsUnitInGroup(unitID int) bool {
	gm.mutex.RLock()
	defer gm.mutex.RUnlock()

	_, exists := gm.unitGroups[unitID]
	return exists
}

// MoveGroup issues a move command to an entire group
func (gm *GroupManager) MoveGroup(groupID int, target Vector3) error {
	gm.mutex.RLock()
	group, exists := gm.groups[groupID]
	gm.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("group %d not found", groupID)
	}

	// Update group formation target
	group.MoveToPosition(target)

	// Issue individual movement commands to units with formation-aware targets
	for unitID, unit := range group.Units {
		if unit.IsAlive() {
			formationPos, hasPos := group.GetFormationPosition(unitID)
			if hasPos {
				// Create move command to formation position
				command := UnitCommand{
					Type:      CommandMove,
					Target:    &formationPos,
					CreatedAt: time.Now(),
					Parameters: map[string]interface{}{
						"group_id":         groupID,
						"formation_move":   true,
						"formation_target": target,
					},
				}

				// Issue command through world's command processor
				if gm.world != nil && gm.world.commandProcessor != nil {
					gm.world.commandProcessor.IssueCommand(unitID, command)
				}
			}
		}
	}

	return nil
}

// SetGroupFormation changes the formation type for a group
func (gm *GroupManager) SetGroupFormation(groupID int, formationType FormationType) error {
	gm.mutex.RLock()
	group, exists := gm.groups[groupID]
	gm.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("group %d not found", groupID)
	}

	group.SetFormation(formationType)
	return nil
}

// Update updates all groups and their formations
func (gm *GroupManager) Update(deltaTime time.Duration) {
	gm.mutex.RLock()
	defer gm.mutex.RUnlock()

	// Update all groups
	for groupID, group := range gm.groups {
		group.Update(deltaTime)

		// Remove empty groups
		if group.IsEmpty() {
			go func(id int) {
				gm.mutex.Lock()
				defer gm.mutex.Unlock()
				gm.removeGroup(id)
			}(groupID)
		}
	}
}

// CleanupDeadUnits removes dead units from all groups
func (gm *GroupManager) CleanupDeadUnits() {
	gm.mutex.Lock()
	defer gm.mutex.Unlock()

	for _, group := range gm.groups {
		deadUnits := make([]int, 0)

		for unitID, unit := range group.Units {
			if !unit.IsAlive() {
				deadUnits = append(deadUnits, unitID)
			}
		}

		// Remove dead units
		for _, unitID := range deadUnits {
			group.RemoveUnit(unitID)
			delete(gm.unitGroups, unitID)
		}

		// Mark group for removal if empty
		if group.IsEmpty() {
			gm.removeGroup(group.ID)
		}
	}
}

// GetGroupStats returns statistics about groups
func (gm *GroupManager) GetGroupStats() map[string]interface{} {
	gm.mutex.RLock()
	defer gm.mutex.RUnlock()

	stats := map[string]interface{}{
		"total_groups": len(gm.groups),
		"total_units":  len(gm.unitGroups),
	}

	// Count by formation type
	formationCounts := make(map[string]int)
	for _, group := range gm.groups {
		formationType := group.Formation.String()
		formationCounts[formationType]++
	}
	stats["formations"] = formationCounts

	return stats
}

// Helper functions (called with mutex held)

// removeUnitFromAnyGroup removes a unit from whatever group it's in
func (gm *GroupManager) removeUnitFromAnyGroup(unitID int) {
	if group, exists := gm.unitGroups[unitID]; exists {
		group.RemoveUnit(unitID)
		delete(gm.unitGroups, unitID)

		if group.IsEmpty() {
			gm.removeGroup(group.ID)
		}
	}
}

// removeGroup removes a group from all tracking structures
func (gm *GroupManager) removeGroup(groupID int) {
	group, exists := gm.groups[groupID]
	if !exists {
		return
	}

	// Remove from player groups
	if playerGroupMap, exists := gm.playerGroups[group.PlayerID]; exists {
		delete(playerGroupMap, groupID)
		if len(playerGroupMap) == 0 {
			delete(gm.playerGroups, group.PlayerID)
		}
	}

	// Remove from main groups map
	delete(gm.groups, groupID)
}

// Formation-aware pathfinding integration

// GetFormationPath generates a path for a group that considers formation constraints
func (gm *GroupManager) GetFormationPath(groupID int, target Vector3) ([]Vector3, error) {
	group, exists := gm.GetGroup(groupID)
	if !exists {
		return nil, fmt.Errorf("group %d not found", groupID)
	}

	// Generate path using world's pathfinding system
	// For now, use basic pathfinding - could be enhanced for formation-aware pathfinding
	if gm.world != nil && gm.world.pathfindingMgr != nil && group.Leader != nil {
		pathResult, err := gm.world.pathfindingMgr.RequestPath(group.Leader, target)
		if err != nil {
			return nil, fmt.Errorf("pathfinding failed: %w", err)
		}

		return pathResult.Path, nil
	}

	return nil, fmt.Errorf("pathfinding manager not available")
}

// Group command shortcuts for common operations

// CreateLineFormation creates a group in line formation
func (gm *GroupManager) CreateLineFormation(playerID int, units []*GameUnit) (*UnitGroup, error) {
	return gm.CreateGroup(playerID, units, FormationLine)
}

// CreateAttackFormation creates a group in wedge formation for attacks
func (gm *GroupManager) CreateAttackFormation(playerID int, units []*GameUnit) (*UnitGroup, error) {
	return gm.CreateGroup(playerID, units, FormationWedge)
}

// CreateDefenseFormation creates a group in circular formation for defense
func (gm *GroupManager) CreateDefenseFormation(playerID int, units []*GameUnit) (*UnitGroup, error) {
	return gm.CreateGroup(playerID, units, FormationCircle)
}

// CreateScoutFormation creates a group in scattered formation for scouting
func (gm *GroupManager) CreateScoutFormation(playerID int, units []*GameUnit) (*UnitGroup, error) {
	return gm.CreateGroup(playerID, units, FormationScatter)
}

// BatchGroupOperations for performance

// CreateMultipleGroups creates multiple groups in a single operation
func (gm *GroupManager) CreateMultipleGroups(playerID int, unitGroups [][](*GameUnit), formations []FormationType) ([]*UnitGroup, error) {
	if len(unitGroups) != len(formations) {
		return nil, fmt.Errorf("number of unit groups must match number of formations")
	}

	groups := make([]*UnitGroup, 0, len(unitGroups))

	gm.mutex.Lock()
	defer gm.mutex.Unlock()

	for i, units := range unitGroups {
		if len(units) == 0 {
			continue
		}

		// Remove units from existing groups
		for _, unit := range units {
			gm.removeUnitFromAnyGroup(unit.ID)
		}

		// Create group
		groupID := gm.nextGroupID
		gm.nextGroupID++

		group := NewUnitGroup(groupID, playerID, units, formations[i])
		gm.groups[groupID] = group

		// Add to player groups
		if gm.playerGroups[playerID] == nil {
			gm.playerGroups[playerID] = make(map[int]*UnitGroup)
		}
		gm.playerGroups[playerID][groupID] = group

		// Map units to group
		for _, unit := range units {
			gm.unitGroups[unit.ID] = group
		}

		groups = append(groups, group)
	}

	return groups, nil
}