package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

	fmt.Println("✓ Authentication successful")
	fmt.Println()

	// Initialize storage
	a.storage, err = NewStorage(a.config.Backup)
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}

	// Ensure backup directory exists
	if err := os.MkdirAll(a.config.Backup.LocalPath, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}

	fmt.Println("=== Starting Backup ===")
	fmt.Println()

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

	// Backup Google Drive Files
	for _, file := range a.config.GoogleDriveFiles {
		if err := a.backupDriveFile(file); err != nil {
			return fmt.Errorf("failed to backup drive file %s: %w", file.Name, err)
		}
	}

	// Generate README.md if using git backup
	if a.config.Backup.Type == "git" {
		fmt.Println()
		fmt.Println("Generating README.md...")
		if err := a.generateReadme(); err != nil {
			return fmt.Errorf("failed to generate README: %w", err)
		}
	}

	// Commit and push if using git
	if a.config.Backup.Type == "git" {
		fmt.Println()
		fmt.Println("Committing to git...")
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

func (a *App) backupDriveFile(file GoogleDriveFile) error {
	fmt.Printf("→ Backing up drive file: %s... ", file.Name)

	data, originalName, err := a.google.DownloadDriveFile(file.URL)
	if err != nil {
		fmt.Println("✗")
		return err
	}

	// Use custom name if provided, otherwise use original filename
	filename := file.Name
	if filename == "" {
		filename = originalName
	}

	path := filepath.Join(a.config.Backup.LocalPath, filename)

	if err := a.storage.Save(path, data); err != nil {
		fmt.Println("✗")
		return err
	}

	fmt.Printf("✓ (%d KB)\n", len(data)/1024)
	return nil
}

func (a *App) generateReadme() error {
	readmePath := filepath.Join(a.config.Backup.LocalPath, "README.md")

	var content strings.Builder

	// Header
	content.WriteString("# Google Docs & Sheets Backup\n\n")
	content.WriteString(fmt.Sprintf("Last backup: %s\n\n", time.Now().Format("2006-01-02 15:04:05")))
	content.WriteString("---\n\n")

	// Google Sheets section
	if len(a.config.GoogleSheets) > 0 {
		content.WriteString("## 📊 Google Sheets\n\n")
		content.WriteString("| File Name | Format | Source URL |\n")
		content.WriteString("|-----------|--------|------------|\n")

		for _, sheet := range a.config.GoogleSheets {
			filename := fmt.Sprintf("%s.%s", sheet.Name, sheet.ExportFormat)
			content.WriteString(fmt.Sprintf("| [%s](%s) | %s | [Open in Google Sheets](%s) |\n",
				filename, filename, sheet.ExportFormat, sheet.URL))
		}
		content.WriteString("\n")
	}

	// Google Docs section
	if len(a.config.GoogleDocs) > 0 {
		content.WriteString("## 📝 Google Docs\n\n")
		content.WriteString("| File Name | Format | Source URL |\n")
		content.WriteString("|-----------|--------|------------|\n")

		for _, doc := range a.config.GoogleDocs {
			filename := fmt.Sprintf("%s.%s", doc.Name, doc.ExportFormat)
			content.WriteString(fmt.Sprintf("| [%s](%s) | %s | [Open in Google Docs](%s) |\n",
				filename, filename, doc.ExportFormat, doc.URL))
		}
		content.WriteString("\n")
	}

	// Google Drive Files section
	if len(a.config.GoogleDriveFiles) > 0 {
		content.WriteString("## 📁 Google Drive Files\n\n")
		content.WriteString("| File Name | Source URL |\n")
		content.WriteString("|-----------|------------|\n")

		for _, file := range a.config.GoogleDriveFiles {
			filename := file.Name
			content.WriteString(fmt.Sprintf("| [%s](%s) | [Open in Google Drive](%s) |\n",
				filename, filename, file.URL))
		}
		content.WriteString("\n")
	}

	// Footer
	content.WriteString("---\n\n")
	content.WriteString("## 📋 Backup Summary\n\n")
	content.WriteString(fmt.Sprintf("- **Total Sheets**: %d\n", len(a.config.GoogleSheets)))
	content.WriteString(fmt.Sprintf("- **Total Docs**: %d\n", len(a.config.GoogleDocs)))
	content.WriteString(fmt.Sprintf("- **Total Drive Files**: %d\n", len(a.config.GoogleDriveFiles)))
	content.WriteString(fmt.Sprintf("- **Total Files**: %d\n", len(a.config.GoogleSheets)+len(a.config.GoogleDocs)+len(a.config.GoogleDriveFiles)))
	content.WriteString(fmt.Sprintf("- **Backup Time**: %s\n", time.Now().Format("2006-01-02 15:04:05")))
	content.WriteString("\n")
	content.WriteString("*This backup was automatically generated by gdocs-backup tool.*\n")

	// Write to file
	return os.WriteFile(readmePath, []byte(content.String()), 0644)
}
