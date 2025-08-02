# Sift - Your Smart File Organizer

`Sift` is a simple, fast, and powerful command-line tool built with Go that helps you organize messy directories in seconds. It intelligently sorts files into categorized folders based on their file type or modification date, making it easy to clean up your "Downloads" or any other folder.

---

## Features

- **Two Organization Modes:** Sort files by **file type** (e.g., `.jpg` → `Images`) or by **date** (`YYYY/MM-Month`).
- **Custom Rules:** Use a `config.yml` file to define your own custom folders and file-type mappings.
- **Recursive Scanning:** Cleans up files in the main directory and all its subdirectories.
- **Safe by Default:** Includes a `-dry-run` mode to preview all changes before they happen.
- **Conflict Resolution:** Automatically renames files (e.g., `photo (1).png`) to prevent overwriting existing data.
- **Cross-Platform:** Built with Go, it can be compiled to run on Windows, macOS, and Linux.
- **Verbose Mode:** Use the `-verbose` flag for detailed output, perfect for debugging.

---

## Installation

To use Sift, you need to have **[Go](https://go.dev/doc/install)** installed on your system.

1. **Clone the repository:**

   ```bash
   git clone https://github.com/ShreyamKundu/sift.git
   cd sift
   ```

2. **Build the binary:**

   ```bash
   go build
   ```

   This will create an executable file named `sift` (or `sift.exe` on Windows) in the directory.

3. **(Optional) Install globally:** For easy access from anywhere, you can place the executable in a directory that is in your system's PATH, or use `go install`:
   ```bash
   go install
   ```

---

## How to Use

The basic command structure is:

```bash
./sift -source="<path_to_your_folder>" [flags]
```

### Example 1: Organize by File Type (Default)

This is the standard mode. It uses built-in rules to sort files into folders like `Images`, `Documents`, etc.

```bash
# Always run a dry run first to see the plan
./sift -source="/path/to/your/Downloads" -dry-run

# If the plan looks good, run it for real
./sift -source="/path/to/your/Downloads"
```

### Example 2: Organize by File Type with Custom Rules

Use the `-config` flag to point to your own `config.yml` file for complete control over the folder structure.

```bash
# Use your custom rules in a dry run
./sift -source="~/Desktop/Projects" -config="my_rules.yml" -dry-run

# Run it for real with your custom rules
./sift -source="~/Desktop/Projects" -config="my_rules.yml"
```

### Example 3: Organize by Date

Use the `-by-date` flag to sort files into folders based on their modification date. This ignores file types.

```bash
# See how files would be sorted by date
./sift -source="~/Pictures/Camera-Roll" -by-date -dry-run

# Run the date-based organization for real
./sift -source="~/Pictures/Camera-Roll" -by-date
```

### Command-Line Flags

- `-source` (required): The path to the directory you want to organize.
- `-dry-run` (optional): Simulates the process without moving any files.
- `-verbose` (optional): Provides detailed logs, such as which directories are being skipped.
- `-config` (optional): Path to a custom `config.yml` file to define your own file-type rules.
- `-by-date` (optional): Organizes files by date (`YYYY/DD-Month`) instead of by file type.
