package main

import (
	"os"
	"path/filepath"
	"testing"
	"io/ioutil"
	"encoding/json"
	"time"
	"strconv"
	"strings"
)

// Helper function to create a temporary directory for testing
func tempDir(t *testing.T) string {
	dir, err := ioutil.TempDir("", "adr-test")
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
	// Override config paths for testing
	originalAdrConfigFolderPath := adrConfigFolderPath
	originalAdrConfigFilePath := adrConfigFilePath
	testDir := tempDir(t)
	adrConfigFolderPath = filepath.Join(testDir, ".adr")
	adrConfigFilePath = filepath.Join(adrConfigFolderPath, "config.json")
	defer func() {
		adrConfigFolderPath = originalAdrConfigFolderPath
		adrConfigFilePath = originalAdrConfigFilePath
		removeTempDir(t, testDir)
	}()

	testBaseDir := filepath.Join(testDir, "adr_docs")
	initConfig(testBaseDir)

	if _, err := os.Stat(adrConfigFilePath); os.IsNotExist(err) {
		t.Fatalf("initConfig failed to create config file at %s", adrConfigFilePath)
	}

	var config AdrConfig
	configBytes, err := ioutil.ReadFile(adrConfigFilePath)
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
	// Override template path for testing
	originalAdrTemplateFilePath := adrTemplateFilePath
	testDir := tempDir(t)
	// Ensure the .adr directory exists for the template file
	adrConfigFolderPath = filepath.Join(testDir, ".adr")
	if _, err := os.Stat(adrConfigFolderPath); os.IsNotExist(err) {
		os.Mkdir(adrConfigFolderPath, 0755)
	}
	adrTemplateFilePath = filepath.Join(adrConfigFolderPath, "template.md")
	defer func() {
		adrTemplateFilePath = originalAdrTemplateFilePath
		removeTempDir(t, testDir)
	}()

	initTemplate()

	if _, err := os.Stat(adrTemplateFilePath); os.IsNotExist(err) {
		t.Fatalf("initTemplate failed to create template file at %s", adrTemplateFilePath)
	}

	// Verify content (basic check for now)
	content, err := ioutil.ReadFile(adrTemplateFilePath)
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
	// Override config paths for testing
	originalAdrConfigFolderPath := adrConfigFolderPath
	originalAdrConfigFilePath := adrConfigFilePath
	testDir := tempDir(t)
	adrConfigFolderPath = filepath.Join(testDir, ".adr")
	adrConfigFilePath = filepath.Join(adrConfigFolderPath, "config.json")
	defer func() {
		adrConfigFolderPath = originalAdrConfigFolderPath
		adrConfigFilePath = originalAdrConfigFilePath
		removeTempDir(t, testDir)
	}()

	// Initialize a config first
	initialBaseDir := filepath.Join(testDir, "initial_docs")
	initConfig(initialBaseDir)


	updatedBaseDir := filepath.Join(testDir, "updated_docs")
	updatedConfig := AdrConfig{BaseDir: updatedBaseDir, CurrentAdr: 5}
	updateConfig(updatedConfig)

	var config AdrConfig
	configBytes, err := ioutil.ReadFile(adrConfigFilePath)
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
	// Override config paths for testing
	originalAdrConfigFolderPath := adrConfigFolderPath
	originalAdrConfigFilePath := adrConfigFilePath
	testDir := tempDir(t)
	adrConfigFolderPath = filepath.Join(testDir, ".adr")
	adrConfigFilePath = filepath.Join(adrConfigFolderPath, "config.json")
	defer func() {
		adrConfigFolderPath = originalAdrConfigFolderPath
		adrConfigFilePath = originalAdrConfigFilePath
		removeTempDir(t, testDir)
	}()

	// Test case 1: Config file exists
	expectedConfig := AdrConfig{BaseDir: filepath.Join(testDir, "my_adrs"), CurrentAdr: 10}
	// Create a dummy config file
	if _, err := os.Stat(adrConfigFolderPath); os.IsNotExist(err) {
		os.Mkdir(adrConfigFolderPath, 0755)
	}
	bytes, _ := json.MarshalIndent(expectedConfig, "", " ")
	ioutil.WriteFile(adrConfigFilePath, bytes, 0644)

	config := getConfig()
	if config.BaseDir != expectedConfig.BaseDir {
		t.Errorf("Expected BaseDir to be %s, got %s", expectedConfig.BaseDir, config.BaseDir)
	}
	if config.CurrentAdr != expectedConfig.CurrentAdr {
		t.Errorf("Expected CurrentAdr to be %d, got %d", expectedConfig.CurrentAdr, config.CurrentAdr)
	}

	// Test case 2: Config file does not exist (handled by os.Exit, difficult to test directly without refactor)
	// We can check if the function attempts to read the correct file,
	// but capturing os.Exit requires more advanced techniques (e.g., running as a separate process).
	// For now, we'll assume the os.Exit path is covered by manual testing or integration tests.
	// To simulate, we remove the config file. The function should print an error and exit.
	os.Remove(adrConfigFilePath)
	// The test will fail here if getConfig doesn't exit.
	// This is a limited way to test this scenario.
	// getConfig() // This would call os.Exit(1)
	// color.Red is called before os.Exit, we can't directly assert that here.
}


// Test for newAdr
func TestNewAdr(t *testing.T) {
	// Override config and template paths for testing
	originalAdrConfigFolderPath := adrConfigFolderPath
	originalAdrConfigFilePath := adrConfigFilePath
	originalAdrTemplateFilePath := adrTemplateFilePath
	testDir := tempDir(t)
	adrBaseDir := filepath.Join(testDir, "adr_repo")
	os.Mkdir(adrBaseDir, 0755) // Create the base ADR directory

	adrConfigFolderPath = filepath.Join(testDir, ".adr")
	adrConfigFilePath = filepath.Join(adrConfigFolderPath, "config.json")
	adrTemplateFilePath = filepath.Join(adrConfigFolderPath, "template.md")

	defer func() {
		adrConfigFolderPath = originalAdrConfigFolderPath
		adrConfigFilePath = originalAdrConfigFilePath
		adrTemplateFilePath = originalAdrTemplateFilePath
		removeTempDir(t, testDir)
	}()

	// Initialize a config and a template
	initConfig(adrBaseDir) // initConfig now uses the overridden adrConfigFilePath
	initTemplate()       // initTemplate now uses the overridden adrTemplateFilePath

	config := getConfig() // getConfig now uses the overridden adrConfigFilePath
	config.CurrentAdr = 1 // Set initial ADR number

	adrTitle := []string{"Test", "ADR", "Creation"}
	newAdr(config, adrTitle)

	expectedAdrNumber := 1
	expectedTitleStr := "Test ADR Creation"
	expectedFileName := strconv.Itoa(expectedAdrNumber) + "-" + strings.Join(strings.Split(strings.Trim(expectedTitleStr, "\n \t"), " "), "-") + ".md"
	expectedFilePath := filepath.Join(adrBaseDir, expectedFileName)

	if _, err := os.Stat(expectedFilePath); os.IsNotExist(err) {
		t.Fatalf("newAdr failed to create ADR file at %s", expectedFilePath)
	}

	content, err := ioutil.ReadFile(expectedFilePath)
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
	if !strings.Contains(string(content), "Date: "+time.Now().Format("02-01-2006")) {
		t.Errorf("ADR file does not contain correct date string (yy-mm-dd)")
	}
}
