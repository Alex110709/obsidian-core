# Obsidian Core

A privacy-focused cryptocurrency with shielded transactions, encrypted memos, and custom token support.

## Features

- **Privacy**: Shielded transactions with zk-SNARK-inspired proofs
- **Tokens**: Custom token creation and transfer without smart contracts
- **Fast**: 2-minute block time with 3.2MB blocks
- **Secure**: Bitcoin-compatible difficulty adjustment
- **Anonymous**: Optional Tor networking
- **Multi-platform**: Linux, macOS, Windows support

## Specifications

- **Block Time**: 2 minutes
- **Block Size**: 3.2MB
- **Total Supply**: 100M OBS
- **Consensus**: DarkMatter PoW (AES/SHA256 hybrid)
- **Privacy**: Shielded addresses (zobs*), encrypted memos
- **Tokens**: ERC-20 style tokens with shielded transfers

## Privacy Features (NEW in v1.1.0)

### Shielded Transactions
Obsidian implements Zcash-style shielded transactions for complete privacy:

- **Shielded Addresses (z-addresses)**: Start with `zobs` prefix
- **Transparent Addresses (t-addresses)**: Standard addresses starting with `Obs`
- **Private Transfers**: Amount, sender, and receiver hidden using cryptographic proofs
- **Encrypted Memos**: Attach encrypted messages to transactions (up to 512 bytes)
- **Selective Disclosure**: View-only keys allow read access without spending permission

### Transaction Types
1. **Transparent (t→t)**: Standard public transactions
2. **Shielded (z→z)**: Fully private transactions
3. **Shielding (t→z)**: Move funds from transparent to shielded pool
4. **Deshielding (z→t)**: Move funds from shielded to transparent pool

## Token System

Create and transfer custom tokens without smart contracts. Supports both transparent and shielded token transactions.

**RPC Methods:**
- `issuetoken` - Create new tokens
- `transfertoken` - Transfer tokens
- `shieldtoken` - Private token transfers
- `gettokenbalance` - Check balances

See [Token Guide](./docs/token-guide.md) for detailed documentation.

## Getting Started

### Prerequisites
- Go 1.20+
- Git

### Building from Source

#### Linux
```bash
# Install Go (Ubuntu/Debian)
sudo apt update
sudo apt install golang-go

# Clone and build
git clone https://github.com/your-org/obsidian-core.git
cd obsidian-core
go mod tidy
go build ./cmd/obsidiand

# Run
./obsidiand
```

#### macOS
```bash
# Install Go using Homebrew
brew install go

# Or download from https://golang.org/dl/

# Clone and build
git clone https://github.com/your-org/obsidian-core.git
cd obsidian-core
go mod tidy
go build ./cmd/obsidiand

# Run
./obsidiand
```

#### Windows
```powershell
# Install Go from https://golang.org/dl/
# Or using Chocolatey
choco install golang

# Clone and build
git clone https://github.com/your-org/obsidian-core.git
cd obsidian-core
go mod tidy
go build ./cmd/obsidiand

# Run
.\obsidiand.exe
```

#### Cross-Platform Build
```bash
# Build for Linux (from any platform)
GOOS=linux GOARCH=amd64 go build ./cmd/obsidiand

# Build for Windows (from any platform)
GOOS=windows GOARCH=amd64 go build ./cmd/obsidiand

# Build for macOS (from any platform)
GOOS=darwin GOARCH=amd64 go build ./cmd/obsidiand
```

### Running the Node
```bash
# Run directly (development)
go run cmd/obsidiand/main.go

# Or run built binary
./obsidiand
```

This will start the node, initialize the blockchain with the Genesis block, and start a CPU miner simulation.

### Environment Variables
```bash
# Mining configuration
SOLO_MINING=true MINER_ADDRESS=YourAddress ./obsidiand

# Pool server
POOL_SERVER=true POOL_LISTEN=0.0.0.0:3333 ./obsidiand

# Tor networking
TOR_ENABLED=true ./obsidiand

# Custom data directory
DATA_DIR=./my-data ./obsidiand
```

## Documentation

- [Protocol Specification](./docs/protocol-spec.md) - Wire protocol and consensus rules
- [Build Guide](./docs/build-guide.md) - Platform-specific build instructions
- [Token Guide](./docs/token-guide.md) - Custom token system documentation
- [Go Implementation](./docs/go-implementation.md) - Reference implementation details
- [Rust Implementation](./docs/rust-implementation.md) - High-performance alternative
- [Python Implementation](./docs/python-implementation.md) - Research and experimentation
- [JavaScript Implementation](./docs/javascript-implementation.md) - Web integration guide
