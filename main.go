package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// --- Structs & Constants ---
const logFileName = ".sift_log"
const logSeparator = "::SFT::"

type Config struct {
	ExcludeFolders []string            `yaml:"exclude_folders"`
	Rules          map[string][]string `yaml:"rules"`
}

var defaultRuleMap = map[string]string{
	//Images
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

// --- Core Functions ---
func loadConfig(configFile string) (map[string]string, map[string]bool, error) {
	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, nil, fmt.Errorf("error reading config file: %w", err)
	}
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, nil, fmt.Errorf("error parsing YAML: %w", err)
	}
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

// --- Sifting Logic (Organize Command) ---
func organizeByFileType(sourceDir string, dryRun, verbose bool, ruleMap map[string]string, exclusionSet map[string]bool) (int, int) {
	var filesMoved, filesSkipped int
	destinationFolders := buildDestinationSet(ruleMap)
	logFilePath := filepath.Join(sourceDir, logFileName)
	logFile, err := os.Create(logFilePath)
	if err != nil {
		log.Printf("Error: Could not create log file for undo: %v", err)
		return 0, 0
	}
	defer logFile.Close()
	logWriter := bufio.NewWriter(logFile)
	defer logWriter.Flush()
	err = filepath.WalkDir(sourceDir, func(currentPath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if currentPath == logFilePath {
			return nil
		}
		if d.IsDir() {
			if _, isExcluded := exclusionSet[d.Name()]; isExcluded {
				if verbose {
					fmt.Printf("Excluding directory: %s\n", currentPath)
				}
				return filepath.SkipDir
			}
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

		logLine := fmt.Sprintf("%s%s%s\n", finalNewPath, logSeparator, currentPath)
		if _, err := logWriter.WriteString(logLine); err != nil {
			log.Printf("Warning: Failed to write to undo log: %v", err)
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
	logFilePath := filepath.Join(sourceDir, logFileName)
	logFile, err := os.Create(logFilePath)
	if err != nil {
		log.Printf("Error: Could not create log file for undo: %v", err)
		return 0, 0
	}
	defer logFile.Close()
	logWriter := bufio.NewWriter(logFile)
	defer logWriter.Flush()
	err = filepath.WalkDir(sourceDir, func(currentPath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if currentPath == logFilePath {
			return nil
		}
		if d.IsDir() {
			if _, isExcluded := exclusionSet[d.Name()]; isExcluded {
				if verbose {
					fmt.Printf("Excluding directory: %s\n", currentPath)
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

		logLine := fmt.Sprintf("%s%s%s\n", finalNewPath, logSeparator, currentPath)
		if _, err := logWriter.WriteString(logLine); err != nil {
			log.Printf("Warning: Failed to write to undo log: %v", err)
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

// --- Undo Logic ---
func performUndo(sourceDir string) {
	logFilePath := filepath.Join(sourceDir, logFileName)
	if _, err := os.Stat(logFilePath); os.IsNotExist(err) {
		log.Fatalf("Error: No undo log file found in %s. Cannot perform undo.", sourceDir)
	}

	logFile, err := os.Open(logFilePath)
	if err != nil {
		log.Fatalf("Error opening undo log file: %v", err)
	}
	defer logFile.Close()

	var lines []string
	scanner := bufio.NewScanner(logFile)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("Error reading undo log file: %v", err)
	}

	fmt.Printf("Found %d operations to undo.\n", len(lines))
	var filesReverted int

	for i := len(lines) - 1; i >= 0; i-- {

		parts := strings.Split(lines[i], logSeparator)
		if len(parts) != 2 {
			log.Printf("Warning: Skipping malformed log entry: %s", lines[i])
			continue
		}
		newPath := parts[0]
		originalPath := parts[1]

		fmt.Printf("Reverting %s -> %s\n", newPath, originalPath)
		if err := os.Rename(newPath, originalPath); err != nil {
			log.Printf("Error reverting file %s: %v", newPath, err)
			log.Println("Stopping undo operation to prevent data loss.")
			return
		}
		filesReverted++
	}

	if err := os.Remove(logFilePath); err != nil {
		log.Printf("Warning: Could not remove undo log file: %v", err)
	}

	fmt.Println("\n--------------------")
	fmt.Println("✅ Undo Complete!")
	fmt.Printf("Files Reverted: %d\n", filesReverted)
	fmt.Println("--------------------")
}

// --- Main Application Logic ---
func main() {
	organizeCmd := flag.NewFlagSet("organize", flag.ExitOnError)
	undoCmd := flag.NewFlagSet("undo", flag.ExitOnError)
	sourceDir := organizeCmd.String("source", "", "The source directory to organize. (Required)")
	dryRun := organizeCmd.Bool("dry-run", false, "Simulate the organization without moving files.")
	verbose := organizeCmd.Bool("verbose", false, "Enable detailed output.")
	configFile := organizeCmd.String("config", "", "Path to a custom config.yml file.")
	byDate := organizeCmd.Bool("by-date", false, "Organize files by date (YYYY/MM-Month).")
	excludeDirs := organizeCmd.String("exclude", "", "Comma-separated list of folder names to exclude.")
	undoSourceDir := undoCmd.String("source", "", "The directory where the organization was performed. (Required)")

	if len(os.Args) < 2 {
		fmt.Println("Expected 'organize' or 'undo' subcommands.")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "organize":
		organizeCmd.Parse(os.Args[2:])
		if *sourceDir == "" {
			log.Println("Error: The -source flag is required for the organize command.")
			organizeCmd.PrintDefaults()
			os.Exit(1)
		}
		fmt.Println("🚀 Welcome to Sift - Your Smart File Organizer!")
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
		if *excludeDirs != "" {
			foldersToExclude := strings.Split(*excludeDirs, ",")
			for _, folder := range foldersToExclude {
				exclusionSet[strings.TrimSpace(folder)] = true
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
		fmt.Println("✅ Sifting Complete!")
		fmt.Printf("Files Moved: %d\n", filesMoved)
		fmt.Printf("Files Skipped (due to errors): %d\n", filesSkipped)
		fmt.Println("--------------------")

	case "undo":
		undoCmd.Parse(os.Args[2:])
		if *undoSourceDir == "" {
			log.Println("Error: The -source flag is required for the undo command.")
			undoCmd.PrintDefaults()
			os.Exit(1)
		}
		fmt.Println("Sift Undo Operation")
		fmt.Println("--------------------")
		performUndo(*undoSourceDir)

	default:
		fmt.Println("Expected 'organize' or 'undo' subcommands.")
		os.Exit(1)
	}
}
