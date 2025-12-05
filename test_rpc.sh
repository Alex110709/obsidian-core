#!/bin/bash

# RPC 테스트 스크립트

HOST="http://localhost:8545"

echo "=== Obsidian RPC Tests ==="
echo ""

echo "1. Get Block Count:"
curl -s -X POST $HOST -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"getblockcount","params":[],"id":1}' | jq .
echo ""

echo "2. Get Blockchain Info:"
curl -s -X POST $HOST -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"getblockchaininfo","params":[],"id":1}' | jq .
echo ""

echo "3. Get Mining Info:"
curl -s -X POST $HOST -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"getmininginfo","params":[],"id":1}' | jq .
echo ""

echo "4. Get New Address:"
curl -s -X POST $HOST -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"getnewaddress","params":[],"id":1}' | jq .
echo ""

echo "5. Get Total Burned:"
curl -s -X POST $HOST -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"gettotalburned","params":[],"id":1}' | jq .
echo ""

echo "6. Get Circulating Supply:"
curl -s -X POST $HOST -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"getcirculatingsupply","params":[],"id":1}' | jq .
echo ""

echo "7. Generate Shielded Address:"
curl -s -X POST $HOST -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"z_getnewaddress","params":[],"id":1}' | jq .
echo ""

echo "=== Tests Complete ==="
