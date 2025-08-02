# Sift - Your Smart File Organizer

Sift is a simple, fast, and powerful command-line tool built with Go that helps you organize messy directories in seconds. It intelligently sorts files based on a powerful and customizable set of rules.

Most importantly, Sift is built to be safe, featuring a "dry-run" mode to preview all changes and a one-click undo command to revert the last operation.

## Features

- **Undo Functionality**: Instantly revert the last organization with the `sift undo` command.
- **Powerful Rule System**: Organize files by file type or by modification date.
- **Fully Customizable**: Use a `config.yml` file to define your own file-type rules and create folder exclusion lists.
- **Folder Exclusion**: Keep specific folders pristine by ignoring them via a config file or a command-line flag.
- **Safe by Default**: Includes a `-dry-run` mode to preview all changes before they happen.
- **Conflict Resolution**: Automatically renames files (e.g., `photo (1).png`) to prevent overwriting existing data.
- **Cross-Platform**: Built with Go, it can be compiled to run on Windows, macOS, and Linux.

## Installation

To use Sift, you need to have Go installed on your system.

Clone the repository:

```bash
git clone https://github.com/ShreyamKundu/sift.git
cd sift
````

Build the binary:

```bash
go build
```

This will create an executable file named `sift` (or `sift.exe` on Windows) in the directory.

(Optional) Install globally: For easy access from anywhere, you can place the executable in a directory that is in your system's PATH, or use `go install`:

```bash
go install
```

## How to Use

Sift uses two main commands: `organize` and `undo`.

### Example 1: Basic Organization (by File Type)

This is the default mode. It uses built-in rules to sort files into folders like Images, Documents, etc.

```bash
# Always run a dry run first to see the plan!
./sift organize -source="/path/to/Downloads" -dry-run

# If the plan looks good, run it for real
./sift organize -source="/path/to/Downloads"
```

### Example 2: Organizing by Date

Use the `-by-date` flag to sort a folder chronologically. This is perfect for photos or camera uploads.

```bash
# Preview how files will be grouped into folders like "2025/08-August"
./sift organize -source="/Pictures/Camera-Roll" -by-date -dry-run

# Run the date-based organization for real
./sift organize -source="/Pictures/Camera-Roll" -by-date
```

### Example 3: Using a Custom Configuration File

For full control, use the `-config` flag to point to your own `config.yml` file.

```bash
# See a preview using your custom rules
./sift organize -source="/home/user/Desktop/Projects" -config="my_rules.yml" -dry-run

# Run it for real with your custom rules
./sift organize -source="/home/user/Desktop/Projects" -config="my_rules.yml"
```

### Example 4: Excluding Specific Folders

You can ignore folders defined in your `config.yml` (like `Backups`) and also add temporary exclusions with the `-exclude` flag (like `in-progress`).

```bash
# Use custom rules, but ignore both permanent and temporary exclusion folders
./sift organize -source="." -config="my_rules.yml" -exclude="in-progress" -verbose -dry-run
```

### Example 5: Undoing a Mistake

If you're not happy with the result of an organization, you can instantly revert it.

```bash
# Put everything back exactly where it was
./sift undo -source="/Downloads"
```

## Configuration (`config.yml`)

Create a `config.yml` file to have full control over Sift's behavior.

### Example `config.yml`

```yaml
# A list of folder names to completely ignore during organization.
# These folders will never be touched.
exclude_folders:
  - "node_modules"
  - ".git"
  - "Backups"

# Define your custom file type to folder mappings.
rules:
  Images & Graphics:
    - .jpg
    - .jpeg
    - .png
  
  Code & Scripts:
    - .go
    - .js
    - .py
```

## Command Reference

### `sift organize`

* `-source` (required): The path to the directory you want to organize.
* `-dry-run` (optional): Simulates the process without moving any files.
* `-verbose` (optional): Provides detailed logs, like which folders are being skipped.
* `-config` (optional): Path to a custom `config.yml` file.
* `-by-date` (optional): Organizes files by date (YYYY/DD-Month) instead of by file type.
* `-exclude` (optional): A comma-separated list of folder names to temporarily exclude for a single run.

### `sift undo`

* `-source` (required): The path to the directory where the organization was performed (this is where Sift looks for the `.sift_log` file).

