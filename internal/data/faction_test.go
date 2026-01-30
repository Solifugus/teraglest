package data

import (
	"testing"
)

func TestFactionParsing(t *testing.T) {
	// Test loading the magic faction
	faction, err := LoadFaction("../../megaglest-source/data/glest_game/techs/megapack/factions/magic/magic.xml")
	if err != nil {
		t.Fatalf("Failed to load magic faction: %v", err)
	}

	// Verify starting resources
	expectedResources := map[string]int{
		"gold":   500,
		"wood":   400,
		"stone":  400,
		"energy": 1,
	}

	if len(faction.StartingResources) != len(expectedResources) {
		t.Errorf("Expected %d starting resources, got %d", len(expectedResources), len(faction.StartingResources))
	}

	for _, res := range faction.StartingResources {
		expected, exists := expectedResources[res.Name]
		if !exists {
			t.Errorf("Unexpected starting resource: %s", res.Name)
			continue
		}
		if res.Amount != expected {
			t.Errorf("Expected %s amount %d, got %d", res.Name, expected, res.Amount)
		}
	}

	// Verify starting units
	expectedUnits := map[string]int{
		"mage_tower":    1,
		"energy_source": 1,
		"initiate":      3,
		"battlemage":    1,
		"summoner":      1,
		"daemon":        1,
		"golem":         1,
	}

	if len(faction.StartingUnits) != len(expectedUnits) {
		t.Errorf("Expected %d starting units, got %d", len(expectedUnits), len(faction.StartingUnits))
	}

	for _, unit := range faction.StartingUnits {
		expected, exists := expectedUnits[unit.Name]
		if !exists {
			t.Errorf("Unexpected starting unit: %s", unit.Name)
			continue
		}
		if unit.Amount != expected {
			t.Errorf("Expected %s amount %d, got %d", unit.Name, expected, unit.Amount)
		}
	}

	// Verify music configuration
	if faction.Music == nil {
		t.Error("Expected faction to have music configuration")
	} else {
		if !faction.Music.Value {
			t.Error("Expected music to be enabled")
		}
		if faction.Music.Path != "music/music_magic.ogg" {
			t.Errorf("Expected music path 'music/music_magic.ogg', got '%s'", faction.Music.Path)
		}
	}

	// Verify AI behavior exists
	if faction.AIBehavior == nil {
		t.Error("Expected faction to have AI behavior configuration")
	} else {
		if len(faction.AIBehavior.WorkerUnits) == 0 {
			t.Error("Expected AI to have worker units configured")
		}
		if len(faction.AIBehavior.WarriorUnits) == 0 {
			t.Error("Expected AI to have warrior units configured")
		}
		if len(faction.AIBehavior.Upgrades) == 0 {
			t.Error("Expected AI to have upgrades configured")
		}
	}
}

func TestLoadAllFactions(t *testing.T) {
	// Test loading all factions from megapack
	factions, err := LoadAllFactions("../../megaglest-source/data/glest_game/techs/megapack/factions")
	if err != nil {
		t.Fatalf("Failed to load all factions: %v", err)
	}

	// Should have 7 factions in megapack
	if len(factions) != 7 {
		t.Errorf("Expected 7 factions, got %d", len(factions))
	}

	// Check that magic faction exists
	magicFaction := GetFactionByName(factions, "magic")
	if magicFaction == nil {
		t.Error("Expected to find magic faction")
	} else {
		// Test helper functions
		goldAmount := magicFaction.GetStartingResource("gold")
		if goldAmount != 500 {
			t.Errorf("Expected magic faction to start with 500 gold, got %d", goldAmount)
		}

		initiateCount := magicFaction.GetStartingUnit("initiate")
		if initiateCount != 3 {
			t.Errorf("Expected magic faction to start with 3 initiates, got %d", initiateCount)
		}

		if !magicFaction.HasMusic() {
			t.Error("Expected magic faction to have music")
		}

		musicPath := magicFaction.GetMusicPath()
		if musicPath != "music/music_magic.ogg" {
			t.Errorf("Expected music path 'music/music_magic.ogg', got '%s'", musicPath)
		}
	}

	// Verify all factions have basic required elements
	for _, faction := range factions {
		if len(faction.Faction.StartingResources) == 0 {
			t.Errorf("Faction %s has no starting resources", faction.Name)
		}
		if len(faction.Faction.StartingUnits) == 0 {
			t.Errorf("Faction %s has no starting units", faction.Name)
		}
	}
}

func TestGetFactionByName(t *testing.T) {
	factions := []FactionDefinition{
		{Name: "magic", Faction: Faction{}},
		{Name: "tech", Faction: Faction{}},
	}

	// Test finding existing faction
	magicFaction := GetFactionByName(factions, "magic")
	if magicFaction == nil {
		t.Error("Expected to find magic faction")
	} else if magicFaction.Name != "magic" {
		t.Errorf("Expected faction name 'magic', got '%s'", magicFaction.Name)
	}

	// Test finding non-existent faction
	nonExistentFaction := GetFactionByName(factions, "nonexistent")
	if nonExistentFaction != nil {
		t.Error("Expected not to find non-existent faction")
	}
}