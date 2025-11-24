# Obsidian Core Rust Implementation

A high-performance implementation of Obsidian Core written in Rust, focusing on memory safety, concurrency, and performance optimization.

## Features

- 🚀 High-performance networking with async/await
- 🔒 Memory safety guarantees
- ⚡ Zero-cost abstractions
- 🏗️ Modular architecture
- 🔧 Easy deployment and maintenance
- 📊 Advanced metrics and monitoring

## Prerequisites

- Rust 1.70+
- Cargo package manager

## Installation

```bash
git clone https://github.com/your-org/obsidian-core-rust.git
cd obsidian-core-rust
cargo build --release
```

## Project Structure

```
obsidian-core-rust/
├── src/
│   ├── main.rs              # Application entry point
│   ├── blockchain/          # Blockchain state management
│   ├── consensus/           # DarkMatter PoW implementation
│   ├── network/             # P2P networking with Tokio
│   ├── mining/              # CPU miner with rayon
│   ├── rpc/                 # JSON-RPC server with warp
│   ├── crypto/              # Cryptographic primitives
│   ├── storage/             # Database abstraction
│   └── config/              # Configuration management
├── Cargo.toml               # Dependencies and metadata
└── docs/                    # Documentation
```

## Quick Start

```bash
# Build and run
cargo run --release

# Run with custom config
cargo run --release -- --config config.toml
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
OBSIDIAN_TOR_ENABLED=false

# Storage
OBSIDIAN_DATA_DIR=./data
```

### Configuration File (config.toml)

```toml
[network]
p2p_port = 8333
rpc_port = 8545
tor_enabled = false
seed_nodes = ["node1.example.com:8333"]

[mining]
enabled = true
miner_address = "your_address_here"
threads = 4

[consensus]
target_block_time = 300  # 5 minutes in seconds
difficulty_adjustment_interval = 2016

[storage]
data_dir = "./data"
db_engine = "sled"  # or "rocksdb"
```

## Core Components

### Blockchain Engine

```rust
use obsidian_core::{Blockchain, Consensus};

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let config = Config::from_env()?;
    let consensus = DarkMatterConsensus::new(config.consensus);
    let blockchain = Blockchain::new(consensus, config.storage).await?;

    // Start the node
    let node = Node::new(blockchain, config.network).await?;
    node.run().await?;

    Ok(())
}
```

### Async Networking

```rust
use tokio::net::TcpListener;
use obsidian_network::{Peer, MessageHandler};

pub struct NetworkManager {
    peers: Arc<RwLock<HashMap<PeerId, Peer>>>,
    listener: TcpListener,
}

impl NetworkManager {
    pub async fn listen(&self) -> Result<(), NetworkError> {
        loop {
            let (socket, addr) = self.listener.accept().await?;
            let peer = Peer::new(socket, addr);

            tokio::spawn(async move {
                peer.handle_connection().await
            });
        }
    }
}
```

### Mining Implementation

```rust
use rayon::prelude::*;
use obsidian_mining::{Miner, Work};

pub struct CpuMiner {
    threads: usize,
    consensus: Arc<Consensus>,
}

impl CpuMiner {
    pub fn mine(&self, work: Work) -> Option<Block> {
        (0..u32::MAX).into_par_iter()
            .find_map_any(|nonce| {
                let mut block = work.template.clone();
                block.header.nonce = nonce;

                if self.consensus.verify_pow(&block.header) {
                    Some(block)
                } else {
                    None
                }
            })
    }
}
```

### RPC Server

```rust
use warp::Filter;
use obsidian_rpc::{RpcServer, RpcMethod};

pub struct JsonRpcServer {
    blockchain: Arc<Blockchain>,
    methods: HashMap<String, RpcMethod>,
}

impl JsonRpcServer {
    pub async fn serve(self, addr: impl Into<SocketAddr>) {
        let routes = self.routes();
        warp::serve(routes).run(addr).await;
    }

    fn routes(&self) -> impl Filter<Extract = impl warp::Reply, Error = warp::Rejection> + Clone {
        let blockchain = Arc::clone(&self.blockchain);

        warp::post()
            .and(warp::body::json())
            .and_then(move |request: RpcRequest| {
                let blockchain = Arc::clone(&blockchain);
                async move {
                    self.handle_request(request, blockchain).await
                }
            })
    }
}
```

## API Reference

### JSON-RPC Methods

```rust
#[derive(Serialize, Deserialize)]
pub struct RpcRequest {
    pub jsonrpc: String,
    pub method: String,
    pub params: Vec<Value>,
    pub id: Value,
}

#[derive(Serialize, Deserialize)]
pub struct RpcResponse {
    pub jsonrpc: String,
    pub result: Value,
    pub id: Value,
}
```

Available methods:
- `getblockcount` - Current block height
- `getbestblockhash` - Best block hash
- `getblock` - Block by hash
- `getmininginfo` - Mining statistics
- `z_getnewaddress` - Generate shielded address
- `z_sendmany` - Send shielded transaction

## Development

### Testing

```bash
# Run all tests
cargo test

# Run with output
cargo test -- --nocapture

# Run specific test
cargo test test_darkmatter_pow

# Run benchmarks
cargo bench
```

### Building

```bash
# Debug build
cargo build

# Release build (optimized)
cargo build --release

# Cross-compilation
cargo build --release --target x86_64-unknown-linux-gnu
```

### Code Quality

```bash
# Format code
cargo fmt

# Lint code
cargo clippy

# Check documentation
cargo doc --open
```

## Performance Optimization

### CPU Mining Optimization

```rust
// Use SIMD for hash calculations
#[cfg(target_arch = "x86_64")]
use std::arch::x86_64::*;

pub fn darkmatter_hash_simd(data: &[u8]) -> [u8; 32] {
    unsafe {
        let mut hash = [0u8; 32];
        // SIMD-accelerated SHA-256 implementation
        // ...
        hash
    }
}
```

### Memory Pool Optimization

```rust
use dashmap::DashMap; // Concurrent HashMap

pub struct MemoryPool {
    transactions: DashMap<TxId, Transaction>,
    by_fee: BTreeMap<Reverse<i64>, TxId>, // Max-heap by fee
}

impl MemoryPool {
    pub fn add_transaction(&self, tx: Transaction) -> Result<(), MempoolError> {
        let txid = tx.id();
        let fee = tx.calculate_fee();

        self.transactions.insert(txid, tx);
        self.by_fee.insert(Reverse(fee), txid);

        Ok(())
    }
}
```

## Deployment

### Docker

```dockerfile
FROM rust:1.70-slim as builder
WORKDIR /app
COPY . .
RUN cargo build --release

FROM debian:bullseye-slim
RUN apt-get update && apt-get install -y ca-certificates
COPY --from=builder /app/target/release/obsidian-core-rust /usr/local/bin/
CMD ["obsidian-core-rust"]
```

### Systemd Service

```ini
[Unit]
Description=Obsidian Core Rust Node
After=network.target

[Service]
Type=simple
User=obsidian
ExecStart=/usr/local/bin/obsidian-core-rust --config /etc/obsidian/config.toml
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

## Monitoring

### Metrics

```rust
use prometheus::{Encoder, TextEncoder, register_counter, register_gauge};

lazy_static! {
    static ref BLOCKS_MINED: Counter = register_counter!(
        "obsidian_blocks_mined_total",
        "Total number of blocks mined"
    ).unwrap();

    static ref PEERS_CONNECTED: Gauge = register_gauge!(
        "obsidian_peers_connected",
        "Number of connected peers"
    ).unwrap();
}
```

### Logging

```rust
use tracing::{info, error, instrument};
use tracing_subscriber;

#[instrument]
pub async fn process_block(&self, block: Block) -> Result<(), Error> {
    info!("Processing block {}", block.hash());

    // Process block logic...

    info!("Block processed successfully");
    Ok(())
}
```

## Contributing

1. Follow Rust coding standards
2. Write comprehensive tests
3. Update documentation
4. Run `cargo fmt` and `cargo clippy` before submitting

## License

Licensed under the Apache License 2.0.</content>
<parameter name="filePath">/Users/yuchan/Desktop/Obsidian Chain/obsidian-core/docs/rust-implementation.md