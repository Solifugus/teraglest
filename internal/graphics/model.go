package graphics

import (
	"fmt"
	"strings"
	"unsafe"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/mathgl/mgl32"

	"teraglest/pkg/formats"
)

// Model represents a 3D model ready for rendering
type Model struct {
	// GPU Resources
	VAO        uint32  // Vertex Array Object
	VBO        uint32  // Vertex Buffer Object
	EBO        uint32  // Element Buffer Object
	IndexCount int32   // Number of indices to render
	TextureID  uint32  // Associated texture

	// Model properties
	Name        string           // Model name
	BoundingBox BoundingBox      // For frustum culling
	Transform   mgl32.Mat4       // Model transformation matrix

	// Animation support (for future)
	FrameCount  int              // Number of animation frames
	CurrentFrame int             // Current animation frame

	// Material properties from G3D
	Material        Material
	AdvancedMaterial *AdvancedMaterial  // Enhanced material system
}

// Vertex represents a single vertex with all attributes
type Vertex struct {
	Position mgl32.Vec3 // X, Y, Z coordinates
	Normal   mgl32.Vec3 // Normal vector for lighting
	TexCoord mgl32.Vec2 // Texture coordinates (U, V)
}

// Material represents material properties from G3D
type Material struct {
	DiffuseColor  mgl32.Vec3 // RGB diffuse color
	SpecularColor mgl32.Vec3 // RGB specular color
	SpecularPower float32    // Shininess/specular power
	Opacity       float32    // Alpha/opacity value
}

// BoundingBox represents an axis-aligned bounding box
type BoundingBox struct {
	Min mgl32.Vec3 // Minimum corner
	Max mgl32.Vec3 // Maximum corner
}

// NewModelFromG3D creates a Model from G3D data
func NewModelFromG3D(g3dModel *formats.G3DModel) (*Model, error) {
	if g3dModel == nil {
		return nil, fmt.Errorf("G3D model is nil")
	}

	if len(g3dModel.Meshes) == 0 {
		return nil, fmt.Errorf("G3D model has no meshes")
	}

	// For now, render only the first mesh (can be enhanced for multi-mesh models)
	mesh := &g3dModel.Meshes[0]

	// Extract vertex data from G3D mesh
	vertices, indices, err := extractG3DVertexData(mesh)
	if err != nil {
		return nil, fmt.Errorf("failed to extract G3D vertex data: %w", err)
	}

	// Convert mesh name from [64]byte to string
	meshName := strings.TrimRight(string(mesh.Header.Name[:]), "\x00")

	// Create OpenGL buffers
	model := &Model{
		Name:         meshName,
		IndexCount:   int32(len(indices)),
		FrameCount:   int(mesh.Header.FrameCount),
		CurrentFrame: 0,
		Transform:    mgl32.Ident4(),
		Material:     extractMaterial(mesh),
	}

	// Generate OpenGL objects
	gl.GenVertexArrays(1, &model.VAO)
	gl.GenBuffers(1, &model.VBO)
	gl.GenBuffers(1, &model.EBO)

	// Bind VAO
	gl.BindVertexArray(model.VAO)

	// Upload vertex data
	gl.BindBuffer(gl.ARRAY_BUFFER, model.VBO)
	vertexDataSize := len(vertices) * int(unsafe.Sizeof(Vertex{}))
	gl.BufferData(gl.ARRAY_BUFFER, vertexDataSize, gl.Ptr(vertices), gl.STATIC_DRAW)

	// Upload index data
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, model.EBO)
	indexDataSize := len(indices) * int(unsafe.Sizeof(uint32(0)))
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, indexDataSize, gl.Ptr(indices), gl.STATIC_DRAW)

	// Configure vertex attributes
	setupVertexAttributes()

	// Calculate bounding box
	model.BoundingBox = calculateBoundingBox(vertices)

	// Unbind VAO
	gl.BindVertexArray(0)

	return model, nil
}

// extractG3DVertexData extracts vertex and index data from a G3D mesh
func extractG3DVertexData(mesh *formats.G3DMesh) ([]Vertex, []uint32, error) {
	// NOW USING REAL G3D VERTEX DATA! Complete parser provides actual geometry.

	// Validate that we have vertex data
	if len(mesh.Vertices) == 0 {
		return nil, nil, fmt.Errorf("mesh has no vertex data: %s", mesh.Name)
	}

	if len(mesh.Indices) == 0 {
		return nil, nil, fmt.Errorf("mesh has no index data: %s", mesh.Name)
	}

	// For now, render only the first frame (frame 0) for non-animated models
	// TODO: Add animation support later
	frameToRender := 0
	vertexCount := int(mesh.Header.VertexCount)

	if frameToRender >= int(mesh.Header.FrameCount) {
		frameToRender = 0 // Fallback to first frame
	}

	// Calculate offset for the frame we want to render
	frameOffset := frameToRender * vertexCount

	// Extract vertices for the current frame
	vertices := make([]Vertex, vertexCount)

	for i := 0; i < vertexCount; i++ {
		vertexIndex := frameOffset + i

		// Validate array bounds
		if vertexIndex >= len(mesh.Vertices) || vertexIndex >= len(mesh.Normals) {
			return nil, nil, fmt.Errorf("vertex index out of bounds: %d >= %d or %d",
				vertexIndex, len(mesh.Vertices), len(mesh.Normals))
		}

		// Convert from G3D Vec3f to mathgl Vec3
		g3dVertex := mesh.Vertices[vertexIndex]
		g3dNormal := mesh.Normals[vertexIndex]

		vertices[i].Position = mgl32.Vec3{g3dVertex.X, g3dVertex.Y, g3dVertex.Z}
		vertices[i].Normal = mgl32.Vec3{g3dNormal.X, g3dNormal.Y, g3dNormal.Z}

		// Texture coordinates (not frame-dependent)
		if i < len(mesh.TexCoords) {
			g3dTexCoord := mesh.TexCoords[i]
			vertices[i].TexCoord = mgl32.Vec2{g3dTexCoord.X, g3dTexCoord.Y}
		} else {
			// Default texture coordinates if not available
			vertices[i].TexCoord = mgl32.Vec2{0.0, 0.0}
		}
	}

	// Copy indices directly - they reference the vertices in the current frame
	indices := make([]uint32, len(mesh.Indices))
	copy(indices, mesh.Indices)

	return vertices, indices, nil
}

// setupVertexAttributes configures OpenGL vertex attribute pointers
func setupVertexAttributes() {
	// Position attribute (location 0)
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, int32(unsafe.Sizeof(Vertex{})), gl.PtrOffset(0))

	// Normal attribute (location 1)
	gl.EnableVertexAttribArray(1)
	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, int32(unsafe.Sizeof(Vertex{})),
		gl.PtrOffset(int(unsafe.Offsetof(Vertex{}.Normal))))

	// Texture coordinate attribute (location 2)
	gl.EnableVertexAttribArray(2)
	gl.VertexAttribPointer(2, 2, gl.FLOAT, false, int32(unsafe.Sizeof(Vertex{})),
		gl.PtrOffset(int(unsafe.Offsetof(Vertex{}.TexCoord))))
}

// extractMaterial extracts material properties from G3D mesh
func extractMaterial(mesh *formats.G3DMesh) Material {
	header := mesh.Header

	return Material{
		DiffuseColor:  mgl32.Vec3{header.DiffuseColor[0], header.DiffuseColor[1], header.DiffuseColor[2]},
		SpecularColor: mgl32.Vec3{header.SpecularColor[0], header.SpecularColor[1], header.SpecularColor[2]},
		SpecularPower: header.SpecularPower,
		Opacity:       header.Opacity,
	}
}

// calculateBoundingBox calculates the axis-aligned bounding box for vertices
func calculateBoundingBox(vertices []Vertex) BoundingBox {
	if len(vertices) == 0 {
		return BoundingBox{}
	}

	min := vertices[0].Position
	max := vertices[0].Position

	for _, vertex := range vertices[1:] {
		pos := vertex.Position

		// Update min
		if pos.X() < min.X() { min[0] = pos.X() }
		if pos.Y() < min.Y() { min[1] = pos.Y() }
		if pos.Z() < min.Z() { min[2] = pos.Z() }

		// Update max
		if pos.X() > max.X() { max[0] = pos.X() }
		if pos.Y() > max.Y() { max[1] = pos.Y() }
		if pos.Z() > max.Z() { max[2] = pos.Z() }
	}

	return BoundingBox{Min: min, Max: max}
}

// SetPosition sets the model's position
func (m *Model) SetPosition(x, y, z float32) {
	m.Transform = mgl32.Translate3D(x, y, z)
}

// SetRotation sets the model's rotation (in radians)
func (m *Model) SetRotation(angleX, angleY, angleZ float32) {
	rotX := mgl32.HomogRotate3DX(angleX)
	rotY := mgl32.HomogRotate3DY(angleY)
	rotZ := mgl32.HomogRotate3DZ(angleZ)
	m.Transform = rotZ.Mul4(rotY).Mul4(rotX)
}

// SetScale sets the model's scale
func (m *Model) SetScale(scaleX, scaleY, scaleZ float32) {
	scale := mgl32.Scale3D(scaleX, scaleY, scaleZ)
	m.Transform = m.Transform.Mul4(scale)
}

// GetModelMatrix returns the complete model transformation matrix
func (m *Model) GetModelMatrix() mgl32.Mat4 {
	return m.Transform
}

// GetNormalMatrix returns the normal transformation matrix for lighting
func (m *Model) GetNormalMatrix() mgl32.Mat3 {
	// Normal matrix is the transpose of the inverse of the model matrix (3x3 portion)
	modelMat3 := m.Transform.Mat3()
	return modelMat3.Inv().Transpose()
}

// GetVertexCount returns the number of vertices in the model
func (m *Model) GetVertexCount() int {
	return int(m.IndexCount) // For indexed rendering, this is the index count
}

// GetTriangleCount returns the number of triangles in the model
func (m *Model) GetTriangleCount() int {
	return int(m.IndexCount) / 3 // 3 indices per triangle
}

// IsVisible checks if the model is within the camera frustum (basic implementation)
func (m *Model) IsVisible(viewMatrix, projMatrix mgl32.Mat4) bool {
	// Transform bounding box center to view space
	center := m.BoundingBox.Min.Add(m.BoundingBox.Max).Mul(0.5)
	worldCenter := m.Transform.Mul4x1(mgl32.Vec4{center.X(), center.Y(), center.Z(), 1.0})

	// Simple visibility check - can be enhanced with proper frustum culling
	viewPos := viewMatrix.Mul4x1(worldCenter)

	// Basic near/far plane check
	return viewPos.Z() < 0 && viewPos.Z() > -1000.0 // Rough visibility test
}

// SetAdvancedMaterial assigns an advanced material to this model
func (m *Model) SetAdvancedMaterial(material *AdvancedMaterial) {
	m.AdvancedMaterial = material

	// Update texture ID if material has diffuse texture
	if material.HasTexture(DiffuseTexture) {
		m.TextureID = material.GetTexture(DiffuseTexture)
	}
}

// GetEffectiveMaterial returns the advanced material or creates one from basic material
func (m *Model) GetEffectiveMaterial() *AdvancedMaterial {
	if m.AdvancedMaterial != nil {
		return m.AdvancedMaterial
	}

	// Create advanced material from basic material properties
	advMaterial := &AdvancedMaterial{
		Type:           BasicMaterial,
		Name:           m.Name + "_material",
		DiffuseColor:   m.Material.DiffuseColor,
		SpecularColor:  m.Material.SpecularColor,
		EmissiveColor:  mgl32.Vec3{0, 0, 0},
		Shininess:      m.Material.SpecularPower,
		Metallic:       0.0,
		Roughness:      1.0 - (m.Material.SpecularPower / 128.0),
		Opacity:        m.Material.Opacity,
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
		Transparent:    m.Material.Opacity < 1.0,
	}

	// Set texture if available
	if m.TextureID != 0 {
		advMaterial.Type = TexturedMaterial
		advMaterial.SetTexture(DiffuseTexture, m.TextureID)
	}

	return advMaterial
}

// Cleanup releases OpenGL resources
func (m *Model) Cleanup() {
	if m.VAO != 0 {
		gl.DeleteVertexArrays(1, &m.VAO)
		m.VAO = 0
	}
	if m.VBO != 0 {
		gl.DeleteBuffers(1, &m.VBO)
		m.VBO = 0
	}
	if m.EBO != 0 {
		gl.DeleteBuffers(1, &m.EBO)
		m.EBO = 0
	}
}

// Render renders the model using the provided shader
func (m *Model) Render(shaderName string, shaderInterface ShaderInterface) error {
	if m.VAO == 0 {
		return fmt.Errorf("model VAO not initialized")
	}

	// Set uniforms for this model
	err := shaderInterface.SetUniformMat4(shaderName, "uModel", m.GetModelMatrix())
	if err != nil {
		return fmt.Errorf("failed to set model matrix: %w", err)
	}

	normalMatrix := m.GetNormalMatrix()
	// Convert Mat3 to Mat4 for uniform (shader expects mat4)
	normalMatrix4 := mgl32.Mat4{
		normalMatrix[0], normalMatrix[1], normalMatrix[2], 0,
		normalMatrix[3], normalMatrix[4], normalMatrix[5], 0,
		normalMatrix[6], normalMatrix[7], normalMatrix[8], 0,
		0, 0, 0, 1,
	}

	err = shaderInterface.SetUniformMat4(shaderName, "uNormalMatrix", normalMatrix4)
	if err != nil {
		return fmt.Errorf("failed to set normal matrix: %w", err)
	}

	// Set material properties (try advanced shader format first, then fallback to basic)
	err = shaderInterface.SetUniformVec3(shaderName, "material.diffuse", m.Material.DiffuseColor)
	if err != nil {
		// Fallback to basic shader material uniforms
		err = shaderInterface.SetUniformVec3(shaderName, "uDiffuseColor", m.Material.DiffuseColor)
		if err != nil {
			return fmt.Errorf("failed to set material diffuse: %w", err)
		}

		err = shaderInterface.SetUniformVec3(shaderName, "uSpecularColor", m.Material.SpecularColor)
		if err != nil {
			return fmt.Errorf("failed to set material specular: %w", err)
		}

		err = shaderInterface.SetUniformFloat(shaderName, "uSpecularPower", m.Material.SpecularPower)
		if err != nil {
			return fmt.Errorf("failed to set material shininess: %w", err)
		}

		err = shaderInterface.SetUniformFloat(shaderName, "uOpacity", m.Material.Opacity)
		if err != nil {
			return fmt.Errorf("failed to set material opacity: %w", err)
		}
	} else {
		// Advanced shader material format
		err = shaderInterface.SetUniformVec3(shaderName, "material.specular", m.Material.SpecularColor)
		if err != nil {
			return fmt.Errorf("failed to set material specular: %w", err)
		}

		err = shaderInterface.SetUniformFloat(shaderName, "material.shininess", m.Material.SpecularPower)
		if err != nil {
			return fmt.Errorf("failed to set material shininess: %w", err)
		}

		err = shaderInterface.SetUniformFloat(shaderName, "material.opacity", m.Material.Opacity)
		if err != nil {
			return fmt.Errorf("failed to set material opacity: %w", err)
		}
	}

	// Bind texture if available
	if m.TextureID != 0 {
		gl.ActiveTexture(gl.TEXTURE0)
		gl.BindTexture(gl.TEXTURE_2D, m.TextureID)

		// Try advanced shader texture uniform first
		err = shaderInterface.SetUniformInt(shaderName, "uDiffuseTexture", 0)
		if err != nil {
			// Fallback to basic shader texture uniform
			err = shaderInterface.SetUniformInt(shaderName, "texture1", 0)
			if err != nil {
				return fmt.Errorf("failed to set texture uniform: %w", err)
			}
		}

		// Set texture usage flag
		err = shaderInterface.SetUniformBool(shaderName, "uUseTexture", true)
		if err != nil {
			// Ignore error if uniform doesn't exist (basic shader might not have it)
		}
	} else {
		// No texture available, set flag to false
		err = shaderInterface.SetUniformBool(shaderName, "uUseTexture", false)
		if err != nil {
			// Ignore error if uniform doesn't exist
		}
	}

	// Render the model
	gl.BindVertexArray(m.VAO)
	gl.DrawElements(gl.TRIANGLES, m.IndexCount, gl.UNSIGNED_INT, gl.PtrOffset(0))
	gl.BindVertexArray(0)

	return nil
}

// RenderWithAdvancedMaterial renders the model using the advanced material system
func (m *Model) RenderWithAdvancedMaterial(materialManager *MaterialManager, shaderInterface ShaderInterface) error {
	if m.VAO == 0 {
		return fmt.Errorf("model VAO not initialized")
	}

	// Get effective material
	material := m.GetEffectiveMaterial()

	// Get appropriate shader for this material
	shaderName := materialManager.GetShaderForMaterial(material)

	// Set uniforms for this model
	err := shaderInterface.SetUniformMat4(shaderName, "uModel", m.GetModelMatrix())
	if err != nil {
		return fmt.Errorf("failed to set model matrix: %w", err)
	}

	normalMatrix := m.GetNormalMatrix()
	normalMatrix4 := mgl32.Mat4{
		normalMatrix[0], normalMatrix[1], normalMatrix[2], 0,
		normalMatrix[3], normalMatrix[4], normalMatrix[5], 0,
		normalMatrix[6], normalMatrix[7], normalMatrix[8], 0,
		0, 0, 0, 1,
	}

	err = shaderInterface.SetUniformMat4(shaderName, "uNormalMatrix", normalMatrix4)
	if err != nil {
		return fmt.Errorf("failed to set normal matrix: %w", err)
	}

	// Apply advanced material properties
	err = material.ApplyMaterial(shaderInterface, shaderName)
	if err != nil {
		return fmt.Errorf("failed to apply material: %w", err)
	}

	// Render the model
	gl.BindVertexArray(m.VAO)
	gl.DrawElements(gl.TRIANGLES, m.IndexCount, gl.UNSIGNED_INT, gl.PtrOffset(0))
	gl.BindVertexArray(0)

	return nil
}