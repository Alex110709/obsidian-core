# Mining Guide

## Overview

Obsidian supports two mining modes:
1. **Solo Mining** - Mine blocks directly to your own address
2. **Pool Mining** - Run a Stratum pool that other miners connect to

Both modes can be toggled with environment variables and run simultaneously.

## Solo Mining

### Configuration

```bash
# Enable solo mining (default: true)
SOLO_MINING=true
MINER_ADDRESS=YourObsidianAddress
```

### Running Solo Miner

```bash
# With Docker
docker compose up -d

# With Go
MINER_ADDRESS=YourObsAddress go run cmd/obsidiand/main.go
```

### Expected Output

```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Mining Configuration
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  Miner Address:  YourObsAddress
  Block Reward:   100 OBS (halves every 420000 blocks)
  Solo Mining:    true
  Pool Server:    false
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
✓ Solo mining started

Miner started. Mining on CPU...
✓ Mined block! Height: 1, Subsidy: 100 OBS, Fees: 0, Total: 100 OBS
✓ Mined block! Height: 2, Subsidy: 100 OBS, Fees: 0, Total: 100 OBS
```

## Pool Mining

### Configuration

```bash
# Enable pool server
POOL_SERVER=true
POOL_LISTEN=0.0.0.0:3333
MINER_ADDRESS=PoolRewardAddress

# Optionally disable solo mining on pool server
SOLO_MINING=false
```

### Running Pool Server

```bash
# With Docker
docker compose -f docker-compose.pool.yml up -d

# With Go
POOL_SERVER=true POOL_LISTEN=0.0.0.0:3333 go run cmd/obsidiand/main.go
```

### Expected Output

```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Mining Configuration
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  Miner Address:  PoolRewardAddress
  Block Reward:   100 OBS (halves every 420000 blocks)
  Solo Mining:    false
  Pool Server:    true
  Pool Listen:    stratum+tcp://0.0.0.0:3333
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
✗ Solo mining disabled
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
⛏️  Stratum Mining Pool Started
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  Listen Address: stratum+tcp://0.0.0.0:3333
  Pool Address:   PoolRewardAddress
  Difficulty:     1.00
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
✓ Pool server started
```

### Connecting Miners to Your Pool

Miners connect using the Stratum protocol: `stratum+tcp://your-server:3333`

**Example with cpuminer:**
```bash
minerd -a sha256d \
  -o stratum+tcp://localhost:3333 \
  -u miner1 \
  -p password
```

**Example with cgminer:**
```bash
cgminer -o stratum+tcp://localhost:3333 -u miner1 -p x
```

### Pool Events

When miners connect and submit shares:

```
⛏️  New miner connected: 192.168.1.100:52341
⛏️  Miner authorized: 192.168.1.100:52341 (user: miner1)
⛏️  New job generated: 0000000000000001 (height: 5, miners: 1)
⛏️  Share submitted by 192.168.1.100:52341 (total: 1)
⛏️  Share submitted by 192.168.1.100:52341 (total: 2)
⛏️  Miner disconnected: 192.168.1.100:52341 (shares: 25)
```

## Stratum Protocol

### Supported Methods

#### Client → Server

**mining.subscribe**
```json
{
  "id": 1,
  "method": "mining.subscribe",
  "params": ["user-agent/1.0"]
}
```

Response:
```json
{
  "id": 1,
  "result": [
    [["mining.notify", "ae6812eb4cd7735a302a8a9dd95c6578"]],
    "08000002",
    4
  ],
  "error": null
}
```

**mining.authorize**
```json
{
  "id": 2,
  "method": "mining.authorize",
  "params": ["username", "password"]
}
```

Response:
```json
{
  "id": 2,
  "result": true,
  "error": null
}
```

**mining.submit**
```json
{
  "id": 3,
  "method": "mining.submit",
  "params": [
    "username",
    "job_id",
    "extranonce2",
    "ntime",
    "nonce"
  ]
}
```

Response:
```json
{
  "id": 3,
  "result": true,
  "error": null
}
```

#### Server → Client

**mining.set_difficulty**
```json
{
  "id": null,
  "method": "mining.set_difficulty",
  "params": [1.0]
}
```

**mining.notify** (new work)
```json
{
  "id": null,
  "method": "mining.notify",
  "params": [
    "job_id",
    "prevhash",
    "coinbase1",
    "coinbase2",
    ["merkle_branch"],
    "version",
    "nbits",
    "ntime",
    true
  ]
}
```

## Pool Statistics

### RPC Method: `getpoolinfo`

Get real-time pool statistics:

```bash
curl -X POST http://localhost:8545 \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "getpoolinfo",
    "params": [],
    "id": 1
  }'
```

**Response (pool enabled):**
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "enabled": true,
    "miners": 3,
    "total_shares": 1523,
    "difficulty": 1.0,
    "pool_address": "PoolRewardAddress"
  }
}
```

**Response (pool disabled):**
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "enabled": false,
    "message": "Pool server not running"
  }
}
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `SOLO_MINING` | `true` | Enable/disable solo mining |
| `POOL_SERVER` | `false` | Enable/disable pool server |
| `POOL_LISTEN` | `0.0.0.0:3333` | Stratum server listen address |
| `MINER_ADDRESS` | Default address | Address to receive mining rewards |

## Running Modes

### Mode 1: Solo Miner Only (Default)
```bash
SOLO_MINING=true
POOL_SERVER=false
```

### Mode 2: Pool Server Only
```bash
SOLO_MINING=false
POOL_SERVER=true
```

### Mode 3: Both Solo + Pool
```bash
SOLO_MINING=true
POOL_SERVER=true
```

### Mode 4: Non-Mining Full Node
```bash
SOLO_MINING=false
POOL_SERVER=false
```

## Docker Examples

### Solo Mining Node
```bash
docker compose up -d
```

### Mining Pool Server
```bash
docker compose -f docker-compose.pool.yml up -d
```

### Hybrid (Solo + Pool)
Edit `docker-compose.yml`:
```yaml
environment:
  SOLO_MINING: "true"
  POOL_SERVER: "true"
  MINER_ADDRESS: "YourAddress"
```

Then run:
```bash
docker compose up -d
```

## Port Reference

| Port | Protocol | Purpose |
|------|----------|---------|
| 8333 | TCP | P2P network |
| 8545 | HTTP | JSON-RPC API |
| 3333 | Stratum | Mining pool |
| 9050 | SOCKS5 | Tor proxy (if enabled) |

## Troubleshooting

### Pool server not starting
- Check if port 3333 is available: `netstat -an | grep 3333`
- Verify `POOL_SERVER=true` is set
- Check logs: `docker logs obsidian-pool`

### Miners can't connect
- Ensure port 3333 is open in firewall
- Check pool is listening: `curl localhost:8545 -d '{"method":"getpoolinfo"}'`
- Verify correct pool URL: `stratum+tcp://your-ip:3333`

### No shares being submitted
- Check miner logs for errors
- Verify miner algorithm matches (DarkMatter uses custom PoW)
- Ensure miner is authorized: check "authorized: true" in pool logs

## Future Enhancements

Planned features for mining:
- [ ] Share validation and difficulty adjustment per miner
- [ ] Payout system based on shares
- [ ] Pool-side block submission
- [ ] Worker statistics and graphs
- [ ] PPLNS/PPS reward systems
- [ ] Pool fee configuration
- [ ] Vardiff (variable difficulty) support
