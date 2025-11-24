# Building Obsidian Core

This guide covers building Obsidian Core from source on different platforms and environments.

## Prerequisites

### System Requirements
- **Go**: Version 1.20 or later
- **Git**: For cloning the repository
- **RAM**: Minimum 4GB (8GB recommended)
- **Disk**: Minimum 10GB free space
- **OS**: Linux, macOS, or Windows

### Go Installation

#### Linux (Ubuntu/Debian)
```bash
# Update package list
sudo apt update

# Install Go
sudo apt install golang-go

# Verify installation
go version
```

#### Linux (CentOS/RHEL/Fedora)
```bash
# CentOS/RHEL
sudo yum install golang

# Fedora
sudo dnf install golang

# Verify
go version
```

#### macOS
```bash
# Using Homebrew (recommended)
brew install go

# Or download from official site
# Visit: https://golang.org/dl/
# Download and install the .pkg file

# Verify
go version
```

#### Windows
```powershell
# Using Chocolatey
choco install golang

# Or download from official site
# Visit: https://golang.org/dl/
# Download and install the .msi file

# Verify (PowerShell)
go version
```

## Building from Source

### Clone Repository
```bash
git clone https://github.com/your-org/obsidian-core.git
cd obsidian-core
```

### Download Dependencies
```bash
go mod tidy
```

### Build Binary
```bash
go build ./cmd/obsidiand
```

### Verify Build
```bash
# Check if binary was created
ls -la obsidiand  # Linux/macOS
dir obsidiand.exe # Windows

# Run version check (if implemented)
./obsidiand --version
```

## Platform-Specific Builds

### Linux (x86_64)
```bash
# Native build
go build ./cmd/obsidiand

# Optimized build
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" ./cmd/obsidiand
```

### Linux (ARM64)
```bash
# For Raspberry Pi or ARM servers
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build ./cmd/obsidiand
```

### macOS (Intel)
```bash
# Native build on Intel Mac
go build ./cmd/obsidiand

# Cross-compile from Intel to Apple Silicon
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build ./cmd/obsidiand
```

### macOS (Apple Silicon)
```bash
# Native build on Apple Silicon Mac
go build ./cmd/obsidiand

# Cross-compile from Apple Silicon to Intel
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build ./cmd/obsidiand
```

### Windows (x86_64)
```powershell
# Native build
go build ./cmd/obsidiand

# Optimized build
$env:CGO_ENABLED="0"
$env:GOOS="windows"
$env:GOARCH="amd64"
go build -ldflags="-s -w" ./cmd/obsidiand
```

### Windows (ARM64)
```powershell
# For Windows on ARM devices
$env:CGO_ENABLED="0"
$env:GOOS="windows"
$env:GOARCH="arm64"
go build ./cmd/obsidiand
```

## Build Options

### Optimization Flags
```bash
# Strip debug information and symbols
go build -ldflags="-s -w" ./cmd/obsidiand

# Include version information
go build -ldflags="-X main.version=1.0.0 -X main.commit=$(git rev-parse HEAD)" ./cmd/obsidiand

# Static linking (no external dependencies)
CGO_ENABLED=0 go build -ldflags="-extldflags '-static'" ./cmd/obsidiand
```

### Debug Build
```bash
# Include debug information
go build -gcflags="all=-N -l" ./cmd/obsidiand

# Enable race detection
go build -race ./cmd/obsidiand
```

## Testing Build

### Run Tests
```bash
# Run all tests
go test ./...

# Run specific package tests
go test ./blockchain
go test ./consensus

# Run with verbose output
go test -v ./...

# Run with race detection
go test -race ./...

# Run benchmarks
go test -bench=. ./...
```

### Integration Testing
```bash
# Build and run integration tests
go test -tags=integration ./...
```

## Docker Build

### Using Docker Compose (Recommended)
```bash
# Quick start
docker compose up -d

# Build custom image
docker compose build
```

### Manual Docker Build
```bash
# Build Docker image
docker build -t obsidian-node .

# Run container
docker run -d -p 8333:8333 -p 8545:8545 obsidian-node
```

### Multi-Architecture Build
```bash
# Build for multiple platforms
docker buildx build --platform linux/amd64,linux/arm64 -t obsidian-node .
```

## CI/CD Build

### GitHub Actions Example
```yaml
name: Build
on: [push, pull_request]

jobs:
  build:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        go-version: [1.20.x, 1.21.x]

    steps:
    - uses: actions/checkout@v3

    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: ${{ matrix.go-version }}

    - name: Cache dependencies
      uses: actions/cache@v3
      with:
        path: ~/go/pkg/mod
        key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}

    - name: Download dependencies
      run: go mod download

    - name: Build
      run: go build ./cmd/obsidiand

    - name: Test
      run: go test ./...

    - name: Upload artifacts
      uses: actions/upload-artifact@v3
      with:
        name: obsidian-${{ runner.os }}
        path: obsidiand
```

### Jenkins Pipeline Example
```groovy
pipeline {
    agent any

    stages {
        stage('Checkout') {
            steps {
                git 'https://github.com/your-org/obsidian-core.git'
            }
        }

        stage('Setup Go') {
            steps {
                sh 'go version'
            }
        }

        stage('Build') {
            steps {
                sh 'go mod tidy'
                sh 'go build ./cmd/obsidiand'
            }
        }

        stage('Test') {
            steps {
                sh 'go test ./...'
            }
        }

        stage('Archive') {
            steps {
                archiveArtifacts artifacts: 'obsidiand', fingerprint: true
            }
        }
    }
}
```

## Troubleshooting

### Common Build Issues

#### Go Version Too Old
```
Error: requires go >= 1.20
```
**Solution**: Upgrade Go to version 1.20 or later.

#### Missing Dependencies
```
go: missing module
```
**Solution**: Run `go mod tidy` to download dependencies.

#### CGO Issues on Windows
```
cgo: exec gcc: exec: "gcc": executable file not found
```
**Solution**: Install MinGW or use `CGO_ENABLED=0` for static builds.

#### macOS Code Signing
```
code signature not valid
```
**Solution**: Run `xattr -rd com.apple.quarantine obsidiand` or build with `CGO_ENABLED=0`.

### Performance Optimization

#### Compiler Optimizations
```bash
# Enable optimizations
go build -gcflags="-l=4" ./cmd/obsidiand

# Profile-guided optimization (Go 1.20+)
go build -pgo=auto ./cmd/obsidiand
```

#### Memory Usage
```bash
# Reduce binary size
go build -ldflags="-s -w" ./cmd/obsidiand

# Static linking
CGO_ENABLED=0 go build ./cmd/obsidiand
```

## Distribution

### Creating Releases
```bash
# Create release directory
mkdir release
cp obsidiand release/

# Create archives
tar -czf obsidian-linux-amd64.tar.gz -C release .
zip obsidian-windows-amd64.zip release/*

# Generate checksums
sha256sum obsidian-*.tar.gz obsidian-*.zip > SHA256SUMS
```

### Package Managers

#### Linux Packages
```bash
# Debian/Ubuntu
# Create .deb package using dpkg-deb

# RPM-based
# Create .rpm package using rpmbuild
```

#### macOS
```bash
# Create .dmg installer
# Use create-dmg tool
```

#### Windows
```bash
# Create .msi installer
# Use WiX Toolset or NSIS
```

## Contributing

When contributing code changes:

1. Ensure your changes build on all supported platforms
2. Run the full test suite: `go test ./...`
3. Follow the existing code style
4. Update documentation if needed

### Pre-commit Hooks
```bash
# Install pre-commit hooks
go install honnef.co/go/tools/cmd/staticcheck@latest
go install golang.org/x/lint/golint@latest

# Run checks
go vet ./...
staticcheck ./...
golint ./...
```

## Support

If you encounter build issues:

1. Check the [GitHub Issues](https://github.com/your-org/obsidian-core/issues)
2. Verify your Go version: `go version`
3. Check dependencies: `go mod verify`
4. Try a clean build: `go clean -cache && go mod tidy && go build`

## License

This build guide is part of the Obsidian Core project, licensed under the MIT License.</content>
<parameter name="filePath">/Users/yuchan/Desktop/Obsidian Chain/obsidian-core/docs/build-guide.md