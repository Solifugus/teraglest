package graphics

import (
	"fmt"
	"math"

	"github.com/go-gl/mathgl/mgl32"
)

// LightType represents different types of lights
type LightType int

const (
	DirectionalLight LightType = 0
	PointLight       LightType = 1
	SpotLight        LightType = 2
)

// Light represents a 3D light source
type Light struct {
	// Basic properties
	Type      LightType   // Type of light
	Position  mgl32.Vec3  // World position (used by point and spot lights)
	Direction mgl32.Vec3  // Light direction (used by directional and spot lights)

	// Color properties
	Color     mgl32.Vec3  // RGB color of the light
	Intensity float32     // Light intensity/brightness

	// Attenuation (for point and spot lights)
	Constant  float32     // Constant attenuation factor
	Linear    float32     // Linear attenuation factor
	Quadratic float32     // Quadratic attenuation factor

	// Spot light properties
	InnerCone float32     // Inner cone angle (radians)
	OuterCone float32     // Outer cone angle (radians)

	// State
	Enabled   bool        // Whether this light is active
	CastsShadows bool     // Whether this light casts shadows
}

// LightManager manages all lights in the scene
type LightManager struct {
	lights       []Light      // All lights in the scene
	maxLights    int          // Maximum number of lights supported
	ambientColor mgl32.Vec3   // Global ambient lighting
	ambientStrength float32   // Ambient light intensity
}

// NewLightManager creates a new light manager
func NewLightManager(maxLights int) *LightManager {
	return &LightManager{
		lights:          make([]Light, 0, maxLights),
		maxLights:       maxLights,
		ambientColor:    mgl32.Vec3{0.2, 0.2, 0.2}, // Soft gray ambient
		ambientStrength: 0.3,
	}
}

// CreateDirectionalLight creates a new directional light (like the sun)
func (lm *LightManager) CreateDirectionalLight(direction mgl32.Vec3, color mgl32.Vec3, intensity float32) (*Light, error) {
	if len(lm.lights) >= lm.maxLights {
		return nil, fmt.Errorf("maximum number of lights (%d) reached", lm.maxLights)
	}

	light := Light{
		Type:         DirectionalLight,
		Direction:    direction.Normalize(),
		Color:        color,
		Intensity:    intensity,
		Enabled:      true,
		CastsShadows: true,
	}

	lm.lights = append(lm.lights, light)
	return &lm.lights[len(lm.lights)-1], nil
}

// CreatePointLight creates a new point light (like a light bulb)
func (lm *LightManager) CreatePointLight(position mgl32.Vec3, color mgl32.Vec3, intensity float32, range_ float32) (*Light, error) {
	if len(lm.lights) >= lm.maxLights {
		return nil, fmt.Errorf("maximum number of lights (%d) reached", lm.maxLights)
	}

	// Calculate attenuation factors for the given range
	constant := 1.0
	linear := 2.0 / range_
	quadratic := 1.0 / (range_ * range_)

	light := Light{
		Type:         PointLight,
		Position:     position,
		Color:        color,
		Intensity:    intensity,
		Constant:     float32(constant),
		Linear:       float32(linear),
		Quadratic:    float32(quadratic),
		Enabled:      true,
		CastsShadows: false, // Point light shadows more complex, disabled for now
	}

	lm.lights = append(lm.lights, light)
	return &lm.lights[len(lm.lights)-1], nil
}

// CreateSpotLight creates a new spot light (like a flashlight)
func (lm *LightManager) CreateSpotLight(position, direction mgl32.Vec3, color mgl32.Vec3, intensity float32, range_ float32, innerAngle, outerAngle float32) (*Light, error) {
	if len(lm.lights) >= lm.maxLights {
		return nil, fmt.Errorf("maximum number of lights (%d) reached", lm.maxLights)
	}

	// Calculate attenuation factors
	constant := 1.0
	linear := 2.0 / range_
	quadratic := 1.0 / (range_ * range_)

	light := Light{
		Type:         SpotLight,
		Position:     position,
		Direction:    direction.Normalize(),
		Color:        color,
		Intensity:    intensity,
		Constant:     float32(constant),
		Linear:       float32(linear),
		Quadratic:    float32(quadratic),
		InnerCone:    innerAngle,
		OuterCone:    outerAngle,
		Enabled:      true,
		CastsShadows: false, // Spot light shadows complex, disabled for now
	}

	lm.lights = append(lm.lights, light)
	return &lm.lights[len(lm.lights)-1], nil
}

// RemoveLight removes a light from the manager
func (lm *LightManager) RemoveLight(index int) error {
	if index < 0 || index >= len(lm.lights) {
		return fmt.Errorf("invalid light index: %d", index)
	}

	// Remove light by swapping with last and truncating
	lm.lights[index] = lm.lights[len(lm.lights)-1]
	lm.lights = lm.lights[:len(lm.lights)-1]

	return nil
}

// GetActiveLights returns all enabled lights
func (lm *LightManager) GetActiveLights() []Light {
	activeLights := make([]Light, 0, len(lm.lights))
	for _, light := range lm.lights {
		if light.Enabled {
			activeLights = append(activeLights, light)
		}
	}
	return activeLights
}

// GetLightCount returns the number of lights
func (lm *LightManager) GetLightCount() int {
	return len(lm.lights)
}

// GetActiveLightCount returns the number of enabled lights
func (lm *LightManager) GetActiveLightCount() int {
	count := 0
	for _, light := range lm.lights {
		if light.Enabled {
			count++
		}
	}
	return count
}

// SetAmbientLight sets the global ambient lighting
func (lm *LightManager) SetAmbientLight(color mgl32.Vec3, strength float32) {
	lm.ambientColor = color
	lm.ambientStrength = strength
}

// GetAmbientLight returns the ambient lighting parameters
func (lm *LightManager) GetAmbientLight() (mgl32.Vec3, float32) {
	return lm.ambientColor, lm.ambientStrength
}

// UpdateShaderUniforms sends light data to the shader
func (lm *LightManager) UpdateShaderUniforms(shaderInterface ShaderInterface, shaderName string) error {
	activeLights := lm.GetActiveLights()
	lightCount := len(activeLights)

	// Send light count to shader
	err := shaderInterface.SetUniformInt(shaderName, "uNumLights", int32(lightCount))
	if err != nil {
		return fmt.Errorf("failed to set light count: %w", err)
	}

	// Send ambient light
	ambientFinal := lm.ambientColor.Mul(lm.ambientStrength)
	err = shaderInterface.SetUniformVec3(shaderName, "uAmbientColor", ambientFinal)
	if err != nil {
		return fmt.Errorf("failed to set ambient color: %w", err)
	}

	// Send individual light data
	for i, light := range activeLights {
		if i >= lm.maxLights {
			break // Don't exceed shader array limits
		}

		prefix := fmt.Sprintf("uLights[%d]", i)

		// Light type
		err = shaderInterface.SetUniformInt(shaderName, prefix+".type", int32(light.Type))
		if err != nil {
			return fmt.Errorf("failed to set light %d type: %w", i, err)
		}

		// Position (for point and spot lights)
		err = shaderInterface.SetUniformVec3(shaderName, prefix+".position", light.Position)
		if err != nil {
			return fmt.Errorf("failed to set light %d position: %w", i, err)
		}

		// Direction (for directional and spot lights)
		err = shaderInterface.SetUniformVec3(shaderName, prefix+".direction", light.Direction)
		if err != nil {
			return fmt.Errorf("failed to set light %d direction: %w", i, err)
		}

		// Color and intensity
		lightColor := light.Color.Mul(light.Intensity)
		err = shaderInterface.SetUniformVec3(shaderName, prefix+".color", lightColor)
		if err != nil {
			return fmt.Errorf("failed to set light %d color: %w", i, err)
		}

		// Attenuation (for point and spot lights)
		err = shaderInterface.SetUniformFloat(shaderName, prefix+".constant", light.Constant)
		if err != nil {
			return fmt.Errorf("failed to set light %d constant: %w", i, err)
		}

		err = shaderInterface.SetUniformFloat(shaderName, prefix+".linear", light.Linear)
		if err != nil {
			return fmt.Errorf("failed to set light %d linear: %w", i, err)
		}

		err = shaderInterface.SetUniformFloat(shaderName, prefix+".quadratic", light.Quadratic)
		if err != nil {
			return fmt.Errorf("failed to set light %d quadratic: %w", i, err)
		}

		// Spot light cone angles
		err = shaderInterface.SetUniformFloat(shaderName, prefix+".innerCone", float32(math.Cos(float64(light.InnerCone))))
		if err != nil {
			return fmt.Errorf("failed to set light %d inner cone: %w", i, err)
		}

		err = shaderInterface.SetUniformFloat(shaderName, prefix+".outerCone", float32(math.Cos(float64(light.OuterCone))))
		if err != nil {
			return fmt.Errorf("failed to set light %d outer cone: %w", i, err)
		}
	}

	return nil
}

// CreateDefaultLighting creates a basic lighting setup for testing
func (lm *LightManager) CreateDefaultLighting() error {
	// Clear existing lights
	lm.lights = make([]Light, 0, lm.maxLights)

	// Create main directional light (sun)
	sunDirection := mgl32.Vec3{-0.3, -0.7, -0.6}.Normalize() // From upper-left
	_, err := lm.CreateDirectionalLight(sunDirection, mgl32.Vec3{1.0, 0.95, 0.8}, 0.8)
	if err != nil {
		return fmt.Errorf("failed to create sun light: %w", err)
	}

	// Create fill light (softer directional)
	fillDirection := mgl32.Vec3{0.5, -0.2, 0.3}.Normalize()
	_, err = lm.CreateDirectionalLight(fillDirection, mgl32.Vec3{0.4, 0.5, 0.6}, 0.3)
	if err != nil {
		return fmt.Errorf("failed to create fill light: %w", err)
	}

	// Set ambient lighting
	lm.SetAmbientLight(mgl32.Vec3{0.15, 0.2, 0.3}, 0.4)

	return nil
}

// SetLightPosition updates a light's position
func (lm *LightManager) SetLightPosition(index int, position mgl32.Vec3) error {
	if index < 0 || index >= len(lm.lights) {
		return fmt.Errorf("invalid light index: %d", index)
	}

	lm.lights[index].Position = position
	return nil
}

// SetLightDirection updates a light's direction
func (lm *LightManager) SetLightDirection(index int, direction mgl32.Vec3) error {
	if index < 0 || index >= len(lm.lights) {
		return fmt.Errorf("invalid light index: %d", index)
	}

	lm.lights[index].Direction = direction.Normalize()
	return nil
}

// SetLightColor updates a light's color and intensity
func (lm *LightManager) SetLightColor(index int, color mgl32.Vec3, intensity float32) error {
	if index < 0 || index >= len(lm.lights) {
		return fmt.Errorf("invalid light index: %d", index)
	}

	lm.lights[index].Color = color
	lm.lights[index].Intensity = intensity
	return nil
}

// SetLightEnabled enables or disables a light
func (lm *LightManager) SetLightEnabled(index int, enabled bool) error {
	if index < 0 || index >= len(lm.lights) {
		return fmt.Errorf("invalid light index: %d", index)
	}

	lm.lights[index].Enabled = enabled
	return nil
}

// GetLightingInfo returns debug information about the lighting setup
func (lm *LightManager) GetLightingInfo() string {
	info := fmt.Sprintf("Lighting System:\n")
	info += fmt.Sprintf("  Ambient: RGB(%.2f, %.2f, %.2f) @ %.2f\n",
		lm.ambientColor.X(), lm.ambientColor.Y(), lm.ambientColor.Z(), lm.ambientStrength)
	info += fmt.Sprintf("  Active Lights: %d/%d\n", lm.GetActiveLightCount(), lm.maxLights)

	for i, light := range lm.lights {
		if !light.Enabled {
			continue
		}

		var lightType string
		switch light.Type {
		case DirectionalLight:
			lightType = "Directional"
		case PointLight:
			lightType = "Point"
		case SpotLight:
			lightType = "Spot"
		}

		info += fmt.Sprintf("  Light %d: %s, RGB(%.2f, %.2f, %.2f) @ %.2f\n",
			i, lightType, light.Color.X(), light.Color.Y(), light.Color.Z(), light.Intensity)
	}

	return info
}