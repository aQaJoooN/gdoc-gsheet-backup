package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: gdocs-backup <config-file>")
		fmt.Println("Example: gdocs-backup config.yaml")
		os.Exit(1)
	}

	configPath := os.Args[1]
	
	config, err := LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	app := NewApp(config)
	
	if err := app.Run(); err != nil {
		log.Fatalf("Backup failed: %v", err)
	}

	fmt.Println("\n✅ Backup completed successfully!")
}
