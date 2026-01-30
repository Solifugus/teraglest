package data

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
)

// Resource represents a game resource definition from resource.xml files
type Resource struct {
	XMLName xml.Name     `xml:"resource"`
	Image   ResourceImage `xml:"image"`
	Type    ResourceType  `xml:"type"`
}

// ResourceImage represents the image element in resource XML
type ResourceImage struct {
	Path string `xml:"path,attr"`
}

// ResourceType contains the resource type configuration
type ResourceType struct {
	Value          string             `xml:"value,attr"`        // "tech", "static", "consumable"
	Model          *ResourceModel     `xml:"model,omitempty"`   // Path to 3D model (optional)
	DefaultAmount  *ResourceAmount    `xml:"default-amount,omitempty"`  // Default amount on map (optional)
	ResourceNumber *ResourceNumber    `xml:"resource-number,omitempty"` // Resource ID number (optional)
}

// ResourceModel represents the model element in resource type
type ResourceModel struct {
	Path string `xml:"path,attr"`
}

// ResourceAmount represents the default-amount element
type ResourceAmount struct {
	Value int `xml:"value,attr"`
}

// ResourceNumber represents the resource-number element
type ResourceNumber struct {
	Value int `xml:"value,attr"`
}

// ResourceDefinition represents a complete resource with its name and configuration
type ResourceDefinition struct {
	Name     string   // Resource name (derived from directory name)
	Resource Resource // Parsed XML data
}

// LoadResource parses a single resource XML file and returns the Resource structure
func LoadResource(xmlPath string) (*Resource, error) {
	// Read the XML file
	data, err := os.ReadFile(xmlPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read resource file %s: %w", xmlPath, err)
	}

	// Parse the XML
	var resource Resource
	err = xml.Unmarshal(data, &resource)
	if err != nil {
		return nil, fmt.Errorf("failed to parse resource XML %s: %w", xmlPath, err)
	}

	return &resource, nil
}

// LoadAllResources loads all resource definitions from a tech tree resources directory
func LoadAllResources(resourcesDir string) ([]ResourceDefinition, error) {
	var resources []ResourceDefinition

	// Read all subdirectories in the resources folder
	entries, err := os.ReadDir(resourcesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read resources directory %s: %w", resourcesDir, err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		resourceName := entry.Name()
		resourceXMLPath := filepath.Join(resourcesDir, resourceName, resourceName+".xml")

		// Check if the resource XML file exists
		if _, err := os.Stat(resourceXMLPath); os.IsNotExist(err) {
			fmt.Printf("Warning: No XML file found for resource %s at %s\n", resourceName, resourceXMLPath)
			continue
		}

		// Load the resource
		resource, err := LoadResource(resourceXMLPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load resource %s: %w", resourceName, err)
		}

		resources = append(resources, ResourceDefinition{
			Name:     resourceName,
			Resource: *resource,
		})
	}

	return resources, nil
}

// PrintResources prints all resources for debugging/validation
func PrintResources(resources []ResourceDefinition) {
	fmt.Println("Resources:")
	for i, res := range resources {
		fmt.Printf("  %d. %s (type: %s", i+1, res.Name, res.Resource.Type.Value)
		if res.Resource.Type.DefaultAmount != nil && res.Resource.Type.DefaultAmount.Value > 0 {
			fmt.Printf(", default: %d", res.Resource.Type.DefaultAmount.Value)
		}
		if res.Resource.Type.ResourceNumber != nil && res.Resource.Type.ResourceNumber.Value > 0 {
			fmt.Printf(", number: %d", res.Resource.Type.ResourceNumber.Value)
		}
		fmt.Println(")")
	}
}

// GetResourceByName finds a resource definition by name
func GetResourceByName(resources []ResourceDefinition, name string) *ResourceDefinition {
	for _, res := range resources {
		if res.Name == name {
			return &res
		}
	}
	return nil
}

// IsStaticResource checks if a resource is static (not gathered/consumed)
func (rd *ResourceDefinition) IsStaticResource() bool {
	return rd.Resource.Type.Value == "static"
}

// IsTechResource checks if a resource is a tech resource (can be gathered)
func (rd *ResourceDefinition) IsTechResource() bool {
	return rd.Resource.Type.Value == "tech"
}

// IsConsumableResource checks if a resource is consumable
func (rd *ResourceDefinition) IsConsumableResource() bool {
	return rd.Resource.Type.Value == "consumable"
}