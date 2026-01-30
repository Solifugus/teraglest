package renderer

import (
	"fmt"
	"log"
	"runtime"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

// RenderContext manages the OpenGL context and window
type RenderContext struct {
	window     *glfw.Window
	width      int
	height     int
	fullscreen bool
	title      string
}

// NewRenderContext creates a new OpenGL context and window
func NewRenderContext(title string, width, height int, fullscreen bool) (*RenderContext, error) {
	// GLFW requires the main OS thread for window operations
	runtime.LockOSThread()

	// Initialize GLFW
	if err := glfw.Init(); err != nil {
		return nil, fmt.Errorf("failed to initialize GLFW: %w", err)
	}

	// Set OpenGL context version to 3.3 core profile
	glfw.WindowHint(glfw.ContextVersionMajor, 3)
	glfw.WindowHint(glfw.ContextVersionMinor, 3)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

	// Enable depth buffer
	glfw.WindowHint(glfw.DepthBits, 24)

	// Enable multisampling for anti-aliasing
	glfw.WindowHint(glfw.Samples, 4)

	// Create window
	var window *glfw.Window
	var err error

	if fullscreen {
		monitor := glfw.GetPrimaryMonitor()
		videoMode := monitor.GetVideoMode()
		window, err = glfw.CreateWindow(videoMode.Width, videoMode.Height, title, monitor, nil)
		width = videoMode.Width
		height = videoMode.Height
	} else {
		window, err = glfw.CreateWindow(width, height, title, nil, nil)
	}

	if err != nil {
		glfw.Terminate()
		return nil, fmt.Errorf("failed to create window: %w", err)
	}

	// Make the OpenGL context current
	window.MakeContextCurrent()

	// Initialize OpenGL bindings
	if err := gl.Init(); err != nil {
		window.Destroy()
		glfw.Terminate()
		return nil, fmt.Errorf("failed to initialize OpenGL: %w", err)
	}

	// Enable V-Sync
	glfw.SwapInterval(1)

	// Set up OpenGL state
	gl.Enable(gl.DEPTH_TEST)
	gl.Enable(gl.CULL_FACE)
	gl.Enable(gl.MULTISAMPLE)
	gl.CullFace(gl.BACK)
	gl.FrontFace(gl.CCW)

	// Set clear color to a pleasant sky blue
	gl.ClearColor(0.53, 0.81, 0.92, 1.0)

	// Set viewport
	gl.Viewport(0, 0, int32(width), int32(height))

	// Set up callback for window resize
	rc := &RenderContext{
		window:     window,
		width:      width,
		height:     height,
		fullscreen: fullscreen,
		title:      title,
	}

	window.SetFramebufferSizeCallback(rc.onFramebufferResize)

	// Print OpenGL version info
	version := gl.GoStr(gl.GetString(gl.VERSION))
	renderer := gl.GoStr(gl.GetString(gl.RENDERER))
	log.Printf("OpenGL Version: %s", version)
	log.Printf("OpenGL Renderer: %s", renderer)

	return rc, nil
}

// onFramebufferResize handles window resize events
func (rc *RenderContext) onFramebufferResize(window *glfw.Window, width, height int) {
	rc.width = width
	rc.height = height
	gl.Viewport(0, 0, int32(width), int32(height))
}

// ShouldClose returns true if the window should close
func (rc *RenderContext) ShouldClose() bool {
	return rc.window.ShouldClose()
}

// SwapBuffers swaps the front and back buffers
func (rc *RenderContext) SwapBuffers() {
	rc.window.SwapBuffers()
}

// PollEvents processes pending events
func (rc *RenderContext) PollEvents() {
	glfw.PollEvents()
}

// GetWidth returns the current window width
func (rc *RenderContext) GetWidth() int {
	return rc.width
}

// GetHeight returns the current window height
func (rc *RenderContext) GetHeight() int {
	return rc.height
}

// GetAspectRatio returns the current aspect ratio
func (rc *RenderContext) GetAspectRatio() float32 {
	return float32(rc.width) / float32(rc.height)
}

// GetWindow returns the GLFW window handle
func (rc *RenderContext) GetWindow() *glfw.Window {
	return rc.window
}

// SetCursorInputMode sets the cursor input mode
func (rc *RenderContext) SetCursorInputMode(mode int) {
	rc.window.SetInputMode(glfw.CursorMode, mode)
}

// GetCursorPos returns the current cursor position
func (rc *RenderContext) GetCursorPos() (float64, float64) {
	return rc.window.GetCursorPos()
}

// IsKeyPressed returns true if the specified key is currently pressed
func (rc *RenderContext) IsKeyPressed(key glfw.Key) bool {
	return rc.window.GetKey(key) == glfw.Press
}

// IsMouseButtonPressed returns true if the specified mouse button is currently pressed
func (rc *RenderContext) IsMouseButtonPressed(button glfw.MouseButton) bool {
	return rc.window.GetMouseButton(button) == glfw.Press
}

// SetKeyCallback sets the key callback function
func (rc *RenderContext) SetKeyCallback(callback glfw.KeyCallback) {
	rc.window.SetKeyCallback(callback)
}

// SetMouseButtonCallback sets the mouse button callback function
func (rc *RenderContext) SetMouseButtonCallback(callback glfw.MouseButtonCallback) {
	rc.window.SetMouseButtonCallback(callback)
}

// SetScrollCallback sets the scroll callback function
func (rc *RenderContext) SetScrollCallback(callback glfw.ScrollCallback) {
	rc.window.SetScrollCallback(callback)
}

// SetCursorPosCallback sets the cursor position callback function
func (rc *RenderContext) SetCursorPosCallback(callback glfw.CursorPosCallback) {
	rc.window.SetCursorPosCallback(callback)
}

// Clear clears the screen
func (rc *RenderContext) Clear() {
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
}

// EnableWireframe enables wireframe rendering mode
func (rc *RenderContext) EnableWireframe() {
	gl.PolygonMode(gl.FRONT_AND_BACK, gl.LINE)
}

// DisableWireframe disables wireframe rendering mode
func (rc *RenderContext) DisableWireframe() {
	gl.PolygonMode(gl.FRONT_AND_BACK, gl.FILL)
}

// Destroy cleans up the context and terminates GLFW
func (rc *RenderContext) Destroy() {
	if rc.window != nil {
		rc.window.Destroy()
	}
	glfw.Terminate()
}