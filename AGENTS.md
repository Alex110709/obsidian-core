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
- Test both success and error paths

## Smart Contract Development

### OCL Language (Obsidian Contract Language)
- Python-like syntax for smart contracts
- Indentation-based block structure
- Supported constructs: functions, if/elif/else, assignments, expressions
- Built-in functions: self.balance, self.storage, send(), etc.

### Example Contract
```ocl
contract MyContract:
    def __init__(self):
        self.owner = msg.sender
        self.balance = 0

    def deposit(self, amount):
        if amount > 0:
            self.balance += amount

    def withdraw(self, amount):
        if self.balance >= amount:
            self.balance -= amount
            send(msg.sender, amount)
```

### Smart Contract Commands
- **Compile contract**: Use smartcontract package to compile OCL to bytecode
- **Deploy contract**: `deploycontract <contract_code>` RPC method
- **Call contract**: `callcontract <address> <function> [args...]` RPC method

## Address Formats

### Transparent Addresses
- **Prefix**: `obs`
- **Format**: obs + base58 encoded hash
- **Example**: obs5pPyd6DA6tYyYwip4hYcBWWFTNf4wj8nn

### Shielded Addresses
- **Prefix**: `zobs`
- **Format**: zobs + base58 encoded shielded data
- **Example**: zobs + encrypted note data

## Privacy Features

### Shield/Unshield Transactions
- **Shield**: Convert transparent funds to shielded: `shield <from> <to> <amount>`
- **Unshield**: Convert shielded funds to transparent: `unshield <from> <to> <amount>`
- Uses zk-SNARK proofs for privacy

## RPC Methods

### New Methods Added
- `deploycontract`: Deploy smart contract
- `callcontract`: Call smart contract function
- `shield`: Shield funds to private address
- `unshield`: Unshield funds to public address
- `z_sendmany`: Send to multiple shielded addresses
- `z_getnewaddress`: Generate new shielded address
- `z_getbalance`: Get shielded balance
- `z_listaddresses`: List shielded addresses

### Usage Examples
```bash
# Deploy contract
curl -X POST -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"deploycontract","params":["contract MyContract:\n    def hello(self):\n        return \"Hello World\""],"id":1}' \
  http://localhost:8545

# Call contract
curl -X POST -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"callcontract","params":["obs123...","hello"],"id":2}' \
  http://localhost:8545

# Shield funds
curl -X POST -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"shield","params":["obs123...","zobs456...",1000],"id":3}' \
  http://localhost:8545
```</content>
<parameter name="filePath">/Users/yuchan/Desktop/Obsidian Chain/obsidian-core/AGENTS.md