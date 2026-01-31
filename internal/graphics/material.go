package graphics

import (
	"fmt"

	"github.com/go-gl/mathgl/mgl32"
)

// MaterialType represents different types of materials
type MaterialType int

const (
	BasicMaterial      MaterialType = 0  // Simple diffuse + specular
	TexturedMaterial   MaterialType = 1  // Diffuse texture + specular
	NormalMappedMaterial MaterialType = 2  // Diffuse + normal map
	PBRMaterial       MaterialType = 3  // Physically Based Rendering
	EmissiveMaterial  MaterialType = 4  // Glowing/emissive material
	MetallicMaterial  MaterialType = 5  // Metallic surfaces
)

// TextureSlot represents different texture slots
type TextureSlot int

const (
	DiffuseTexture   TextureSlot = 0
	NormalTexture    TextureSlot = 1
	SpecularTexture  TextureSlot = 2
	EmissiveTexture  TextureSlot = 3
	MetallicTexture  TextureSlot = 4
	RoughnessTexture TextureSlot = 5
	AOTexture        TextureSlot = 6  // Ambient Occlusion
	HeightTexture    TextureSlot = 7  // Height/Displacement
)

// AdvancedMaterial represents an enhanced material with multiple texture support
type AdvancedMaterial struct {
	// Basic properties
	Type    MaterialType
	Name    string

	// Color properties
	DiffuseColor   mgl32.Vec3  // Base diffuse color
	SpecularColor  mgl32.Vec3  // Specular reflection color
	EmissiveColor  mgl32.Vec3  // Self-illumination color

	// Surface properties
	Shininess      float32     // Specular shininess/power
	Metallic       float32     // Metallic factor (0-1)
	Roughness      float32     // Surface roughness (0-1)
	Opacity        float32     // Material transparency (0-1)

	// Advanced properties
	NormalStrength    float32  // Normal map intensity
	EmissiveStrength  float32  // Emissive intensity
	HeightScale       float32  // Height/parallax scale

	// Texture assignments
	Textures       map[TextureSlot]uint32  // Texture slot -> OpenGL texture ID
	TextureEnabled map[TextureSlot]bool    // Which texture slots are active

	// Tiling and offset
	TextureScale   mgl32.Vec2  // UV scale for texture tiling
	TextureOffset  mgl32.Vec2  // UV offset for texture shifting

	// Rendering flags
	DoubleSided    bool        // Disable backface culling
	CastShadows    bool        // Whether this material casts shadows
	ReceiveShadows bool        // Whether this material receives shadows
	Transparent    bool        // Whether this material needs alpha blending
}

// MaterialManager manages advanced materials and their shaders
type MaterialManager struct {
	materials    map[string]*AdvancedMaterial  // Name -> material
	shaderMap    map[MaterialType]string       // Material type -> shader name
	defaultMat   *AdvancedMaterial             // Default fallback material
}

// NewMaterialManager creates a new material manager
func NewMaterialManager() *MaterialManager {
	mm := &MaterialManager{
		materials: make(map[string]*AdvancedMaterial),
		shaderMap: make(map[MaterialType]string),
	}

	// Setup default shader mappings
	mm.shaderMap[BasicMaterial] = "basic_material"
	mm.shaderMap[TexturedMaterial] = "textured_material"
	mm.shaderMap[NormalMappedMaterial] = "normal_mapped_material"
	mm.shaderMap[PBRMaterial] = "pbr_material"
	mm.shaderMap[EmissiveMaterial] = "emissive_material"
	mm.shaderMap[MetallicMaterial] = "metallic_material"

	// Create default material
	mm.defaultMat = mm.createDefaultMaterial()
	mm.materials["default"] = mm.defaultMat

	return mm
}

// createDefaultMaterial creates a basic default material
func (mm *MaterialManager) createDefaultMaterial() *AdvancedMaterial {
	return &AdvancedMaterial{
		Type:           BasicMaterial,
		Name:           "default",
		DiffuseColor:   mgl32.Vec3{0.8, 0.8, 0.8},  // Light gray
		SpecularColor:  mgl32.Vec3{1.0, 1.0, 1.0},  // White specular
		EmissiveColor:  mgl32.Vec3{0.0, 0.0, 0.0},  // No emission
		Shininess:      32.0,
		Metallic:       0.0,
		Roughness:      0.5,
		Opacity:        1.0,
		NormalStrength: 1.0,
		EmissiveStrength: 0.0,
		HeightScale:    0.02,
		Textures:       make(map[TextureSlot]uint32),
		TextureEnabled: make(map[TextureSlot]bool),
		TextureScale:   mgl32.Vec2{1.0, 1.0},
		TextureOffset:  mgl32.Vec2{0.0, 0.0},
		DoubleSided:    false,
		CastShadows:    true,
		ReceiveShadows: true,
		Transparent:    false,
	}
}

// CreateBasicMaterial creates a simple diffuse + specular material
func (mm *MaterialManager) CreateBasicMaterial(name string, diffuse, specular mgl32.Vec3, shininess float32) *AdvancedMaterial {
	material := &AdvancedMaterial{
		Type:           BasicMaterial,
		Name:           name,
		DiffuseColor:   diffuse,
		SpecularColor:  specular,
		EmissiveColor:  mgl32.Vec3{0.0, 0.0, 0.0},
		Shininess:      shininess,
		Metallic:       0.0,
		Roughness:      1.0 - (shininess / 128.0), // Convert shininess to roughness
		Opacity:        1.0,
		NormalStrength: 1.0,
		EmissiveStrength: 0.0,
		HeightScale:    0.02,
		Textures:       make(map[TextureSlot]uint32),
		TextureEnabled: make(map[TextureSlot]bool),
		TextureScale:   mgl32.Vec2{1.0, 1.0},
		TextureOffset:  mgl32.Vec2{0.0, 0.0},
		DoubleSided:    false,
		CastShadows:    true,
		ReceiveShadows: true,
		Transparent:    false,
	}

	mm.materials[name] = material
	return material
}

// CreateTexturedMaterial creates a material with diffuse texture
func (mm *MaterialManager) CreateTexturedMaterial(name string, diffuseTexture uint32) *AdvancedMaterial {
	material := mm.CreateBasicMaterial(name, mgl32.Vec3{1, 1, 1}, mgl32.Vec3{1, 1, 1}, 32.0)
	material.Type = TexturedMaterial
	material.SetTexture(DiffuseTexture, diffuseTexture)
	return material
}

// CreatePBRMaterial creates a physically based rendering material
func (mm *MaterialManager) CreatePBRMaterial(name string, baseColor mgl32.Vec3, metallic, roughness float32) *AdvancedMaterial {
	material := &AdvancedMaterial{
		Type:           PBRMaterial,
		Name:           name,
		DiffuseColor:   baseColor,
		SpecularColor:  mgl32.Vec3{0.04, 0.04, 0.04}, // Dielectric default
		EmissiveColor:  mgl32.Vec3{0.0, 0.0, 0.0},
		Shininess:      (2.0 / (roughness * roughness)) - 2.0, // Convert roughness to shininess
		Metallic:       metallic,
		Roughness:      roughness,
		Opacity:        1.0,
		NormalStrength: 1.0,
		EmissiveStrength: 0.0,
		HeightScale:    0.02,
		Textures:       make(map[TextureSlot]uint32),
		TextureEnabled: make(map[TextureSlot]bool),
		TextureScale:   mgl32.Vec2{1.0, 1.0},
		TextureOffset:  mgl32.Vec2{0.0, 0.0},
		DoubleSided:    false,
		CastShadows:    true,
		ReceiveShadows: true,
		Transparent:    false,
	}

	mm.materials[name] = material
	return material
}

// CreateEmissiveMaterial creates a glowing material
func (mm *MaterialManager) CreateEmissiveMaterial(name string, emissiveColor mgl32.Vec3, strength float32) *AdvancedMaterial {
	material := mm.CreateBasicMaterial(name, mgl32.Vec3{0.1, 0.1, 0.1}, mgl32.Vec3{0, 0, 0}, 1.0)
	material.Type = EmissiveMaterial
	material.EmissiveColor = emissiveColor
	material.EmissiveStrength = strength
	material.CastShadows = false // Emissive materials typically don't cast shadows
	return material
}

// SetTexture assigns a texture to a specific slot
func (mat *AdvancedMaterial) SetTexture(slot TextureSlot, textureID uint32) {
	mat.Textures[slot] = textureID
	mat.TextureEnabled[slot] = textureID != 0
}

// RemoveTexture removes a texture from a slot
func (mat *AdvancedMaterial) RemoveTexture(slot TextureSlot) {
	delete(mat.Textures, slot)
	mat.TextureEnabled[slot] = false
}

// HasTexture checks if a material has a texture in a specific slot
func (mat *AdvancedMaterial) HasTexture(slot TextureSlot) bool {
	return mat.TextureEnabled[slot] && mat.Textures[slot] != 0
}

// GetTexture returns the texture ID for a slot
func (mat *AdvancedMaterial) GetTexture(slot TextureSlot) uint32 {
	if !mat.HasTexture(slot) {
		return 0
	}
	return mat.Textures[slot]
}

// SetTextureTransform sets texture scaling and offset
func (mat *AdvancedMaterial) SetTextureTransform(scale, offset mgl32.Vec2) {
	mat.TextureScale = scale
	mat.TextureOffset = offset
}

// ApplyMaterial applies this material's properties to a shader
func (mat *AdvancedMaterial) ApplyMaterial(shaderInterface ShaderInterface, shaderName string) error {
	// Basic material properties
	err := shaderInterface.SetUniformVec3(shaderName, "material.diffuse", mat.DiffuseColor)
	if err != nil {
		return fmt.Errorf("failed to set diffuse color: %w", err)
	}

	err = shaderInterface.SetUniformVec3(shaderName, "material.specular", mat.SpecularColor)
	if err != nil {
		return fmt.Errorf("failed to set specular color: %w", err)
	}

	err = shaderInterface.SetUniformVec3(shaderName, "material.emissive", mat.EmissiveColor)
	if err != nil {
		// Ignore if uniform doesn't exist (basic shaders might not have it)
	}

	err = shaderInterface.SetUniformFloat(shaderName, "material.shininess", mat.Shininess)
	if err != nil {
		return fmt.Errorf("failed to set shininess: %w", err)
	}

	err = shaderInterface.SetUniformFloat(shaderName, "material.opacity", mat.Opacity)
	if err != nil {
		return fmt.Errorf("failed to set opacity: %w", err)
	}

	// Advanced properties (ignore errors for basic shaders)
	shaderInterface.SetUniformFloat(shaderName, "material.metallic", mat.Metallic)
	shaderInterface.SetUniformFloat(shaderName, "material.roughness", mat.Roughness)
	shaderInterface.SetUniformFloat(shaderName, "material.normalStrength", mat.NormalStrength)
	shaderInterface.SetUniformFloat(shaderName, "material.emissiveStrength", mat.EmissiveStrength)

	// Texture transform
	shaderInterface.SetUniformVec2(shaderName, "material.textureScale", mat.TextureScale)
	shaderInterface.SetUniformVec2(shaderName, "material.textureOffset", mat.TextureOffset)

	// Texture bindings
	textureUnit := int32(0)
	for slot, textureID := range mat.Textures {
		if !mat.TextureEnabled[slot] || textureID == 0 {
			continue
		}

		// Bind texture
		// gl.ActiveTexture(gl.TEXTURE0 + uint32(textureUnit))
		// gl.BindTexture(gl.TEXTURE_2D, textureID)

		// Set shader uniform
		var uniformName string
		switch slot {
		case DiffuseTexture:
			uniformName = "uDiffuseTexture"
		case NormalTexture:
			uniformName = "uNormalTexture"
		case SpecularTexture:
			uniformName = "uSpecularTexture"
		case EmissiveTexture:
			uniformName = "uEmissiveTexture"
		case MetallicTexture:
			uniformName = "uMetallicTexture"
		case RoughnessTexture:
			uniformName = "uRoughnessTexture"
		case AOTexture:
			uniformName = "uAOTexture"
		case HeightTexture:
			uniformName = "uHeightTexture"
		}

		if uniformName != "" {
			shaderInterface.SetUniformInt(shaderName, uniformName, textureUnit)
			textureUnit++
		}
	}

	// Texture enabled flags
	shaderInterface.SetUniformBool(shaderName, "uUseDiffuseTexture", mat.HasTexture(DiffuseTexture))
	shaderInterface.SetUniformBool(shaderName, "uUseNormalTexture", mat.HasTexture(NormalTexture))
	shaderInterface.SetUniformBool(shaderName, "uUseSpecularTexture", mat.HasTexture(SpecularTexture))
	shaderInterface.SetUniformBool(shaderName, "uUseEmissiveTexture", mat.HasTexture(EmissiveTexture))

	return nil
}

// GetMaterial returns a material by name
func (mm *MaterialManager) GetMaterial(name string) *AdvancedMaterial {
	if material, exists := mm.materials[name]; exists {
		return material
	}
	return mm.defaultMat // Return default if not found
}

// GetShaderForMaterial returns the appropriate shader for a material type
func (mm *MaterialManager) GetShaderForMaterial(material *AdvancedMaterial) string {
	if shaderName, exists := mm.shaderMap[material.Type]; exists {
		return shaderName
	}
	return "advanced_model" // Default shader
}

// ListMaterials returns all material names
func (mm *MaterialManager) ListMaterials() []string {
	names := make([]string, 0, len(mm.materials))
	for name := range mm.materials {
		names = append(names, name)
	}
	return names
}

// GetMaterialCount returns the number of loaded materials
func (mm *MaterialManager) GetMaterialCount() int {
	return len(mm.materials)
}

// RemoveMaterial removes a material from the manager
func (mm *MaterialManager) RemoveMaterial(name string) {
	if name != "default" { // Don't allow removing default material
		delete(mm.materials, name)
	}
}

// SetShaderMapping sets the shader to use for a specific material type
func (mm *MaterialManager) SetShaderMapping(materialType MaterialType, shaderName string) {
	mm.shaderMap[materialType] = shaderName
}

// CloneMaterial creates a copy of an existing material with a new name
func (mm *MaterialManager) CloneMaterial(originalName, newName string) *AdvancedMaterial {
	original := mm.GetMaterial(originalName)
	if original == nil {
		return nil
	}

	clone := &AdvancedMaterial{
		Type:           original.Type,
		Name:           newName,
		DiffuseColor:   original.DiffuseColor,
		SpecularColor:  original.SpecularColor,
		EmissiveColor:  original.EmissiveColor,
		Shininess:      original.Shininess,
		Metallic:       original.Metallic,
		Roughness:      original.Roughness,
		Opacity:        original.Opacity,
		NormalStrength: original.NormalStrength,
		EmissiveStrength: original.EmissiveStrength,
		HeightScale:    original.HeightScale,
		Textures:       make(map[TextureSlot]uint32),
		TextureEnabled: make(map[TextureSlot]bool),
		TextureScale:   original.TextureScale,
		TextureOffset:  original.TextureOffset,
		DoubleSided:    original.DoubleSided,
		CastShadows:    original.CastShadows,
		ReceiveShadows: original.ReceiveShadows,
		Transparent:    original.Transparent,
	}

	// Copy textures
	for slot, textureID := range original.Textures {
		clone.Textures[slot] = textureID
	}
	for slot, enabled := range original.TextureEnabled {
		clone.TextureEnabled[slot] = enabled
	}

	mm.materials[newName] = clone
	return clone
}