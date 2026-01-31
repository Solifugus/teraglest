package graphics

import (
	"testing"

	"github.com/go-gl/mathgl/mgl32"
)

// TestMaterialManager tests the material management system
func TestMaterialManager(t *testing.T) {
	materialMgr := NewMaterialManager()

	// Test initial state
	if materialMgr.GetMaterialCount() != 1 { // Should have default material
		t.Errorf("Expected 1 material initially (default), got %d", materialMgr.GetMaterialCount())
	}

	// Test default material
	defaultMat := materialMgr.GetMaterial("default")
	if defaultMat == nil {
		t.Fatal("Default material should exist")
	}

	if defaultMat.Type != BasicMaterial {
		t.Errorf("Expected BasicMaterial for default, got %d", defaultMat.Type)
	}

	if defaultMat.Name != "default" {
		t.Errorf("Expected name 'default', got '%s'", defaultMat.Name)
	}

	// Test creating basic material
	redMaterial := materialMgr.CreateBasicMaterial(
		"red_plastic",
		mgl32.Vec3{0.8, 0.1, 0.1}, // Red diffuse
		mgl32.Vec3{1.0, 1.0, 1.0}, // White specular
		64.0, // High shininess
	)

	if redMaterial == nil {
		t.Fatal("Failed to create red material")
	}

	if redMaterial.Type != BasicMaterial {
		t.Errorf("Expected BasicMaterial, got %d", redMaterial.Type)
	}

	if materialMgr.GetMaterialCount() != 2 {
		t.Errorf("Expected 2 materials after creation, got %d", materialMgr.GetMaterialCount())
	}

	// Test PBR material
	pbrMaterial := materialMgr.CreatePBRMaterial(
		"metal_surface",
		mgl32.Vec3{0.7, 0.7, 0.7}, // Gray base color
		0.9,  // High metallic
		0.1,  // Low roughness (shiny)
	)

	if pbrMaterial == nil {
		t.Fatal("Failed to create PBR material")
	}

	if pbrMaterial.Type != PBRMaterial {
		t.Errorf("Expected PBRMaterial, got %d", pbrMaterial.Type)
	}

	if pbrMaterial.Metallic != 0.9 {
		t.Errorf("Expected metallic 0.9, got %f", pbrMaterial.Metallic)
	}

	// Test emissive material
	glowMaterial := materialMgr.CreateEmissiveMaterial(
		"neon_light",
		mgl32.Vec3{0.0, 1.0, 0.3}, // Green glow
		2.0, // High intensity
	)

	if glowMaterial == nil {
		t.Fatal("Failed to create emissive material")
	}

	if glowMaterial.Type != EmissiveMaterial {
		t.Errorf("Expected EmissiveMaterial, got %d", glowMaterial.Type)
	}

	if glowMaterial.EmissiveStrength != 2.0 {
		t.Errorf("Expected emissive strength 2.0, got %f", glowMaterial.EmissiveStrength)
	}

	// Test material retrieval
	retrievedMat := materialMgr.GetMaterial("red_plastic")
	if retrievedMat == nil {
		t.Error("Failed to retrieve created material")
	}

	if retrievedMat.Name != "red_plastic" {
		t.Errorf("Retrieved wrong material: expected 'red_plastic', got '%s'", retrievedMat.Name)
	}

	// Test non-existent material (should return default)
	nonExistent := materialMgr.GetMaterial("does_not_exist")
	if nonExistent != defaultMat {
		t.Error("Non-existent material should return default material")
	}
}

// TestAdvancedMaterialProperties tests advanced material property manipulation
func TestAdvancedMaterialProperties(t *testing.T) {
	materialMgr := NewMaterialManager()

	// Create test material
	material := materialMgr.CreateBasicMaterial("test", mgl32.Vec3{1, 1, 1}, mgl32.Vec3{1, 1, 1}, 32.0)

	// Test texture assignment
	testTextureID := uint32(123)
	material.SetTexture(DiffuseTexture, testTextureID)

	if !material.HasTexture(DiffuseTexture) {
		t.Error("Material should have diffuse texture after assignment")
	}

	if material.GetTexture(DiffuseTexture) != testTextureID {
		t.Errorf("Expected texture ID %d, got %d", testTextureID, material.GetTexture(DiffuseTexture))
	}

	// Test multiple textures
	normalTextureID := uint32(456)
	material.SetTexture(NormalTexture, normalTextureID)
	material.Type = NormalMappedMaterial

	if materialMgr.GetShaderForMaterial(material) != "normal_mapped_material" {
		t.Errorf("Expected normal mapped shader for normal mapped material, got %s",
			materialMgr.GetShaderForMaterial(material))
	}

	// Test texture removal
	material.RemoveTexture(DiffuseTexture)
	if material.HasTexture(DiffuseTexture) {
		t.Error("Material should not have diffuse texture after removal")
	}

	// Test texture transform
	scale := mgl32.Vec2{2.0, 3.0}
	offset := mgl32.Vec2{0.1, 0.2}
	material.SetTextureTransform(scale, offset)

	if material.TextureScale != scale {
		t.Errorf("Expected texture scale %v, got %v", scale, material.TextureScale)
	}

	if material.TextureOffset != offset {
		t.Errorf("Expected texture offset %v, got %v", offset, material.TextureOffset)
	}
}

// TestMaterialCloning tests material cloning functionality
func TestMaterialCloning(t *testing.T) {
	materialMgr := NewMaterialManager()

	// Create original material
	original := materialMgr.CreatePBRMaterial("original", mgl32.Vec3{0.5, 0.3, 0.1}, 0.8, 0.3)
	original.SetTexture(DiffuseTexture, 100)
	original.SetTexture(NormalTexture, 200)
	original.SetTextureTransform(mgl32.Vec2{2, 2}, mgl32.Vec2{0.5, 0.5})

	// Clone the material
	clone := materialMgr.CloneMaterial("original", "clone")
	if clone == nil {
		t.Fatal("Failed to clone material")
	}

	// Test that clone has same properties
	if clone.Type != original.Type {
		t.Errorf("Clone type mismatch: expected %d, got %d", original.Type, clone.Type)
	}

	if clone.DiffuseColor != original.DiffuseColor {
		t.Errorf("Clone diffuse color mismatch: expected %v, got %v", original.DiffuseColor, clone.DiffuseColor)
	}

	if clone.Metallic != original.Metallic {
		t.Errorf("Clone metallic mismatch: expected %f, got %f", original.Metallic, clone.Metallic)
	}

	if clone.GetTexture(DiffuseTexture) != original.GetTexture(DiffuseTexture) {
		t.Error("Clone should have same diffuse texture")
	}

	if clone.TextureScale != original.TextureScale {
		t.Errorf("Clone texture scale mismatch: expected %v, got %v", original.TextureScale, clone.TextureScale)
	}

	// Test that they are independent
	clone.DiffuseColor = mgl32.Vec3{1, 0, 0}
	if original.DiffuseColor == clone.DiffuseColor {
		t.Error("Clone and original should be independent")
	}

	// Test that both exist in manager
	if materialMgr.GetMaterialCount() != 3 { // default + original + clone
		t.Errorf("Expected 3 materials, got %d", materialMgr.GetMaterialCount())
	}
}

// TestShaderMapping tests shader mapping for different material types
func TestShaderMapping(t *testing.T) {
	materialMgr := NewMaterialManager()

	// Test default mappings
	basicMat := materialMgr.CreateBasicMaterial("basic", mgl32.Vec3{1, 1, 1}, mgl32.Vec3{1, 1, 1}, 32)
	if materialMgr.GetShaderForMaterial(basicMat) != "basic_material" {
		t.Errorf("Expected 'basic_material' shader for BasicMaterial, got '%s'",
			materialMgr.GetShaderForMaterial(basicMat))
	}

	pbrMat := materialMgr.CreatePBRMaterial("pbr", mgl32.Vec3{1, 1, 1}, 0.5, 0.5)
	if materialMgr.GetShaderForMaterial(pbrMat) != "pbr_material" {
		t.Errorf("Expected 'pbr_material' shader for PBRMaterial, got '%s'",
			materialMgr.GetShaderForMaterial(pbrMat))
	}

	// Test custom shader mapping
	materialMgr.SetShaderMapping(BasicMaterial, "custom_basic_shader")
	if materialMgr.GetShaderForMaterial(basicMat) != "custom_basic_shader" {
		t.Errorf("Expected 'custom_basic_shader' after mapping change, got '%s'",
			materialMgr.GetShaderForMaterial(basicMat))
	}
}

// MockAdvancedShaderInterface for testing material application
type MockAdvancedShaderInterface struct {
	uniforms map[string]interface{}
}

func (m *MockAdvancedShaderInterface) SetUniformMat4(shaderName, uniformName string, matrix mgl32.Mat4) error {
	if m.uniforms == nil {
		m.uniforms = make(map[string]interface{})
	}
	m.uniforms[shaderName+"."+uniformName] = matrix
	return nil
}

func (m *MockAdvancedShaderInterface) SetUniformVec3(shaderName, uniformName string, vector mgl32.Vec3) error {
	if m.uniforms == nil {
		m.uniforms = make(map[string]interface{})
	}
	m.uniforms[shaderName+"."+uniformName] = vector
	return nil
}

func (m *MockAdvancedShaderInterface) SetUniformVec2(shaderName, uniformName string, vector mgl32.Vec2) error {
	if m.uniforms == nil {
		m.uniforms = make(map[string]interface{})
	}
	m.uniforms[shaderName+"."+uniformName] = vector
	return nil
}

func (m *MockAdvancedShaderInterface) SetUniformFloat(shaderName, uniformName string, value float32) error {
	if m.uniforms == nil {
		m.uniforms = make(map[string]interface{})
	}
	m.uniforms[shaderName+"."+uniformName] = value
	return nil
}

func (m *MockAdvancedShaderInterface) SetUniformInt(shaderName, uniformName string, value int32) error {
	if m.uniforms == nil {
		m.uniforms = make(map[string]interface{})
	}
	m.uniforms[shaderName+"."+uniformName] = value
	return nil
}

func (m *MockAdvancedShaderInterface) SetUniformBool(shaderName, uniformName string, value bool) error {
	if m.uniforms == nil {
		m.uniforms = make(map[string]interface{})
	}
	m.uniforms[shaderName+"."+uniformName] = value
	return nil
}

// TestMaterialApplication tests applying materials to shaders
func TestMaterialApplication(t *testing.T) {
	materialMgr := NewMaterialManager()
	mock := &MockAdvancedShaderInterface{}

	// Create test material with textures
	material := materialMgr.CreateTexturedMaterial("textured_test", 123)
	material.SetTexture(NormalTexture, 456)
	material.SetTexture(SpecularTexture, 789)

	// Apply material to shader
	err := material.ApplyMaterial(mock, "test_shader")
	if err != nil {
		t.Errorf("Failed to apply material: %v", err)
	}

	// Verify that uniforms were set
	expectedUniforms := []string{
		"test_shader.material.diffuse",
		"test_shader.material.specular",
		"test_shader.material.shininess",
		"test_shader.material.opacity",
		"test_shader.uUseDiffuseTexture",
		"test_shader.uUseNormalTexture",
		"test_shader.uUseSpecularTexture",
	}

	for _, uniformName := range expectedUniforms {
		if _, exists := mock.uniforms[uniformName]; !exists {
			t.Errorf("Expected uniform %s was not set", uniformName)
		}
	}

	// Verify texture usage flags
	diffuseFlag, exists := mock.uniforms["test_shader.uUseDiffuseTexture"]
	if !exists || diffuseFlag != true {
		t.Error("Diffuse texture usage flag should be true")
	}

	normalFlag, exists := mock.uniforms["test_shader.uUseNormalTexture"]
	if !exists || normalFlag != true {
		t.Error("Normal texture usage flag should be true")
	}

	emissiveFlag, exists := mock.uniforms["test_shader.uUseEmissiveTexture"]
	if !exists || emissiveFlag != false {
		t.Error("Emissive texture usage flag should be false")
	}

	t.Logf("Successfully set %d material uniforms", len(mock.uniforms))
}