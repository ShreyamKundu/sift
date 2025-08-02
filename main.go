package main

import (
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// ruleMap defines the mapping from file extension to destination folder.
var ruleMap = map[string]string{
	// Images
	".jpg":  "Images",
	".jpeg": "Images",
	".png":  "Images",
	".gif":  "Images",
	".webp": "Images",
	".bmp":  "Images",
	".svg":  "Images",

	// Documents
	".pdf":  "Documents",
	".docx": "Documents",
	".doc":  "Documents",
	".txt":  "Documents",
	".ppt":  "Documents",
	".pptx": "Documents",
	".xls":  "Documents",
	".xlsx": "Documents",
	".md":   "Documents",

	// Audio
	".mp3":  "Audio",
	".wav":  "Audio",
	".m4a":  "Audio",
	".flac": "Audio",

	// Video
	".mp4":  "Videos",
	".mov":  "Videos",
	".avi":  "Videos",
	".mkv":  "Videos",
	".webm": "Videos",

	// Archives
	".zip": "Archives",
	".rar": "Archives",
	".7z":  "Archives",
	".tar": "Archives",
	".gz":  "Archives",
}

func main() {
	fmt.Println("Welcome to Sift - Your Smart File Organizer!")

	// CLI Flags
	sourceDir := flag.String("source", "", "The source directory to organize.")
	dryRun := flag.Bool("dry-run", false, "Simulate the organization without moving files.")
	flag.Parse()

	// Input Validation
	if *sourceDir == "" {
		log.Fatalln("Error: The -source flag is required. Please specify a directory to organize.")
	}

	if *dryRun {
		fmt.Println("\n⚠️  DRY RUN MODE ENABLED: No files will be moved. ⚠️")
	}

	fmt.Printf("\nScanning directory: %s\n\n", *sourceDir)

	err := filepath.WalkDir(*sourceDir, func(currentPath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if currentPath == *sourceDir || d.IsDir() {
			return nil // Skip the root directory and other subdirectories
		}

		ext := strings.ToLower(filepath.Ext(currentPath))
		destSubFolder, ok := ruleMap[ext]
		if !ok {
			destSubFolder = "Others"
		}

		destDir := filepath.Join(*sourceDir, destSubFolder)
		fileName := filepath.Base(currentPath)
		newPath := filepath.Join(destDir, fileName)

		if *dryRun {
			fmt.Printf("[Dry Run] Move %s -> %s\n", currentPath, newPath)
			return nil
		}

		if err := os.MkdirAll(destDir, 0755); err != nil {
			log.Printf("Error creating directory %s: %v\n", destDir, err)
			return err
		}

		fmt.Printf("Moving %s -> %s\n", currentPath, newPath)
		if err := os.Rename(currentPath, newPath); err != nil {
			log.Printf("Error moving file %s: %v\n", currentPath, err)
			return err
		}

		return nil
	})

	if err != nil {
		log.Fatalf("Error processing directory %q: %v\n", *sourceDir, err)
	}

	fmt.Println("\nSifting complete!")
}
