# Obsidian Core Python Implementation

A Python implementation of Obsidian Core designed for research, experimentation, and rapid prototyping. This implementation prioritizes readability and ease of modification over raw performance.

## Features

- 🐍 Pure Python implementation
- 🔬 Research-friendly architecture
- 📚 Educational code structure
- 🧪 Easy testing and experimentation
- 📊 Built-in analytics and visualization
- 🔧 Extensible plugin system

## Prerequisites

- Python 3.9+
- pip package manager

## Installation

```bash
git clone https://github.com/your-org/obsidian-core-python.git
cd obsidian-core-python
pip install -r requirements.txt
```

## Project Structure

```
obsidian-core-python/
├── obsidian/
│   ├── __init__.py
│   ├── blockchain.py         # Blockchain state management
│   ├── consensus.py          # DarkMatter PoW implementation
│   ├── network.py            # P2P networking with asyncio
│   ├── mining.py             # CPU miner
│   ├── rpc.py                # JSON-RPC server
│   ├── crypto/               # Cryptographic functions
│   ├── storage/              # Database abstraction
│   └── config.py             # Configuration management
├── scripts/
│   ├── run_node.py           # Main node runner
│   ├── miner.py              # Standalone miner
│   └── rpc_client.py         # RPC client for testing
├── tests/                    # Unit and integration tests
├── examples/                 # Usage examples
├── requirements.txt
├── setup.py
└── README.md
```

## Quick Start

```bash
# Install dependencies
pip install -r requirements.txt

# Run the node
python scripts/run_node.py

# Run with custom config
python scripts/run_node.py --config config.json
```

## Configuration

### Environment Variables

```bash
# Mining
OBSIDIAN_MINING_ENABLED=true
OBSIDIAN_MINER_ADDRESS=your_address_here

# Network
OBSIDIAN_P2P_PORT=8333
OBSIDIAN_RPC_PORT=8545

# Storage
OBSIDIAN_DATA_DIR=./data
```

### Configuration File (config.json)

```json
{
  "network": {
    "p2p_port": 8333,
    "rpc_port": 8545,
    "tor_enabled": false,
    "seed_nodes": ["node1.example.com:8333"]
  },
  "mining": {
    "enabled": true,
    "miner_address": "your_address_here",
    "threads": 4
  },
  "consensus": {
    "target_block_time": 300,
    "difficulty_adjustment_interval": 2016
  },
  "storage": {
    "data_dir": "./data",
    "db_engine": "sqlite"
  },
  "logging": {
    "level": "INFO",
    "file": "obsidian.log"
  }
}
```

## Core Components

### Blockchain Engine

```python
from obsidian import Blockchain, DarkMatterConsensus

async def main():
    config = load_config()
    consensus = DarkMatterConsensus(config['consensus'])
    blockchain = Blockchain(consensus, config['storage'])

    # Initialize genesis block
    await blockchain.initialize()

    # Start processing blocks
    await blockchain.run()

if __name__ == "__main__":
    asyncio.run(main())
```

### Async Networking

```python
import asyncio
from obsidian.network import PeerManager, MessageHandler

class NetworkManager:
    def __init__(self, config):
        self.config = config
        self.peers = {}
        self.server = None

    async def start(self):
        self.server = await asyncio.start_server(
            self.handle_connection,
            self.config['host'],
            self.config['port']
        )

        async with self.server:
            await self.server.serve_forever()

    async def handle_connection(self, reader, writer):
        peer = Peer(reader, writer)
        self.peers[peer.id] = peer

        try:
            await peer.handle_messages()
        finally:
            del self.peers[peer.id]
```

### Mining Implementation

```python
import hashlib
import multiprocessing as mp
from obsidian.consensus import DarkMatterConsensus

class CpuMiner:
    def __init__(self, consensus, threads=None):
        self.consensus = consensus
        self.threads = threads or mp.cpu_count()

    def mine_block(self, block_template):
        """Mine a block using multiple processes"""
        with mp.Pool(self.threads) as pool:
            results = pool.imap_unordered(
                self.try_nonce,
                [(block_template, nonce) for nonce in range(0, 2**32, self.threads)]
            )

            for result in results:
                if result:
                    return result
        return None

    def try_nonce(self, args):
        """Try a range of nonces"""
        template, start_nonce = args

        for nonce in range(start_nonce, start_nonce + 1000):
            block = template.copy()
            block.header.nonce = nonce

            if self.consensus.verify_pow(block.header):
                return block
        return None
```

### RPC Server

```python
from aiohttp import web
from obsidian.rpc import RpcHandler

class JsonRpcServer:
    def __init__(self, blockchain):
        self.blockchain = blockchain
        self.app = web.Application()
        self.setup_routes()

    def setup_routes(self):
        self.app.router.add_post('/rpc', self.handle_rpc)

    async def handle_rpc(self, request):
        data = await request.json()

        method = data.get('method')
        params = data.get('params', [])
        rpc_id = data.get('id')

        handler = RpcHandler(self.blockchain)
        result = await handler.call_method(method, params)

        return web.json_response({
            'jsonrpc': '2.0',
            'result': result,
            'id': rpc_id
        })

    def run(self, host='localhost', port=8545):
        web.run_app(self.app, host=host, port=port)
```

## API Reference

### JSON-RPC Methods

```python
# Available RPC methods
rpc_methods = {
    'getblockcount': 'Get current block height',
    'getbestblockhash': 'Get best block hash',
    'getblock': 'Get block by hash',
    'getmininginfo': 'Get mining information',
    'z_getnewaddress': 'Generate shielded address',
    'z_sendmany': 'Send shielded transaction',
    'z_getbalance': 'Get shielded balance'
}

# Example RPC call
async def get_block_count():
    async with aiohttp.ClientSession() as session:
        payload = {
            'jsonrpc': '2.0',
            'method': 'getblockcount',
            'params': [],
            'id': 1
        }

        async with session.post('http://localhost:8545/rpc', json=payload) as resp:
            data = await resp.json()
            return data['result']
```

## Development

### Testing

```bash
# Run all tests
python -m pytest

# Run with coverage
python -m pytest --cov=obsidian --cov-report=html

# Run specific test
python -m pytest tests/test_consensus.py::test_darkmatter_pow

# Run integration tests
python -m pytest tests/integration/
```

### Code Quality

```bash
# Format code
black obsidian/
isort obsidian/

# Lint code
flake8 obsidian/
mypy obsidian/

# Check security
bandit -r obsidian/
```

### Debugging

```python
import logging
logging.basicConfig(level=logging.DEBUG)

# Enable debug logging for specific modules
logging.getLogger('obsidian.network').setLevel(logging.DEBUG)
logging.getLogger('obsidian.consensus').setLevel(logging.DEBUG)
```

## Research and Experimentation

### Custom Consensus Rules

```python
class ExperimentalConsensus(DarkMatterConsensus):
    def __init__(self, config):
        super().__init__(config)
        self.experimental_param = config.get('experimental_param', 1.0)

    def verify_pow(self, header):
        # Custom verification logic
        base_result = super().verify_pow(header)

        # Add experimental modifications
        experimental_check = self.experimental_verification(header)

        return base_result and experimental_check

    def experimental_verification(self, header):
        # Implement your experimental consensus rules
        return True
```

### Network Analysis

```python
class NetworkAnalyzer:
    def __init__(self, network_manager):
        self.network = network_manager
        self.stats = {
            'messages_received': 0,
            'messages_sent': 0,
            'peers_connected': 0,
            'blocks_propagated': 0
        }

    def analyze_traffic(self):
        """Analyze network traffic patterns"""
        # Implementation for traffic analysis
        pass

    def plot_network_graph(self):
        """Generate network topology visualization"""
        import networkx as nx
        import matplotlib.pyplot as plt

        G = nx.Graph()
        for peer_id, peer in self.network.peers.items():
            G.add_node(peer_id)
            for neighbor in peer.neighbors:
                G.add_edge(peer_id, neighbor)

        nx.draw(G, with_labels=True)
        plt.show()
```

### Performance Benchmarking

```python
import time
import cProfile

def benchmark_mining():
    """Benchmark mining performance"""
    miner = CpuMiner(DarkMatterConsensus())
    template = create_test_block_template()

    start_time = time.time()
    pr = cProfile.Profile()
    pr.enable()

    result = miner.mine_block(template)

    pr.disable()
    end_time = time.time()

    print(f"Mining took {end_time - start_time:.2f} seconds")
    pr.print_stats(sort='cumulative')

    return result
```

## Educational Examples

### Simple Block Creation

```python
from obsidian import Block, BlockHeader, Transaction

def create_genesis_block():
    """Create the genesis block"""
    header = BlockHeader(
        version=1,
        prev_block_hash='0' * 64,
        merkle_root=calculate_merkle_root([]),
        timestamp=int(time.time()),
        bits=0x2000ffff,
        nonce=0
    )

    # Find valid nonce
    consensus = DarkMatterConsensus()
    while not consensus.verify_pow(header):
        header.nonce += 1

    return Block(header, [])
```

### Wallet Implementation

```python
from obsidian.crypto import KeyPair, Address

class SimpleWallet:
    def __init__(self):
        self.keypair = KeyPair.generate()
        self.address = Address.from_public_key(self.keypair.public_key)

    def create_transaction(self, recipient, amount):
        """Create a simple transaction"""
        tx = Transaction(
            sender=self.address,
            recipient=recipient,
            amount=amount,
            fee=10000
        )

        # Sign transaction
        tx.signature = self.keypair.sign(tx.hash())

        return tx

    def get_balance(self, blockchain):
        """Get wallet balance"""
        balance = 0
        for block in blockchain.blocks:
            for tx in block.transactions:
                if tx.recipient == self.address:
                    balance += tx.amount
                if tx.sender == self.address:
                    balance -= tx.amount + tx.fee
        return balance
```

## Deployment

### Docker

```dockerfile
FROM python:3.9-slim

WORKDIR /app
COPY requirements.txt .
RUN pip install -r requirements.txt

COPY . .
EXPOSE 8333 8545

CMD ["python", "scripts/run_node.py"]
```

### Docker Compose

```yaml
services:
  obsidian-python:
    build: .
    ports:
      - "8333:8333"
      - "8545:8545"
    volumes:
      - obsidian-data:/app/data
    environment:
      - OBSIDIAN_MINING_ENABLED=true
      - OBSIDIAN_MINER_ADDRESS=your_address

volumes:
  obsidian-data:
```

## Contributing

1. Follow PEP 8 style guidelines
2. Write comprehensive tests
3. Update documentation
4. Run `black` and `flake8` before submitting

## License

Licensed under the MIT License.</content>
<parameter name="filePath">/Users/yuchan/Desktop/Obsidian Chain/obsidian-core/docs/python-implementation.md