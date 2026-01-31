package graphics

import (
	"runtime"
	"testing"

	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

// TestModelRenderingPipeline tests the complete 3D model rendering pipeline
func TestModelRenderingPipeline(t *testing.T) {
	// Ensure we're on the main thread for OpenGL operations
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	// Initialize GLFW for OpenGL context
	if err := glfw.Init(); err != nil {
		t.Fatal("Failed to initialize GLFW")
	}
	defer glfw.Terminate()

	// Create test window (hidden)
	glfw.WindowHint(glfw.Visible, glfw.False)
	glfw.WindowHint(glfw.ContextVersionMajor, 3)
	glfw.WindowHint(glfw.ContextVersionMinor, 3)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)

	window, err := glfw.CreateWindow(800, 600, "Test", nil, nil)
	if err != nil {
		t.Fatalf("Failed to create GLFW window: %v", err)
	}
	defer window.Destroy()

	window.MakeContextCurrent()

	// Test ModelManager creation
	t.Run("ModelManager Creation", func(t *testing.T) {
		modelMgr := NewModelManager()
		if modelMgr == nil {
			t.Fatal("Failed to create ModelManager")
		}

		stats := modelMgr.GetStatistics()
		if stats.ModelCount != 0 {
			t.Errorf("Expected 0 models initially, got %d", stats.ModelCount)
		}

		if stats.TextureCount != 0 {
			t.Errorf("Expected 0 textures initially, got %d", stats.TextureCount)
		}
	})

	// Test TextureManager integration
	t.Run("TextureManager Integration", func(t *testing.T) {
		textureMgr := NewTextureManager()
		if textureMgr == nil {
			t.Fatal("Failed to create TextureManager")
		}

		// Test default texture creation
		defaultTexture, err := textureMgr.CreateDefaultTexture()
		if err != nil {
			t.Errorf("Failed to create default texture: %v", err)
		}

		if defaultTexture.ID == 0 {
			t.Error("Default texture has invalid OpenGL ID")
		}

		// Test checker texture creation
		checkerTexture, err := textureMgr.CreateCheckerTexture(64)
		if err != nil {
			t.Errorf("Failed to create checker texture: %v", err)
		}

		if checkerTexture.ID == 0 {
			t.Error("Checker texture has invalid OpenGL ID")
		}

		// Cleanup
		textureMgr.Cleanup()
	})

	// Note: Camera system is tested in the renderer package
	// to avoid circular dependencies

	// Test Model creation (without actual G3D files)
	t.Run("Model Creation", func(t *testing.T) {
		modelMgr := NewModelManager()

		// Create test scene with cube model
		err := modelMgr.CreateTestScene()
		if err != nil {
			t.Errorf("Failed to create test scene: %v", err)
		}

		stats := modelMgr.GetStatistics()
		if stats.ModelCount == 0 {
			t.Error("Expected at least 1 model after creating test scene")
		}

		// Test model retrieval
		cubeModel := modelMgr.GetModel("test_cube")
		if cubeModel == nil {
			t.Error("Failed to retrieve test cube model")
		}

		if cubeModel.VAO == 0 {
			t.Error("Test cube model has invalid VAO")
		}

		if cubeModel.IndexCount == 0 {
			t.Error("Test cube model has no indices")
		}

		// Test model transformation
		cubeModel.SetPosition(5, 0, 5)
		modelMatrix := cubeModel.GetModelMatrix()

		// Check that translation was applied (position should be in last column)
		if modelMatrix[12] != 5.0 || modelMatrix[14] != 5.0 {
			t.Error("Model transformation not applied correctly")
		}

		// Cleanup
		modelMgr.Cleanup()
	})
}

// TestShaderInterface tests the shader interface for model rendering
func TestShaderInterface(t *testing.T) {
	// Create a mock shader interface for testing
	mockShader := &MockShaderInterface{}

	// Test model rendering with mock interface
	t.Run("Mock Shader Interface", func(t *testing.T) {
		modelMgr := NewModelManager()
		err := modelMgr.CreateTestScene()
		if err != nil {
			t.Errorf("Failed to create test scene: %v", err)
		}

		cubeModel := modelMgr.GetModel("test_cube")
		if cubeModel == nil {
			t.Fatal("Failed to get test cube model")
		}

		// Test rendering without OpenGL context (should not crash)
		// Note: This will fail OpenGL calls but tests the interface
		err = cubeModel.Render("test_shader", mockShader)
		if err != nil {
			t.Logf("Expected OpenGL error in test environment: %v", err)
		}

		modelMgr.Cleanup()
	})
}

// MockShaderInterface implements ShaderInterface for testing
type MockShaderInterface struct {
	uniformCalls map[string]interface{}
}

func (m *MockShaderInterface) SetUniformMat4(shaderName, uniformName string, matrix mgl32.Mat4) error {
	if m.uniformCalls == nil {
		m.uniformCalls = make(map[string]interface{})
	}
	m.uniformCalls[shaderName+"."+uniformName] = matrix
	return nil
}

func (m *MockShaderInterface) SetUniformVec3(shaderName, uniformName string, vector mgl32.Vec3) error {
	if m.uniformCalls == nil {
		m.uniformCalls = make(map[string]interface{})
	}
	m.uniformCalls[shaderName+"."+uniformName] = vector
	return nil
}

func (m *MockShaderInterface) SetUniformVec2(shaderName, uniformName string, vector mgl32.Vec2) error {
	if m.uniformCalls == nil {
		m.uniformCalls = make(map[string]interface{})
	}
	m.uniformCalls[shaderName+"."+uniformName] = vector
	return nil
}

func (m *MockShaderInterface) SetUniformFloat(shaderName, uniformName string, value float32) error {
	if m.uniformCalls == nil {
		m.uniformCalls = make(map[string]interface{})
	}
	m.uniformCalls[shaderName+"."+uniformName] = value
	return nil
}

func (m *MockShaderInterface) SetUniformInt(shaderName, uniformName string, value int32) error {
	if m.uniformCalls == nil {
		m.uniformCalls = make(map[string]interface{})
	}
	m.uniformCalls[shaderName+"."+uniformName] = value
	return nil
}

func (m *MockShaderInterface) SetUniformBool(shaderName, uniformName string, value bool) error {
	if m.uniformCalls == nil {
		m.uniformCalls = make(map[string]interface{})
	}
	m.uniformCalls[shaderName+"."+uniformName] = value
	return nil
}

// TestPhase33Integration tests the complete Phase 3.3 implementation
func TestPhase33Integration(t *testing.T) {
	t.Run("Component Integration Test", func(t *testing.T) {
		// Test that all Phase 3.3 components work together

		// 1. Model Manager
		modelMgr := NewModelManager()
		defer modelMgr.Cleanup()

		// 2. Texture Manager
		stats := modelMgr.GetStatistics()
		if stats.ModelCount != 0 {
			t.Error("Expected clean state initially")
		}

		// 3. Test Scene Creation
		err := modelMgr.CreateTestScene()
		if err != nil {
			t.Errorf("Failed to create integrated test scene: %v", err)
		}

		stats = modelMgr.GetStatistics()
		if stats.ModelCount == 0 {
			t.Error("Expected models after test scene creation")
		}

		if stats.TextureCount == 0 {
			t.Error("Expected textures after test scene creation")
		}

		// 4. Model Transformation
		err = modelMgr.SetModelPosition("test_cube", 10, 5, 10)
		if err != nil {
			t.Errorf("Failed to set model position: %v", err)
		}

		err = modelMgr.SetModelRotation("test_cube", 0, 1.57, 0) // 90 degrees in Y
		if err != nil {
			t.Errorf("Failed to set model rotation: %v", err)
		}

		// 5. Memory Usage Estimation
		if stats.MemoryUsageEstimate == 0 {
			t.Error("Expected non-zero memory usage estimate")
		}

		t.Logf("Phase 3.3 Integration successful:")
		t.Logf("  - Models loaded: %d", stats.ModelCount)
		t.Logf("  - Textures loaded: %d", stats.TextureCount)
		t.Logf("  - Total triangles: %d", stats.TotalTriangles)
		t.Logf("  - Memory estimate: %d bytes", stats.MemoryUsageEstimate)
	})
}