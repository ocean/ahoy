package main

import (
	"slices"
	"testing"
)

func TestRunConfigValidate_ConfigNotExists(t *testing.T) {
	result := RunConfigValidate("nonexistent.ahoy.yml")

	if result.ConfigExists {
		t.Error("Expected ConfigExists to be false for nonexistent file")
	}
	if result.ConfigValid {
		t.Error("Expected ConfigValid to be false for nonexistent file")
	}
	if len(result.Recommendations) == 0 {
		t.Error("Expected recommendations for missing config file")
	}

	expectedRec := "Create a .ahoy.yml file using 'ahoy config init'"
	if !slices.Contains(result.Recommendations, expectedRec) {
		t.Errorf("Expected recommendation %q not found in: %v", expectedRec, result.Recommendations)
	}
}

func TestRunConfigValidate_InvalidYAML(t *testing.T) {
	result := RunConfigValidate("testdata/invalid-yaml.ahoy.yml")

	if !result.ConfigExists {
		t.Error("Expected ConfigExists to be true for existing file")
	}
	if result.ConfigValid {
		t.Error("Expected ConfigValid to be false for invalid YAML")
	}

	expectedRec := "Fix YAML syntax errors in configuration file"
	if !slices.Contains(result.Recommendations, expectedRec) {
		t.Errorf("Expected recommendation %q not found in: %v", expectedRec, result.Recommendations)
	}
}

func TestRunConfigValidate_ValidConfig(t *testing.T) {
	result := RunConfigValidate("testdata/simple.ahoy.yml")

	if !result.ConfigExists {
		t.Error("Expected ConfigExists to be true for existing file")
	}
	if !result.ConfigValid {
		t.Error("Expected ConfigValid to be true for valid YAML")
	}
	if result.APIVersion != "v2" {
		t.Errorf("Expected APIVersion 'v2', got %q", result.APIVersion)
	}
	if result.AhoyVersion == "" {
		t.Error("Expected AhoyVersion to be set")
	}
}

func TestRunConfigValidate_WithEnvironmentFiles(t *testing.T) {
	result := RunConfigValidate("testdata/with-env-files.ahoy.yml")

	if len(result.EnvFiles) != 3 {
		t.Errorf("Expected 3 environment files, got %d", len(result.EnvFiles))
	}

	envPaths := make(map[string]bool)
	for _, envFile := range result.EnvFiles {
		envPaths[envFile.Path] = true
	}

	for _, expected := range []string{".env", ".env.local", ".env.command"} {
		if !envPaths[expected] {
			t.Errorf("Expected to find env file %q in validation result", expected)
		}
	}
}

func TestRunConfigValidate_WithImportFiles(t *testing.T) {
	result := RunConfigValidate("testdata/with-imports.ahoy.yml")

	if len(result.ImportFiles) != 3 {
		t.Errorf("Expected 3 import files, got %d", len(result.ImportFiles))
	}

	importsByPath := make(map[string]ImportFileStatus)
	for _, importFile := range result.ImportFiles {
		importsByPath[importFile.Path] = importFile
	}

	if !importsByPath["simple.ahoy.yml"].Exists {
		t.Error("Expected simple.ahoy.yml to exist")
	}
	if importsByPath["missing-import.ahoy.yml"].Exists {
		t.Error("Expected missing-import.ahoy.yml to not exist")
	}
	if !importsByPath["another-missing.ahoy.yml"].Optional {
		t.Error("Expected another-missing.ahoy.yml to be optional")
	}
}

func TestGenerateRecommendations_VersionMismatch(t *testing.T) {
	result := ConfigReport{
		ValidationResult: ValidationResult{
			Issues: []ValidationIssue{
				{Type: "version_mismatch", Severity: "error", Message: "Version mismatch"},
			},
		},
	}

	recommendations := generateRecommendations(result)

	expectedRec := "Upgrade Ahoy to the latest version for full feature support"
	if !slices.Contains(recommendations, expectedRec) {
		t.Errorf("Expected recommendation %q not found in: %v", expectedRec, recommendations)
	}
}

func TestGenerateRecommendations_MissingImportFiles(t *testing.T) {
	result := ConfigReport{
		ImportFiles: []ImportFileStatus{
			{Path: "missing.ahoy.yml", Exists: false, Optional: false},
		},
		ValidationResult: ValidationResult{Issues: []ValidationIssue{}},
	}

	recommendations := generateRecommendations(result)

	expectedRec := "Create missing import files or mark them as optional"
	if !slices.Contains(recommendations, expectedRec) {
		t.Errorf("Expected recommendation %q not found in: %v", expectedRec, recommendations)
	}
}

func TestGenerateRecommendations_MissingEnvFiles(t *testing.T) {
	result := ConfigReport{
		EnvFiles: []EnvFileStatus{
			{Path: ".env", Exists: false},
		},
		ValidationResult: ValidationResult{Issues: []ValidationIssue{}},
	}

	recommendations := generateRecommendations(result)

	expectedRec := "Consider creating missing environment files or removing them from configuration"
	if !slices.Contains(recommendations, expectedRec) {
		t.Errorf("Expected recommendation %q not found in: %v", expectedRec, recommendations)
	}
}

func TestGenerateRecommendations_NewerFeatures(t *testing.T) {
	result := ConfigReport{
		ValidationResult: ValidationResult{
			Issues: []ValidationIssue{
				{Type: "version_mismatch", Severity: "warning", Message: "Using newer features"},
			},
		},
	}

	recommendations := generateRecommendations(result)

	expectedRec := "Consider upgrading to a newer Ahoy version for better support of advanced features"
	if !slices.Contains(recommendations, expectedRec) {
		t.Errorf("Expected recommendation %q not found in: %v", expectedRec, recommendations)
	}
}

func TestGenerateRecommendations_NoIssues(t *testing.T) {
	result := ConfigReport{
		ValidationResult: ValidationResult{Issues: []ValidationIssue{}},
		EnvFiles:         []EnvFileStatus{},
		ImportFiles:      []ImportFileStatus{},
	}

	recommendations := generateRecommendations(result)

	expectedRec := "Configuration looks good! No issues found."
	if !slices.Contains(recommendations, expectedRec) {
		t.Errorf("Expected recommendation %q not found in: %v", expectedRec, recommendations)
	}
}

func TestCheckEnvironmentFiles(t *testing.T) {
	// Provide a config path in testdata so expandPath() resolves correctly.
	configFile := "testdata/with-env-files.ahoy.yml"

	config := Config{
		Env: StringArray{".env.test", ".env.missing"},
		Commands: map[string]Command{
			"test": {Env: StringArray{".env.command"}},
		},
	}

	envFiles := checkEnvironmentFiles(config, configFile)

	if len(envFiles) != 3 {
		t.Errorf("Expected 3 environment files, got %d", len(envFiles))
	}

	globalCount := 0
	for _, envFile := range envFiles {
		if envFile.Global {
			globalCount++
		}
	}
	if globalCount != 2 {
		t.Errorf("Expected 2 global environment files, got %d", globalCount)
	}

	// testdata/.env.test exists.
	found := false
	for _, envFile := range envFiles {
		if envFile.Path == ".env.test" && envFile.Exists {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected .env.test to be found and exist")
	}
}

func TestCheckImportFiles(t *testing.T) {
	configFile := "testdata/with-imports.ahoy.yml"

	config := Config{
		Commands: map[string]Command{
			"test1": {Imports: []string{"simple.ahoy.yml", "missing-import.ahoy.yml"}},
			"test2": {Imports: []string{"another-missing.ahoy.yml"}, Optional: true},
		},
	}

	importFiles := checkImportFiles(config, configFile)

	if len(importFiles) != 3 {
		t.Errorf("Expected 3 import files, got %d", len(importFiles))
	}

	// simple.ahoy.yml should exist in testdata/.
	found := false
	for _, importFile := range importFiles {
		if importFile.Path == "simple.ahoy.yml" && importFile.Exists {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected simple.ahoy.yml to exist")
	}

	// another-missing.ahoy.yml should be optional.
	found = false
	for _, importFile := range importFiles {
		if importFile.Path == "another-missing.ahoy.yml" && importFile.Optional {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected another-missing.ahoy.yml to be marked as optional")
	}
}

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		v1, v2   string
		expected int
	}{
		{"v2.1.0", "v2.0.0", 1},
		{"v2.0.0", "v2.1.0", -1},
		{"v2.1.0", "v2.1.0", 0},
		{"v2.1.0", "v2.1.1", -1},
		{"v3.0.0", "v2.9.9", 1},
		{"v2.1.0-alpha", "v2.1.0", -1},
		{"v2.1.0", "v2.1.0-alpha", 1},
	}

	for _, tt := range tests {
		result := compareVersions(tt.v1, tt.v2)
		if result != tt.expected {
			t.Errorf("compareVersions(%q, %q) = %d, want %d", tt.v1, tt.v2, result, tt.expected)
		}
	}
}

func TestVersionSupports(t *testing.T) {
	// Development version should support all features.
	if !VersionSupports("development", "command_aliases") {
		t.Error("development version should support all features")
	}

	// Version that supports the feature.
	if !VersionSupports("v2.2.0", "optional_imports") {
		t.Error("v2.2.0 should support optional_imports")
	}

	// Version that doesn't support the feature.
	if VersionSupports("v2.0.0", "optional_imports") {
		t.Error("v2.0.0 should not support optional_imports")
	}
}

func TestConfigReport_Fields(t *testing.T) {
	result := ConfigReport{
		ConfigFile:       "test.ahoy.yml",
		ConfigExists:     true,
		ConfigValid:      true,
		APIVersion:       "v2",
		AhoyVersion:      "v2.3.0",
		ValidationResult: ValidationResult{},
		EnvFiles:         []EnvFileStatus{},
		ImportFiles:      []ImportFileStatus{},
		Recommendations:  []string{},
	}

	if result.ConfigFile != "test.ahoy.yml" {
		t.Error("ConfigFile field not working")
	}
	if !result.ConfigExists {
		t.Error("ConfigExists field not working")
	}
	if !result.ConfigValid {
		t.Error("ConfigValid field not working")
	}
	if result.APIVersion != "v2" {
		t.Error("APIVersion field not working")
	}
	if result.AhoyVersion != "v2.3.0" {
		t.Error("AhoyVersion field not working")
	}
}
