# Obsidian Core Go Implementation

This is the reference implementation of Obsidian Core written in Go. It provides a complete, production-ready cryptocurrency node with all features including shielded transactions, mining, and networking.

## Features

- ✅ Full node implementation
- ✅ CPU mining (DarkMatter PoW)
- ✅ Stratum mining pool server
- ✅ JSON-RPC API
- ✅ P2P networking with Tor support
- ✅ Shielded transactions (zk-SNARK-inspired)
- ✅ Persistent storage (BoltDB)
- ✅ Encrypted memos
- ✅ Viewing keys for privacy

## Project Structure

```
obsidian-core/
├── cmd/obsidiand/          # Main daemon entry point
├── blockchain/             # Blockchain state management
├── consensus/              # DarkMatter PoW implementation
├── chaincfg/               # Network parameters and configuration
├── wire/                   # Wire protocol data structures
├── network/                # P2P networking and peer management
├── mining/                 # CPU miner implementation
├── stratum/                # Stratum mining pool server
├── rpc/                    # JSON-RPC API server
├── crypto/                 # Cryptographic functions
├── database/               # BoltDB storage layer
├── tor/                    # Tor integration
└── docs/                   # Documentation
```

## Quick Start

### Prerequisites

- Go 1.20+
- Git

### Installation

```bash
git clone https://github.com/your-org/obsidian-core.git
cd obsidian-core
go mod tidy
go build ./cmd/obsidiand
```

### Running

```bash
./obsidiand
```

Or with environment variables:

```bash
SOLO_MINING=true MINER_ADDRESS=YourAddress ./obsidiand
```

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `SOLO_MINING` | `"true"` | Enable solo CPU mining |
| `POOL_SERVER` | `"false"` | Enable Stratum pool server |
| `MINER_ADDRESS` | Auto-generated | Mining reward address |
| `POOL_LISTEN` | `"0.0.0.0:3333"` | Stratum server address |
| `TOR_ENABLED` | `"false"` | Enable Tor networking |
| `RPC_ADDR` | `"0.0.0.0:8545"` | RPC server address |
| `DATA_DIR` | `"./"` | Data directory |

### Network Parameters

Located in `chaincfg/params.go`:

```go
type Params struct {
    Name              string
    BaseBlockReward   int64
    HalvingInterval   int64
    MaxMoney          int64
    BlockMaxSize      int32
    TargetTimePerBlock time.Duration
    // ... more parameters
}
```

## Core Components

### Blockchain Engine

```go
// Initialize blockchain
params := chaincfg.MainNetParams
pow := consensus.NewDarkMatterEngine()
bc, err := blockchain.NewBlockchain(&params, pow)

// Add new block
block := wire.NewBlock(header, transactions)
err := bc.ProcessBlock(block)
```

### Mining

```go
// Start CPU miner
miner := mining.NewCPUMiner(blockchain, &params, pow, minerAddress)
go miner.Start()
```

### Networking

```go
// Start P2P server
server := network.NewServer(&params)
server.Start()
```

### RPC API

```go
// Start RPC server
rpcServer := rpc.NewServer(blockchain, miner, "0.0.0.0:8545")
rpcServer.Start()
```

## API Reference

### JSON-RPC Methods

- `getblockcount` - Get current block height
- `getbestblockhash` - Get best block hash
- `getblock` - Get block by hash
- `getmininginfo` - Get mining information
- `z_getnewaddress` - Generate shielded address
- `z_sendmany` - Send shielded transaction
- `z_getbalance` - Get shielded balance

### Wire Protocol

See `wire/` package for message structures:

```go
type BlockHeader struct {
    Version       int32
    PrevBlock     Hash
    MerkleRoot    Hash
    Timestamp     time.Time
    Bits          uint32
    Nonce         uint32
    // Shielded transaction root
    ShieldedRoot  Hash
}
```

## Development

### Testing

```bash
# Run all tests
go test ./...

# Run specific test
go test -run TestDarkMatterDeterministic ./consensus

# Run with race detection
go test -race ./...
```

### Building

```bash
# Standard build
go build ./cmd/obsidiand

# Cross-compilation
GOOS=linux GOARCH=amd64 go build ./cmd/obsidiand

# Docker build
docker build -t obsidian-node .
```

### Code Style

Follow Go conventions:
- Use `gofmt` for formatting
- Run `go vet` for static analysis
- Write tests for new features
- Use meaningful variable names

## Deployment

### Docker

```bash
docker run -d \
  -p 8333:8333 \
  -p 8545:8545 \
  -v obsidian-data:/root \
  yuchanshin/obsidian-node:latest
```

### Docker Compose

```yaml
services:
  obsidian-node:
    image: yuchanshin/obsidian-node:latest
    environment:
      - SOLO_MINING=true
      - MINER_ADDRESS=YourAddress
    ports:
      - "8333:8333"
      - "8545:8545"
    volumes:
      - obsidian-data:/root
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Write tests for new functionality
4. Ensure all tests pass
5. Submit a pull request

## License

This implementation is licensed under the MIT License.</content>
<parameter name="filePath">/Users/yuchan/Desktop/Obsidian Chain/obsidian-core/docs/go-implementation.md