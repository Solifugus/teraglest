package engine

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"strings"
)

// Constants from MegaGlest map format specification
const (
	MaxTitleLength       = 128
	MaxAuthorLength      = 128
	MaxDescriptionLength = 256
	MaxDescriptionLengthV2 = 128

	MinMapCellDimension = 16
	MaxMapCellDimension = 1024

	MinMapCellHeight     = 0
	MaxMapCellHeight     = 20
	DefaultMapCellHeight = 10

	MinMapFactionCount = 1
	MaxMapFactionCount = 8

	// Magic number for version 2 format validation
	MapVersion2Magic = 0x01020304
)

// MapVersion represents the map file format version
type MapVersion int32

const (
	MapVersionGBM MapVersion = 1 // Glest Binary Map format
	MapVersionMGM MapVersion = 2 // MegaGlest Map format
)

// MapSurfaceType represents terrain surface types
type MapSurfaceType int8

const (
	SurfaceGrass MapSurfaceType = iota + 1
	SurfaceSecondaryGrass
	SurfaceRoad
	SurfaceStone
	SurfaceGround
)

// String returns the string representation of the surface type
func (s MapSurfaceType) String() string {
	switch s {
	case SurfaceGrass:
		return "Grass"
	case SurfaceSecondaryGrass:
		return "Secondary_Grass"
	case SurfaceRoad:
		return "Road"
	case SurfaceStone:
		return "Stone"
	case SurfaceGround:
		return "Ground"
	default:
		return "Unknown"
	}
}

// MapFileHeader represents the binary header of a map file
type MapFileHeader struct {
	Version     int32     // Format version (1=GBM, 2=MGM)
	MaxFactions int32     // Maximum number of players/factions
	Width       int32     // Map width in tiles
	Height      int32     // Map height in tiles
	HeightFactor int32    // Height scaling factor
	WaterLevel  int32     // Water level
	Title       [128]byte // Map title (null-terminated)
	Author      [128]byte // Map author (null-terminated)
	Description [256]byte // Description or extended data for v2
}

// MapFileHeaderV2 represents the extended header for version 2 maps
type MapFileHeaderV2 struct {
	ShortDesc    [128]byte // Short description
	Magic        int32     // 0x01020304 for validation
	CliffLevel   int32     // Cliff level
	CameraHeight int32     // Camera height
	Meta         [116]byte // Reserved space
}

// StartPosition represents a player starting position
type StartPosition struct {
	X int32
	Y int32
}

// Map represents a loaded game map with all terrain data
type Map struct {
	// Header information
	Title        string       `json:"title"`
	Author       string       `json:"author"`
	Description  string       `json:"description"`
	Width        int          `json:"width"`        // Map width in tiles
	Height       int          `json:"height"`       // Map height in tiles
	MaxPlayers   int          `json:"max_players"`  // Maximum players supported
	Version      MapVersion   `json:"version"`      // Map format version

	// Terrain data
	HeightMap       [][]float32    `json:"-"` // Terrain heights [y][x]
	SurfaceMap      [][]int8       `json:"-"` // Surface type indices [y][x]
	ObjectMap       [][]int8       `json:"-"` // Terrain object placements [y][x]
	StartPositions  []Vector2i     `json:"start_positions"` // Player starting positions

	// Rendering and gameplay data
	TilesetName     string         `json:"tileset_name"`
	Tileset         *Tileset       `json:"-"` // Loaded tileset (populated later)
	WaterLevel      float32        `json:"water_level"`
	HeightFactor    float32        `json:"height_factor"`
	CliffLevel      float32        `json:"cliff_level"`     // Version 2 only
	CameraHeight    float32        `json:"camera_height"`   // Version 2 only

	// Metadata
	FilePath        string         `json:"file_path"`       // Original file path
	FileSize        int64          `json:"file_size"`       // File size in bytes
}

// MapLoader handles parsing of MegaGlest map files
type MapLoader struct {
	// Could add asset manager reference for tileset loading later
}

// NewMapLoader creates a new map loader instance
func NewMapLoader() *MapLoader {
	return &MapLoader{}
}

// ParseMapFile parses a .mgm or .gbm map file and returns a Map structure
func (ml *MapLoader) ParseMapFile(filePath string) (*Map, error) {
	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open map file %s: %w", filePath, err)
	}
	defer file.Close()

	// Get file size
	fileInfo, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get file info for %s: %w", filePath, err)
	}

	// Parse the file
	mapData, err := ml.parseMapData(file)
	if err != nil {
		return nil, fmt.Errorf("failed to parse map data from %s: %w", filePath, err)
	}

	// Set metadata
	mapData.FilePath = filePath
	mapData.FileSize = fileInfo.Size()

	// Determine tileset name from file or use default
	mapData.TilesetName = ml.determineTilesetName(mapData)

	return mapData, nil
}

// parseMapData parses the binary map data from a reader
func (ml *MapLoader) parseMapData(reader io.Reader) (*Map, error) {
	// Read and parse header
	header, err := ml.parseHeader(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to parse map header: %w", err)
	}

	// Validate header
	if err := ml.validateHeader(header); err != nil {
		return nil, fmt.Errorf("invalid map header: %w", err)
	}

	// Create map structure
	mapData := &Map{
		Title:        ml.extractString(header.Title[:]),
		Author:       ml.extractString(header.Author[:]),
		Width:        int(header.Width),
		Height:       int(header.Height),
		MaxPlayers:   int(header.MaxFactions),
		Version:      MapVersion(header.Version),
		WaterLevel:   float32(header.WaterLevel),
		HeightFactor: float32(header.HeightFactor),
	}

	// Handle version-specific data
	if mapData.Version == MapVersionMGM {
		// Parse version 2 extended header
		v2Header, err := ml.parseHeaderV2(reader, header.Description[:])
		if err != nil {
			return nil, fmt.Errorf("failed to parse version 2 header: %w", err)
		}
		mapData.Description = ml.extractString(v2Header.ShortDesc[:])
		mapData.CliffLevel = float32(v2Header.CliffLevel)
		mapData.CameraHeight = float32(v2Header.CameraHeight)
	} else {
		mapData.Description = ml.extractString(header.Description[:])
	}

	// Parse start positions
	startPositions, err := ml.parseStartPositions(reader, mapData.MaxPlayers)
	if err != nil {
		return nil, fmt.Errorf("failed to parse start positions: %w", err)
	}
	mapData.StartPositions = startPositions

	// Parse terrain data
	if err := ml.parseTerrainData(reader, mapData); err != nil {
		return nil, fmt.Errorf("failed to parse terrain data: %w", err)
	}

	return mapData, nil
}

// parseHeader reads and parses the map file header
func (ml *MapLoader) parseHeader(reader io.Reader) (*MapFileHeader, error) {
	header := &MapFileHeader{}

	// Read header in little-endian format (MegaGlest standard)
	if err := binary.Read(reader, binary.LittleEndian, header); err != nil {
		return nil, fmt.Errorf("failed to read map header: %w", err)
	}

	return header, nil
}

// parseHeaderV2 parses the extended header for version 2 maps
func (ml *MapLoader) parseHeaderV2(reader io.Reader, descData []byte) (*MapFileHeaderV2, error) {
	v2Header := &MapFileHeaderV2{}

	// The description field contains the v2 header data
	buf := bytes.NewReader(descData)
	if err := binary.Read(buf, binary.LittleEndian, v2Header); err != nil {
		return nil, fmt.Errorf("failed to read version 2 header: %w", err)
	}

	// Validate magic number
	if v2Header.Magic != MapVersion2Magic {
		return nil, fmt.Errorf("invalid version 2 magic number: expected %x, got %x",
			MapVersion2Magic, v2Header.Magic)
	}

	return v2Header, nil
}

// parseStartPositions reads player starting positions
func (ml *MapLoader) parseStartPositions(reader io.Reader, maxPlayers int) ([]Vector2i, error) {
	positions := make([]Vector2i, maxPlayers)

	for i := 0; i < maxPlayers; i++ {
		var pos StartPosition
		if err := binary.Read(reader, binary.LittleEndian, &pos); err != nil {
			return nil, fmt.Errorf("failed to read start position %d: %w", i, err)
		}
		positions[i] = Vector2i{X: int(pos.X), Y: int(pos.Y)}
	}

	return positions, nil
}

// parseTerrainData reads height map, surface map, and object map data
func (ml *MapLoader) parseTerrainData(reader io.Reader, mapData *Map) error {
	width, height := mapData.Width, mapData.Height

	// Initialize 2D arrays
	mapData.HeightMap = make([][]float32, height)
	mapData.SurfaceMap = make([][]int8, height)
	mapData.ObjectMap = make([][]int8, height)

	for y := 0; y < height; y++ {
		mapData.HeightMap[y] = make([]float32, width)
		mapData.SurfaceMap[y] = make([]int8, width)
		mapData.ObjectMap[y] = make([]int8, width)
	}

	// Read height data (width * height * float32)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			var height float32
			if err := binary.Read(reader, binary.LittleEndian, &height); err != nil {
				return fmt.Errorf("failed to read height at (%d,%d): %w", x, y, err)
			}
			mapData.HeightMap[y][x] = height
		}
	}

	// Read surface data (width * height * int8)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			var surface int8
			if err := binary.Read(reader, binary.LittleEndian, &surface); err != nil {
				return fmt.Errorf("failed to read surface at (%d,%d): %w", x, y, err)
			}
			mapData.SurfaceMap[y][x] = surface
		}
	}

	// Read object data (width * height * int8)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			var object int8
			if err := binary.Read(reader, binary.LittleEndian, &object); err != nil {
				return fmt.Errorf("failed to read object at (%d,%d): %w", x, y, err)
			}
			mapData.ObjectMap[y][x] = object
		}
	}

	return nil
}

// validateHeader performs validation on the parsed header
func (ml *MapLoader) validateHeader(header *MapFileHeader) error {
	// Check version
	if header.Version != int32(MapVersionGBM) && header.Version != int32(MapVersionMGM) {
		return fmt.Errorf("unsupported map version: %d", header.Version)
	}

	// Check dimensions
	if header.Width < MinMapCellDimension || header.Width > MaxMapCellDimension {
		return fmt.Errorf("invalid map width: %d (must be %d-%d)",
			header.Width, MinMapCellDimension, MaxMapCellDimension)
	}

	if header.Height < MinMapCellDimension || header.Height > MaxMapCellDimension {
		return fmt.Errorf("invalid map height: %d (must be %d-%d)",
			header.Height, MinMapCellDimension, MaxMapCellDimension)
	}

	// Check player count
	if header.MaxFactions < MinMapFactionCount || header.MaxFactions > MaxMapFactionCount {
		return fmt.Errorf("invalid max factions: %d (must be %d-%d)",
			header.MaxFactions, MinMapFactionCount, MaxMapFactionCount)
	}

	return nil
}

// extractString extracts a null-terminated string from a byte array
func (ml *MapLoader) extractString(data []byte) string {
	// Find null terminator
	nullIndex := bytes.IndexByte(data, 0)
	if nullIndex == -1 {
		nullIndex = len(data)
	}

	// Convert to string and trim whitespace
	return strings.TrimSpace(string(data[:nullIndex]))
}

// determineTilesetName determines the tileset name for the map
func (ml *MapLoader) determineTilesetName(mapData *Map) string {
	// For now, use a simple heuristic based on map name or default to meadow
	// This could be enhanced to read from map metadata or user preference

	mapName := strings.ToLower(mapData.Title)

	// Common tileset mappings based on map name patterns
	if strings.Contains(mapName, "desert") {
		return "desert2"
	}
	if strings.Contains(mapName, "winter") || strings.Contains(mapName, "snow") {
		return "winter"
	}
	if strings.Contains(mapName, "forest") || strings.Contains(mapName, "wood") {
		return "forest"
	}
	if strings.Contains(mapName, "jungle") {
		return "jungle"
	}
	if strings.Contains(mapName, "hell") || strings.Contains(mapName, "lava") {
		return "hell"
	}

	// Default to meadow tileset
	return "meadow"
}

// GetHeightAt returns the terrain height at the specified coordinates
func (m *Map) GetHeightAt(x, y int) float32 {
	if y >= 0 && y < len(m.HeightMap) && x >= 0 && x < len(m.HeightMap[y]) {
		return m.HeightMap[y][x]
	}
	return 0.0
}

// GetSurfaceAt returns the surface type at the specified coordinates
func (m *Map) GetSurfaceAt(x, y int) MapSurfaceType {
	if y >= 0 && y < len(m.SurfaceMap) && x >= 0 && x < len(m.SurfaceMap[y]) {
		return MapSurfaceType(m.SurfaceMap[y][x])
	}
	return SurfaceGrass // Default to grass
}

// GetObjectAt returns the object type at the specified coordinates
func (m *Map) GetObjectAt(x, y int) int8 {
	if y >= 0 && y < len(m.ObjectMap) && x >= 0 && x < len(m.ObjectMap[y]) {
		return m.ObjectMap[y][x]
	}
	return 0 // No object
}

// IsValidPosition checks if the given coordinates are within map bounds
func (m *Map) IsValidPosition(x, y int) bool {
	return x >= 0 && x < m.Width && y >= 0 && y < m.Height
}

// GetPlayerStartPosition returns the starting position for the specified player (0-indexed)
func (m *Map) GetPlayerStartPosition(playerIndex int) (Vector2i, bool) {
	if playerIndex >= 0 && playerIndex < len(m.StartPositions) {
		return m.StartPositions[playerIndex], true
	}
	return Vector2i{}, false
}

// PrintSummary prints a summary of the map information
func (m *Map) PrintSummary() {
	fmt.Printf("Map: %s\n", m.Title)
	fmt.Printf("  Author: %s\n", m.Author)
	fmt.Printf("  Description: %s\n", m.Description)
	fmt.Printf("  Dimensions: %dx%d\n", m.Width, m.Height)
	fmt.Printf("  Max Players: %d\n", m.MaxPlayers)
	fmt.Printf("  Version: %d\n", int(m.Version))
	fmt.Printf("  Tileset: %s\n", m.TilesetName)
	fmt.Printf("  Water Level: %.1f\n", m.WaterLevel)
	fmt.Printf("  Height Factor: %.1f\n", m.HeightFactor)
	if m.Version == MapVersionMGM {
		fmt.Printf("  Cliff Level: %.1f\n", m.CliffLevel)
		fmt.Printf("  Camera Height: %.1f\n", m.CameraHeight)
	}

	fmt.Printf("  Start Positions:\n")
	for i, pos := range m.StartPositions {
		fmt.Printf("    Player %d: (%d, %d)\n", i+1, pos.X, pos.Y)
	}
}