package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type App struct {
	config  *Config
	google  *GoogleClient
	storage Storage
}

func NewApp(config *Config) *App {
	return &App{
		config: config,
	}
}

func (a *App) Run() error {
	fmt.Println("Initializing Google API client...")
	
	var err error
	a.google, err = NewGoogleClient(a.config.GoogleAccount)
	if err != nil {
		return fmt.Errorf("failed to initialize Google client: %w", err)
	}
	
	fmt.Println("✓ Authentication successful\n")

	// Initialize storage
	a.storage, err = NewStorage(a.config.Backup)
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}

	// Ensure backup directory exists
	if err := os.MkdirAll(a.config.Backup.LocalPath, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}

	fmt.Println("=== Starting Backup ===\n")

	// Backup Google Sheets
	for _, sheet := range a.config.GoogleSheets {
		if err := a.backupSheet(sheet); err != nil {
			return fmt.Errorf("failed to backup sheet %s: %w", sheet.Name, err)
		}
	}

	// Backup Google Docs
	for _, doc := range a.config.GoogleDocs {
		if err := a.backupDoc(doc); err != nil {
			return fmt.Errorf("failed to backup doc %s: %w", doc.Name, err)
		}
	}

	// Commit and push if using git
	if a.config.Backup.Type == "git" {
		fmt.Println("\nCommitting to git...")
		if err := a.storage.Commit(time.Now()); err != nil {
			return fmt.Errorf("failed to commit: %w", err)
		}
	}

	return nil
}

func (a *App) backupSheet(sheet GoogleSheet) error {
	fmt.Printf("→ Backing up sheet: %s (%s)... ", sheet.Name, sheet.ExportFormat)
	
	data, err := a.google.ExportSheet(sheet.URL, sheet.ExportFormat)
	if err != nil {
		fmt.Println("✗")
		return err
	}

	filename := fmt.Sprintf("%s.%s", sheet.Name, sheet.ExportFormat)
	path := filepath.Join(a.config.Backup.LocalPath, filename)
	
	if err := a.storage.Save(path, data); err != nil {
		fmt.Println("✗")
		return err
	}

	fmt.Printf("✓ (%d KB)\n", len(data)/1024)
	return nil
}

func (a *App) backupDoc(doc GoogleDoc) error {
	fmt.Printf("→ Backing up doc: %s (%s)... ", doc.Name, doc.ExportFormat)
	
	data, err := a.google.ExportDoc(doc.URL, doc.ExportFormat)
	if err != nil {
		fmt.Println("✗")
		return err
	}

	filename := fmt.Sprintf("%s.%s", doc.Name, doc.ExportFormat)
	path := filepath.Join(a.config.Backup.LocalPath, filename)
	
	if err := a.storage.Save(path, data); err != nil {
		fmt.Println("✗")
		return err
	}

	fmt.Printf("✓ (%d KB)\n", len(data)/1024)
	return nil
}
