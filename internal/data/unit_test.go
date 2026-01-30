package data

import (
	"testing"
)

func TestUnitParsing(t *testing.T) {
	// Test loading the initiate unit
	unit, err := LoadUnit("../../megaglest-source/data/glest_game/techs/megapack/factions/magic/units/initiate/initiate.xml")
	if err != nil {
		t.Fatalf("Failed to load initiate unit: %v", err)
	}

	// Verify basic parameters
	if unit.Parameters.Size.Value != 1 {
		t.Errorf("Expected size 1, got %d", unit.Parameters.Size.Value)
	}

	if unit.Parameters.Height.Value != 2 {
		t.Errorf("Expected height 2, got %d", unit.Parameters.Height.Value)
	}

	if unit.Parameters.MaxHP.Value != 450 {
		t.Errorf("Expected max HP 450, got %d", unit.Parameters.MaxHP.Value)
	}

	if unit.Parameters.MaxHP.Regeneration != 1 {
		t.Errorf("Expected HP regeneration 1, got %d", unit.Parameters.MaxHP.Regeneration)
	}

	if unit.Parameters.Armor.Value != 0 {
		t.Errorf("Expected armor 0, got %d", unit.Parameters.Armor.Value)
	}

	if unit.Parameters.ArmorType.Value != "leather" {
		t.Errorf("Expected armor type 'leather', got '%s'", unit.Parameters.ArmorType.Value)
	}

	if unit.Parameters.Sight.Value != 9 {
		t.Errorf("Expected sight 9, got %d", unit.Parameters.Sight.Value)
	}

	// Verify resource requirements (cost)
	expectedCosts := map[string]int{
		"gold":   75,
		"energy": 1,
	}

	if len(unit.Parameters.ResourceRequirements) != len(expectedCosts) {
		t.Errorf("Expected %d resource requirements, got %d", len(expectedCosts), len(unit.Parameters.ResourceRequirements))
	}

	for _, req := range unit.Parameters.ResourceRequirements {
		expected, exists := expectedCosts[req.Name]
		if !exists {
			t.Errorf("Unexpected resource requirement: %s", req.Name)
			continue
		}
		if req.Amount != expected {
			t.Errorf("Expected %s cost %d, got %d", req.Name, expected, req.Amount)
		}
	}

	// Verify skills exist
	if len(unit.Skills) == 0 {
		t.Error("Expected unit to have skills")
	}

	// Find and verify a specific skill (stop_skill)
	var stopSkill *Skill
	for _, skill := range unit.Skills {
		if skill.Name.Value == "stop_skill" {
			stopSkill = &skill
			break
		}
	}

	if stopSkill == nil {
		t.Error("Expected to find 'stop_skill'")
	} else {
		if stopSkill.Type.Value != "stop" {
			t.Errorf("Expected stop skill type 'stop', got '%s'", stopSkill.Type.Value)
		}
		if stopSkill.Animation.Path != "models/initiate_standing.g3d" {
			t.Errorf("Expected stop skill animation 'models/initiate_standing.g3d', got '%s'", stopSkill.Animation.Path)
		}
	}

	// Verify commands exist
	if len(unit.Commands) == 0 {
		t.Error("Expected unit to have commands")
	}

	// Find and verify a specific command (move)
	var moveCommand *Command
	for _, command := range unit.Commands {
		if command.Type.Value == "move" {
			moveCommand = &command
			break
		}
	}

	if moveCommand == nil {
		t.Error("Expected to find move command")
	} else {
		if moveCommand.MoveSkill == nil {
			t.Error("Expected move command to have move skill")
		} else {
			if moveCommand.MoveSkill.Value != "move_skill" {
				t.Errorf("Expected move skill 'move_skill', got '%s'", moveCommand.MoveSkill.Value)
			}
		}
	}

	// Verify fields (terrain types)
	if len(unit.Parameters.Fields) == 0 {
		t.Error("Expected unit to have field types")
	} else {
		landFound := false
		for _, field := range unit.Parameters.Fields {
			if field.Value == "land" {
				landFound = true
				break
			}
		}
		if !landFound {
			t.Error("Expected unit to have 'land' field type")
		}
	}
}

func TestLoadAllUnitsFromFaction(t *testing.T) {
	// Test loading all units from magic faction
	units, err := LoadAllUnitsFromFaction("../../megaglest-source/data/glest_game/techs/megapack/factions/magic/units")
	if err != nil {
		t.Fatalf("Failed to load units from magic faction: %v", err)
	}

	// Magic faction should have multiple units
	if len(units) == 0 {
		t.Error("Expected magic faction to have units")
	}

	// Check that initiate unit exists
	initiateUnit := GetUnitByName(units, "initiate")
	if initiateUnit == nil {
		t.Error("Expected to find initiate unit")
	} else {
		// Test helper functions
		goldCost := initiateUnit.GetResourceCost("gold")
		if goldCost != 75 {
			t.Errorf("Expected initiate gold cost 75, got %d", goldCost)
		}

		energyCost := initiateUnit.GetResourceCost("energy")
		if energyCost != 1 {
			t.Errorf("Expected initiate energy cost 1, got %d", energyCost)
		}

		nonExistentCost := initiateUnit.GetResourceCost("nonexistent")
		if nonExistentCost != 0 {
			t.Errorf("Expected non-existent resource cost 0, got %d", nonExistentCost)
		}

		if !initiateUnit.HasField("land") {
			t.Error("Expected initiate to have land field")
		}

		if initiateUnit.HasField("air") {
			t.Error("Expected initiate not to have air field")
		}

		stopSkill := initiateUnit.GetSkillByName("stop_skill")
		if stopSkill == nil {
			t.Error("Expected initiate to have stop_skill")
		}

		nonExistentSkill := initiateUnit.GetSkillByName("nonexistent")
		if nonExistentSkill != nil {
			t.Error("Expected not to find non-existent skill")
		}
	}

	// Verify all units have basic required elements
	for _, unit := range units {
		if unit.Unit.Parameters.MaxHP.Value <= 0 {
			t.Errorf("Unit %s has invalid HP %d", unit.Name, unit.Unit.Parameters.MaxHP.Value)
		}
		if len(unit.Unit.Skills) == 0 {
			t.Errorf("Unit %s has no skills", unit.Name)
		}
		// Note: Some units like buildings may not have commands, so we don't enforce this
	}
}

func TestGetUnitByName(t *testing.T) {
	units := []UnitDefinition{
		{Name: "initiate", Unit: Unit{}},
		{Name: "battlemage", Unit: Unit{}},
	}

	// Test finding existing unit
	initiateUnit := GetUnitByName(units, "initiate")
	if initiateUnit == nil {
		t.Error("Expected to find initiate unit")
	} else if initiateUnit.Name != "initiate" {
		t.Errorf("Expected unit name 'initiate', got '%s'", initiateUnit.Name)
	}

	// Test finding non-existent unit
	nonExistentUnit := GetUnitByName(units, "nonexistent")
	if nonExistentUnit != nil {
		t.Error("Expected not to find non-existent unit")
	}
}