# Synology Cloud Sync Decrypt Tool - GUI Version

A graphical user interface version of the Synology Cloud Sync decryption tool, built with Fyne framework.

## Features

- 🎨 Intuitive and user-friendly interface
- 🔒 Password input with masking for security
- 🗂️ File and directory selection dialogs
- 📊 Real-time log display with color-coded entries
- ⏳ Progress indication
- 📁 Support for both single file and batch directory decryption
- 🖥️ Cross-platform support (Windows, macOS, Linux)

## Screenshots

![GUI Interface](https://via.placeholder.com/800x600?text=Synology+Decrypt+GUI)

## Installation

### Prerequisites

- Go 1.21 or higher
- lz4 command-line tool

### Installing lz4

```bash
# macOS
brew install lz4

# Ubuntu/Debian
apt-get install lz4

# Fedora/RHEL
dnf install lz4

# Arch Linux
pacman -S lz4
```

### Build from Source

```bash
# Clone repository
git clone git@github.com:cdredfox/synology-cloud-sync-decrypt-tool.git
cd synology-cloud-sync-decrypt-tool

# Switch to GUI branch (if not on main)
git checkout fyne-gui-version

# Build GUI version
go build -o syndecrypt-gui cmd/syndecrypt-gui/main.go

# Run the GUI application
./syndecrypt-gui
```

## Usage

### Basic Workflow

1. **Launch the Application**
   ```bash
   ./syndecrypt-gui
   ```

2. **Enter Password**
   - Type the decryption password in the password field
   - The password is masked for security

3. **Select Input**
   - Click "Browse" next to "Input (File/Dir)"
   - Choose whether you want to decrypt:
     - **Single File**: Select one encrypted file (.cse, .enc, .cloudsync, .csenc)
     - **Directory**: Select a folder containing encrypted files

4. **Select Output Directory**
   - Click "Browse" next to "Output Directory"
   - Choose where to save decrypted files
   - Default is "output" folder in current directory

5. **Configure Options**
   - **Process subdirectories recursively**: Check this if you want to decrypt all files in subdirectories (automatically enabled when selecting a directory)

6. **Start Decryption**
   - Click "Start Decryption" button
   - Watch the progress in the log area
   - Decrypted files will be saved to the output directory

### Log Messages Explained

The log area displays color-coded messages:

- **Green (Success)**: ✅ File decrypted successfully
- **Red (Error)**: ❌ Decryption failed with error message
- **Yellow (Progress)**: ⏳ Current operation information
- **Gray (Info)**: General information messages

### Example Log Output

```
[17:23:45] Starting decryption: /Users/user/encrypted/backup.cse
[17:23:45] Output directory: /Users/user/decrypted
[17:23:45] Decrypting: backup.cse
[17:23:48] ✅ Successfully decrypted: backup.cse
[17:23:48] Output saved to: /Users/user/decrypted/backup (Time: 3.2s)
```

## Keyboard Shortcuts

- `Enter`: Start decryption (when form is filled)
- `Escape`: Cancel operation

## Troubleshooting

### lz4 Not Found

If you see errors about lz4, ensure it's installed:
```bash
# Verify installation
which lz4

# Install if missing
# macOS: brew install lz4
# Linux: apt-get install lz4 (Debian/Ubuntu)
# Linux: dnf install lz4 (Fedora/RHEL)
```

### Application Won't Start

If the GUI application doesn't start:
1. Check that you have a display (X11 on Linux, or running in a desktop environment)
2. Try running from terminal to see error messages: `./syndecrypt-gui 2>&1`
3. Ensure all dependencies are installed: `go mod download`

### Decryption Fails

If decryption fails:
1. Verify the password is correct
2. Ensure the encrypted file is not corrupted
3. Check that the file has a supported extension (.cse, .enc, .cloudsync, .csenc)
4. Try the command-line version to compare: `./syndecrypt -p yourpassword -O output/ file.cse`

## Building for Different Platforms

### Windows
```bash
GOOS=windows GOARCH=amd64 go build -o syndecrypt-gui.exe cmd/syndecrypt-gui/main.go
```

### macOS (Intel)
```bash
GOOS=darwin GOARCH=amd64 go build -o syndecrypt-gui-mac-intel cmd/syndecrypt-gui/main.go
```

### macOS (Apple Silicon)
```bash
GOOS=darwin GOARCH=arm64 go build -o syndecrypt-gui-mac-m1 cmd/syndecrypt-gui/main.go
```

### Linux
```bash
GOOS=linux GOARCH=amd64 go build -o syndecrypt-gui-linux cmd/syndecrypt-gui/main.go
```

## Differences from Command-Line Version

| Feature | Command-Line | GUI Version |
|---------|--------------|-------------|
| Password Input | Direct in command | Password field with masking |
| File Selection | Path as argument | File dialog browser |
| Progress | Text output | Visual progress bar + logs |
| Batch Processing | Multiple paths | Directory selection |
| Ease of Use | Requires command knowledge | Point-and-click interface |
| Best For | Scripts, automation | Interactive use |

## Technical Details

- **Framework**: Fyne v2.7.1
- **Language**: Go 1.21+
- **GUI Toolkit**: Native OpenGL/DirectX with platform abstraction
- **Size**: ~24MB (includes Fyne framework)
- **Memory**: Efficient rendering, minimal resources

## Contributing

To contribute to the GUI version:

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/amazing-feature`
3. Make your changes
4. Test the application: `go run cmd/syndecrypt-gui/main.go`
5. Commit and push your changes
6. Open a Pull Request

## License

Same as the main project - MIT License

## Support

For issues and questions:
1. Check the troubleshooting section above
2. Open an issue on GitHub
3. Provide the log output from the application
4. Include your OS and Go version information
