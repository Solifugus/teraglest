package data

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
)

// Unit represents a complete unit definition from unit.xml files
type Unit struct {
	XMLName    xml.Name     `xml:"unit"`
	Parameters UnitParameters `xml:"parameters"`
	Skills     []Skill      `xml:"skills>skill"`
	Commands   []Command    `xml:"commands>command"`
}

// UnitParameters contains the basic unit stats and configuration
type UnitParameters struct {
	Size                 UnitSize              `xml:"size"`
	Height               UnitHeight            `xml:"height"`
	MaxHP                UnitHP                `xml:"max-hp"`
	MaxEP                *UnitEP               `xml:"max-ep,omitempty"`
	Armor                UnitArmor             `xml:"armor"`
	ArmorType            UnitArmorType         `xml:"armor-type"`
	Sight                UnitSight             `xml:"sight"`
	Time                 UnitTime              `xml:"time"`
	MultiSelection       UnitMultiSelection    `xml:"multi-selection"`
	Cellmap              UnitCellmap           `xml:"cellmap"`
	Fields               []Field               `xml:"fields>field"`
	ResourceRequirements []ResourceRequirement `xml:"resource-requirements>resource"`
	Image                UnitImage             `xml:"image"`
	ImageCancel          UnitImageCancel       `xml:"image-cancel"`
	MeetingPoint         UnitMeetingPoint      `xml:"meeting-point"`
	SelectionSounds      *SoundGroup           `xml:"selection-sounds,omitempty"`
	CommandSounds        *SoundGroup           `xml:"command-sounds,omitempty"`
}

// Unit parameter helper structs for XML parsing
type UnitSize struct {
	Value int `xml:"value,attr"`
}

type UnitHeight struct {
	Value int `xml:"value,attr"`
}

type UnitArmor struct {
	Value int `xml:"value,attr"`
}

type UnitArmorType struct {
	Value string `xml:"value,attr"`
}

type UnitSight struct {
	Value int `xml:"value,attr"`
}

type UnitTime struct {
	Value int `xml:"value,attr"`
}

type UnitMultiSelection struct {
	Value bool `xml:"value,attr"`
}

type UnitCellmap struct {
	Value bool `xml:"value,attr"`
}

type UnitImage struct {
	Path string `xml:"path,attr"`
}

type UnitImageCancel struct {
	Path string `xml:"path,attr"`
}

type UnitMeetingPoint struct {
	Value bool `xml:"value,attr"`
}

// UnitHP represents health points configuration
type UnitHP struct {
	Value        int `xml:"value,attr"`
	Regeneration int `xml:"regeneration,attr"`
}

// UnitEP represents energy points configuration (optional)
type UnitEP struct {
	Value           int `xml:"value,attr"`
	Regeneration    int `xml:"regeneration,attr"`
	StartPercentage int `xml:"start-percentage,attr"`
}

// Field represents terrain types the unit can occupy
type Field struct {
	Value string `xml:"value,attr"`
}

// ResourceRequirement represents the cost to create/upgrade this unit
type ResourceRequirement struct {
	Name   string `xml:"name,attr"`
	Amount int    `xml:"amount,attr"`
}

// SoundGroup represents a collection of sound files for unit feedback
type SoundGroup struct {
	Enabled bool        `xml:"enabled,attr"`
	Sounds  []SoundFile `xml:"sound"`
}

// SoundFile represents an individual sound file reference
type SoundFile struct {
	Path string `xml:"path,attr"`
}

// Skill represents a unit ability/animation like movement, attacking, building
type Skill struct {
	Type        SkillType        `xml:"type"`
	Name        SkillName        `xml:"name"`
	EPCost      SkillEPCost      `xml:"ep-cost"`
	Speed       SkillSpeed       `xml:"speed"`
	AnimSpeed   SkillAnimSpeed   `xml:"anim-speed"`
	Animation   SkillAnimation   `xml:"animation"`
	Sound       *SkillSound      `xml:"sound,omitempty"`

	// Attack-specific fields
	AttackStrength  *SkillAttackStrength  `xml:"attack-strenght,omitempty"` // Note: typo in original XML
	AttackVar       *SkillAttackVar       `xml:"attack-var,omitempty"`
	AttackRange     *SkillAttackRange     `xml:"attack-range,omitempty"`
	AttackType      *SkillAttackType      `xml:"attack-type,omitempty"`
	AttackFields    []Field               `xml:"attack-fields>field,omitempty"`
	AttackStartTime *SkillAttackStartTime `xml:"attack-start-time,omitempty"`
	Projectile      *Projectile           `xml:"projectile,omitempty"`
}

// Skill helper structs for XML parsing
type SkillType struct {
	Value string `xml:"value,attr"`
}

type SkillName struct {
	Value string `xml:"value,attr"`
}

type SkillEPCost struct {
	Value int `xml:"value,attr"`
}

type SkillSpeed struct {
	Value int `xml:"value,attr"`
}

type SkillAnimSpeed struct {
	Value int `xml:"value,attr"`
}

type SkillAnimation struct {
	Path string `xml:"path,attr"`
}

type SkillAttackStrength struct {
	Value int `xml:"value,attr"`
}

type SkillAttackVar struct {
	Value int `xml:"value,attr"`
}

type SkillAttackRange struct {
	Value int `xml:"value,attr"`
}

type SkillAttackType struct {
	Value string `xml:"value,attr"`
}

type SkillAttackStartTime struct {
	Value float64 `xml:"value,attr"`
}

// SkillSound represents sound configuration for skills
type SkillSound struct {
	Enabled   bool        `xml:"enabled,attr"`
	StartTime *float64    `xml:"start-time,attr,omitempty"`
	SoundFiles []SoundFile `xml:"sound-file"`
}

// Projectile represents projectile configuration for ranged attacks
type Projectile struct {
	Value    bool               `xml:"value,attr"`
	Particle *ProjectileParticle `xml:"particle,omitempty"`
	Sound    *SkillSound        `xml:"sound,omitempty"`
}

// ProjectileParticle represents particle effects for projectiles
type ProjectileParticle struct {
	Value bool   `xml:"value,attr"`
	Path  string `xml:"path,attr"`
}

// Command represents an action the unit can perform (move, attack, build, etc.)
type Command struct {
	Type        CommandType  `xml:"type"`
	Name        CommandName  `xml:"name"`
	Image       CommandImage `xml:"image"`

	// Skill references for different command phases
	MoveSkill       *CommandMoveSkill    `xml:"move-skill,omitempty"`
	AttackSkill     *CommandAttackSkill  `xml:"attack-skill,omitempty"`
	BuildSkill      *CommandBuildSkill   `xml:"build-skill,omitempty"`
	HarvestSkill    *CommandHarvestSkill `xml:"harvest-skill,omitempty"`
	RepairSkill     *CommandRepairSkill  `xml:"repair-skill,omitempty"`
	MorphSkill      *CommandMorphSkill   `xml:"morph-skill,omitempty"`
	StopSkill       *CommandStopSkill    `xml:"stop-skill,omitempty"`

	// Command-specific configuration
	AttackRange        *CommandAttackRange `xml:"attack-range,omitempty"`
	AttackType         *CommandAttackType  `xml:"attack-type,omitempty"`
	Buildings          []Building          `xml:"buildings>building,omitempty"`
	HarvestedResources []HarvestedResource `xml:"harvested-resources>resource,omitempty"`
	MaxLoad            *CommandMaxLoad     `xml:"max-load,omitempty"`
	HitsPerUnit        *CommandHitsPerUnit `xml:"hits-per-unit,omitempty"`
	MorphUnit          *CommandMorphUnit   `xml:"morph-unit,omitempty"`
	Discount           *CommandDiscount    `xml:"discount,omitempty"`
}

// Command helper structs for XML parsing
type CommandType struct {
	Value string `xml:"value,attr"`
}

type CommandName struct {
	Value string `xml:"value,attr"`
}

type CommandImage struct {
	Path string `xml:"path,attr"`
}

type CommandMoveSkill struct {
	Value string `xml:"value,attr"`
}

type CommandAttackSkill struct {
	Value string `xml:"value,attr"`
}

type CommandBuildSkill struct {
	Value string `xml:"value,attr"`
}

type CommandHarvestSkill struct {
	Value string `xml:"value,attr"`
}

type CommandRepairSkill struct {
	Value string `xml:"value,attr"`
}

type CommandMorphSkill struct {
	Value string `xml:"value,attr"`
}

type CommandStopSkill struct {
	Value string `xml:"value,attr"`
}

type CommandAttackRange struct {
	Value int `xml:"value,attr"`
}

type CommandAttackType struct {
	Value string `xml:"value,attr"`
}

type CommandMaxLoad struct {
	Value int `xml:"value,attr"`
}

type CommandHitsPerUnit struct {
	Value int `xml:"value,attr"`
}

type CommandMorphUnit struct {
	Name string `xml:"name,attr"`
}

type CommandDiscount struct {
	Value int `xml:"value,attr"`
}

// Building represents a building type that can be constructed
type Building struct {
	Name string `xml:"name,attr"`
}

// HarvestedResource represents a resource type that can be harvested
type HarvestedResource struct {
	Name string `xml:"name,attr"`
}

// UnitDefinition represents a complete unit with its name and parsed data
type UnitDefinition struct {
	Name string // Unit name (derived from directory name)
	Unit Unit   // Parsed XML data
}

// LoadUnit parses a single unit XML file and returns the Unit structure
func LoadUnit(xmlPath string) (*Unit, error) {
	// Read the XML file
	data, err := os.ReadFile(xmlPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read unit file %s: %w", xmlPath, err)
	}

	// Parse the XML
	var unit Unit
	err = xml.Unmarshal(data, &unit)
	if err != nil {
		return nil, fmt.Errorf("failed to parse unit XML %s: %w", xmlPath, err)
	}

	return &unit, nil
}

// LoadAllUnitsFromFaction loads all unit definitions from a faction's units directory
func LoadAllUnitsFromFaction(unitsDir string) ([]UnitDefinition, error) {
	var units []UnitDefinition

	// Read all subdirectories in the units folder
	entries, err := os.ReadDir(unitsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read units directory %s: %w", unitsDir, err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		unitName := entry.Name()
		unitXMLPath := filepath.Join(unitsDir, unitName, unitName+".xml")

		// Check if the unit XML file exists
		if _, err := os.Stat(unitXMLPath); os.IsNotExist(err) {
			fmt.Printf("Warning: No XML file found for unit %s at %s\n", unitName, unitXMLPath)
			continue
		}

		// Load the unit
		unit, err := LoadUnit(unitXMLPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load unit %s: %w", unitName, err)
		}

		units = append(units, UnitDefinition{
			Name: unitName,
			Unit: *unit,
		})
	}

	return units, nil
}

// PrintUnits prints all units with their basic stats and costs
func PrintUnits(units []UnitDefinition) {
	fmt.Println("Units:")
	for i, unit := range units {
		fmt.Printf("  %d. %s\n", i+1, unit.Name)
		fmt.Printf("    HP: %d (armor: %d %s)\n",
			unit.Unit.Parameters.MaxHP.Value,
			unit.Unit.Parameters.Armor,
			unit.Unit.Parameters.ArmorType)

		if len(unit.Unit.Parameters.ResourceRequirements) > 0 {
			fmt.Print("    Cost: ")
			for j, req := range unit.Unit.Parameters.ResourceRequirements {
				if j > 0 {
					fmt.Print(", ")
				}
				fmt.Printf("%s %d", req.Name, req.Amount)
			}
			fmt.Println()
		}

		fmt.Printf("    Skills: %d, Commands: %d\n",
			len(unit.Unit.Skills),
			len(unit.Unit.Commands))
		fmt.Println()
	}
}

// GetUnitByName finds a unit definition by name
func GetUnitByName(units []UnitDefinition, name string) *UnitDefinition {
	for _, unit := range units {
		if unit.Name == name {
			return &unit
		}
	}
	return nil
}

// GetResourceCost returns the cost for a specific resource type
func (ud *UnitDefinition) GetResourceCost(resourceName string) int {
	for _, req := range ud.Unit.Parameters.ResourceRequirements {
		if req.Name == resourceName {
			return req.Amount
		}
	}
	return 0
}

// HasField checks if the unit can move on a specific terrain type
func (ud *UnitDefinition) HasField(fieldType string) bool {
	for _, field := range ud.Unit.Parameters.Fields {
		if field.Value == fieldType {
			return true
		}
	}
	return false
}

// GetSkillByName finds a skill by its name
func (ud *UnitDefinition) GetSkillByName(skillName string) *Skill {
	for _, skill := range ud.Unit.Skills {
		if skill.Name.Value == skillName {
			return &skill
		}
	}
	return nil
}

// GetCommandByName finds a command by its name
func (ud *UnitDefinition) GetCommandByName(commandName string) *Command {
	for _, command := range ud.Unit.Commands {
		if command.Name.Value == commandName {
			return &command
		}
	}
	return nil
}