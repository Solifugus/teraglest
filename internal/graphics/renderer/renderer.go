package renderer

import (
	"fmt"
	"log"
	"time"

	"teraglest/internal/data"
	"teraglest/internal/engine"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
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
	context   *RenderContext
	assetMgr  *data.AssetManager

	// GPU resource caches
	modelCache   map[string]*GPUModel   // Path -> GPU model
	textureCache map[string]*GPUTexture // Path -> GPU texture

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

	renderer := &Renderer{
		context:       context,
		assetMgr:      assetMgr,
		modelCache:    make(map[string]*GPUModel),
		textureCache:  make(map[string]*GPUTexture),
		lastFrameTime: time.Now(),
		wireframe:     false,
		showStats:     true,
	}

	// Set up input callbacks
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

// ShouldClose returns true if the window should close
func (r *Renderer) ShouldClose() bool {
	return r.context.ShouldClose()
}

// GetContext returns the render context
func (r *Renderer) GetContext() *RenderContext {
	return r.context
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

	// TODO: Actual rendering will go here
	// For now, just clear to a blue background to show the window is working

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

	// TODO: Implement world rendering
	// This will include:
	// 1. Terrain rendering from world.Map and heightMap
	// 2. Unit rendering from world.ObjectManager.GetAllUnits()
	// 3. Building rendering from world.ObjectManager buildings
	// 4. Resource node rendering from world.GetAllResourceNodes()

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

	// Destroy OpenGL context
	r.context.Destroy()

	log.Printf("Renderer destroyed after %d frames", r.frameCount)
}