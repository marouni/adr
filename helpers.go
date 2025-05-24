package main

import (
	"encoding/json"
	"html/template"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
)

// PathConfig holds all path-related configurations.
type PathConfig struct {
	ConfigFolderName  string // e.g., ".adr"
	ConfigFileName    string // e.g., "config.json"
	TemplateFileName  string // e.g., "template.md"
	UserHomeDir       string // User's home directory
	ConfigFolderPath  string // Full path to the configuration folder
	ConfigFilePath    string // Full path to the configuration file
	TemplateFilePath  string // Full path to the template file
	DefaultBaseFolder string // Default base directory for ADRs
}

// AdrConfig ADR configuration, loaded and used by each sub-command
type AdrConfig struct {
	BaseDir    string `json:"base_directory"`
	CurrentAdr int    `json:"current_id"`
}

// Adr basic structure
type Adr struct {
	Number int
	Title  string
	Date   string
	Status AdrStatus
}

// AdrStatus type
type AdrStatus string

// ADR status enums
const (
	PROPOSED   AdrStatus = "Proposed"
	ACCEPTED   AdrStatus = "Accepted"
	DEPRECATED AdrStatus = "Deprecated"
	SUPERSEDED AdrStatus = "Superseded"
)

var pathCfg *PathConfig

// NewPathConfig initializes a new PathConfig instance.
func NewPathConfig() (*PathConfig, error) {
	usr, err := user.Current()
	if err != nil {
		return nil, err
	}

	cfg := &PathConfig{
		ConfigFolderName: ".adr",
		ConfigFileName:   "config.json",
		TemplateFileName: "template.md",
		UserHomeDir:      usr.HomeDir,
	}

	cfg.ConfigFolderPath = filepath.Join(cfg.UserHomeDir, cfg.ConfigFolderName)
	cfg.ConfigFilePath = filepath.Join(cfg.ConfigFolderPath, cfg.ConfigFileName)
	cfg.TemplateFilePath = filepath.Join(cfg.ConfigFolderPath, cfg.TemplateFileName)
	cfg.DefaultBaseFolder = filepath.Join(cfg.UserHomeDir, "adr")

	return cfg, nil
}

// GetDefaultBaseFolder returns the default base directory for ADRs.
// It's populated during pathCfg initialization.
func GetDefaultBaseFolder() string {
	if pathCfg == nil {
		// This should ideally not happen if init() runs correctly.
		// Consider logging an error or returning a sensible default/error.
		// For now, returning empty string or panicking might be alternatives.
		// However, the init() function already panics if pathCfg isn't set.
		return "" // Or handle error more gracefully
	}
	return pathCfg.DefaultBaseFolder
}

func init() {
	var err error
	pathCfg, err = NewPathConfig()
	if err != nil {
		// Panicking here because these paths are essential for the application to run.
		panic("Failed to initialize path configuration: " + err.Error())
	}
}

func initBaseDir(baseDir string) {
	if _, err := os.Stat(baseDir); os.IsNotExist(err) {
		// Consider returning error from os.Mkdir if it fails.
		// For now, keeping behavior similar to original.
		os.Mkdir(baseDir, 0744) 
	} else {
		color.Red(baseDir + " already exists, skipping folder creation")
	}
}

func initConfig(baseDir string) error {
	if _, err := os.Stat(pathCfg.ConfigFolderPath); os.IsNotExist(err) {
		err := os.Mkdir(pathCfg.ConfigFolderPath, 0744)
		if err != nil {
			return err
		}
	}
	config := AdrConfig{baseDir, 0}
	bytes, err := json.MarshalIndent(config, "", " ")
	if err != nil {
		return err
	}
	return os.WriteFile(pathCfg.ConfigFilePath, bytes, 0644)
}

func initTemplate() error {
	body := []byte(`
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

`)

	return os.WriteFile(pathCfg.TemplateFilePath, body, 0644)
}

func updateConfig(config AdrConfig) error {
	bytes, err := json.MarshalIndent(config, "", " ")
	if err != nil {
		return err
	}
	return os.WriteFile(pathCfg.ConfigFilePath, bytes, 0644)
}

func getConfig() (AdrConfig, error) {
	var currentConfig AdrConfig

	bytes, err := os.ReadFile(pathCfg.ConfigFilePath)
	if err != nil {
		return currentConfig, err
	}

	err = json.Unmarshal(bytes, &currentConfig)
	return currentConfig, err
}

func newAdr(config AdrConfig, adrName []string) error {
	adr := Adr{
		Title:  strings.Join(adrName, " "),
		Date:   time.Now().Format("02-01-2006 15:04:05"),
		Number: config.CurrentAdr,
		Status: PROPOSED,
	}
	tmpl, err := template.ParseFiles(pathCfg.TemplateFilePath)
	if err != nil {
		return err
	}
	adrFileName := strconv.Itoa(adr.Number) + "-" + strings.Join(strings.Split(strings.Trim(adr.Title, "\n \t"), " "), "-") + ".md"
	adrFullPath := filepath.Join(config.BaseDir, adrFileName)
	f, err := os.Create(adrFullPath)
	if err != nil {
		return err
	}
	defer f.Close()
	err = tmpl.Execute(f, adr)
	if err != nil {
		return err
	}
	color.Green("ADR number " + strconv.Itoa(adr.Number) + " was successfully written to : " + adrFullPath)
	return nil
}
