package engine

import (
	"fmt"
	"time"
)

// GroupCommand represents a command that operates on multiple units as a group
type GroupCommand struct {
	Type        GroupCommandType       `json:"type"`
	GroupID     int                    `json:"group_id"`
	UnitIDs     []int                  `json:"unit_ids"`
	Target      *Vector3               `json:"target"`
	Formation   FormationType          `json:"formation"`
	Parameters  map[string]interface{} `json:"parameters"`
	CreatedAt   time.Time              `json:"created_at"`
	Priority    int                    `json:"priority"`
}

// GroupCommandType represents different types of group commands
type GroupCommandType int

const (
	GroupCommandCreate     GroupCommandType = iota // Create a new group
	GroupCommandMove                               // Move group to location
	GroupCommandAttack                             // Group attack command
	GroupCommandDefend                             // Group defensive stance
	GroupCommandSetFormation                       // Change group formation
	GroupCommandDisband                            // Disband the group
	GroupCommandAddUnits                           // Add units to group
	GroupCommandRemoveUnits                        // Remove units from group
)

// String returns the string representation of GroupCommandType
func (gct GroupCommandType) String() string {
	switch gct {
	case GroupCommandCreate:
		return "Create"
	case GroupCommandMove:
		return "Move"
	case GroupCommandAttack:
		return "Attack"
	case GroupCommandDefend:
		return "Defend"
	case GroupCommandSetFormation:
		return "SetFormation"
	case GroupCommandDisband:
		return "Disband"
	case GroupCommandAddUnits:
		return "AddUnits"
	case GroupCommandRemoveUnits:
		return "RemoveUnits"
	default:
		return "Unknown"
	}
}

// Group command processing methods for CommandProcessor

// IssueGroupCommand processes a group command
func (cp *CommandProcessor) IssueGroupCommand(playerID int, command GroupCommand) error {
	command.CreatedAt = time.Now()

	switch command.Type {
	case GroupCommandCreate:
		return cp.processCreateGroupCommand(playerID, command)
	case GroupCommandMove:
		return cp.processGroupMoveCommand(playerID, command)
	case GroupCommandAttack:
		return cp.processGroupAttackCommand(playerID, command)
	case GroupCommandSetFormation:
		return cp.processSetFormationCommand(playerID, command)
	case GroupCommandDisband:
		return cp.processDisbandGroupCommand(playerID, command)
	case GroupCommandAddUnits:
		return cp.processAddUnitsCommand(playerID, command)
	case GroupCommandRemoveUnits:
		return cp.processRemoveUnitsCommand(playerID, command)
	default:
		return fmt.Errorf("unknown group command type: %v", command.Type)
	}
}

// processCreateGroupCommand creates a new unit group
func (cp *CommandProcessor) processCreateGroupCommand(playerID int, command GroupCommand) error {
	if cp.world.groupMgr == nil {
		return fmt.Errorf("group manager not initialized")
	}

	// Get units by IDs
	units := make([]*GameUnit, 0, len(command.UnitIDs))
	for _, unitID := range command.UnitIDs {
		unit := cp.world.ObjectManager.GetUnit(unitID)
		if unit == nil {
			continue // Skip invalid unit IDs
		}
		if unit.PlayerID != playerID {
			continue // Skip units not owned by player
		}
		units = append(units, unit)
	}

	if len(units) == 0 {
		return fmt.Errorf("no valid units provided for group creation")
	}

	// Create group with specified formation
	formation := FormationLine // Default formation
	if command.Formation >= FormationLine && command.Formation <= FormationCustom {
		formation = command.Formation
	}

	_, err := cp.world.groupMgr.CreateGroup(playerID, units, formation)
	return err
}

// processGroupMoveCommand moves an entire group to a target location
func (cp *CommandProcessor) processGroupMoveCommand(playerID int, command GroupCommand) error {
	if cp.world.groupMgr == nil {
		return fmt.Errorf("group manager not initialized")
	}

	if command.Target == nil {
		return fmt.Errorf("group move command requires target position")
	}

	// Verify group ownership
	group, exists := cp.world.groupMgr.GetGroup(command.GroupID)
	if !exists {
		return fmt.Errorf("group %d not found", command.GroupID)
	}

	if group.PlayerID != playerID {
		return fmt.Errorf("player %d does not own group %d", playerID, command.GroupID)
	}

	// Execute group movement
	return cp.world.groupMgr.MoveGroup(command.GroupID, *command.Target)
}

// processGroupAttackCommand commands a group to attack a target
func (cp *CommandProcessor) processGroupAttackCommand(playerID int, command GroupCommand) error {
	if cp.world.groupMgr == nil {
		return fmt.Errorf("group manager not initialized")
	}

	// Get target position
	var targetPos Vector3
	if command.Target != nil {
		targetPos = *command.Target
	} else {
		return fmt.Errorf("group attack command requires target")
	}

	// Verify group ownership
	group, exists := cp.world.groupMgr.GetGroup(command.GroupID)
	if !exists {
		return fmt.Errorf("group %d not found", command.GroupID)
	}

	if group.PlayerID != playerID {
		return fmt.Errorf("player %d does not own group %d", playerID, command.GroupID)
	}

	// Issue attack commands to all units in the group
	for _, unit := range group.Units {
		if unit.IsAlive() {
			attackCommand := UnitCommand{
				Type:      CommandAttack,
				Target:    &targetPos,
				CreatedAt: time.Now(),
				Parameters: map[string]interface{}{
					"group_id":     command.GroupID,
					"group_attack": true,
				},
			}

			err := cp.IssueCommand(unit.ID, attackCommand)
			if err != nil {
				// Log error but continue with other units
				continue
			}
		}
	}

	return nil
}

// processSetFormationCommand changes a group's formation
func (cp *CommandProcessor) processSetFormationCommand(playerID int, command GroupCommand) error {
	if cp.world.groupMgr == nil {
		return fmt.Errorf("group manager not initialized")
	}

	// Verify group ownership
	group, exists := cp.world.groupMgr.GetGroup(command.GroupID)
	if !exists {
		return fmt.Errorf("group %d not found", command.GroupID)
	}

	if group.PlayerID != playerID {
		return fmt.Errorf("player %d does not own group %d", playerID, command.GroupID)
	}

	// Set new formation
	return cp.world.groupMgr.SetGroupFormation(command.GroupID, command.Formation)
}

// processDisbandGroupCommand disbands a group
func (cp *CommandProcessor) processDisbandGroupCommand(playerID int, command GroupCommand) error {
	if cp.world.groupMgr == nil {
		return fmt.Errorf("group manager not initialized")
	}

	// Verify group ownership
	group, exists := cp.world.groupMgr.GetGroup(command.GroupID)
	if !exists {
		return fmt.Errorf("group %d not found", command.GroupID)
	}

	if group.PlayerID != playerID {
		return fmt.Errorf("player %d does not own group %d", playerID, command.GroupID)
	}

	// Disband the group
	return cp.world.groupMgr.DisbandGroup(command.GroupID)
}

// processAddUnitsCommand adds units to an existing group
func (cp *CommandProcessor) processAddUnitsCommand(playerID int, command GroupCommand) error {
	if cp.world.groupMgr == nil {
		return fmt.Errorf("group manager not initialized")
	}

	// Verify group ownership
	group, exists := cp.world.groupMgr.GetGroup(command.GroupID)
	if !exists {
		return fmt.Errorf("group %d not found", command.GroupID)
	}

	if group.PlayerID != playerID {
		return fmt.Errorf("player %d does not own group %d", playerID, command.GroupID)
	}

	// Get valid units to add
	units := make([]*GameUnit, 0, len(command.UnitIDs))
	for _, unitID := range command.UnitIDs {
		unit := cp.world.ObjectManager.GetUnit(unitID)
		if unit == nil {
			continue // Skip invalid unit IDs
		}
		if unit.PlayerID != playerID {
			continue // Skip units not owned by player
		}
		units = append(units, unit)
	}

	// Add units to group
	return cp.world.groupMgr.AddUnitsToGroup(command.GroupID, units)
}

// processRemoveUnitsCommand removes units from a group
func (cp *CommandProcessor) processRemoveUnitsCommand(playerID int, command GroupCommand) error {
	if cp.world.groupMgr == nil {
		return fmt.Errorf("group manager not initialized")
	}

	// Verify group ownership
	group, exists := cp.world.groupMgr.GetGroup(command.GroupID)
	if !exists {
		return fmt.Errorf("group %d not found", command.GroupID)
	}

	if group.PlayerID != playerID {
		return fmt.Errorf("player %d does not own group %d", playerID, command.GroupID)
	}

	// Remove units from group
	return cp.world.groupMgr.RemoveUnitsFromGroup(command.GroupID, command.UnitIDs)
}

// Helper functions for creating common group commands

// CreateGroupCommand creates a command to form a new group
func CreateGroupCommand(unitIDs []int, formation FormationType) GroupCommand {
	return GroupCommand{
		Type:      GroupCommandCreate,
		UnitIDs:   unitIDs,
		Formation: formation,
		CreatedAt: time.Now(),
	}
}

// CreateGroupMoveCommand creates a command to move a group
func CreateGroupMoveCommand(groupID int, target Vector3) GroupCommand {
	return GroupCommand{
		Type:      GroupCommandMove,
		GroupID:   groupID,
		Target:    &target,
		CreatedAt: time.Now(),
	}
}

// CreateGroupAttackCommand creates a command for group attack
func CreateGroupAttackCommand(groupID int, target Vector3) GroupCommand {
	return GroupCommand{
		Type:      GroupCommandAttack,
		GroupID:   groupID,
		Target:    &target,
		CreatedAt: time.Now(),
	}
}

// CreateSetFormationCommand creates a command to change group formation
func CreateSetFormationCommand(groupID int, formation FormationType) GroupCommand {
	return GroupCommand{
		Type:      GroupCommandSetFormation,
		GroupID:   groupID,
		Formation: formation,
		CreatedAt: time.Now(),
	}
}

// CreateDisbandGroupCommand creates a command to disband a group
func CreateDisbandGroupCommand(groupID int) GroupCommand {
	return GroupCommand{
		Type:      GroupCommandDisband,
		GroupID:   groupID,
		CreatedAt: time.Now(),
	}
}

// Convenience methods for World

// CreateGroup creates a new unit group with specified formation
func (w *World) CreateGroup(playerID int, unitIDs []int, formation FormationType) (*UnitGroup, error) {
	if w.groupMgr == nil {
		return nil, fmt.Errorf("group manager not initialized")
	}

	// Get units by IDs
	units := make([]*GameUnit, 0, len(unitIDs))
	for _, unitID := range unitIDs {
		unit := w.ObjectManager.GetUnit(unitID)
		if unit == nil {
			continue
		}
		if unit.PlayerID != playerID {
			continue
		}
		units = append(units, unit)
	}

	if len(units) == 0 {
		return nil, fmt.Errorf("no valid units provided")
	}

	return w.groupMgr.CreateGroup(playerID, units, formation)
}

// MoveGroup moves an entire group to a target position
func (w *World) MoveGroup(playerID int, groupID int, target Vector3) error {
	if w.groupMgr == nil {
		return fmt.Errorf("group manager not initialized")
	}

	// Verify ownership
	group, exists := w.groupMgr.GetGroup(groupID)
	if !exists {
		return fmt.Errorf("group %d not found", groupID)
	}

	if group.PlayerID != playerID {
		return fmt.Errorf("player %d does not own group %d", playerID, groupID)
	}

	return w.groupMgr.MoveGroup(groupID, target)
}

// GetPlayerGroups returns all groups for a player
func (w *World) GetPlayerGroups(playerID int) []*UnitGroup {
	if w.groupMgr == nil {
		return []*UnitGroup{}
	}

	return w.groupMgr.GetPlayerGroups(playerID)
}

// GetUnitGroup returns the group containing a specific unit
func (w *World) GetUnitGroup(unitID int) (*UnitGroup, bool) {
	if w.groupMgr == nil {
		return nil, false
	}

	return w.groupMgr.GetUnitGroup(unitID)
}

// IsUnitInGroup checks if a unit is part of any group
func (w *World) IsUnitInGroup(unitID int) bool {
	if w.groupMgr == nil {
		return false
	}

	return w.groupMgr.IsUnitInGroup(unitID)
}