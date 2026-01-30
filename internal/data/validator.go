package data

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ValidationSeverity represents the severity level of a validation issue
type ValidationSeverity int

const (
	ValidationError   ValidationSeverity = iota // Critical issues that prevent proper functioning
	ValidationWarning                           // Non-critical issues that may affect gameplay
	ValidationInfo                              // Informational messages about the data
)

// ValidationIssue represents a single validation problem found in the data
type ValidationIssue struct {
	Severity    ValidationSeverity
	Category    string    // e.g., "XML Reference", "Asset Missing", "Data Consistency"
	Message     string    // Human-readable description
	File        string    // File where the issue was found
	Line        int       // Line number (if available)
	Field       string    // Specific field or element name
	Value       string    // The problematic value
	Context     string    // Additional context information
	Suggestion  string    // Suggested fix (if available)
	Timestamp   time.Time
}

// ValidationReport contains all validation issues found during validation
type ValidationReport struct {
	Issues       []ValidationIssue
	ErrorCount   int
	WarningCount int
	InfoCount    int
	FilesChecked int
	Duration     time.Duration
	Timestamp    time.Time
}

// DataValidator performs comprehensive validation of game data
type DataValidator struct {
	techTreeRoot     string
	validationRules  []ValidationRule
	assetManager     *AssetManager
	techTree         *TechTree
	resources        []ResourceDefinition
	factions         []FactionDefinition
	enableFileChecks bool // Whether to perform file existence checks
}

// ValidationRule defines a validation rule that can be applied to game data
type ValidationRule interface {
	Name() string
	Description() string
	Validate(validator *DataValidator, report *ValidationReport) error
}

// NewDataValidator creates a new data validator
func NewDataValidator(techTreeRoot string, assetManager *AssetManager) *DataValidator {
	return &DataValidator{
		techTreeRoot:     techTreeRoot,
		validationRules:  getDefaultValidationRules(),
		assetManager:     assetManager,
		enableFileChecks: true,
	}
}

// ValidateAllData performs comprehensive validation of all game data
func (v *DataValidator) ValidateAllData() (*ValidationReport, error) {
	startTime := time.Now()

	report := &ValidationReport{
		Issues:    make([]ValidationIssue, 0),
		Timestamp: startTime,
	}

	// Load core data structures for validation
	if err := v.loadCoreData(report); err != nil {
		return report, fmt.Errorf("failed to load core data for validation: %w", err)
	}

	// Apply all validation rules
	for _, rule := range v.validationRules {
		if err := rule.Validate(v, report); err != nil {
			v.addIssue(report, ValidationError, "Validation System",
				fmt.Sprintf("Validation rule '%s' failed: %v", rule.Name(), err),
				"", 0, rule.Name(), "", "", "Check validation system implementation")
		}
	}

	// Count issues by severity
	v.countIssues(report)

	report.Duration = time.Since(startTime)
	return report, nil
}

// ValidateTechTree validates tech tree data and references
func (v *DataValidator) ValidateTechTree() (*ValidationReport, error) {
	report := &ValidationReport{
		Issues:    make([]ValidationIssue, 0),
		Timestamp: time.Now(),
	}

	if v.techTree == nil {
		var err error
		v.techTree, err = v.assetManager.LoadTechTree()
		if err != nil {
			v.addIssue(report, ValidationError, "Data Loading",
				"Failed to load tech tree for validation",
				"megapack.xml", 0, "tech-tree", "", "", "Ensure tech tree XML is valid")
			return report, err
		}
	}

	report.FilesChecked++

	// Validate attack types
	if len(v.techTree.AttackTypes) == 0 {
		v.addIssue(report, ValidationWarning, "Data Completeness",
			"No attack types defined in tech tree",
			"megapack.xml", 0, "attack-types", "", "", "Add at least one attack type")
	}

	// Validate armor types
	if len(v.techTree.ArmorTypes) == 0 {
		v.addIssue(report, ValidationWarning, "Data Completeness",
			"No armor types defined in tech tree",
			"megapack.xml", 0, "armor-types", "", "", "Add at least one armor type")
	}

	// Validate damage multipliers reference valid attack and armor types
	v.validateDamageMultipliers(report)

	v.countIssues(report)
	return report, nil
}

// ValidateFaction validates a specific faction and its data
func (v *DataValidator) ValidateFaction(factionName string) (*ValidationReport, error) {
	report := &ValidationReport{
		Issues:    make([]ValidationIssue, 0),
		Timestamp: time.Now(),
	}

	// Load faction data
	factions, err := v.assetManager.LoadFactions()
	if err != nil {
		v.addIssue(report, ValidationError, "Data Loading",
			"Failed to load factions for validation",
			"factions directory", 0, "faction", factionName, "", "Check faction XML files")
		return report, err
	}

	// Find the specific faction
	var faction *FactionDefinition
	for _, f := range factions {
		if f.Name == factionName {
			faction = &f
			break
		}
	}

	if faction == nil {
		v.addIssue(report, ValidationError, "Data Reference",
			fmt.Sprintf("Faction '%s' not found", factionName),
			fmt.Sprintf("factions/%s/%s.xml", factionName, factionName), 0, "faction", factionName, "",
			"Ensure faction XML file exists and is properly named")
		v.countIssues(report)
		return report, nil
	}

	report.FilesChecked++

	// Validate starting resources reference valid resources
	v.validateFactionStartingResources(faction, report)

	// Validate starting units exist
	v.validateFactionStartingUnits(faction, report)

	// Load and validate all units in this faction
	v.validateFactionUnits(faction, report)

	v.countIssues(report)
	return report, nil
}

// ValidateAssetReferences validates that all asset files referenced in XML exist
func (v *DataValidator) ValidateAssetReferences() (*ValidationReport, error) {
	report := &ValidationReport{
		Issues:    make([]ValidationIssue, 0),
		Timestamp: time.Now(),
	}

	if !v.enableFileChecks {
		v.addIssue(report, ValidationInfo, "Asset Validation",
			"Asset file existence checks are disabled",
			"", 0, "config", "enableFileChecks", "false", "Enable file checks for complete validation")
		return report, nil
	}

	// Load all factions to check their assets
	factions, err := v.assetManager.LoadFactions()
	if err != nil {
		v.addIssue(report, ValidationError, "Data Loading",
			"Failed to load factions for asset validation",
			"factions directory", 0, "factions", "", "", "Check faction directory structure")
		return report, err
	}

	// Check assets for each faction
	for _, faction := range factions {
		v.validateFactionAssets(&faction, report)
	}

	v.countIssues(report)
	return report, nil
}

// Helper method to load core data structures needed for validation
func (v *DataValidator) loadCoreData(report *ValidationReport) error {
	var err error

	// Load tech tree
	if v.techTree == nil {
		v.techTree, err = v.assetManager.LoadTechTree()
		if err != nil {
			v.addIssue(report, ValidationError, "Data Loading",
				"Failed to load tech tree", "megapack.xml", 0, "tech-tree", "", "",
				"Ensure megapack.xml exists and is valid")
			return err
		}
	}

	// Load resources
	if v.resources == nil {
		v.resources, err = v.assetManager.LoadResources()
		if err != nil {
			v.addIssue(report, ValidationError, "Data Loading",
				"Failed to load resources", "resources directory", 0, "resources", "", "",
				"Check resources directory and XML files")
			// Don't return error, continue with other validation
		}
	}

	// Load factions
	if v.factions == nil {
		v.factions, err = v.assetManager.LoadFactions()
		if err != nil {
			v.addIssue(report, ValidationError, "Data Loading",
				"Failed to load factions", "factions directory", 0, "factions", "", "",
				"Check factions directory and XML files")
			// Don't return error, continue with other validation
		}
	}

	return nil
}

// Validate damage multipliers reference valid attack and armor types
func (v *DataValidator) validateDamageMultipliers(report *ValidationReport) {
	attackTypeMap := make(map[string]bool)
	armorTypeMap := make(map[string]bool)

	// Build maps of valid types
	for _, at := range v.techTree.AttackTypes {
		attackTypeMap[at.Name] = true
	}
	for _, at := range v.techTree.ArmorTypes {
		armorTypeMap[at.Name] = true
	}

	// Check each damage multiplier
	for _, dm := range v.techTree.DamageMultipliers {
		if !attackTypeMap[dm.Attack] {
			v.addIssue(report, ValidationError, "XML Reference",
				fmt.Sprintf("Damage multiplier references unknown attack type '%s'", dm.Attack),
				"megapack.xml", 0, "damage-multiplier", dm.Attack,
				fmt.Sprintf("attack:%s -> armor:%s = %.1f", dm.Attack, dm.Armor, dm.Value),
				"Ensure attack type is defined in attack-types section")
		}

		if !armorTypeMap[dm.Armor] {
			v.addIssue(report, ValidationError, "XML Reference",
				fmt.Sprintf("Damage multiplier references unknown armor type '%s'", dm.Armor),
				"megapack.xml", 0, "damage-multiplier", dm.Armor,
				fmt.Sprintf("attack:%s -> armor:%s = %.1f", dm.Attack, dm.Armor, dm.Value),
				"Ensure armor type is defined in armor-types section")
		}
	}
}

// Validate faction starting resources reference valid resources
func (v *DataValidator) validateFactionStartingResources(faction *FactionDefinition, report *ValidationReport) {
	if v.resources == nil {
		return // Skip if resources couldn't be loaded
	}

	resourceMap := make(map[string]bool)
	for _, res := range v.resources {
		resourceMap[res.Name] = true
	}

	for _, startRes := range faction.Faction.StartingResources {
		if !resourceMap[startRes.Name] {
			v.addIssue(report, ValidationError, "XML Reference",
				fmt.Sprintf("Faction starting resource '%s' is not defined in resources", startRes.Name),
				fmt.Sprintf("factions/%s/%s.xml", faction.Name, faction.Name), 0,
				"starting-resources", startRes.Name,
				fmt.Sprintf("amount: %d", startRes.Amount),
				"Ensure resource is defined in resources directory")
		}

		if startRes.Amount < 0 {
			v.addIssue(report, ValidationWarning, "Data Consistency",
				fmt.Sprintf("Faction starting resource '%s' has negative amount", startRes.Name),
				fmt.Sprintf("factions/%s/%s.xml", faction.Name, faction.Name), 0,
				"starting-resources", startRes.Name,
				fmt.Sprintf("amount: %d", startRes.Amount),
				"Use non-negative amount for starting resources")
		}
	}
}

// Validate faction starting units exist
func (v *DataValidator) validateFactionStartingUnits(faction *FactionDefinition, report *ValidationReport) {
	for _, startUnit := range faction.Faction.StartingUnits {
		// Check if unit XML file exists
		unitPath := filepath.Join(v.techTreeRoot, "factions", faction.Name, "units", startUnit.Name, startUnit.Name+".xml")
		if _, err := os.Stat(unitPath); os.IsNotExist(err) {
			v.addIssue(report, ValidationError, "Asset Missing",
				fmt.Sprintf("Starting unit '%s' XML file not found", startUnit.Name),
				fmt.Sprintf("factions/%s/%s.xml", faction.Name, faction.Name), 0,
				"starting-units", startUnit.Name,
				fmt.Sprintf("amount: %d", startUnit.Amount),
				fmt.Sprintf("Create unit XML file at %s", unitPath))
		}

		if startUnit.Amount <= 0 {
			v.addIssue(report, ValidationWarning, "Data Consistency",
				fmt.Sprintf("Starting unit '%s' has invalid amount", startUnit.Name),
				fmt.Sprintf("factions/%s/%s.xml", faction.Name, faction.Name), 0,
				"starting-units", startUnit.Name,
				fmt.Sprintf("amount: %d", startUnit.Amount),
				"Use positive amount for starting units")
		}
	}
}

// Validate all units in a faction
func (v *DataValidator) validateFactionUnits(faction *FactionDefinition, report *ValidationReport) {
	unitsDir := filepath.Join(v.techTreeRoot, "factions", faction.Name, "units")
	entries, err := os.ReadDir(unitsDir)
	if err != nil {
		v.addIssue(report, ValidationError, "Asset Missing",
			fmt.Sprintf("Cannot read units directory for faction '%s'", faction.Name),
			fmt.Sprintf("factions/%s/units/", faction.Name), 0, "units", "", "",
			"Ensure units directory exists and is readable")
		return
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		unitName := entry.Name()

		// Try to load the unit to validate its XML structure
		unit, err := v.assetManager.LoadUnit(faction.Name, unitName)
		if err != nil {
			v.addIssue(report, ValidationError, "XML Structure",
				fmt.Sprintf("Failed to parse unit '%s' XML: %v", unitName, err),
				fmt.Sprintf("factions/%s/units/%s/%s.xml", faction.Name, unitName, unitName), 0,
				"unit", unitName, "", "Check XML syntax and structure")
			continue
		}

		report.FilesChecked++
		v.validateUnit(faction, unit, report)
	}
}

// Validate individual unit data
func (v *DataValidator) validateUnit(faction *FactionDefinition, unit *UnitDefinition, report *ValidationReport) {
	unitFile := fmt.Sprintf("factions/%s/units/%s/%s.xml", faction.Name, unit.Name, unit.Name)

	// Validate resource requirements reference valid resources
	if v.resources != nil {
		resourceMap := make(map[string]bool)
		for _, res := range v.resources {
			resourceMap[res.Name] = true
		}

		for _, req := range unit.Unit.Parameters.ResourceRequirements {
			if !resourceMap[req.Name] {
				v.addIssue(report, ValidationError, "XML Reference",
					fmt.Sprintf("Unit '%s' requires unknown resource '%s'", unit.Name, req.Name),
					unitFile, 0, "resource-requirements", req.Name,
					fmt.Sprintf("amount: %d", req.Amount),
					"Ensure resource is defined in resources directory")
			}

			if req.Amount <= 0 {
				v.addIssue(report, ValidationWarning, "Data Consistency",
					fmt.Sprintf("Unit '%s' resource requirement '%s' has invalid amount", unit.Name, req.Name),
					unitFile, 0, "resource-requirements", req.Name,
					fmt.Sprintf("amount: %d", req.Amount),
					"Use positive amounts for resource requirements")
			}
		}
	}

	// Validate armor type references valid armor types
	if v.techTree != nil {
		armorTypeMap := make(map[string]bool)
		for _, at := range v.techTree.ArmorTypes {
			armorTypeMap[at.Name] = true
		}

		if !armorTypeMap[unit.Unit.Parameters.ArmorType.Value] {
			v.addIssue(report, ValidationError, "XML Reference",
				fmt.Sprintf("Unit '%s' uses unknown armor type '%s'", unit.Name, unit.Unit.Parameters.ArmorType.Value),
				unitFile, 0, "armor-type", unit.Unit.Parameters.ArmorType.Value, "",
				"Use armor type defined in tech tree")
		}
	}

	// Basic parameter validation
	if unit.Unit.Parameters.MaxHP.Value <= 0 {
		v.addIssue(report, ValidationError, "Data Consistency",
			fmt.Sprintf("Unit '%s' has invalid max HP", unit.Name),
			unitFile, 0, "max-hp", "",
			fmt.Sprintf("value: %d", unit.Unit.Parameters.MaxHP.Value),
			"Units must have positive HP values")
	}

	if unit.Unit.Parameters.Size.Value <= 0 {
		v.addIssue(report, ValidationWarning, "Data Consistency",
			fmt.Sprintf("Unit '%s' has invalid size", unit.Name),
			unitFile, 0, "size", "",
			fmt.Sprintf("value: %d", unit.Unit.Parameters.Size.Value),
			"Units should have positive size values")
	}
}

// Validate assets referenced by a faction exist
func (v *DataValidator) validateFactionAssets(faction *FactionDefinition, report *ValidationReport) {
	// Check faction units directory exists
	unitsDir := filepath.Join(v.techTreeRoot, "factions", faction.Name, "units")
	if _, err := os.Stat(unitsDir); os.IsNotExist(err) {
		v.addIssue(report, ValidationError, "Asset Missing",
			fmt.Sprintf("Units directory missing for faction '%s'", faction.Name),
			fmt.Sprintf("factions/%s/units/", faction.Name), 0, "directory", "units", "",
			"Create units directory for faction")
		return
	}

	// Check each unit's model and texture files
	entries, err := os.ReadDir(unitsDir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		unitName := entry.Name()
		v.validateUnitAssets(faction.Name, unitName, report)
	}
}

// Validate assets for a specific unit
func (v *DataValidator) validateUnitAssets(factionName, unitName string, report *ValidationReport) {
	unitDir := filepath.Join(v.techTreeRoot, "factions", factionName, "units", unitName)

	// Check unit XML file exists
	xmlFile := filepath.Join(unitDir, unitName+".xml")
	if _, err := os.Stat(xmlFile); os.IsNotExist(err) {
		v.addIssue(report, ValidationError, "Asset Missing",
			fmt.Sprintf("Unit XML file missing: %s", unitName),
			fmt.Sprintf("factions/%s/units/%s/%s.xml", factionName, unitName, unitName), 0,
			"file", "xml", "", "Create unit XML file")
		return // Can't validate further without XML
	}

	// Check models directory
	modelsDir := filepath.Join(unitDir, "models")
	if _, err := os.Stat(modelsDir); os.IsNotExist(err) {
		v.addIssue(report, ValidationWarning, "Asset Missing",
			fmt.Sprintf("Models directory missing for unit '%s'", unitName),
			fmt.Sprintf("factions/%s/units/%s/models/", factionName, unitName), 0,
			"directory", "models", "", "Create models directory and add G3D files")
		return
	}

	// Check for at least one model file
	modelEntries, err := os.ReadDir(modelsDir)
	if err != nil {
		return
	}

	hasModels := false
	for _, entry := range modelEntries {
		if strings.HasSuffix(entry.Name(), ".g3d") {
			hasModels = true

			// Try to load the model to validate it
			modelPath := filepath.Join("factions", factionName, "units", unitName, "models", entry.Name())
			_, err := v.assetManager.LoadG3DModel(modelPath)
			if err != nil {
				v.addIssue(report, ValidationError, "Asset Corrupt",
					fmt.Sprintf("Cannot load G3D model: %s", entry.Name()),
					modelPath, 0, "model", entry.Name(), "",
					"Check G3D file format and integrity")
			}
		}
	}

	if !hasModels {
		v.addIssue(report, ValidationWarning, "Asset Missing",
			fmt.Sprintf("No G3D model files found for unit '%s'", unitName),
			fmt.Sprintf("factions/%s/units/%s/models/", factionName, unitName), 0,
			"models", "*.g3d", "", "Add at least one G3D model file")
	}

	// Check images directory (optional)
	imagesDir := filepath.Join(unitDir, "images")
	if _, err := os.Stat(imagesDir); err == nil {
		// Directory exists, check for texture files
		imageEntries, err := os.ReadDir(imagesDir)
		if err == nil {
			hasTextures := false
			for _, entry := range imageEntries {
				ext := strings.ToLower(filepath.Ext(entry.Name()))
				if ext == ".png" || ext == ".jpg" || ext == ".jpeg" {
					hasTextures = true
					break
				}
			}

			if !hasTextures {
				v.addIssue(report, ValidationInfo, "Asset Optional",
					fmt.Sprintf("No texture files found for unit '%s'", unitName),
					fmt.Sprintf("factions/%s/units/%s/images/", factionName, unitName), 0,
					"textures", "*.png/*.jpg", "", "Add texture files if unit uses custom textures")
			}
		}
	}
}

// Helper method to add an issue to the validation report
func (v *DataValidator) addIssue(report *ValidationReport, severity ValidationSeverity, category, message, file string, line int, field, value, context, suggestion string) {
	issue := ValidationIssue{
		Severity:   severity,
		Category:   category,
		Message:    message,
		File:       file,
		Line:       line,
		Field:      field,
		Value:      value,
		Context:    context,
		Suggestion: suggestion,
		Timestamp:  time.Now(),
	}

	report.Issues = append(report.Issues, issue)
}

// Helper method to count issues by severity
func (v *DataValidator) countIssues(report *ValidationReport) {
	report.ErrorCount = 0
	report.WarningCount = 0
	report.InfoCount = 0

	for _, issue := range report.Issues {
		switch issue.Severity {
		case ValidationError:
			report.ErrorCount++
		case ValidationWarning:
			report.WarningCount++
		case ValidationInfo:
			report.InfoCount++
		}
	}
}

// String methods for enums
func (s ValidationSeverity) String() string {
	switch s {
	case ValidationError:
		return "ERROR"
	case ValidationWarning:
		return "WARNING"
	case ValidationInfo:
		return "INFO"
	default:
		return "UNKNOWN"
	}
}

// PrintReport prints a formatted validation report
func (report *ValidationReport) PrintReport() {
	fmt.Println("=== Data Validation Report ===")
	fmt.Printf("Files Checked: %d\n", report.FilesChecked)
	fmt.Printf("Duration: %v\n", report.Duration)
	fmt.Printf("Issues Found: %d (%d errors, %d warnings, %d info)\n",
		len(report.Issues), report.ErrorCount, report.WarningCount, report.InfoCount)
	fmt.Println()

	if len(report.Issues) == 0 {
		fmt.Println("✅ No validation issues found!")
		return
	}

	// Group issues by severity
	errors := make([]ValidationIssue, 0)
	warnings := make([]ValidationIssue, 0)
	info := make([]ValidationIssue, 0)

	for _, issue := range report.Issues {
		switch issue.Severity {
		case ValidationError:
			errors = append(errors, issue)
		case ValidationWarning:
			warnings = append(warnings, issue)
		case ValidationInfo:
			info = append(info, issue)
		}
	}

	// Print errors first
	if len(errors) > 0 {
		fmt.Printf("❌ ERRORS (%d):\n", len(errors))
		for i, issue := range errors {
			printIssue(i+1, issue)
		}
		fmt.Println()
	}

	// Print warnings
	if len(warnings) > 0 {
		fmt.Printf("⚠️  WARNINGS (%d):\n", len(warnings))
		for i, issue := range warnings {
			printIssue(i+1, issue)
		}
		fmt.Println()
	}

	// Print info
	if len(info) > 0 {
		fmt.Printf("ℹ️  INFO (%d):\n", len(info))
		for i, issue := range info {
			printIssue(i+1, issue)
		}
	}
}

// Helper function to print individual issues
func printIssue(num int, issue ValidationIssue) {
	fmt.Printf("  %d. [%s] %s\n", num, issue.Category, issue.Message)
	if issue.File != "" {
		fmt.Printf("     File: %s", issue.File)
		if issue.Line > 0 {
			fmt.Printf(":%d", issue.Line)
		}
		fmt.Println()
	}
	if issue.Field != "" && issue.Value != "" {
		fmt.Printf("     Field: %s = %s\n", issue.Field, issue.Value)
	}
	if issue.Context != "" {
		fmt.Printf("     Context: %s\n", issue.Context)
	}
	if issue.Suggestion != "" {
		fmt.Printf("     Suggestion: %s\n", issue.Suggestion)
	}
	fmt.Println()
}

// Validation Rules Implementation

// TechTreeValidationRule validates tech tree structure and references
type TechTreeValidationRule struct{}

func (r *TechTreeValidationRule) Name() string        { return "Tech Tree Validation" }
func (r *TechTreeValidationRule) Description() string { return "Validates tech tree structure and cross-references" }

func (r *TechTreeValidationRule) Validate(validator *DataValidator, report *ValidationReport) error {
	if validator.techTree == nil {
		return fmt.Errorf("tech tree not loaded")
	}

	// Validate attack types
	if len(validator.techTree.AttackTypes) == 0 {
		validator.addIssue(report, ValidationWarning, "Tech Tree Structure",
			"No attack types defined", "megapack.xml", 0, "attack-types", "", "",
			"Define at least one attack type")
	}

	// Validate armor types
	if len(validator.techTree.ArmorTypes) == 0 {
		validator.addIssue(report, ValidationWarning, "Tech Tree Structure",
			"No armor types defined", "megapack.xml", 0, "armor-types", "", "",
			"Define at least one armor type")
	}

	// Validate damage multipliers
	validator.validateDamageMultipliers(report)

	// Check for duplicate attack type names
	attackNames := make(map[string]bool)
	for _, at := range validator.techTree.AttackTypes {
		if attackNames[at.Name] {
			validator.addIssue(report, ValidationError, "Data Consistency",
				fmt.Sprintf("Duplicate attack type name: %s", at.Name),
				"megapack.xml", 0, "attack-type", at.Name, "",
				"Use unique names for attack types")
		}
		attackNames[at.Name] = true
	}

	// Check for duplicate armor type names
	armorNames := make(map[string]bool)
	for _, at := range validator.techTree.ArmorTypes {
		if armorNames[at.Name] {
			validator.addIssue(report, ValidationError, "Data Consistency",
				fmt.Sprintf("Duplicate armor type name: %s", at.Name),
				"megapack.xml", 0, "armor-type", at.Name, "",
				"Use unique names for armor types")
		}
		armorNames[at.Name] = true
	}

	return nil
}

// ResourceValidationRule validates resource definitions
type ResourceValidationRule struct{}

func (r *ResourceValidationRule) Name() string        { return "Resource Validation" }
func (r *ResourceValidationRule) Description() string { return "Validates resource definitions and properties" }

func (r *ResourceValidationRule) Validate(validator *DataValidator, report *ValidationReport) error {
	if validator.resources == nil {
		return fmt.Errorf("resources not loaded")
	}

	if len(validator.resources) == 0 {
		validator.addIssue(report, ValidationError, "Resource Structure",
			"No resources defined", "resources directory", 0, "resources", "", "",
			"Define at least one resource")
		return nil
	}

	// Check for duplicate resource names
	resourceNames := make(map[string]bool)
	for _, res := range validator.resources {
		if resourceNames[res.Name] {
			validator.addIssue(report, ValidationError, "Data Consistency",
				fmt.Sprintf("Duplicate resource name: %s", res.Name),
				fmt.Sprintf("resources/%s.xml", res.Name), 0, "resource", res.Name, "",
				"Use unique names for resources")
		}
		resourceNames[res.Name] = true

		// Validate resource properties
		if res.Resource.Type.Value == "tech" && res.Resource.Type.DefaultAmount != nil && res.Resource.Type.DefaultAmount.Value < 0 {
			validator.addIssue(report, ValidationWarning, "Resource Properties",
				fmt.Sprintf("Tech resource '%s' has negative default value", res.Name),
				fmt.Sprintf("resources/%s.xml", res.Name), 0, "default-amount", fmt.Sprintf("%d", res.Resource.Type.DefaultAmount.Value), "",
				"Use non-negative default values for tech resources")
		}
	}

	// Check for essential resources
	essentialResources := []string{"gold", "wood", "stone", "energy"}
	for _, essential := range essentialResources {
		found := false
		for _, res := range validator.resources {
			if res.Name == essential {
				found = true
				break
			}
		}
		if !found {
			validator.addIssue(report, ValidationWarning, "Resource Completeness",
				fmt.Sprintf("Essential resource '%s' not found", essential),
				"resources directory", 0, "resource", essential, "",
				fmt.Sprintf("Consider adding %s resource definition", essential))
		}
	}

	return nil
}

// FactionValidationRule validates faction data and references
type FactionValidationRule struct{}

func (r *FactionValidationRule) Name() string        { return "Faction Validation" }
func (r *FactionValidationRule) Description() string { return "Validates faction definitions and unit references" }

func (r *FactionValidationRule) Validate(validator *DataValidator, report *ValidationReport) error {
	if validator.factions == nil {
		return fmt.Errorf("factions not loaded")
	}

	if len(validator.factions) == 0 {
		validator.addIssue(report, ValidationError, "Faction Structure",
			"No factions defined", "factions directory", 0, "factions", "", "",
			"Define at least one faction")
		return nil
	}

	// Validate each faction
	for _, faction := range validator.factions {
		validator.validateFactionStartingResources(&faction, report)
		validator.validateFactionStartingUnits(&faction, report)

		// Check faction has at least one starting unit
		if len(faction.Faction.StartingUnits) == 0 {
			validator.addIssue(report, ValidationWarning, "Faction Balance",
				fmt.Sprintf("Faction '%s' has no starting units", faction.Name),
				fmt.Sprintf("factions/%s/%s.xml", faction.Name, faction.Name), 0,
				"starting-units", "", "", "Add at least one starting unit")
		}

		// Check faction has starting resources
		if len(faction.Faction.StartingResources) == 0 {
			validator.addIssue(report, ValidationWarning, "Faction Balance",
				fmt.Sprintf("Faction '%s' has no starting resources", faction.Name),
				fmt.Sprintf("factions/%s/%s.xml", faction.Name, faction.Name), 0,
				"starting-resources", "", "", "Add starting resource amounts")
		}
	}

	// Check for duplicate faction names
	factionNames := make(map[string]bool)
	for _, faction := range validator.factions {
		if factionNames[faction.Name] {
			validator.addIssue(report, ValidationError, "Data Consistency",
				fmt.Sprintf("Duplicate faction name: %s", faction.Name),
				fmt.Sprintf("factions/%s/%s.xml", faction.Name, faction.Name), 0,
				"faction", faction.Name, "", "Use unique names for factions")
		}
		factionNames[faction.Name] = true
	}

	return nil
}

// AssetExistenceRule validates that referenced asset files exist
type AssetExistenceRule struct{}

func (r *AssetExistenceRule) Name() string        { return "Asset Existence Validation" }
func (r *AssetExistenceRule) Description() string { return "Validates that all referenced asset files exist on disk" }

func (r *AssetExistenceRule) Validate(validator *DataValidator, report *ValidationReport) error {
	if !validator.enableFileChecks {
		validator.addIssue(report, ValidationInfo, "Asset Validation",
			"Asset file existence checks are disabled", "", 0, "config", "enableFileChecks", "false",
			"Enable file checks for complete validation")
		return nil
	}

	if validator.factions == nil {
		return fmt.Errorf("factions not loaded for asset validation")
	}

	// Check assets for each faction
	for _, faction := range validator.factions {
		validator.validateFactionAssets(&faction, report)
	}

	// Validate tech tree file exists
	techTreePath := filepath.Join(validator.techTreeRoot, "megapack.xml")
	if _, err := os.Stat(techTreePath); os.IsNotExist(err) {
		validator.addIssue(report, ValidationError, "Asset Missing",
			"Tech tree file not found", "megapack.xml", 0, "file", "megapack.xml", "",
			"Ensure megapack.xml exists in tech tree root")
	}

	// Validate resources directory exists
	resourcesPath := filepath.Join(validator.techTreeRoot, "resources")
	if _, err := os.Stat(resourcesPath); os.IsNotExist(err) {
		validator.addIssue(report, ValidationError, "Asset Missing",
			"Resources directory not found", "resources/", 0, "directory", "resources", "",
			"Create resources directory with resource XML files")
	}

	// Validate factions directory exists
	factionsPath := filepath.Join(validator.techTreeRoot, "factions")
	if _, err := os.Stat(factionsPath); os.IsNotExist(err) {
		validator.addIssue(report, ValidationError, "Asset Missing",
			"Factions directory not found", "factions/", 0, "directory", "factions", "",
			"Create factions directory with faction subdirectories")
	}

	return nil
}

// Get default validation rules
func getDefaultValidationRules() []ValidationRule {
	return []ValidationRule{
		&TechTreeValidationRule{},
		&ResourceValidationRule{},
		&FactionValidationRule{},
		&AssetExistenceRule{},
	}
}