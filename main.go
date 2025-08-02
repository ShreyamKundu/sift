package main

import (
	"fmt"
	"io/fs"
	"log"
	"path/filepath"
)

func main() {
	fmt.Print("Welcome to Sift - Your Smart File Organizer!\n")

	const sourceDir = "test-folder"

	fmt.Printf("Scanning directory: %s\n", sourceDir)

	err := filepath.WalkDir(sourceDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() {
			fmt.Printf("Found file: %s\n", path)
		}

		return nil
	})

	if err != nil {
		log.Fatalf("Error walking the path %q: %v\n", sourceDir, err)
	}
}
