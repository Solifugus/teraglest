package ui

import (
	"fmt"
	"sync"
	"time"

	"teraglest/internal/engine"
)

// SimpleUIManager is a minimal UI manager without ImGui dependencies for testing
type SimpleUIManager struct {
	// Core components
	world *engine.World

	// Input state
	selectedUnits    []*engine.GameUnit
	selectedBuilding *engine.GameBuilding

	// UI state
	showDebugInfo bool

	// Threading
	mutex sync.RWMutex
}

// NewSimpleUIManager creates a new simple UI manager without ImGui
func NewSimpleUIManager(world *engine.World) *SimpleUIManager {
	return &SimpleUIManager{
		world:         world,
		selectedUnits: make([]*engine.GameUnit, 0),
		showDebugInfo: false,
	}
}

// Update updates the UI system
func (ui *SimpleUIManager) Update(deltaTime time.Duration) {
	ui.mutex.Lock()
	defer ui.mutex.Unlock()

	// Simple update logic - no ImGui components to update
	// This is where we would update UI components if they existed
}

// Render renders the UI (minimal implementation)
func (ui *SimpleUIManager) Render() {
	// For now, just log selection changes
	// In a full implementation, this would render UI elements
}

// GetSelectedUnits returns currently selected units
func (ui *SimpleUIManager) GetSelectedUnits() []*engine.GameUnit {
	ui.mutex.RLock()
	defer ui.mutex.RUnlock()

	// Return copy to avoid race conditions
	result := make([]*engine.GameUnit, len(ui.selectedUnits))
	copy(result, ui.selectedUnits)
	return result
}

// GetSelectedBuilding returns currently selected building
func (ui *SimpleUIManager) GetSelectedBuilding() *engine.GameBuilding {
	ui.mutex.RLock()
	defer ui.mutex.RUnlock()

	return ui.selectedBuilding
}

// SelectUnits sets the selected units
func (ui *SimpleUIManager) SelectUnits(units []*engine.GameUnit) {
	ui.mutex.Lock()
	defer ui.mutex.Unlock()

	ui.selectedUnits = make([]*engine.GameUnit, len(units))
	copy(ui.selectedUnits, units)
	ui.selectedBuilding = nil // Clear building selection

	if len(units) > 0 {
		fmt.Printf("Selected %d units\n", len(units))
	}
}

// SelectBuilding sets the selected building
func (ui *SimpleUIManager) SelectBuilding(building *engine.GameBuilding) {
	ui.mutex.Lock()
	defer ui.mutex.Unlock()

	ui.selectedBuilding = building
	ui.selectedUnits = ui.selectedUnits[:0] // Clear unit selection

	if building != nil {
		fmt.Printf("Selected building: %s\n", building.BuildingType)
	}
}

// ClearSelection clears all selections
func (ui *SimpleUIManager) ClearSelection() {
	ui.mutex.Lock()
	defer ui.mutex.Unlock()

	ui.selectedUnits = ui.selectedUnits[:0]
	ui.selectedBuilding = nil
	fmt.Println("Selection cleared")
}

// IssueCommand issues a command to selected units
func (ui *SimpleUIManager) IssueCommand(commandType engine.CommandType, params map[string]interface{}) error {
	ui.mutex.RLock()
	defer ui.mutex.RUnlock()

	if len(ui.selectedUnits) == 0 {
		return fmt.Errorf("no units selected")
	}

	// Create command for each selected unit
	for _, unit := range ui.selectedUnits {
		command := engine.UnitCommand{
			Type:       commandType,
			Parameters: params,
			CreatedAt:  time.Now(),
			IsQueued:   false, // TODO: Check for shift key modifier
		}

		// Issue command through world's command processor
		world := ui.world
		if world == nil {
			return fmt.Errorf("world is nil")
		}

		// Create command processor for this command
		commandProcessor := engine.NewCommandProcessor(world)
		err := commandProcessor.IssueCommand(unit.GetID(), command)
		if err != nil {
			return fmt.Errorf("failed to issue command to unit %d: %w", unit.GetID(), err)
		}
	}

	fmt.Printf("Issued %s command to %d units\n", commandType, len(ui.selectedUnits))
	return nil
}

// IsMouseOverUI returns false for simple UI (no UI elements to check)
func (ui *SimpleUIManager) IsMouseOverUI() bool {
	return false
}

// Cleanup cleans up UI resources (no-op for simple UI)
func (ui *SimpleUIManager) Cleanup() {
	// Nothing to clean up in simple implementation
}