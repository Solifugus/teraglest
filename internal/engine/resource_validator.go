package engine

import (
	"fmt"
)

// ResourceValidator provides centralized resource validation for game actions
type ResourceValidator struct {
	world *World
}

// ResourceCheck represents a resource validation request
type ResourceCheck struct {
	PlayerID int               // Player requesting the action
	Required map[string]int    // Resources required for the action
	Purpose  string            // Purpose of the resource check (for logging)
}

// ValidationResult represents the result of a resource validation
type ValidationResult struct {
	Valid   bool               // Whether the player has sufficient resources
	Error   string             // Error message if validation failed
	Missing map[string]int     // Resources the player is missing (if any)
}

// NewResourceValidator creates a new resource validator
func NewResourceValidator(world *World) *ResourceValidator {
	return &ResourceValidator{
		world: world,
	}
}

// ValidateResources checks if a player has sufficient resources for an action
func (rv *ResourceValidator) ValidateResources(check ResourceCheck) ValidationResult {
	if rv.world == nil {
		return ValidationResult{
			Valid: false,
			Error: "world is nil",
		}
	}

	player := rv.world.GetPlayer(check.PlayerID)
	if player == nil {
		return ValidationResult{
			Valid: false,
			Error: fmt.Sprintf("player %d not found", check.PlayerID),
		}
	}

	// Check each required resource
	missing := make(map[string]int)
	for resourceType, requiredAmount := range check.Required {
		currentAmount, exists := player.Resources[resourceType]
		if !exists {
			currentAmount = 0
		}

		if currentAmount < requiredAmount {
			missing[resourceType] = requiredAmount - currentAmount
		}
	}

	// If any resources are missing, validation fails
	if len(missing) > 0 {
		// Build detailed error message
		errorMsg := fmt.Sprintf("insufficient resources for %s:", check.Purpose)
		for resourceType := range missing {
			current := 0
			if amount, exists := player.Resources[resourceType]; exists {
				current = amount
			}
			required := check.Required[resourceType]
			errorMsg += fmt.Sprintf(" %s (have %d, need %d)", resourceType, current, required)
		}

		return ValidationResult{
			Valid:   false,
			Error:   errorMsg,
			Missing: missing,
		}
	}

	return ValidationResult{
		Valid: true,
	}
}

// CanAfford is a convenience method to check if a player can afford a cost
func (rv *ResourceValidator) CanAfford(playerID int, cost map[string]int) bool {
	if len(cost) == 0 {
		return true
	}

	result := rv.ValidateResources(ResourceCheck{
		PlayerID: playerID,
		Required: cost,
		Purpose:  "cost check",
	})

	return result.Valid
}

// GetMissingResources returns what resources a player lacks for a cost
func (rv *ResourceValidator) GetMissingResources(playerID int, cost map[string]int) map[string]int {
	result := rv.ValidateResources(ResourceCheck{
		PlayerID: playerID,
		Required: cost,
		Purpose:  "missing resource check",
	})

	return result.Missing
}

// ValidateUnitCost validates if a player can afford to create a unit
func (rv *ResourceValidator) ValidateUnitCost(playerID int, unitType string) ValidationResult {
	cost := rv.getUnitCost(playerID, unitType)
	if cost == nil {
		return ValidationResult{
			Valid: false,
			Error: fmt.Sprintf("could not determine cost for unit type %s", unitType),
		}
	}

	return rv.ValidateResources(ResourceCheck{
		PlayerID: playerID,
		Required: cost,
		Purpose:  fmt.Sprintf("unit creation (%s)", unitType),
	})
}

// ValidateBuildingCost validates if a player can afford to construct a building
func (rv *ResourceValidator) ValidateBuildingCost(playerID int, buildingType string) ValidationResult {
	cost := rv.getBuildingCost(playerID, buildingType)
	if cost == nil {
		return ValidationResult{
			Valid: false,
			Error: fmt.Sprintf("could not determine cost for building type %s", buildingType),
		}
	}

	return rv.ValidateResources(ResourceCheck{
		PlayerID: playerID,
		Required: cost,
		Purpose:  fmt.Sprintf("building construction (%s)", buildingType),
	})
}

// Helper method to get unit cost from AssetManager
func (rv *ResourceValidator) getUnitCost(playerID int, unitType string) map[string]int {
	player := rv.world.GetPlayer(playerID)
	if player == nil || player.FactionData == nil {
		return nil
	}

	// Load unit from AssetManager
	unit, err := rv.world.assetMgr.LoadUnit(player.FactionName, unitType)
	if err != nil {
		return nil
	}

	// Extract resource requirements
	costs := make(map[string]int)
	for _, req := range unit.Unit.Parameters.ResourceRequirements {
		costs[req.Name] = req.Amount
	}

	return costs
}

// Helper method to get building cost from AssetManager
func (rv *ResourceValidator) getBuildingCost(playerID int, buildingType string) map[string]int {
	player := rv.world.GetPlayer(playerID)
	if player == nil || player.FactionData == nil {
		return nil
	}

	// Load building from AssetManager
	building, err := rv.world.assetMgr.LoadUnit(player.FactionName, buildingType)
	if err != nil {
		return nil
	}

	// Extract resource requirements
	costs := make(map[string]int)
	for _, req := range building.Unit.Parameters.ResourceRequirements {
		costs[req.Name] = req.Amount
	}

	return costs
}