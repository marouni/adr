package main

import (
	"context" // Correctly placed context import
	"encoding/json"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/urfave/cli/v3"
)

var originalHomeDir string
var tempHomeDir string

// testMainSetup sets up a temporary home directory for tests.
func testMainSetup(t *testing.T) {
	var err error
	originalHomeDir = usr.HomeDir // Save original user's home directory
	tempHomeDir, err = ioutil.TempDir("", "adr-test-home")
	if err != nil {
		t.Fatalf("Failed to create temp home dir: %v", err)
	}
	usr.HomeDir = tempHomeDir // Override user's home directory for this test

	// Update path variables that depend on usr.HomeDir
	adrConfigFolderName = ".adr"
	adrConfigFileName = "config.json"
	adrConfigTemplateName = "template.md"
	adrConfigFolderPath = filepath.Join(usr.HomeDir, adrConfigFolderName)
	adrConfigFilePath = filepath.Join(adrConfigFolderPath, adrConfigFileName)
	adrTemplateFilePath = filepath.Join(adrConfigFolderPath, adrConfigTemplateName)
	adrDefaultBaseFolder = filepath.Join(usr.HomeDir, "adr")
}

// testMainTeardown cleans up the temporary home directory.
func testMainTeardown(t *testing.T) {
	if err := os.RemoveAll(tempHomeDir); err != nil {
		t.Logf("Warning: failed to remove temp home dir %s: %v", tempHomeDir, err)
	}
	usr.HomeDir = originalHomeDir // Restore original user's home directory

	// Restore original path variables
	usr, _ = user.Current() // Re-fetch current user to reset HomeDir correctly for subsequent package-level var initializations if any
	adrConfigFolderPath = filepath.Join(usr.HomeDir, adrConfigFolderName)
	adrConfigFilePath = filepath.Join(adrConfigFolderPath, adrConfigFileName)
	adrTemplateFilePath = filepath.Join(adrConfigFolderPath, adrConfigTemplateName)
	adrDefaultBaseFolder = filepath.Join(usr.HomeDir, "adr")
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
	testMainSetup(t)
	defer testMainTeardown(t)

	args := []string{"adr", "init"}
	err := runApp(args)
	if err != nil {
		t.Fatalf("init command failed: %v", err)
	}

	// 1. Test that it creates the base directory (adrDefaultBaseFolder)
	if _, err := os.Stat(adrDefaultBaseFolder); os.IsNotExist(err) {
		t.Errorf("Default base directory %s was not created", adrDefaultBaseFolder)
	}

	// 2. Test that it creates the configuration file (adrConfigFilePath)
	if _, err := os.Stat(adrConfigFilePath); os.IsNotExist(err) {
		t.Fatalf("Config file %s was not created", adrConfigFilePath)
	}

	// Verify config.json content
	var config AdrConfig
	configBytes, _ := ioutil.ReadFile(adrConfigFilePath)
	json.Unmarshal(configBytes, &config)
	if config.BaseDir != adrDefaultBaseFolder {
		t.Errorf("Expected BaseDir in config to be %s, got %s", adrDefaultBaseFolder, config.BaseDir)
	}
	if config.CurrentAdr != 0 {
		t.Errorf("Expected CurrentAdr in config to be 0, got %d", config.CurrentAdr)
	}

	// 3. Test that it creates the template file (adrTemplateFilePath)
	if _, err := os.Stat(adrTemplateFilePath); os.IsNotExist(err) {
		t.Fatalf("Template file %s was not created", adrTemplateFilePath)
	}
	// Verify template.md content
	content, _ := ioutil.ReadFile(adrTemplateFilePath)
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
	testMainSetup(t)
	defer testMainTeardown(t)

	customBaseDir := filepath.Join(tempHomeDir, "my_custom_adrs")
	args := []string{"adr", "init", customBaseDir}
	err := runApp(args)
	if err != nil {
		t.Fatalf("init command with argument failed: %v", err)
	}

	// 1. Test that it creates the custom base directory
	if _, err := os.Stat(customBaseDir); os.IsNotExist(err) {
		t.Errorf("Custom base directory %s was not created", customBaseDir)
	}

	// 2. Test that it creates the configuration file (adrConfigFilePath)
	if _, err := os.Stat(adrConfigFilePath); os.IsNotExist(err) {
		t.Fatalf("Config file %s was not created", adrConfigFilePath)
	}

	// Verify config.json content
	var config AdrConfig
	configBytes, _ := ioutil.ReadFile(adrConfigFilePath)
	json.Unmarshal(configBytes, &config)
	if config.BaseDir != customBaseDir {
		t.Errorf("Expected BaseDir in config to be %s, got %s", customBaseDir, config.BaseDir)
	}
	if config.CurrentAdr != 0 {
		t.Errorf("Expected CurrentAdr in config to be 0, got %d", config.CurrentAdr)
	}

	// 3. Test that it creates the template file (adrTemplateFilePath) - should still be in ~/.adr
	if _, err := os.Stat(adrTemplateFilePath); os.IsNotExist(err) {
		t.Fatalf("Template file %s was not created", adrTemplateFilePath)
	}
}

// TestInitCommandBaseDirExists tests the 'init' command when the base directory already exists.
func TestInitCommandBaseDirExists(t *testing.T) {
	testMainSetup(t)
	defer testMainTeardown(t)

	// Create the default base directory beforehand
	if err := os.MkdirAll(adrDefaultBaseFolder, 0755); err != nil {
		t.Fatalf("Failed to pre-create base directory: %v", err)
	}

	args := []string{"adr", "init"}
	err := runApp(args) // Should not fail, but print a message (can't test stdout easily here)
	if err != nil {
		t.Fatalf("init command failed when base directory already exists: %v", err)
	}

	// Check config and template are still created
	if _, err := os.Stat(adrConfigFilePath); os.IsNotExist(err) {
		t.Fatalf("Config file %s was not created when base dir existed", adrConfigFilePath)
	}
	if _, err := os.Stat(adrTemplateFilePath); os.IsNotExist(err) {
		t.Fatalf("Template file %s was not created when base dir existed", adrTemplateFilePath)
	}
}

// TestNewCommand tests the 'new' command.
func TestNewCommand(t *testing.T) {
	testMainSetup(t)
	defer testMainTeardown(t)

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
	// adrDefaultBaseFolder is already set up by testMainSetup to be tempHomeDir/adr
	expectedFilePath := filepath.Join(adrDefaultBaseFolder, expectedFileName)

	if _, err := os.Stat(expectedFilePath); os.IsNotExist(err) {
		t.Fatalf("New ADR file %s was not created. Content of %s: %v", expectedFilePath, adrDefaultBaseFolder, listDir(t, adrDefaultBaseFolder))
	}

	// 4. Verify ADR file content (basic check based on template)
	content, err := ioutil.ReadFile(expectedFilePath)
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
	configBytes, _ := ioutil.ReadFile(adrConfigFilePath)
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
	expectedFilePath2 := filepath.Join(adrDefaultBaseFolder, expectedFileName2)

	if _, err := os.Stat(expectedFilePath2); os.IsNotExist(err) {
		t.Fatalf("Second New ADR file %s was not created. Content of %s: %v", expectedFilePath2, adrDefaultBaseFolder, listDir(t, adrDefaultBaseFolder))
	}
	
	configBytes2, _ := ioutil.ReadFile(adrConfigFilePath)
	json.Unmarshal(configBytes2, &config)
	if config.CurrentAdr != expectedAdrNumber2 {
		t.Errorf("Expected CurrentAdr in config to be %d after second ADR, got %d", expectedAdrNumber2, config.CurrentAdr)
	}
}

// listDir is a helper to list directory contents for debugging.
func listDir(t *testing.T, dir string) []string {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		t.Logf("Error listing directory %s: %v", dir, err)
		return nil
	}
	var names []string
	for _, f := range files {
		names = append(names, f.Name())
	}
	return names
}

// TODO: Add more tests, e.g. for 'new' command when init has not been run (should fail gracefully).
// TODO: Test 'new' command with multi-word title arguments. (Covered by current TestNewCommand)
// TODO: Test edge cases for file system permissions (harder to test reliably in unit/integration tests).
