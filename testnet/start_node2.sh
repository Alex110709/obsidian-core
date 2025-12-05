#!/bin/bash

# Node 2 - Connects to Node 1
echo "Starting Node 2..."

cd /Users/yuchan/Desktop/Obsidian\ Chain/obsidian-core

# Kill existing process (if running)
pkill -f "DATA_DIR=./testnet/node2"

# Wait a bit for node 1 to start
sleep 2

# Create data directory
mkdir -p ./testnet/node2

# Start node 2
P2P_ADDR=0.0.0.0:8334 \
RPC_ADDR=0.0.0.0:8546 \
DATA_DIR=./testnet/node2 \
MINER_ADDRESS=Obs5AaMK1wATxMu5mzRDcUY3t7ArhgcGoRUD \
SOLO_MINING=true \
SEED_NODES=127.0.0.1:8333 \
LOG_LEVEL=info \
./obsidiand 2>&1 | tee ./testnet/node2/node2.log
