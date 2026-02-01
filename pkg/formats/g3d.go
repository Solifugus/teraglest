package formats

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

// Complete G3D file format implementation with vertex data parsing
// Based on MegaGlest C++ source code analysis

// G3D file format constants
const (
	G3DVersion2   = 2
	G3DVersion3   = 3
	G3DVersion4   = 4
	MapPathSize   = 64
	MeshNameSize  = 64
)

// G3DMeshType represents the mesh type
type G3DMeshType uint8

const (
	MorphMesh G3DMeshType = 0
)

// Vec3f represents a 3D vector with float32 components
type Vec3f struct {
	X, Y, Z float32
}

// Vec2f represents a 2D vector with float32 components
type Vec2f struct {
	X, Y float32
}

// G3DFileHeader represents the file header structure
type G3DFileHeader struct {
	ID      [3]byte // Should be "G3D"
	Version uint8   // Version number (2, 3, or 4)
}

// G3DModelHeader represents the model header structure
type G3DModelHeader struct {
	MeshCount uint16      // Number of meshes in this model
	Type      G3DMeshType // Mesh type (always MorphMesh = 0)
}

// G3DMeshHeader represents the mesh header structure
type G3DMeshHeader struct {
	Name           [64]byte   // Mesh name (null-terminated string)
	FrameCount     uint32     // Number of animation frames
	VertexCount    uint32     // Number of vertices per frame
	IndexCount     uint32     // Number of triangle indices
	DiffuseColor   [3]float32 // RGB diffuse color
	SpecularColor  [3]float32 // RGB specular color
	SpecularPower  float32    // Specular power/shininess
	Opacity        float32    // Opacity/alpha value
	Properties     uint32     // Property flags
	Textures       uint32     // Texture type flags
}

// G3DMesh represents a complete mesh with ALL its data INCLUDING vertex data
type G3DMesh struct {
	Header       G3DMeshHeader
	TextureNames []string // Texture file names

	// ACTUAL VERTEX DATA - This was missing in the original!
	Vertices   []Vec3f  // frameCount * vertexCount vertices
	Normals    []Vec3f  // frameCount * vertexCount normals
	TexCoords  []Vec2f  // vertexCount texture coordinates
	Indices    []uint32 // indexCount triangle indices

	// Derived properties
	Name         string  // Name as string
	TwoSided     bool    // Two-sided mesh flag
	CustomColor  bool    // Custom color flag
	NoSelect     bool    // No selection flag
	Glow         bool    // Glow effect flag
}

// G3DModel represents a complete G3D model file with full vertex data
type G3DModel struct {
	FileHeader  G3DFileHeader
	ModelHeader G3DModelHeader
	Meshes      []G3DMesh
}

// LoadG3D loads and parses a G3D model file with COMPLETE vertex data parsing
func LoadG3D(filepath string) (*G3DModel, error) {
	// Open the file
	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open G3D file %s: %w", filepath, err)
	}
	defer file.Close()

	// Read entire file into memory for binary parsing
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read G3D file: %w", err)
	}

	if len(data) < 7 { // Minimum size for headers
		return nil, fmt.Errorf("G3D file too small: %d bytes", len(data))
	}

	reader := bytes.NewReader(data)
	model := &G3DModel{}

	// Read file header
	err = binary.Read(reader, binary.LittleEndian, &model.FileHeader)
	if err != nil {
		return nil, fmt.Errorf("failed to read G3D file header: %w", err)
	}

	// Validate file header
	if string(model.FileHeader.ID[:]) != "G3D" {
		return nil, fmt.Errorf("invalid G3D file: expected 'G3D', got '%s'", string(model.FileHeader.ID[:]))
	}

	// Check version support
	if model.FileHeader.Version < 2 || model.FileHeader.Version > 4 {
		return nil, fmt.Errorf("unsupported G3D version: %d", model.FileHeader.Version)
	}

	// Read model header
	err = binary.Read(reader, binary.LittleEndian, &model.ModelHeader)
	if err != nil {
		return nil, fmt.Errorf("failed to read G3D model header: %w", err)
	}

	// Initialize meshes array
	model.Meshes = make([]G3DMesh, model.ModelHeader.MeshCount)

	// Read each mesh with COMPLETE data including vertex arrays
	for i := 0; i < int(model.ModelHeader.MeshCount); i++ {
		mesh, err := readG3DMeshComplete(reader, model.FileHeader.Version)
		if err != nil {
			return nil, fmt.Errorf("failed to read mesh %d: %w", i, err)
		}
		model.Meshes[i] = *mesh
	}

	return model, nil
}

// readG3DMeshComplete reads a complete mesh including ALL vertex data
func readG3DMeshComplete(reader *bytes.Reader, version uint8) (*G3DMesh, error) {
	mesh := &G3DMesh{}

	// Read mesh header
	err := binary.Read(reader, binary.LittleEndian, &mesh.Header)
	if err != nil {
		return nil, fmt.Errorf("failed to read mesh header: %w", err)
	}

	// Parse name from header
	mesh.Name = string(bytes.TrimRight(mesh.Header.Name[:], "\x00"))

	// Parse property flags based on MegaGlest implementation
	mesh.TwoSided = (mesh.Header.Properties & (1 << 0)) != 0    // mpfTwoSided
	mesh.CustomColor = (mesh.Header.Properties & (1 << 1)) != 0 // mpfCustomColor
	mesh.NoSelect = (mesh.Header.Properties & (1 << 2)) != 0    // mpfNoSelect
	mesh.Glow = (mesh.Header.Properties & (1 << 3)) != 0        // mpfGlow

	// Read texture names if textures are present
	textureFlag := uint32(1)
	maxTextures := 8 // meshTextureCount from MegaGlest source
	for i := 0; i < maxTextures; i++ {
		if mesh.Header.Textures & textureFlag != 0 {
			// Read texture path
			texturePath := make([]byte, MapPathSize)
			_, err := reader.Read(texturePath)
			if err != nil {
				return nil, fmt.Errorf("failed to read texture path %d: %w", i, err)
			}

			// Convert to string and store
			textureStr := string(bytes.TrimRight(texturePath, "\x00"))
			mesh.TextureNames = append(mesh.TextureNames, textureStr)
		}
		textureFlag <<= 1
	}

	// NOW READ THE ACTUAL VERTEX DATA - This was the critical missing piece!

	frameCount := mesh.Header.FrameCount
	vertexCount := mesh.Header.VertexCount
	indexCount := mesh.Header.IndexCount

	// Read vertices (frameCount * vertexCount) - positions for all animation frames
	totalVertices := frameCount * vertexCount
	if totalVertices > 0 {
		mesh.Vertices = make([]Vec3f, totalVertices)
		err = binary.Read(reader, binary.LittleEndian, mesh.Vertices)
		if err != nil {
			return nil, fmt.Errorf("failed to read vertices: %w", err)
		}
	}

	// Read normals (frameCount * vertexCount) - normals for all animation frames
	if totalVertices > 0 {
		mesh.Normals = make([]Vec3f, totalVertices)
		err = binary.Read(reader, binary.LittleEndian, mesh.Normals)
		if err != nil {
			return nil, fmt.Errorf("failed to read normals: %w", err)
		}
	}

	// Read texture coordinates (vertexCount) - only if textures are present
	if mesh.Header.Textures != 0 && vertexCount > 0 {
		mesh.TexCoords = make([]Vec2f, vertexCount)
		err = binary.Read(reader, binary.LittleEndian, mesh.TexCoords)
		if err != nil {
			return nil, fmt.Errorf("failed to read texture coordinates: %w", err)
		}
	}

	// Read indices (indexCount) - triangle indices for rendering
	if indexCount > 0 {
		mesh.Indices = make([]uint32, indexCount)
		err = binary.Read(reader, binary.LittleEndian, mesh.Indices)
		if err != nil {
			return nil, fmt.Errorf("failed to read indices: %w", err)
		}
	}

	return mesh, nil
}

// Helper methods for the complete model

func (model *G3DModel) GetTotalVertexCount() uint32 {
	total := uint32(0)
	for _, mesh := range model.Meshes {
		total += mesh.Header.VertexCount * mesh.Header.FrameCount
	}
	return total
}

func (model *G3DModel) GetTotalTriangleCount() uint32 {
	total := uint32(0)
	for _, mesh := range model.Meshes {
		total += mesh.Header.IndexCount / 3
	}
	return total
}

func (model *G3DModel) HasTextures() bool {
	for _, mesh := range model.Meshes {
		if mesh.Header.Textures != 0 {
			return true
		}
	}
	return false
}

func (model *G3DModel) IsAnimated() bool {
	for _, mesh := range model.Meshes {
		if mesh.Header.FrameCount > 1 {
			return true
		}
	}
	return false
}

// PrintSummary prints a summary of the complete model
func (model *G3DModel) PrintSummary() {
	fmt.Printf("G3D Model Summary:\n")
	fmt.Printf("  Version: %d\n", model.FileHeader.Version)
	fmt.Printf("  Meshes: %d\n", model.ModelHeader.MeshCount)
	fmt.Printf("  Total Vertices: %d\n", model.GetTotalVertexCount())
	fmt.Printf("  Total Triangles: %d\n", model.GetTotalTriangleCount())
	fmt.Printf("  Has Textures: %v\n", model.HasTextures())
	fmt.Printf("  Is Animated: %v\n", model.IsAnimated())

	for i, mesh := range model.Meshes {
		fmt.Printf("  Mesh %d: %s\n", i, mesh.Name)
		fmt.Printf("    Frames: %d\n", mesh.Header.FrameCount)
		fmt.Printf("    Vertices: %d (loaded: %d)\n", mesh.Header.VertexCount, len(mesh.Vertices))
		fmt.Printf("    Indices: %d (loaded: %d)\n", mesh.Header.IndexCount, len(mesh.Indices))
		fmt.Printf("    Textures: %d files\n", len(mesh.TextureNames))
		for j, texName := range mesh.TextureNames {
			fmt.Printf("      Texture %d: %s\n", j, texName)
		}
	}
}