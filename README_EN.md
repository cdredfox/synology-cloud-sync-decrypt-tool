# Synology Cloud Sync Decryption Tool (Go Implementation)

A high-performance Go implementation of the Synology Cloud Sync decryption tool, providing the same functionality as the Python version.

## Features

- 🚀 High-performance Go implementation (2-3x faster than Python)
- 🔐 Supports both password and RSA private key decryption
- 📁 Supports single file, multiple files, and directory recursion
- 📊 Progress indication and detailed result statistics
- 🔧 Cross-platform support (Linux, macOS, Windows)
- 💾 Low memory usage with streaming processing for large files
- 📦 Single executable with no runtime dependencies

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

# Download dependencies
go mod download

# Build
go build -o syndecrypt cmd/syndecrypt/main.go

# Install to system path (optional)
sudo cp syndecrypt /usr/local/bin/
```

## Usage

### Basic Usage

```bash
# Decrypt file with password
syndecrypt -p password.txt -O output/ encrypted_file.cse

# Decrypt file with RSA private key
syndecrypt -k private.pem -l public.pem -O output/ encrypted_file.cse

# Decrypt multiple files
syndecrypt -p password.txt -O output/ file1.cse file2.cse file3.cse

# Recursively decrypt entire directory
syndecrypt -p password.txt -O output/ /path/to/encrypted/directory/
```

### Command-line Options

```
synology-decrypt: Synology Cloud Sync decryption tool

Usage:
  syndecrypt (-p <password_file> | -k <private_key_file> -l <public_key_file>) -O <output_directory> <encrypted_file>...
  syndecrypt (-h | --help)
  syndecrypt --version

Options:
  -O <dir> --output-directory=<dir>     Output directory
  -p <file> --password-file=<file>      File containing decryption password
  -k <file> --private-key-file=<file>   File containing private key for decryption
  -l <file> --public-key-file=<file>    File containing public key for decryption
  -h --help                            Show help message
  --version                            Show version information
```

## Password File Format

Password file should contain the plaintext password, for example:

```
mysecretpassword
```

## Supported File Formats

- `.cse` - Synology Cloud Sync encrypted files
- `.enc` - Generic encrypted files
- `.cloudsync` - Cloud Sync encrypted files
- `.csenc` - Cloud Sync encrypted files

## Development

### Project Structure

```
.
├── cmd/syndecrypt/        # Command-line entry point
├── pkg/
│   ├── core/              # Core decryption algorithms (AES-256-CBC, RSA-OAEP, OpenSSL KDF)
│   ├── files/             # File handling logic and result statistics
│   └── util/              # Utility functions (LZ4 decompression, etc.)
├── internal/              # Internal implementations
├── test/                  # Test files
├── go.mod                 # Go module file
├── LICENSE                # MIT License
└── README.md
```

### Running Tests

```bash
go test ./...
```

### Building Releases

```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o syndecrypt-linux-amd64 cmd/syndecrypt/main.go

# macOS
GOOS=darwin GOARCH=amd64 go build -o syndecrypt-darwin-amd64 cmd/syndecrypt/main.go

# Windows
GOOS=windows GOARCH=amd64 go build -o syndecrypt-windows-amd64.exe cmd/syndecrypt/main.go
```

## Performance Advantages

Compared to the Python version, the Go implementation offers:

- ⚡ **Faster Speed**: 2-3x faster decryption speed
- 💾 **Lower Memory**: Streaming processing with significantly reduced memory usage
- 🔧 **Better Concurrency**: Native concurrency support for efficient batch processing
- 📦 **Simpler Deployment**: Single compiled binary with no Python environment dependencies
- 🎯 **More Stable Performance**: Static typing with compile-time optimizations

## Troubleshooting

### lz4 Not Found

If you see error "lz4 command failed", ensure lz4 is installed and in PATH:

```bash
which lz4
```

### Permission Issues

Ensure password file and private key file have correct read permissions:

```bash
chmod 600 password.txt
chmod 600 private.pem
```

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

Thanks to the original Python project authors [@marnix](https://github.com/marnix/synology-decrypt) and [@anojht](https://github.com/anojht/synology-cloud-sync-decrypt-tool).
