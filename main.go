package main

import (
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
	fmt.Print("Welcome to Sift - Your Smart File Organizer!\n")

	const sourceDir = "test-folder"

	fmt.Printf("\nScanning directory: %s\n\n", sourceDir)

	err := filepath.WalkDir(sourceDir, func(currentPath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if currentPath == sourceDir {
			return nil // Skip the root directory
		}

		// Skip directories
		if d.IsDir() {
			return nil
		}

		//Get the file extension
		ext := strings.ToLower(filepath.Ext(currentPath))

		destSubFolder, ok := ruleMap[ext]

		if !ok {
			destSubFolder = "Others" // Default folder for unmatched files
		}

		// Construct the full destination path.
		// All organized files will be placed inside the sourceDir.
		// e.g., test-folder/Images/
		destDir := filepath.Join(sourceDir, destSubFolder)

		// Create the destination subdirectory if it doesn't exist.
		if err := os.MkdirAll(destDir, 0755); err != nil {
			log.Printf("Error creating directory %q: %v\n", destDir, err)
			return err
		}

		// Construct the final path for the file in its new home.
		fileName := filepath.Base(currentPath)
		newPath := filepath.Join(destDir, fileName)

		// Move the file.
		fmt.Printf("Moving %s -> %s\n", currentPath, newPath)
		if err := os.Rename(currentPath, newPath); err != nil {
			log.Printf("Error moving file %s: %v\n", currentPath, err)
			return err
		}

		return nil
	})

	if err != nil {
		log.Fatalf("Error processing directory %q: %v\n", sourceDir, err)
	}

	fmt.Println("\nSifting complete!")
}
