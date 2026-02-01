package renderer

import (
	"fmt"
	"log"
	"strings"
	"time"

	"teraglest/internal/data"
	"teraglest/internal/engine"
	"teraglest/internal/graphics"
	"teraglest/pkg/formats"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

// GPUTexture represents a texture uploaded to the GPU
type GPUTexture struct {
	ID     uint32
	Width  int32
	Height int32
	Format uint32
}

// GPUModel represents a 3D model uploaded to the GPU
type GPUModel struct {
	VAO         uint32    // Vertex Array Object
	VBO         uint32    // Vertex Buffer Object
	EBO         uint32    // Element Buffer Object
	IndexCount  int32     // Number of indices to draw
	TextureID   uint32    // Texture ID (if any)
	BoundingBox [6]float32 // Min/max X, Y, Z for culling
}

// Renderer orchestrates all rendering operations
type Renderer struct {
	// Core components
	context     *RenderContext
	assetMgr    *data.AssetManager
	shaderMgr   *ShaderManager
	camera      *Camera
	modelMgr    *graphics.ModelManager
	lightMgr    *graphics.LightManager
	materialMgr *graphics.MaterialManager

	// GPU resource caches
	modelCache   map[string]*GPUModel   // Path -> GPU model
	textureCache map[string]*GPUTexture // Path -> GPU texture

	// Placeholder rendering
	cubeVAO     uint32 // VAO for rendering unit placeholders
	basicShader uint32 // Basic shader for placeholder rendering

	// Rendering statistics
	frameCount    uint64
	lastFrameTime time.Time
	fps           float32

	// Debug settings
	wireframe bool
	showStats bool
}

// NewRenderer creates a new renderer instance
func NewRenderer(assetMgr *data.AssetManager, title string, width, height int) (*Renderer, error) {
	// Create OpenGL context
	context, err := NewRenderContext(title, width, height, false)
	if err != nil {
		return nil, fmt.Errorf("failed to create render context: %w", err)
	}

	// Initialize shader manager
	shaderMgr := NewShaderManager()

	// Create camera for 3D rendering
	camera := NewRTSCamera(width, height, 64.0) // Default map size of 64

	// Create model manager for 3D model handling
	modelMgr := graphics.NewModelManager()

	// Create lighting manager with support for up to 8 lights
	lightMgr := graphics.NewLightManager(8)

	// Create material manager for advanced material support
	materialMgr := graphics.NewMaterialManager()

	renderer := &Renderer{
		context:       context,
		assetMgr:      assetMgr,
		shaderMgr:     shaderMgr,
		camera:        camera,
		modelMgr:      modelMgr,
		lightMgr:      lightMgr,
		materialMgr:   materialMgr,
		modelCache:    make(map[string]*GPUModel),
		textureCache:  make(map[string]*GPUTexture),
		lastFrameTime: time.Now(),
		wireframe:     false,
		showStats:     true,
	}

	// Initialize default lighting
	err = renderer.setupDefaultLighting()
	if err != nil {
		log.Printf("Warning: Failed to setup default lighting: %v", err)
	}

	// Load advanced shaders
	err = renderer.loadAdvancedShaders()
	if err != nil {
		log.Printf("Warning: Failed to load advanced shaders: %v", err)
	}

	// Set up basic input callbacks (can be enhanced later with game input handler)
	renderer.setupInputCallbacks()

	log.Printf("Renderer initialized: %dx%d", width, height)
	return renderer, nil
}

// setupInputCallbacks configures keyboard and mouse input handling
func (r *Renderer) setupInputCallbacks() {
	window := r.context.GetWindow()

	// Keyboard callback
	window.SetKeyCallback(func(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
		if action == glfw.Press {
			switch key {
			case glfw.KeyEscape:
				w.SetShouldClose(true)
			case glfw.KeyF1:
				r.wireframe = !r.wireframe
				if r.wireframe {
					r.context.EnableWireframe()
					log.Println("Wireframe mode enabled")
				} else {
					r.context.DisableWireframe()
					log.Println("Wireframe mode disabled")
				}
			case glfw.KeyF2:
				r.showStats = !r.showStats
				log.Printf("Stats display: %v", r.showStats)
			}
		}
	})
}

// SetupGameInputCallbacks configures input callbacks for game integration
func (r *Renderer) SetupGameInputCallbacks(inputHandler interface{}) {
	window := r.context.GetWindow()

	// Type assertion to check for required methods
	// This allows us to avoid importing ui package here
	type InputHandler interface {
		HandleMouseButton(window *glfw.Window, button glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey)
		HandleMouseMove(window *glfw.Window, xpos, ypos float64)
		HandleKeyboard(window *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey)
	}

	if handler, ok := inputHandler.(InputHandler); ok {
		// Setup mouse button callback
		window.SetMouseButtonCallback(func(w *glfw.Window, button glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
			handler.HandleMouseButton(w, button, action, mods)
		})

		// Setup mouse movement callback
		window.SetCursorPosCallback(func(w *glfw.Window, xpos, ypos float64) {
			handler.HandleMouseMove(w, xpos, ypos)
		})

		// Setup keyboard callback (enhanced version)
		window.SetKeyCallback(func(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
			// Handle basic renderer controls first
			if action == glfw.Press {
				switch key {
				case glfw.KeyEscape:
					w.SetShouldClose(true)
					return
				case glfw.KeyF1:
					r.wireframe = !r.wireframe
					if r.wireframe {
						r.context.EnableWireframe()
						log.Println("Wireframe mode enabled")
					} else {
						r.context.DisableWireframe()
						log.Println("Wireframe mode disabled")
					}
					return
				case glfw.KeyF2:
					r.showStats = !r.showStats
					log.Printf("Stats display: %v", r.showStats)
					return
				}
			}

			// Forward to game input handler
			handler.HandleKeyboard(w, key, scancode, action, mods)
		})

		log.Println("Game input callbacks configured")
	} else {
		log.Println("Warning: Invalid input handler provided to SetupGameInputCallbacks")
	}
}

// ShouldClose returns true if the window should close
func (r *Renderer) ShouldClose() bool {
	return r.context.ShouldClose()
}

// GetContext returns the render context
func (r *Renderer) GetContext() *RenderContext {
	return r.context
}

// GetDisplaySize returns the current display dimensions
func (r *Renderer) GetDisplaySize() (int, int) {
	return r.context.GetWidth(), r.context.GetHeight()
}

// updateStats updates rendering statistics
func (r *Renderer) updateStats() {
	now := time.Now()
	deltaTime := now.Sub(r.lastFrameTime)

	if deltaTime > 0 {
		r.fps = float32(time.Second) / float32(deltaTime)
	}

	r.frameCount++
	r.lastFrameTime = now

	// Log stats every 60 frames
	if r.showStats && r.frameCount%60 == 0 {
		log.Printf("Frame %d: FPS=%.1f, Models cached=%d, Textures cached=%d",
			r.frameCount, r.fps, len(r.modelCache), len(r.textureCache))
	}
}

// RenderFrame renders a complete frame
func (r *Renderer) RenderFrame() error {
	// Update rendering statistics
	r.updateStats()

	// Clear the screen
	r.context.Clear()

	// Setup 3D rendering for basic frame
	err := r.setup3DRendering()
	if err != nil {
		return fmt.Errorf("failed to setup 3D rendering: %w", err)
	}

	// TODO: Render test objects here if needed

	// Swap buffers and poll events
	r.context.SwapBuffers()
	r.context.PollEvents()

	return nil
}

// RenderWorld renders the entire game world (main entry point for game rendering)
func (r *Renderer) RenderWorld(world *engine.World) error {
	if world == nil {
		return fmt.Errorf("world is nil")
	}

	// Update rendering statistics
	r.updateStats()

	// Clear the screen
	r.context.Clear()

	// Setup 3D rendering pipeline
	err := r.setup3DRendering()
	if err != nil {
		return fmt.Errorf("failed to setup 3D rendering: %w", err)
	}

	// Render world objects
	err = r.renderWorldObjects(world)
	if err != nil {
		return fmt.Errorf("failed to render world objects: %w", err)
	}

	// For now, log that we're rendering a world
	if r.frameCount%120 == 0 { // Log every 2 seconds at 60 FPS
		allUnits := 0
		for _, player := range world.GetAllPlayers() {
			allUnits += len(world.ObjectManager.GetUnitsForPlayer(player.ID))
		}

		log.Printf("Rendering world: %dx%d, %d players, %d units, %d resources",
			world.Width, world.Height,
			len(world.GetAllPlayers()),
			allUnits,
			len(world.GetAllResourceNodes()))
	}

	// Swap buffers and poll events
	r.context.SwapBuffers()
	r.context.PollEvents()

	return nil
}

// setup3DRendering configures the rendering pipeline for 3D rendering
func (r *Renderer) setup3DRendering() error {
	// Enable depth testing for proper 3D rendering
	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LESS)

	// Enable back-face culling for performance
	gl.Enable(gl.CULL_FACE)
	gl.CullFace(gl.BACK)

	// Use the advanced model shader for 3D rendering
	shaderName := "advanced_model"
	err := r.shaderMgr.UseShader(shaderName)
	if err != nil {
		// Fallback to basic model shader if advanced shader not available
		shaderName = "model"
		err = r.shaderMgr.UseShader(shaderName)
		if err != nil {
			return fmt.Errorf("failed to use any model shader: %w", err)
		}
	}

	// Set view and projection matrices
	viewMatrix := r.camera.GetViewMatrix()
	projMatrix := r.camera.GetProjectionMatrix()

	err = r.shaderMgr.SetUniformMat4(shaderName, "uView", viewMatrix)
	if err != nil {
		return fmt.Errorf("failed to set view matrix: %w", err)
	}

	err = r.shaderMgr.SetUniformMat4(shaderName, "uProjection", projMatrix)
	if err != nil {
		return fmt.Errorf("failed to set projection matrix: %w", err)
	}

	// Set camera position for lighting calculations
	err = r.shaderMgr.SetUniformVec3(shaderName, "uViewPosition", r.camera.Position)
	if err != nil {
		return fmt.Errorf("failed to set view position: %w", err)
	}

	// Update lighting uniforms if using advanced shader
	if shaderName == "advanced_model" {
		err = r.lightMgr.UpdateShaderUniforms(r.shaderMgr, shaderName)
		if err != nil {
			return fmt.Errorf("failed to update lighting uniforms: %w", err)
		}
	}

	return nil
}

// renderWorldObjects renders all objects in the game world
func (r *Renderer) renderWorldObjects(world *engine.World) error {
	// 1. Render terrain (simplified grid for now)
	err := r.renderTerrain(world)
	if err != nil {
		return fmt.Errorf("failed to render terrain: %w", err)
	}

	// 2. Render all units from the game world
	err = r.renderUnits(world)
	if err != nil {
		return fmt.Errorf("failed to render units: %w", err)
	}

	// 3. Render all buildings from the game world
	err = r.renderBuildings(world)
	if err != nil {
		return fmt.Errorf("failed to render buildings: %w", err)
	}

	// 4. Render resource nodes
	err = r.renderResourceNodes(world)
	if err != nil {
		return fmt.Errorf("failed to render resource nodes: %w", err)
	}

	// 5. Render any additional test models from model manager
	err = r.modelMgr.RenderAllModels("model", r.shaderMgr)
	if err != nil {
		return fmt.Errorf("failed to render test models: %w", err)
	}

	return nil
}

// initializeCubeGeometry creates VAO/VBO for rendering simple cube placeholders
func (r *Renderer) initializeCubeGeometry() {
	// Simple cube vertices (position only)
	vertices := []float32{
		// Front face
		-0.5, -0.5,  0.5,   0.5, -0.5,  0.5,   0.5,  0.5,  0.5,  -0.5,  0.5,  0.5,
		// Back face
		-0.5, -0.5, -0.5,  -0.5,  0.5, -0.5,   0.5,  0.5, -0.5,   0.5, -0.5, -0.5,
	}

	// Cube indices (2 triangles per face, 6 faces)
	indices := []uint32{
		// Front face
		0, 1, 2,   2, 3, 0,
		// Back face
		4, 5, 6,   6, 7, 4,
		// Left face
		4, 0, 3,   3, 5, 4,
		// Right face
		1, 7, 6,   6, 2, 1,
		// Top face
		3, 2, 6,   6, 5, 3,
		// Bottom face
		4, 7, 1,   1, 0, 4,
	}

	// Generate and bind VAO
	gl.GenVertexArrays(1, &r.cubeVAO)
	gl.BindVertexArray(r.cubeVAO)

	// Generate and bind VBO
	var vbo uint32
	gl.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	// Generate and bind EBO
	var ebo uint32
	gl.GenBuffers(1, &ebo)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, ebo)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(indices)*4, gl.Ptr(indices), gl.STATIC_DRAW)

	// Configure vertex attributes (position)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 3*4, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(0)

	// Unbind VAO
	gl.BindVertexArray(0)

	log.Printf("✅ Cube geometry initialized for unit placeholders")
}

// initializeBasicShader creates a simple shader for rendering colored placeholders
func (r *Renderer) initializeBasicShader() error {
	// Simple vertex shader source
	vertexShaderSource := `#version 330 core
layout (location = 0) in vec3 aPos;

uniform mat4 model;
uniform mat4 view;
uniform mat4 projection;

void main() {
    gl_Position = projection * view * model * vec4(aPos, 1.0);
}` + "\x00"

	// Simple fragment shader source
	fragmentShaderSource := `#version 330 core
uniform vec3 color;
out vec4 FragColor;

void main() {
    FragColor = vec4(color, 1.0);
}` + "\x00"

	// Compile vertex shader
	vertexShader, err := compileShader(vertexShaderSource, gl.VERTEX_SHADER)
	if err != nil {
		return fmt.Errorf("vertex shader compilation failed: %v", err)
	}
	defer gl.DeleteShader(vertexShader)

	// Compile fragment shader
	fragmentShader, err := compileShader(fragmentShaderSource, gl.FRAGMENT_SHADER)
	if err != nil {
		return fmt.Errorf("fragment shader compilation failed: %v", err)
	}
	defer gl.DeleteShader(fragmentShader)

	// Create shader program
	r.basicShader = gl.CreateProgram()
	gl.AttachShader(r.basicShader, vertexShader)
	gl.AttachShader(r.basicShader, fragmentShader)
	gl.LinkProgram(r.basicShader)

	// Check linking status
	var status int32
	gl.GetProgramiv(r.basicShader, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(r.basicShader, gl.INFO_LOG_LENGTH, &logLength)
		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetProgramInfoLog(r.basicShader, logLength, nil, gl.Str(log))
		return fmt.Errorf("program linking failed: %v", log)
	}

	log.Printf("✅ Basic shader initialized for unit placeholders")
	return nil
}

// compileShader compiles a shader from source
func compileShader(source string, shaderType uint32) (uint32, error) {
	shader := gl.CreateShader(shaderType)
	csource, free := gl.Strs(source)
	gl.ShaderSource(shader, 1, csource, nil)
	free()
	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)
		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))
		return 0, fmt.Errorf("shader compilation failed: %v", log)
	}

	return shader, nil
}

// GetCamera returns the renderer's camera for external manipulation
func (r *Renderer) GetCamera() *Camera {
	return r.camera
}

// SetCamera updates the renderer's camera
func (r *Renderer) SetCamera(camera *Camera) {
	r.camera = camera
}

// GetModelManager returns the renderer's model manager
func (r *Renderer) GetModelManager() *graphics.ModelManager {
	return r.modelMgr
}

// LoadG3DModel loads a G3D model through the model manager
func (r *Renderer) LoadG3DModel(filePath string) (*graphics.Model, error) {
	return r.modelMgr.LoadG3DModel(filePath)
}

// CreateTestScene creates a test scene with basic models
func (r *Renderer) CreateTestScene() error {
	return r.modelMgr.CreateTestScene()
}

// ResizeViewport handles window resizing
func (r *Renderer) ResizeViewport(width, height int) {
	gl.Viewport(0, 0, int32(width), int32(height))
	r.camera.SetAspectRatio(width, height)
}

// RenderModel renders a 3D model with the given transformation
func (r *Renderer) RenderModel(model *graphics.Model) error {
	if model == nil {
		return fmt.Errorf("model is nil")
	}

	// Ensure we're using the model shader
	err := r.shaderMgr.UseShader("advanced_model")
	if err != nil {
		return fmt.Errorf("failed to use model shader: %w", err)
	}

	// Render the model (it will set its own model matrix and material properties)
	err = model.Render("advanced_model", r.shaderMgr)
	if err != nil {
		return fmt.Errorf("failed to render model: %w", err)
	}

	return nil
}

// RenderModelAt renders a 3D model at a specific position
func (r *Renderer) RenderModelAt(model *graphics.Model, x, y, z float32) error {
	if model == nil {
		return fmt.Errorf("model is nil")
	}

	// Save original transform
	originalTransform := model.GetModelMatrix()

	// Set position
	model.SetPosition(x, y, z)

	// Render model
	err := r.RenderModel(model)

	// Restore original transform
	model.Transform = originalTransform

	return err
}

// setupDefaultLighting initializes the default lighting configuration
func (r *Renderer) setupDefaultLighting() error {
	// Create default lighting setup
	err := r.lightMgr.CreateDefaultLighting()
	if err != nil {
		return fmt.Errorf("failed to create default lighting: %w", err)
	}

	log.Println("Default lighting setup completed")
	return nil
}

// loadAdvancedShaders loads the advanced shader programs
func (r *Renderer) loadAdvancedShaders() error {
	// Load advanced model shader
	err := r.shaderMgr.LoadShader(
		"advanced_model",
		"internal/graphics/shaders/advanced_model.vert",
		"internal/graphics/shaders/advanced_model.frag",
	)
	if err != nil {
		return fmt.Errorf("failed to load advanced model shader: %w", err)
	}

	// Load normal mapped material shader
	err = r.shaderMgr.LoadShader(
		"normal_mapped_material",
		"internal/graphics/shaders/normal_mapped_material.vert",
		"internal/graphics/shaders/normal_mapped_material.frag",
	)
	if err != nil {
		log.Printf("Warning: Failed to load normal mapped shader: %v", err)
	}

	// Set up material shader mappings
	r.materialMgr.SetShaderMapping(graphics.BasicMaterial, "advanced_model")
	r.materialMgr.SetShaderMapping(graphics.TexturedMaterial, "advanced_model")
	r.materialMgr.SetShaderMapping(graphics.NormalMappedMaterial, "normal_mapped_material")
	r.materialMgr.SetShaderMapping(graphics.EmissiveMaterial, "advanced_model")

	log.Println("Advanced shaders and materials loaded successfully")
	return nil
}

// GetLightManager returns the renderer's light manager
func (r *Renderer) GetLightManager() *graphics.LightManager {
	return r.lightMgr
}

// GetMaterialManager returns the renderer's material manager
func (r *Renderer) GetMaterialManager() *graphics.MaterialManager {
	return r.materialMgr
}

// LoadTexture loads a texture from the AssetManager and uploads it to GPU
func (r *Renderer) LoadTexture(texturePath string) (*GPUTexture, error) {
	// Check cache first
	if texture, exists := r.textureCache[texturePath]; exists {
		return texture, nil
	}

	// Load image from AssetManager
	_, err := r.assetMgr.LoadTexture(texturePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load texture from AssetManager: %w", err)
	}

	// TODO: Convert image.Image to OpenGL texture
	// For now, create a placeholder GPU texture
	var textureID uint32
	gl.GenTextures(1, &textureID)

	gpuTexture := &GPUTexture{
		ID:     textureID,
		Width:  100, // Placeholder
		Height: 100, // Placeholder
		Format: gl.RGBA,
	}

	// Cache the texture
	r.textureCache[texturePath] = gpuTexture

	log.Printf("Loaded texture: %s (ID=%d)", texturePath, textureID)
	return gpuTexture, nil
}

// LoadModel loads a G3D model from the AssetManager and uploads it to GPU
func (r *Renderer) LoadModel(modelPath string) (*GPUModel, error) {
	// Check cache first
	if model, exists := r.modelCache[modelPath]; exists {
		return model, nil
	}

	// Load G3D model from AssetManager
	g3dModel, err := r.assetMgr.LoadG3DModel(modelPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load G3D model from AssetManager: %w", err)
	}

	// TODO: Convert G3D model data to OpenGL buffers
	// For now, create a placeholder GPU model
	var vao, vbo, ebo uint32
	gl.GenVertexArrays(1, &vao)
	gl.GenBuffers(1, &vbo)
	gl.GenBuffers(1, &ebo)

	gpuModel := &GPUModel{
		VAO:        vao,
		VBO:        vbo,
		EBO:        ebo,
		IndexCount: 36, // Placeholder for a cube
		TextureID:  0,  // No texture initially
	}

	// Cache the model
	r.modelCache[modelPath] = gpuModel

	log.Printf("Loaded model: %s (VAO=%d, vertices=%d, triangles=%d)",
		modelPath, vao, g3dModel.GetTotalVertexCount(), g3dModel.GetTotalTriangleCount())
	return gpuModel, nil
}

// GetFPS returns the current frames per second
func (r *Renderer) GetFPS() float32 {
	return r.fps
}

// GetFrameCount returns the total number of frames rendered
func (r *Renderer) GetFrameCount() uint64 {
	return r.frameCount
}

// renderTerrain renders the game world terrain
func (r *Renderer) renderTerrain(world *engine.World) error {
	// For now, render a simple grid to represent the game world bounds
	// TODO: Implement proper heightmap-based terrain rendering

	// Skip terrain rendering if world dimensions are too large to avoid performance issues
	if world.Width > 100 || world.Height > 100 {
		return nil
	}

	// Simple grid visualization using lines (placeholder implementation)
	// In a full implementation, this would render textured quads or heightmap meshes

	return nil // Terrain rendering placeholder
}

// renderUnits renders all units from the game world
func (r *Renderer) renderUnits(world *engine.World) error {
	allPlayers := world.GetAllPlayers()

	for _, player := range allPlayers {
		units := world.ObjectManager.GetUnitsForPlayer(player.ID)

		for _, unit := range units {
			// Skip dead units
			if unit.Health <= 0 {
				continue
			}

			err := r.renderUnitWithFaction(unit, player.FactionName)
			if err != nil {
				// Log error but continue rendering other units
				log.Printf("Warning: Failed to render unit %d: %v", unit.ID, err)
				continue
			}
		}
	}

	return nil
}

// renderUnit renders a single game unit (legacy method, use renderUnitWithFaction)
func (r *Renderer) renderUnit(unit *engine.GameUnit) error {
	return r.renderUnitWithFaction(unit, "magic") // fallback to magic for backward compatibility
}

// renderUnitWithFaction renders a single game unit using the correct faction
func (r *Renderer) renderUnitWithFaction(unit *engine.GameUnit, faction string) error {
	// Get unit position from game state
	pos := unit.GetPosition()

	// Load G3D model using the CORRECT faction instead of hardcoding "magic"
	// Try multiple naming patterns for better compatibility
	var g3dModel *formats.G3DModel
	var err error

	// Pattern 1: Try with _standing suffix
	modelPath := fmt.Sprintf("factions/%s/units/%s/models/%s_standing.g3d", faction, unit.UnitType, unit.UnitType)
	log.Printf("🔍 DEBUG: Attempting to load model: %s for unit %s (faction: %s)", modelPath, unit.UnitType, faction)
	g3dModel, err = r.assetMgr.LoadG3DModel(modelPath)

	if err != nil {
		// Pattern 2: Fallback - try without _standing suffix
		modelPath = fmt.Sprintf("factions/%s/units/%s/models/%s.g3d", faction, unit.UnitType, unit.UnitType)
		log.Printf("🔄 Fallback: Attempting model without _standing: %s", modelPath)
		g3dModel, err = r.assetMgr.LoadG3DModel(modelPath)

		if err != nil {
			// Model loading failed with both patterns - render placeholder
			log.Printf("❌ BOTH MODEL PATTERNS FAILED for unit %s:", unit.UnitType)
			log.Printf("   Pattern 1 error: %v", err)
			log.Printf("   Pattern 2 error: %v", err)
			log.Printf("   Rendering placeholder at (%.1f, %.1f, %.1f)", pos.X, pos.Y, pos.Z)

			// Use the terrain rendering to draw a simple colored indicator
			// This ensures units are ALWAYS visible even without proper models
			return r.renderUnitPlaceholder(unit, pos)
		} else {
			log.Printf("✅ SUCCESS: Loaded model with fallback pattern: %s", modelPath)
		}
	} else {
		log.Printf("✅ SUCCESS: Loaded model with primary pattern: %s", modelPath)
	}

	// Convert G3DModel to our internal Model format for rendering (using ModelManager's logic)
	log.Printf("🔄 Converting G3D model to internal format for unit %s...", unit.UnitType)
	model, err := graphics.NewModelFromG3D(g3dModel)
	if err != nil {
		log.Printf("❌ CONVERSION FAILED: G3D to internal model conversion failed for unit %s: %v", unit.UnitType, err)
		return r.renderUnitPlaceholder(unit, pos)
	}
	log.Printf("✅ CONVERSION SUCCESS: G3D model converted successfully for unit %s", unit.UnitType)

	// Create transformation matrix for unit position
	// TODO: Add rotation based on unit facing direction
	// TODO: Add animation state based on unit.State (moving, attacking, etc.)

	log.Printf("🎨 About to render model for unit %s at position (%.1f, %.1f, %.1f)...", unit.UnitType, pos.X, pos.Y, pos.Z)
	err = r.RenderModel(model)
	if err != nil {
		// If model rendering fails, fallback to placeholder
		log.Printf("❌ RENDER FAILED: OpenGL rendering failed for unit %d (%s): %v", unit.ID, unit.UnitType, err)
		return r.renderUnitPlaceholder(unit, pos)
	}
	log.Printf("✅ RENDER SUCCESS: Model rendered successfully for unit %s", unit.UnitType)

	return nil
}

// renderUnitPlaceholder renders a simple visible placeholder for units without models
func (r *Renderer) renderUnitPlaceholder(unit *engine.GameUnit, pos engine.Vector3) error {
	// Create a simple colored indicator that's definitely visible
	log.Printf("🔲 Rendering placeholder for unit %d ('%s') at (%.1f, %.1f, %.1f)",
		unit.ID, unit.UnitType, pos.X, pos.Y, pos.Z)

	// Choose color based on unit type for visual distinction
	var color [3]float32
	switch unit.UnitType {
	case "worker":
		color = [3]float32{0.8, 0.8, 0.2} // Yellow for workers
	case "guard", "archer":
		color = [3]float32{0.2, 0.8, 0.2} // Green for military
	case "initiate", "battlemage", "daemon":
		color = [3]float32{0.8, 0.2, 0.8} // Purple for magic
	default:
		color = [3]float32{0.5, 0.5, 0.5} // Gray for unknown
	}

	// Render a simple colored cube using the terrain shader
	// This ensures maximum compatibility with existing rendering pipeline
	return r.renderColoredCube(pos, color, 1.0) // 1.0 unit size
}

// renderColoredCube renders a simple colored cube at the given position
func (r *Renderer) renderColoredCube(pos engine.Vector3, color [3]float32, size float32) error {
	// Initialize basic shader if not done yet
	if r.basicShader == 0 {
		err := r.initializeBasicShader()
		if err != nil {
			return fmt.Errorf("failed to initialize basic shader: %v", err)
		}
	}

	// Use the basic shader to render a simple cube
	gl.UseProgram(r.basicShader)

	// Set up transformation matrix for the cube position
	translation := mgl32.Translate3D(float32(pos.X), float32(pos.Y), float32(pos.Z))
	scale := mgl32.Scale3D(size, size, size)
	modelMatrix := translation.Mul4(scale)

	// Set uniforms for the basic shader
	modelLoc := gl.GetUniformLocation(r.basicShader, gl.Str("model\x00"))
	gl.UniformMatrix4fv(modelLoc, 1, false, &modelMatrix[0])

	viewLoc := gl.GetUniformLocation(r.basicShader, gl.Str("view\x00"))
	gl.UniformMatrix4fv(viewLoc, 1, false, &r.camera.ViewMatrix[0])

	projLoc := gl.GetUniformLocation(r.basicShader, gl.Str("projection\x00"))
	gl.UniformMatrix4fv(projLoc, 1, false, &r.camera.ProjectionMatrix[0])

	// Set color uniform
	colorLoc := gl.GetUniformLocation(r.basicShader, gl.Str("color\x00"))
	gl.Uniform3f(colorLoc, color[0], color[1], color[2])

	// Render a simple cube (8 vertices, 12 triangles)
	if r.cubeVAO == 0 {
		r.initializeCubeGeometry()
	}

	gl.BindVertexArray(r.cubeVAO)
	gl.DrawElements(gl.TRIANGLES, 36, gl.UNSIGNED_INT, gl.PtrOffset(0))
	gl.BindVertexArray(0)

	return nil
}

// renderBuildings renders all buildings from the game world
func (r *Renderer) renderBuildings(world *engine.World) error {
	allPlayers := world.GetAllPlayers()

	for _, player := range allPlayers {
		buildings := world.ObjectManager.GetBuildingsForPlayer(player.ID)

		for _, building := range buildings {
			// Skip buildings that haven't finished construction
			if !building.IsBuilt {
				continue
			}

			err := r.renderBuilding(building)
			if err != nil {
				// Log error but continue rendering other buildings
				log.Printf("Warning: Failed to render building %d: %v", building.ID, err)
				continue
			}
		}
	}

	return nil
}

// renderBuilding renders a single building
func (r *Renderer) renderBuilding(building *engine.GameBuilding) error {
	// Get building position from game state (TODO: use for transformation)
	_ = building.Position

	// For now, try to load a G3D model for this building type
	// TODO: Cache loaded models and use proper asset management
	modelPath := fmt.Sprintf("factions/magic/buildings/%s/models/%s.g3d", building.BuildingType, building.BuildingType)

	model, err := r.LoadG3DModel(modelPath)
	if err != nil {
		// If specific model not found, use a default placeholder
		return nil // Skip rendering rather than error
	}

	// Create transformation matrix for building position and rotation
	// TODO: Apply building.Rotation for proper orientation

	err = r.RenderModel(model)
	if err != nil {
		return fmt.Errorf("failed to render building model: %w", err)
	}

	return nil
}

// renderResourceNodes renders resource nodes on the map
func (r *Renderer) renderResourceNodes(world *engine.World) error {
	resourceNodes := world.GetAllResourceNodes()

	for _, node := range resourceNodes {
		err := r.renderResourceNode(node)
		if err != nil {
			// Log error but continue rendering other nodes
			log.Printf("Warning: Failed to render resource node %d: %v", node.ID, err)
			continue
		}
	}

	return nil
}

// renderResourceNode renders a single resource node
func (r *Renderer) renderResourceNode(node *engine.ResourceNode) error {
	// Get resource node position (TODO: use for transformation)
	_ = node.Position

	// Try to load resource-specific model
	// TODO: Use different models for different resource types (gold, wood, stone)
	modelPath := fmt.Sprintf("resources/%s/%s.g3d", node.ResourceType, node.ResourceType)

	model, err := r.LoadG3DModel(modelPath)
	if err != nil {
		// If specific model not found, skip rendering
		return nil
	}

	err = r.RenderModel(model)
	if err != nil {
		return fmt.Errorf("failed to render resource node model: %w", err)
	}

	return nil
}

// Destroy cleans up the renderer and releases resources
func (r *Renderer) Destroy() {
	// Clean up GPU models
	for path, model := range r.modelCache {
		gl.DeleteVertexArrays(1, &model.VAO)
		gl.DeleteBuffers(1, &model.VBO)
		gl.DeleteBuffers(1, &model.EBO)
		log.Printf("Cleaned up model: %s", path)
	}

	// Clean up GPU textures
	for path, texture := range r.textureCache {
		gl.DeleteTextures(1, &texture.ID)
		log.Printf("Cleaned up texture: %s", path)
	}

	// Clean up model manager
	if r.modelMgr != nil {
		r.modelMgr.Cleanup()
	}

	// Clean up lighting manager (no cleanup needed - just references)
	r.lightMgr = nil

	// Clean up shader manager
	if r.shaderMgr != nil {
		r.shaderMgr.Destroy()
	}

	// Destroy OpenGL context
	r.context.Destroy()

	log.Printf("Renderer destroyed after %d frames", r.frameCount)
}