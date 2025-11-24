# Obsidian Core JavaScript Implementation

A JavaScript/Node.js implementation of Obsidian Core designed for web applications, browser-based wallets, and server-side integrations. This implementation provides both Node.js and browser-compatible versions.

## Features

- 🌐 Browser and Node.js compatible
- 📱 Web wallet integration
- 🔗 RESTful and WebSocket APIs
- 📦 NPM package ecosystem
- 🖥️ Electron app support
- 🌍 CORS-enabled for web applications

## Prerequisites

- Node.js 16+
- npm or yarn package manager

## Installation

```bash
# For Node.js applications
npm install obsidian-core

# For development
git clone https://github.com/your-org/obsidian-core-js.git
cd obsidian-core-js
npm install
```

## Project Structure

```
obsidian-core-js/
├── src/
│   ├── index.js              # Main exports
│   ├── blockchain/           # Blockchain state management
│   ├── consensus/            # DarkMatter PoW implementation
│   ├── network/              # P2P networking
│   ├── mining/               # CPU miner
│   ├── rpc/                  # JSON-RPC server
│   ├── crypto/               # Cryptographic functions
│   ├── storage/              # Database abstraction
│   ├── wallet/               # Wallet functionality
│   └── utils/                # Utility functions
├── browser/                  # Browser-specific code
├── examples/                 # Usage examples
├── test/                     # Unit and integration tests
├── package.json
├── webpack.config.js         # Browser bundle config
└── README.md
```

## Quick Start

### Node.js

```javascript
const { ObsidianNode } = require('obsidian-core');

async function main() {
  const node = new ObsidianNode({
    mining: { enabled: true },
    network: { port: 8333 },
    rpc: { port: 8545 }
  });

  await node.start();
  console.log('Obsidian node started');
}

main().catch(console.error);
```

### Browser

```html
<!DOCTYPE html>
<html>
<head>
  <title>Obsidian Web Wallet</title>
</head>
<body>
  <script src="https://cdn.jsdelivr.net/npm/obsidian-core@latest/browser/obsidian.min.js"></script>
  <script>
    const { ObsidianWallet } = window.Obsidian;

    async function initWallet() {
      const wallet = new ObsidianWallet();
      await wallet.initialize();

      const address = wallet.getNewAddress();
      console.log('New address:', address);
    }

    initWallet();
  </script>
</body>
</html>
```

## Configuration

### Configuration Object

```javascript
const config = {
  network: {
    port: 8333,
    rpcPort: 8545,
    torEnabled: false,
    seedNodes: ['node1.example.com:8333']
  },
  mining: {
    enabled: true,
    minerAddress: 'your_address_here',
    threads: 4
  },
  consensus: {
    targetBlockTime: 300,
    difficultyAdjustmentInterval: 2016
  },
  storage: {
    dataDir: './data',
    dbEngine: 'leveldown'
  },
  wallet: {
    encrypted: true,
    autoBackup: true
  }
};

const node = new ObsidianNode(config);
```

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

## Core Components

### Blockchain Engine

```javascript
const { Blockchain, DarkMatterConsensus } = require('obsidian-core');

class CustomNode {
  constructor(config) {
    this.consensus = new DarkMatterConsensus(config.consensus);
    this.blockchain = new Blockchain(this.consensus, config.storage);
  }

  async initialize() {
    await this.blockchain.initialize();

    // Set up event listeners
    this.blockchain.on('block', (block) => {
      console.log('New block:', block.hash);
    });

    this.blockchain.on('transaction', (tx) => {
      console.log('New transaction:', tx.id);
    });
  }

  async start() {
    await this.initialize();
    await this.blockchain.startSync();
  }
}
```

### WebSocket Networking

```javascript
const WebSocket = require('ws');
const { PeerManager } = require('obsidian-core');

class WebSocketPeerManager extends PeerManager {
  constructor(config) {
    super(config);
    this.wss = null;
  }

  async start() {
    this.wss = new WebSocket.Server({ port: this.config.port });

    this.wss.on('connection', (ws) => {
      const peer = new WebSocketPeer(ws);
      this.addPeer(peer);

      ws.on('message', (data) => {
        this.handleMessage(peer, JSON.parse(data));
      });

      ws.on('close', () => {
        this.removePeer(peer);
      });
    });
  }

  broadcast(message) {
    this.wss.clients.forEach(client => {
      if (client.readyState === WebSocket.OPEN) {
        client.send(JSON.stringify(message));
      }
    });
  }
}
```

### Mining Implementation

```javascript
const { CpuMiner } = require('obsidian-core');

class WebMiner extends CpuMiner {
  constructor(consensus, threads = navigator.hardwareConcurrency || 4) {
    super(consensus, threads);
    this.isBrowser = typeof window !== 'undefined';
  }

  async mineBlock(template) {
    if (this.isBrowser) {
      // Use Web Workers for browser mining
      return this.mineWithWebWorkers(template);
    } else {
      // Use standard mining for Node.js
      return super.mineBlock(template);
    }
  }

  async mineWithWebWorkers(template) {
    return new Promise((resolve) => {
      const workers = [];

      for (let i = 0; i < this.threads; i++) {
        const worker = new Worker('./miner-worker.js');
        workers.push(worker);

        worker.postMessage({
          type: 'mine',
          template,
          startNonce: i * 1000000,
          endNonce: (i + 1) * 1000000
        });

        worker.onmessage = (e) => {
          if (e.data.type === 'found') {
            // Stop all workers
            workers.forEach(w => w.terminate());
            resolve(e.data.block);
          }
        };
      }
    });
  }
}
```

### Express RPC Server

```javascript
const express = require('express');
const { JsonRpcServer } = require('obsidian-core');

class ExpressRpcServer extends JsonRpcServer {
  constructor(blockchain, config) {
    super(blockchain);
    this.app = express();
    this.config = config;
    this.setupMiddleware();
    this.setupRoutes();
  }

  setupMiddleware() {
    this.app.use(express.json());

    // CORS for web applications
    this.app.use((req, res, next) => {
      res.header('Access-Control-Allow-Origin', '*');
      res.header('Access-Control-Allow-Methods', 'GET, POST, OPTIONS');
      res.header('Access-Control-Allow-Headers', 'Content-Type, Authorization');
      next();
    });
  }

  setupRoutes() {
    // JSON-RPC endpoint
    this.app.post('/rpc', async (req, res) => {
      try {
        const result = await this.handleRequest(req.body);
        res.json(result);
      } catch (error) {
        res.status(500).json({
          jsonrpc: '2.0',
          error: { code: -32603, message: error.message },
          id: req.body.id
        });
      }
    });

    // RESTful endpoints
    this.app.get('/blocks/:hash', async (req, res) => {
      const block = await this.blockchain.getBlock(req.params.hash);
      res.json(block);
    });

    this.app.get('/transactions/:id', async (req, res) => {
      const tx = await this.blockchain.getTransaction(req.params.id);
      res.json(tx);
    });
  }

  listen() {
    this.app.listen(this.config.port, () => {
      console.log(`RPC server listening on port ${this.config.port}`);
    });
  }
}
```

## API Reference

### JSON-RPC Methods

```javascript
// Available RPC methods
const rpcMethods = {
  getblockcount: 'Get current block height',
  getbestblockhash: 'Get best block hash',
  getblock: 'Get block by hash',
  getmininginfo: 'Get mining information',
  z_getnewaddress: 'Generate shielded address',
  z_sendmany: 'Send shielded transaction',
  z_getbalance: 'Get shielded balance'
};

// Example RPC call with fetch
async function getBlockCount() {
  const response = await fetch('http://localhost:8545/rpc', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      jsonrpc: '2.0',
      method: 'getblockcount',
      params: [],
      id: 1
    })
  });

  const data = await response.json();
  return data.result;
}
```

### WebSocket API

```javascript
const WebSocket = require('ws');

class WebSocketClient {
  constructor(url) {
    this.ws = new WebSocket(url);
    this.setupEventHandlers();
  }

  setupEventHandlers() {
    this.ws.on('open', () => {
      console.log('Connected to Obsidian node');
    });

    this.ws.on('message', (data) => {
      const message = JSON.parse(data);
      this.handleMessage(message);
    });

    this.ws.on('close', () => {
      console.log('Disconnected from Obsidian node');
    });
  }

  send(method, params = []) {
    const message = {
      jsonrpc: '2.0',
      method,
      params,
      id: Date.now()
    };

    this.ws.send(JSON.stringify(message));
  }

  subscribe(event) {
    this.send('subscribe', [event]);
  }

  handleMessage(message) {
    if (message.method === 'block') {
      console.log('New block:', message.params[0]);
    } else if (message.method === 'transaction') {
      console.log('New transaction:', message.params[0]);
    }
  }
}

// Usage
const client = new WebSocketClient('ws://localhost:8545/ws');
client.subscribe('blocks');
client.subscribe('transactions');
```

## Wallet Integration

### Browser Wallet

```javascript
const { ObsidianWallet } = require('obsidian-core');

class WebWallet {
  constructor() {
    this.wallet = new ObsidianWallet();
    this.node = null;
  }

  async initialize() {
    await this.wallet.initialize();

    // Connect to a public node
    this.node = new ObsidianNode({
      network: { connectOnly: true },
      rpc: { url: 'https://api.obsidian.network' }
    });
  }

  async getBalance() {
    const addresses = this.wallet.getAddresses();
    let totalBalance = 0;

    for (const address of addresses) {
      const balance = await this.node.getBalance(address);
      totalBalance += balance;
    }

    return totalBalance;
  }

  async sendTransaction(recipient, amount) {
    const utxos = await this.node.getUtxos(this.wallet.getAddresses());
    const transaction = this.wallet.createTransaction(utxos, recipient, amount);

    // Sign transaction
    const signedTx = this.wallet.signTransaction(transaction);

    // Broadcast transaction
    return await this.node.broadcastTransaction(signedTx);
  }
}
```

### React Hook

```javascript
import { useState, useEffect } from 'react';
import { ObsidianWallet } from 'obsidian-core';

export function useObsidianWallet() {
  const [wallet, setWallet] = useState(null);
  const [balance, setBalance] = useState(0);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    initializeWallet();
  }, []);

  const initializeWallet = async () => {
    try {
      const obsidianWallet = new ObsidianWallet();
      await obsidianWallet.initialize();
      setWallet(obsidianWallet);

      // Get initial balance
      const currentBalance = await getWalletBalance(obsidianWallet);
      setBalance(currentBalance);
    } catch (error) {
      console.error('Failed to initialize wallet:', error);
    } finally {
      setLoading(false);
    }
  };

  const getWalletBalance = async (walletInstance) => {
    // Implementation to get balance from connected node
    return 0; // Placeholder
  };

  const sendTransaction = async (recipient, amount) => {
    if (!wallet) return;

    try {
      const tx = await wallet.sendTransaction(recipient, amount);
      // Refresh balance after transaction
      const newBalance = await getWalletBalance(wallet);
      setBalance(newBalance);
      return tx;
    } catch (error) {
      throw error;
    }
  };

  return {
    wallet,
    balance,
    loading,
    sendTransaction
  };
}
```

## Development

### Testing

```bash
# Run all tests
npm test

# Run with coverage
npm run test:coverage

# Run specific test
npm test -- --grep "mining"

# Run integration tests
npm run test:integration
```

### Building

```bash
# Build for Node.js
npm run build

# Build for browser
npm run build:browser

# Build for Electron
npm run build:electron
```

### Code Quality

```bash
# Lint code
npm run lint

# Format code
npm run format

# Type checking (if using TypeScript)
npm run type-check
```

## Browser Compatibility

### Webpack Configuration

```javascript
// webpack.config.js
const path = require('path');

module.exports = {
  entry: './src/index.js',
  output: {
    path: path.resolve(__dirname, 'dist'),
    filename: 'obsidian.min.js',
    library: 'Obsidian',
    libraryTarget: 'umd'
  },
  resolve: {
    fallback: {
      crypto: require.resolve('crypto-browserify'),
      stream: require.resolve('stream-browserify'),
      buffer: require.resolve('buffer'),
      util: require.resolve('util')
    }
  },
  plugins: [
    new webpack.ProvidePlugin({
      Buffer: ['buffer', 'Buffer'],
      process: 'process/browser'
    })
  ]
};
```

### Browser Limitations

```javascript
// Detect browser environment
const isBrowser = typeof window !== 'undefined';

// Adjust mining threads for browser
const recommendedThreads = isBrowser
  ? Math.min(navigator.hardwareConcurrency || 4, 2) // Limit for browser
  : require('os').cpus().length;

// Use appropriate storage backend
const storageBackend = isBrowser
  ? new IndexedDBStorage()
  : new LevelDBStorage();
```

## Deployment

### NPM Package

```json
{
  "name": "obsidian-core",
  "version": "1.0.0",
  "main": "dist/index.js",
  "browser": "dist/browser/obsidian.min.js",
  "scripts": {
    "build": "webpack",
    "test": "mocha",
    "lint": "eslint src/"
  },
  "dependencies": {
    "leveldown": "^6.1.0",
    "ws": "^8.0.0",
    "express": "^4.18.0"
  },
  "devDependencies": {
    "webpack": "^5.0.0",
    "mocha": "^10.0.0",
    "eslint": "^8.0.0"
  }
}
```

### Docker

```dockerfile
FROM node:16-alpine

WORKDIR /app
COPY package*.json ./
RUN npm ci --only=production

COPY . .
EXPOSE 8333 8545

CMD ["node", "src/index.js"]
```

### Vercel Deployment

```javascript
// api/rpc.js
const { JsonRpcServer } = require('obsidian-core');

export default async function handler(req, res) {
  if (req.method !== 'POST') {
    return res.status(405).json({ error: 'Method not allowed' });
  }

  // Handle CORS
  res.setHeader('Access-Control-Allow-Origin', '*');
  res.setHeader('Access-Control-Allow-Methods', 'POST');
  res.setHeader('Access-Control-Allow-Headers', 'Content-Type');

  try {
    const server = new JsonRpcServer();
    const result = await server.handleRequest(req.body);
    res.status(200).json(result);
  } catch (error) {
    res.status(500).json({
      jsonrpc: '2.0',
      error: { code: -32603, message: error.message },
      id: req.body.id
    });
  }
}
```

## Contributing

1. Follow JavaScript/Node.js best practices
2. Write comprehensive tests
3. Update documentation
4. Run `npm run lint` and `npm test` before submitting

## License

Licensed under the MIT License.</content>
<parameter name="filePath">/Users/yuchan/Desktop/Obsidian Chain/obsidian-core/docs/javascript-implementation.md