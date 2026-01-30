package data

import (
	"testing"
)

func TestNewDataValidator(t *testing.T) {
	techTreeRoot := "../../megaglest-source/data/glest_game/techs/megapack"
	assetManager := NewAssetManager(techTreeRoot)
	validator := NewDataValidator(techTreeRoot, assetManager)

	if validator == nil {
		t.Error("NewDataValidator returned nil")
	}

	if validator.techTreeRoot != techTreeRoot {
		t.Errorf("Expected techTreeRoot %s, got %s", techTreeRoot, validator.techTreeRoot)
	}

	if validator.assetManager != assetManager {
		t.Error("AssetManager not set correctly")
	}

	if !validator.enableFileChecks {
		t.Error("File checks should be enabled by default")
	}

	if len(validator.validationRules) == 0 {
		t.Error("No validation rules loaded")
	}
}

func TestValidationSeverityString(t *testing.T) {
	tests := []struct {
		severity ValidationSeverity
		expected string
	}{
		{ValidationError, "ERROR"},
		{ValidationWarning, "WARNING"},
		{ValidationInfo, "INFO"},
	}

	for _, test := range tests {
		result := test.severity.String()
		if result != test.expected {
			t.Errorf("Expected %s, got %s", test.expected, result)
		}
	}
}

func TestValidateTechTree(t *testing.T) {
	techTreeRoot := "../../megaglest-source/data/glest_game/techs/megapack"
	assetManager := NewAssetManager(techTreeRoot)
	validator := NewDataValidator(techTreeRoot, assetManager)

	report, err := validator.ValidateTechTree()
	if err != nil {
		t.Fatalf("ValidateTechTree failed: %v", err)
	}

	if report == nil {
		t.Error("ValidateTechTree returned nil report")
	}

	if report.FilesChecked == 0 {
		t.Error("Expected at least one file to be checked")
	}

	// Should have mostly successful validation for the real tech tree
	t.Logf("Tech tree validation: %d errors, %d warnings, %d info messages",
		report.ErrorCount, report.WarningCount, report.InfoCount)

	// Print issues for debugging
	for _, issue := range report.Issues {
		t.Logf("[%s] %s: %s", issue.Severity, issue.Category, issue.Message)
	}
}

func TestValidateFaction(t *testing.T) {
	techTreeRoot := "../../megaglest-source/data/glest_game/techs/megapack"
	assetManager := NewAssetManager(techTreeRoot)
	validator := NewDataValidator(techTreeRoot, assetManager)

	// Test validating magic faction
	report, err := validator.ValidateFaction("magic")
	if err != nil {
		t.Fatalf("ValidateFaction failed: %v", err)
	}

	if report == nil {
		t.Error("ValidateFaction returned nil report")
	}

	if report.FilesChecked == 0 {
		t.Error("Expected at least one file to be checked")
	}

	t.Logf("Magic faction validation: %d errors, %d warnings, %d info messages",
		report.ErrorCount, report.WarningCount, report.InfoCount)

	// Test validating non-existent faction
	nonExistentReport, err := validator.ValidateFaction("nonexistent")
	if err != nil {
		t.Fatalf("ValidateFaction for nonexistent faction failed: %v", err)
	}

	// Should have at least one error for faction not found
	if nonExistentReport.ErrorCount == 0 {
		t.Error("Expected validation error for nonexistent faction")
		t.Logf("Report had %d issues total", len(nonExistentReport.Issues))
		for i, issue := range nonExistentReport.Issues {
			t.Logf("Issue %d: [%s] %s - %s", i+1, issue.Severity, issue.Category, issue.Message)
		}
	} else {
		// Should find the "faction not found" error
		foundError := false
		for _, issue := range nonExistentReport.Issues {
			if issue.Severity == ValidationError && issue.Value == "nonexistent" {
				foundError = true
				break
			}
		}
		if !foundError {
			t.Error("Expected to find faction not found error")
		}
	}
}

func TestValidateAssetReferences(t *testing.T) {
	techTreeRoot := "../../megaglest-source/data/glest_game/techs/megapack"
	assetManager := NewAssetManager(techTreeRoot)
	validator := NewDataValidator(techTreeRoot, assetManager)

	report, err := validator.ValidateAssetReferences()
	if err != nil {
		t.Fatalf("ValidateAssetReferences failed: %v", err)
	}

	if report == nil {
		t.Error("ValidateAssetReferences returned nil report")
	}

	t.Logf("Asset reference validation: %d errors, %d warnings, %d info messages",
		report.ErrorCount, report.WarningCount, report.InfoCount)

	// Test with file checks disabled
	validator.enableFileChecks = false
	disabledReport, err := validator.ValidateAssetReferences()
	if err != nil {
		t.Fatalf("ValidateAssetReferences with disabled checks failed: %v", err)
	}

	// Should have info message about disabled checks
	foundInfo := false
	for _, issue := range disabledReport.Issues {
		if issue.Severity == ValidationInfo && issue.Category == "Asset Validation" {
			foundInfo = true
			break
		}
	}
	if !foundInfo {
		t.Error("Expected info message about disabled asset checks")
	}
}

func TestValidateAllData(t *testing.T) {
	techTreeRoot := "../../megaglest-source/data/glest_game/techs/megapack"
	assetManager := NewAssetManager(techTreeRoot)
	validator := NewDataValidator(techTreeRoot, assetManager)

	report, err := validator.ValidateAllData()
	if err != nil {
		t.Fatalf("ValidateAllData failed: %v", err)
	}

	if report == nil {
		t.Error("ValidateAllData returned nil report")
	}

	if report.Duration <= 0 {
		t.Error("Expected positive duration for validation")
	}

	t.Logf("Complete data validation: %d errors, %d warnings, %d info messages",
		report.ErrorCount, report.WarningCount, report.InfoCount)
	t.Logf("Validation took: %v", report.Duration)

	// Validation should complete successfully even if there are issues
	if len(report.Issues) > 0 {
		t.Logf("Found %d validation issues (expected for real data)", len(report.Issues))

		// Print first few issues for debugging
		for i, issue := range report.Issues {
			if i >= 5 {
				break // Limit output
			}
			t.Logf("Issue %d: [%s] %s - %s", i+1, issue.Severity, issue.Category, issue.Message)
		}
	}
}

func TestValidationReport_PrintReport(t *testing.T) {
	// Create a test report with various issue types
	report := &ValidationReport{
		Issues: []ValidationIssue{
			{
				Severity:   ValidationError,
				Category:   "Test Error",
				Message:    "This is a test error",
				File:       "test.xml",
				Line:       10,
				Field:      "test-field",
				Value:      "test-value",
				Context:    "Test context",
				Suggestion: "Fix the test error",
			},
			{
				Severity:   ValidationWarning,
				Category:   "Test Warning",
				Message:    "This is a test warning",
				File:       "test2.xml",
				Field:      "warning-field",
				Value:      "warning-value",
				Suggestion: "Consider fixing this warning",
			},
			{
				Severity: ValidationInfo,
				Category: "Test Info",
				Message:  "This is test information",
			},
		},
		FilesChecked: 5,
		Duration:     100000, // 100ms in nanoseconds
	}

	// Count issues
	errorCount := 0
	warningCount := 0
	infoCount := 0

	for _, issue := range report.Issues {
		switch issue.Severity {
		case ValidationError:
			errorCount++
		case ValidationWarning:
			warningCount++
		case ValidationInfo:
			infoCount++
		}
	}

	report.ErrorCount = errorCount
	report.WarningCount = warningCount
	report.InfoCount = infoCount

	// This should not panic and should produce readable output
	report.PrintReport()

	// Verify counts are correct
	if report.ErrorCount != 1 {
		t.Errorf("Expected 1 error, got %d", report.ErrorCount)
	}
	if report.WarningCount != 1 {
		t.Errorf("Expected 1 warning, got %d", report.WarningCount)
	}
	if report.InfoCount != 1 {
		t.Errorf("Expected 1 info, got %d", report.InfoCount)
	}
}

func TestValidationRules(t *testing.T) {
	// Test that all default validation rules can be instantiated
	rules := getDefaultValidationRules()

	if len(rules) == 0 {
		t.Error("No default validation rules found")
	}

	expectedRules := []string{
		"Tech Tree Validation",
		"Resource Validation",
		"Faction Validation",
		"Asset Existence Validation",
	}

	for i, rule := range rules {
		if i >= len(expectedRules) {
			break
		}

		name := rule.Name()
		if name == "" {
			t.Errorf("Rule %d has empty name", i)
		}

		description := rule.Description()
		if description == "" {
			t.Errorf("Rule %d (%s) has empty description", i, name)
		}

		expectedName := expectedRules[i]
		if name != expectedName {
			t.Errorf("Expected rule name '%s', got '%s'", expectedName, name)
		}

		t.Logf("Rule %d: %s - %s", i+1, name, description)
	}
}

func TestTechTreeValidationRule(t *testing.T) {
	techTreeRoot := "../../megaglest-source/data/glest_game/techs/megapack"
	assetManager := NewAssetManager(techTreeRoot)
	validator := NewDataValidator(techTreeRoot, assetManager)

	// Load tech tree for validation
	_, err := validator.ValidateTechTree()
	if err != nil {
		t.Fatalf("Failed to load tech tree for rule testing: %v", err)
	}

	report := &ValidationReport{
		Issues: make([]ValidationIssue, 0),
	}

	rule := &TechTreeValidationRule{}
	err = rule.Validate(validator, report)
	if err != nil {
		t.Fatalf("TechTreeValidationRule failed: %v", err)
	}

	// Should validate successfully with real data
	t.Logf("Tech tree rule found %d issues", len(report.Issues))
}

func TestResourceValidationRule(t *testing.T) {
	techTreeRoot := "../../megaglest-source/data/glest_game/techs/megapack"
	assetManager := NewAssetManager(techTreeRoot)
	validator := NewDataValidator(techTreeRoot, assetManager)

	// Load resources for validation
	resources, err := assetManager.LoadResources()
	if err != nil {
		t.Fatalf("Failed to load resources for rule testing: %v", err)
	}
	validator.resources = resources

	report := &ValidationReport{
		Issues: make([]ValidationIssue, 0),
	}

	rule := &ResourceValidationRule{}
	err = rule.Validate(validator, report)
	if err != nil {
		t.Fatalf("ResourceValidationRule failed: %v", err)
	}

	t.Logf("Resource rule found %d issues", len(report.Issues))

	// Check that it validates for essential resources
	// Note: We don't check for specific missing resources because megaglest should have them all defined
}

func TestFactionValidationRule(t *testing.T) {
	techTreeRoot := "../../megaglest-source/data/glest_game/techs/megapack"
	assetManager := NewAssetManager(techTreeRoot)
	validator := NewDataValidator(techTreeRoot, assetManager)

	// Load factions for validation
	factions, err := assetManager.LoadFactions()
	if err != nil {
		t.Fatalf("Failed to load factions for rule testing: %v", err)
	}
	validator.factions = factions

	// Also load resources for cross-reference validation
	resources, err := assetManager.LoadResources()
	if err != nil {
		t.Fatalf("Failed to load resources for faction rule testing: %v", err)
	}
	validator.resources = resources

	report := &ValidationReport{
		Issues: make([]ValidationIssue, 0),
	}

	rule := &FactionValidationRule{}
	err = rule.Validate(validator, report)
	if err != nil {
		t.Fatalf("FactionValidationRule failed: %v", err)
	}

	t.Logf("Faction rule found %d issues", len(report.Issues))
}

func TestAssetExistenceRule(t *testing.T) {
	techTreeRoot := "../../megaglest-source/data/glest_game/techs/megapack"
	assetManager := NewAssetManager(techTreeRoot)
	validator := NewDataValidator(techTreeRoot, assetManager)

	// Load factions for asset validation
	factions, err := assetManager.LoadFactions()
	if err != nil {
		t.Fatalf("Failed to load factions for asset rule testing: %v", err)
	}
	validator.factions = factions

	report := &ValidationReport{
		Issues: make([]ValidationIssue, 0),
	}

	rule := &AssetExistenceRule{}
	err = rule.Validate(validator, report)
	if err != nil {
		t.Fatalf("AssetExistenceRule failed: %v", err)
	}

	t.Logf("Asset existence rule found %d issues", len(report.Issues))
}

func TestValidationWithMissingFiles(t *testing.T) {
	// Test validation behavior with missing tech tree root
	invalidRoot := "/nonexistent/path"
	assetManager := NewAssetManager(invalidRoot)
	validator := NewDataValidator(invalidRoot, assetManager)

	// This should handle errors gracefully
	report, err := validator.ValidateAllData()
	if err != nil {
		// Error is expected, but should still return a report
		if report == nil {
			t.Error("Expected report even with errors")
		}
		t.Logf("Expected error with invalid path: %v", err)
	}

	// Should have validation errors for missing files (if report exists) OR an error should have occurred
	if err == nil && (report == nil || report.ErrorCount == 0) {
		t.Error("Expected either error or validation errors for missing files")
	}
}

func TestValidationIssueFields(t *testing.T) {
	issue := ValidationIssue{
		Severity:   ValidationError,
		Category:   "Test Category",
		Message:    "Test message",
		File:       "test.xml",
		Line:       42,
		Field:      "test-field",
		Value:      "test-value",
		Context:    "Test context",
		Suggestion: "Test suggestion",
	}

	if issue.Severity != ValidationError {
		t.Error("Severity not set correctly")
	}
	if issue.Category != "Test Category" {
		t.Error("Category not set correctly")
	}
	if issue.Message != "Test message" {
		t.Error("Message not set correctly")
	}
	if issue.File != "test.xml" {
		t.Error("File not set correctly")
	}
	if issue.Line != 42 {
		t.Error("Line not set correctly")
	}
	if issue.Field != "test-field" {
		t.Error("Field not set correctly")
	}
	if issue.Value != "test-value" {
		t.Error("Value not set correctly")
	}
	if issue.Context != "Test context" {
		t.Error("Context not set correctly")
	}
	if issue.Suggestion != "Test suggestion" {
		t.Error("Suggestion not set correctly")
	}
}

func TestValidationPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	techTreeRoot := "../../megaglest-source/data/glest_game/techs/megapack"
	assetManager := NewAssetManager(techTreeRoot)
	validator := NewDataValidator(techTreeRoot, assetManager)

	// Disable file checks for faster testing
	validator.enableFileChecks = false

	report, err := validator.ValidateAllData()
	if err != nil {
		t.Fatalf("Performance test validation failed: %v", err)
	}

	// Should complete within reasonable time (less than 5 seconds for disabled file checks)
	if report.Duration.Seconds() > 5.0 {
		t.Errorf("Validation took too long: %v", report.Duration)
	}

	t.Logf("Validation performance: %v for %d files", report.Duration, report.FilesChecked)
}