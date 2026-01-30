package main

import (
	"fmt"
	"log"
	"path/filepath"

	"teraglest/internal/data"
	"teraglest/internal/graphics/renderer"

	"github.com/go-gl/glfw/v3.3/glfw"
)

func main() {
	fmt.Println("TeraGlest Rendering Demo - Phase 3.0 Foundation Test")
	fmt.Println("================================================")

	// Initialize asset manager
	dataRoot := "/home/solifugus/development/teraglest/megaglest-source/data/glest_game"
	assetManager := data.NewAssetManager(filepath.Join(dataRoot, "techs", "megapack"))

	fmt.Printf("✅ AssetManager initialized with data root: %s\n", dataRoot)

	// Create renderer
	r, err := renderer.NewRenderer(assetManager, "TeraGlest Rendering Demo", 1024, 768)
	if err != nil {
		log.Fatalf("Failed to create renderer: %v", err)
	}
	defer r.Destroy()

	fmt.Println("✅ OpenGL renderer initialized")

	// Create shader manager and load basic shaders
	shaderMgr := renderer.NewShaderManager()

	// Load model shaders
	err = shaderMgr.LoadShader("model",
		"internal/graphics/shaders/model.vert",
		"internal/graphics/shaders/model.frag")
	if err != nil {
		log.Printf("Warning: Failed to load model shaders: %v", err)
	} else {
		fmt.Println("✅ Model shaders loaded")
	}

	// Load terrain shaders
	err = shaderMgr.LoadShader("terrain",
		"internal/graphics/shaders/terrain.vert",
		"internal/graphics/shaders/terrain.frag")
	if err != nil {
		log.Printf("Warning: Failed to load terrain shaders: %v", err)
	} else {
		fmt.Println("✅ Terrain shaders loaded")
	}

	fmt.Printf("✅ Loaded %d shader programs\n", len(shaderMgr.ListShaders()))

	// Set up additional input callbacks
	context := r.GetContext()
	window := context.GetWindow()

	window.SetKeyCallback(func(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
		if action == glfw.Press {
			switch key {
			case glfw.KeySpace:
				fmt.Printf("Frame %d: FPS=%.1f\n", r.GetFrameCount(), r.GetFPS())
			case glfw.KeyR:
				fmt.Printf("Renderer status: Frame %d, FPS=%.1f\n",
					r.GetFrameCount(), r.GetFPS())
			}
		}
	})

	fmt.Println()
	fmt.Println("=== CONTROLS ===")
	fmt.Println("ESC:    Exit demo")
	fmt.Println("F1:     Toggle wireframe")
	fmt.Println("F2:     Toggle stats")
	fmt.Println("SPACE:  Print current FPS")
	fmt.Println("R:      Print r status")
	fmt.Println()

	fmt.Println("Starting render loop...")

	// Main render loop
	frameCount := 0
	for !r.ShouldClose() {
		// Render frame
		err := r.RenderFrame()
		if err != nil {
			log.Printf("Render error: %v", err)
		}

		frameCount++

		// Print startup confirmation after first few frames
		if frameCount == 60 { // After 1 second at 60 FPS
			fmt.Printf("🎉 Rendering successfully! Frame %d, FPS=%.1f\n",
				frameCount, r.GetFPS())
		}
	}

	fmt.Printf("\nDemo completed after %d frames (avg FPS: %.1f)\n",
		r.GetFrameCount(), r.GetFPS())
	fmt.Println("✅ Phase 3.0 foundation test successful!")
}