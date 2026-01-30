package formats

import (
	"testing"
)

func TestLoadG3D(t *testing.T) {
	// Test loading the initiate standing model
	modelPath := "../../megaglest-source/data/glest_game/techs/megapack/factions/magic/units/initiate/models/initiate_standing.g3d"

	model, err := LoadG3D(modelPath)
	if err != nil {
		t.Fatalf("Failed to load G3D model: %v", err)
	}

	// Verify file header
	if string(model.FileHeader.ID[:]) != "G3D" {
		t.Errorf("Expected file ID 'G3D', got '%s'", string(model.FileHeader.ID[:]))
	}

	if model.FileHeader.Version != G3DVersion4 {
		t.Errorf("Expected version 4, got %d", model.FileHeader.Version)
	}

	// Verify model header
	if model.ModelHeader.MeshCount == 0 {
		t.Error("Expected at least one mesh")
	}

	if model.ModelHeader.Type != MorphMesh {
		t.Errorf("Expected mesh type MorphMesh (%d), got %d", MorphMesh, model.ModelHeader.Type)
	}

	// Verify meshes have valid header data (Phase 1.4 focus)
	for i, mesh := range model.Meshes {
		if mesh.Header.FrameCount == 0 {
			t.Errorf("Mesh %d has no animation frames", i)
		}

		if mesh.Header.VertexCount == 0 {
			t.Errorf("Mesh %d has no vertices", i)
		}

		if mesh.Header.IndexCount == 0 {
			t.Errorf("Mesh %d has no indices", i)
		}

		if mesh.Header.IndexCount%3 != 0 {
			t.Errorf("Mesh %d index count %d is not divisible by 3 (not triangles)", i, mesh.Header.IndexCount)
		}
	}

	// Test utility functions
	totalVertices := model.GetTotalVertexCount()
	if totalVertices == 0 {
		t.Error("Expected non-zero total vertex count")
	}

	totalTriangles := model.GetTotalTriangleCount()
	if totalTriangles == 0 {
		t.Error("Expected non-zero total triangle count")
	}

	// The initiate model should have textures
	if !model.HasTextures() {
		t.Error("Expected initiate model to have textures")
	}

	// Check specific values for initiate model
	if model.ModelHeader.MeshCount != 2 {
		t.Errorf("Expected initiate model to have 2 meshes, got %d", model.ModelHeader.MeshCount)
	}

	// Verify first mesh has reasonable values
	if len(model.Meshes) > 0 {
		firstMesh := model.Meshes[0]
		if firstMesh.Header.FrameCount == 0 {
			t.Error("Expected first mesh to have frames")
		}
		if firstMesh.Header.VertexCount == 0 {
			t.Error("Expected first mesh to have vertices")
		}
		if firstMesh.Header.IndexCount == 0 {
			t.Error("Expected first mesh to have indices")
		}
	}
}

func TestLoadG3DAnimated(t *testing.T) {
	// Test loading an animated model
	modelPath := "../../megaglest-source/data/glest_game/techs/megapack/factions/magic/units/initiate/models/initiate_walking.g3d"

	model, err := LoadG3D(modelPath)
	if err != nil {
		t.Fatalf("Failed to load animated G3D model: %v", err)
	}

	// Verify it's actually animated (should have multiple frames)
	if !model.IsAnimated() {
		t.Error("IsAnimated() should return true for walking model")
	}

	// Check that at least one mesh has multiple frames
	foundAnimatedMesh := false
	for _, mesh := range model.Meshes {
		if mesh.Header.FrameCount > 1 {
			foundAnimatedMesh = true
			break
		}
	}

	if !foundAnimatedMesh {
		t.Error("Expected walking model to have at least one mesh with multiple frames")
	}
}

func TestLoadG3DInvalid(t *testing.T) {
	// Test loading a non-existent file
	_, err := LoadG3D("nonexistent.g3d")
	if err == nil {
		t.Error("Expected error when loading non-existent file")
	}

	// Test loading a non-G3D file (use a text file)
	_, err = LoadG3D("../../megaglest-source/BUILD.md")
	if err == nil {
		t.Error("Expected error when loading non-G3D file")
	}
}

func TestG3DModelUtilityFunctions(t *testing.T) {
	// Load a test model
	modelPath := "../../megaglest-source/data/glest_game/techs/megapack/factions/magic/units/initiate/models/initiate_standing.g3d"
	model, err := LoadG3D(modelPath)
	if err != nil {
		t.Fatalf("Failed to load test model: %v", err)
	}

	// Test GetTotalVertexCount
	totalVertices := model.GetTotalVertexCount()
	expectedVertices := uint32(0)
	for _, mesh := range model.Meshes {
		expectedVertices += mesh.Header.FrameCount * mesh.Header.VertexCount
	}
	if totalVertices != expectedVertices {
		t.Errorf("GetTotalVertexCount(): expected %d, got %d", expectedVertices, totalVertices)
	}

	// Test GetTotalTriangleCount
	totalTriangles := model.GetTotalTriangleCount()
	expectedTriangles := uint32(0)
	for _, mesh := range model.Meshes {
		expectedTriangles += mesh.Header.IndexCount / 3
	}
	if totalTriangles != expectedTriangles {
		t.Errorf("GetTotalTriangleCount(): expected %d, got %d", expectedTriangles, totalTriangles)
	}

	// Test HasTextures
	expectedHasTextures := false
	for _, mesh := range model.Meshes {
		if len(mesh.TextureNames) > 0 {
			expectedHasTextures = true
			break
		}
	}
	if model.HasTextures() != expectedHasTextures {
		t.Errorf("HasTextures(): expected %t, got %t", expectedHasTextures, model.HasTextures())
	}
}