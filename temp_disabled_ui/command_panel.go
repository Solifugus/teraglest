package ui

import (
	"fmt"
	"time"

	"teraglest/internal/engine"

	"github.com/inkyblackness/imgui-go/v4"
)

// CommandPanel provides interface for issuing commands to selected units
type CommandPanel struct {
	world     *engine.World
	uiManager *UIManager

	// Command state
	activeCommandType engine.CommandType
	commandParameters map[string]interface{}

	// UI state
	showAdvancedCommands bool
	buttonSize          imgui.Vec2
}

// NewCommandPanel creates a new command panel
func NewCommandPanel(world *engine.World, uiManager *UIManager) *CommandPanel {
	return &CommandPanel{
		world:                world,
		uiManager:            uiManager,
		commandParameters:    make(map[string]interface{}),
		showAdvancedCommands: false,
		buttonSize:           imgui.Vec2{X: 60, Y: 40},
	}
}

// Update updates the command panel state
func (cp *CommandPanel) Update(deltaTime time.Duration, selectedUnits []*engine.GameUnit, selectedBuilding *engine.GameBuilding) {
	// Update based on current selection
	// For now, no active updates needed
}

// Render renders the command panel
func (cp *CommandPanel) Render() {
	// Position at bottom-center of screen
	displaySize := imgui.Vec2{X: 1024, Y: 768} // Default size, will be updated by UI manager
	panelWidth := float32(400)
	panelHeight := float32(120)

	imgui.SetNextWindowPos(imgui.Vec2{
		X: (displaySize.X - panelWidth) / 2,
		Y: displaySize.Y - panelHeight - 20,
	})
	imgui.SetNextWindowSize(imgui.Vec2{X: panelWidth, Y: panelHeight})

	if imgui.Begin("Commands") {
		cp.renderCommandContent()
	}
	imgui.End()
}

// RenderWithSelection renders the command panel with specific selection
func (cp *CommandPanel) RenderWithSelection(selectedUnits []*engine.GameUnit, selectedBuilding *engine.GameBuilding) {
	// Position at bottom-center of screen
	displaySize := imgui.Vec2{X: 1024, Y: 768} // Default size, will be updated by UI manager
	panelWidth := float32(400)
	panelHeight := float32(120)

	imgui.SetNextWindowPos(imgui.Vec2{
		X: (displaySize.X - panelWidth) / 2,
		Y: displaySize.Y - panelHeight - 20,
	})
	imgui.SetNextWindowSize(imgui.Vec2{X: panelWidth, Y: panelHeight})

	if imgui.Begin("Commands##WithSelection") {
		if len(selectedUnits) > 0 {
			cp.renderUnitCommands(selectedUnits)
		} else if selectedBuilding != nil {
			cp.renderBuildingCommands(selectedBuilding)
		} else {
			imgui.Text("Select units or buildings to issue commands")
		}
	}
	imgui.End()
}

// renderCommandContent renders the main command panel content (legacy method)
func (cp *CommandPanel) renderCommandContent() {
	imgui.Text("Command Panel")
	imgui.Text("Select units to issue commands")
}

// renderUnitCommands renders available commands for selected units
func (cp *CommandPanel) renderUnitCommands(selectedUnits []*engine.GameUnit) {
	// Basic movement and combat commands
	cp.renderBasicCommands()

	// Advanced commands if enabled
	if cp.showAdvancedCommands {
		imgui.Separator()
		cp.renderAdvancedCommands(selectedUnits)
	}

	// Toggle for advanced commands
	imgui.Separator()
	imgui.Checkbox("Show Advanced", &cp.showAdvancedCommands)
}

// renderBuildingCommands renders available commands for selected buildings
func (cp *CommandPanel) renderBuildingCommands(building *engine.GameBuilding) {
	if !building.IsBuilt {
		imgui.Text("Building under construction...")
		return
	}

	imgui.Text(fmt.Sprintf("%s Commands:", building.BuildingType))
	imgui.Separator()

	// Production commands for production buildings
	cp.renderProductionCommands(building)

	// Research commands for research buildings
	cp.renderResearchCommands(building)

	// Upgrade commands
	cp.renderUpgradeCommands(building)
}

// renderBasicCommands renders basic unit command buttons
func (cp *CommandPanel) renderBasicCommands() {
	// First row of commands
	if cp.renderCommandButton("Move", "📍", "Move to target location") {
		cp.issueCommand(engine.CommandMove, nil)
	}
	imgui.SameLine()

	if cp.renderCommandButton("Attack", "⚔️", "Attack target") {
		cp.issueCommand(engine.CommandAttack, nil)
	}
	imgui.SameLine()

	if cp.renderCommandButton("Stop", "🛑", "Stop current action") {
		cp.issueCommand(engine.CommandStop, nil)
	}
	imgui.SameLine()

	if cp.renderCommandButton("Hold", "🛡️", "Hold position") {
		cp.issueCommand(engine.CommandHold, nil)
	}

	// Second row of commands
	if cp.renderCommandButton("Gather", "🌲", "Gather resources") {
		cp.issueCommand(engine.CommandGather, nil)
	}
	imgui.SameLine()

	if cp.renderCommandButton("Build", "🏗️", "Construct building") {
		cp.issueCommand(engine.CommandBuild, nil)
	}
	imgui.SameLine()

	if cp.renderCommandButton("Repair", "🔧", "Repair building") {
		cp.issueCommand(engine.CommandRepair, nil)
	}
	imgui.SameLine()

	if cp.renderCommandButton("Patrol", "🔄", "Patrol area") {
		cp.issueCommand(engine.CommandPatrol, nil)
	}
}

// renderAdvancedCommands renders advanced unit command options
func (cp *CommandPanel) renderAdvancedCommands(selectedUnits []*engine.GameUnit) {
	imgui.Text("Advanced Commands:")

	// Formation commands
	if cp.renderCommandButton("Formation", "📐", "Formation commands") {
		// Open formation selection popup
		imgui.OpenPopup("Formation Selection")
	}

	if imgui.BeginPopup("Formation Selection") {
		formations := []struct {
			name string
			icon string
		}{
			{"Line", "—"},
			{"Column", "|"},
			{"Wedge", "▲"},
			{"Circle", "●"},
			{"Square", "■"},
			{"Skirmish", "☰"},
			{"Phalanx", "████"},
		}

		for _, formation := range formations {
			if imgui.Selectable(fmt.Sprintf("%s %s", formation.icon, formation.name)) {
				params := map[string]interface{}{
					"formation": formation.name,
				}
				cp.issueCommand(engine.CommandFormation, params)
				imgui.CloseCurrentPopup()
			}
		}

		imgui.EndPopup()
	}

	imgui.SameLine()

	// Group commands
	if cp.renderCommandButton("Group", "👥", "Group movement") {
		cp.issueCommand(engine.CommandGroupMove, nil)
	}

	imgui.SameLine()

	// Follow command
	if cp.renderCommandButton("Follow", "👤", "Follow target unit") {
		cp.issueCommand(engine.CommandFollow, nil)
	}

	imgui.SameLine()

	// Guard command
	if cp.renderCommandButton("Guard", "🔒", "Guard target") {
		cp.issueCommand(engine.CommandGuard, nil)
	}
}

// renderProductionCommands renders production commands for buildings
func (cp *CommandPanel) renderProductionCommands(building *engine.GameBuilding) {
	// Get available units that can be produced by this building
	availableUnits := cp.getAvailableUnitsForBuilding(building.BuildingType)

	if len(availableUnits) > 0 {
		imgui.Text("Produce Units:")

		for _, unitType := range availableUnits {
			if cp.renderProductionButton(unitType, cp.getUnitIcon(unitType)) {
				cp.issueProductionCommand(building, unitType)
			}

			// Show multiple buttons per row
			if len(availableUnits) > 1 {
				imgui.SameLine()
			}
		}
	}
}

// renderResearchCommands renders research commands for buildings
func (cp *CommandPanel) renderResearchCommands(building *engine.GameBuilding) {
	if !cp.canBuildingResearch(building.BuildingType) {
		return
	}

	// Get available technologies
	techTree := cp.world.GetProductionSystem().GetTechnologyTree()
	availableTechs := techTree.GetAvailableTechnologies(building.PlayerID)

	if len(availableTechs) > 0 {
		imgui.Text("Research:")

		for i, tech := range availableTechs {
			if i >= 3 { // Limit displayed technologies
				break
			}

			if cp.renderResearchButton(tech.DisplayName, "🔬") {
				cp.issueResearchCommand(building, tech.Name)
			}

			if i < 2 && i < len(availableTechs)-1 {
				imgui.SameLine()
			}

			// Show tooltip with tech details
			if imgui.IsItemHovered() {
				tooltipText := fmt.Sprintf("%s\n%s", tech.DisplayName, tech.Description)
				if len(tech.Cost) > 0 {
					tooltipText += "\n\nCost:"
					for resource, amount := range tech.Cost {
						tooltipText += fmt.Sprintf("\n  %s: %d", resource, amount)
					}
				}
				imgui.SetTooltip(tooltipText)
			}
		}
	}
}

// renderUpgradeCommands renders upgrade commands for buildings
func (cp *CommandPanel) renderUpgradeCommands(building *engine.GameBuilding) {
	if building.UpgradeLevel >= building.MaxUpgradeLevel {
		return
	}

	imgui.Text("Upgrades:")

	upgradeTypes := []string{"durability", "production_speed", "resource_efficiency"}
	for i, upgradeType := range upgradeTypes {
		if cp.renderUpgradeButton(upgradeType, "⬆️") {
			cp.issueUpgradeCommand(building, upgradeType)
		}

		if i < len(upgradeTypes)-1 {
			imgui.SameLine()
		}

		// Show tooltip with upgrade details
		if imgui.IsItemHovered() {
			tooltipText := cp.getUpgradeDescription(upgradeType)
			imgui.SetTooltip(tooltipText)
		}
	}
}

// renderCommandButton renders a command button with icon and tooltip
func (cp *CommandPanel) renderCommandButton(label, icon, tooltip string) bool {
	buttonText := fmt.Sprintf("%s\n%s", icon, label)
	pressed := imgui.Button(buttonText)

	if imgui.IsItemHovered() {
		imgui.SetTooltip(tooltip)
	}

	return pressed
}

// renderProductionButton renders a unit production button
func (cp *CommandPanel) renderProductionButton(unitType, icon string) bool {
	buttonText := fmt.Sprintf("%s\n%s", icon, unitType)
	pressed := imgui.Button(buttonText)

	if imgui.IsItemHovered() {
		// Show unit cost tooltip
		cost := cp.getUnitCost(unitType)
		if len(cost) > 0 {
			tooltipText := fmt.Sprintf("Produce %s\nCost:", unitType)
			for resource, amount := range cost {
				tooltipText += fmt.Sprintf("\n  %s: %d", resource, amount)
			}
			imgui.SetTooltip(tooltipText)
		}
	}

	return pressed
}

// renderResearchButton renders a research button
func (cp *CommandPanel) renderResearchButton(techName, icon string) bool {
	buttonText := fmt.Sprintf("%s\n%s", icon, techName)
	return imgui.Button(buttonText)
}

// renderUpgradeButton renders an upgrade button
func (cp *CommandPanel) renderUpgradeButton(upgradeType, icon string) bool {
	buttonText := fmt.Sprintf("%s\n%s", icon, upgradeType)
	return imgui.Button(buttonText)
}

// issueCommand issues a command to selected units
func (cp *CommandPanel) issueCommand(commandType engine.CommandType, params map[string]interface{}) {
	if params == nil {
		params = make(map[string]interface{})
	}

	err := cp.uiManager.IssueCommand(commandType, params)
	if err != nil {
		fmt.Printf("Failed to issue command %s: %v\n", commandType.String(), err)
	}
}

// issueProductionCommand issues a unit production command
func (cp *CommandPanel) issueProductionCommand(building *engine.GameBuilding, unitType string) {
	commandProcessorInterface := cp.world.GetCommandProcessor()
	commandProcessor, ok := commandProcessorInterface.(*engine.CommandProcessor)
	if !ok {
		fmt.Printf("Failed to get command processor for production command\n")
		return
	}
	err := commandProcessor.IssueUnitProductionCommand(building.ID, unitType)
	if err != nil {
		fmt.Printf("Failed to issue production command for %s: %v\n", unitType, err)
	}
}

// issueResearchCommand issues a research command
func (cp *CommandPanel) issueResearchCommand(building *engine.GameBuilding, technologyName string) {
	commandProcessorInterface := cp.world.GetCommandProcessor()
	commandProcessor, ok := commandProcessorInterface.(*engine.CommandProcessor)
	if !ok {
		fmt.Printf("Failed to get command processor for research command\n")
		return
	}
	err := commandProcessor.StartResearchCommand(building.ID, technologyName)
	if err != nil {
		fmt.Printf("Failed to issue research command for %s: %v\n", technologyName, err)
	}
}

// issueUpgradeCommand issues an upgrade command
func (cp *CommandPanel) issueUpgradeCommand(building *engine.GameBuilding, upgradeType string) {
	// Issue to building (buildings don't use the unit command system directly)
	fmt.Printf("Upgrade command for %s: %s\n", building.BuildingType, upgradeType)
}

// Helper methods

// getAvailableUnitsForBuilding returns units that can be produced by a building type
func (cp *CommandPanel) getAvailableUnitsForBuilding(buildingType string) []string {
	unitsByBuilding := map[string][]string{
		"barracks":       {"swordman", "archer"},
		"archery_range":  {"archer", "horseman"},
		"castle":         {"swordman", "horseman", "catapult"},
		"mage_tower":     {"initiate", "battlemage"},
		"summoner_guild": {"daemon", "summoner"},
		"dragon_lair":    {"dragon"},
	}

	if units, exists := unitsByBuilding[buildingType]; exists {
		return units
	}
	return []string{}
}

// getUnitIcon returns an icon for a unit type
func (cp *CommandPanel) getUnitIcon(unitType string) string {
	icons := map[string]string{
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

	if icon, exists := icons[unitType]; exists {
		return icon
	}
	return "🎮" // Default icon
}

// getUnitCost returns the resource cost for a unit type
func (cp *CommandPanel) getUnitCost(unitType string) map[string]int {
	costs := map[string]map[string]int{
		"worker":     {"wood": 50, "gold": 25},
		"swordman":   {"wood": 75, "gold": 50},
		"archer":     {"wood": 60, "gold": 40},
		"horseman":   {"wood": 100, "gold": 75},
		"catapult":   {"wood": 150, "gold": 100, "stone": 50},
		"initiate":   {"energy": 75, "gold": 50},
		"battlemage": {"energy": 150, "gold": 100},
		"daemon":     {"energy": 100, "gold": 75},
		"summoner":   {"energy": 200, "gold": 150},
		"dragon":     {"energy": 300, "gold": 250},
	}

	if cost, exists := costs[unitType]; exists {
		return cost
	}
	return map[string]int{"wood": 50, "gold": 25} // Default cost
}

// canBuildingResearch checks if a building can conduct research
func (cp *CommandPanel) canBuildingResearch(buildingType string) bool {
	researchBuildings := map[string]bool{
		"mage_tower":     true,
		"summoner_guild": true,
		"library":        true,
		"blacksmith":     true,
		"castle":         true,
	}

	return researchBuildings[buildingType]
}

// getUpgradeDescription returns a description for an upgrade type
func (cp *CommandPanel) getUpgradeDescription(upgradeType string) string {
	descriptions := map[string]string{
		"durability":          "Increase building health by 20%",
		"production_speed":    "Increase production speed by 25%",
		"resource_efficiency": "Increase resource generation by 25%",
	}

	if desc, exists := descriptions[upgradeType]; exists {
		return desc
	}
	return "Building upgrade"
}