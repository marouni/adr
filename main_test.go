package main

import (
	"context"
	"encoding/json"
	"os"
	// "io/ioutil" // No longer needed directly
	// "os/user" // No longer needed directly by test setup
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/urfave/cli/v3"
)

// testMainSetup modifies the global pathCfg to use temporary paths for testing.
// It returns the temporary home directory path and the original PathConfig to be restored later.
func testMainSetup(t *testing.T) (string, PathConfig) {
	originalGlobalPathCfg := *pathCfg // Save a copy of the original pathCfg values

	tempTestHome, err := os.MkdirTemp("", "adr-test-home-") // Use os.MkdirTemp
	if err != nil {
		t.Fatalf("Failed to create temp test home dir: %v", err)
	}

	// Create a new PathConfig pointing to the temporary directory
	testSpecificPathCfg := PathConfig{
		ConfigFolderName:  ".adr",
		ConfigFileName:    "config.json",
		TemplateFileName:  "template.md",
		UserHomeDir:       tempTestHome, // This field is mostly for reference within PathConfig itself
		ConfigFolderPath:  filepath.Join(tempTestHome, ".adr"),
		ConfigFilePath:    filepath.Join(tempTestHome, ".adr", "config.json"),
		TemplateFilePath:  filepath.Join(tempTestHome, ".adr", "template.md"),
		DefaultBaseFolder: filepath.Join(tempTestHome, "adr"), // Default ADRs storage
	}

	*pathCfg = testSpecificPathCfg // Point the global pathCfg to our test version

	return tempTestHome, originalGlobalPathCfg
}

// testMainTeardown restores the original pathCfg and removes the temporary directory.
func testMainTeardown(t *testing.T, tempTestHome string, originalGlobalPathCfg PathConfig) {
	*pathCfg = originalGlobalPathCfg // Restore original pathCfg

	if err := os.RemoveAll(tempTestHome); err != nil {
		t.Logf("Warning: failed to remove temp test home dir %s: %v", tempTestHome, err)
	}
}

// Helper to run the CLI app with specific arguments
func runApp(args []string) error {
	// In urfave/cli v3, the root is typically a Command.
	cmd := &cli.Command{
		Name:    "adr",
		Usage:   "A simple CLI to manage Architecture Decision Records",
		Version: "0.2.0", // Ensure this matches or is irrelevant to test logic
		Commands: []*cli.Command{ // Correct field for subcommands in v3 is Commands
			&InitCmd, // Commands are now pointers
			&NewCmd,
		},
	}
	// The Run method for a command takes a context and arguments.
	// os.Args from the actual execution needs to be simulated here.
	// The first element of `args` is typically the program name.
	return cmd.Run(context.Background(), args)
}

// TestInitCommandDefault tests the 'init' command with default settings.
func TestInitCommandDefault(t *testing.T) {
	tempHome, originalCfg := testMainSetup(t) // Use new setup
	defer testMainTeardown(t, tempHome, originalCfg) // Use new teardown

	args := []string{"adr", "init"}
	err := runApp(args)
	if err != nil {
		t.Fatalf("init command failed: %v", err)
	}

	// 1. Test that it creates the base directory (using pathCfg)
	if _, err := os.Stat(pathCfg.DefaultBaseFolder); os.IsNotExist(err) {
		t.Errorf("Default base directory %s was not created", pathCfg.DefaultBaseFolder)
	}

	// 2. Test that it creates the configuration file (using pathCfg)
	if _, err := os.Stat(pathCfg.ConfigFilePath); os.IsNotExist(err) {
		t.Fatalf("Config file %s was not created", pathCfg.ConfigFilePath)
	}

	// Verify config.json content
	var config AdrConfig
	configBytes, _ := os.ReadFile(pathCfg.ConfigFilePath) // Use os.ReadFile
	json.Unmarshal(configBytes, &config)
	if config.BaseDir != pathCfg.DefaultBaseFolder {
		t.Errorf("Expected BaseDir in config to be %s, got %s", pathCfg.DefaultBaseFolder, config.BaseDir)
	}
	if config.CurrentAdr != 0 {
		t.Errorf("Expected CurrentAdr in config to be 0, got %d", config.CurrentAdr)
	}

	// 3. Test that it creates the template file (using pathCfg)
	if _, err := os.Stat(pathCfg.TemplateFilePath); os.IsNotExist(err) {
		t.Fatalf("Template file %s was not created", pathCfg.TemplateFilePath)
	}
	// Verify template.md content
	content, _ := os.ReadFile(pathCfg.TemplateFilePath) // Use os.ReadFile
	expectedTemplateContent := `
# {{.Number}}. {{.Title}}
======
Date: {{.Date}}

## Status
======
{{.Status}}

## Context
======

## Decision
======

## Consequences
======

`
	if string(content) != expectedTemplateContent {
		t.Errorf("Template content mismatch. Expected:\n%s\nGot:\n%s", expectedTemplateContent, string(content))
	}
}

// TestInitCommandWithArg tests the 'init' command with a base directory argument.
func TestInitCommandWithArg(t *testing.T) {
	tempHome, originalCfg := testMainSetup(t) // Use new setup
	defer testMainTeardown(t, tempHome, originalCfg) // Use new teardown

	customBaseDir := filepath.Join(tempHome, "my_custom_adrs") // Ensure custom path is within tempHome
	args := []string{"adr", "init", customBaseDir}
	err := runApp(args)
	if err != nil {
		t.Fatalf("init command with argument failed: %v", err)
	}

	// 1. Test that it creates the custom base directory
	if _, err := os.Stat(customBaseDir); os.IsNotExist(err) {
		t.Errorf("Custom base directory %s was not created", customBaseDir)
	}

	// 2. Test that it creates the configuration file (using pathCfg)
	if _, err := os.Stat(pathCfg.ConfigFilePath); os.IsNotExist(err) {
		t.Fatalf("Config file %s was not created", pathCfg.ConfigFilePath)
	}

	// Verify config.json content
	var config AdrConfig
	configBytes, _ := os.ReadFile(pathCfg.ConfigFilePath) // Use os.ReadFile
	json.Unmarshal(configBytes, &config)
	if config.BaseDir != customBaseDir {
		t.Errorf("Expected BaseDir in config to be %s, got %s", customBaseDir, config.BaseDir)
	}
	if config.CurrentAdr != 0 {
		t.Errorf("Expected CurrentAdr in config to be 0, got %d", config.CurrentAdr)
	}

	// 3. Test that it creates the template file (using pathCfg)
	if _, err := os.Stat(pathCfg.TemplateFilePath); os.IsNotExist(err) {
		t.Fatalf("Template file %s was not created", pathCfg.TemplateFilePath)
	}
}

// TestInitCommandBaseDirExists tests the 'init' command when the base directory already exists.
func TestInitCommandBaseDirExists(t *testing.T) {
	tempHome, originalCfg := testMainSetup(t) // Use new setup
	defer testMainTeardown(t, tempHome, originalCfg) // Use new teardown

	// Create the default base directory beforehand (using pathCfg)
	if err := os.MkdirAll(pathCfg.DefaultBaseFolder, 0755); err != nil {
		t.Fatalf("Failed to pre-create base directory: %v", err)
	}

	args := []string{"adr", "init"}
	err := runApp(args) // Should not fail, but print a message (can't test stdout easily here)
	if err != nil {
		t.Fatalf("init command failed when base directory already exists: %v", err)
	}

	// Check config and template are still created (using pathCfg)
	if _, err := os.Stat(pathCfg.ConfigFilePath); os.IsNotExist(err) {
		t.Fatalf("Config file %s was not created when base dir existed", pathCfg.ConfigFilePath)
	}
	if _, err := os.Stat(pathCfg.TemplateFilePath); os.IsNotExist(err) {
		t.Fatalf("Template file %s was not created when base dir existed", pathCfg.TemplateFilePath)
	}
}

// TestNewCommand tests the 'new' command.
func TestNewCommand(t *testing.T) {
	tempHome, originalCfg := testMainSetup(t) // Use new setup
	defer testMainTeardown(t, tempHome, originalCfg) // Use new teardown

	// 1. Initialize ADR first
	initArgs := []string{"adr", "init"}
	if err := runApp(initArgs); err != nil {
		t.Fatalf("Prerequisite init command failed: %v", err)
	}

	// 2. Run the 'new' command
	adrTitle := "My First ADR"
	newArgs := []string{"adr", "new", adrTitle}
	if err := runApp(newArgs); err != nil {
		t.Fatalf("new command failed: %v", err)
	}

	// 3. Verify ADR file creation
	// Config would have been created with CurrentAdr = 0 by init.
	// newAdr increments it *before* creating the file in the actual code,
	// so the first ADR is 1.
	// However, the `new` command in `commands.go` fetches config, increments, *then* calls `newAdr` which uses the passed (already incremented) config.
	// Let's re-check commands.go newCmd logic:
	// currentConfig := getConfig()
	// currentConfig.CurrentAdr++ -> current_id becomes 1
	// updateConfig(currentConfig) -> config file now has current_id = 1
	// newAdr(currentConfig, c.Args()) -> newAdr receives config with CurrentAdr = 1, creates file 1-...
	// This means the first ADR created will be numbered 1.

	expectedAdrNumber := 1
	// Corrected filename expectation: ADR title parts are not lowercased by newAdr function.
	expectedFileName := strconv.Itoa(expectedAdrNumber) + "-" + strings.ReplaceAll(adrTitle, " ", "-") + ".md"
	// ADRs are created in the DefaultBaseFolder defined by pathCfg
	expectedFilePath := filepath.Join(pathCfg.DefaultBaseFolder, expectedFileName)

	if _, err := os.Stat(expectedFilePath); os.IsNotExist(err) {
		t.Fatalf("New ADR file %s was not created. Content of %s: %v", expectedFilePath, pathCfg.DefaultBaseFolder, listDir(t, pathCfg.DefaultBaseFolder))
	}

	// 4. Verify ADR file content (basic check based on template)
	content, err := os.ReadFile(expectedFilePath) // Use os.ReadFile
	if err != nil {
		t.Fatalf("Failed to read new ADR file: %v", err)
	}
	if !strings.Contains(string(content), "# "+strconv.Itoa(expectedAdrNumber)+". "+adrTitle) {
		t.Errorf("ADR file does not contain correct title string. Got: %s", string(content))
	}
	if !strings.Contains(string(content), "## Status\n======\nProposed") {
		t.Errorf("ADR file does not contain correct status string. Got: %s", string(content))
	}

	// 5. Verify config.json is updated (CurrentAdr incremented)
	var config AdrConfig
	configBytes, _ := os.ReadFile(pathCfg.ConfigFilePath) // Use os.ReadFile
	json.Unmarshal(configBytes, &config)
	if config.CurrentAdr != expectedAdrNumber { // After creating ADR #1, CurrentAdr should be 1
		t.Errorf("Expected CurrentAdr in config to be %d, got %d", expectedAdrNumber, config.CurrentAdr)
	}

	// 6. Create a second ADR to ensure increment logic is correct
	adrTitle2 := "My Second ADR"
	newArgs2 := []string{"adr", "new", adrTitle2}
	if err := runApp(newArgs2); err != nil {
		t.Fatalf("second new command failed: %v", err)
	}

	expectedAdrNumber2 := 2
	// Corrected filename expectation for second ADR
	expectedFileName2 := strconv.Itoa(expectedAdrNumber2) + "-" + strings.ReplaceAll(adrTitle2, " ", "-") + ".md"
	expectedFilePath2 := filepath.Join(pathCfg.DefaultBaseFolder, expectedFileName2)

	if _, err := os.Stat(expectedFilePath2); os.IsNotExist(err) {
		t.Fatalf("Second New ADR file %s was not created. Content of %s: %v", expectedFilePath2, pathCfg.DefaultBaseFolder, listDir(t, pathCfg.DefaultBaseFolder))
	}
	
	configBytes2, _ := os.ReadFile(pathCfg.ConfigFilePath) // Use os.ReadFile
	json.Unmarshal(configBytes2, &config)
	if config.CurrentAdr != expectedAdrNumber2 {
		t.Errorf("Expected CurrentAdr in config to be %d after second ADR, got %d", expectedAdrNumber2, config.CurrentAdr)
	}
}

// listDir is a helper to list directory contents for debugging.
func listDir(t *testing.T, dir string) []string {
	dirEntries, err := os.ReadDir(dir) // Changed from ioutil.ReadDir to os.ReadDir
	if err != nil {
		t.Logf("Error listing directory %s: %v", dir, err)
		return nil
	}
	var names []string
	for _, de := range dirEntries { // Iterate over os.DirEntry
		names = append(names, de.Name())
	}
	return names
}

// TestNewCommandBeforeInit tests that 'new' command fails if 'init' was not run.
func TestNewCommandBeforeInit(t *testing.T) {
	tempHome, originalCfg := testMainSetup(t)
	defer testMainTeardown(t, tempHome, originalCfg)

	// Do not run 'init' command.

	adrTitle := "Should Fail ADR"
	newArgs := []string{"adr", "new", adrTitle}
	err := runApp(newArgs)

	if err == nil {
		t.Fatalf("'new' command should have failed because 'init' was not run, but it succeeded.")
	}
	// As per commands.go, NewCmd returns the error from getConfig() if config is not found.
	// The error from getConfig will be an os.PathError if the file doesn't exist.
	// We can check for this. The error message "No ADR configuration is found!" is printed by
	// commands.go, but the actual error returned is from os.ReadFile.
	if !os.IsNotExist(err) {
		t.Logf("Received error: %v. Type: %T", err, err)
		// This check is important. If it's not an IsNotExist error, it might be a different problem.
		// For example, if getConfig returned a different error type, this test might still pass
		// the err == nil check but for the wrong reason.
		t.Errorf("Expected a file not found error (os.IsNotExist), but got a different error type.")
	}
}

// TODO: Test 'new' command with multi-word title arguments. (Covered by current TestNewCommand)
// TODO: Test edge cases for file system permissions (harder to test reliably in unit/integration tests).
