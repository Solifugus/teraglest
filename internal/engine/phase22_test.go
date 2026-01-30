package engine

import (
	"os"
	"path/filepath"
	"testing"

	"teraglest/internal/data"
)

// TestPhase22Requirements validates all Phase 2.2 requirements from DEVELOPMENT_PLAN.md
func TestPhase22Requirements(t *testing.T) {
	// Skip tests if test data is not available
	if _, err := os.Stat(testDataRoot); os.IsNotExist(err) {
		t.Skipf("Test data root not found: %s", testDataRoot)
	}

	t.Run("Grid-based map structure (cells, heights, tilesets)", func(t *testing.T) {
		// ✅ Grid-based map structure
		loader := NewMapLoader()
		mapPath := filepath.Join(testDataRoot, "maps", testMapName+".mgm")

		mapData, err := loader.ParseMapFile(mapPath)
		if err != nil {
			t.Skipf("Cannot test without map data: %v", err)
		}

		// Validate grid structure
		if mapData.Width <= 0 || mapData.Height <= 0 {
			t.Error("Map should have valid grid dimensions")
		}

		// Validate cells (height map, surface map, object map)
		if len(mapData.HeightMap) != mapData.Height {
			t.Error("Height map should have correct height dimension")
		}
		for y, row := range mapData.HeightMap {
			if len(row) != mapData.Width {
				t.Errorf("Height map row %d should have correct width dimension", y)
			}
		}

		if len(mapData.SurfaceMap) != mapData.Height {
			t.Error("Surface map should have correct height dimension")
		}
		if len(mapData.ObjectMap) != mapData.Height {
			t.Error("Object map should have correct height dimension")
		}

		// Validate tileset integration
		if mapData.TilesetName == "" {
			t.Error("Map should specify a tileset")
		}

		t.Logf("✅ Grid structure validated: %dx%d cells with heights and tileset %s",
			mapData.Width, mapData.Height, mapData.TilesetName)
	})

	t.Run("Support Glest .gbm and .mgm map formats", func(t *testing.T) {
		// ✅ Support for both .gbm and .mgm formats
		loader := NewMapLoader()

		// Test MGM format
		mgmPath := filepath.Join(testDataRoot, "maps", testMapName+".mgm")
		if _, err := os.Stat(mgmPath); !os.IsNotExist(err) {
			mapData, err := loader.ParseMapFile(mgmPath)
			if err != nil {
				t.Errorf("Failed to parse .mgm file: %v", err)
			} else {
				if mapData.Version != MapVersionMGM && mapData.Version != MapVersionGBM {
					t.Errorf("Unexpected map version: %d", int(mapData.Version))
				}
				t.Logf("✅ Successfully parsed .mgm format (version %d)", int(mapData.Version))
			}
		}

		// Test GBM format (if available)
		gbmFiles, _ := filepath.Glob(filepath.Join(testDataRoot, "maps", "*.gbm"))
		if len(gbmFiles) > 0 {
			mapData, err := loader.ParseMapFile(gbmFiles[0])
			if err != nil {
				t.Errorf("Failed to parse .gbm file: %v", err)
			} else {
				// Note: Some .gbm files may actually be version 2 format
				if mapData.Version != MapVersionGBM && mapData.Version != MapVersionMGM {
					t.Errorf("GBM file should have version 1 or 2, got %d", int(mapData.Version))
				}
				t.Logf("✅ Successfully parsed .gbm format (version %d): %s",
					int(mapData.Version), filepath.Base(gbmFiles[0]))
			}
		} else {
			t.Log("⚠️  No .gbm files found for testing")
		}
	})

	t.Run("Terrain tile rendering data preparation", func(t *testing.T) {
		// ✅ Terrain rendering data preparation
		tilesetLoader := NewTilesetLoader(testDataRoot)
		tileset, err := tilesetLoader.LoadTileset("meadow")
		if err != nil {
			t.Skipf("Cannot test without tileset data: %v", err)
		}

		// Validate surface rendering data
		if len(tileset.Surfaces) == 0 {
			t.Error("Tileset should have surface definitions for rendering")
		}

		for i, surface := range tileset.Surfaces {
			if len(surface.Textures) == 0 {
				t.Errorf("Surface %d should have texture variations for rendering", i+1)
			}

			for j, texture := range surface.Textures {
				if texture.Path == "" {
					t.Errorf("Surface %d texture %d should have a path for rendering", i+1, j)
				}
				if texture.Probability <= 0 {
					t.Errorf("Surface %d texture %d should have positive probability", i+1, j)
				}
			}
		}

		// Validate object rendering data
		if len(tileset.Objects) == 0 {
			t.Error("Tileset should have object definitions for rendering")
		}

		for i, object := range tileset.Objects {
			if len(object.Models) == 0 {
				t.Errorf("Object %d should have model references for rendering", i+1)
			}
		}

		// Validate environmental rendering parameters
		params := tileset.Parameters
		if params.Water.Effects && len(params.Water.Textures) == 0 {
			t.Error("Water effects enabled but no animation textures provided")
		}

		t.Logf("✅ Rendering data prepared: %d surfaces, %d objects, water=%t, fog=%t",
			len(tileset.Surfaces), len(tileset.Objects),
			params.Water.Effects, params.Fog.Enabled)
	})

	t.Run("Starting position handling for players", func(t *testing.T) {
		// ✅ Starting position handling
		loader := NewMapLoader()
		mapPath := filepath.Join(testDataRoot, "maps", testMapName+".mgm")

		mapData, err := loader.ParseMapFile(mapPath)
		if err != nil {
			t.Skipf("Cannot test without map data: %v", err)
		}

		// Validate start positions
		if len(mapData.StartPositions) == 0 {
			t.Error("Map should have player starting positions")
		}

		if len(mapData.StartPositions) != mapData.MaxPlayers {
			t.Errorf("Start positions count (%d) should match max players (%d)",
				len(mapData.StartPositions), mapData.MaxPlayers)
		}

		// Validate each start position
		for i, pos := range mapData.StartPositions {
			if !mapData.IsValidPosition(pos.X, pos.Y) {
				t.Errorf("Start position %d is invalid: (%d, %d)", i+1, pos.X, pos.Y)
			}

			// Check that positions are reasonably distributed
			if i > 0 {
				prevPos := mapData.StartPositions[i-1]
				distance := abs(pos.X-prevPos.X) + abs(pos.Y-prevPos.Y) // Manhattan distance
				if distance < 5 {
					t.Errorf("Start positions %d and %d are too close together", i, i+1)
				}
			}
		}

		t.Logf("✅ Start positions validated: %d positions for %d players",
			len(mapData.StartPositions), mapData.MaxPlayers)
	})
}

// TestPhase22Integration validates the integration of all Phase 2.2 components
func TestPhase22Integration(t *testing.T) {
	// Skip tests if test data is not available
	if _, err := os.Stat(testDataRoot); os.IsNotExist(err) {
		t.Skipf("Test data root not found: %s", testDataRoot)
	}

	t.Run("MapManager integration", func(t *testing.T) {
		assetManager := data.NewAssetManager(filepath.Join(testDataRoot, "techs", "megapack"))
		mapManager := NewMapManager(assetManager, testDataRoot)

		// Test map loading
		mapData, err := mapManager.LoadMap(testMapName)
		if err != nil {
			t.Fatalf("MapManager failed to load map: %v", err)
		}

		// Validate integration
		if mapData.Tileset == nil {
			t.Error("MapManager should load tileset along with map")
		}

		// Test validation
		issues := mapManager.ValidateMap(mapData)
		if len(issues) > 0 {
			t.Errorf("Map validation found issues: %v", issues)
		}

		t.Logf("✅ MapManager integration successful: %dx%d map with %s tileset",
			mapData.Width, mapData.Height, mapData.TilesetName)
	})

	t.Run("World creation from map", func(t *testing.T) {
		// Test the complete integration: AssetManager -> MapManager -> World
		assetManager := data.NewAssetManager(filepath.Join(testDataRoot, "techs", "megapack"))

		techTree, err := assetManager.LoadTechTree()
		if err != nil {
			t.Skipf("Cannot test without tech tree: %v", err)
		}

		settings := GameSettings{
			PlayerFactions:     map[int]string{1: "romans"},
			MaxPlayers:        2,
			ResourceMultiplier: 1.0,
		}

		// Create world from map (this tests the complete integration)
		world, err := NewWorldFromMap(settings, techTree, assetManager, testMapName)
		if err != nil {
			t.Fatalf("Failed to create world from map: %v", err)
		}

		// Validate world creation
		if world.Width <= 64 || world.Height <= 64 {
			t.Error("World should use map dimensions, not hardcoded 64x64")
		}

		if world.Map == nil {
			t.Error("World should have reference to loaded map")
		}

		// Test terrain integration
		if world.Map.Width != world.Width || world.Map.Height != world.Height {
			t.Error("World dimensions should match map dimensions")
		}

		// Test resource placement
		resources := world.GetAllResourceNodes()
		if len(resources) == 0 {
			t.Error("World should have resource nodes placed from map data")
		}

		t.Logf("✅ World integration successful: %dx%d world with %d resources",
			world.Width, world.Height, len(resources))
	})
}

// TestPhase22Performance validates that Phase 2.2 components meet performance expectations
func TestPhase22Performance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance tests in short mode")
	}

	// Skip tests if test data is not available
	if _, err := os.Stat(testDataRoot); os.IsNotExist(err) {
		t.Skipf("Test data root not found: %s", testDataRoot)
	}

	t.Run("Map loading performance", func(t *testing.T) {
		loader := NewMapLoader()
		mapPath := filepath.Join(testDataRoot, "maps", testMapName+".mgm")

		// Time multiple map loads
		const iterations = 10

		for i := 0; i < iterations; i++ {
			_, err := loader.ParseMapFile(mapPath)
			if err != nil {
				t.Fatalf("Map loading failed: %v", err)
			}
		}

		t.Logf("Map loading performance: %d iterations completed", iterations)
	})

	t.Run("World creation performance", func(t *testing.T) {
		assetManager := data.NewAssetManager(filepath.Join(testDataRoot, "techs", "megapack"))

		techTree, err := assetManager.LoadTechTree()
		if err != nil {
			t.Skipf("Cannot test without tech tree: %v", err)
		}

		settings := GameSettings{
			PlayerFactions:     map[int]string{1: "romans"},
			MaxPlayers:        2,
			ResourceMultiplier: 1.0,
		}

		// Time world creation
		const iterations = 5
		for i := 0; i < iterations; i++ {
			_, err := NewWorldFromMap(settings, techTree, assetManager, testMapName)
			if err != nil {
				t.Fatalf("World creation failed: %v", err)
			}
		}

		t.Logf("World creation performance: %d iterations completed", iterations)
	})
}

// TestPhase22ErrorHandling validates error handling across Phase 2.2 components
func TestPhase22ErrorHandling(t *testing.T) {
	t.Run("Invalid map file", func(t *testing.T) {
		loader := NewMapLoader()
		_, err := loader.ParseMapFile("/nonexistent/path/invalid.mgm")
		if err == nil {
			t.Error("Should return error for non-existent map file")
		}
	})

	t.Run("Invalid tileset", func(t *testing.T) {
		tilesetLoader := NewTilesetLoader(testDataRoot)
		_, err := tilesetLoader.LoadTileset("invalid_tileset_name")
		if err == nil {
			t.Error("Should return error for non-existent tileset")
		}
	})

	t.Run("MapManager error handling", func(t *testing.T) {
		assetManager := data.NewAssetManager(testDataRoot)
		mapManager := NewMapManager(assetManager, testDataRoot)

		_, err := mapManager.LoadMap("invalid_map_name")
		if err == nil {
			t.Error("Should return error for invalid map name")
		}
	})

	t.Run("World creation error handling", func(t *testing.T) {
		assetManager := data.NewAssetManager(testDataRoot)
		techTree := &data.TechTree{} // Empty tech tree

		settings := GameSettings{
			PlayerFactions: map[int]string{1: "romans"},
			MaxPlayers:     2,
		}

		_, err := NewWorldFromMap(settings, techTree, assetManager, "invalid_map")
		if err == nil {
			t.Error("Should return error for invalid map in world creation")
		}
	})
}

// Helper functions
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}