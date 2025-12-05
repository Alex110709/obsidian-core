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
	GetPeerCount() int
}

type CPUMiner struct {
	chain       *blockchain.BlockChain
	params      *chaincfg.Params
	pow         consensus.PowEngine
	minerAddr   string
	syncManager BlockBroadcaster

	// Hash rate tracking
	hashCount    uint64
	startTime    time.Time
	lastHashRate float64

	// Control
	stopChan chan struct{}
	running  bool
}

func NewCPUMiner(chain *blockchain.BlockChain, params *chaincfg.Params, pow consensus.PowEngine, minerAddr string) *CPUMiner {
	return &CPUMiner{
		chain:     chain,
		params:    params,
		pow:       pow,
		minerAddr: minerAddr,
		startTime: time.Now(),
		stopChan:  make(chan struct{}),
		running:   false,
	}
}

// SetSyncManager sets the sync manager for broadcasting blocks
func (m *CPUMiner) SetSyncManager(sm BlockBroadcaster) {
	m.syncManager = sm
}

// GetHashRate returns the current hash rate in hashes per second
func (m *CPUMiner) GetHashRate() float64 {
	elapsed := time.Since(m.startTime).Seconds()
	if elapsed == 0 {
		return 0
	}
	return float64(m.hashCount) / elapsed
}

// UpdateHashCount adds hashes to the counter
func (m *CPUMiner) UpdateHashCount(count uint64) {
	m.hashCount += count
}

// Stop stops the miner
func (m *CPUMiner) Stop() {
	if m.running {
		close(m.stopChan)
		m.running = false
		fmt.Println("Miner stopped")
	}
}

func (m *CPUMiner) Start() {
	fmt.Println("Miner started. Mining on CPU...")
	m.running = true

	// Check if we need to mine genesis block
	if m.chain.Height() == 0 {
		fmt.Println("[MINING] No blocks found - mining genesis block for fair launch!")
		m.mineGenesisBlock()
	}

	for {
		select {
		case <-m.stopChan:
			return
		default:
		}
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

		// Add pending transactions from mempool
		mempool := m.chain.Mempool()
		if mempool != nil {
			// Get transactions by priority (fee per KB)
			pendingTxs := mempool.GetTransactionsByPriority(100) // Max 100 txs per block
			for _, tx := range pendingTxs {
				// Skip coinbase transactions
				if !tx.IsCoinbase() {
					newBlock.AddTransaction(tx)
				}
			}
		}

		// Calculate total fees from transactions
		// TODO: Implement proper fee calculation with UTXO lookup
		totalFees := int64(0) // For now, no fees until UTXO lookup is implemented

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

		// Update hash count for rate calculation
		m.UpdateHashCount(uint64(maxAttempts))

		if found {
			newBlock.Header.Nonce = nonce
			newBlock.Header.DarkMatterSolution = solution
			fmt.Printf("[OK] Found nonce: %d\n", nonce)
		} else {
			fmt.Printf("[ERROR] Mining failed after %d attempts, skipping block\n", maxAttempts)
			time.Sleep(5 * time.Second)
			continue // Skip this block attempt
		}

		// 4. Add block to blockchain
		err = m.chain.ProcessBlock(newBlock, m.pow)
		if err != nil {
			fmt.Printf("[ERROR] Failed to process block: %v\n", err)
			time.Sleep(5 * time.Second)
			continue
		}

		// Print success message
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Println("[MINING] NEW BLOCK MINED!")
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
			fmt.Printf("[BROADCAST] Block broadcast to %d peer(s)\n", peerCount)
		} else {
			fmt.Println("[BROADCAST] No peers connected (mining solo)")
		}

		// Wait before next block (target: 5 minutes, but faster for demo)
		time.Sleep(30 * time.Second)
	}
}

// mineGenesisBlock mines the genesis block for fair launch
func (m *CPUMiner) mineGenesisBlock() {
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("[LAUNCH] FAIR LAUNCH - Mining Genesis Block")
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
	fmt.Println("\n[MINING] Mining genesis block...")
	startTime := time.Now()

	maxAttempts := uint32(10000000) // 10 million attempts
	nonce, solution, found := m.pow.SolveWithLimit(&genesis.Header, maxAttempts)

	// Update hash count for rate calculation
	m.UpdateHashCount(uint64(maxAttempts))

	if !found {
		fmt.Printf("[ERROR] Genesis mining failed after %d attempts\n", maxAttempts)
		fmt.Println("💡 Tip: Genesis difficulty may be too high. Adjust PowLimitBits in params.go")
		return
	}

	genesis.Header.Nonce = nonce
	genesis.Header.DarkMatterSolution = solution

	elapsed := time.Since(startTime)
	genesisHash := genesis.BlockHash()

	fmt.Printf("[SUCCESS] Genesis block mined!\n")
	fmt.Printf("   Hash:     %s\n", genesisHash.String())
	fmt.Printf("   Nonce:    %d\n", nonce)
	fmt.Printf("   Time:     %v\n", elapsed)
	fmt.Printf("   Attempts: %d\n", nonce)
	fmt.Printf("   Reward:   100 OBS\n")

	// Save genesis block to blockchain
	err := m.chain.ProcessBlock(genesis, m.pow)
	if err != nil {
		fmt.Printf("[ERROR] Failed to process genesis block: %v\n", err)
		return
	}

	fmt.Println("\n[SUCCESS] Genesis block accepted into blockchain!")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("[SUCCESS] FAIR LAUNCH SUCCESSFUL")
	fmt.Println("   All 100 million OBS will be distributed to miners")
	fmt.Println("   No premine, no founder allocation")
	fmt.Println("   Pure Proof-of-Work distribution")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Broadcast genesis block to peers if sync manager is available
	if m.syncManager != nil {
		m.syncManager.BroadcastBlock(genesis)
		fmt.Println("[BROADCAST] Genesis block broadcast to peers")
	}
}
