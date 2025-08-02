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

// Config defines the structure of our config.yml file.
type Config struct {
	ExcludeFolders []string            `yaml:"exclude_folders"`
	Rules          map[string][]string `yaml:"rules"`
}

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

// loadConfig reads a YAML file and converts it into our ruleMap and an exclusion set.
func loadConfig(configFile string) (map[string]string, map[string]bool, error) {
	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, nil, fmt.Errorf("error reading config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, nil, fmt.Errorf("error parsing YAML: %w", err)
	}

	// Create the ruleMap from the 'rules' section
	ruleMap := make(map[string]string)
	for folderName, extensions := range config.Rules {
		for _, ext := range extensions {
			normalizedExt := strings.ToLower(ext)
			if !strings.HasPrefix(normalizedExt, ".") {
				normalizedExt = "." + normalizedExt
			}
			ruleMap[normalizedExt] = folderName
		}
	}

	// Create the exclusionSet from the 'exclude_folders' section for fast lookups
	exclusionSet := make(map[string]bool)
	for _, folder := range config.ExcludeFolders {
		exclusionSet[folder] = true
	}

	return ruleMap, exclusionSet, nil
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

// Main Sifting Logic
func organizeByFileType(sourceDir string, dryRun, verbose bool, ruleMap map[string]string, exclusionSet map[string]bool) (int, int) {
	var filesMoved, filesSkipped int
	destinationFolders := buildDestinationSet(ruleMap)

	err := filepath.WalkDir(sourceDir, func(currentPath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			// Check if the directory name is in our exclusion set.
			if _, isExcluded := exclusionSet[d.Name()]; isExcluded {
				if verbose {
					fmt.Printf("Excluding directory as per config: %s\n", currentPath)
				}
				return filepath.SkipDir
			}

			// Check if it's an already organized destination folder
			if _, isDest := destinationFolders[d.Name()]; isDest {
				if verbose {
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
		destDir := filepath.Join(sourceDir, destSubFolder)
		fileName := filepath.Base(currentPath)
		finalNewPath := getNewPathWithSuffix(filepath.Join(destDir, fileName))

		if dryRun {
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
		log.Printf("Error during file type organization: %v\n", err)
	}
	return filesMoved, filesSkipped
}

func organizeByDate(sourceDir string, dryRun, verbose bool, exclusionSet map[string]bool) (int, int) {
	var filesMoved, filesSkipped int
	err := filepath.WalkDir(sourceDir, func(currentPath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			// Apply exclusions to date-based organization
			if _, isExcluded := exclusionSet[d.Name()]; isExcluded {
				if verbose {
					fmt.Printf("Excluding directory as per config: %s\n", currentPath)
				}
				return filepath.SkipDir
			}
			return nil
		}

		info, err := d.Info()
		if err != nil {
			log.Printf("Could not get file info for %s: %v", currentPath, err)
			filesSkipped++
			return nil
		}
		modTime := info.ModTime()
		year := modTime.Format("2006")
		month := modTime.Format("01-January")
		destDir := filepath.Join(sourceDir, year, month)
		fileName := filepath.Base(currentPath)
		finalNewPath := getNewPathWithSuffix(filepath.Join(destDir, fileName))

		if dryRun {
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
		log.Printf("Error during date organization: %v\n", err)
	}
	return filesMoved, filesSkipped
}

func main() {
	fmt.Println("Welcome to Sift - Your Smart File Organizer!")

	sourceDir := flag.String("source", "", "The source directory to organize.")
	dryRun := flag.Bool("dry-run", false, "Simulate the organization without moving files.")
	verbose := flag.Bool("verbose", false, "Enable detailed output.")
	configFile := flag.String("config", "", "Path to a custom config.yml file.")
	byDate := flag.Bool("by-date", false, "Organize files by date (YYYY/MM-Month).")
	flag.Parse()

	if *sourceDir == "" {
		log.Fatalln("Error: The -source flag is required.")
	}

	ruleMap := defaultRuleMap
	exclusionSet := make(map[string]bool)

	if *configFile != "" {
		fmt.Printf("Loading custom rules from: %s\n", *configFile)
		var err error
		ruleMap, exclusionSet, err = loadConfig(*configFile)
		if err != nil {
			log.Fatalf("Error loading configuration: %v", err)
		}
	}

	if *dryRun {
		fmt.Println("\n⚠️  DRY RUN MODE ENABLED: No files will be moved. ⚠️")
	}
	fmt.Printf("\nProcessing directory: %s\n\n", *sourceDir)

	var filesMoved, filesSkipped int

	if *byDate {
		fmt.Println("Organizing by date...")
		filesMoved, filesSkipped = organizeByDate(*sourceDir, *dryRun, *verbose, exclusionSet)
	} else {
		fmt.Println("Organizing by file type...")
		filesMoved, filesSkipped = organizeByFileType(*sourceDir, *dryRun, *verbose, ruleMap, exclusionSet)
	}

	fmt.Println("\n--------------------")
	fmt.Println("Sifting Complete!")
	fmt.Printf("Files Moved: %d\n", filesMoved)
	fmt.Printf("Files Skipped (due to errors): %d\n", filesSkipped)
	fmt.Println("--------------------")
}
