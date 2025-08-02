package main

import (
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Default Rules
var defaultRuleMap = map[string]string{
	// Images
	".jpg": "Images", ".jpeg": "Images", ".png": "Images", ".gif": "Images",
	".webp": "Images", ".bmp": "Images", ".svg": "Images",

	// Documents
	".pdf": "Documents", ".docx": "Documents", ".doc": "Documents", ".txt": "Documents",
	".ppt": "Documents", ".pptx": "Documents", ".xls": "Documents", ".xlsx": "Documents",
	".md": "Documents",

	// Audio
	".mp3": "Audio", ".wav": "Audio", ".m4a": "Audio", ".flac": "Audio",

	// Video
	".mp4": "Videos", ".mov": "Videos", ".avi": "Videos", ".mkv": "Videos",
	".webm": "Videos",

	// Archives
	".zip": "Archives", ".rar": "Archives", ".7z": "Archives", ".tar": "Archives",
	".gz": "Archives",
}

// loadRulesFromConfig reads a YAML file and converts it into our ruleMap format.
func loadRulesFromConfig(configFile string) (map[string]string, error) {
	// Read the raw YAML file content
	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	// The YAML is structured as "FolderName: [extensions...]", so we parse it into this temporary map.
	var parsedConfig map[string][]string
	if err := yaml.Unmarshal(data, &parsedConfig); err != nil {
		return nil, fmt.Errorf("error parsing YAML: %w", err)
	}

	// Now, we invert the map to create our final ruleMap: {".ext": "FolderName"}
	ruleMap := make(map[string]string)
	for folderName, extensions := range parsedConfig {
		for _, ext := range extensions {
			// Ensure the extension starts with a dot and is lowercase for consistent matching.
			normalizedExt := strings.ToLower(ext)
			if !strings.HasPrefix(normalizedExt, ".") {
				normalizedExt = "." + normalizedExt
			}
			ruleMap[normalizedExt] = folderName
		}
	}

	return ruleMap, nil
}

func buildDestinationSet(rules map[string]string) map[string]bool {
	set := make(map[string]bool)
	for _, folder := range rules {
		set[folder] = true
	}
	set["Others"] = true
	return set
}

func getNewPathWithSuffix(path string) string {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return path
	}
	dir := filepath.Dir(path)
	ext := filepath.Ext(path)
	baseName := strings.TrimSuffix(filepath.Base(path), ext)
	for i := 1; ; i++ {
		newBaseName := fmt.Sprintf("%s (%d)", baseName, i)
		newPath := filepath.Join(dir, newBaseName+ext)
		if _, err := os.Stat(newPath); os.IsNotExist(err) {
			return newPath
		}
	}
}

func main() {
	fmt.Println("Welcome to Sift - Your Smart File Organizer!")

	// CLI Flags
	sourceDir := flag.String("source", "", "The source directory to organize.")
	dryRun := flag.Bool("dry-run", false, "Simulate the organization without moving files.")
	verbose := flag.Bool("verbose", false, "Enable detailed output.")
	configFile := flag.String("config", "", "Path to a custom config.yml file.")
	flag.Parse()

	// Input Validation
	if *sourceDir == "" {
		log.Fatalln("Error: The -source flag is required.")
	}

	// Rule Loading
	ruleMap := defaultRuleMap
	if *configFile != "" {
		fmt.Printf("Loading custom rules from: %s\n", *configFile)
		var err error
		ruleMap, err = loadRulesFromConfig(*configFile)
		if err != nil {
			log.Fatalf("Error loading configuration: %v", err)
		}
	}

	if *dryRun {
		fmt.Println("\n⚠️  DRY RUN MODE ENABLED: No files will be moved. ⚠️")
	}
	fmt.Printf("\nScanning directory: %s\n\n", *sourceDir)

	var filesMoved, filesSkipped int
	destinationFolders := buildDestinationSet(ruleMap)

	err := filepath.WalkDir(*sourceDir, func(currentPath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			if _, isDest := destinationFolders[d.Name()]; isDest {
				if *verbose {
					fmt.Printf("Skipping already organized directory: %s\n", currentPath)
				}
				return filepath.SkipDir
			}
			return nil
		}

		ext := strings.ToLower(filepath.Ext(currentPath))
		destSubFolder, ok := ruleMap[ext]
		if !ok {
			destSubFolder = "Others"
		}

		destDir := filepath.Join(*sourceDir, destSubFolder)
		fileName := filepath.Base(currentPath)
		potentialNewPath := filepath.Join(destDir, fileName)
		finalNewPath := getNewPathWithSuffix(potentialNewPath)

		if *dryRun {
			fmt.Printf("[Dry Run] Move %s -> %s\n", currentPath, finalNewPath)
			filesMoved++
			return nil
		}

		if err := os.MkdirAll(destDir, 0755); err != nil {
			log.Printf("Error creating directory %s: %v\n", destDir, err)
			filesSkipped++
			return err
		}

		if err := os.Rename(currentPath, finalNewPath); err != nil {
			log.Printf("Error moving file %s: %v\n", currentPath, err)
			filesSkipped++
			return err
		}

		fmt.Printf("Moved %s -> %s\n", currentPath, finalNewPath)
		filesMoved++

		return nil
	})

	if err != nil {
		log.Fatalf("Error processing directory %q: %v\n", *sourceDir, err)
	}

	fmt.Println("\n--------------------")
	fmt.Println("Sifting Complete!")
	fmt.Printf("Files Moved: %d\n", filesMoved)
	fmt.Printf("Files Skipped (due to errors): %d\n", filesSkipped)
	fmt.Println("--------------------")
}
