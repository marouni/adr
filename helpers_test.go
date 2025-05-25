package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

// Helper function to create a temporary directory for testing
func tempDir(t *testing.T) string {
	dir, err := os.MkdirTemp("", "adr-test") // Changed from ioutil.TempDir
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	return dir
}

// Helper function to remove a temporary directory after testing
func removeTempDir(t *testing.T, dir string) {
	err := os.RemoveAll(dir)
	if err != nil {
		t.Fatalf("Failed to remove temp dir: %v", err)
	}
}

// Test for initBaseDir
func TestInitBaseDir(t *testing.T) {
	testDir := tempDir(t)
	defer removeTempDir(t, testDir)

	// Test creating a new directory
	newDir := filepath.Join(testDir, "newBaseDir")
	initBaseDir(newDir)
	if _, err := os.Stat(newDir); os.IsNotExist(err) {
		t.Errorf("initBaseDir failed to create directory %s", newDir)
	}

	// Test attempting to create an existing directory
	// Should not return an error, but print a message (cannot verify message here)
	initBaseDir(newDir)
	if _, err := os.Stat(newDir); os.IsNotExist(err) {
		t.Errorf("initBaseDir failed when directory %s already exists", newDir)
	}
}

// Test for initConfig
func TestInitConfig(t *testing.T) {
	originalPathCfg := *pathCfg // Dereference to copy values
	testDir := tempDir(t)
	defer removeTempDir(t, testDir)

	// Modify global pathCfg for this test
	pathCfg.ConfigFolderPath = filepath.Join(testDir, ".adr")
	pathCfg.ConfigFilePath = filepath.Join(pathCfg.ConfigFolderPath, "config.json")
	// Ensure .adr directory exists if initConfig doesn't create it (it should)
	// os.MkdirAll(pathCfg.ConfigFolderPath, 0755) // initConfig should handle this

	defer func() {
		*pathCfg = originalPathCfg // Restore original pathCfg values
	}()

	testBaseDir := filepath.Join(testDir, "adr_docs")
	err := initConfig(testBaseDir) // initConfig now returns an error
	if err != nil {
		t.Fatalf("initConfig failed: %v", err)
	}

	if _, err := os.Stat(pathCfg.ConfigFilePath); os.IsNotExist(err) {
		t.Fatalf("initConfig failed to create config file at %s", pathCfg.ConfigFilePath)
	}

	var config AdrConfig
	configBytes, err := os.ReadFile(pathCfg.ConfigFilePath) // Changed from ioutil.ReadFile
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}
	err = json.Unmarshal(configBytes, &config)
	if err != nil {
		t.Fatalf("Failed to unmarshal config: %v", err)
	}

	if config.BaseDir != testBaseDir {
		t.Errorf("Expected BaseDir to be %s, got %s", testBaseDir, config.BaseDir)
	}
	if config.CurrentAdr != 0 {
		t.Errorf("Expected CurrentAdr to be 0, got %d", config.CurrentAdr)
	}
}

// Test for initTemplate
func TestInitTemplate(t *testing.T) {
	originalPathCfg := *pathCfg // Dereference to copy values
	testDir := tempDir(t)
	defer removeTempDir(t, testDir)

	// Modify global pathCfg for this test
	pathCfg.ConfigFolderPath = filepath.Join(testDir, ".adr") // Used by initTemplate to place template.md
	pathCfg.TemplateFilePath = filepath.Join(pathCfg.ConfigFolderPath, "template.md")

	// initTemplate expects ConfigFolderPath to exist.
	// In the main code, initConfig usually creates this.
	// For this unit test, we might need to ensure it exists if initTemplate doesn't create it.
	// However, initConfig is responsible for creating ConfigFolderPath.
	// initTemplate just writes the template file into it.
	// Let's assume for this specific test, the folder might not exist if initConfig wasn't called.
	// The real application flow ensures initConfig (which creates .adr folder) is called before template generation.
	// For an isolated test of initTemplate, we should create its parent directory.
	err := os.MkdirAll(pathCfg.ConfigFolderPath, 0755)
	if err != nil {
		t.Fatalf("Failed to create .adr directory for template: %v", err)
	}


	defer func() {
		*pathCfg = originalPathCfg // Restore original pathCfg values
	}()

	err = initTemplate() // initTemplate now returns an error
	if err != nil {
		t.Fatalf("initTemplate failed: %v", err)
	}

	if _, err := os.Stat(pathCfg.TemplateFilePath); os.IsNotExist(err) {
		t.Fatalf("initTemplate failed to create template file at %s", pathCfg.TemplateFilePath)
	}

	// Verify content (basic check for now)
	content, err := os.ReadFile(pathCfg.TemplateFilePath) // Changed from ioutil.ReadFile
	if err != nil {
		t.Fatalf("Failed to read template file: %v", err)
	}
	expectedContent := `
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
	if string(content) != expectedContent {
		t.Errorf("Template content mismatch. Expected:\n%s\nGot:\n%s", expectedContent, string(content))
	}
}

// Test for updateConfig
func TestUpdateConfig(t *testing.T) {
	originalPathCfg := *pathCfg // Dereference to copy values
	testDir := tempDir(t)
	defer removeTempDir(t, testDir)

	// Modify global pathCfg for this test
	pathCfg.ConfigFolderPath = filepath.Join(testDir, ".adr")
	pathCfg.ConfigFilePath = filepath.Join(pathCfg.ConfigFolderPath, "config.json")
	// Ensure .adr directory exists
	os.MkdirAll(pathCfg.ConfigFolderPath, 0755)


	defer func() {
		*pathCfg = originalPathCfg // Restore original pathCfg values
	}()

	// Initialize a config first
	initialBaseDir := filepath.Join(testDir, "initial_docs")
	err := initConfig(initialBaseDir) // Uses modified pathCfg.ConfigFilePath
	if err != nil {
		t.Fatalf("Initial initConfig failed: %v", err)
	}


	updatedBaseDir := filepath.Join(testDir, "updated_docs")
	updatedConfigData := AdrConfig{BaseDir: updatedBaseDir, CurrentAdr: 5}
	err = updateConfig(updatedConfigData) // updateConfig now returns an error
	if err != nil {
		t.Fatalf("updateConfig failed: %v", err)
	}

	var config AdrConfig
	configBytes, err := os.ReadFile(pathCfg.ConfigFilePath) // Changed from ioutil.ReadFile
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}
	err = json.Unmarshal(configBytes, &config)
	if err != nil {
		t.Fatalf("Failed to unmarshal config: %v", err)
	}

	if config.BaseDir != updatedBaseDir {
		t.Errorf("Expected updated BaseDir to be %s, got %s", updatedBaseDir, config.BaseDir)
	}
	if config.CurrentAdr != 5 {
		t.Errorf("Expected updated CurrentAdr to be 5, got %d", config.CurrentAdr)
	}
}

// Test for getConfig
func TestGetConfig(t *testing.T) {
	originalPathCfg := *pathCfg // Dereference to copy values
	testDir := tempDir(t)
	defer removeTempDir(t, testDir)

	// Modify global pathCfg for this test
	pathCfg.ConfigFolderPath = filepath.Join(testDir, ".adr")
	pathCfg.ConfigFilePath = filepath.Join(pathCfg.ConfigFolderPath, "config.json")
	// Ensure .adr directory exists
	os.MkdirAll(pathCfg.ConfigFolderPath, 0755)


	defer func() {
		*pathCfg = originalPathCfg // Restore original pathCfg values
	}()

	// Test case 1: Config file exists
	expectedConfigData := AdrConfig{BaseDir: filepath.Join(testDir, "my_adrs"), CurrentAdr: 10}
	bytes, _ := json.MarshalIndent(expectedConfigData, "", " ")
	err := os.WriteFile(pathCfg.ConfigFilePath, bytes, 0644) // Use os.WriteFile
	if err != nil {
		t.Fatalf("Failed to write dummy config file: %v", err)
	}

	config, err := getConfig() // getConfig now returns (AdrConfig, error)
	if err != nil {
		t.Fatalf("getConfig failed when config file exists: %v", err)
	}
	if config.BaseDir != expectedConfigData.BaseDir {
		t.Errorf("Expected BaseDir to be %s, got %s", expectedConfigData.BaseDir, config.BaseDir)
	}
	if config.CurrentAdr != expectedConfigData.CurrentAdr {
		t.Errorf("Expected CurrentAdr to be %d, got %d", expectedConfigData.CurrentAdr, config.CurrentAdr)
	}

	// Test case 2: Config file does not exist
	err = os.Remove(pathCfg.ConfigFilePath)
	if err != nil {
		t.Fatalf("Failed to remove config file for non-existence test: %v", err)
	}

	_, err = getConfig()
	if err == nil {
		t.Errorf("getConfig should have failed when config file does not exist, but it succeeded.")
	}
	// Further checks could assert the type of error, e.g., os.IsNotExist(err)
}


// Test for newAdr
func TestNewAdr(t *testing.T) {
	originalPathCfg := *pathCfg // Dereference to copy values
	testDir := tempDir(t)
	defer removeTempDir(t, testDir)

	adrBaseDirForTest := filepath.Join(testDir, "adr_repo")
	err := os.Mkdir(adrBaseDirForTest, 0755) // Create the base ADR directory
	if err != nil {
		t.Fatalf("Failed to create adrBaseDirForTest: %v", err)
	}

	// Modify global pathCfg for this test
	pathCfg.ConfigFolderPath = filepath.Join(testDir, ".adr")
	pathCfg.ConfigFilePath = filepath.Join(pathCfg.ConfigFolderPath, "config.json")
	pathCfg.TemplateFilePath = filepath.Join(pathCfg.ConfigFolderPath, "template.md")
	// Ensure .adr directory exists for config and template
	os.MkdirAll(pathCfg.ConfigFolderPath, 0755)


	defer func() {
		*pathCfg = originalPathCfg // Restore original pathCfg values
	}()

	// Initialize a config and a template using the modified pathCfg
	err = initConfig(adrBaseDirForTest)
	if err != nil {
		t.Fatalf("initConfig for TestNewAdr failed: %v", err)
	}
	err = initTemplate()
	if err != nil {
		t.Fatalf("initTemplate for TestNewAdr failed: %v", err)
	}

	currentConfig, err := getConfig()
	if err != nil {
		t.Fatalf("getConfig for TestNewAdr failed: %v", err)
	}
	currentConfig.CurrentAdr = 1 // Set initial ADR number

	adrTitle := []string{"Test", "ADR", "Creation"}
	err = newAdr(currentConfig, adrTitle) // newAdr now returns an error
	if err != nil {
		t.Fatalf("newAdr failed: %v", err)
	}


	expectedAdrNumber := 1
	expectedTitleStr := "Test ADR Creation"
	expectedFileName := strconv.Itoa(expectedAdrNumber) + "-" + strings.Join(strings.Split(strings.Trim(expectedTitleStr, "\n \t"), " "), "-") + ".md"
	// newAdr uses currentConfig.BaseDir, which was set by initConfig(adrBaseDirForTest)
	expectedFilePath := filepath.Join(adrBaseDirForTest, expectedFileName)


	if _, err := os.Stat(expectedFilePath); os.IsNotExist(err) {
		t.Fatalf("newAdr failed to create ADR file at %s", expectedFilePath)
	}

	content, err := os.ReadFile(expectedFilePath) // Changed from ioutil.ReadFile
	if err != nil {
		t.Fatalf("Failed to read ADR file: %v", err)
	}

	// Check some content
	if !strings.Contains(string(content), "# "+strconv.Itoa(expectedAdrNumber)+". "+expectedTitleStr) {
		t.Errorf("ADR file does not contain correct title string")
	}
	if !strings.Contains(string(content), "## Status\n======\nProposed") {
		t.Errorf("ADR file does not contain correct status string")
	}
	// Check date (presence and rough format, exact time is tricky)
	// Using strings.Contains for date is safer due to potential slight time differences during test run.
	if !strings.Contains(string(content), "Date: "+time.Now().Format("02-01-2006")) { // Check for YYYY-MM-DD part
		t.Errorf("ADR file does not contain correct date string (expected format like 02-01-2006)")
	}
}

// Test for NewPathConfig - success path
func TestNewPathConfig_Success(t *testing.T) {
	// This test relies on user.Current() succeeding.
	// We are mostly testing that paths are joined correctly.
	cfg, err := NewPathConfig()
	if err != nil {
		t.Fatalf("NewPathConfig() failed: %v", err)
	}

	if cfg.UserHomeDir == "" {
		t.Error("Expected UserHomeDir to be non-empty")
	}
	expectedConfigFolder := ".adr"
	if cfg.ConfigFolderName != expectedConfigFolder {
		t.Errorf("Expected ConfigFolderName to be '%s', got '%s'", expectedConfigFolder, cfg.ConfigFolderName)
	}
	expectedConfigPath := filepath.Join(cfg.UserHomeDir, expectedConfigFolder)
	if cfg.ConfigFolderPath != expectedConfigPath {
		t.Errorf("Expected ConfigFolderPath to be '%s', got '%s'", expectedConfigPath, cfg.ConfigFolderPath)
	}
	expectedDefaultBase := filepath.Join(cfg.UserHomeDir, "adr")
	if cfg.DefaultBaseFolder != expectedDefaultBase {
		t.Errorf("Expected DefaultBaseFolder to be '%s', got '%s'", expectedDefaultBase, cfg.DefaultBaseFolder)
	}
	// Add more checks for other paths if necessary
}


// Test for GetDefaultBaseFolder
func TestGetDefaultBaseFolder(t *testing.T) {
	// Relies on pathCfg being initialized by the init() function in helpers.go
	// We are testing if our getter retrieves the value correctly.
	// If pathCfg could be nil (e.g. user.Current() failed in init()), this test would be problematic.
	// However, init() in helpers.go panics if NewPathConfig() fails.
	
	// For a robust test, we might want to save and restore original pathCfg if we were to modify it.
	// But here, we assume pathCfg is valid due to init() in helpers.go.
	// If pathCfg was not initialized, GetDefaultBaseFolder() has a check, but init() should prevent that.

	expectedDefaultFolder, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Could not get user home dir for comparison: %v", err)
	}
	expectedDefaultFolder = filepath.Join(expectedDefaultFolder, "adr")

	defaultFolder := GetDefaultBaseFolder()
	if defaultFolder == "" && pathCfg == nil {
		// This case implies NewPathConfig() failed in the global init() and GetDefaultBaseFolder() handled nil pathCfg.
		// This is an edge case, as init() panics.
		t.Logf("pathCfg is nil, GetDefaultBaseFolder returned empty string as expected in such a case.")
	} else if defaultFolder != expectedDefaultFolder {
		t.Errorf("Expected DefaultBaseFolder to be '%s', got '%s'", expectedDefaultFolder, defaultFolder)
	}

	// Test the nil pathCfg scenario for GetDefaultBaseFolder explicitly if possible,
	// though it's hard because init() panics.
	// One way: temporarily set global pathCfg to nil (if it's exported or accessible).
	// This is more of a test for GetDefaultBaseFolder's internal nil check.
	
	// Store current global pathCfg, set to nil, test, then restore.
	// This is slightly risky if other tests run in parallel and depend on pathCfg.
	// However, tests are usually run sequentially by default.
	originalGlobalPathCfg := pathCfg 
	pathCfg = nil // Simulate init failure for this specific check
	
	if GetDefaultBaseFolder() != "" {
		t.Errorf("GetDefaultBaseFolder should return empty string if global pathCfg is nil, but it did not.")
	}
	pathCfg = originalGlobalPathCfg // Restore
}

// TODO: Add tests for error conditions for initConfig, initTemplate, updateConfig, newAdr
// For example:
// - initConfig: test failure if os.Mkdir fails (e.g. permissions - hard to mock)
// - initTemplate: test failure if os.WriteFile fails
// - getConfig: test failure if json.Unmarshal fails (corrupted config file)
// - newAdr: test failure if template parsing fails, or os.Create fails.
