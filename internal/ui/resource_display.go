package ui

import (
	"fmt"
	"time"

	"teraglest/internal/engine"

	"github.com/inkyblackness/imgui-go/v4"
)

// ResourceDisplay manages the display of player resources
type ResourceDisplay struct {
	world    *engine.World
	playerID int // Which player's resources to display

	// Resource tracking for animations
	lastResources map[string]int
	resourceDelta map[string]int
	deltaTime     map[string]time.Time

	// Display settings
	showBackground bool
	compactMode    bool
}

// NewResourceDisplay creates a new resource display component
func NewResourceDisplay(world *engine.World) (*ResourceDisplay, error) {
	return &ResourceDisplay{
		world:          world,
		playerID:       1, // Default to player 1 (human player)
		lastResources:  make(map[string]int),
		resourceDelta:  make(map[string]int),
		deltaTime:      make(map[string]time.Time),
		showBackground: true,
		compactMode:    false,
	}, nil
}

// Update updates the resource display state
func (rd *ResourceDisplay) Update(deltaTime time.Duration) {
	player := rd.world.GetPlayer(rd.playerID)
	if player == nil {
		return
	}

	currentTime := time.Now()

	// Track resource changes for animations
	for resourceType, currentAmount := range player.Resources {
		if lastAmount, exists := rd.lastResources[resourceType]; exists {
			if currentAmount != lastAmount {
				rd.resourceDelta[resourceType] = currentAmount - lastAmount
				rd.deltaTime[resourceType] = currentTime
			}
		}
		rd.lastResources[resourceType] = currentAmount
	}

	// Clean up old deltas (after 2 seconds)
	for resourceType, deltaTime := range rd.deltaTime {
		if currentTime.Sub(deltaTime) > 2*time.Second {
			delete(rd.resourceDelta, resourceType)
			delete(rd.deltaTime, resourceType)
		}
	}
}

// Render renders the resource display
func (rd *ResourceDisplay) Render() {
	player := rd.world.GetPlayer(rd.playerID)
	if player == nil {
		return
	}

	// Window flags simplified for compatibility

	// Background handling simplified for compatibility

	// Position at top of screen
	displaySize := imgui.Vec2{X: 1024, Y: 768} // Default size, will be updated by UI manager
	windowWidth := float32(400)
	if rd.compactMode {
		windowWidth = 300
	}

	imgui.SetNextWindowPos(imgui.Vec2{
		X: (displaySize.X - windowWidth) / 2,
		Y: 10,
	})
	imgui.SetNextWindowSize(imgui.Vec2{X: windowWidth, Y: 0})

	if imgui.Begin("Resources") {
		rd.renderResourceBar(player)
	}
	imgui.End()
}

// renderResourceBar renders the main resource bar
func (rd *ResourceDisplay) renderResourceBar(player *engine.Player) {
	// Define resource order and colors
	resourceInfo := map[string]struct {
		icon  string
		color imgui.Vec4
	}{
		"wood":   {"🌲", imgui.Vec4{X: 0.4, Y: 0.2, Z: 0.0, W: 1.0}}, // Brown
		"gold":   {"🪙", imgui.Vec4{X: 1.0, Y: 0.8, Z: 0.0, W: 1.0}}, // Gold
		"stone":  {"🪨", imgui.Vec4{X: 0.5, Y: 0.5, Z: 0.5, W: 1.0}}, // Gray
		"energy": {"⚡", imgui.Vec4{X: 0.0, Y: 0.4, Z: 1.0, W: 1.0}}, // Blue
		"food":   {"🌾", imgui.Vec4{X: 0.8, Y: 0.6, Z: 0.2, W: 1.0}}, // Yellow-brown
	}

	// Resource order for consistent display
	resourceOrder := []string{"wood", "gold", "stone", "energy", "food"}

	if rd.compactMode {
		rd.renderCompactMode(player, resourceInfo, resourceOrder)
	} else {
		rd.renderFullMode(player, resourceInfo, resourceOrder)
	}

	// Population display if production system available
	if rd.world.GetProductionSystem() != nil {
		imgui.SameLine()
		imgui.Spacing()
		imgui.SameLine()
		rd.renderPopulationDisplay()
	}
}

// renderFullMode renders resources in full mode with icons and labels
func (rd *ResourceDisplay) renderFullMode(player *engine.Player, resourceInfo map[string]struct {
	icon  string
	color imgui.Vec4
}, resourceOrder []string) {
	for i, resourceType := range resourceOrder {
		amount, exists := player.Resources[resourceType]
		if !exists {
			continue
		}

		if i > 0 {
			imgui.SameLine()
			imgui.Spacing()
			imgui.SameLine()
		}

		info := resourceInfo[resourceType]

		// Resource icon and amount
		// Color styling simplified for compatibility
		imgui.Text(fmt.Sprintf("%s %d", info.icon, amount))
		imgui.PopStyleColor()

		// Show delta if recent change
		if delta, hasDelta := rd.resourceDelta[resourceType]; hasDelta {
			imgui.SameLine()
			// Delta color logic simplified for compatibility

			// Color styling simplified for compatibility
			deltaText := fmt.Sprintf("%+d", delta)
			imgui.Text(deltaText)
			imgui.PopStyleColor()
		}

		// Tooltip with resource type name
		if imgui.IsItemHovered() {
			imgui.SetTooltip(fmt.Sprintf("%s: %d", resourceType, amount))
		}
	}
}

// renderCompactMode renders resources in compact mode
func (rd *ResourceDisplay) renderCompactMode(player *engine.Player, resourceInfo map[string]struct {
	icon  string
	color imgui.Vec4
}, resourceOrder []string) {
	for i, resourceType := range resourceOrder {
		amount, exists := player.Resources[resourceType]
		if !exists {
			continue
		}

		if i > 0 {
			imgui.SameLine()
		}

		info := resourceInfo[resourceType]

		// Just icon and number in compact mode
		// Color styling simplified for compatibility
		imgui.Text(fmt.Sprintf("%s%d", info.icon, amount))
		imgui.PopStyleColor()

		// Tooltip with full information
		if imgui.IsItemHovered() {
			tooltipText := fmt.Sprintf("%s: %d", resourceType, amount)
			if delta, hasDelta := rd.resourceDelta[resourceType]; hasDelta {
				tooltipText += fmt.Sprintf(" (%+d)", delta)
			}
			imgui.SetTooltip(tooltipText)
		}
	}
}

// renderPopulationDisplay renders the population information
func (rd *ResourceDisplay) renderPopulationDisplay() {
	popMgr := rd.world.GetProductionSystem().GetPopulationManager()
	popStats := popMgr.GetPopulationStatus(rd.playerID)

	// Population icon and count color logic simplified for compatibility

	// Color styling simplified for compatibility
	imgui.Text(fmt.Sprintf("👥 %d/%d", popStats.CurrentPopulation, popStats.MaxPopulation))
	imgui.PopStyleColor()

	// Population tooltip
	if imgui.IsItemHovered() {
		tooltipText := fmt.Sprintf("Population: %d/%d\nHousing Buildings: %d",
			popStats.CurrentPopulation, popStats.MaxPopulation, len(popStats.HousingBuildings))

		// Add unit breakdown
		if len(popStats.PopulationUnits) > 0 {
			tooltipText += fmt.Sprintf("\n\nPopulation Units: %d", len(popStats.PopulationUnits))
		}

		imgui.SetTooltip(tooltipText)
	}
}

// SetPlayerID sets which player's resources to display
func (rd *ResourceDisplay) SetPlayerID(playerID int) {
	rd.playerID = playerID
	// Reset tracking when switching players
	rd.lastResources = make(map[string]int)
	rd.resourceDelta = make(map[string]int)
	rd.deltaTime = make(map[string]time.Time)
}

// SetCompactMode enables or disables compact display mode
func (rd *ResourceDisplay) SetCompactMode(compact bool) {
	rd.compactMode = compact
}

// SetShowBackground enables or disables the background
func (rd *ResourceDisplay) SetShowBackground(show bool) {
	rd.showBackground = show
}

// Cleanup cleans up resources used by the display
func (rd *ResourceDisplay) Cleanup() {
	// Nothing to clean up for now
}