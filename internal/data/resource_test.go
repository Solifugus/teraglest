package data

import (
	"testing"
)

func TestResourceParsing(t *testing.T) {
	// Test loading individual resource
	goldResource, err := LoadResource("../../megaglest-source/data/glest_game/techs/megapack/resources/gold/gold.xml")
	if err != nil {
		t.Fatalf("Failed to load gold resource: %v", err)
	}

	// Verify gold resource properties
	if goldResource.Image.Path != "images/gold.bmp" {
		t.Errorf("Expected gold image path 'images/gold.bmp', got '%s'", goldResource.Image.Path)
	}

	if goldResource.Type.Value != "tech" {
		t.Errorf("Expected gold type 'tech', got '%s'", goldResource.Type.Value)
	}

	if goldResource.Type.Model == nil {
		t.Error("Expected gold to have a model, but got nil")
	} else if goldResource.Type.Model.Path != "models/gold.g3d" {
		t.Errorf("Expected gold model path 'models/gold.g3d', got '%s'", goldResource.Type.Model.Path)
	}

	if goldResource.Type.DefaultAmount == nil {
		t.Error("Expected gold to have a default amount, but got nil")
	} else if goldResource.Type.DefaultAmount.Value != 2000 {
		t.Errorf("Expected gold default amount 2000, got %d", goldResource.Type.DefaultAmount.Value)
	}

	if goldResource.Type.ResourceNumber == nil {
		t.Error("Expected gold to have a resource number, but got nil")
	} else if goldResource.Type.ResourceNumber.Value != 1 {
		t.Errorf("Expected gold resource number 1, got %d", goldResource.Type.ResourceNumber.Value)
	}

	// Test loading energy resource (static type with no model)
	energyResource, err := LoadResource("../../megaglest-source/data/glest_game/techs/megapack/resources/energy/energy.xml")
	if err != nil {
		t.Fatalf("Failed to load energy resource: %v", err)
	}

	if energyResource.Type.Value != "static" {
		t.Errorf("Expected energy type 'static', got '%s'", energyResource.Type.Value)
	}

	if energyResource.Type.Model != nil {
		t.Error("Expected energy to have no model, but got one")
	}
}

func TestLoadAllResources(t *testing.T) {
	// Test loading all resources from megapack
	resources, err := LoadAllResources("../../megaglest-source/data/glest_game/techs/megapack/resources")
	if err != nil {
		t.Fatalf("Failed to load all resources: %v", err)
	}

	// Should have at least 6 basic resources
	if len(resources) < 6 {
		t.Errorf("Expected at least 6 resources, got %d", len(resources))
	}

	// Check that gold resource exists
	goldRes := GetResourceByName(resources, "gold")
	if goldRes == nil {
		t.Error("Expected to find gold resource")
	} else {
		if !goldRes.IsTechResource() {
			t.Error("Expected gold to be a tech resource")
		}
		if goldRes.IsStaticResource() {
			t.Error("Expected gold not to be a static resource")
		}
	}

	// Check that energy resource exists
	energyRes := GetResourceByName(resources, "energy")
	if energyRes == nil {
		t.Error("Expected to find energy resource")
	} else {
		if !energyRes.IsStaticResource() {
			t.Error("Expected energy to be a static resource")
		}
		if energyRes.IsTechResource() {
			t.Error("Expected energy not to be a tech resource")
		}
	}

	// Check that food resource exists and is consumable
	foodRes := GetResourceByName(resources, "food")
	if foodRes == nil {
		t.Error("Expected to find food resource")
	} else {
		if !foodRes.IsConsumableResource() {
			t.Error("Expected food to be a consumable resource")
		}
	}
}

func TestGetResourceByName(t *testing.T) {
	resources := []ResourceDefinition{
		{Name: "gold", Resource: Resource{}},
		{Name: "wood", Resource: Resource{}},
	}

	// Test finding existing resource
	goldRes := GetResourceByName(resources, "gold")
	if goldRes == nil {
		t.Error("Expected to find gold resource")
	} else if goldRes.Name != "gold" {
		t.Errorf("Expected resource name 'gold', got '%s'", goldRes.Name)
	}

	// Test finding non-existent resource
	nonExistentRes := GetResourceByName(resources, "nonexistent")
	if nonExistentRes != nil {
		t.Error("Expected not to find non-existent resource")
	}
}