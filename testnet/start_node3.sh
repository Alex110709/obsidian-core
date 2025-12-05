#!/bin/bash

# Node 3 - Connects to Node 1
echo "Starting Node 3..."

cd /Users/yuchan/Desktop/Obsidian\ Chain/obsidian-core

# Kill existing process (if running)
pkill -f "DATA_DIR=./testnet/node3"

# Wait for other nodes
sleep 3

# Create data directory
mkdir -p ./testnet/node3

# Start node 3
P2P_ADDR=0.0.0.0:8335 \
RPC_ADDR=0.0.0.0:8547 \
DATA_DIR=./testnet/node3 \
MINER_ADDRESS=ObsEUHSGcLLp1enXu2aPqCWPUDZ6MM6QEUun \
SOLO_MINING=true \
SEED_NODES=127.0.0.1:8333 \
LOG_LEVEL=info \
./obsidiand 2>&1 | tee ./testnet/node3/node3.log
