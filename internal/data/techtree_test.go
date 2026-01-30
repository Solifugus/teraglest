package data

import (
	"testing"
)

func TestTechTreeParsing(t *testing.T) {
	// Test loading the megapack tech tree
	techTree, err := LoadTechTree("../../megaglest-source/data/glest_game/techs/megapack/megapack.xml")
	if err != nil {
		t.Fatalf("Failed to load tech tree: %v", err)
	}

	// Verify description
	if techTree.Description.Value != "magitech tech tree" {
		t.Errorf("Expected description 'magitech tech tree', got '%s'", techTree.Description.Value)
	}

	// Verify attack types
	expectedAttackTypes := []string{"slashing", "piercing", "impact", "energy", "sword", "arrow", "magic", "beat"}
	if len(techTree.AttackTypes) != len(expectedAttackTypes) {
		t.Errorf("Expected %d attack types, got %d", len(expectedAttackTypes), len(techTree.AttackTypes))
	}

	for i, expected := range expectedAttackTypes {
		if i >= len(techTree.AttackTypes) {
			t.Errorf("Missing attack type: %s", expected)
			continue
		}
		if techTree.AttackTypes[i].Name != expected {
			t.Errorf("Expected attack type '%s' at position %d, got '%s'", expected, i, techTree.AttackTypes[i].Name)
		}
	}

	// Verify armor types
	expectedArmorTypes := []string{"organic", "leather", "wood", "metal", "stone"}
	if len(techTree.ArmorTypes) != len(expectedArmorTypes) {
		t.Errorf("Expected %d armor types, got %d", len(expectedArmorTypes), len(techTree.ArmorTypes))
	}

	for i, expected := range expectedArmorTypes {
		if i >= len(techTree.ArmorTypes) {
			t.Errorf("Missing armor type: %s", expected)
			continue
		}
		if techTree.ArmorTypes[i].Name != expected {
			t.Errorf("Expected armor type '%s' at position %d, got '%s'", expected, i, techTree.ArmorTypes[i].Name)
		}
	}

	// Test damage multiplier functionality
	arrowVsOrganic := techTree.GetDamageMultiplier("arrow", "organic")
	if arrowVsOrganic != 1.25 {
		t.Errorf("Expected arrow vs organic multiplier 1.25, got %f", arrowVsOrganic)
	}

	impactVsMetal := techTree.GetDamageMultiplier("impact", "metal")
	if impactVsMetal != 1.5 {
		t.Errorf("Expected impact vs metal multiplier 1.5, got %f", impactVsMetal)
	}

	// Test default multiplier for non-existent combination
	defaultMultiplier := techTree.GetDamageMultiplier("nonexistent", "nonexistent")
	if defaultMultiplier != 1.0 {
		t.Errorf("Expected default multiplier 1.0, got %f", defaultMultiplier)
	}

	// Test attack type existence
	if !techTree.HasAttackType("sword") {
		t.Error("Expected HasAttackType('sword') to return true")
	}

	if techTree.HasAttackType("nonexistent") {
		t.Error("Expected HasAttackType('nonexistent') to return false")
	}

	// Test armor type existence
	if !techTree.HasArmorType("metal") {
		t.Error("Expected HasArmorType('metal') to return true")
	}

	if techTree.HasArmorType("nonexistent") {
		t.Error("Expected HasArmorType('nonexistent') to return false")
	}
}