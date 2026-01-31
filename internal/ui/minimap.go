package ui

import (
	"image"
	"image/color"
	"time"

	"teraglest/internal/engine"

	"github.com/inkyblackness/imgui-go/v4"
)

// Minimap provides a strategic overview of the game world
type Minimap struct {
	world   *engine.World
	width   int
	height  int

	// Minimap data
	terrainTexture uint32 // OpenGL texture ID for terrain
	mapImage      *image.RGBA
	needsUpdate   bool

	// Display settings
	showUnits      bool
	showBuildings  bool
	showResources  bool
	showFogOfWar   bool
	scale          float32

	// Player colors for units/buildings
	playerColors map[int]color.RGBA

	// Update timing
	lastUpdate     time.Time
	updateInterval time.Duration
}

// NewMinimap creates a new minimap component
func NewMinimap(world *engine.World, width, height int) (*Minimap, error) {
	minimap := &Minimap{
		world:          world,
		width:          width,
		height:         height,
		mapImage:       image.NewRGBA(image.Rect(0, 0, width, height)),
		needsUpdate:    true,
		showUnits:      true,
		showBuildings:  true,
		showResources:  true,
		showFogOfWar:   false,
		scale:          1.0,
		playerColors:   make(map[int]color.RGBA),
		updateInterval: 100 * time.Millisecond, // Update 10 times per second
	}

	// Initialize player colors
	minimap.initializePlayerColors()

	// Generate initial terrain texture
	minimap.generateTerrainTexture()

	return minimap, nil
}

// initializePlayerColors sets up colors for different players
func (m *Minimap) initializePlayerColors() {
	m.playerColors[1] = color.RGBA{R: 0, G: 100, B: 255, A: 255}   // Blue
	m.playerColors[2] = color.RGBA{R: 255, G: 0, B: 0, A: 255}     // Red
	m.playerColors[3] = color.RGBA{R: 0, G: 255, B: 0, A: 255}     // Green
	m.playerColors[4] = color.RGBA{R: 255, G: 255, B: 0, A: 255}   // Yellow
	m.playerColors[5] = color.RGBA{R: 255, G: 0, B: 255, A: 255}   // Magenta
	m.playerColors[6] = color.RGBA{R: 0, G: 255, B: 255, A: 255}   // Cyan
	m.playerColors[7] = color.RGBA{R: 255, G: 128, B: 0, A: 255}   // Orange
	m.playerColors[8] = color.RGBA{R: 128, G: 0, B: 128, A: 255}   // Purple
}

// Update updates the minimap data
func (m *Minimap) Update(deltaTime time.Duration) {
	now := time.Now()
	if now.Sub(m.lastUpdate) < m.updateInterval {
		return
	}

	m.lastUpdate = now
	m.needsUpdate = true

	// Update minimap content if needed
	if m.needsUpdate {
		m.updateMinimapImage()
		m.needsUpdate = false
	}
}

// Render renders the minimap
func (m *Minimap) Render() {
	// Position minimap in top-right corner
	displaySize := imgui.Vec2{X: 1024, Y: 768} // Default size, will be updated by UI manager
	minimapSize := imgui.Vec2{X: float32(m.width), Y: float32(m.height)}

	imgui.SetNextWindowPos(imgui.Vec2{
		X: displaySize.X - minimapSize.X - 20,
		Y: 60, // Below resource bar
	})
	imgui.SetNextWindowSize(imgui.Vec2{
		X: minimapSize.X + 20,
		Y: minimapSize.Y + 50,
	})

	if imgui.Begin("Minimap") {
		// Render minimap image
		m.renderMinimapImage()

		// Render controls
		imgui.Separator()
		m.renderMinimapControls()
	}
	imgui.End()
}

// renderMinimapImage renders the actual minimap visualization
func (m *Minimap) renderMinimapImage() {
	// For now, render a simple representation using ImGui drawing commands
	drawList := imgui.WindowDrawList()
	windowPos := imgui.CursorScreenPos()

	// Background (terrain)
	backgroundColor := imgui.PackedColorFromVec4(imgui.Vec4{X: 0.2, Y: 0.4, Z: 0.2, W: 1.0}) // Dark green
	drawList.AddRectFilled(
		windowPos,
		imgui.Vec2{X: windowPos.X + float32(m.width), Y: windowPos.Y + float32(m.height)},
		backgroundColor,
	)

	// World dimensions for scaling
	worldWidth := float32(m.world.Width)
	worldHeight := float32(m.world.Height)
	scaleX := float32(m.width) / worldWidth
	scaleY := float32(m.height) / worldHeight

	// Render resources if enabled
	if m.showResources {
		m.renderMinimapResources(drawList, windowPos, scaleX, scaleY)
	}

	// Render buildings if enabled
	if m.showBuildings {
		m.renderMinimapBuildings(drawList, windowPos, scaleX, scaleY)
	}

	// Render units if enabled
	if m.showUnits {
		m.renderMinimapUnits(drawList, windowPos, scaleX, scaleY)
	}

	// Create invisible button over minimap for click handling
	imgui.SetCursorScreenPos(windowPos)
	imgui.InvisibleButton("minimap_area", imgui.Vec2{X: float32(m.width), Y: float32(m.height)})

	if imgui.IsItemClicked() {
		// Handle minimap click for camera movement
		m.handleMinimapClick()
	}
}

// renderMinimapResources renders resource nodes on the minimap
func (m *Minimap) renderMinimapResources(drawList imgui.DrawList, windowPos imgui.Vec2, scaleX, scaleY float32) {
	resourceColor := imgui.PackedColorFromVec4(imgui.Vec4{X: 0.8, Y: 0.6, Z: 0.0, W: 1.0}) // Gold

	// Get all resource nodes from the world
	for _, resource := range m.world.GetResources() {
		if resource.Amount > 0 {
			// Convert world coordinates to minimap coordinates
			minimapX := windowPos.X + float32(resource.Position.X)*scaleX
			minimapY := windowPos.Y + float32(resource.Position.Z)*scaleY

			// Draw resource as small circle
			drawList.AddCircleFilled(
				imgui.Vec2{X: minimapX, Y: minimapY},
				2.0,
				resourceColor,
			)
		}
	}
}

// renderMinimapBuildings renders buildings on the minimap
func (m *Minimap) renderMinimapBuildings(drawList imgui.DrawList, windowPos imgui.Vec2, scaleX, scaleY float32) {
	// Render buildings for all players
	for playerID := range m.world.GetPlayers() {
		playerColor, exists := m.playerColors[playerID]
		if !exists {
			playerColor = color.RGBA{R: 128, G: 128, B: 128, A: 255} // Default gray
		}

		buildings := m.world.ObjectManager.GetBuildingsForPlayer(playerID)
		for _, building := range buildings {
			if building.IsAlive() {
				// Convert world coordinates to minimap coordinates
				minimapX := windowPos.X + float32(building.Position.X)*scaleX
				minimapY := windowPos.Y + float32(building.Position.Z)*scaleY

				// Buildings are larger squares
				size := float32(3.0)
				if !building.IsBuilt {
					size = 2.0 // Smaller for under construction
				}

				buildingColorPacked := imgui.PackedColorFromVec4(imgui.Vec4{
					X: float32(playerColor.R) / 255.0,
					Y: float32(playerColor.G) / 255.0,
					Z: float32(playerColor.B) / 255.0,
					W: float32(playerColor.A) / 255.0,
				})

				drawList.AddRectFilled(
					imgui.Vec2{X: minimapX - size, Y: minimapY - size},
					imgui.Vec2{X: minimapX + size, Y: minimapY + size},
					buildingColorPacked,
				)
			}
		}
	}
}

// renderMinimapUnits renders units on the minimap
func (m *Minimap) renderMinimapUnits(drawList imgui.DrawList, windowPos imgui.Vec2, scaleX, scaleY float32) {
	// Render units for all players
	for playerID := range m.world.GetPlayers() {
		playerColor, exists := m.playerColors[playerID]
		if !exists {
			playerColor = color.RGBA{R: 128, G: 128, B: 128, A: 255} // Default gray
		}

		units := m.world.ObjectManager.GetUnitsForPlayer(playerID)
		for _, unit := range units {
			if unit.IsAlive() {
				// Convert world coordinates to minimap coordinates
				minimapX := windowPos.X + float32(unit.Position.X)*scaleX
				minimapY := windowPos.Y + float32(unit.Position.Z)*scaleY

				// Units are small dots
				unitColorPacked := imgui.PackedColorFromVec4(imgui.Vec4{
					X: float32(playerColor.R) / 255.0,
					Y: float32(playerColor.G) / 255.0,
					Z: float32(playerColor.B) / 255.0,
					W: float32(playerColor.A) / 255.0,
				})

				drawList.AddCircleFilled(
					imgui.Vec2{X: minimapX, Y: minimapY},
					1.5,
					unitColorPacked,
				)
			}
		}
	}
}

// renderMinimapControls renders minimap control buttons
func (m *Minimap) renderMinimapControls() {
	// Toggle buttons for different display options
	if imgui.Checkbox("Units", &m.showUnits) {
		m.needsUpdate = true
	}
	imgui.SameLine()

	if imgui.Checkbox("Buildings", &m.showBuildings) {
		m.needsUpdate = true
	}
	imgui.SameLine()

	if imgui.Checkbox("Resources", &m.showResources) {
		m.needsUpdate = true
	}

	// Scale slider
	if imgui.SliderFloat("Scale", &m.scale, 0.5, 2.0) {
		// Update minimap scaling
		m.needsUpdate = true
	}
}

// handleMinimapClick handles mouse clicks on the minimap
func (m *Minimap) handleMinimapClick() {
	// Get mouse position relative to minimap
	mousePos := imgui.Vec2{X: 0, Y: 0} // TODO: Get proper mouse position from IO
	windowPos := imgui.CursorScreenPos()

	// Convert minimap coordinates to world coordinates
	relativePosX := mousePos.X - windowPos.X
	relativePosY := mousePos.Y - windowPos.Y

	worldX := (relativePosX / float32(m.width)) * float32(m.world.Width)
	worldZ := (relativePosY / float32(m.height)) * float32(m.world.Height)

	// TODO: Integrate with camera system to move camera to clicked location
	// For now, this is a placeholder for future camera integration
	_ = worldX
	_ = worldZ
}

// generateTerrainTexture generates the base terrain texture
func (m *Minimap) generateTerrainTexture() {
	// Generate a simple terrain representation
	// This is a placeholder - real implementation would use actual terrain data

	for y := 0; y < m.height; y++ {
		for x := 0; x < m.width; x++ {
			// Simple terrain coloring based on position
			terrainColor := color.RGBA{
				R: uint8(50 + (x*100)/m.width),  // Gradient
				G: uint8(100 + (y*50)/m.height), // Green terrain
				B: 50,
				A: 255,
			}

			m.mapImage.Set(x, y, terrainColor)
		}
	}
}

// updateMinimapImage updates the minimap image with current game state
func (m *Minimap) updateMinimapImage() {
	// This would update the texture with current game state
	// For now, we're using direct rendering with ImGui, so no image update needed
}

// OnResize handles window resize events
func (m *Minimap) OnResize(windowWidth, windowHeight int) {
	// Adjust minimap position if needed
	// Currently positioned relative to window, so should adapt automatically
}

// SetScale sets the minimap zoom scale
func (m *Minimap) SetScale(scale float32) {
	m.scale = scale
	m.needsUpdate = true
}

// SetShowUnits enables/disables unit display
func (m *Minimap) SetShowUnits(show bool) {
	m.showUnits = show
	m.needsUpdate = true
}

// SetShowBuildings enables/disables building display
func (m *Minimap) SetShowBuildings(show bool) {
	m.showBuildings = show
	m.needsUpdate = true
}

// SetShowResources enables/disables resource display
func (m *Minimap) SetShowResources(show bool) {
	m.showResources = show
	m.needsUpdate = true
}

// GetWorldCoordinatesFromMinimap converts minimap coordinates to world coordinates
func (m *Minimap) GetWorldCoordinatesFromMinimap(minimapX, minimapY float32) (float32, float32) {
	worldX := (minimapX / float32(m.width)) * float32(m.world.Width)
	worldZ := (minimapY / float32(m.height)) * float32(m.world.Height)
	return worldX, worldZ
}

// Cleanup cleans up minimap resources
func (m *Minimap) Cleanup() {
	// Clean up OpenGL texture if created
	// For now, using direct ImGui rendering, so nothing to clean up
}