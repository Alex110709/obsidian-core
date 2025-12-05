#!/bin/bash

# Node 1 - Seed Node
echo "Starting Node 1 (Seed Node)..."

cd /Users/yuchan/Desktop/Obsidian\ Chain/obsidian-core

# Kill existing process
pkill -f "obsidiand"

# Create data directory
mkdir -p ./testnet/node1

# Start node 1
P2P_ADDR=0.0.0.0:8333 \
RPC_ADDR=0.0.0.0:8545 \
DATA_DIR=./testnet/node1 \
MINER_ADDRESS=Obs8Uz3Bz9gfPYGZie6DGwBQUoH1a9M4XNEh \
SOLO_MINING=true \
LOG_LEVEL=info \
./obsidiand 2>&1 | tee ./testnet/node1/node1.log
