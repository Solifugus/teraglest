package graphics

import "github.com/go-gl/mathgl/mgl32"

// ShaderInterface defines the methods needed by Model for shader operations
// This interface breaks the circular dependency between graphics and renderer packages
type ShaderInterface interface {
	// SetUniformMat4 sets a 4x4 matrix uniform
	SetUniformMat4(shaderName, uniformName string, matrix mgl32.Mat4) error

	// SetUniformVec3 sets a 3-component vector uniform
	SetUniformVec3(shaderName, uniformName string, vector mgl32.Vec3) error

	// SetUniformVec2 sets a 2-component vector uniform
	SetUniformVec2(shaderName, uniformName string, vector mgl32.Vec2) error

	// SetUniformFloat sets a float uniform
	SetUniformFloat(shaderName, uniformName string, value float32) error

	// SetUniformInt sets an integer uniform
	SetUniformInt(shaderName, uniformName string, value int32) error

	// SetUniformBool sets a boolean uniform
	SetUniformBool(shaderName, uniformName string, value bool) error
}