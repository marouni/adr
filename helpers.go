package main

import (
	"encoding/json"
	"html/template"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
)

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

var usr, err = user.Current()
var adrConfigFolderName = ".adr"
var adrConfigFileName = "config.json"
var adrConfigTemplateName = "template.md"
var adrConfigFolderPath = filepath.Join(usr.HomeDir, adrConfigFolderName)
var adrConfigFilePath = filepath.Join(adrConfigFolderPath, adrConfigFileName)
var adrTemplateFilePath = filepath.Join(adrConfigFolderPath, adrConfigTemplateName)
var adrDefaultBaseFolder = filepath.Join(usr.HomeDir, "adr")

func initBaseDir(baseDir string) {
	if _, err := os.Stat(baseDir); os.IsNotExist(err) {
		os.Mkdir(baseDir, 0744)
	} else {
		color.Red(baseDir + " already exists, skipping folder creation")
	}
}

func initConfig(baseDir string) {
	if _, err := os.Stat(adrConfigFolderPath); os.IsNotExist(err) {
		os.Mkdir(adrConfigFolderPath, 0744)
	}
	config := AdrConfig{baseDir, 0}
	bytes, err := json.MarshalIndent(config, "", " ")
	if err != nil {
		panic(err)
	}
	ioutil.WriteFile(adrConfigFilePath, bytes, 0644)
}

func initTemplate() {
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

	ioutil.WriteFile(adrTemplateFilePath, body, 0644)
}

func updateConfig(config AdrConfig) {
	bytes, err := json.MarshalIndent(config, "", " ")
	if err != nil {
		panic(err)
	}
	ioutil.WriteFile(adrConfigFilePath, bytes, 0644)
}

func getConfig() AdrConfig {
	var currentConfig AdrConfig

	bytes, err := ioutil.ReadFile(adrConfigFilePath)
	if err != nil {
		color.Red("No ADR configuration is found!")
		color.HiGreen("Start by initializing ADR configuration, check 'adr init --help' for more help")
		os.Exit(1)
	}

	json.Unmarshal(bytes, &currentConfig)
	return currentConfig
}

func newAdr(config AdrConfig, adrName []string) {
	adr := Adr{
		Title:  strings.Join(adrName, " "),
		Date:   time.Now().Format("02-01-2006 15:04:05"),
		Number: config.CurrentAdr,
		Status: PROPOSED,
	}
	template, err := template.ParseFiles(adrTemplateFilePath)
	if err != nil {
		panic(err)
	}
	adrFileName := strconv.Itoa(adr.Number) + "-" + strings.Join(strings.Split(strings.Trim(adr.Title, "\n \t"), " "), "-") + ".md"
	adrFullPath := filepath.Join(config.BaseDir, adrFileName)
	f, err := os.Create(adrFullPath)
	if err != nil {
		panic(err)
	}
	template.Execute(f, adr)
	f.Close()
	color.Green("ADR number " + strconv.Itoa(adr.Number) + " was successfully written to : " + adrFullPath)
}
