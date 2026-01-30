package engine

import (
	"os"
	"path/filepath"
	"testing"
)

// Test data paths
const (
	testDataRoot = "/home/solifugus/development/teraglest/megaglest-source/data/glest_game"
	testMapName  = "2rivers"
)

func TestMapLoaderCreation(t *testing.T) {
	loader := NewMapLoader()
	if loader == nil {
		t.Fatal("NewMapLoader() returned nil")
	}
}

func TestMapFileExists(t *testing.T) {
	mapPath := filepath.Join(testDataRoot, "maps", testMapName+".mgm")
	if _, err := os.Stat(mapPath); os.IsNotExist(err) {
		t.Skipf("Test map file not found: %s", mapPath)
	}
}

func TestParseMapFile(t *testing.T) {
	loader := NewMapLoader()
	mapPath := filepath.Join(testDataRoot, "maps", testMapName+".mgm")

	// Skip test if map file doesn't exist
	if _, err := os.Stat(mapPath); os.IsNotExist(err) {
		t.Skipf("Test map file not found: %s", mapPath)
	}

	mapData, err := loader.ParseMapFile(mapPath)
	if err != nil {
		t.Fatalf("Failed to parse map file: %v", err)
	}

	// Validate basic map properties
	if mapData == nil {
		t.Fatal("ParseMapFile returned nil map data")
	}

	if mapData.Width <= 0 || mapData.Height <= 0 {
		t.Errorf("Invalid map dimensions: %dx%d", mapData.Width, mapData.Height)
	}

	if mapData.MaxPlayers <= 0 || mapData.MaxPlayers > 8 {
		t.Errorf("Invalid max players: %d", mapData.MaxPlayers)
	}

	if len(mapData.StartPositions) != mapData.MaxPlayers {
		t.Errorf("Start positions count (%d) doesn't match max players (%d)",
			len(mapData.StartPositions), mapData.MaxPlayers)
	}

	t.Logf("Successfully parsed map: %dx%d, %d players, tileset: %s",
		mapData.Width, mapData.Height, mapData.MaxPlayers, mapData.TilesetName)
}

func TestMapDataAccess(t *testing.T) {
	loader := NewMapLoader()
	mapPath := filepath.Join(testDataRoot, "maps", testMapName+".mgm")

	// Skip test if map file doesn't exist
	if _, err := os.Stat(mapPath); os.IsNotExist(err) {
		t.Skipf("Test map file not found: %s", mapPath)
	}

	mapData, err := loader.ParseMapFile(mapPath)
	if err != nil {
		t.Fatalf("Failed to parse map file: %v", err)
	}

	// Test terrain data access
	testCases := []struct {
		name string
		x, y int
	}{
		{"Origin", 0, 0},
		{"Center", mapData.Width / 2, mapData.Height / 2},
		{"Corner", mapData.Width - 1, mapData.Height - 1},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if !mapData.IsValidPosition(tc.x, tc.y) {
				t.Errorf("Position (%d, %d) should be valid for %dx%d map",
					tc.x, tc.y, mapData.Width, mapData.Height)
			}

			height := mapData.GetHeightAt(tc.x, tc.y)
			if height < -100 || height > 100 {
				t.Errorf("Height at (%d, %d) seems unrealistic: %.2f", tc.x, tc.y, height)
			}

			surface := mapData.GetSurfaceAt(tc.x, tc.y)
			if surface < SurfaceGrass || surface > SurfaceGround {
				t.Errorf("Surface type at (%d, %d) is invalid: %d", tc.x, tc.y, int(surface))
			}

			object := mapData.GetObjectAt(tc.x, tc.y)
			if object < 0 {
				t.Errorf("Object index at (%d, %d) should not be negative: %d", tc.x, tc.y, object)
			}
		})
	}
}

func TestMapStartPositions(t *testing.T) {
	loader := NewMapLoader()
	mapPath := filepath.Join(testDataRoot, "maps", testMapName+".mgm")

	// Skip test if map file doesn't exist
	if _, err := os.Stat(mapPath); os.IsNotExist(err) {
		t.Skipf("Test map file not found: %s", mapPath)
	}

	mapData, err := loader.ParseMapFile(mapPath)
	if err != nil {
		t.Fatalf("Failed to parse map file: %v", err)
	}

	// Test start positions
	for i := 0; i < mapData.MaxPlayers; i++ {
		pos, exists := mapData.GetPlayerStartPosition(i)
		if !exists {
			t.Errorf("Start position for player %d not found", i+1)
			continue
		}

		if !mapData.IsValidPosition(pos.X, pos.Y) {
			t.Errorf("Start position for player %d is out of bounds: (%d, %d)",
				i+1, pos.X, pos.Y)
		}

		t.Logf("Player %d start position: (%d, %d)", i+1, pos.X, pos.Y)
	}

	// Test invalid player index
	_, exists := mapData.GetPlayerStartPosition(-1)
	if exists {
		t.Error("GetPlayerStartPosition should return false for invalid player index")
	}

	_, exists = mapData.GetPlayerStartPosition(mapData.MaxPlayers)
	if exists {
		t.Error("GetPlayerStartPosition should return false for player index >= MaxPlayers")
	}
}

func TestMapVersionHandling(t *testing.T) {
	// Test version string representation
	testCases := []struct {
		version  MapVersion
		expected string
	}{
		{MapVersionGBM, "GBM"},
		{MapVersionMGM, "MGM"},
	}

	for _, tc := range testCases {
		if tc.version == MapVersionGBM {
			// We expect version 1 for GBM format
			if int(tc.version) != 1 {
				t.Errorf("MapVersionGBM should be 1, got %d", int(tc.version))
			}
		} else if tc.version == MapVersionMGM {
			// We expect version 2 for MGM format
			if int(tc.version) != 2 {
				t.Errorf("MapVersionMGM should be 2, got %d", int(tc.version))
			}
		}
	}
}

func TestMapSurfaceTypes(t *testing.T) {
	testCases := []struct {
		surface  MapSurfaceType
		expected string
	}{
		{SurfaceGrass, "Grass"},
		{SurfaceSecondaryGrass, "Secondary_Grass"},
		{SurfaceRoad, "Road"},
		{SurfaceStone, "Stone"},
		{SurfaceGround, "Ground"},
	}

	for _, tc := range testCases {
		if tc.surface.String() != tc.expected {
			t.Errorf("Surface type %d string representation: expected %s, got %s",
				int(tc.surface), tc.expected, tc.surface.String())
		}
	}
}

func TestMapPrintSummary(t *testing.T) {
	loader := NewMapLoader()
	mapPath := filepath.Join(testDataRoot, "maps", testMapName+".mgm")

	// Skip test if map file doesn't exist
	if _, err := os.Stat(mapPath); os.IsNotExist(err) {
		t.Skipf("Test map file not found: %s", mapPath)
	}

	mapData, err := loader.ParseMapFile(mapPath)
	if err != nil {
		t.Fatalf("Failed to parse map file: %v", err)
	}

	// Test that PrintSummary doesn't panic
	// This is mainly a smoke test
	mapData.PrintSummary()
	t.Log("PrintSummary completed without panic")
}

func TestTilesetNameDetermination(t *testing.T) {
	loader := NewMapLoader()

	testCases := []struct {
		mapTitle string
		expected string
	}{
		{"Desert Battle", "desert2"},
		{"Winter Conquest", "winter"},
		{"Forest Ambush", "forest"},
		{"Jungle Warfare", "jungle"},
		{"Hell's Gate", "hell"},
		{"Regular Map", "meadow"}, // Default
	}

	for _, tc := range testCases {
		mapData := &Map{Title: tc.mapTitle}
		tileset := loader.determineTilesetName(mapData)

		if tileset != tc.expected {
			t.Errorf("Map title '%s': expected tileset '%s', got '%s'",
				tc.mapTitle, tc.expected, tileset)
		}
	}
}

// Benchmark tests
func BenchmarkMapParsing(b *testing.B) {
	mapPath := filepath.Join(testDataRoot, "maps", testMapName+".mgm")

	// Skip benchmark if map file doesn't exist
	if _, err := os.Stat(mapPath); os.IsNotExist(err) {
		b.Skipf("Test map file not found: %s", mapPath)
	}

	loader := NewMapLoader()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := loader.ParseMapFile(mapPath)
		if err != nil {
			b.Fatalf("Failed to parse map file: %v", err)
		}
	}
}

func BenchmarkMapDataAccess(b *testing.B) {
	loader := NewMapLoader()
	mapPath := filepath.Join(testDataRoot, "maps", testMapName+".mgm")

	// Skip benchmark if map file doesn't exist
	if _, err := os.Stat(mapPath); os.IsNotExist(err) {
		b.Skipf("Test map file not found: %s", mapPath)
	}

	mapData, err := loader.ParseMapFile(mapPath)
	if err != nil {
		b.Fatalf("Failed to parse map file: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Access terrain data
		x := i % mapData.Width
		y := (i / mapData.Width) % mapData.Height

		_ = mapData.GetHeightAt(x, y)
		_ = mapData.GetSurfaceAt(x, y)
		_ = mapData.GetObjectAt(x, y)
	}
}