package ui

import (
	"fmt"
	"math"

	"teraglest/internal/engine"
	"teraglest/internal/graphics/renderer"

	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

// InputHandler manages game input events for unit selection and commands
type InputHandler struct {
	world     *engine.World
	uiManager *UIManager

	// Mouse state
	lastMouseX   float64
	lastMouseY   float64
	isDragging   bool
	dragStartX   float64
	dragStartY   float64

	// Selection state
	selectionBox SelectionBox
	isSelecting  bool

	// Camera reference for world coordinate conversion
	camera *renderer.Camera

	// Screen dimensions for coordinate conversion
	screenWidth  int
	screenHeight int
}

// SelectionBox represents a selection rectangle
type SelectionBox struct {
	StartX, StartY float64
	EndX, EndY     float64
	Active         bool
}

// NewInputHandler creates a new input handler
func NewInputHandler(world *engine.World, uiManager *UIManager) *InputHandler {
	return &InputHandler{
		world:     world,
		uiManager: uiManager,
	}
}

// SetCamera sets the camera reference for coordinate conversion
func (ih *InputHandler) SetCamera(camera *renderer.Camera) {
	ih.camera = camera
}

// SetScreenDimensions sets the screen dimensions for coordinate conversion
func (ih *InputHandler) SetScreenDimensions(width, height int) {
	ih.screenWidth = width
	ih.screenHeight = height
}

// getCurrentPlayerID returns the current player's ID (for now, assumes player 1)
func (ih *InputHandler) getCurrentPlayerID() int {
	// TODO: In a full multiplayer implementation, this would determine
	// the actual human player from the world/game state
	// For now, assume the first human player (typically player 1)

	players := ih.world.GetAllPlayers()
	for _, player := range players {
		if !player.IsAI {
			return player.ID
		}
	}

	// Fallback to player 1 if no human players found
	return 1
}

// selectAllPlayerUnits selects all units belonging to the current player
func (ih *InputHandler) selectAllPlayerUnits() {
	playerID := ih.getCurrentPlayerID()
	allUnits := ih.world.ObjectManager.GetUnitsForPlayer(playerID)

	// Filter out dead units
	var livingUnits []*engine.GameUnit
	for _, unit := range allUnits {
		if unit.Health > 0 {
			livingUnits = append(livingUnits, unit)
		}
	}

	ih.uiManager.SelectUnits(livingUnits)
}

// deleteSelectedUnits removes selected units (for testing)
func (ih *InputHandler) deleteSelectedUnits() {
	selectedUnits := ih.uiManager.GetSelectedUnits()
	for _, unit := range selectedUnits {
		unit.Health = 0 // Mark as dead
	}
	ih.uiManager.ClearSelection() // Clear selection
}

// groupSelectedUnits groups selected units (placeholder for future group management)
func (ih *InputHandler) groupSelectedUnits() {
	selectedUnits := ih.uiManager.GetSelectedUnits()
	if len(selectedUnits) > 0 {
		fmt.Printf("Grouped %d units\n", len(selectedUnits))
		// TODO: Implement actual group management
	}
}

// issueHoldCommand makes selected units hold their position
func (ih *InputHandler) issueHoldCommand() {
	selectedUnits := ih.uiManager.GetSelectedUnits()
	if len(selectedUnits) > 0 {
		params := map[string]interface{}{}
		ih.uiManager.IssueCommand(engine.CommandHold, params)
	}
}

// issueStopCommand makes selected units stop their current action
func (ih *InputHandler) issueStopCommand() {
	selectedUnits := ih.uiManager.GetSelectedUnits()
	if len(selectedUnits) > 0 {
		params := map[string]interface{}{}
		ih.uiManager.IssueCommand(engine.CommandStop, params)
	}
}

// HandleMouseButton processes mouse button events
func (ih *InputHandler) HandleMouseButton(window *glfw.Window, button glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
	// Skip if UI wants to capture mouse
	if ih.uiManager.IsMouseOverUI() {
		return
	}

	xpos, ypos := window.GetCursorPos()

	switch button {
	case glfw.MouseButtonLeft:
		if action == glfw.Press {
			ih.handleLeftMousePress(xpos, ypos, mods)
		} else if action == glfw.Release {
			ih.handleLeftMouseRelease(xpos, ypos, mods)
		}

	case glfw.MouseButtonRight:
		if action == glfw.Press {
			ih.handleRightMousePress(xpos, ypos, mods)
		}
	}
}

// HandleMouseMove processes mouse movement events
func (ih *InputHandler) HandleMouseMove(window *glfw.Window, xpos, ypos float64) {
	ih.lastMouseX = xpos
	ih.lastMouseY = ypos

	// Update selection box if dragging
	if ih.isDragging && ih.isSelecting {
		ih.selectionBox.EndX = xpos
		ih.selectionBox.EndY = ypos
		ih.selectionBox.Active = true
	}
}

// HandleKeyboard processes keyboard events
func (ih *InputHandler) HandleKeyboard(window *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if action == glfw.Press || action == glfw.Repeat {
		switch key {
		case glfw.KeyA:
			// Select all units
			if (mods & glfw.ModControl) != 0 {
				ih.selectAllPlayerUnits()
			}
		case glfw.KeyDelete:
			// Delete selected units (for debugging/testing)
			ih.deleteSelectedUnits()
		case glfw.KeyG:
			// Group selected units (for future group management)
			ih.groupSelectedUnits()
		case glfw.KeyH:
			// Hold position command
			ih.issueHoldCommand()
		case glfw.KeyS:
			// Stop command
			ih.issueStopCommand()
		}
	}
}

// handleLeftMousePress handles left mouse button press
func (ih *InputHandler) handleLeftMousePress(xpos, ypos float64, mods glfw.ModifierKey) {
	// Check if shift is held for additive selection
	additive := (mods & glfw.ModShift) != 0

	// Convert screen coordinates to world coordinates
	worldX, worldZ := ih.screenToWorld(xpos, ypos)

	// Try to select unit or building at clicked position
	selectedUnit := ih.findUnitAtPosition(worldX, worldZ)
	selectedBuilding := ih.findBuildingAtPosition(worldX, worldZ)

	if selectedUnit != nil {
		if additive {
			// Add to existing selection
			currentUnits := ih.uiManager.GetSelectedUnits()
			currentUnits = append(currentUnits, selectedUnit)
			ih.uiManager.SelectUnits(currentUnits)
		} else {
			// New selection
			ih.uiManager.SelectUnits([]*engine.GameUnit{selectedUnit})
		}
	} else if selectedBuilding != nil {
		ih.uiManager.SelectBuilding(selectedBuilding)
	} else if !additive {
		// Start drag selection
		ih.startDragSelection(xpos, ypos)
	}
}

// handleLeftMouseRelease handles left mouse button release
func (ih *InputHandler) handleLeftMouseRelease(xpos, ypos float64, mods glfw.ModifierKey) {
	if ih.isDragging && ih.isSelecting {
		ih.finishDragSelection(mods)
	}

	// Reset drag state
	ih.isDragging = false
	ih.isSelecting = false
	ih.selectionBox.Active = false
}

// handleRightMousePress handles right mouse button press (issue commands)
func (ih *InputHandler) handleRightMousePress(xpos, ypos float64, mods glfw.ModifierKey) {
	selectedUnits := ih.uiManager.GetSelectedUnits()
	if len(selectedUnits) == 0 {
		return
	}

	// Convert screen coordinates to world coordinates
	worldX, worldZ := ih.screenToWorld(xpos, ypos)

	// Check if clicking on an enemy unit (attack command)
	targetUnit := ih.findUnitAtPosition(worldX, worldZ)
	if targetUnit != nil && targetUnit.PlayerID != selectedUnits[0].PlayerID {
		// Issue attack command
		params := map[string]interface{}{
			"target_unit": targetUnit,
		}
		ih.uiManager.IssueCommand(engine.CommandAttack, params)
		return
	}

	// Check if clicking on a resource node (gather command)
	resourceNode := ih.findResourceAtPosition(worldX, worldZ)
	if resourceNode != nil {
		params := map[string]interface{}{
			"target_resource": resourceNode,
		}
		ih.uiManager.IssueCommand(engine.CommandGather, params)
		return
	}

	// Check if clicking on a building (could be repair or other interaction)
	targetBuilding := ih.findBuildingAtPosition(worldX, worldZ)
	if targetBuilding != nil && targetBuilding.PlayerID != selectedUnits[0].PlayerID {
		// Attack building
		params := map[string]interface{}{
			"target_building": targetBuilding,
		}
		ih.uiManager.IssueCommand(engine.CommandAttack, params)
		return
	} else if targetBuilding != nil && targetBuilding.Health < targetBuilding.MaxHealth {
		// Repair friendly building
		params := map[string]interface{}{
			"target_building": targetBuilding,
		}
		ih.uiManager.IssueCommand(engine.CommandRepair, params)
		return
	}

	// Default: move command
	queueCommand := (mods & glfw.ModShift) != 0
	params := map[string]interface{}{
		"target_x":     worldX,
		"target_z":     worldZ,
		"queue":        queueCommand,
	}
	ih.uiManager.IssueCommand(engine.CommandMove, params)
}

// startDragSelection begins a drag selection operation
func (ih *InputHandler) startDragSelection(xpos, ypos float64) {
	ih.isDragging = true
	ih.isSelecting = true
	ih.dragStartX = xpos
	ih.dragStartY = ypos
	ih.selectionBox.StartX = xpos
	ih.selectionBox.StartY = ypos
	ih.selectionBox.EndX = xpos
	ih.selectionBox.EndY = ypos
	ih.selectionBox.Active = true
}

// finishDragSelection completes a drag selection operation
func (ih *InputHandler) finishDragSelection(mods glfw.ModifierKey) {
	// Calculate selection rectangle bounds
	minX := math.Min(ih.selectionBox.StartX, ih.selectionBox.EndX)
	maxX := math.Max(ih.selectionBox.StartX, ih.selectionBox.EndX)
	minY := math.Min(ih.selectionBox.StartY, ih.selectionBox.EndY)
	maxY := math.Max(ih.selectionBox.StartY, ih.selectionBox.EndY)

	// Only proceed if drag distance is significant
	dragDistance := math.Sqrt(math.Pow(maxX-minX, 2) + math.Pow(maxY-minY, 2))
	if dragDistance < 5.0 { // Minimum drag distance threshold
		return
	}

	// Convert screen rectangle to world coordinates
	worldMinX, worldMinZ := ih.screenToWorld(minX, maxY) // Note: Y is flipped
	worldMaxX, worldMaxZ := ih.screenToWorld(maxX, minY)

	// Find all units within the selection rectangle
	selectedUnits := ih.findUnitsInRectangle(worldMinX, worldMinZ, worldMaxX, worldMaxZ)

	// Filter to only player's own units (can't select enemy units)
	playerID := ih.getCurrentPlayerID()
	filteredUnits := make([]*engine.GameUnit, 0, len(selectedUnits))
	for _, unit := range selectedUnits {
		if unit.PlayerID == playerID {
			filteredUnits = append(filteredUnits, unit)
		}
	}

	// Apply selection
	additive := (mods & glfw.ModShift) != 0
	if additive && len(filteredUnits) > 0 {
		// Add to existing selection
		currentUnits := ih.uiManager.GetSelectedUnits()
		currentUnits = append(currentUnits, filteredUnits...)
		ih.uiManager.SelectUnits(currentUnits)
	} else if len(filteredUnits) > 0 {
		// New selection
		ih.uiManager.SelectUnits(filteredUnits)
	}
}

// screenToWorld converts screen coordinates to world coordinates using camera ray casting
func (ih *InputHandler) screenToWorld(screenX, screenY float64) (float64, float64) {
	// Check if camera is available
	if ih.camera == nil {
		// Fallback to simple conversion if no camera
		worldX := screenX / 10.0
		worldZ := screenY / 10.0
		return worldX, worldZ
	}

	// Use camera to convert screen coordinates to world ray
	rayOrigin, rayDirection := ih.camera.ScreenToWorldRay(
		int(screenX), int(screenY),
		ih.screenWidth, ih.screenHeight)

	// Find intersection with ground plane (Y = 0)
	// Ray equation: P = origin + t * direction
	// Ground plane: Y = 0
	// Solve: origin.Y + t * direction.Y = 0
	// Therefore: t = -origin.Y / direction.Y

	if math.Abs(float64(rayDirection.Y())) < 0.0001 {
		// Ray is parallel to ground plane, use camera position projection
		worldX := float64(rayOrigin.X())
		worldZ := float64(rayOrigin.Z())
		return worldX, worldZ
	}

	t := -rayOrigin.Y() / rayDirection.Y()

	// Calculate intersection point
	intersectionPoint := rayOrigin.Add(rayDirection.Mul(t))

	worldX := float64(intersectionPoint.X())
	worldZ := float64(intersectionPoint.Z())

	return worldX, worldZ
}

// findUnitAtPosition finds a unit at the given world position
func (ih *InputHandler) findUnitAtPosition(worldX, worldZ float64) *engine.GameUnit {
	// Search radius for unit selection
	searchRadius := 1.0

	// Get all units from all players
	for playerID := range ih.world.GetPlayers() {
		units := ih.world.ObjectManager.GetUnitsForPlayer(playerID)
		for _, unit := range units {
			if unit.IsAlive() {
				// Calculate distance to unit
				dx := unit.Position.X - worldX
				dz := unit.Position.Z - worldZ
				distance := math.Sqrt(dx*dx + dz*dz)

				if distance <= searchRadius {
					return unit
				}
			}
		}
	}

	return nil
}

// findBuildingAtPosition finds a building at the given world position
func (ih *InputHandler) findBuildingAtPosition(worldX, worldZ float64) *engine.GameBuilding {
	// Search radius for building selection (larger than units)
	searchRadius := 2.0

	// Get all buildings from all players
	for playerID := range ih.world.GetPlayers() {
		buildings := ih.world.ObjectManager.GetBuildingsForPlayer(playerID)
		for _, building := range buildings {
			if building.IsAlive() {
				// Calculate distance to building
				dx := building.Position.X - worldX
				dz := building.Position.Z - worldZ
				distance := math.Sqrt(dx*dx + dz*dz)

				if distance <= searchRadius {
					return building
				}
			}
		}
	}

	return nil
}

// findResourceAtPosition finds a resource node at the given world position
func (ih *InputHandler) findResourceAtPosition(worldX, worldZ float64) *engine.ResourceNode {
	// Search radius for resource selection
	searchRadius := 1.5

	// Get all resources
	resources := ih.world.GetResources()
	for _, resource := range resources {
		if resource.Amount > 0 {
			// Calculate distance to resource
			dx := resource.Position.X - worldX
			dz := resource.Position.Z - worldZ
			distance := math.Sqrt(dx*dx + dz*dz)

			if distance <= searchRadius {
				return resource
			}
		}
	}

	return nil
}

// findUnitsInRectangle finds all units within a rectangular area
func (ih *InputHandler) findUnitsInRectangle(minX, minZ, maxX, maxZ float64) []*engine.GameUnit {
	var selectedUnits []*engine.GameUnit

	// Ensure min/max are ordered correctly
	if minX > maxX {
		minX, maxX = maxX, minX
	}
	if minZ > maxZ {
		minZ, maxZ = maxZ, minZ
	}

	// Search all units
	for playerID := range ih.world.GetPlayers() {
		units := ih.world.ObjectManager.GetUnitsForPlayer(playerID)
		for _, unit := range units {
			if unit.IsAlive() {
				// Check if unit is within rectangle
				if unit.Position.X >= minX && unit.Position.X <= maxX &&
					unit.Position.Z >= minZ && unit.Position.Z <= maxZ {
					selectedUnits = append(selectedUnits, unit)
				}
			}
		}
	}

	return selectedUnits
}

// GetSelectionBox returns the current selection box for rendering
func (ih *InputHandler) GetSelectionBox() SelectionBox {
	return ih.selectionBox
}

