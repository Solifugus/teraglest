package data

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"teraglest/pkg/formats"
)

// AssetType represents different types of game assets
type AssetType string

const (
	AssetTypeXML     AssetType = "xml"
	AssetTypeG3D     AssetType = "g3d"
	AssetTypeTexture AssetType = "texture"
	AssetTypeAudio   AssetType = "audio"
)

// AssetManager handles loading, caching, and managing all game assets
type AssetManager struct {
	cache        *AssetCache
	techTreeRoot string    // Root path for tech tree assets
	mutex        sync.Mutex // For thread-safe operations

	// Preloaded common data
	techTree  *TechTree
	resources []ResourceDefinition
	factions  []FactionDefinition
}

// NewAssetManager creates a new asset manager with the specified tech tree root
func NewAssetManager(techTreeRoot string) *AssetManager {
	return &AssetManager{
		cache:        NewAssetCache(512, 1000), // 512MB cache, max 1000 entries
		techTreeRoot: techTreeRoot,
	}
}

// LoadTechTree loads and caches the main tech tree data
func (am *AssetManager) LoadTechTree() (*TechTree, error) {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	if am.techTree != nil {
		return am.techTree, nil
	}

	techTreePath := filepath.Join(am.techTreeRoot, "megapack.xml")

	// Check cache first
	if cached, found := am.cache.Get(techTreePath); found {
		am.techTree = cached.(*TechTree)
		return am.techTree, nil
	}

	// Load from file
	techTree, err := LoadTechTree(techTreePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load tech tree: %w", err)
	}

	// Cache the result
	fileInfo, _ := os.Stat(techTreePath)
	size := int64(0)
	if fileInfo != nil {
		size = fileInfo.Size()
	}

	err = am.cache.Put(techTreePath, techTree, string(AssetTypeXML), size)
	if err != nil {
		fmt.Printf("Warning: Failed to cache tech tree: %v\n", err)
	}

	am.techTree = techTree
	return techTree, nil
}

// LoadResources loads and caches all resource definitions
func (am *AssetManager) LoadResources() ([]ResourceDefinition, error) {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	if am.resources != nil {
		return am.resources, nil
	}

	resourcesPath := filepath.Join(am.techTreeRoot, "resources")

	// Check cache first
	if cached, found := am.cache.Get(resourcesPath); found {
		am.resources = cached.([]ResourceDefinition)
		return am.resources, nil
	}

	// Load from files
	resources, err := LoadAllResources(resourcesPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load resources: %w", err)
	}

	// Cache the result (estimate size)
	size := int64(len(resources) * 1024) // Rough estimate
	err = am.cache.Put(resourcesPath, resources, string(AssetTypeXML), size)
	if err != nil {
		fmt.Printf("Warning: Failed to cache resources: %v\n", err)
	}

	am.resources = resources
	return resources, nil
}

// LoadFactions loads and caches all faction definitions
func (am *AssetManager) LoadFactions() ([]FactionDefinition, error) {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	if am.factions != nil {
		return am.factions, nil
	}

	factionsPath := filepath.Join(am.techTreeRoot, "factions")

	// Check cache first
	if cached, found := am.cache.Get(factionsPath); found {
		am.factions = cached.([]FactionDefinition)
		return am.factions, nil
	}

	// Load from files
	factions, err := LoadAllFactions(factionsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load factions: %w", err)
	}

	// Cache the result
	size := int64(len(factions) * 2048) // Rough estimate
	err = am.cache.Put(factionsPath, factions, string(AssetTypeXML), size)
	if err != nil {
		fmt.Printf("Warning: Failed to cache factions: %v\n", err)
	}

	am.factions = factions
	return factions, nil
}

// LoadUnit loads and caches a specific unit definition
func (am *AssetManager) LoadUnit(factionName, unitName string) (*UnitDefinition, error) {
	unitPath := filepath.Join(am.techTreeRoot, "factions", factionName, "units", unitName, unitName+".xml")

	// Check cache first
	if cached, found := am.cache.Get(unitPath); found {
		return cached.(*UnitDefinition), nil
	}

	// Load from file
	unit, err := LoadUnit(unitPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load unit %s/%s: %w", factionName, unitName, err)
	}

	unitDef := &UnitDefinition{
		Name: unitName,
		Unit: *unit,
	}

	// Cache the result
	fileInfo, _ := os.Stat(unitPath)
	size := int64(4096) // Estimate for unit XML
	if fileInfo != nil {
		size = fileInfo.Size()
	}

	err = am.cache.Put(unitPath, unitDef, string(AssetTypeXML), size)
	if err != nil {
		fmt.Printf("Warning: Failed to cache unit %s: %v\n", unitName, err)
	}

	return unitDef, nil
}

// LoadG3DModel loads and caches a G3D model
func (am *AssetManager) LoadG3DModel(modelPath string) (*formats.G3DModel, error) {
	// Resolve relative path
	fullPath := am.resolvePath(modelPath)

	// Check cache first
	if cached, found := am.cache.Get(fullPath); found {
		return cached.(*formats.G3DModel), nil
	}

	// Load from file
	model, err := formats.LoadG3D(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load G3D model %s: %w", modelPath, err)
	}

	// Cache the result
	fileInfo, _ := os.Stat(fullPath)
	size := int64(0)
	if fileInfo != nil {
		size = fileInfo.Size()
	}

	err = am.cache.Put(fullPath, model, string(AssetTypeG3D), size)
	if err != nil {
		fmt.Printf("Warning: Failed to cache G3D model %s: %v\n", modelPath, err)
	}

	return model, nil
}

// LoadTexture loads and caches a texture image
func (am *AssetManager) LoadTexture(texturePath string) (image.Image, error) {
	// Resolve relative path
	fullPath := am.resolvePath(texturePath)

	// Check cache first
	if cached, found := am.cache.Get(fullPath); found {
		return cached.(image.Image), nil
	}

	// Load from file
	file, err := os.Open(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open texture %s: %w", texturePath, err)
	}
	defer file.Close()

	var img image.Image
	ext := strings.ToLower(filepath.Ext(fullPath))

	switch ext {
	case ".png":
		img, err = png.Decode(file)
	case ".jpg", ".jpeg":
		img, err = jpeg.Decode(file)
	default:
		return nil, fmt.Errorf("unsupported texture format: %s", ext)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to decode texture %s: %w", texturePath, err)
	}

	// Cache the result
	fileInfo, _ := os.Stat(fullPath)
	size := int64(0)
	if fileInfo != nil {
		size = fileInfo.Size()
	}

	err = am.cache.Put(fullPath, img, string(AssetTypeTexture), size)
	if err != nil {
		fmt.Printf("Warning: Failed to cache texture %s: %v\n", texturePath, err)
	}

	return img, nil
}

// LoadAudio loads and caches audio data (placeholder for now)
func (am *AssetManager) LoadAudio(audioPath string) ([]byte, error) {
	// Resolve relative path
	fullPath := am.resolvePath(audioPath)

	// Check cache first
	if cached, found := am.cache.Get(fullPath); found {
		return cached.([]byte), nil
	}

	// Load from file (raw bytes for now)
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load audio %s: %w", audioPath, err)
	}

	// Cache the result
	size := int64(len(data))
	err = am.cache.Put(fullPath, data, string(AssetTypeAudio), size)
	if err != nil {
		fmt.Printf("Warning: Failed to cache audio %s: %v\n", audioPath, err)
	}

	return data, nil
}

// LoadFactionComplete loads a complete faction with all its units and their models
func (am *AssetManager) LoadFactionComplete(factionName string) (*FactionCompleteData, error) {
	// Load faction definition
	factions, err := am.LoadFactions()
	if err != nil {
		return nil, err
	}

	var faction *FactionDefinition
	for _, f := range factions {
		if f.Name == factionName {
			faction = &f
			break
		}
	}

	if faction == nil {
		return nil, fmt.Errorf("faction %s not found", factionName)
	}

	result := &FactionCompleteData{
		Faction: *faction,
		Units:   make(map[string]*UnitDefinition),
		Models:  make(map[string]*formats.G3DModel),
	}

	// Load all units for this faction
	unitsDir := filepath.Join(am.techTreeRoot, "factions", factionName, "units")
	entries, err := os.ReadDir(unitsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read units directory for faction %s: %w", factionName, err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		unitName := entry.Name()

		// Load unit definition
		unit, err := am.LoadUnit(factionName, unitName)
		if err != nil {
			fmt.Printf("Warning: Failed to load unit %s: %v\n", unitName, err)
			continue
		}

		result.Units[unitName] = unit

		// Load unit models (from skills animations)
		modelsDir := filepath.Join(unitsDir, unitName, "models")
		if _, err := os.Stat(modelsDir); err == nil {
			modelEntries, err := os.ReadDir(modelsDir)
			if err == nil {
				for _, modelEntry := range modelEntries {
					if strings.HasSuffix(modelEntry.Name(), ".g3d") {
						modelPath := filepath.Join("factions", factionName, "units", unitName, "models", modelEntry.Name())
						model, err := am.LoadG3DModel(modelPath)
						if err != nil {
							fmt.Printf("Warning: Failed to load model %s: %v\n", modelPath, err)
							continue
						}
						result.Models[modelEntry.Name()] = model
					}
				}
			}
		}
	}

	return result, nil
}

// FactionCompleteData represents a fully loaded faction with all its assets
type FactionCompleteData struct {
	Faction FactionDefinition                  // Faction definition
	Units   map[string]*UnitDefinition         // All unit definitions
	Models  map[string]*formats.G3DModel       // All 3D models
}

// resolvePath resolves a relative asset path to an absolute path
func (am *AssetManager) resolvePath(assetPath string) string {
	if filepath.IsAbs(assetPath) {
		return assetPath
	}
	return filepath.Join(am.techTreeRoot, assetPath)
}

// GetCacheStats returns current cache statistics
func (am *AssetManager) GetCacheStats() CacheStats {
	return am.cache.GetStats()
}

// PrintCacheStats prints cache statistics for debugging
func (am *AssetManager) PrintCacheStats() {
	am.cache.PrintStats()
}

// ClearCache clears all cached assets
func (am *AssetManager) ClearCache() {
	am.cache.Clear()
	am.mutex.Lock()
	am.techTree = nil
	am.resources = nil
	am.factions = nil
	am.mutex.Unlock()
}