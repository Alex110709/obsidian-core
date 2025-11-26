# Agent Guidelines for Obsidian Core

## Build Commands
- **Build**: `go build ./cmd/obsidiand`
- **Clean build**: `go clean && go build ./cmd/obsidiand`
- **Docker build**: `./build_and_push.sh [version]`

## Test Commands
- **All tests**: `go test ./...`
- **Single test**: `go test -run TestFunctionName ./package`
- **Verbose tests**: `go test -v ./...`
- **Race detection**: `go test -race ./...`

## Lint/Type Check Commands
- **Vet**: `go vet ./...`
- **Format check**: `gofmt -l .` (should return no output)
- **Format fix**: `gofmt -w .`

## Code Style Guidelines

### Imports
- Group imports: standard library, blank line, third-party, blank line, local packages
- Use full import paths (no dot imports)
- Example:
```go
import (
    "fmt"
    "math/big"

    "github.com/btcsuite/btcutil"
    "golang.org/x/net"

    "obsidian-core/chaincfg"
    "obsidian-core/wire"
)
```

### Naming Conventions
- **Packages**: lowercase, single word (e.g., `blockchain`, `consensus`)
- **Functions/Methods**: PascalCase for exported, camelCase for unexported
- **Variables**: camelCase, descriptive names
- **Constants**: PascalCase for exported, camelCase for unexported
- **Types**: PascalCase for exported structs/interfaces

### Error Handling
- Return errors as last return value
- Use `fmt.Errorf` for wrapping errors: `return fmt.Errorf("failed to connect: %v", err)`
- Check errors immediately after operations
- Don't ignore errors (use `_` only when intentionally ignoring)

### Types and Structs
- Use pointer receivers for methods that modify the receiver
- Define interfaces where abstraction is needed
- Use meaningful field names in structs

### Comments
- No comments needed for self-explanatory code
- Comment exported functions/types with complete sentences
- Use `//` for single-line comments

### Formatting
- Use `gofmt` for consistent formatting
- 4-space indentation (Go standard)
- Max line length: reasonable, break long lines naturally

### Testing
- Test files end with `_test.go`
- Test functions start with `Test` followed by PascalCase name
- Use table-driven tests for multiple test cases
- Test both success and error paths</content>
<parameter name="filePath">/Users/yuchan/Desktop/Obsidian Chain/obsidian-core/AGENTS.md