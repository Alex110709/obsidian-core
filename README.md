# Obsidian Core

Obsidian is a privacy-focused cryptocurrency based on the Zcash protocol with **shielded transactions** and **encrypted memos**.

## Specifications
- **Block Size**: 6MB
- **Block Time**: 5 minutes (target)
- **Difficulty Adjustment**: Every 2,016 blocks (~1 week), max 4x change (Bitcoin algorithm)
- **Total Supply**: 100,000,000 OBS
- **Initial Supply**: 1,000,000 OBS (pre-mine)
- **Initial Block Reward**: 100 OBS
- **Halving Interval**: 420,000 blocks (~4 years)
- **Consensus**: Proof of Work (DarkMatter - AES/SHA256 Hybrid)
- **Storage**: Persistent storage using BoltDB
- **Privacy**: Shielded transactions (zk-SNARK-inspired), encrypted memos
- **Security**: Block validation, PoW verification, and size limits

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

## Emission Schedule

### Supply Distribution
- **Total Supply**: 100,000,000 OBS
- **Pre-mine**: 1,000,000 OBS (1%)
- **Minable**: 99,000,000 OBS (99%)

### Block Rewards
- **Initial**: 100 OBS per block
- **Halving**: Every 420,000 blocks (~4 years)
- **Timeline**: 
  - Year 0-4: 100 OBS/block → 42M OBS
  - Year 4-8: 50 OBS/block → 21M OBS
  - Year 8-12: 25 OBS/block → 10.5M OBS
  - Year 12+: Decreasing halvings...

See [EMISSION.md](EMISSION.md) for full emission schedule and economic model.

## Getting Started

### Prerequisites
- Go 1.20+

### Running the Node
```bash
go run cmd/obsidiand/main.go
```

This will start the node, initialize the blockchain with the Genesis block, and start a CPU miner simulation.

## Project Structure
- `cmd/obsidiand`: Main daemon entry point.
- `chaincfg`: Chain configuration and parameters.
- `wire`: Wire protocol data structures (Blocks, Transactions).
- `blockchain`: Blockchain state management with persistent storage.
- `consensus`: Consensus rules and DarkMatter PoW implementation.
- `mining`: CPU miner implementation.
- `stratum`: Stratum mining pool server implementation.
- `database`: BoltDB storage layer.
- `tor`: Tor integration for anonymous networking.
- `network`: P2P network and peer management.

## Production Features

### Security Enhancements
1. **Block Validation**: Validates block headers, PoW, and transactions before accepting.
2. **Persistent Storage**: Uses BoltDB for production-grade data persistence.
3. **PoW Verification**: Real DarkMatter algorithm implementation with difficulty checking.
4. **Size Limits**: Enforces 6MB block size and transaction limits.
5. **Duplicate Prevention**: Prevents duplicate blocks from being accepted.
6. **Tor Integration**: Optional Tor support for anonymous P2P networking and .onion peers.

### DarkMatter PoW Algorithm
- **Hybrid Approach**: Combines SHA-256 for hashing with AES encryption for memory-hardness.
- **Security**: Resistant to ASIC mining while maintaining verification efficiency.
- **Steps**:
  1. SHA-256 hash of block header
  2. AES-CTR encryption using hash as key
  3. Final SHA-256 hash of ciphertext
  4. Compare result against difficulty target

### Difficulty Adjustment (Bitcoin Algorithm)
Obsidian uses **Bitcoin's exact difficulty adjustment algorithm**, adapted for 5-minute blocks:

- **Target Block Time**: 5 minutes
- **Retarget Interval**: 2,016 blocks (exactly 1 week at target rate)
- **Algorithm**: Same as Bitcoin
  1. Measures actual time to mine last 2,016 blocks
  2. Calculates: `new_target = old_target × (actual_time / target_time)`
  3. Clamps adjustment to prevent extreme changes:
     - Minimum: 1/4 of target time (if blocks 4x faster)
     - Maximum: 4x target time (if blocks 4x slower)
  4. Never allows difficulty below PowLimit (minimum difficulty)

**Example:**
```
If 2,016 blocks mined in 3.5 days instead of 7 days:
  → Blocks found 2x faster than target
  → Difficulty doubles (target halves)
  → Next 2,016 blocks will be harder to mine
```

This is identical to Bitcoin's mechanism, but optimized for 5-minute blocks instead of 10-minute blocks.

## Mining Modes

Obsidian supports **two mining modes** that can be toggled with environment variables:

### 1. Solo Mining (Default)
Mine blocks directly to your own address:
```bash
SOLO_MINING=true
MINER_ADDRESS=YourObsidianAddress
```

### 2. Mining Pool Server
Run a Stratum mining pool that miners can connect to:
```bash
POOL_SERVER=true
POOL_LISTEN=0.0.0.0:3333
MINER_ADDRESS=PoolRewardAddress
```

**Pool Protocol**: `stratum+tcp://your-server:3333`

**Supported Methods:**
- `mining.subscribe` - Subscribe to pool
- `mining.authorize` - Authorize with username/password
- `mining.submit` - Submit shares
- `mining.notify` - Receive new work (server → client)
- `mining.set_difficulty` - Set difficulty (server → client)

**RPC Methods:**
```bash
# Get pool statistics
curl -X POST http://localhost:8545 -d '{
  "jsonrpc":"2.0",
  "method":"getpoolinfo",
  "params":[],
  "id":1
}'
```

### Running Both Modes
You can run solo mining and pool server simultaneously:
```bash
SOLO_MINING=true
POOL_SERVER=true
```

Or disable both for a non-mining full node:
```bash
SOLO_MINING=false
POOL_SERVER=false
```

## Docker Deployment

### Environment Variables

Obsidian Core supports the following environment variables for configuration:

#### Mining Configuration
- `SOLO_MINING`: Enable/disable solo CPU mining (default: `"true"`)
- `POOL_SERVER`: Enable/disable Stratum mining pool server (default: `"false"`)
- `MINER_ADDRESS`: Mining reward address (default: `"ObsidianDefaultMinerAddress123456789"`)
- `POOL_LISTEN`: Stratum pool server listen address (default: `"0.0.0.0:3333"`)

#### Network Configuration
- `TOR_ENABLED`: Enable Tor for anonymous P2P networking (default: `"false"`)
- `SEED_NODES`: Comma-separated list of seed nodes (optional)
- `DATA_DIR`: Custom data directory path (default: `"/root"`)

#### RPC Configuration
- `RPC_ADDR`: RPC server listen address (default: `"0.0.0.0:8545"`)

### Docker Compose Configurations

#### Quick Start with Docker Compose (Recommended)

The easiest way to run Obsidian node:

```bash
docker compose up -d
```

Stop the node:
```bash
docker compose down
```

View logs:
```bash
docker compose logs -f
```

This will:
- Pull the latest multi-architecture image (supports AMD64 and ARM64)
- Create a persistent volume for blockchain data
- Expose ports: 8333 (P2P), 8545 (RPC), 3333 (Stratum), 9050 (Tor)
- Enable solo mining by default
- Auto-restart on failure

#### Running a Mining Pool

To run a Stratum mining pool server:

```bash
docker compose -f docker-compose.pool.yml up -d
```

This will:
- Disable solo mining
- Enable Stratum pool server on port 3333
- Miners can connect to: `stratum+tcp://your-server-ip:3333`

**Connect miners to your pool:**
```bash
# Example with cpuminer
minerd -a sha256d -o stratum+tcp://localhost:3333 -u miner1 -p password
```

**Monitor pool statistics:**
```bash
curl -X POST http://localhost:8545 -d '{"jsonrpc":"2.0","method":"getpoolinfo","params":[],"id":1}'
```

#### Running with Tor (Anonymous Networking)

For enhanced privacy with anonymous P2P connections:

```bash
docker compose -f docker-compose.tor.yml up -d
```

Or modify the default `docker-compose.yml`:
```yaml
environment:
  - TOR_ENABLED=true
```

This enables:
- Tor SOCKS5 proxy on port 9050
- Anonymous P2P connections through Tor network
- Support for .onion peer addresses

#### Cluster Setup (Multiple Nodes)

For running multiple interconnected nodes:

```bash
docker compose -f docker-compose.cluster.yml up -d
```

This creates a 3-node cluster with:
- Automatic peer discovery
- Load balancing across nodes
- Shared mining rewards

#### Synology NAS Deployment

For Synology NAS users, multiple deployment options are available:

```bash
# Simple setup
docker compose -f docker-compose.synology-simple.yml up -d

# Full-featured setup
docker compose -f docker-compose.synology-final.yml up -d
```

#### Seed Node Setup

To run dedicated seed nodes for network bootstrapping:

```bash
docker compose -f docker-compose.seeds.yml up -d
```

### Custom Configuration

Create a custom `docker-compose.override.yml`:

```yaml
services:
  obsidian-node:
    environment:
      - SOLO_MINING=true
      - MINER_ADDRESS=YourCustomAddress
      - TOR_ENABLED=true
      - SEED_NODES=node1.example.com:8333,node2.example.com:8333
    volumes:
      - ./custom-data:/root
```

Then run:
```bash
docker compose -f docker-compose.yml -f docker-compose.override.yml up -d
```

### Port Reference

| Port | Protocol | Description |
|------|----------|-------------|
| 8333 | TCP | P2P network connections |
| 8545 | TCP | JSON-RPC API |
| 3333 | TCP | Stratum mining pool |
| 9050 | TCP | Tor SOCKS5 proxy (when enabled) |

### Volume Management

Blockchain data is stored in named Docker volumes:
- `obsidian-data`: Main node data
- `obsidian-pool-data`: Pool server data

To backup data:
```bash
docker run --rm -v obsidian-data:/data -v $(pwd):/backup alpine tar czf /backup/obsidian-backup.tar.gz -C /data .
```

To restore:
```bash
docker run --rm -v obsidian-data:/data -v $(pwd):/backup alpine tar xzf /backup/obsidian-backup.tar.gz -C /data
```

#### Running with Tor (Anonymous Networking)

Tor is now integrated directly into the Obsidian node. To enable it, uncomment the `TOR_ENABLED=true` line in `docker-compose.yml`:

```yaml
environment:
  - TOR_ENABLED=true  # Uncomment this line
```

Then start the node:

```bash
docker compose up -d
```

This will:
- Automatically start Tor process inside the container
- Route all P2P connections through Tor SOCKS proxy
- Support .onion peer addresses for fully anonymous networking
- Expose Tor SOCKS proxy on port 9050

Check node logs (including Tor startup):
```bash
docker compose logs -f obsidian-node
```

### Using Pre-built Docker Image

Pull and run the official image:
```bash
docker pull yuchanshin/obsidian-node:latest
docker run -d -p 8333:8333 -v obsidian-data:/root --name obsidian-node yuchanshin/obsidian-node:latest
```

### Building Your Own Image

Build the Docker image:
```bash
./build_and_push.sh YOUR_DOCKER_USERNAME
```

Or manually:
```bash
docker build -t YOUR_USERNAME/obsidian-node:latest .
docker push YOUR_USERNAME/obsidian-node:latest
```

Multi-architecture build (AMD64 + ARM64):
```bash
docker buildx build --platform linux/amd64,linux/arm64 -t YOUR_USERNAME/obsidian-node:latest --push .
```

## Tor Support

Obsidian Core includes built-in Tor support for enhanced privacy and anonymous P2P networking.

### Features
- **Anonymous Connections**: All P2P traffic routed through Tor network
- **.onion Peer Support**: Connect to hidden service peers
- **Optional**: Tor can be enabled/disabled via environment variable
- **SOCKS5 Proxy**: Uses standard Tor SOCKS5 proxy (default: 127.0.0.1:9050)

### Running with Tor

Tor is now fully integrated into Obsidian Core. When enabled, it automatically starts and manages the Tor process.

**Using Docker Compose** (Recommended):
Edit `docker-compose.yml` and uncomment `TOR_ENABLED=true`, then:
```bash
docker compose up -d
```

**Using Environment Variable**:
```bash
TOR_ENABLED=true go run cmd/obsidiand/main.go
```

**Features:**
- Automatic Tor process management (no separate installation needed)
- Creates and manages Tor configuration automatically
- Graceful shutdown of Tor process on exit
- For .onion addresses, Tor is used automatically regardless of the setting

## JSON-RPC API

Obsidian Core provides a JSON-RPC 2.0 API for interacting with the blockchain.

### Connection

**Default RPC Endpoint**: `http://localhost:8545`

The RPC server starts automatically with the node and listens on port 8545 by default. You can customize this using the `RPC_ADDR` environment variable.

### API Methods

#### `ping`
Test server connectivity.

**Request:**
```bash
curl -X POST http://localhost:8545 \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"ping","params":[],"id":1}'
```

**Response:**
```json
{"jsonrpc":"2.0","result":"pong","id":1}
```

#### `getblockcount`
Returns the current blockchain height.

**Request:**
```bash
curl -X POST http://localhost:8545 \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"getblockcount","params":[],"id":1}'
```

**Response:**
```json
{"jsonrpc":"2.0","result":42,"id":1}
```

#### `getbestblockhash`
Returns the hash of the best (tip) block.

**Request:**
```bash
curl -X POST http://localhost:8545 \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"getbestblockhash","params":[],"id":1}'
```

**Response:**
```json
{"jsonrpc":"2.0","result":"000000abcd1234...","id":1}
```

#### `getblock`
Returns information about a block by hash.

**Parameters:**
- `blockhash` (string): The block hash

**Request:**
```bash
curl -X POST http://localhost:8545 \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"getblock","params":["000000abcd1234..."],"id":1}'
```

**Response:**
```json
{
  "jsonrpc":"2.0",
  "result":{
    "hash":"000000abcd1234...",
    "height":42,
    "version":1,
    "previousblockhash":"000000xyz...",
    "merkleroot":"abcd1234...",
    "time":1700000000,
    "bits":486604799,
    "nonce":12345,
    "tx_count":1
  },
  "id":1
}
```

#### `getblockchaininfo`
Returns general blockchain information.

**Request:**
```bash
curl -X POST http://localhost:8545 \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"getblockchaininfo","params":[],"id":1}'
```

**Response:**
```json
{
  "jsonrpc":"2.0",
  "result":{
    "chain":"mainnet",
    "blocks":42,
    "bestblockhash":"000000abcd1234...",
    "difficulty":486604799,
    "maxmoney":100000000,
    "initialsupply":1000000
  },
  "id":1
}
```

#### `getmininginfo`
Returns mining-related information.

**Request:**
```bash
curl -X POST http://localhost:8545 \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"getmininginfo","params":[],"id":1}'
```

**Response:**
```json
{
  "jsonrpc":"2.0",
  "result":{
    "blocks":42,
    "currentblockhash":"000000abcd1234...",
    "difficulty":486604799,
    "generate":true,
    "hashespersec":0
  },
  "id":1
}
```

### Error Handling

If an error occurs, the RPC server returns a JSON-RPC error response:

```json
{
  "jsonrpc":"2.0",
  "error":{
    "code":-32603,
    "message":"Internal error",
    "data":null
  },
  "id":1
}
```

### Using with JavaScript

```javascript
async function getBlockCount() {
  const response = await fetch('http://localhost:8545', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      jsonrpc: '2.0',
      method: 'getblockcount',
      params: [],
      id: 1
    })
  });
  const data = await response.json();
  console.log('Block count:', data.result);
}
```

### Using with Python

```python
import requests

def get_blockchain_info():
    payload = {
        "jsonrpc": "2.0",
        "method": "getblockchaininfo",
        "params": [],
        "id": 1
    }
    response = requests.post('http://localhost:8545', json=payload)
    return response.json()['result']
```

## Shielded Transaction API (v1.1.0+)

Obsidian supports Zcash-compatible shielded transaction methods for privacy-preserving payments.

### `z_getnewaddress`
Generate a new shielded z-address.

**Request:**
```bash
curl -X POST http://localhost:8545 \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"z_getnewaddress","params":[],"id":1}'
```

**Response:**
```json
{"jsonrpc":"2.0","result":"zobs1abc123...","id":1}
```

### `z_listaddresses`
List all z-addresses in wallet.

**Request:**
```bash
curl -X POST http://localhost:8545 \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"z_listaddresses","params":[],"id":1}'
```

### `z_getbalance`
Get shielded balance for a z-address.

**Parameters:**
- `address` (string): The z-address

**Request:**
```bash
curl -X POST http://localhost:8545 \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"z_getbalance","params":["zobs1abc123..."],"id":1}'
```

**Response:**
```json
{
  "jsonrpc":"2.0",
  "result":{
    "address":"zobs1abc123...",
    "balance":1000000000,
    "balance_obs":10.0
  },
  "id":1
}
```

### `z_sendmany`
Send from z-address to multiple recipients (transparent or shielded).

**Parameters:**
- `from_address` (string): Source z-address
- `amounts` (array): Array of {address, amount, memo} objects
- `memo` (string, optional): Default memo for all recipients

**Request:**
```bash
curl -X POST http://localhost:8545 \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc":"2.0",
    "method":"z_sendmany",
    "params":[
      "zobs1sender...",
      [
        {"address":"zobs1recipient1...","amount":1.5,"memo":"Payment 1"},
        {"address":"Obs_recipient2","amount":2.0,"memo":"Payment 2"}
      ]
    ],
    "id":1
  }'
```

**Response:**
```json
{
  "jsonrpc":"2.0",
  "result":{
    "txid":"abc123...",
    "from":"zobs1sender...",
    "recipients":2
  },
  "id":1
}
```

### `z_gettotalbalance`
Get total balance (transparent + shielded).

**Request:**
```bash
curl -X POST http://localhost:8545 \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"z_gettotalbalance","params":[],"id":1}'
```

**Response:**
```json
{
  "jsonrpc":"2.0",
  "result":{
    "transparent":5.0,
    "shielded":10.0,
    "total":15.0
  },
  "id":1
}
```

### `z_exportviewingkey`
Export viewing key for a z-address (allows read-only access).

**Parameters:**
- `address` (string): The z-address

**Request:**
```bash
curl -X POST http://localhost:8545 \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"z_exportviewingkey","params":["zobs1abc..."],"id":1}'
```

### `z_importviewingkey`
Import a viewing key for read-only access.

**Parameters:**
- `viewingkey` (string): The viewing key to import

**Request:**
```bash
curl -X POST http://localhost:8545 \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"z_importviewingkey","params":["vk_abc123..."],"id":1}'
```

### `z_shieldcoinbase`
Shield transparent coinbase funds to a z-address.

**Parameters:**
- `to_address` (string): Destination z-address

**Request:**
```bash
curl -X POST http://localhost:8545 \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"z_shieldcoinbase","params":["zobs1abc..."],"id":1}'
```

### Example: Private Payment with Memo

```bash
# 1. Generate new z-address
curl -X POST http://localhost:8545 \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"z_getnewaddress","params":[],"id":1}'

# 2. Send private payment with memo
curl -X POST http://localhost:8545 \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc":"2.0",
    "method":"z_sendmany",
    "params":[
      "zobs1sender...",
      [{"address":"zobs1recipient...","amount":10.5,"memo":"Invoice #1234"}]
    ],
    "id":1
  }'

# 3. Check balance
curl -X POST http://localhost:8545 \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"z_getbalance","params":["zobs1recipient..."],"id":1}'
```

## Economic Model

### Block Rewards & Fees

- **Initial Reward**: 50 OBS per block
- **Halving Schedule**: Every 420,000 blocks (~4 years at 5-minute intervals)
- **Minimum Fee**: 0.0001 OBS (10,000 satoshis)
- **Fee Calculation**: 10 satoshis per byte + base fee

**Example Fee Calculation:**
```bash
curl -X POST http://localhost:8545 \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"estimatefee","params":[250],"id":1}'
```

Response:
```json
{
  "size_bytes": 250,
  "fee_satoshis": 12500,
  "fee_obs": 0.000125,
  "fee_per_byte": 10
}
```

### Supply Schedule

| Halving | Block Range | Reward | Supply Added |
|---------|-------------|--------|--------------|
| 0 (Initial) | 0 - 420,000 | 50 OBS | 21,000,000 OBS |
| 1 | 420,001 - 840,000 | 25 OBS | 10,500,000 OBS |
| 2 | 840,001 - 1,260,000 | 12.5 OBS | 5,250,000 OBS |
| 3+ | ... | ... | ... |

**Total Cap**: 100,000,000 OBS (including 1M initial supply)

## Version History

### v1.1.0 (Latest)
- ✅ Shielded transactions (z-addresses)
- ✅ Encrypted memos (up to 512 bytes)
- ✅ zk-SNARK-inspired privacy proofs
- ✅ Viewing keys for read-only access
- ✅ Shielded pool management
- ✅ Z-address RPC methods (z_sendmany, z_getbalance, etc.)

### v1.0.2
- ✅ Network fee system
- ✅ Transaction fee calculation
- ✅ Enhanced RPC methods

### v1.0.1
- ✅ Block reward system
- ✅ Halving mechanism
- ✅ JSON-RPC 2.0 API

### v1.0.0
- ✅ Initial release
- ✅ DarkMatter PoW
- ✅ Tor integration
- ✅ Persistent storage
