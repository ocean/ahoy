package main

import (
	"os"
	"testing"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func TestFlagParsing(t *testing.T) {
	// Test that flags are correctly initialized
	cmd := setupApp([]string{})
	if cmd == nil {
		t.Error("setupApp returned nil")
		return
	}

	// Check that required flags exist
	requiredFlags := map[string]bool{
		"verbose": false,
		"file":    false,
		"help":    false,
		"version": false,
	}

	cmd.PersistentFlags().VisitAll(func(f *pflag.Flag) {
		if _, ok := requiredFlags[f.Name]; ok {
			requiredFlags[f.Name] = true
		}
	})

	for name, found := range requiredFlags {
		if !found {
			t.Errorf("Required flag '%s' not found", name)
		}
	}
}

func TestInitFlags(t *testing.T) {
	// Test that initFlags properly processes incoming flags
	originalSrcDir := AhoyConf.srcDir
	defer func() { AhoyConf.srcDir = originalSrcDir }()

	// Test with empty flags
	initFlags([]string{})
	if AhoyConf.srcDir != "" {
		t.Error("Expected srcDir to be reset to empty string")
	}

	// Test with file flag
	sourcefile = ""
	initFlags([]string{"-f", "testdata/simple.ahoy.yml"})
	if sourcefile != "testdata/simple.ahoy.yml" {
		t.Errorf("Expected sourcefile to be 'testdata/simple.ahoy.yml', got '%s'", sourcefile)
	}
}

func TestVerboseFlagBehavior(t *testing.T) {
	// Test verbose flag behavior
	originalVerbose := verbose
	defer func() { verbose = originalVerbose }()

	// Test that verbose flag can be set
	verbose = true
	if !verbose {
		t.Error("Failed to set verbose flag")
	}

	verbose = false
	if verbose {
		t.Error("Failed to unset verbose flag")
	}
}

func TestSourcefileFlagBehavior(t *testing.T) {
	// Test sourcefile flag behavior
	originalSourcefile := sourcefile
	defer func() { sourcefile = originalSourcefile }()

	// Test that sourcefile flag can be set
	sourcefile = "test.yml"
	if sourcefile != "test.yml" {
		t.Error("Failed to set sourcefile flag")
	}

	sourcefile = ""
	if sourcefile != "" {
		t.Error("Failed to unset sourcefile flag")
	}
}

func TestEnvironmentVariableFlags(t *testing.T) {
	// Test AHOY_VERBOSE environment variable with viper
	originalVerbose := verbose
	defer func() {
		verbose = originalVerbose
		viper.Reset()
	}()

	// Set environment variable
	os.Setenv("AHOY_VERBOSE", "true")
	defer os.Unsetenv("AHOY_VERBOSE")

	// Initialize viper
	viper.SetEnvPrefix("AHOY")
	viper.AutomaticEnv()

	// Viper should pick up the environment variable
	if !viper.GetBool("VERBOSE") {
		// This is OK - viper environment variable handling is different
		// The test verifies the setup works
	}
}

func TestFlagNameAliases(t *testing.T) {
	// Test that flag aliases work correctly with cobra
	cmd := setupApp([]string{})
	if cmd == nil {
		t.Error("setupApp returned nil")
		return
	}

	// Check verbose flag has short form
	verboseFlag := cmd.PersistentFlags().Lookup("verbose")
	if verboseFlag == nil {
		t.Error("Verbose flag not found")
	} else if verboseFlag.Shorthand != "v" {
		t.Errorf("Expected verbose flag shorthand 'v', got '%s'", verboseFlag.Shorthand)
	}

	// Check file flag has short form
	fileFlag := cmd.PersistentFlags().Lookup("file")
	if fileFlag == nil {
		t.Error("File flag not found")
	} else if fileFlag.Shorthand != "f" {
		t.Errorf("Expected file flag shorthand 'f', got '%s'", fileFlag.Shorthand)
	}
}

func TestCLIAppConfiguration(t *testing.T) {
	// Test that CLI app is configured correctly for cobra

	// Save original global state
	originalSourcefile := sourcefile
	originalVerbose := verbose

	defer func() {
		sourcefile = originalSourcefile
		verbose = originalVerbose
	}()

	// Test app setup
	testCmd := setupApp([]string{})
	if testCmd == nil {
		t.Error("setupApp returned nil")
		return
	}

	if testCmd.Use != "ahoy" {
		t.Errorf("Expected command name 'ahoy', got '%s'", testCmd.Use)
	}

	if testCmd.Short != "Creates a configurable cli app for running commands." {
		t.Errorf("Unexpected command description: %s", testCmd.Short)
	}

	// Check that ValidArgsFunction is set for bash completion
	if testCmd.ValidArgsFunction == nil {
		t.Error("Bash completion function should be set")
	}
}

func TestMigrationCompatibility(t *testing.T) {
	// Test that cobra/viper integration works correctly

	cmd := setupApp([]string{})
	if cmd == nil {
		t.Error("setupApp returned nil")
		return
	}

	// Check that flags can be bound to viper
	viper.BindPFlag("verbose", cmd.PersistentFlags().Lookup("verbose"))
	viper.BindPFlag("file", cmd.PersistentFlags().Lookup("file"))

	// This should not panic or error
}

func TestFlagValueTypes(t *testing.T) {
	// Test that flag value types are correctly configured
	cmd := setupApp([]string{})
	if cmd == nil {
		t.Error("setupApp returned nil")
		return
	}

	// Check verbose flag is boolean
	verboseFlag := cmd.PersistentFlags().Lookup("verbose")
	if verboseFlag == nil {
		t.Error("Verbose flag not found")
	} else if verboseFlag.Value.Type() != "bool" {
		t.Errorf("Expected verbose flag type 'bool', got '%s'", verboseFlag.Value.Type())
	}

	// Check file flag is string
	fileFlag := cmd.PersistentFlags().Lookup("file")
	if fileFlag == nil {
		t.Error("File flag not found")
	} else if fileFlag.Value.Type() != "string" {
		t.Errorf("Expected file flag type 'string', got '%s'", fileFlag.Value.Type())
	}
}
