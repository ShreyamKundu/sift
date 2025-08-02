# Sift - Your Smart File Organizer

`Sift` is a simple, fast, and powerful command-line tool built with Go that helps you organize messy directories in seconds. It intelligently sorts files into categorized folders based on their file type, making it easy to clean up your "Downloads" or any other folder.

---

## Features

- **Intelligent Categorization:** Sorts files into folders like `Images`, `Documents`, `Audio`, `Videos`, `Archives`, and `Others`.
- **Recursive Scanning:** Cleans up files in the main directory and all its subdirectories.
- **Safe by Default:** Includes a `-dry-run` mode to preview changes before they happen.
- **Conflict Resolution:** Automatically renames files (e.g., `photo (1).png`) to prevent overwriting existing files.
- **Cross-Platform:** Built with Go, it can be compiled to run on Windows, macOS, and Linux.
- **Verbose Mode:** Use the `-verbose` flag for detailed output, perfect for debugging.

---

## Installation

To use Sift, you need to have Go installed on your system.

1.  **Clone the repository:**

    ```bash
    git clone https://github.com/ShreyamKundu/sift.git
    cd sift
    ```

2.  **Build the binary:**

    ```bash
    go build
    ```

    This will create an executable file named `sift` (or `sift.exe` on Windows) in the directory.

3.  **(Optional) Install globally:** For easy access from anywhere, you can place the executable in a directory that is in your system's PATH, or use `go install`:
    ```bash
    go install
    ```

---

## How to Use

The basic command structure is:

```bash
./sift -source="<path_to_your_folder>"
```
