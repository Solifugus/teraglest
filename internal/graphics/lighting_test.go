package graphics

import (
	"testing"

	"github.com/go-gl/mathgl/mgl32"
)

// TestLightManager tests the light management system
func TestLightManager(t *testing.T) {
	// Create light manager with max 4 lights
	lightMgr := NewLightManager(4)

	// Test initial state
	if lightMgr.GetLightCount() != 0 {
		t.Errorf("Expected 0 lights initially, got %d", lightMgr.GetLightCount())
	}

	if lightMgr.GetActiveLightCount() != 0 {
		t.Errorf("Expected 0 active lights initially, got %d", lightMgr.GetActiveLightCount())
	}

	// Test ambient light
	ambientColor, ambientStrength := lightMgr.GetAmbientLight()
	if ambientColor == (mgl32.Vec3{}) {
		t.Error("Expected non-zero ambient color")
	}

	if ambientStrength <= 0 {
		t.Errorf("Expected positive ambient strength, got %f", ambientStrength)
	}

	// Test creating directional light
	sunDirection := mgl32.Vec3{-1, -1, 0}.Normalize()
	sunColor := mgl32.Vec3{1, 0.9, 0.8}
	sunLight, err := lightMgr.CreateDirectionalLight(sunDirection, sunColor, 0.8)
	if err != nil {
		t.Errorf("Failed to create directional light: %v", err)
	}

	if sunLight.Type != DirectionalLight {
		t.Errorf("Expected DirectionalLight, got %d", sunLight.Type)
	}

	if !sunLight.Enabled {
		t.Error("Expected light to be enabled by default")
	}

	if lightMgr.GetLightCount() != 1 {
		t.Errorf("Expected 1 light after creation, got %d", lightMgr.GetLightCount())
	}

	// Test creating point light
	lampPosition := mgl32.Vec3{5, 3, 2}
	lampColor := mgl32.Vec3{1, 1, 0.9}
	lampLight, err := lightMgr.CreatePointLight(lampPosition, lampColor, 1.0, 10.0)
	if err != nil {
		t.Errorf("Failed to create point light: %v", err)
	}

	if lampLight.Type != PointLight {
		t.Errorf("Expected PointLight, got %d", lampLight.Type)
	}

	if lightMgr.GetLightCount() != 2 {
		t.Errorf("Expected 2 lights after creation, got %d", lightMgr.GetLightCount())
	}

	// Test creating spot light
	flashlightPos := mgl32.Vec3{0, 2, 5}
	flashlightDir := mgl32.Vec3{0, -0.5, -1}.Normalize()
	flashlightColor := mgl32.Vec3{1, 1, 1}
	spotLight, err := lightMgr.CreateSpotLight(flashlightPos, flashlightDir, flashlightColor, 0.8, 8.0, 0.2, 0.5)
	if err != nil {
		t.Errorf("Failed to create spot light: %v", err)
	}

	if spotLight.Type != SpotLight {
		t.Errorf("Expected SpotLight, got %d", spotLight.Type)
	}

	if lightMgr.GetLightCount() != 3 {
		t.Errorf("Expected 3 lights after creation, got %d", lightMgr.GetLightCount())
	}

	// Test active lights
	activeLights := lightMgr.GetActiveLights()
	if len(activeLights) != 3 {
		t.Errorf("Expected 3 active lights, got %d", len(activeLights))
	}

	// Test disabling a light
	err = lightMgr.SetLightEnabled(1, false)
	if err != nil {
		t.Errorf("Failed to disable light: %v", err)
	}

	if lightMgr.GetActiveLightCount() != 2 {
		t.Errorf("Expected 2 active lights after disabling one, got %d", lightMgr.GetActiveLightCount())
	}

	// Test removing a light
	err = lightMgr.RemoveLight(0)
	if err != nil {
		t.Errorf("Failed to remove light: %v", err)
	}

	if lightMgr.GetLightCount() != 2 {
		t.Errorf("Expected 2 lights after removal, got %d", lightMgr.GetLightCount())
	}

	// Test light limit
	for i := 0; i < 10; i++ {
		_, err = lightMgr.CreateDirectionalLight(mgl32.Vec3{0, -1, 0}, mgl32.Vec3{1, 1, 1}, 0.5)
		if err != nil && i < 2 { // We should be able to create at least 2 more (4 total - 2 existing)
			t.Errorf("Failed to create light %d when under limit: %v", i, err)
		}
		if err != nil && i >= 2 { // After limit, should get error
			break
		}
	}

	// Should have reached the limit
	if lightMgr.GetLightCount() != 4 {
		t.Errorf("Expected to reach light limit of 4, got %d", lightMgr.GetLightCount())
	}
}

// TestDefaultLighting tests the default lighting setup
func TestDefaultLighting(t *testing.T) {
	lightMgr := NewLightManager(8)

	err := lightMgr.CreateDefaultLighting()
	if err != nil {
		t.Errorf("Failed to create default lighting: %v", err)
	}

	// Should have created some lights
	if lightMgr.GetLightCount() == 0 {
		t.Error("Default lighting should create at least one light")
	}

	// Should have at least one directional light (sun)
	activeLights := lightMgr.GetActiveLights()
	hasDirectional := false
	for _, light := range activeLights {
		if light.Type == DirectionalLight {
			hasDirectional = true
			break
		}
	}

	if !hasDirectional {
		t.Error("Default lighting should include at least one directional light")
	}

	// Test lighting info
	info := lightMgr.GetLightingInfo()
	if len(info) == 0 {
		t.Error("Lighting info should not be empty")
	}

	t.Logf("Default lighting setup:\n%s", info)
}

// TestLightManipulation tests light property manipulation
func TestLightManipulation(t *testing.T) {
	lightMgr := NewLightManager(4)

	// Create a test light
	originalPos := mgl32.Vec3{1, 2, 3}
	light, err := lightMgr.CreatePointLight(originalPos, mgl32.Vec3{1, 1, 1}, 1.0, 5.0)
	if err != nil {
		t.Errorf("Failed to create test light: %v", err)
	}

	// Test position update
	newPos := mgl32.Vec3{4, 5, 6}
	err = lightMgr.SetLightPosition(0, newPos)
	if err != nil {
		t.Errorf("Failed to set light position: %v", err)
	}

	if light.Position != newPos {
		t.Errorf("Light position not updated correctly: expected %v, got %v", newPos, light.Position)
	}

	// Test direction update
	newDir := mgl32.Vec3{0, -1, 0}
	err = lightMgr.SetLightDirection(0, newDir)
	if err != nil {
		t.Errorf("Failed to set light direction: %v", err)
	}

	// Direction should be normalized
	expectedDir := newDir.Normalize()
	if light.Direction != expectedDir {
		t.Errorf("Light direction not updated correctly: expected %v, got %v", expectedDir, light.Direction)
	}

	// Test color update
	newColor := mgl32.Vec3{0.8, 0.6, 0.4}
	newIntensity := float32(1.5)
	err = lightMgr.SetLightColor(0, newColor, newIntensity)
	if err != nil {
		t.Errorf("Failed to set light color: %v", err)
	}

	if light.Color != newColor {
		t.Errorf("Light color not updated correctly: expected %v, got %v", newColor, light.Color)
	}

	if light.Intensity != newIntensity {
		t.Errorf("Light intensity not updated correctly: expected %f, got %f", newIntensity, light.Intensity)
	}

	// Test ambient light modification
	newAmbient := mgl32.Vec3{0.3, 0.3, 0.4}
	newAmbientStrength := float32(0.6)
	lightMgr.SetAmbientLight(newAmbient, newAmbientStrength)

	ambient, strength := lightMgr.GetAmbientLight()
	if ambient != newAmbient {
		t.Errorf("Ambient color not updated correctly: expected %v, got %v", newAmbient, ambient)
	}

	if strength != newAmbientStrength {
		t.Errorf("Ambient strength not updated correctly: expected %f, got %f", newAmbientStrength, strength)
	}
}

// MockShaderInterface for testing lighting uniform updates
type MockLightingShaderInterface struct {
	uniforms map[string]interface{}
}

func (m *MockLightingShaderInterface) SetUniformMat4(shaderName, uniformName string, matrix mgl32.Mat4) error {
	if m.uniforms == nil {
		m.uniforms = make(map[string]interface{})
	}
	m.uniforms[shaderName+"."+uniformName] = matrix
	return nil
}

func (m *MockLightingShaderInterface) SetUniformVec3(shaderName, uniformName string, vector mgl32.Vec3) error {
	if m.uniforms == nil {
		m.uniforms = make(map[string]interface{})
	}
	m.uniforms[shaderName+"."+uniformName] = vector
	return nil
}

func (m *MockLightingShaderInterface) SetUniformVec2(shaderName, uniformName string, vector mgl32.Vec2) error {
	if m.uniforms == nil {
		m.uniforms = make(map[string]interface{})
	}
	m.uniforms[shaderName+"."+uniformName] = vector
	return nil
}

func (m *MockLightingShaderInterface) SetUniformFloat(shaderName, uniformName string, value float32) error {
	if m.uniforms == nil {
		m.uniforms = make(map[string]interface{})
	}
	m.uniforms[shaderName+"."+uniformName] = value
	return nil
}

func (m *MockLightingShaderInterface) SetUniformInt(shaderName, uniformName string, value int32) error {
	if m.uniforms == nil {
		m.uniforms = make(map[string]interface{})
	}
	m.uniforms[shaderName+"."+uniformName] = value
	return nil
}

func (m *MockLightingShaderInterface) SetUniformBool(shaderName, uniformName string, value bool) error {
	if m.uniforms == nil {
		m.uniforms = make(map[string]interface{})
	}
	m.uniforms[shaderName+"."+uniformName] = value
	return nil
}

// TestShaderUniformUpdates tests shader uniform updates for lighting
func TestShaderUniformUpdates(t *testing.T) {
	lightMgr := NewLightManager(4)
	mock := &MockLightingShaderInterface{}

	// Create some test lights
	_, err := lightMgr.CreateDirectionalLight(mgl32.Vec3{0, -1, 0}, mgl32.Vec3{1, 1, 1}, 0.8)
	if err != nil {
		t.Errorf("Failed to create directional light: %v", err)
	}

	_, err = lightMgr.CreatePointLight(mgl32.Vec3{5, 5, 5}, mgl32.Vec3{1, 0.5, 0.2}, 1.0, 8.0)
	if err != nil {
		t.Errorf("Failed to create point light: %v", err)
	}

	// Update shader uniforms
	err = lightMgr.UpdateShaderUniforms(mock, "test_shader")
	if err != nil {
		t.Errorf("Failed to update shader uniforms: %v", err)
	}

	// Verify that uniforms were set
	expectedUniforms := []string{
		"test_shader.uNumLights",
		"test_shader.uAmbientColor",
		"test_shader.uLights[0].type",
		"test_shader.uLights[0].color",
		"test_shader.uLights[1].type",
		"test_shader.uLights[1].position",
	}

	for _, uniformName := range expectedUniforms {
		if _, exists := mock.uniforms[uniformName]; !exists {
			t.Errorf("Expected uniform %s was not set", uniformName)
		}
	}

	// Verify light count
	lightCount, exists := mock.uniforms["test_shader.uNumLights"]
	if !exists {
		t.Error("Light count uniform not set")
	} else if lightCount != int32(2) {
		t.Errorf("Expected light count 2, got %v", lightCount)
	}

	t.Logf("Successfully set %d shader uniforms for lighting", len(mock.uniforms))
}