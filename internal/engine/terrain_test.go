package engine

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTilesetLoaderCreation(t *testing.T) {
	loader := NewTilesetLoader(testDataRoot)
	if loader == nil {
		t.Fatal("NewTilesetLoader() returned nil")
	}

	if loader.basePath != testDataRoot {
		t.Errorf("Expected base path %s, got %s", testDataRoot, loader.basePath)
	}
}

func TestTilesetFileExists(t *testing.T) {
	tilesetName := "meadow"
	xmlPath := filepath.Join(testDataRoot, "tilesets", tilesetName, tilesetName+".xml")

	if _, err := os.Stat(xmlPath); os.IsNotExist(err) {
		t.Skipf("Test tileset file not found: %s", xmlPath)
	}
}

func TestLoadTileset(t *testing.T) {
	loader := NewTilesetLoader(testDataRoot)
	tilesetName := "meadow"

	// Skip test if tileset file doesn't exist
	xmlPath := filepath.Join(testDataRoot, "tilesets", tilesetName, tilesetName+".xml")
	if _, err := os.Stat(xmlPath); os.IsNotExist(err) {
		t.Skipf("Test tileset file not found: %s", xmlPath)
	}

	tileset, err := loader.LoadTileset(tilesetName)
	if err != nil {
		t.Fatalf("Failed to load tileset: %v", err)
	}

	// Validate basic tileset properties
	if tileset == nil {
		t.Fatal("LoadTileset returned nil tileset")
	}

	if tileset.Name != tilesetName {
		t.Errorf("Expected tileset name %s, got %s", tilesetName, tileset.Name)
	}

	if len(tileset.Surfaces) == 0 {
		t.Error("Tileset should have at least one surface")
	}

	if len(tileset.Objects) == 0 {
		t.Error("Tileset should have at least one object")
	}

	if tileset.AmbientSounds == nil {
		t.Error("Tileset should have ambient sounds configuration")
	}

	t.Logf("Successfully loaded tileset: %s with %d surfaces, %d objects",
		tileset.Name, len(tileset.Surfaces), len(tileset.Objects))
}

func TestTilesetSurfaces(t *testing.T) {
	loader := NewTilesetLoader(testDataRoot)
	tilesetName := "meadow"

	// Skip test if tileset file doesn't exist
	xmlPath := filepath.Join(testDataRoot, "tilesets", tilesetName, tilesetName+".xml")
	if _, err := os.Stat(xmlPath); os.IsNotExist(err) {
		t.Skipf("Test tileset file not found: %s", xmlPath)
	}

	tileset, err := loader.LoadTileset(tilesetName)
	if err != nil {
		t.Fatalf("Failed to load tileset: %v", err)
	}

	// Test surface access
	for i, surface := range tileset.Surfaces {
		if surface.Index != i+1 {
			t.Errorf("Surface %d has incorrect index: expected %d, got %d",
				i, i+1, surface.Index)
		}

		if len(surface.Textures) == 0 {
			t.Errorf("Surface %d should have at least one texture", i+1)
		}

		if surface.TotalProbability <= 0 {
			t.Errorf("Surface %d should have positive total probability, got %.3f",
				i+1, surface.TotalProbability)
		}

		// Test texture variations
		for j, texture := range surface.Textures {
			if texture.Path == "" {
				t.Errorf("Surface %d, texture %d has empty path", i+1, j)
			}

			if texture.Probability <= 0 {
				t.Errorf("Surface %d, texture %d has invalid probability: %.3f",
					i+1, j, texture.Probability)
			}
		}

		// Test random texture selection
		randomTexture := surface.GetRandomTexture(0.5)
		if randomTexture == "" {
			t.Errorf("Surface %d GetRandomTexture returned empty string", i+1)
		}
	}
}

func TestTilesetObjects(t *testing.T) {
	loader := NewTilesetLoader(testDataRoot)
	tilesetName := "meadow"

	// Skip test if tileset file doesn't exist
	xmlPath := filepath.Join(testDataRoot, "tilesets", tilesetName, tilesetName+".xml")
	if _, err := os.Stat(xmlPath); os.IsNotExist(err) {
		t.Skipf("Test tileset file not found: %s", xmlPath)
	}

	tileset, err := loader.LoadTileset(tilesetName)
	if err != nil {
		t.Fatalf("Failed to load tileset: %v", err)
	}

	// Test object access
	for i, object := range tileset.Objects {
		if object.Index != i+1 {
			t.Errorf("Object %d has incorrect index: expected %d, got %d",
				i, i+1, object.Index)
		}

		if len(object.Models) == 0 {
			t.Errorf("Object %d should have at least one model", i+1)
		}

		// Test model paths
		for j, modelPath := range object.Models {
			if modelPath == "" {
				t.Errorf("Object %d, model %d has empty path", i+1, j)
			}

			// Check that model path has .g3d extension
			if filepath.Ext(modelPath) != ".g3d" {
				t.Logf("Warning: Object %d, model %d doesn't have .g3d extension: %s",
					i+1, j, modelPath)
			}
		}

		t.Logf("Object %d: %d models, walkable=%t", i+1, len(object.Models), object.Walkable)
	}
}

func TestTilesetAccessMethods(t *testing.T) {
	loader := NewTilesetLoader(testDataRoot)
	tilesetName := "meadow"

	// Skip test if tileset file doesn't exist
	xmlPath := filepath.Join(testDataRoot, "tilesets", tilesetName, tilesetName+".xml")
	if _, err := os.Stat(xmlPath); os.IsNotExist(err) {
		t.Skipf("Test tileset file not found: %s", xmlPath)
	}

	tileset, err := loader.LoadTileset(tilesetName)
	if err != nil {
		t.Fatalf("Failed to load tileset: %v", err)
	}

	// Test GetSurface method
	surface := tileset.GetSurface(1)
	if surface == nil {
		t.Error("GetSurface(1) should return a surface")
	}

	surface = tileset.GetSurface(0)
	if surface != nil {
		t.Error("GetSurface(0) should return nil (invalid index)")
	}

	surface = tileset.GetSurface(len(tileset.Surfaces) + 1)
	if surface != nil {
		t.Error("GetSurface(out_of_bounds) should return nil")
	}

	// Test GetObject method
	object := tileset.GetObject(1)
	if object == nil {
		t.Error("GetObject(1) should return an object")
	}

	object = tileset.GetObject(0)
	if object != nil {
		t.Error("GetObject(0) should return nil (invalid index)")
	}

	object = tileset.GetObject(len(tileset.Objects) + 1)
	if object != nil {
		t.Error("GetObject(out_of_bounds) should return nil")
	}

	// Test IsObjectWalkable method
	walkable := tileset.IsObjectWalkable(1)
	if walkable != tileset.Objects[0].Walkable {
		t.Error("IsObjectWalkable(1) should match first object's walkable property")
	}

	walkable = tileset.IsObjectWalkable(0)
	if !walkable {
		t.Error("IsObjectWalkable(0) should return true (no object = walkable)")
	}
}

func TestTilesetAmbientSounds(t *testing.T) {
	loader := NewTilesetLoader(testDataRoot)
	tilesetName := "meadow"

	// Skip test if tileset file doesn't exist
	xmlPath := filepath.Join(testDataRoot, "tilesets", tilesetName, tilesetName+".xml")
	if _, err := os.Stat(xmlPath); os.IsNotExist(err) {
		t.Skipf("Test tileset file not found: %s", xmlPath)
	}

	tileset, err := loader.LoadTileset(tilesetName)
	if err != nil {
		t.Fatalf("Failed to load tileset: %v", err)
	}

	if tileset.AmbientSounds == nil {
		t.Fatal("Tileset should have ambient sounds configuration")
	}

	sounds := tileset.AmbientSounds

	// Test sound configurations
	testCases := []struct {
		name   string
		config AudioConfig
	}{
		{"Day Sound", sounds.DaySound},
		{"Night Sound", sounds.NightSound},
		{"Rain Sound", sounds.RainSound},
		{"Snow Sound", sounds.SnowSound},
		{"Day Start Sound", sounds.DayStartSound},
		{"Night Start Sound", sounds.NightStartSound},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.config.Volume < 0.0 || tc.config.Volume > 1.0 {
				t.Errorf("%s volume should be between 0.0 and 1.0, got %.2f",
					tc.name, tc.config.Volume)
			}

			if tc.config.Enabled && tc.config.Path == "" {
				t.Errorf("%s is enabled but has empty path", tc.name)
			}
		})
	}
}

func TestTilesetParameters(t *testing.T) {
	loader := NewTilesetLoader(testDataRoot)
	tilesetName := "meadow"

	// Skip test if tileset file doesn't exist
	xmlPath := filepath.Join(testDataRoot, "tilesets", tilesetName, tilesetName+".xml")
	if _, err := os.Stat(xmlPath); os.IsNotExist(err) {
		t.Skipf("Test tileset file not found: %s", xmlPath)
	}

	tileset, err := loader.LoadTileset(tilesetName)
	if err != nil {
		t.Fatalf("Failed to load tileset: %v", err)
	}

	params := tileset.Parameters

	// Test water configuration
	water := params.Water
	if water.Effects && len(water.Textures) == 0 {
		t.Error("Water effects enabled but no textures specified")
	}

	if water.FrameCount != len(water.Textures) {
		t.Errorf("Water frame count (%d) doesn't match texture count (%d)",
			water.FrameCount, len(water.Textures))
	}

	// Test fog configuration
	fog := params.Fog
	if fog.Density < 0.0 || fog.Density > 1.0 {
		t.Errorf("Fog density should be between 0.0 and 1.0, got %.3f", fog.Density)
	}

	if fog.Red < 0.0 || fog.Red > 1.0 {
		t.Errorf("Fog red component should be between 0.0 and 1.0, got %.3f", fog.Red)
	}

	// Test lighting configuration
	lighting := params.Lighting
	if lighting.SunRed < 0.0 || lighting.SunRed > 1.0 {
		t.Errorf("Sun red component should be between 0.0 and 1.0, got %.3f", lighting.SunRed)
	}

	if lighting.DayTime <= 0 {
		t.Errorf("Day time should be positive, got %.1f", lighting.DayTime)
	}

	// Test weather configuration
	weather := params.Weather
	totalProbability := weather.SunProbability + weather.RainProbability + weather.SnowProbability
	if totalProbability < 0.9 || totalProbability > 1.1 {
		t.Errorf("Weather probabilities should sum to ~1.0, got %.3f", totalProbability)
	}
}

func TestTilesetPrintSummary(t *testing.T) {
	loader := NewTilesetLoader(testDataRoot)
	tilesetName := "meadow"

	// Skip test if tileset file doesn't exist
	xmlPath := filepath.Join(testDataRoot, "tilesets", tilesetName, tilesetName+".xml")
	if _, err := os.Stat(xmlPath); os.IsNotExist(err) {
		t.Skipf("Test tileset file not found: %s", xmlPath)
	}

	tileset, err := loader.LoadTileset(tilesetName)
	if err != nil {
		t.Fatalf("Failed to load tileset: %v", err)
	}

	// Test that PrintSummary doesn't panic
	// This is mainly a smoke test
	tileset.PrintSummary()
	t.Log("PrintSummary completed without panic")
}

func TestTilesetNonExistent(t *testing.T) {
	loader := NewTilesetLoader(testDataRoot)

	_, err := loader.LoadTileset("nonexistent_tileset")
	if err == nil {
		t.Error("LoadTileset should return error for non-existent tileset")
	}
}

func TestTilesetMultipleLoads(t *testing.T) {
	loader := NewTilesetLoader(testDataRoot)
	tilesetName := "meadow"

	// Skip test if tileset file doesn't exist
	xmlPath := filepath.Join(testDataRoot, "tilesets", tilesetName, tilesetName+".xml")
	if _, err := os.Stat(xmlPath); os.IsNotExist(err) {
		t.Skipf("Test tileset file not found: %s", xmlPath)
	}

	// Load tileset multiple times to ensure consistency
	tileset1, err := loader.LoadTileset(tilesetName)
	if err != nil {
		t.Fatalf("First load failed: %v", err)
	}

	tileset2, err := loader.LoadTileset(tilesetName)
	if err != nil {
		t.Fatalf("Second load failed: %v", err)
	}

	// Compare basic properties
	if tileset1.Name != tileset2.Name {
		t.Error("Multiple loads produced different tileset names")
	}

	if len(tileset1.Surfaces) != len(tileset2.Surfaces) {
		t.Error("Multiple loads produced different surface counts")
	}

	if len(tileset1.Objects) != len(tileset2.Objects) {
		t.Error("Multiple loads produced different object counts")
	}
}

// Benchmark tests
func BenchmarkTilesetLoading(b *testing.B) {
	tilesetName := "meadow"
	xmlPath := filepath.Join(testDataRoot, "tilesets", tilesetName, tilesetName+".xml")

	// Skip benchmark if tileset file doesn't exist
	if _, err := os.Stat(xmlPath); os.IsNotExist(err) {
		b.Skipf("Test tileset file not found: %s", xmlPath)
	}

	loader := NewTilesetLoader(testDataRoot)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := loader.LoadTileset(tilesetName)
		if err != nil {
			b.Fatalf("Failed to load tileset: %v", err)
		}
	}
}

func BenchmarkTilesetRandomTexture(b *testing.B) {
	loader := NewTilesetLoader(testDataRoot)
	tilesetName := "meadow"

	// Skip benchmark if tileset file doesn't exist
	xmlPath := filepath.Join(testDataRoot, "tilesets", tilesetName, tilesetName+".xml")
	if _, err := os.Stat(xmlPath); os.IsNotExist(err) {
		b.Skipf("Test tileset file not found: %s", xmlPath)
	}

	tileset, err := loader.LoadTileset(tilesetName)
	if err != nil {
		b.Fatalf("Failed to load tileset: %v", err)
	}

	if len(tileset.Surfaces) == 0 {
		b.Skip("No surfaces available for benchmark")
	}

	surface := &tileset.Surfaces[0]

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		randomValue := float32(i%1000) / 1000.0 // 0.0 to 0.999
		_ = surface.GetRandomTexture(randomValue)
	}
}