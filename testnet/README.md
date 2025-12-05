# Obsidian Testnet Guide

## Overview
This testnet setup allows you to run multiple Obsidian nodes locally for testing P2P networking, mining, and transactions.

## Architecture
```
Node 1 (Seed)     Node 2           Node 3
Port: 8333        Port: 8334       Port: 8335
RPC: 8545         RPC: 8546        RPC: 8547
    |                |                |
    +----------------+----------------+
           P2P Network
```

## Quick Start

### 1. Start All Nodes
```bash
# Terminal 1 - Start Node 1 (Seed Node)
cd testnet
./start_node1.sh

# Terminal 2 - Start Node 2
./start_node2.sh

# Terminal 3 - Start Node 3
./start_node3.sh
```

### 2. Run Network Tests
```bash
# Terminal 4 - Run test suite
./test_network.sh
```

## Manual Testing

### Check Block Count
```bash
# Node 1
curl -s -X POST http://localhost:8545 \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"getblockcount","params":[],"id":1}' | jq .

# Node 2
curl -s -X POST http://localhost:8546 \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"getblockcount","params":[],"id":1}' | jq .
```

### Generate Address
```bash
curl -s -X POST http://localhost:8545 \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"getnewaddress","params":[],"id":1}' | jq .
```

### Send Transaction
```bash
# Send from Node 1 address to Node 2 address
curl -s -X POST http://localhost:8545 \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc":"2.0",
    "method":"sendtoaddress",
    "params":["<from_addr>","<to_addr>",10.5],
    "id":1
  }' | jq .
```

### Check Balance
```bash
curl -s -X POST http://localhost:8545 \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc":"2.0",
    "method":"getbalance",
    "params":["<address>"],
    "id":1
  }' | jq .
```

### Burn OBS
```bash
curl -s -X POST http://localhost:8545 \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc":"2.0",
    "method":"burnobs",
    "params":["<from_addr>",100],
    "id":1
  }' | jq .
```

### Check Circulating Supply
```bash
curl -s -X POST http://localhost:8545 \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"getcirculatingsupply","params":[],"id":1}' | jq .
```

## Test Scenarios

### Scenario 1: P2P Synchronization
1. Start all 3 nodes
2. Wait for nodes to discover each other
3. Check that block heights are synchronized
4. Verify blocks propagate across network

### Scenario 2: Mining Competition
1. All nodes mine simultaneously
2. Verify that blocks from different miners are accepted
3. Check for chain reorganizations
4. Verify consensus is maintained

### Scenario 3: Transaction Propagation
1. Generate addresses on Node 1 and Node 2
2. Mine blocks to get balance on Node 1
3. Send transaction from Node 1 to Node 2
4. Verify transaction appears in Node 2's mempool
5. Mine block and verify transaction is confirmed
6. Check balance on Node 2

### Scenario 4: Shielded Transactions
1. Generate shielded addresses on both nodes
2. Shield funds (transparent → shielded)
3. Send shielded transaction
4. Unshield funds (shielded → transparent)
5. Verify privacy features work correctly

### Scenario 5: Burn and Redistribution
1. Burn OBS tokens
2. Continue mining
3. Verify burned coins are redistributed in block rewards
4. Check circulating supply decreases

## Logs

Each node stores logs in its own directory:
- Node 1: `./testnet/node1/node1.log`
- Node 2: `./testnet/node2/node2.log`
- Node 3: `./testnet/node3/node3.log`

## Cleanup

```bash
# Stop all nodes
pkill obsidiand

# Clean all data
rm -rf testnet/node*/blocks testnet/node*/chainstate testnet/node*/*.log
```

## Troubleshooting

### Nodes not connecting
- Check firewall settings
- Verify seed node (Node 1) is running first
- Check logs for connection errors

### Mining not working
- Verify `--mine=true` flag
- Check CPU usage
- Verify miner address is set

### Transactions not confirming
- Wait for next block (20 seconds)
- Check mempool with `getmempoolinfo`
- Verify transaction fee is sufficient

## Advanced Testing

### Using Python Wallet
```bash
# Generate wallet
python3 wallet.py --rpc-host=http://localhost:8545 generate

# Create address
python3 wallet.py --rpc-host=http://localhost:8545 create-address transparent

# Send transaction
python3 wallet.py --rpc-host=http://localhost:8545 send <from> <to> 10.5
```

### Docker Testing
```bash
# Use docker-compose for isolated testing
docker-compose -f docker-compose.yml up
```

## Performance Metrics

Monitor:
- Block propagation time
- Transaction confirmation time (should be ~20 seconds)
- P2P message latency
- Memory usage per node
- Disk I/O for blockchain storage

## Security Testing

Test vectors:
- Double spend attempts
- Invalid signatures
- Replay attacks
- Chain reorganization (51% attack simulation)
- P2P DDoS resistance

## Feature Checklist

- [ ] P2P peer discovery
- [ ] Block synchronization
- [ ] Mining and PoW verification
- [ ] Transaction creation and validation
- [ ] Transparent addresses (obs...)
- [ ] Shielded addresses (zobs...)
- [ ] Auto shield/unshield
- [ ] Gas limit enforcement
- [ ] Burn mechanism
- [ ] Burn redistribution
- [ ] Smart contracts
- [ ] Token issuance
- [ ] HD wallet support
