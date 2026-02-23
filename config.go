package main

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	GoogleAccount GoogleAccount `yaml:"google_account"`
	GoogleSheets  []GoogleSheet `yaml:"google_sheets"`
	GoogleDocs    []GoogleDoc   `yaml:"google_docs"`
	Backup        BackupConfig  `yaml:"backup"`
}

type GoogleAccount struct {
	CredentialsFile string `yaml:"credentials_file"`
	TokenFile       string `yaml:"token_file"`
}

type GoogleSheet struct {
	URL          string `yaml:"url"`
	ExportFormat string `yaml:"export_format"`
	Name         string `yaml:"name"`
}

type GoogleDoc struct {
	URL          string `yaml:"url"`
	ExportFormat string `yaml:"export_format"`
	Name         string `yaml:"name"`
}

type BackupConfig struct {
	Type      string    `yaml:"type"`
	LocalPath string    `yaml:"local_path"`
	Git       GitConfig `yaml:"git"`
}

type GitConfig struct {
	Repository    string         `yaml:"repository"`
	Branch        string         `yaml:"branch"`
	Credentials   GitCredentials `yaml:"credentials"`
	CommitMessage string         `yaml:"commit_message"`
}

type GitCredentials struct {
	Username string `yaml:"username"`
	Token    string `yaml:"token"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
