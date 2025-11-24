package blockchain

import (
	"fmt"
	"math/big"
	"obsidian-core/consensus"
	"obsidian-core/wire"
)

// ChainReorgResult represents the result of a chain reorganization
type ChainReorgResult struct {
	OldChainTip  wire.Hash
	NewChainTip  wire.Hash
	Disconnected []*wire.MsgBlock
	Connected    []*wire.MsgBlock
}

// MaybeReorg checks if a new block causes a chain reorganization
func (b *BlockChain) MaybeReorg(newBlock *wire.MsgBlock, pow consensus.PowEngine) (*ChainReorgResult, error) {
	newBlockHash := newBlock.BlockHash()

	// Calculate work for the new block's chain
	newChainWork, err := b.calculateChainWork(newBlock)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate new chain work: %v", err)
	}

	// Get current chain work
	currentChainWork, err := b.getChainWork(b.bestHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get current chain work: %v", err)
	}

	// If new chain has more work, reorganize
	if newChainWork.Int.Cmp(currentChainWork.Int) > 0 {
		return b.reorganizeChain(newBlockHash, pow)
	}

	return nil, nil
}

// reorganizeChain reorganizes the blockchain to the new best chain
func (b *BlockChain) reorganizeChain(newTip wire.Hash, pow consensus.PowEngine) (*ChainReorgResult, error) {
	// Find the fork point
	forkPoint, err := b.findForkPoint(b.bestHash, newTip)
	if err != nil {
		return nil, fmt.Errorf("failed to find fork point: %v", err)
	}

	// Get blocks to disconnect (old chain)
	disconnectBlocks, err := b.getBlocksBetween(forkPoint, b.bestHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get disconnect blocks: %v", err)
	}

	// Get blocks to connect (new chain)
	connectBlocks, err := b.getBlocksBetween(forkPoint, newTip)
	if err != nil {
		return nil, fmt.Errorf("failed to get connect blocks: %v", err)
	}

	fmt.Printf("🔄 Chain reorganization: disconnecting %d blocks, connecting %d blocks\n",
		len(disconnectBlocks), len(connectBlocks))

	result := &ChainReorgResult{
		OldChainTip:  b.bestHash,
		NewChainTip:  newTip,
		Disconnected: disconnectBlocks,
		Connected:    connectBlocks,
	}

	// Disconnect old blocks (in reverse order)
	for i := len(disconnectBlocks) - 1; i >= 0; i-- {
		if err := b.disconnectBlock(disconnectBlocks[i]); err != nil {
			return nil, fmt.Errorf("failed to disconnect block: %v", err)
		}
	}

	// Connect new blocks
	for _, block := range connectBlocks {
		if err := b.connectBlock(block, pow); err != nil {
			// Attempt to rollback
			b.rollbackReorg(result)
			return nil, fmt.Errorf("failed to connect block: %v", err)
		}
	}

	// Update chain tip
	b.bestHash = newTip
	b.height = b.height - int32(len(disconnectBlocks)) + int32(len(connectBlocks))

	fmt.Printf("✅ Chain reorganization complete: new height %d, tip %s\n", b.height, newTip.String())

	return result, nil
}

// findForkPoint finds the common ancestor of two chain tips
func (b *BlockChain) findForkPoint(hash1, hash2 wire.Hash) (wire.Hash, error) {
	// Build path from hash1 to genesis
	path1 := make(map[wire.Hash]bool)
	current := hash1

	for current != (wire.Hash{}) {
		path1[current] = true
		block, err := b.GetBlock(current[:])
		if err != nil {
			return wire.Hash{}, err
		}
		current = block.Header.PrevBlock
	}

	// Walk from hash2 to genesis until we find a block in path1
	current = hash2
	for current != (wire.Hash{}) {
		if path1[current] {
			return current, nil
		}
		block, err := b.GetBlock(current[:])
		if err != nil {
			return wire.Hash{}, err
		}
		current = block.Header.PrevBlock
	}

	return wire.Hash{}, fmt.Errorf("no common ancestor found")
}

// getBlocksBetween returns blocks between two hashes (exclusive of start, inclusive of end)
func (b *BlockChain) getBlocksBetween(start, end wire.Hash) ([]*wire.MsgBlock, error) {
	var blocks []*wire.MsgBlock
	current := end

	for current != start && current != (wire.Hash{}) {
		block, err := b.GetBlock(current[:])
		if err != nil {
			return nil, err
		}
		blocks = append([]*wire.MsgBlock{block}, blocks...) // Prepend to maintain order
		current = block.Header.PrevBlock
	}

	return blocks, nil
}

// disconnectBlock removes a block from the active chain
func (b *BlockChain) disconnectBlock(block *wire.MsgBlock) error {
	fmt.Printf("⬅️  Disconnecting block at height %d\n", b.height)

	// Rollback UTXO set
	if err := b.utxoSet.RollbackBlock(block); err != nil {
		return fmt.Errorf("failed to rollback UTXO set: %v", err)
	}

	// Add transactions back to mempool (except coinbase)
	for _, tx := range block.Transactions {
		if !tx.IsCoinbase() {
			fee, _ := b.CalculateTransactionFee(tx, b.utxoSet)
			b.mempool.AddTransaction(tx, b.height, fee)
		}
	}

	// Rollback shielded pool
	for _, tx := range block.Transactions {
		if tx.IsShielded() {
			// TODO: Implement shielded pool rollback
		}
	}

	b.height--

	return nil
}

// connectBlock adds a block to the active chain
func (b *BlockChain) connectBlock(block *wire.MsgBlock, pow consensus.PowEngine) error {
	blockHash := block.BlockHash()
	fmt.Printf("➡️  Connecting block %s at height %d\n", blockHash.String(), b.height+1)

	// Validate block
	if err := b.validateBlockHeader(&block.Header); err != nil {
		return fmt.Errorf("invalid block header: %v", err)
	}

	// Verify PoW
	if !pow.Verify(&block.Header) {
		return fmt.Errorf("invalid proof of work")
	}

	// Validate transactions
	for _, tx := range block.Transactions {
		if !tx.IsCoinbase() {
			if err := b.ValidateTransaction(tx, b.utxoSet); err != nil {
				return fmt.Errorf("invalid transaction: %v", err)
			}
		}
	}

	// Apply UTXO changes
	if err := b.utxoSet.ApplyBlock(block, b.height+1); err != nil {
		return fmt.Errorf("failed to apply UTXO changes: %v", err)
	}

	// Remove transactions from mempool
	for _, tx := range block.Transactions {
		txHash := tx.TxHash()
		b.mempool.RemoveTransaction(txHash)
	}

	// Process shielded transactions
	for _, tx := range block.Transactions {
		if tx.IsShielded() {
			if err := b.shieldedPool.ProcessShieldedTransaction(tx); err != nil {
				return fmt.Errorf("failed to process shielded transaction: %v", err)
			}
		}
	}

	// Save block
	if err := b.db.SaveBlock(block); err != nil {
		return fmt.Errorf("failed to save block: %v", err)
	}

	b.height++

	return nil
}

// rollbackReorg attempts to rollback a failed reorganization
func (b *BlockChain) rollbackReorg(result *ChainReorgResult) {
	fmt.Println("⚠️  Rolling back failed reorganization...")

	// Disconnect blocks that were connected
	for i := len(result.Connected) - 1; i >= 0; i-- {
		if err := b.disconnectBlock(result.Connected[i]); err != nil {
			fmt.Printf("❌ Failed to rollback block: %v\n", err)
		}
	}

	// Reconnect blocks that were disconnected
	for _, block := range result.Disconnected {
		// Note: Using nil for pow as we're just reconnecting previously valid blocks
		if err := b.connectBlock(block, b.pow); err != nil {
			fmt.Printf("❌ Failed to reconnect block: %v\n", err)
		}
	}

	// Restore original chain tip
	b.bestHash = result.OldChainTip
}

// calculateChainWork calculates the total work for a chain ending at the given block
func (b *BlockChain) calculateChainWork(block *wire.MsgBlock) (*BigInt, error) {
	// Simplified work calculation
	// In production, sum the work of all blocks in the chain
	work := NewBigInt()
	current := block.BlockHash()

	for current != (wire.Hash{}) {
		blk, err := b.GetBlock(current[:])
		if err != nil {
			return nil, err
		}

		// Add this block's work
		blockWork := calculateBlockWork(blk.Header.Bits)
		work = work.Add(blockWork)

		current = blk.Header.PrevBlock
	}

	return work, nil
}

// getChainWork returns the total work for the chain ending at the given hash
func (b *BlockChain) getChainWork(hash wire.Hash) (*BigInt, error) {
	block, err := b.GetBlock(hash[:])
	if err != nil {
		return nil, err
	}
	return b.calculateChainWork(block)
}

// calculateBlockWork calculates the work for a single block based on its difficulty
func calculateBlockWork(bits uint32) *BigInt {
	// Work = 2^256 / (target + 1)
	// Simplified: just return inverse of target
	target := consensus.CompactToBig(bits)

	// Calculate 2^256
	max := new(big.Int).Lsh(big.NewInt(1), 256)

	// Divide by (target + 1)
	target.Add(target, big.NewInt(1))
	work := new(big.Int).Div(max, target)

	return &BigInt{work}
}

// BigInt wrapper for big.Int with additional methods
type BigInt struct {
	*big.Int
}

func NewBigInt() *BigInt {
	return &BigInt{big.NewInt(0)}
}

func (b *BigInt) Add(other *BigInt) *BigInt {
	b.Int.Add(b.Int, other.Int)
	return b
}

func (b *BigInt) Div(numerator, denominator *BigInt) *BigInt {
	b.Int.Div(numerator.Int, denominator.Int)
	return b
}

func (b *BigInt) SetUint64(val uint64) *BigInt {
	b.Int.SetUint64(val)
	return b
}

func (b *BigInt) Lsh(x *BigInt, n uint) *BigInt {
	b.Int.Lsh(x.Int, n)
	return b
}
