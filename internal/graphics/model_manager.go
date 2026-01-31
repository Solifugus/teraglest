package graphics

import (
	"fmt"
	"log"
	"strings"

	"teraglest/pkg/formats"
)

// ModelManager manages loading, caching, and rendering of 3D models
type ModelManager struct {
	modelCache   map[string]*Model         // Path -> loaded model
	textureManager *TextureManager       // Texture management
	loadedModels []*Model                 // All loaded models for batch operations
}

// NewModelManager creates a new model manager
func NewModelManager() *ModelManager {
	return &ModelManager{
		modelCache:     make(map[string]*Model),
		textureManager: NewTextureManager(),
		loadedModels:   make([]*Model, 0),
	}
}

// LoadG3DModel loads a G3D model from a file path
func (mm *ModelManager) LoadG3DModel(filePath string) (*Model, error) {
	// Check cache first
	if model, exists := mm.modelCache[filePath]; exists {
		return model, nil
	}

	// Load G3D file
	g3dModel, err := formats.LoadG3D(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load G3D file %s: %w", filePath, err)
	}

	// Convert G3D to OpenGL model
	model, err := NewModelFromG3D(g3dModel)
	if err != nil {
		return nil, fmt.Errorf("failed to convert G3D model to OpenGL: %w", err)
	}

	// Load and assign texture
	err = mm.loadModelTexture(model, g3dModel, filePath)
	if err != nil {
		log.Printf("Warning: failed to load texture for model %s: %v", filePath, err)
		// Continue without texture - model will use default white texture
	}

	// Cache the model
	mm.modelCache[filePath] = model
	mm.loadedModels = append(mm.loadedModels, model)

	log.Printf("Loaded G3D model: %s (%d vertices, %d triangles)",
		filePath, model.GetVertexCount(), model.GetTriangleCount())

	return model, nil
}

// loadModelTexture loads and assigns the appropriate texture for a G3D model
func (mm *ModelManager) loadModelTexture(model *Model, g3dModel *formats.G3DModel, modelPath string) error {
	if len(g3dModel.Meshes) == 0 {
		return fmt.Errorf("no meshes in G3D model")
	}

	// Get texture names from the first mesh
	mesh := &g3dModel.Meshes[0]
	textureNames := make([]string, 0)

	// Extract texture name from mesh header (implementation depends on G3D format)
	// For now, we'll use the mesh name to derive texture names
	meshName := strings.TrimRight(string(mesh.Header.Name[:]), "\x00")
	if meshName != "" {
		textureNames = append(textureNames, meshName)
	}

	// Try to find and load texture
	texture, err := mm.textureManager.FindTextureForModel(modelPath, textureNames)
	if err != nil {
		return fmt.Errorf("failed to find texture: %w", err)
	}

	// Assign texture to model
	model.TextureID = texture.ID

	return nil
}

// LoadModelFromMemory creates a model from an already-loaded G3D structure
func (mm *ModelManager) LoadModelFromMemory(g3dModel *formats.G3DModel, cacheKey string) (*Model, error) {
	// Check cache first
	if model, exists := mm.modelCache[cacheKey]; exists {
		return model, nil
	}

	// Convert G3D to OpenGL model
	model, err := NewModelFromG3D(g3dModel)
	if err != nil {
		return nil, fmt.Errorf("failed to convert G3D model to OpenGL: %w", err)
	}

	// Create default texture for in-memory models
	defaultTexture, err := mm.textureManager.CreateDefaultTexture()
	if err != nil {
		return nil, fmt.Errorf("failed to create default texture: %w", err)
	}
	model.TextureID = defaultTexture.ID

	// Cache the model
	mm.modelCache[cacheKey] = model
	mm.loadedModels = append(mm.loadedModels, model)

	return model, nil
}

// GetModel returns a cached model by path
func (mm *ModelManager) GetModel(filePath string) *Model {
	return mm.modelCache[filePath]
}

// GetAllModels returns all loaded models
func (mm *ModelManager) GetAllModels() []*Model {
	return mm.loadedModels
}

// GetModelCount returns the number of loaded models
func (mm *ModelManager) GetModelCount() int {
	return len(mm.loadedModels)
}

// UnloadModel removes a model from cache and GPU
func (mm *ModelManager) UnloadModel(filePath string) {
	if model, exists := mm.modelCache[filePath]; exists {
		// Remove from loaded models slice
		for i, m := range mm.loadedModels {
			if m == model {
				mm.loadedModels = append(mm.loadedModels[:i], mm.loadedModels[i+1:]...)
				break
			}
		}

		// Cleanup GPU resources
		model.Cleanup()

		// Remove from cache
		delete(mm.modelCache, filePath)

		log.Printf("Unloaded model: %s", filePath)
	}
}

// RenderAllModels renders all loaded models using the provided shader interface
func (mm *ModelManager) RenderAllModels(shaderName string, shaderInterface ShaderInterface) error {
	for _, model := range mm.loadedModels {
		err := model.Render(shaderName, shaderInterface)
		if err != nil {
			return fmt.Errorf("failed to render model %s: %w", model.Name, err)
		}
	}
	return nil
}

// RenderModel renders a specific model
func (mm *ModelManager) RenderModel(model *Model, shaderName string, shaderInterface ShaderInterface) error {
	if model == nil {
		return fmt.Errorf("model is nil")
	}
	return model.Render(shaderName, shaderInterface)
}

// CreateTestScene creates a test scene with basic geometric models
func (mm *ModelManager) CreateTestScene() error {
	// Create test cube model
	cubeModel, err := mm.createTestCubeModel()
	if err != nil {
		return fmt.Errorf("failed to create test cube: %w", err)
	}

	// Position the test models
	cubeModel.SetPosition(0, 0, 0)

	// Cache test models
	mm.modelCache["test_cube"] = cubeModel
	mm.loadedModels = append(mm.loadedModels, cubeModel)

	log.Println("Created test scene with cube model")
	return nil
}

// createTestCubeModel creates a simple cube model for testing
func (mm *ModelManager) createTestCubeModel() (*Model, error) {
	// Create a simple G3D-like structure for a cube
	// This is for testing the rendering pipeline

	// Create test G3D model structure
	g3dModel := &formats.G3DModel{
		FileHeader: formats.G3DFileHeader{
			ID: [3]byte{'G', '3', 'D'},
			Version: 4,
		},
		ModelHeader: formats.G3DModelHeader{
			MeshCount: 1,
			Type: formats.MorphMesh,
		},
		Meshes: []formats.G3DMesh{
			{
				Header: formats.G3DMeshHeader{
					Name: [64]byte{'T', 'e', 's', 't', 'C', 'u', 'b', 'e'},
					FrameCount: 1,
					VertexCount: 24, // 6 faces * 4 vertices per face
					IndexCount: 36, // 6 faces * 2 triangles * 3 indices per triangle
					DiffuseColor: [3]float32{0.8, 0.8, 0.8},  // Light gray
					SpecularColor: [3]float32{1.0, 1.0, 1.0}, // White specular
					SpecularPower: 32.0,
					Opacity: 1.0,
				},
			},
		},
	}

	// Convert to OpenGL model (uses the hardcoded cube vertices from extractG3DVertexData)
	model, err := NewModelFromG3D(g3dModel)
	if err != nil {
		return nil, fmt.Errorf("failed to create test cube model: %w", err)
	}

	// Assign default texture
	defaultTexture, err := mm.textureManager.CreateDefaultTexture()
	if err != nil {
		return nil, fmt.Errorf("failed to create default texture for test cube: %w", err)
	}
	model.TextureID = defaultTexture.ID

	return model, nil
}

// GetStatistics returns statistics about loaded models and textures
func (mm *ModelManager) GetStatistics() ModelStatistics {
	return ModelStatistics{
		ModelCount:   len(mm.modelCache),
		TextureCount: mm.textureManager.GetCacheSize(),
		TotalTriangles: mm.getTotalTriangles(),
		MemoryUsageEstimate: mm.getMemoryUsageEstimate(),
	}
}

// ModelStatistics contains information about loaded models
type ModelStatistics struct {
	ModelCount    int     // Number of loaded models
	TextureCount  int     // Number of loaded textures
	TotalTriangles int    // Total triangles across all models
	MemoryUsageEstimate int64 // Rough memory usage in bytes
}

// getTotalTriangles calculates the total number of triangles across all models
func (mm *ModelManager) getTotalTriangles() int {
	total := 0
	for _, model := range mm.loadedModels {
		total += int(model.IndexCount) / 3 // 3 indices per triangle
	}
	return total
}

// getMemoryUsageEstimate provides a rough estimate of GPU memory usage
func (mm *ModelManager) getMemoryUsageEstimate() int64 {
	var total int64 = 0

	// Estimate model memory (vertices + indices)
	for _, model := range mm.loadedModels {
		// Rough estimate: vertex data + index data + VAO/VBO overhead
		vertexMemory := int64(24) * int64(model.IndexCount) // ~24 bytes per vertex (pos+normal+texcoord)
		indexMemory := int64(4) * int64(model.IndexCount)   // 4 bytes per index
		overhead := int64(1024)                              // VAO/VBO overhead
		total += vertexMemory + indexMemory + overhead
	}

	// Add texture memory estimate (rough: assume 512x512 RGBA textures)
	textureMemory := int64(mm.textureManager.GetCacheSize()) * 512 * 512 * 4
	total += textureMemory

	return total
}

// Cleanup releases all models and GPU resources
func (mm *ModelManager) Cleanup() {
	// Cleanup all models
	for path, model := range mm.modelCache {
		model.Cleanup()
		log.Printf("Cleaned up model: %s", path)
	}

	// Cleanup texture manager
	mm.textureManager.Cleanup()

	// Clear caches
	mm.modelCache = make(map[string]*Model)
	mm.loadedModels = make([]*Model, 0)

	log.Println("ModelManager cleanup completed")
}

// LoadModelsFromDirectory loads all G3D models from a directory
func (mm *ModelManager) LoadModelsFromDirectory(dirPath string) ([]string, error) {
	// This function would scan a directory for .g3d files and load them
	// Implementation depends on whether you want recursive scanning
	loadedPaths := make([]string, 0)

	// For now, return empty slice as this requires filesystem scanning
	// which should be implemented based on specific requirements

	log.Printf("LoadModelsFromDirectory not fully implemented for: %s", dirPath)
	return loadedPaths, nil
}

// SetModelPosition sets the position of a cached model
func (mm *ModelManager) SetModelPosition(filePath string, x, y, z float32) error {
	model := mm.GetModel(filePath)
	if model == nil {
		return fmt.Errorf("model not found: %s", filePath)
	}
	model.SetPosition(x, y, z)
	return nil
}

// SetModelRotation sets the rotation of a cached model
func (mm *ModelManager) SetModelRotation(filePath string, angleX, angleY, angleZ float32) error {
	model := mm.GetModel(filePath)
	if model == nil {
		return fmt.Errorf("model not found: %s", filePath)
	}
	model.SetRotation(angleX, angleY, angleZ)
	return nil
}

// SetModelScale sets the scale of a cached model
func (mm *ModelManager) SetModelScale(filePath string, scaleX, scaleY, scaleZ float32) error {
	model := mm.GetModel(filePath)
	if model == nil {
		return fmt.Errorf("model not found: %s", filePath)
	}
	model.SetScale(scaleX, scaleY, scaleZ)
	return nil
}