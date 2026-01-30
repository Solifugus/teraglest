package data

import (
	"encoding/xml"
	"fmt"
	"os"
)

// TechTree represents the complete tech tree structure from megapack.xml
type TechTree struct {
	XMLName          xml.Name           `xml:"tech-tree"`
	Description      TechTreeDescription `xml:"description"`
	AttackTypes      []AttackType       `xml:"attack-types>attack-type"`
	ArmorTypes       []ArmorType        `xml:"armor-types>armor-type"`
	DamageMultipliers []DamageMultiplier `xml:"damage-multipliers>damage-multiplier"`
}

// TechTreeDescription represents the description element
type TechTreeDescription struct {
	Value string `xml:"value,attr"`
}

// AttackType represents an attack type definition
type AttackType struct {
	Name string `xml:"name,attr"`
}

// ArmorType represents an armor type definition
type ArmorType struct {
	Name string `xml:"name,attr"`
}

// DamageMultiplier represents damage calculation rules between attack and armor types
type DamageMultiplier struct {
	Attack string  `xml:"attack,attr"`
	Armor  string  `xml:"armor,attr"`
	Value  float64 `xml:"value,attr"`
}

// LoadTechTree parses a tech tree XML file and returns the TechTree structure
func LoadTechTree(xmlPath string) (*TechTree, error) {
	// Read the XML file
	data, err := os.ReadFile(xmlPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read tech tree file %s: %w", xmlPath, err)
	}

	// Parse the XML
	var techTree TechTree
	err = xml.Unmarshal(data, &techTree)
	if err != nil {
		return nil, fmt.Errorf("failed to parse tech tree XML %s: %w", xmlPath, err)
	}

	return &techTree, nil
}

// PrintAttackTypes prints all attack types for debugging/validation
func (tt *TechTree) PrintAttackTypes() {
	fmt.Println("Attack Types:")
	for i, attackType := range tt.AttackTypes {
		fmt.Printf("  %d. %s\n", i+1, attackType.Name)
	}
}

// PrintArmorTypes prints all armor types for debugging/validation
func (tt *TechTree) PrintArmorTypes() {
	fmt.Println("Armor Types:")
	for i, armorType := range tt.ArmorTypes {
		fmt.Printf("  %d. %s\n", i+1, armorType.Name)
	}
}

// PrintDamageMultipliers prints all damage multipliers for debugging/validation
func (tt *TechTree) PrintDamageMultipliers() {
	fmt.Println("Damage Multipliers:")
	for _, dm := range tt.DamageMultipliers {
		fmt.Printf("  %s vs %s: %.2fx\n", dm.Attack, dm.Armor, dm.Value)
	}
}

// GetDamageMultiplier returns the damage multiplier for a specific attack vs armor combination
// Returns 1.0 if no specific multiplier is defined
func (tt *TechTree) GetDamageMultiplier(attackType, armorType string) float64 {
	for _, dm := range tt.DamageMultipliers {
		if dm.Attack == attackType && dm.Armor == armorType {
			return dm.Value
		}
	}
	return 1.0 // Default multiplier when no specific rule exists
}

// HasAttackType checks if an attack type exists in the tech tree
func (tt *TechTree) HasAttackType(attackType string) bool {
	for _, at := range tt.AttackTypes {
		if at.Name == attackType {
			return true
		}
	}
	return false
}

// HasArmorType checks if an armor type exists in the tech tree
func (tt *TechTree) HasArmorType(armorType string) bool {
	for _, at := range tt.ArmorTypes {
		if at.Name == armorType {
			return true
		}
	}
	return false
}