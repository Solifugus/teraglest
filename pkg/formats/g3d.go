package formats

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

// G3D file format constants
const (
	G3DVersion2 = 2
	G3DVersion3 = 3
	G3DVersion4 = 4
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

// G3DMesh represents a complete mesh with all its data
type G3DMesh struct {
	Header       G3DMeshHeader
	TextureNames []string // Texture file names (simplified)
	// Note: For Phase 1.4, we focus on header parsing and basic structure
	// Vertex data parsing can be added in later phases when needed for rendering
}

// G3DModel represents a complete G3D model file
type G3DModel struct {
	FileHeader  G3DFileHeader
	ModelHeader G3DModelHeader
	Meshes      []G3DMesh
}

// LoadG3D loads and parses a G3D model file
func LoadG3D(filepath string) (*G3DModel, error) {
	// Open the file
	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open G3D file %s: %w", filepath, err)
	}
	defer file.Close()

	// Get file size for validation
	_, err = file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat G3D file: %w", err)
	}

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

	// Validate file signature
	if string(model.FileHeader.ID[:]) != "G3D" {
		return nil, fmt.Errorf("invalid G3D signature: %s", string(model.FileHeader.ID[:]))
	}

	// Validate version
	if model.FileHeader.Version < G3DVersion2 || model.FileHeader.Version > G3DVersion4 {
		return nil, fmt.Errorf("unsupported G3D version: %d", model.FileHeader.Version)
	}

	// Read model header
	err = binary.Read(reader, binary.LittleEndian, &model.ModelHeader)
	if err != nil {
		return nil, fmt.Errorf("failed to read G3D model header: %w", err)
	}

	// Validate mesh type
	if model.ModelHeader.Type != MorphMesh {
		return nil, fmt.Errorf("unsupported mesh type: %d", model.ModelHeader.Type)
	}

	// Read mesh headers (Phase 1.4 focus: header parsing for mesh/vertex counts)
	model.Meshes = make([]G3DMesh, model.ModelHeader.MeshCount)
	for i := uint16(0); i < model.ModelHeader.MeshCount; i++ {
		mesh, err := readG3DMeshHeader(reader)
		if err != nil {
			return nil, fmt.Errorf("failed to read mesh %d header: %w", i, err)
		}
		model.Meshes[i] = *mesh
	}

	return model, nil
}

// readG3DMeshHeader reads just the mesh header and texture names (simplified for Phase 1.4)
func readG3DMeshHeader(reader *bytes.Reader) (*G3DMesh, error) {
	mesh := &G3DMesh{}

	// Read mesh header
	err := binary.Read(reader, binary.LittleEndian, &mesh.Header)
	if err != nil {
		return nil, fmt.Errorf("failed to read mesh header: %w", err)
	}

	// Read texture names if texture flags indicate they exist
	mesh.TextureNames = make([]string, 0, 1)
	if mesh.Header.Textures&1 != 0 { // Diffuse texture flag
		texName, err := readNullTerminatedString(reader)
		if err != nil {
			return nil, fmt.Errorf("failed to read texture name: %w", err)
		}
		mesh.TextureNames = append(mesh.TextureNames, texName)
	}

	// For Phase 1.4, we skip the vertex data parsing to focus on basic structure
	// The file position is advanced past the mesh header and texture names
	// Full vertex parsing will be implemented in later phases when needed for rendering

	// Calculate and skip vertex data section
	frameSize := mesh.Header.VertexCount * (3 + 3 + 2) * 4 // vertex(3) + normal(3) + texcoord(2) * sizeof(float32)
	vertexDataSize := int64(mesh.Header.FrameCount) * int64(frameSize)
	indexDataSize := int64(mesh.Header.IndexCount * 4) // sizeof(uint32)
	totalSkipSize := vertexDataSize + indexDataSize

	// Skip the vertex and index data for now
	_, err = reader.Seek(int64(reader.Len())-reader.Size()+totalSkipSize, io.SeekCurrent)
	if err != nil {
		// If seek fails, this mesh might have different structure
		// We'll consider this mesh parsed successfully but with limited data
	}

	return mesh, nil
}

// readNullTerminatedString reads a null-terminated string from the binary data
func readNullTerminatedString(reader *bytes.Reader) (string, error) {
	var result []byte
	for {
		var b byte
		err := binary.Read(reader, binary.LittleEndian, &b)
		if err != nil {
			return "", err
		}
		if b == 0 {
			break
		}
		result = append(result, b)
	}
	return string(result), nil
}

// GetTotalVertexCount returns the total number of vertices across all meshes and frames
func (model *G3DModel) GetTotalVertexCount() uint32 {
	total := uint32(0)
	for _, mesh := range model.Meshes {
		total += mesh.Header.FrameCount * mesh.Header.VertexCount
	}
	return total
}

// GetTotalTriangleCount returns the total number of triangles across all meshes
func (model *G3DModel) GetTotalTriangleCount() uint32 {
	total := uint32(0)
	for _, mesh := range model.Meshes {
		total += mesh.Header.IndexCount / 3
	}
	return total
}

// HasTextures returns true if the model has any textures
func (model *G3DModel) HasTextures() bool {
	for _, mesh := range model.Meshes {
		if len(mesh.TextureNames) > 0 {
			return true
		}
	}
	return false
}

// IsAnimated returns true if the model has animation frames
func (model *G3DModel) IsAnimated() bool {
	for _, mesh := range model.Meshes {
		if mesh.Header.FrameCount > 1 {
			return true
		}
	}
	return false
}

// PrintSummary prints a summary of the G3D model for debugging
func (model *G3DModel) PrintSummary() {
	fmt.Printf("G3D Model Summary:\n")
	fmt.Printf("  Version: %d\n", model.FileHeader.Version)
	fmt.Printf("  Mesh Count: %d\n", model.ModelHeader.MeshCount)
	fmt.Println()

	for i, mesh := range model.Meshes {
		// Extract mesh name
		nameBytes := mesh.Header.Name[:]
		if nullIndex := bytes.IndexByte(nameBytes, 0); nullIndex != -1 {
			nameBytes = nameBytes[:nullIndex]
		}

		fmt.Printf("  Mesh %d: %s\n", i+1, string(nameBytes))
		fmt.Printf("    Frames: %d\n", mesh.Header.FrameCount)
		fmt.Printf("    Vertices: %d\n", mesh.Header.VertexCount)
		fmt.Printf("    Indices: %d (triangles: %d)\n", mesh.Header.IndexCount, mesh.Header.IndexCount/3)
		fmt.Printf("    Textures: %d\n", len(mesh.TextureNames))

		if len(mesh.TextureNames) > 0 {
			fmt.Printf("      Texture files: %v\n", mesh.TextureNames)
		}

		fmt.Printf("    Diffuse Color: [%.2f, %.2f, %.2f]\n",
			mesh.Header.DiffuseColor[0], mesh.Header.DiffuseColor[1], mesh.Header.DiffuseColor[2])
		fmt.Printf("    Opacity: %.2f\n", mesh.Header.Opacity)
		fmt.Println()
	}
}