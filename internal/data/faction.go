package data

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
)

// Faction represents a complete faction definition from faction.xml
type Faction struct {
	XMLName                xml.Name             `xml:"faction"`
	StartingResources      []StartingResource   `xml:"starting-resources>resource"`
	StartingUnits          []StartingUnit       `xml:"starting-units>unit"`
	Music                  *Music               `xml:"music,omitempty"`
	FlatParticlePositions  *FlatParticlePositions `xml:"flat-particle-positions,omitempty"`
	AIBehavior             *AIBehavior          `xml:"ai-behavior,omitempty"`
}

// FlatParticlePositions represents the flat-particle-positions configuration
type FlatParticlePositions struct {
	Value bool `xml:"value,attr"`
}

// StartingResource represents a resource amount given to the faction at game start
type StartingResource struct {
	Name   string `xml:"name,attr"`
	Amount int    `xml:"amount,attr"`
}

// StartingUnit represents a unit type and count given to the faction at game start
type StartingUnit struct {
	Name   string `xml:"name,attr"`
	Amount int    `xml:"amount,attr"`
}

// Music represents faction music configuration
type Music struct {
	Value bool   `xml:"value,attr"`
	Path  string `xml:"path,attr"`
}

// AIBehavior represents AI configuration for computer players using this faction
type AIBehavior struct {
	WorkerUnits          []AIUnit      `xml:"worker-units>unit"`
	WarriorUnits         []AIUnit      `xml:"warrior-units>unit"`
	ResourceProducerUnits []AIUnit     `xml:"resource-producer-units>unit"`
	BuildingUnits        []AIUnit      `xml:"building-units>unit"`
	Upgrades             []AIUpgrade   `xml:"upgrades>upgrade"`
	StaticValues         []StaticValue `xml:"static-values>static"`
}

// AIUnit represents an AI unit configuration with minimum requirements
type AIUnit struct {
	Name    string `xml:"name,attr"`
	Minimum int    `xml:"minimum,attr"`
}

// AIUpgrade represents an AI upgrade preference
type AIUpgrade struct {
	Name string `xml:"name,attr"`
}

// StaticValue represents AI static configuration values
type StaticValue struct {
	TypeName string `xml:"type-name,attr"`
	Value    string `xml:"value,attr"`
}

// FactionDefinition represents a complete faction with its name and parsed data
type FactionDefinition struct {
	Name    string  // Faction name (derived from directory name)
	Faction Faction // Parsed XML data
}

// LoadFaction parses a single faction XML file and returns the Faction structure
func LoadFaction(xmlPath string) (*Faction, error) {
	// Read the XML file
	data, err := os.ReadFile(xmlPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read faction file %s: %w", xmlPath, err)
	}

	// Parse the XML
	var faction Faction
	err = xml.Unmarshal(data, &faction)
	if err != nil {
		return nil, fmt.Errorf("failed to parse faction XML %s: %w", xmlPath, err)
	}

	return &faction, nil
}

// LoadAllFactions loads all faction definitions from a tech tree factions directory
func LoadAllFactions(factionsDir string) ([]FactionDefinition, error) {
	var factions []FactionDefinition

	// Read all subdirectories in the factions folder
	entries, err := os.ReadDir(factionsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read factions directory %s: %w", factionsDir, err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		factionName := entry.Name()
		factionXMLPath := filepath.Join(factionsDir, factionName, factionName+".xml")

		// Check if the faction XML file exists
		if _, err := os.Stat(factionXMLPath); os.IsNotExist(err) {
			fmt.Printf("Warning: No XML file found for faction %s at %s\n", factionName, factionXMLPath)
			continue
		}

		// Load the faction
		faction, err := LoadFaction(factionXMLPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load faction %s: %w", factionName, err)
		}

		factions = append(factions, FactionDefinition{
			Name:    factionName,
			Faction: *faction,
		})
	}

	return factions, nil
}

// PrintFactions prints all factions with their starting units and resources
func PrintFactions(factions []FactionDefinition) {
	fmt.Println("Factions:")
	for i, faction := range factions {
		fmt.Printf("  %d. %s\n", i+1, faction.Name)

		// Print starting resources
		fmt.Println("    Starting Resources:")
		for _, res := range faction.Faction.StartingResources {
			fmt.Printf("      %s: %d\n", res.Name, res.Amount)
		}

		// Print starting units
		fmt.Println("    Starting Units:")
		for _, unit := range faction.Faction.StartingUnits {
			fmt.Printf("      %s: %d\n", unit.Name, unit.Amount)
		}

		// Print music info if available
		if faction.Faction.Music != nil && faction.Faction.Music.Value {
			fmt.Printf("    Music: %s\n", faction.Faction.Music.Path)
		}

		// Print AI behavior summary if available
		if faction.Faction.AIBehavior != nil {
			fmt.Printf("    AI: %d worker units, %d warrior units, %d upgrades\n",
				len(faction.Faction.AIBehavior.WorkerUnits),
				len(faction.Faction.AIBehavior.WarriorUnits),
				len(faction.Faction.AIBehavior.Upgrades))
		}

		fmt.Println()
	}
}

// GetFactionByName finds a faction definition by name
func GetFactionByName(factions []FactionDefinition, name string) *FactionDefinition {
	for _, faction := range factions {
		if faction.Name == name {
			return &faction
		}
	}
	return nil
}

// GetStartingResource returns the starting amount for a specific resource
func (fd *FactionDefinition) GetStartingResource(resourceName string) int {
	for _, res := range fd.Faction.StartingResources {
		if res.Name == resourceName {
			return res.Amount
		}
	}
	return 0
}

// GetStartingUnit returns the starting count for a specific unit type
func (fd *FactionDefinition) GetStartingUnit(unitName string) int {
	for _, unit := range fd.Faction.StartingUnits {
		if unit.Name == unitName {
			return unit.Amount
		}
	}
	return 0
}

// HasMusic checks if the faction has music enabled
func (fd *FactionDefinition) HasMusic() bool {
	return fd.Faction.Music != nil && fd.Faction.Music.Value
}

// GetMusicPath returns the faction's music path if available
func (fd *FactionDefinition) GetMusicPath() string {
	if fd.HasMusic() {
		return fd.Faction.Music.Path
	}
	return ""
}