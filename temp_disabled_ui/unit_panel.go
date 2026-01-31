package ui

import (
	"fmt"
	"time"

	"teraglest/internal/engine"

	"github.com/inkyblackness/imgui-go/v4"
)

// UnitPanel displays information about selected units
type UnitPanel struct {
	world *engine.World

	// Display settings
	showDetailedStats bool
	showUnitPortraits bool
	compactMode       bool
}

// NewUnitPanel creates a new unit panel
func NewUnitPanel(world *engine.World) *UnitPanel {
	return &UnitPanel{
		world:             world,
		showDetailedStats: true,
		showUnitPortraits: false, // Portraits not implemented yet
		compactMode:       false,
	}
}

// Update updates the unit panel state
func (up *UnitPanel) Update(deltaTime time.Duration, selectedUnits []*engine.GameUnit, selectedBuilding *engine.GameBuilding) {
	// Panel updates based on selection changes
	// For now, no active updates needed
}

// Render renders the unit panel
func (up *UnitPanel) Render() {
	// Position at bottom-left of screen
	displaySize := imgui.CurrentIO().DisplaySize()
	panelWidth := float32(350)
	panelHeight := float32(200)

	imgui.SetNextWindowPos(imgui.Vec2{
		X: 20,
		Y: displaySize.Y - panelHeight - 20,
	})
	imgui.SetNextWindowSize(imgui.Vec2{X: panelWidth, Y: panelHeight})

	flags := imgui.WindowFlagsNoResize | imgui.WindowFlagsNoMove |
		imgui.WindowFlagsNoCollapse

	if imgui.Begin("Unit Information", nil, flags) {
		up.renderUnitContent()
	}
	imgui.End()
}

// RenderWithSelection renders the unit panel with specific selection
func (up *UnitPanel) RenderWithSelection(selectedUnits []*engine.GameUnit, selectedBuilding *engine.GameBuilding) {
	// Position at bottom-left of screen
	displaySize := imgui.CurrentIO().DisplaySize()
	panelWidth := float32(350)
	panelHeight := float32(200)

	imgui.SetNextWindowPos(imgui.Vec2{
		X: 20,
		Y: displaySize.Y - panelHeight - 20,
	})
	imgui.SetNextWindowSize(imgui.Vec2{X: panelWidth, Y: panelHeight})

	flags := imgui.WindowFlagsNoResize | imgui.WindowFlagsNoMove |
		imgui.WindowFlagsNoCollapse

	if imgui.Begin("Selection", nil, flags) {
		if len(selectedUnits) > 0 {
			up.renderUnitsInfo(selectedUnits)
		} else if selectedBuilding != nil {
			up.renderBuildingInfo(selectedBuilding)
		} else {
			imgui.Text("No selection")
		}
	}
	imgui.End()
}

// renderUnitContent renders the main unit panel content (legacy method)
func (up *UnitPanel) renderUnitContent() {
	imgui.Text("Select units to view information")
	imgui.Separator()

	// Settings
	imgui.Checkbox("Detailed Stats", &up.showDetailedStats)
	imgui.Checkbox("Compact Mode", &up.compactMode)
}

// renderUnitsInfo renders information about selected units
func (up *UnitPanel) renderUnitsInfo(selectedUnits []*engine.GameUnit) {
	if len(selectedUnits) == 1 {
		up.renderSingleUnitInfo(selectedUnits[0])
	} else {
		up.renderMultipleUnitsInfo(selectedUnits)
	}
}

// renderSingleUnitInfo renders detailed info for a single selected unit
func (up *UnitPanel) renderSingleUnitInfo(unit *engine.GameUnit) {
	// Unit name and type
	imgui.Text(fmt.Sprintf("%s (%s)", unit.Name, unit.UnitType))
	imgui.Separator()

	// Health bar
	healthPercent := float32(unit.Health) / float32(unit.MaxHealth)
	healthColor := imgui.Vec4{X: 1.0 - healthPercent, Y: healthPercent, Z: 0.0, W: 1.0}

	imgui.PushStyleColorVec4(imgui.StyleColorPlotHistogram, healthColor)
	imgui.ProgressBar(healthPercent, imgui.Vec2{X: -1, Y: 0}, fmt.Sprintf("Health: %d/%d", unit.Health, unit.MaxHealth))
	imgui.PopStyleColor()

	// Energy bar (if unit has energy)
	if unit.MaxEnergy > 0 {
		energyPercent := float32(unit.Energy) / float32(unit.MaxEnergy)
		energyColor := imgui.Vec4{X: 0.0, Y: 0.4, Z: 1.0, W: 1.0} // Blue

		imgui.PushStyleColorVec4(imgui.StyleColorPlotHistogram, energyColor)
		imgui.ProgressBar(energyPercent, imgui.Vec2{X: -1, Y: 0}, fmt.Sprintf("Energy: %d/%d", unit.Energy, unit.MaxEnergy))
		imgui.PopStyleColor()
	}

	// Unit state and current action
	imgui.Text(fmt.Sprintf("State: %s", unit.State.String()))

	if unit.CurrentCommand != nil {
		imgui.Text(fmt.Sprintf("Command: %s", unit.CurrentCommand.Type.String()))
	}

	// Position information
	imgui.Text(fmt.Sprintf("Position: (%.1f, %.1f, %.1f)", unit.Position.X, unit.Position.Y, unit.Position.Z))

	// Combat stats if detailed view enabled
	if up.showDetailedStats {
		imgui.Separator()
		up.renderUnitCombatStats(unit)
	}

	// Resource carrying (if applicable)
	if len(unit.CarriedResources) > 0 {
		imgui.Separator()
		imgui.Text("Carrying:")
		for resource, amount := range unit.CarriedResources {
			if amount > 0 {
				imgui.Text(fmt.Sprintf("  %s: %d", resource, amount))
			}
		}
	}

	// Command queue info
	if len(unit.CommandQueue) > 0 {
		imgui.Separator()
		imgui.Text(fmt.Sprintf("Queued commands: %d", len(unit.CommandQueue)))

		// Show next few commands
		for i, cmd := range unit.CommandQueue {
			if i >= 3 { // Only show first 3 queued commands
				break
			}
			imgui.Text(fmt.Sprintf("  %d. %s", i+1, cmd.Type.String()))
		}
	}
}

// renderMultipleUnitsInfo renders info for multiple selected units
func (up *UnitPanel) renderMultipleUnitsInfo(selectedUnits []*engine.GameUnit) {
	imgui.Text(fmt.Sprintf("Selected: %d units", len(selectedUnits)))
	imgui.Separator()

	// Group units by type
	unitTypes := make(map[string]int)
	totalHealth := 0
	maxTotalHealth := 0
	aliveUnits := 0

	for _, unit := range selectedUnits {
		if unit.IsAlive() {
			unitTypes[unit.UnitType]++
			totalHealth += unit.Health
			maxTotalHealth += unit.MaxHealth
			aliveUnits++
		}
	}

	// Show unit type breakdown
	imgui.Text("Unit Types:")
	for unitType, count := range unitTypes {
		imgui.Text(fmt.Sprintf("  %s: %d", unitType, count))
	}

	// Overall health
	if aliveUnits > 0 {
		healthPercent := float32(totalHealth) / float32(maxTotalHealth)
		healthColor := imgui.Vec4{X: 1.0 - healthPercent, Y: healthPercent, Z: 0.0, W: 1.0}

		imgui.Separator()
		imgui.PushStyleColorVec4(imgui.StyleColorPlotHistogram, healthColor)
		imgui.ProgressBar(healthPercent, imgui.Vec2{X: -1, Y: 0}, fmt.Sprintf("Total Health: %d/%d", totalHealth, maxTotalHealth))
		imgui.PopStyleColor()
	}

	// Common commands available
	imgui.Separator()
	imgui.Text("Available group commands:")
	imgui.Text("  - Move")
	imgui.Text("  - Attack")
	imgui.Text("  - Stop")
	imgui.Text("  - Hold Position")
}

// renderBuildingInfo renders information about a selected building
func (up *UnitPanel) renderBuildingInfo(building *engine.GameBuilding) {
	// Building name and type
	imgui.Text(fmt.Sprintf("%s (%s)", building.Name, building.BuildingType))
	imgui.Separator()

	// Construction status
	if !building.IsBuilt {
		imgui.Text("Under Construction")

		// Construction progress bar
		imgui.PushStyleColorVec4(imgui.StyleColorPlotHistogram, imgui.Vec4{X: 0.8, Y: 0.6, Z: 0.2, W: 1.0}) // Orange
		imgui.ProgressBar(building.BuildProgress, imgui.Vec2{X: -1, Y: 0}, fmt.Sprintf("Progress: %.1f%%", building.BuildProgress*100))
		imgui.PopStyleColor()
	} else {
		imgui.Text("Completed")
	}

	// Health bar
	healthPercent := float32(building.Health) / float32(building.MaxHealth)
	healthColor := imgui.Vec4{X: 1.0 - healthPercent, Y: healthPercent, Z: 0.0, W: 1.0}

	imgui.PushStyleColorVec4(imgui.StyleColorPlotHistogram, healthColor)
	imgui.ProgressBar(healthPercent, imgui.Vec2{X: -1, Y: 0}, fmt.Sprintf("Health: %d/%d", building.Health, building.MaxHealth))
	imgui.PopStyleColor()

	// Building-specific information
	if building.IsBuilt {
		// Production queue if applicable
		if len(building.ProductionQueue) > 0 || building.CurrentProduction != nil {
			imgui.Separator()
			up.renderBuildingProduction(building)
		}

		// Resource generation if applicable
		if len(building.ResourceGeneration) > 0 {
			imgui.Separator()
			imgui.Text("Resource Generation:")
			for resource, rate := range building.ResourceGeneration {
				imgui.Text(fmt.Sprintf("  %s: %.1f/sec", resource, rate))
			}
		}

		// Upgrade information
		if building.CurrentUpgrade != nil {
			imgui.Separator()
			imgui.Text("Upgrading...")

			upgrade := building.CurrentUpgrade
			imgui.PushStyleColorVec4(imgui.StyleColorPlotHistogram, imgui.Vec4{X: 0.2, Y: 0.8, Z: 0.2, W: 1.0}) // Green
			imgui.ProgressBar(upgrade.Progress, imgui.Vec2{X: -1, Y: 0}, fmt.Sprintf("%s: %.1f%%", upgrade.UpgradeName, upgrade.Progress*100))
			imgui.PopStyleColor()
		}
	}

	// Position information
	imgui.Text(fmt.Sprintf("Position: (%.1f, %.1f, %.1f)", building.Position.X, building.Position.Y, building.Position.Z))
}

// renderBuildingProduction renders building production queue information
func (up *UnitPanel) renderBuildingProduction(building *engine.GameBuilding) {
	imgui.Text("Production:")

	// Current production
	if building.CurrentProduction != nil {
		production := building.CurrentProduction
		imgui.PushStyleColorVec4(imgui.StyleColorPlotHistogram, imgui.Vec4{X: 0.0, Y: 0.6, Z: 1.0, W: 1.0}) // Blue
		imgui.ProgressBar(production.Progress, imgui.Vec2{X: -1, Y: 0}, fmt.Sprintf("Producing %s: %.1f%%", production.ItemName, production.Progress*100))
		imgui.PopStyleColor()
	}

	// Production queue
	if len(building.ProductionQueue) > 0 {
		imgui.Text(fmt.Sprintf("Queue (%d items):", len(building.ProductionQueue)))
		for i, item := range building.ProductionQueue {
			if i >= 3 { // Only show first 3 queued items
				imgui.Text("  ...")
				break
			}
			imgui.Text(fmt.Sprintf("  %d. %s", i+1, item.ItemName))
		}
	}
}

// renderUnitCombatStats renders detailed combat statistics for a unit
func (up *UnitPanel) renderUnitCombatStats(unit *engine.GameUnit) {
	imgui.Text("Combat Stats:")
	imgui.Text(fmt.Sprintf("  Attack: %d", unit.AttackDamage))
	imgui.Text(fmt.Sprintf("  Armor: %d", unit.Armor))
	imgui.Text(fmt.Sprintf("  Range: %.1f", unit.AttackRange))
	imgui.Text(fmt.Sprintf("  Speed: %.1f", unit.Speed))

	// Last attack cooldown
	if !unit.LastAttack.IsZero() {
		timeSinceAttack := time.Since(unit.LastAttack)
		cooldownTime := time.Duration(1.0/unit.AttackSpeed) * time.Second

		if timeSinceAttack < cooldownTime {
			remaining := cooldownTime - timeSinceAttack
			imgui.Text(fmt.Sprintf("  Cooldown: %.1fs", remaining.Seconds()))
		} else {
			imgui.Text("  Ready to attack")
		}
	}
}

// SetShowDetailedStats enables/disables detailed statistics view
func (up *UnitPanel) SetShowDetailedStats(show bool) {
	up.showDetailedStats = show
}

// SetCompactMode enables/disables compact display mode
func (up *UnitPanel) SetCompactMode(compact bool) {
	up.compactMode = compact
}