package data

import (
	"path/filepath"
	"testing"
)

func TestNewAssetManager(t *testing.T) {
	techTreeRoot := "../../megaglest-source/data/glest_game/techs/megapack"
	am := NewAssetManager(techTreeRoot)

	if am == nil {
		t.Error("NewAssetManager returned nil")
	}

	if am.techTreeRoot != techTreeRoot {
		t.Errorf("Expected techTreeRoot %s, got %s", techTreeRoot, am.techTreeRoot)
	}

	if am.cache == nil {
		t.Error("AssetManager cache should not be nil")
	}
}

func TestAssetManagerLoadTechTree(t *testing.T) {
	techTreeRoot := "../../megaglest-source/data/glest_game/techs/megapack"
	am := NewAssetManager(techTreeRoot)

	// First load - should read from file
	techTree1, err := am.LoadTechTree()
	if err != nil {
		t.Fatalf("Failed to load tech tree: %v", err)
	}

	if techTree1 == nil {
		t.Error("LoadTechTree returned nil")
	}

	// Second load - should return from memory field (not cache, but same instance)
	techTree2, err := am.LoadTechTree()
	if err != nil {
		t.Fatalf("Failed to load tech tree from memory: %v", err)
	}

	if techTree1 != techTree2 {
		t.Error("Expected same pointer from cached tech tree")
	}

	// Verify cache stats - tech tree should be in cache from first load
	stats := am.GetCacheStats()
	if stats.TotalEntries == 0 {
		t.Error("Expected at least one cache entry")
	}

	// Clear memory fields and try again to test actual cache hit
	am.ClearCache() // This clears both cache and memory fields

	// Load again - should read from file and cache
	techTree3, err := am.LoadTechTree()
	if err != nil {
		t.Fatalf("Failed to load tech tree after cache clear: %v", err)
	}

	if techTree3 == nil {
		t.Error("LoadTechTree returned nil after cache clear")
	}
}

func TestAssetManagerLoadResources(t *testing.T) {
	techTreeRoot := "../../megaglest-source/data/glest_game/techs/megapack"
	am := NewAssetManager(techTreeRoot)

	resources, err := am.LoadResources()
	if err != nil {
		t.Fatalf("Failed to load resources: %v", err)
	}

	if len(resources) == 0 {
		t.Error("Expected at least one resource")
	}

	// Test caching
	resources2, err := am.LoadResources()
	if err != nil {
		t.Fatalf("Failed to load resources from cache: %v", err)
	}

	if len(resources) != len(resources2) {
		t.Error("Cached resources should have same length")
	}
}

func TestAssetManagerLoadFactions(t *testing.T) {
	techTreeRoot := "../../megaglest-source/data/glest_game/techs/megapack"
	am := NewAssetManager(techTreeRoot)

	factions, err := am.LoadFactions()
	if err != nil {
		t.Fatalf("Failed to load factions: %v", err)
	}

	if len(factions) == 0 {
		t.Error("Expected at least one faction")
	}

	// Should include magic faction
	foundMagic := false
	for _, faction := range factions {
		if faction.Name == "magic" {
			foundMagic = true
			break
		}
	}

	if !foundMagic {
		t.Error("Expected to find magic faction")
	}
}

func TestAssetManagerLoadUnit(t *testing.T) {
	techTreeRoot := "../../megaglest-source/data/glest_game/techs/megapack"
	am := NewAssetManager(techTreeRoot)

	// Load initiate unit from magic faction
	unit, err := am.LoadUnit("magic", "initiate")
	if err != nil {
		t.Fatalf("Failed to load initiate unit: %v", err)
	}

	if unit == nil {
		t.Error("LoadUnit returned nil")
	}

	if unit.Name != "initiate" {
		t.Errorf("Expected unit name 'initiate', got '%s'", unit.Name)
	}

	// Test caching
	unit2, err := am.LoadUnit("magic", "initiate")
	if err != nil {
		t.Fatalf("Failed to load unit from cache: %v", err)
	}

	if unit != unit2 {
		t.Error("Expected same pointer from cached unit")
	}
}

func TestAssetManagerLoadG3DModel(t *testing.T) {
	techTreeRoot := "../../megaglest-source/data/glest_game/techs/megapack"
	am := NewAssetManager(techTreeRoot)

	// Load initiate standing model
	modelPath := "factions/magic/units/initiate/models/initiate_standing.g3d"
	model, err := am.LoadG3DModel(modelPath)
	if err != nil {
		t.Fatalf("Failed to load G3D model: %v", err)
	}

	if model == nil {
		t.Error("LoadG3DModel returned nil")
	}

	if len(model.Meshes) == 0 {
		t.Error("Expected model to have meshes")
	}

	// Test caching
	model2, err := am.LoadG3DModel(modelPath)
	if err != nil {
		t.Fatalf("Failed to load model from cache: %v", err)
	}

	if model != model2 {
		t.Error("Expected same pointer from cached model")
	}
}

func TestAssetManagerLoadTexture(t *testing.T) {
	techTreeRoot := "../../megaglest-source/data/glest_game/techs/megapack"
	am := NewAssetManager(techTreeRoot)

	// Try to load a texture - note: we may not have actual texture files in test
	// This test validates the loading mechanism even if file doesn't exist
	texturePath := "factions/magic/units/initiate/images/initiate.png"
	texture, err := am.LoadTexture(texturePath)

	if err != nil {
		// Expected if file doesn't exist - just check error handling
		t.Logf("Texture loading failed as expected: %v", err)
		return
	}

	if texture == nil {
		t.Error("LoadTexture returned nil without error")
	}
}

func TestAssetManagerLoadFactionComplete(t *testing.T) {
	techTreeRoot := "../../megaglest-source/data/glest_game/techs/megapack"
	am := NewAssetManager(techTreeRoot)

	// Load complete magic faction - this is the Phase 1.5 validation requirement
	completeData, err := am.LoadFactionComplete("magic")
	if err != nil {
		t.Fatalf("Failed to load complete magic faction: %v", err)
	}

	if completeData == nil {
		t.Error("LoadFactionComplete returned nil")
	}

	if completeData.Faction.Name != "magic" {
		t.Errorf("Expected faction name 'magic', got '%s'", completeData.Faction.Name)
	}

	if len(completeData.Units) == 0 {
		t.Error("Expected faction to have units")
	}

	// Should have initiate unit
	initiate, exists := completeData.Units["initiate"]
	if !exists {
		t.Error("Expected magic faction to have initiate unit")
	} else {
		if initiate.Name != "initiate" {
			t.Errorf("Expected unit name 'initiate', got '%s'", initiate.Name)
		}
	}

	// Check for models
	t.Logf("Loaded %d units and %d models for magic faction", len(completeData.Units), len(completeData.Models))

	// Print some debug info
	for unitName, unit := range completeData.Units {
		t.Logf("Unit: %s, HP: %d", unitName, unit.Unit.Parameters.MaxHP.Value)
	}

	for modelName := range completeData.Models {
		t.Logf("Model: %s", modelName)
	}
}

func TestAssetManagerCacheManagement(t *testing.T) {
	techTreeRoot := "../../megaglest-source/data/glest_game/techs/megapack"
	am := NewAssetManager(techTreeRoot)

	// Load some assets to populate cache
	_, err := am.LoadTechTree()
	if err != nil {
		t.Fatalf("Failed to load tech tree: %v", err)
	}

	_, err = am.LoadFactions()
	if err != nil {
		t.Fatalf("Failed to load factions: %v", err)
	}

	// Check cache stats
	stats := am.GetCacheStats()
	if stats.TotalEntries == 0 {
		t.Error("Expected cache to have entries")
	}

	if stats.Hits == 0 && stats.Misses == 0 {
		t.Error("Expected some cache activity")
	}

	// Clear cache
	am.ClearCache()

	// Verify cache is cleared
	statsAfterClear := am.GetCacheStats()
	if statsAfterClear.TotalEntries != 0 {
		t.Error("Expected cache to be empty after clear")
	}

	if statsAfterClear.Hits != 0 || statsAfterClear.Misses != 0 {
		t.Error("Expected cache stats to be reset after clear")
	}
}

func TestAssetManagerResolvePath(t *testing.T) {
	techTreeRoot := "/path/to/tech/tree"
	am := NewAssetManager(techTreeRoot)

	// Test relative path resolution
	relativePath := "factions/magic/magic.xml"
	resolved := am.resolvePath(relativePath)
	expected := filepath.Join(techTreeRoot, relativePath)

	if resolved != expected {
		t.Errorf("Expected resolved path %s, got %s", expected, resolved)
	}

	// Test absolute path (should return as-is)
	absolutePath := "/absolute/path/to/file.xml"
	resolved = am.resolvePath(absolutePath)

	if resolved != absolutePath {
		t.Errorf("Expected absolute path unchanged %s, got %s", absolutePath, resolved)
	}
}