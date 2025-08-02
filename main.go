package main

import (
	"fmt"
	"io/fs"
	"log"
	"path/filepath"
	"strings"
)

// ruleMap defines the mapping from file extension to destination folder.
var ruleMap = map[string]string{
	".jpg":  "Images",
	".jpeg": "Images",
	".png":  "Images",
	".gif":  "Images",

	".pdf":  "Documents",
	".docx": "Documents",
	".txt":  "Documents",

	".mp3": "Audio",
	".wav": "Audio",
	".mp4": "Videos",
}

func main() {
	fmt.Print("Welcome to Sift - Your Smart File Organizer!\n")

	const sourceDir = "test-folder"

	fmt.Printf("\nScanning directory: %s\n\n", sourceDir)

	err := filepath.WalkDir(sourceDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if d.IsDir() {
			return nil
		}

		//Get the file extension
		ext := strings.ToLower(filepath.Ext(path))

		destFolder, ok := ruleMap[ext]

		if ok {
			fmt.Printf("File: %s -> belongs in: %s\n", path, destFolder)
		} else {
			fmt.Printf("File: %s -> belongs in: Others\n", path)
		}

		return nil
	})

	if err != nil {
		log.Fatalf("Error walking the path %q: %v\n", sourceDir, err)
	}
}
