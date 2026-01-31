package ui

import (
	"fmt"
	"time"

	"teraglest/internal/engine"

	"github.com/inkyblackness/imgui-go/v4"
)

// ProductionUI manages the building production and research interface
type ProductionUI struct {
	world     *engine.World
	uiManager *UIManager

	// UI state
	showProductionDetails bool
	showResearchDetails   bool
	autoScrollQueue       bool
}

// NewProductionUI creates a new production UI component
func NewProductionUI(world *engine.World, uiManager *UIManager) *ProductionUI {
	return &ProductionUI{
		world:                 world,
		uiManager:             uiManager,
		showProductionDetails: true,
		showResearchDetails:   true,
		autoScrollQueue:       false,
	}
}

// Update updates the production UI state
func (pui *ProductionUI) Update(deltaTime time.Duration, selectedBuilding *engine.GameBuilding) {
	// Update based on selected building production state
}

// Render renders the production UI
func (pui *ProductionUI) Render() {
	// This is the legacy render method - the main rendering happens in RenderWithBuilding
}

// RenderWithBuilding renders the production UI for a specific building
func (pui *ProductionUI) RenderWithBuilding(building *engine.GameBuilding) {
	if building == nil || !building.IsBuilt {
		return
	}

	// Position at bottom-right of screen
	displaySize := imgui.Vec2{X: 1024, Y: 768} // Default size, will be updated by UI manager
	panelWidth := float32(350)
	panelHeight := float32(250)

	imgui.SetNextWindowPos(imgui.Vec2{
		X: displaySize.X - panelWidth - 20,
		Y: displaySize.Y - panelHeight - 20,
	})
	imgui.SetNextWindowSize(imgui.Vec2{X: panelWidth, Y: panelHeight})

	windowTitle := fmt.Sprintf("%s Production", building.BuildingType)
	if imgui.Begin(windowTitle) {
		pui.renderProductionContent(building)
	}
	imgui.End()
}

// renderProductionContent renders the main production interface
func (pui *ProductionUI) renderProductionContent(building *engine.GameBuilding) {
	// Production queue section
	if pui.showProductionDetails {
		pui.renderProductionQueue(building)
	}

	imgui.Separator()

	// Research section
	if pui.showResearchDetails && pui.canBuildingResearch(building.BuildingType) {
		pui.renderResearchInterface(building)
	}

	imgui.Separator()

	// Building upgrade section
	pui.renderUpgradeInterface(building)

	// UI options
	imgui.Separator()
	imgui.Checkbox("Production Details", &pui.showProductionDetails)
	imgui.SameLine()
	imgui.Checkbox("Research Details", &pui.showResearchDetails)
}

// renderProductionQueue renders the production queue interface
func (pui *ProductionUI) renderProductionQueue(building *engine.GameBuilding) {
	if imgui.CollapsingHeader("Production Queue") {
		productionSys := pui.world.GetProductionSystem()
		if productionSys == nil {
			imgui.Text("Production system not available")
			return
		}

		// Get production queue information
		queue, current, err := productionSys.GetProductionQueue(building.ID)
		if err != nil {
			imgui.Text(fmt.Sprintf("Error: %v", err))
			return
		}

		// Current production
		if current != nil {
			pui.renderCurrentProduction(current)
		} else {
			imgui.Text("No current production")
		}

		// Production queue
		if len(queue) > 0 {
			imgui.Text(fmt.Sprintf("Queue (%d items):", len(queue)))

			// Scrollable queue list
			if imgui.BeginChild("ProductionQueue") {
				for i, item := range queue {
					pui.renderQueueItem(i+1, &item)
				}
			}
			imgui.EndChild()

			// Queue management buttons
			if imgui.Button("Clear Queue") {
				pui.clearProductionQueue(building)
			}
			imgui.SameLine()
			if imgui.Button("Cancel Current") {
				pui.cancelCurrentProduction(building)
			}
		} else {
			imgui.Text("Production queue empty")
		}
	}
}

// renderCurrentProduction renders the currently producing item
func (pui *ProductionUI) renderCurrentProduction(current *engine.ProductionItem) {
	imgui.Text(fmt.Sprintf("Producing: %s", current.ItemName))

	// Progress bar
		if current.ItemType == "research" {
		// Blue progress color
	} else if current.ItemType == "upgrade" {
		// Orange progress color
	}

	// Color styling simplified for compatibility

	progressText := fmt.Sprintf("%.1f%%", current.Progress*100)
	if !current.StartTime.IsZero() {
		elapsed := time.Since(current.StartTime)
		remaining := current.Duration - elapsed
		if remaining > 0 {
			progressText += fmt.Sprintf(" (%.1fs)", remaining.Seconds())
		}
	}

	imgui.ProgressBar(current.Progress)
			imgui.SameLine()
			imgui.Text(progressText)
	imgui.PopStyleColor()

	// Show cost information
	if len(current.Cost) > 0 {
		imgui.Text("Cost:")
		imgui.SameLine()
		for resource, amount := range current.Cost {
			imgui.Text(fmt.Sprintf("%s:%d ", resource, amount))
			imgui.SameLine()
		}
		imgui.Text("")
	}
}

// renderQueueItem renders a single item in the production queue
func (pui *ProductionUI) renderQueueItem(position int, item *engine.ProductionItem) {
	// Item icon and name
	icon := pui.getItemIcon(item.ItemType, item.ItemName)
	imgui.Text(fmt.Sprintf("%d. %s %s", position, icon, item.ItemName))

	// Show cost on same line
	if len(item.Cost) > 0 {
		imgui.SameLine()
		imgui.Text("(")
		imgui.SameLine()
		first := true
		for resource, amount := range item.Cost {
			if !first {
				imgui.SameLine()
				imgui.Text(",")
				imgui.SameLine()
			}
			imgui.Text(fmt.Sprintf("%s:%d", resource, amount))
			imgui.SameLine()
			first = false
		}
		imgui.Text(")")
	}

	// Context menu for queue item management
	if imgui.BeginPopupContextItem() {
		if imgui.MenuItem("Remove from queue") {
			// TODO: Implement queue item removal
		}
		if imgui.MenuItem("Move to top") {
			// TODO: Implement queue reordering
		}
		imgui.EndPopup()
	}
}

// renderResearchInterface renders the research interface
func (pui *ProductionUI) renderResearchInterface(building *engine.GameBuilding) {
	if imgui.CollapsingHeader("Research") {
		techTree := pui.world.GetProductionSystem().GetTechnologyTree()

		// Current research
		current, queue := techTree.GetResearchProgress(building.PlayerID)
		if current != nil {
			imgui.Text("Researching:")
			pui.renderCurrentResearch(current)
		} else {
			imgui.Text("No current research")
		}

		// Research queue
		if len(queue) > 0 {
			imgui.Text(fmt.Sprintf("Research Queue (%d):", len(queue)))
			for i, research := range queue {
				imgui.Text(fmt.Sprintf("  %d. %s", i+1, research.TechName))
			}
		}

		imgui.Separator()

		// Available technologies
		availableTechs := techTree.GetAvailableTechnologies(building.PlayerID)
		if len(availableTechs) > 0 {
			imgui.Text("Available Research:")

			if imgui.BeginChild("AvailableResearch") {
				for _, tech := range availableTechs {
					pui.renderAvailableTechnology(building, tech)
				}
			}
			imgui.EndChild()
		} else {
			imgui.Text("No research available")
		}
	}
}

// renderCurrentResearch renders current research progress
func (pui *ProductionUI) renderCurrentResearch(research *engine.ResearchItem) {
	imgui.Text(fmt.Sprintf("🔬 %s", research.TechName))

	// Progress bar
		// Color styling simplified for compatibility

	progressText := fmt.Sprintf("%.1f%%", research.Progress*100)
	if !research.StartTime.IsZero() {
		elapsed := time.Since(research.StartTime)
		remaining := research.Duration - elapsed
		if remaining > 0 {
			progressText += fmt.Sprintf(" (%.1fs)", remaining.Seconds())
		}
	}

	imgui.ProgressBar(research.Progress)
			imgui.SameLine()
			imgui.Text(progressText)
	imgui.PopStyleColor()
}

// renderAvailableTechnology renders an available technology for research
func (pui *ProductionUI) renderAvailableTechnology(building *engine.GameBuilding, tech *engine.TechnologyDefinition) {
	// Technology button
	if imgui.Button(fmt.Sprintf("🔬 %s", tech.DisplayName)) {
		pui.startResearch(building, tech.Name)
	}

	// Show tooltip with details
	if imgui.IsItemHovered() {
		tooltipText := fmt.Sprintf("%s\n%s", tech.DisplayName, tech.Description)

		if len(tech.Cost) > 0 {
			tooltipText += "\n\nCost:"
			for resource, amount := range tech.Cost {
				tooltipText += fmt.Sprintf("\n  %s: %d", resource, amount)
			}
		}

		tooltipText += fmt.Sprintf("\nDuration: %.0f seconds", tech.Duration.Seconds())

		if len(tech.Effects) > 0 {
			tooltipText += "\n\nEffects:"
			for _, effect := range tech.Effects {
				tooltipText += fmt.Sprintf("\n  %s", effect.Description)
			}
		}

		imgui.SetTooltip(tooltipText)
	}
}

// renderUpgradeInterface renders the building upgrade interface
func (pui *ProductionUI) renderUpgradeInterface(building *engine.GameBuilding) {
	if building.CurrentUpgrade != nil {
		imgui.Text("Upgrading:")
		pui.renderCurrentUpgrade(building.CurrentUpgrade)
	} else if building.UpgradeLevel < building.MaxUpgradeLevel {
		if imgui.CollapsingHeader("Upgrades") {
			pui.renderAvailableUpgrades(building)
		}
	} else {
		imgui.Text("Maximum upgrade level reached")
	}
}

// renderCurrentUpgrade renders current upgrade progress
func (pui *ProductionUI) renderCurrentUpgrade(upgrade *engine.UpgradeItem) {
	imgui.Text(fmt.Sprintf("⬆️ %s", upgrade.UpgradeName))

	// Progress bar
		// Color styling simplified for compatibility

	progressText := fmt.Sprintf("%.1f%%", upgrade.Progress*100)
	if !upgrade.StartTime.IsZero() {
		elapsed := time.Since(upgrade.StartTime)
		remaining := upgrade.Duration - elapsed
		if remaining > 0 {
			progressText += fmt.Sprintf(" (%.1fs)", remaining.Seconds())
		}
	}

	imgui.ProgressBar(upgrade.Progress)
			imgui.SameLine()
			imgui.Text(progressText)
	imgui.PopStyleColor()
}

// renderAvailableUpgrades renders available upgrade options
func (pui *ProductionUI) renderAvailableUpgrades(building *engine.GameBuilding) {
	upgradeTypes := []struct {
		name        string
		icon        string
		description string
		cost        map[string]int
	}{
		{"durability", "🛡️", "Increase building health by 20%", map[string]int{"stone": 100, "gold": 50}},
		{"production_speed", "⚡", "Increase production speed by 25%", map[string]int{"wood": 150, "gold": 75}},
		{"resource_efficiency", "💰", "Increase resource generation by 25%", map[string]int{"wood": 100, "stone": 50}},
	}

	for _, upgrade := range upgradeTypes {
		if imgui.Button(fmt.Sprintf("%s %s", upgrade.icon, upgrade.name)) {
			pui.startUpgrade(building, upgrade.name, upgrade.cost)
		}

		if imgui.IsItemHovered() {
			tooltipText := upgrade.description + "\n\nCost:"
			for resource, amount := range upgrade.cost {
				tooltipText += fmt.Sprintf("\n  %s: %d", resource, amount)
			}
			imgui.SetTooltip(tooltipText)
		}
	}
}

// Helper methods

// getItemIcon returns an appropriate icon for an item type/name
func (pui *ProductionUI) getItemIcon(itemType, itemName string) string {
	if itemType == "research" {
		return "🔬"
	} else if itemType == "upgrade" {
		return "⬆️"
	}

	// Unit icons
	unitIcons := map[string]string{
		"worker":     "👷",
		"swordman":   "⚔️",
		"archer":     "🏹",
		"horseman":   "🐎",
		"catapult":   "🎯",
		"initiate":   "🧙",
		"battlemage": "🔮",
		"daemon":     "👹",
		"summoner":   "🌟",
		"dragon":     "🐉",
	}

	if icon, exists := unitIcons[itemName]; exists {
		return icon
	}

	return "🎮" // Default icon
}

// canBuildingResearch checks if a building can conduct research
func (pui *ProductionUI) canBuildingResearch(buildingType string) bool {
	researchBuildings := map[string]bool{
		"mage_tower":     true,
		"summoner_guild": true,
		"library":        true,
		"blacksmith":     true,
		"castle":         true,
	}

	return researchBuildings[buildingType]
}

// startResearch initiates research for a technology
func (pui *ProductionUI) startResearch(building *engine.GameBuilding, technologyName string) {
	commandProcessorInterface := pui.world.GetCommandProcessor()
	commandProcessor, ok := commandProcessorInterface.(*engine.CommandProcessor)
	if !ok {
		fmt.Printf("Failed to get command processor for research command\n")
		return
	}
	err := commandProcessor.StartResearchCommand(building.ID, technologyName)
	if err != nil {
		fmt.Printf("Failed to start research %s: %v\n", technologyName, err)
	}
}

// startUpgrade initiates a building upgrade
func (pui *ProductionUI) startUpgrade(building *engine.GameBuilding, upgradeType string, cost map[string]int) {
	// Validate resources before starting upgrade
	player := pui.world.GetPlayer(building.PlayerID)
	if player != nil {
		for resource, amount := range cost {
			if player.Resources[resource] < amount {
				fmt.Printf("Insufficient %s for upgrade (need %d, have %d)\n",
					resource, amount, player.Resources[resource])
				return
			}
		}
	}

	// TODO: Implement upgrade command through command system
	fmt.Printf("Starting upgrade %s for building %d\n", upgradeType, building.ID)
}

// clearProductionQueue clears the entire production queue
func (pui *ProductionUI) clearProductionQueue(building *engine.GameBuilding) {
	productionSys := pui.world.GetProductionSystem()
	if productionSys != nil {
		// TODO: Implement queue clearing
		fmt.Printf("Clearing production queue for building %d\n", building.ID)
	}
}

// cancelCurrentProduction cancels the current production
func (pui *ProductionUI) cancelCurrentProduction(building *engine.GameBuilding) {
	productionSys := pui.world.GetProductionSystem()
	if productionSys != nil {
		err := productionSys.CancelProduction(building.ID)
		if err != nil {
			fmt.Printf("Failed to cancel production: %v\n", err)
		}
	}
}