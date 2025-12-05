#!/bin/bash

# Obsidian Network Test Suite
# Tests P2P, mining, and transactions

echo "═══════════════════════════════════════════════════════"
echo "   Obsidian Network Integration Test"
echo "═══════════════════════════════════════════════════════"
echo ""

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test functions
test_rpc() {
    local port=$1
    local method=$2
    local params=$3
    
    curl -s -X POST http://localhost:$port \
        -H "Content-Type: application/json" \
        -d "{\"jsonrpc\":\"2.0\",\"method\":\"$method\",\"params\":$params,\"id\":1}"
}

print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

print_info() {
    echo -e "${YELLOW}→${NC} $1"
}

# Test 1: Check if nodes are running
echo "Test 1: Checking node connectivity..."
for port in 8545 8546 8547; do
    result=$(test_rpc $port "getblockcount" "[]" 2>/dev/null)
    if [[ $result == *"result"* ]]; then
        blocks=$(echo $result | jq -r '.result')
        print_success "Node on port $port: $blocks blocks"
    else
        print_error "Node on port $port: NOT RESPONDING"
    fi
done
echo ""

# Test 2: Get blockchain info from each node
echo "Test 2: Getting blockchain info..."
for port in 8545 8546 8547; do
    result=$(test_rpc $port "getblockchaininfo" "[]" 2>/dev/null)
    if [[ $result == *"result"* ]]; then
        chain=$(echo $result | jq -r '.result.chain')
        blocks=$(echo $result | jq -r '.result.blocks')
        print_info "Port $port: Chain=$chain, Blocks=$blocks"
    fi
done
echo ""

# Test 3: Check mining info
echo "Test 3: Checking mining status..."
for port in 8545 8546 8547; do
    result=$(test_rpc $port "getmininginfo" "[]" 2>/dev/null)
    if [[ $result == *"result"* ]]; then
        mining=$(echo $result | jq -r '.result.generate')
        blocks=$(echo $result | jq -r '.result.blocks')
        reward=$(echo $result | jq -r '.result.blockreward')
        print_info "Port $port: Mining=$mining, Blocks=$blocks, Reward=$reward OBS"
    fi
done
echo ""

# Test 4: Generate addresses
echo "Test 4: Generating test addresses..."
addr1=$(test_rpc 8545 "getnewaddress" "[]" 2>/dev/null | jq -r '.result' 2>/dev/null)
addr2=$(test_rpc 8546 "getnewaddress" "[]" 2>/dev/null | jq -r '.result' 2>/dev/null)

if [[ -n "$addr1" && "$addr1" != "null" ]]; then
    print_success "Node 1 address: $addr1"
else
    print_error "Failed to generate address on Node 1"
fi

if [[ -n "$addr2" && "$addr2" != "null" ]]; then
    print_success "Node 2 address: $addr2"
else
    print_error "Failed to generate address on Node 2"
fi
echo ""

# Test 5: Check balances (should be 0 initially)
echo "Test 5: Checking balances..."
if [[ -n "$addr1" && "$addr1" != "null" ]]; then
    result=$(test_rpc 8545 "getbalance" "[\"$addr1\"]" 2>/dev/null)
    if [[ $result == *"result"* ]]; then
        balance=$(echo $result | jq -r '.result.balance_obs' 2>/dev/null || echo "0")
        print_info "Address $addr1: $balance OBS"
    fi
fi
echo ""

# Test 6: Generate shielded addresses
echo "Test 6: Generating shielded addresses..."
zaddr1=$(test_rpc 8545 "z_getnewaddress" "[]" 2>/dev/null | jq -r '.result' 2>/dev/null)
zaddr2=$(test_rpc 8546 "z_getnewaddress" "[]" 2>/dev/null | jq -r '.result' 2>/dev/null)

if [[ -n "$zaddr1" && "$zaddr1" != "null" ]]; then
    print_success "Node 1 shielded address: ${zaddr1:0:20}..."
else
    print_info "Shielded address generation not available on Node 1"
fi

if [[ -n "$zaddr2" && "$zaddr2" != "null" ]]; then
    print_success "Node 2 shielded address: ${zaddr2:0:20}..."
else
    print_info "Shielded address generation not available on Node 2"
fi
echo ""

# Test 7: Check burn info
echo "Test 7: Checking burn and supply info..."
result=$(test_rpc 8545 "gettotalburned" "[]" 2>/dev/null)
if [[ $result == *"result"* ]]; then
    burned=$(echo $result | jq -r '.result.total_burned_obs' 2>/dev/null || echo "0")
    print_info "Total burned: $burned OBS"
else
    print_info "Burn tracking not yet available"
fi

result=$(test_rpc 8545 "getcirculatingsupply" "[]" 2>/dev/null)
if [[ $result == *"result"* ]]; then
    supply=$(echo $result | jq -r '.result.circulating_supply_obs' 2>/dev/null || echo "0")
    print_info "Circulating supply: $supply OBS"
else
    print_info "Supply info not yet available"
fi
echo ""

# Test 8: Wait and check block synchronization
echo "Test 8: Checking block synchronization..."
sleep 5

blocks1=$(test_rpc 8545 "getblockcount" "[]" 2>/dev/null | jq -r '.result')
blocks2=$(test_rpc 8546 "getblockcount" "[]" 2>/dev/null | jq -r '.result')
blocks3=$(test_rpc 8547 "getblockcount" "[]" 2>/dev/null | jq -r '.result')

echo "Block heights:"
print_info "Node 1 (8545): $blocks1 blocks"
print_info "Node 2 (8546): $blocks2 blocks"
print_info "Node 3 (8547): $blocks3 blocks"

if [[ -n "$blocks1" && -n "$blocks2" && -n "$blocks3" ]]; then
    diff12=$((blocks1 - blocks2))
    diff13=$((blocks1 - blocks3))
    diff23=$((blocks2 - blocks3))
    
    if [[ ${diff12#-} -le 2 && ${diff13#-} -le 2 && ${diff23#-} -le 2 ]]; then
        print_success "Nodes are synchronized (within 2 blocks)"
    else
        print_error "Nodes are NOT synchronized (difference > 2 blocks)"
    fi
fi
echo ""

echo "═══════════════════════════════════════════════════════"
echo "   Test Summary"
echo "═══════════════════════════════════════════════════════"
print_info "All basic tests completed"
print_info "Check individual node logs for detailed information"
echo ""
