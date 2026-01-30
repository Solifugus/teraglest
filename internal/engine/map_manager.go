package engine

import (
	"fmt"
	"os"
	"path/filepath"

	"teraglest/internal/data"
)

// MapManager handles loading and caching of maps using AssetManager
type MapManager struct {
	assetManager *data.AssetManager
	dataRoot     string // Root path for game data (maps, tilesets)
}

// NewMapManager creates a new map manager with the specified asset manager and data root
func NewMapManager(assetManager *data.AssetManager, dataRoot string) *MapManager {
	return &MapManager{
		assetManager: assetManager,
		dataRoot:     dataRoot,
	}
}

// LoadMap loads a map by name, using AssetManager for caching
func (mm *MapManager) LoadMap(mapName string) (*Map, error) {
	// Construct map file path
	mapPath := filepath.Join(mm.dataRoot, "maps", mapName+".mgm")

	// Check if .mgm file exists, if not try .gbm
	if !mm.fileExists(mapPath) {
		mapPath = filepath.Join(mm.dataRoot, "maps", mapName+".gbm")
		if !mm.fileExists(mapPath) {
			return nil, fmt.Errorf("map file not found: %s (.mgm or .gbm)", mapName)
		}
	}

	// Create cache key
	cacheKey := "map:" + mapName

	// Check cache first (using AssetManager's cache)
	if cached := mm.getCachedMap(cacheKey); cached != nil {
		return cached, nil
	}

	// Load map from file
	mapLoader := NewMapLoader()
	mapData, err := mapLoader.ParseMapFile(mapPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load map %s: %w", mapName, err)
	}

	// Load associated tileset
	tileset, err := mm.LoadTileset(mapData.TilesetName)
	if err != nil {
		return nil, fmt.Errorf("failed to load tileset for map %s: %w", mapName, err)
	}
	mapData.Tileset = tileset

	// Cache the loaded map (store in AssetManager cache)
	mm.cacheMap(cacheKey, mapData)

	return mapData, nil
}

// LoadTileset loads a tileset by name, using AssetManager for caching
func (mm *MapManager) LoadTileset(tilesetName string) (*Tileset, error) {
	// Create cache key
	cacheKey := "tileset:" + tilesetName

	// Check cache first
	if cached := mm.getCachedTileset(cacheKey); cached != nil {
		return cached, nil
	}

	// Load tileset from file
	tilesetLoader := NewTilesetLoader(mm.dataRoot)
	tileset, err := tilesetLoader.LoadTileset(tilesetName)
	if err != nil {
		return nil, fmt.Errorf("failed to load tileset %s: %w", tilesetName, err)
	}

	// Cache the loaded tileset
	mm.cacheTileset(cacheKey, tileset)

	return tileset, nil
}

// LoadMapWithTileset loads a map and its associated tileset in one operation
func (mm *MapManager) LoadMapWithTileset(mapName string) (*Map, error) {
	// This is the same as LoadMap since LoadMap already loads the tileset
	return mm.LoadMap(mapName)
}

// GetAvailableMaps returns a list of available map names
func (mm *MapManager) GetAvailableMaps() ([]string, error) {
	mapsDir := filepath.Join(mm.dataRoot, "maps")

	// Use AssetManager's file operations if available, otherwise use direct file access
	files, err := filepath.Glob(filepath.Join(mapsDir, "*.mgm"))
	if err != nil {
		return nil, fmt.Errorf("failed to scan maps directory: %w", err)
	}

	// Also check for .gbm files
	gbmFiles, err := filepath.Glob(filepath.Join(mapsDir, "*.gbm"))
	if err == nil {
		files = append(files, gbmFiles...)
	}

	// Extract map names (without extension)
	mapNames := make([]string, 0, len(files))
	for _, file := range files {
		base := filepath.Base(file)
		name := base[:len(base)-4] // Remove .mgm or .gbm extension
		mapNames = append(mapNames, name)
	}

	return mapNames, nil
}

// GetAvailableTilesets returns a list of available tileset names
func (mm *MapManager) GetAvailableTilesets() ([]string, error) {
	tilesetsDir := filepath.Join(mm.dataRoot, "tilesets")

	// Scan for tileset directories
	entries, err := filepath.Glob(filepath.Join(tilesetsDir, "*"))
	if err != nil {
		return nil, fmt.Errorf("failed to scan tilesets directory: %w", err)
	}

	tilesetNames := make([]string, 0)
	for _, entry := range entries {
		if mm.isDirectory(entry) {
			name := filepath.Base(entry)
			// Verify that the tileset XML file exists
			xmlPath := filepath.Join(entry, name+".xml")
			if mm.fileExists(xmlPath) {
				tilesetNames = append(tilesetNames, name)
			}
		}
	}

	return tilesetNames, nil
}

// ValidateMap performs validation on a loaded map
func (mm *MapManager) ValidateMap(mapData *Map) []string {
	var issues []string

	// Check basic map properties
	if mapData.Width < 16 || mapData.Width > 1024 {
		issues = append(issues, fmt.Sprintf("invalid map width: %d", mapData.Width))
	}
	if mapData.Height < 16 || mapData.Height > 1024 {
		issues = append(issues, fmt.Sprintf("invalid map height: %d", mapData.Height))
	}
	if mapData.MaxPlayers < 1 || mapData.MaxPlayers > 8 {
		issues = append(issues, fmt.Sprintf("invalid max players: %d", mapData.MaxPlayers))
	}

	// Check that start positions are within bounds
	for i, pos := range mapData.StartPositions {
		if !mapData.IsValidPosition(pos.X, pos.Y) {
			issues = append(issues, fmt.Sprintf("start position %d out of bounds: (%d, %d)", i+1, pos.X, pos.Y))
		}
	}

	// Check tileset availability
	if mapData.Tileset == nil {
		issues = append(issues, "tileset not loaded")
	}

	return issues
}

// Helper methods for caching integration with AssetManager

// getCachedMap retrieves a cached map from AssetManager
func (mm *MapManager) getCachedMap(cacheKey string) *Map {
	// For now, we'll implement a simple approach
	// In a full implementation, this would integrate with AssetManager's cache
	// but since AssetManager is in data package and Map is in engine package,
	// we need to be careful about circular imports
	return nil // TODO: implement proper caching integration
}

// cacheMap stores a map in AssetManager's cache
func (mm *MapManager) cacheMap(cacheKey string, mapData *Map) {
	// TODO: implement proper caching integration
	// This would store the map in AssetManager's cache with appropriate size estimation
}

// getCachedTileset retrieves a cached tileset from AssetManager
func (mm *MapManager) getCachedTileset(cacheKey string) *Tileset {
	// TODO: implement proper caching integration
	return nil
}

// cacheTileset stores a tileset in AssetManager's cache
func (mm *MapManager) cacheTileset(cacheKey string, tileset *Tileset) {
	// TODO: implement proper caching integration
}

// fileExists checks if a file exists
func (mm *MapManager) fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// isDirectory checks if a path is a directory
func (mm *MapManager) isDirectory(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// PrintSummary prints a summary of available maps and tilesets
func (mm *MapManager) PrintSummary() {
	fmt.Println("Map Manager Summary:")
	fmt.Printf("  Data Root: %s\n", mm.dataRoot)

	maps, err := mm.GetAvailableMaps()
	if err != nil {
		fmt.Printf("  Maps: Error - %v\n", err)
	} else {
		fmt.Printf("  Available Maps: %d\n", len(maps))
		if len(maps) > 0 && len(maps) <= 10 {
			for i, mapName := range maps {
				if i >= 5 { // Limit output
					fmt.Printf("    ... and %d more\n", len(maps)-5)
					break
				}
				fmt.Printf("    - %s\n", mapName)
			}
		}
	}

	tilesets, err := mm.GetAvailableTilesets()
	if err != nil {
		fmt.Printf("  Tilesets: Error - %v\n", err)
	} else {
		fmt.Printf("  Available Tilesets: %d\n", len(tilesets))
		if len(tilesets) > 0 && len(tilesets) <= 10 {
			for i, tilesetName := range tilesets {
				if i >= 5 { // Limit output
					fmt.Printf("    ... and %d more\n", len(tilesets)-5)
					break
				}
				fmt.Printf("    - %s\n", tilesetName)
			}
		}
	}
}