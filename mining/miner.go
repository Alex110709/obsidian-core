package mining

import (
	"fmt"
	"obsidian-core/blockchain"
	"obsidian-core/chaincfg"
	"obsidian-core/consensus"
	"obsidian-core/wire"
	"time"
)

// BlockBroadcaster is an interface for broadcasting newly mined blocks
type BlockBroadcaster interface {
	BroadcastBlock(block *wire.MsgBlock)
}

type CPUMiner struct {
	chain       *blockchain.BlockChain
	params      *chaincfg.Params
	pow         consensus.PowEngine
	minerAddr   string
	syncManager BlockBroadcaster
}

func NewCPUMiner(chain *blockchain.BlockChain, params *chaincfg.Params, pow consensus.PowEngine, minerAddr string) *CPUMiner {
	return &CPUMiner{
		chain:     chain,
		params:    params,
		pow:       pow,
		minerAddr: minerAddr,
	}
}

// SetSyncManager sets the sync manager for broadcasting blocks
func (m *CPUMiner) SetSyncManager(sm BlockBroadcaster) {
	m.syncManager = sm
}

func (m *CPUMiner) Start() {
	fmt.Println("Miner started. Mining on CPU...")

	// Check if we need to mine genesis block
	if m.chain.Height() == 0 {
		fmt.Println("⛏️  No blocks found - mining genesis block for fair launch!")
		m.mineGenesisBlock()
	}

	for {
		// 1. Get Best Block
		best, err := m.chain.BestBlock()
		if err != nil {
			fmt.Printf("Error getting best block: %v\n", err)
			time.Sleep(5 * time.Second)
			continue
		}

		// 2. Create new block template
		bestHash := best.BlockHash()
		currentHeight := m.chain.Height() + 1

		// Calculate block subsidy
		blockSubsidy := m.params.CalcBlockSubsidy(currentHeight)

		// Use previous block's difficulty - ProcessBlock will correct it if needed
		newTimestamp := time.Now()

		newBlock := wire.NewMsgBlock(&wire.BlockHeader{
			Version:   1,
			PrevBlock: bestHash,
			Timestamp: newTimestamp,
			Bits:      best.Header.Bits,
			Nonce:     0,
		})

		// TODO: Add pending transactions from mempool
		// For now, only coinbase transaction

		// Calculate total fees from transactions (currently 0, no mempool)
		totalFees := chaincfg.CalcBlockFees(newBlock.Transactions)

		// Total reward = subsidy + fees
		totalReward := blockSubsidy + totalFees

		// Add coinbase transaction with reward + fees
		coinbaseTx := wire.NewCoinbaseTx(currentHeight, totalReward, m.minerAddr)
		newBlock.AddTransaction(coinbaseTx)

		// 3. Solve PoW
		fmt.Printf("Mining block at height %d...\n", currentHeight)

		// Try to find a valid nonce
		maxAttempts := uint32(1000000) // Reasonable attempts for production
		nonce, solution, found := m.pow.SolveWithLimit(&newBlock.Header, maxAttempts)

		if found {
			newBlock.Header.Nonce = nonce
			newBlock.Header.DarkMatterSolution = solution
			fmt.Printf("✓ Found nonce: %d\n", nonce)
		} else {
			fmt.Printf("✗ Mining failed after %d attempts, skipping block\n", maxAttempts)
			time.Sleep(5 * time.Second)
			continue // Skip this block attempt
		}

		// 4. Add block to blockchain
		err = m.chain.ProcessBlock(newBlock, m.pow)
		if err != nil {
			fmt.Printf("✗ Failed to process block: %v\n", err)
			time.Sleep(5 * time.Second)
			continue
		}

		// Print success message
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Println("⛏️  NEW BLOCK MINED!")
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Printf("  Height:       %d\n", currentHeight)
		fmt.Printf("  Hash:         %x\n", newBlock.BlockHash())
		fmt.Printf("  Nonce:        %d\n", nonce)
		fmt.Printf("  Difficulty:   0x%08x\n", best.Header.Bits)
		fmt.Printf("  Transactions: %d\n", len(newBlock.Transactions))
		fmt.Printf("  Subsidy:      %d OBS\n", blockSubsidy)
		fmt.Printf("  Fees:         %d OBS\n", totalFees)
		fmt.Printf("  Total Reward: %d OBS\n", totalReward)
		fmt.Printf("  Miner:        %s\n", m.minerAddr)
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

		// Broadcast block to peers if sync manager is available
		if m.syncManager != nil {
			m.syncManager.BroadcastBlock(newBlock)
			peerCount := m.syncManager.GetPeerCount()
			fmt.Printf("📡 Block broadcast to %d peer(s)\n", peerCount)
		} else {
			fmt.Println("📡 No peers connected (mining solo)")
		}

		// Wait before next block (target: 5 minutes, but faster for demo)
		time.Sleep(30 * time.Second)
	}
}

// mineGenesisBlock mines the genesis block for fair launch
func (m *CPUMiner) mineGenesisBlock() {
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("🚀 FAIR LAUNCH - Mining Genesis Block")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("  Network:        Obsidian Mainnet")
	fmt.Println("  Max Supply:     100,000,000 OBS")
	fmt.Println("  Initial Supply: 0 OBS (NO PREMINE)")
	fmt.Println("  Block Reward:   100 OBS")
	fmt.Println("  Block Time:     5 minutes")
	fmt.Println("  Halving:        Every 420,000 blocks (~4 years)")
	fmt.Println("  Miner Address: ", m.minerAddr)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Get genesis block template
	genesis := m.params.GenesisBlock

	// Set miner address in coinbase output
	genesis.Transactions[0].TxOut[0].PkScript = []byte(m.minerAddr)

	// Mine genesis block
	fmt.Println("\n⛏️  Mining genesis block...")
	startTime := time.Now()

	maxAttempts := uint32(10000000) // 10 million attempts
	nonce, solution, found := m.pow.SolveWithLimit(&genesis.Header, maxAttempts)

	if !found {
		fmt.Printf("✗ Genesis mining failed after %d attempts\n", maxAttempts)
		fmt.Println("💡 Tip: Genesis difficulty may be too high. Adjust PowLimitBits in params.go")
		return
	}

	genesis.Header.Nonce = nonce
	genesis.Header.DarkMatterSolution = solution

	elapsed := time.Since(startTime)
	genesisHash := genesis.BlockHash()

	fmt.Printf("✅ Genesis block mined!\n")
	fmt.Printf("   Hash:     %s\n", genesisHash.String())
	fmt.Printf("   Nonce:    %d\n", nonce)
	fmt.Printf("   Time:     %v\n", elapsed)
	fmt.Printf("   Attempts: %d\n", nonce)
	fmt.Printf("   Reward:   100 OBS\n")

	// Save genesis block to blockchain
	err := m.chain.ProcessBlock(genesis, m.pow)
	if err != nil {
		fmt.Printf("✗ Failed to process genesis block: %v\n", err)
		return
	}

	fmt.Println("\n🎉 Genesis block accepted into blockchain!")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("✅ FAIR LAUNCH SUCCESSFUL")
	fmt.Println("   All 100 million OBS will be distributed to miners")
	fmt.Println("   No premine, no founder allocation")
	fmt.Println("   Pure Proof-of-Work distribution")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Broadcast genesis block to peers if sync manager is available
	if m.syncManager != nil {
		m.syncManager.BroadcastBlock(genesis)
		fmt.Println("📡 Genesis block broadcast to peers")
	}
}
