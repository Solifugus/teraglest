package ui

import (
	"fmt"
	"sync"
	"time"

	"teraglest/internal/engine"
	"teraglest/internal/graphics/renderer"

	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/inkyblackness/imgui-go/v4"
)

// UIManager manages the game's user interface
type UIManager struct {
	// Core components
	world        *engine.World
	renderer     *renderer.Renderer
	imgui        *imgui.Context
	imguiRenderer *ImGuiRenderer

	// UI components
	resourceDisplay *ResourceDisplay
	minimap        *Minimap
	unitPanel      *UnitPanel
	commandPanel   *CommandPanel
	productionUI   *ProductionUI

	// Input state
	selectedUnits   []*engine.GameUnit
	selectedBuilding *engine.GameBuilding
	mousePos        [2]float32

	// UI state
	showDebugInfo   bool
	showMinimap     bool
	showResourceBar bool
	uiScale         float32

	// Threading
	mutex          sync.RWMutex
}

// NewUIManager creates a new UI manager
func NewUIManager(world *engine.World, renderer *renderer.Renderer) (*UIManager, error) {
	// Create ImGui context
	imgui := imgui.CreateContext(nil)
	if imgui == nil {
		return nil, fmt.Errorf("failed to create ImGui context")
	}

	// ImGui configuration simplified for compatibility

	ui := &UIManager{
		world:           world,
		renderer:        renderer,
		imgui:          imgui,
		selectedUnits:   make([]*engine.GameUnit, 0),
		showDebugInfo:   false,
		showMinimap:     true,
		showResourceBar: true,
		uiScale:         1.0,
	}

	// Initialize ImGui renderer
	imguiRenderer, err := NewImGuiRenderer()
	if err != nil {
		// Simplified for compatibility
		return nil, fmt.Errorf("failed to create ImGui renderer: %w", err)
	}
	ui.imguiRenderer = imguiRenderer

	// Initialize UI components
	if err := ui.initializeComponents(); err != nil {
		ui.imguiRenderer.Cleanup()
		// Simplified for compatibility
		return nil, fmt.Errorf("failed to initialize UI components: %w", err)
	}

	return ui, nil
}

// initializeComponents creates all UI component instances
func (ui *UIManager) initializeComponents() error {
	var err error

	// Create resource display
	ui.resourceDisplay, err = NewResourceDisplay(ui.world)
	if err != nil {
		return fmt.Errorf("failed to create resource display: %w", err)
	}

	// Create minimap
	ui.minimap, err = NewMinimap(ui.world, 256, 256)
	if err != nil {
		return fmt.Errorf("failed to create minimap: %w", err)
	}

	// Create unit panel
	ui.unitPanel = NewUnitPanel(ui.world)

	// Create command panel
	ui.commandPanel = NewCommandPanel(ui.world, ui)

	// Create production UI
	ui.productionUI = NewProductionUI(ui.world, ui)

	return nil
}

// Update updates the UI system
func (ui *UIManager) Update(deltaTime time.Duration) {
	ui.mutex.Lock()
	defer ui.mutex.Unlock()

	// Update ImGui frame timing
	io := imgui.CurrentIO()
	io.SetDeltaTime(float32(deltaTime.Seconds()))

	// Update UI components
	ui.resourceDisplay.Update(deltaTime)
	ui.minimap.Update(deltaTime)
	ui.unitPanel.Update(deltaTime, ui.selectedUnits, ui.selectedBuilding)
	ui.commandPanel.Update(deltaTime, ui.selectedUnits, ui.selectedBuilding)
	ui.productionUI.Update(deltaTime, ui.selectedBuilding)

	// Update selection based on user input
	ui.updateSelection()
}

// Render renders the UI overlay
func (ui *UIManager) Render() {
	ui.mutex.RLock()
	defer ui.mutex.RUnlock()

	// Start new ImGui frame
	imgui.NewFrame()

	// Render UI components based on visibility settings
	if ui.showResourceBar {
		ui.resourceDisplay.Render()
	}

	if ui.showMinimap {
		ui.minimap.Render()
	}

	// Always render unit/building panels if something is selected
	if len(ui.selectedUnits) > 0 {
		ui.unitPanel.RenderWithSelection(ui.selectedUnits, nil)
		ui.commandPanel.RenderWithSelection(ui.selectedUnits, nil)
	}

	if ui.selectedBuilding != nil {
		ui.unitPanel.RenderWithSelection(nil, ui.selectedBuilding)
		ui.commandPanel.RenderWithSelection(nil, ui.selectedBuilding)
		ui.productionUI.RenderWithBuilding(ui.selectedBuilding)
	}

	// Debug information
	if ui.showDebugInfo {
		ui.renderDebugInfo()
	}

	// Render ImGui frame
	imgui.Render()

	// Render with OpenGL backend
	if ui.imguiRenderer != nil {
		ui.imguiRenderer.Render(imgui.RenderedDrawData())
	}
}

// ProcessInput handles input events for the UI
func (ui *UIManager) ProcessInput(window *glfw.Window) {
	ui.mutex.Lock()
	defer ui.mutex.Unlock()

	// Get mouse position
	xpos, ypos := window.GetCursorPos()
	ui.mousePos[0] = float32(xpos)
	ui.mousePos[1] = float32(ypos)

	// Input handling simplified for compatibility - these would normally update ImGui IO

	// Handle UI toggle keys
	if window.GetKey(glfw.KeyF1) == glfw.Press {
		ui.showDebugInfo = !ui.showDebugInfo
	}
	if window.GetKey(glfw.KeyM) == glfw.Press {
		ui.showMinimap = !ui.showMinimap
	}
}

// OnResize handles window resize events
func (ui *UIManager) OnResize(width, height int) {
	ui.mutex.Lock()
	defer ui.mutex.Unlock()

	// Update ImGui display size
	io := imgui.CurrentIO()
	io.SetDisplaySize(imgui.Vec2{X: float32(width), Y: float32(height)})

	// Update component sizes
	ui.minimap.OnResize(width, height)
}

// updateSelection handles unit/building selection logic
func (ui *UIManager) updateSelection() {
	// This will be called from mouse input processing
	// For now, it's a placeholder for selection logic integration
}

// GetSelectedUnits returns currently selected units
func (ui *UIManager) GetSelectedUnits() []*engine.GameUnit {
	ui.mutex.RLock()
	defer ui.mutex.RUnlock()

	// Return copy to avoid race conditions
	result := make([]*engine.GameUnit, len(ui.selectedUnits))
	copy(result, ui.selectedUnits)
	return result
}

// GetSelectedBuilding returns currently selected building
func (ui *UIManager) GetSelectedBuilding() *engine.GameBuilding {
	ui.mutex.RLock()
	defer ui.mutex.RUnlock()

	return ui.selectedBuilding
}

// SelectUnits sets the selected units
func (ui *UIManager) SelectUnits(units []*engine.GameUnit) {
	ui.mutex.Lock()
	defer ui.mutex.Unlock()

	ui.selectedUnits = make([]*engine.GameUnit, len(units))
	copy(ui.selectedUnits, units)
	ui.selectedBuilding = nil // Clear building selection
}

// SelectBuilding sets the selected building
func (ui *UIManager) SelectBuilding(building *engine.GameBuilding) {
	ui.mutex.Lock()
	defer ui.mutex.Unlock()

	ui.selectedBuilding = building
	ui.selectedUnits = ui.selectedUnits[:0] // Clear unit selection
}

// ClearSelection clears all selections
func (ui *UIManager) ClearSelection() {
	ui.mutex.Lock()
	defer ui.mutex.Unlock()

	ui.selectedUnits = ui.selectedUnits[:0]
	ui.selectedBuilding = nil
}

// IssueCommand issues a command to selected units
func (ui *UIManager) IssueCommand(commandType engine.CommandType, params map[string]interface{}) error {
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

		// Issue command through command processor
		commandProcessorInterface := ui.world.GetCommandProcessor()
		commandProcessor, ok := commandProcessorInterface.(*engine.CommandProcessor)
		if !ok {
			return fmt.Errorf("failed to get command processor for unit command")
		}
		err := commandProcessor.IssueCommand(unit.GetID(), command)
		if err != nil {
			return fmt.Errorf("failed to issue command to unit %d: %w", unit.GetID(), err)
		}
	}

	return nil
}

// renderDebugInfo renders debug information overlay
func (ui *UIManager) renderDebugInfo() {
	imgui.Begin("Debug Information")

	// World statistics
	imgui.Text(fmt.Sprintf("Game Time: %.1f seconds", ui.world.GetGameTime().Seconds()))

	// Object counts
	objectStats := ui.world.ObjectManager.GetStats()
	imgui.Text(fmt.Sprintf("Total Units: %d", objectStats.TotalUnits))
	imgui.Text(fmt.Sprintf("Total Buildings: %d", objectStats.TotalBuildings))

	// Player information
	for playerID := range ui.world.GetPlayers() {
		player := ui.world.GetPlayer(playerID)
		if player != nil {
			if imgui.TreeNode(fmt.Sprintf("Player %d (%s)", playerID, player.Name)) {
				// Resources
				for resource, amount := range player.Resources {
					imgui.Text(fmt.Sprintf("%s: %d", resource, amount))
				}

				// Population if available
				if ui.world.GetProductionSystem() != nil {
					popMgr := ui.world.GetProductionSystem().GetPopulationManager()
					popStats := popMgr.GetPopulationStatus(playerID)
					imgui.Text(fmt.Sprintf("Population: %d/%d", popStats.CurrentPopulation, popStats.MaxPopulation))
				}

				imgui.TreePop()
			}
		}
	}

	// Selection info
	imgui.Separator()
	imgui.Text(fmt.Sprintf("Selected Units: %d", len(ui.selectedUnits)))
	if ui.selectedBuilding != nil {
		imgui.Text(fmt.Sprintf("Selected Building: %s", ui.selectedBuilding.BuildingType))
	}

	imgui.End()
}

// Cleanup cleans up UI resources
func (ui *UIManager) Cleanup() {
	ui.mutex.Lock()
	defer ui.mutex.Unlock()

	// Cleanup components
	if ui.resourceDisplay != nil {
		ui.resourceDisplay.Cleanup()
	}
	if ui.minimap != nil {
		ui.minimap.Cleanup()
	}

	// Cleanup ImGui renderer
	if ui.imguiRenderer != nil {
		ui.imguiRenderer.Cleanup()
	}

	// Destroy ImGui context
	if ui.imgui != nil {
		imgui.DestroyContext(ui.imgui)
	}
}

// IsMouseOverUI returns true if the mouse is over a UI element
func (ui *UIManager) IsMouseOverUI() bool {
	io := imgui.CurrentIO()
	return io.WantCaptureMouse()
}

// SetUIScale sets the UI scaling factor
func (ui *UIManager) SetUIScale(scale float32) {
	ui.mutex.Lock()
	defer ui.mutex.Unlock()

	ui.uiScale = scale

	// Update ImGui font scale
	imgui.GetStyle().SetScaleAllSizes(scale)
}

// GetWorld returns the world instance for UI components
func (ui *UIManager) GetWorld() *engine.World {
	return ui.world
}

// GetCommandProcessor returns the command processor for issuing commands
func (ui *UIManager) GetCommandProcessor() interface{} {
	return ui.world.GetCommandProcessor()
}