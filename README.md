# Obsidian Core

A privacy-focused cryptocurrency with shielded transactions, encrypted memos, and custom token support.

## Features

- **Privacy**: Shielded transactions with zk-SNARK-inspired proofs
- **Tokens**: Custom token creation and transfer without smart contracts
- **Fast**: 20-second block time with 30M gas limit
- **Secure**: Bitcoin-compatible difficulty adjustment
- **Anonymous**: Optional Tor networking
- **Multi-platform**: Linux, macOS, Windows support
- **GUI Wallet**: Beautiful desktop wallet with full privacy support

## Specifications

- **Block Time**: 20 seconds
- **Gas Limit**: 30M (EIP-1559 style)
- **Total Supply**: 100M OBS
- **Block Reward**: 25 OBS (annual halving every 1,577,000 blocks)
- **Consensus**: DarkMatter PoW (AES/SHA256 hybrid)
- **Privacy**: Shielded addresses (zobs*), encrypted memos
- **Tokens**: ERC-20 style tokens with shielded transfers
- **Burn Redistribution**: 0.1% of burned OBS redistributed to miners

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

Create and manage custom tokens without smart contracts. Full token lifecycle support with minting, burning, and ownership transfers.

**Token Features:**
- **Issuance**: Create tokens with custom parameters
- **Minting**: Additional token creation (if enabled)
- **Burning**: Permanent token destruction
- **Transfers**: Send tokens between addresses
- **Shielding**: Private token transactions
- **Ownership**: Transfer token control

**RPC Methods:**
- `issuetoken` - Create new tokens
- `minttoken` - Mint additional tokens
- `burntoken` - Burn tokens permanently
- `transfertoken` - Transfer tokens
- `transfertokenownership` - Change token owner
- `shieldtoken` - Private token transfers
- `gettokenbalance` - Check balances

See [Token Guide](./docs/token-guide.md) for detailed documentation.

## GUI Wallet

Obsidian includes a beautiful desktop wallet with full privacy support.

### Features

- 🌑 **Dark Modern UI**: Clean, professional interface
- 🔐 **BIP39 Recovery**: 24-word mnemonic phrase backup
- 👁️ **Balance Overview**: Real-time balance across all addresses
- 📤 **Easy Sending**: Send to transparent or shielded addresses
- 📥 **Address Generation**: Create new transparent/shielded addresses
- 🔒 **Auto Shield/Unshield**: Automatic privacy routing
- 📊 **Transaction History**: Track all your transactions
- 💾 **Wallet Backup**: Export and import wallet files

### Quick Start

```bash
# Install dependencies
pip3 install base58 ecdsa requests coincurve mnemonic

# Run GUI wallet
python3 wallet_gui.py
```

### Usage

1. **Create New Wallet**: Generate a new wallet with 24-word recovery phrase
2. **Load Wallet**: Import existing wallet from file or recovery phrase
3. **Receive**: Generate new addresses (transparent or shielded)
4. **Send**: Send OBS to any address (auto shield/unshield)
5. **Backup**: Export wallet file for safekeeping

### Screenshots

The GUI wallet features:
- **Overview Tab**: View total balance and all addresses
- **Send Tab**: Send transactions with memo support
- **Receive Tab**: Generate new addresses
- **History Tab**: View all transactions
- **Settings Tab**: Manage wallet and view recovery phrase

## Getting Started

### Prerequisites
- Go 1.20+
- Git
- Docker (optional)

### Quick Start with Docker

The fastest way to run Obsidian Core:

```bash
# Pull and run the latest version
docker pull yuchanshin/obsidian-node:latest
docker run -d -p 8333:8333 -p 8545:8545 yuchanshin/obsidian-node:latest

# With custom configuration
docker run -d \
  -p 8333:8333 -p 8545:8545 \
  -e SOLO_MINING=true \
  -e MINER_ADDRESS=YourAddress \
  -v obsidian-data:/home/obsidian/data \
  yuchanshin/obsidian-node:latest
```

### Docker Compose Deployments

Multiple deployment configurations available:

```bash
# Standard single node
docker-compose up -d

# Mining pool
docker-compose -f docker-compose.pool.yml up -d

# Seed nodes
docker-compose -f docker-compose.seeds.yml up -d

# Cluster deployment
docker-compose -f docker-compose.cluster.yml up -d

# Tor-enabled node
docker-compose -f docker-compose.tor.yml up -d
```

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

## Environment Variables

Obsidian Core can be configured using environment variables. Below is a complete list of all available variables:

### Network Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `NETWORK` | `mainnet` | Network type (mainnet, testnet, regtest) |
| `P2P_ADDR` | `0.0.0.0:8333` | P2P server listen address |
| `RPC_ADDR` | `0.0.0.0:8545` | RPC server listen address |
| `SEED_NODES` | - | Comma-separated list of seed nodes (e.g., `node1:8333,node2:8333`) |

### Peer Management

| Variable | Default | Description |
|----------|---------|-------------|
| `MAX_PEERS` | `125` | Maximum number of peer connections |
| `MIN_PEERS` | `8` | Minimum number of peer connections to maintain |
| `CONNECT_TIMEOUT` | `30s` | Timeout for establishing peer connections |
| `MESSAGE_TIMEOUT` | `300s` | Timeout for receiving peer messages (5 minutes) |
| `MAX_MESSAGE_SIZE` | `10485760` | Maximum P2P message size in bytes (10MB) |
| `BAN_DURATION` | `24h` | Duration to ban misbehaving peers |

### Mining Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `SOLO_MINING` | `true` | Enable/disable solo mining |
| `MINER_ADDRESS` | `ObsidianDefaultMinerAddress123456789` | Address to receive mining rewards |
| `POOL_SERVER` | `false` | Enable/disable Stratum pool server |
| `POOL_ADDR` | `0.0.0.0:3333` | Pool server listen address |
| `POOL_LISTEN` | `0.0.0.0:3333` | Pool server listen address (alternative) |

### Logging

| Variable | Default | Description |
|----------|---------|-------------|
| `LOG_LEVEL` | `info` | Logging level (debug, info, warn, error) |
| `LOG_FILE` | - | Path to log file (empty = stdout only) |

### Storage

| Variable | Default | Description |
|----------|---------|-------------|
| `DATA_DIR` | `.` | Directory for blockchain data storage |

### Privacy & Tor

| Variable | Default | Description |
|----------|---------|-------------|
| `TOR_ENABLED` | `false` | Enable/disable Tor networking |
| `TOR_PROXY_ADDR` | `127.0.0.1:9050` | Tor SOCKS5 proxy address |

### Go Runtime (Performance Tuning)

| Variable | Default | Description |
|----------|---------|-------------|
| `GOGC` | `100` | Go garbage collection target percentage (lower = more aggressive GC) |
| `GOMAXPROCS` | CPU cores | Maximum number of OS threads for Go runtime |

### Example Configurations

#### Solo Mining Node
```bash
SOLO_MINING=true \
MINER_ADDRESS=YourObsidianAddress \
LOG_LEVEL=info \
./obsidiand
```

#### Mining Pool Server
```bash
POOL_SERVER=true \
POOL_LISTEN=0.0.0.0:3333 \
MINER_ADDRESS=PoolOwnerAddress \
SOLO_MINING=false \
./obsidiand
```

#### Tor-Enabled Privacy Node
```bash
TOR_ENABLED=true \
TOR_PROXY_ADDR=127.0.0.1:9050 \
LOG_LEVEL=debug \
./obsidiand
```

#### Seed Node
```bash
P2P_ADDR=0.0.0.0:8333 \
MAX_PEERS=500 \
SOLO_MINING=false \
DATA_DIR=/var/lib/obsidian \
LOG_FILE=/var/log/obsidian/node.log \
./obsidiand
```

#### Full Node with Custom Configuration
```bash
NETWORK=mainnet \
P2P_ADDR=0.0.0.0:8333 \
RPC_ADDR=127.0.0.1:8545 \
SEED_NODES=seed1.obsidian.network:8333,seed2.obsidian.network:8333 \
MAX_PEERS=125 \
MIN_PEERS=8 \
SOLO_MINING=true \
MINER_ADDRESS=obs5pPyd6DA6tYyYwip4hYcBWWFTNf4wj8nn \
DATA_DIR=/home/obsidian/data \
LOG_LEVEL=info \
LOG_FILE=/home/obsidian/logs/obsidian.log \
./obsidiand
```

#### Docker Environment File (.env)
```bash
# Network
NETWORK=mainnet
P2P_ADDR=0.0.0.0:8333
RPC_ADDR=0.0.0.0:8545

# Peers
MAX_PEERS=125
MIN_PEERS=8
SEED_NODES=seed1:8333,seed2:8333

# Mining
SOLO_MINING=true
MINER_ADDRESS=your_address_here

# Logging
LOG_LEVEL=info
LOG_FILE=/home/obsidian/logs/obsidian.log

# Storage
DATA_DIR=/home/obsidian/data

# Privacy
TOR_ENABLED=false

# Performance
GOGC=50
GOMAXPROCS=4
```

## Documentation

- [Protocol Specification](./docs/protocol-spec.md) - Wire protocol and consensus rules
- [Build Guide](./docs/build-guide.md) - Platform-specific build instructions
- [Token Guide](./docs/token-guide.md) - Custom token system documentation
- [Go Implementation](./docs/go-implementation.md) - Reference implementation details
- [Rust Implementation](./docs/rust-implementation.md) - High-performance alternative
- [Python Implementation](./docs/python-implementation.md) - Research and experimentation
- [JavaScript Implementation](./docs/javascript-implementation.md) - Web integration guide

## Docker Hub

Official Docker images are available at:
```bash
# Latest version
docker pull yuchanshin/obsidian-node:latest

# Specific versions
docker pull yuchanshin/obsidian-node:v1.2.3
docker pull yuchanshin/obsidian-node:v1.2.2
docker pull yuchanshin/obsidian-node:v1.2.1
```

**Supported architectures:**
- linux/amd64 (x86_64)
- linux/arm64 (ARM 64-bit)

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is open source and available under the MIT License.

## Community

- **GitHub**: https://github.com/your-org/obsidian-core
- **Issues**: https://github.com/your-org/obsidian-core/issues

## Production Deployment

### System Requirements

**Minimum:**
- CPU: 2 cores
- RAM: 2GB
- Storage: 50GB SSD
- Network: 1 Mbps

**Recommended:**
- CPU: 4 cores
- RAM: 8GB
- Storage: 200GB SSD
- Network: 10 Mbps

### Configuration

See the [Environment Variables](#environment-variables) section above for a complete list of all configuration options.

Quick configuration example:

```bash
# Create .env file
cat > .env << EOF
NETWORK=mainnet
P2P_ADDR=0.0.0.0:8333
RPC_ADDR=0.0.0.0:8545
LOG_LEVEL=info
LOG_FILE=/var/log/obsidian/obsidian.log
SOLO_MINING=true
MINER_ADDRESS=your_address_here
MAX_PEERS=125
BAN_DURATION=24h
DATA_DIR=/var/lib/obsidian
EOF
```

### Security Best Practices

1. **Run as non-root user**: The Docker image uses a dedicated user
2. **Use firewall rules**: Only expose necessary ports
3. **Enable rate limiting**: RPC requests are rate-limited by default
4. **Regular backups**: Backup your data directory regularly
5. **Monitor logs**: Check logs for suspicious activity
6. **Use HTTPS**: Put RPC behind a reverse proxy with TLS
7. **Strong authentication**: Implement authentication for RPC access

### Monitoring

Check node health:
```bash
# Health check endpoint
curl http://localhost:8545/health

# Metrics endpoint
curl http://localhost:8545/metrics

# RPC methods
curl -X POST -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"getblockchaininfo","params":[],"id":1}' \
  http://localhost:8545
```

### Backup and Recovery

```bash
# Backup blockchain data
docker exec obsidian-node tar czf /backup/obsidian-data.tar.gz /home/obsidian/data

# Restore from backup
docker cp obsidian-data.tar.gz obsidian-node:/tmp/
docker exec obsidian-node tar xzf /tmp/obsidian-data.tar.gz -C /home/obsidian/
```

## Performance Tuning

### Go Runtime

Set Go environment variables for optimal performance:

```bash
GOGC=50              # Aggressive garbage collection
GOMAXPROCS=4         # Number of CPU cores to use
```

### Database

The node uses BoltDB for blockchain storage. For best performance:
- Use SSD storage
- Ensure sufficient disk space
- Regular database compaction

## Troubleshooting

### Common Issues

**Node won't start:**
- Check logs: `docker logs obsidian-node`
- Verify ports are not in use: `lsof -i :8333`
- Check disk space: `df -h`

**No peer connections:**
- Check firewall rules
- Verify seed nodes are reachable
- Enable Tor if behind NAT

**High memory usage:**
- Reduce `MAX_PEERS`
- Adjust `GOGC` value
- Check for memory leaks in logs

## Recent Updates

### v1.2.3 (Latest)
- **Logging**: Replaced emojis with text labels for better terminal compatibility
- **Format**: `[MINING]`, `[OK]`, `[SUCCESS]`, `[ERROR]`, `[BROADCAST]`, `[LAUNCH]`, `[PEER]`
- **Compatibility**: Works on all terminals and logging systems
- **Multi-arch**: linux/amd64 and linux/arm64 support

### v1.2.2
- **Logging**: Consolidated startup logs into single comprehensive line
- **Monitoring**: All critical info in one line (Network, PoW, Tor, P2P, Mining, Height)
- **UX**: Dramatically reduced log spam when running multiple nodes
- **Documentation**: Complete environment variables reference with examples

### v1.2.1
- **Environment Variables**: Added comprehensive documentation for all configuration options
- **Logging**: Improved structured logging with reduced verbosity

### v1.2.0 (Production Release)
- **Security**: Rate limiting, request validation, non-root Docker user
- **Logging**: Structured logging with file output support
- **Monitoring**: Health check and metrics endpoints
- **Docker**: Optimized multi-stage build with security hardening
- **Configuration**: Comprehensive environment variable support
- **Error Handling**: Improved error handling and graceful shutdown
- **Resource Management**: Memory and CPU limits in Docker Compose
- **Testing**: All tests passing with increased coverage

### v1.1.0
- Enhanced validation and token mint processing
- Complete token management features (minting, burning, ownership transfer)
- Improved security and vulnerability fixes
- P2P networking tests and improvements
- Updated dependencies
