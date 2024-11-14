Here’s a `README.md` file for your project, **Cabinete**:

```markdown
# Cabinete

Cabinete is a command-line tool for organizing photos (or other files) by creation date. It scans a specified directory, identifies the creation date of each file, and moves files into subdirectories named by the day of the month they were created (e.g., `01`, `02`, ... `31`). The tool provides a live interface that displays the directories being created, the number of files in each, and a progress counter of files processed versus remaining files.

## Features

- Organizes files by day of creation.
- Creates subdirectories for each day in the month.
- Provides a live interface to show the organization progress:
  - Displays the day directories and the total file count for each.
  - Shows a real-time counter for processed and pending files.
  
## Installation

### Prerequisites

- **Go** (v1.17+)
- **Cobra** and **tview** libraries (automatically installed through `go.mod`)

### Clone the Repository

```bash
git clone https://github.com/your-username/cabinete.git
cd cabinete
```

### Build the Binary

To compile the project, use the provided Makefile:

```bash
make build
```

### Install

To install the `cabinete` binary into `/usr/local/bin`:

```bash
sudo make install
```

## Usage

### Basic Usage

Run the command specifying the directory to organize with the `-d` flag:

```bash
cabinete -d /path/to/your/directory
```

### Command-Line Options

- `-d, --dir`: Directory path containing files to organize (required).

### Example

Organize photos in `/Users/johndoe/Photos`:

```bash
cabinete -d /Users/johndoe/Photos
```

This command will organize photos into subdirectories based on their creation day.

## Development

### Project Structure

```
cabinete/
├── cmd/                # Contains Cobra command implementations
├── main.go             # Entry point for the application
├── Makefile            # Build and installation commands
├── go.mod              # Go modules configuration
└── README.md           # Project README
```

### Makefile Commands

- **`make build`**: Compiles the `cabinete` binary into the current directory.
- **`make install`**: Installs the binary to `/usr/local/bin`.
- **`make clean`**: Cleans up compiled binaries.

## Contributing

Contributions are welcome! If you have ideas for new features, find a bug, or would like to improve the documentation, feel free to open an issue or submit a pull request.

### How to Contribute

1. Fork the repository.
2. Create a new branch for your feature or fix.
3. Make your changes and add tests if applicable.
4. Open a pull request detailing the changes you’ve made.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

---

Enjoy organizing your files with **Cabinete**! 🎉
```

### Explanation

This `README.md` includes:
- **Project Description**: Overview and core functionality.
- **Installation**: How to build and install the tool.
- **Usage Instructions**: Basic command options and an example.
- **Development Guide**: Project structure and makefile commands.
- **Contribution Guidelines**: Information on contributing to the project.
- **License**: Standard license information.
