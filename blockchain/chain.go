package blockchain

import (
	"bytes"
	"fmt"
	"math/big"
	"obsidian-core/chaincfg"
	"obsidian-core/consensus"
	"obsidian-core/database"
	"obsidian-core/wire"
	"strconv"
	"strings"
)

// BlockChain provides functions for working with the bitcoin block chain.
// It includes functionality such as rejecting duplicate blocks, ensuring blocks
// follow all rules, orphan handling, checkpointing, and best chain selection
// with reorganization.
type BlockChain struct {
	params       *chaincfg.Params
	db           *database.Storage
	pow          consensus.PowEngine
	bestHash     wire.Hash
	height       int32
	shieldedPool *ShieldedPool
	utxoSet      *UTXOSet
	mempool      *Mempool
	feeEstimator *FeeEstimator
	tokenStore   *TokenStore
}

// TokenStore manages token operations
type TokenStore struct {
	tokens   map[wire.Hash]*Token
	balances map[string]map[wire.Hash]int64 // address -> tokenID -> balance
}

// NewTokenStore creates a new token store
func NewTokenStore() *TokenStore {
	return &TokenStore{
		tokens:   make(map[wire.Hash]*Token),
		balances: make(map[string]map[wire.Hash]int64),
	}
}

// Token represents a token
type Token struct {
	ID          wire.Hash
	Symbol      string
	Name        string
	Decimals    int
	Supply      int64
	TotalSupply int64
	Owner       string
	Mintable    bool
	Created     int64
}

// TokenBalance tracks token balances for addresses
type TokenBalance struct {
	balances map[string]map[wire.Hash]int64 // address -> tokenID -> balance
}

// GetToken retrieves a token by ID
func (ts *TokenStore) GetToken(id wire.Hash) (*Token, error) {
	token, ok := ts.tokens[id]
	if !ok {
		return nil, fmt.Errorf("token not found")
	}
	return token, nil
}

// GetTokenBySymbol retrieves a token by symbol
func (ts *TokenStore) GetTokenBySymbol(symbol string) (*Token, error) {
	for _, token := range ts.tokens {
		if token.Symbol == symbol {
			return token, nil
		}
	}
	return nil, fmt.Errorf("token not found")
}

// GetBalance gets the balance for an address and token
func (ts *TokenStore) GetBalance(address string, tokenID wire.Hash) int64 {
	// Simplified: return 0 for now
	return 0
}

// TransferToken transfers tokens between addresses
func (ts *TokenStore) TransferToken(tokenID wire.Hash, from, to string, amount int64) error {
	// Simplified: no-op
	return nil
}

// ListTokens returns all tokens
func (ts *TokenStore) ListTokens() []*Token {
	tokens := make([]*Token, 0, len(ts.tokens))
	for _, token := range ts.tokens {
		tokens = append(tokens, token)
	}
	return tokens
}

// GetAddressTokens returns tokens held by an address with balances
func (ts *TokenStore) GetAddressTokens(address string) map[wire.Hash]int64 {
	if addressBalances, ok := ts.balances[address]; ok {
		return addressBalances
	}
	return make(map[wire.Hash]int64)
}

// NewBlockchain returns a BlockChain instance using the provided configuration
// details.
func NewBlockchain(params *chaincfg.Params, pow consensus.PowEngine) (*BlockChain, error) {
	db, err := database.NewStorage()
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	// Get the underlying bolt database
	boltDB := db.DB()

	bc := &BlockChain{
		params:       params,
		db:           db,
		pow:          pow,
		height:       0,
		shieldedPool: NewShieldedPool(),
		utxoSet:      NewUTXOSet(boltDB),
		mempool:      NewMempool(),
		feeEstimator: NewFeeEstimator(),
		tokenStore:   NewTokenStore(),
	}

	// Save genesis block if it doesn't exist
	genesisHash := params.GenesisBlock.BlockHash()
	_, err = db.GetBlock(genesisHash[:])
	if err != nil {
		// Genesis doesn't exist, save it
		if err := db.SaveBlock(params.GenesisBlock); err != nil {
			return nil, fmt.Errorf("failed to save genesis: %v", err)
		}
	}
	bc.bestHash = genesisHash

	return bc, nil
}

// BestBlock returns the block at the tip of the chain.
func (b *BlockChain) BestBlock() (*wire.MsgBlock, error) {
	return b.db.GetBlock(b.bestHash[:])
}

// Height returns the height of the best block.
func (b *BlockChain) Height() int32 {
	return b.height
}

// Close closes the database.
func (b *BlockChain) Close() {
	b.db.Close()
}

// GetBlock retrieves a block by its hash.
func (b *BlockChain) GetBlock(hash []byte) (*wire.MsgBlock, error) {
	return b.db.GetBlock(hash)
}

// Params returns the chain parameters.
func (b *BlockChain) Params() *chaincfg.Params {
	return b.params
}

// ShieldedPool returns the shielded pool.
func (b *BlockChain) ShieldedPool() *ShieldedPool {
	return b.shieldedPool
}

// Mempool returns the mempool.
func (b *BlockChain) Mempool() *Mempool {
	return b.mempool
}

// FeeEstimator returns the fee estimator.
func (b *BlockChain) FeeEstimator() *FeeEstimator {
	return b.feeEstimator
}

// GetTokenStore returns the token store
func (b *BlockChain) GetTokenStore() *TokenStore {
	return b.tokenStore
}

// ProcessBlock is the main workhorse for handling insertion of new blocks into
// the block chain.  It includes functionality such as rejecting duplicate
// blocks, ensuring blocks follow all rules, orphan handling, and best chain
// selection.
func (b *BlockChain) ProcessBlock(block *wire.MsgBlock, pow consensus.PowEngine) error {
	// Use provided PoW engine if given, otherwise use chain's default
	if pow == nil {
		pow = b.pow
	}
	// 1. Calculate correct difficulty for this block
	currentHeight := b.height + 1
	correctDifficulty, err := b.calculateDifficultyForHeight(currentHeight, block.Header.Timestamp.Unix())
	if err != nil {
		return fmt.Errorf("failed to calculate difficulty: %v", err)
	}

	// Update block header with correct difficulty if different
	if block.Header.Bits != correctDifficulty {
		fmt.Printf("Correcting block difficulty: %08x -> %08x\n", block.Header.Bits, correctDifficulty)
		block.Header.Bits = correctDifficulty
	}

	// 2. Validate block header
	if err := b.validateBlockHeader(&block.Header); err != nil {
		return fmt.Errorf("invalid block header: %v", err)
	}

	// 3. Verify PoW
	if !pow.Verify(&block.Header) {
		return fmt.Errorf("invalid proof of work")
	}

	// 4. Check if block already exists
	blockHash := block.BlockHash()

	// 4a. Check against checkpoints
	if err := b.validateCheckpoint(currentHeight, blockHash); err != nil {
		return fmt.Errorf("checkpoint validation failed: %v", err)
	}
	if _, err := b.db.GetBlock(blockHash[:]); err == nil {
		return fmt.Errorf("block already exists")
	}

	// 4. Validate transactions
	if len(block.Transactions) == 0 {
		return fmt.Errorf("block has no transactions")
	}

	// 4a. Validate shielded transactions
	for _, tx := range block.Transactions {
		if tx.IsShielded() {
			if err := b.shieldedPool.ValidateShieldedTransaction(tx); err != nil {
				return fmt.Errorf("invalid shielded transaction: %v", err)
			}
		}
	}

	// 5. Validate block reward
	if err := b.validateBlockReward(block); err != nil {
		return fmt.Errorf("invalid block reward: %v", err)
	}

	// 5a. Process shielded transactions
	for _, tx := range block.Transactions {
		if tx.IsShielded() {
			if err := b.shieldedPool.ProcessShieldedTransaction(tx); err != nil {
				return fmt.Errorf("failed to process shielded transaction: %v", err)
			}
		}
	}

	// 6. Save block
	if err := b.db.SaveBlock(block); err != nil {
		return fmt.Errorf("failed to save block: %v", err)
	}

	// 7. Update chain state
	b.bestHash = blockHash
	b.height++

	// 8. Update fee estimator
	b.feeEstimator.AddBlock(block, b.height)

	fmt.Printf("Block accepted! Height: %d, Hash: %s\n", b.height, blockHash.String())
	return nil
}

// validateBlockHeader performs header validation
func (b *BlockChain) validateBlockHeader(header *wire.BlockHeader) error {
	// Check block size limit via DarkMatterSolution size
	if len(header.DarkMatterSolution) > 1024 {
		return fmt.Errorf("solution too large")
	}

	// Check target difficulty (simplified)
	if header.Bits == 0 {
		return fmt.Errorf("invalid difficulty bits")
	}

	return nil
}

// validateBlockReward validates the coinbase transaction reward.
func (b *BlockChain) validateBlockReward(block *wire.MsgBlock) error {
	if len(block.Transactions) == 0 {
		return fmt.Errorf("no coinbase transaction")
	}

	coinbaseTx := block.Transactions[0]

	// Check if it's a coinbase (first input should have null hash)
	if len(coinbaseTx.TxIn) == 0 {
		return fmt.Errorf("coinbase has no inputs")
	}

	// Calculate expected block subsidy
	expectedReward := b.params.CalcBlockSubsidy(b.height + 1)

	// Calculate total fees from transactions
	totalFees := chaincfg.CalcBlockFees(block.Transactions)

	// Total allowed = subsidy + fees
	maxAllowed := expectedReward + totalFees

	// Sum all outputs in coinbase
	totalOutput := int64(0)
	for _, txOut := range coinbaseTx.TxOut {
		totalOutput += txOut.Value
	}

	// Reward must not exceed expected amount + fees
	if totalOutput > maxAllowed {
		return fmt.Errorf("coinbase output %d exceeds reward %d + fees %d", totalOutput, expectedReward, totalFees)
	}

	return nil
}

// CompactToBig converts a compact representation to a big.Int
func CompactToBig(compact uint32) *big.Int {
	mantissa := compact & 0x007fffff
	isNegative := compact&0x00800000 != 0
	exponent := uint(compact >> 24)

	var bn *big.Int
	if exponent <= 3 {
		mantissa >>= 8 * (3 - exponent)
		bn = big.NewInt(int64(mantissa))
	} else {
		bn = big.NewInt(int64(mantissa))
		bn.Lsh(bn, 8*(exponent-3))
	}

	if isNegative {
		bn = bn.Neg(bn)
	}

	return bn
}

// calculateDifficultyForHeight calculates the correct difficulty for a given block height
func (b *BlockChain) calculateDifficultyForHeight(height int32, blockTime int64) (uint32, error) {
	// Get the previous block
	prevBlock, err := b.BestBlock()
	if err != nil {
		// Genesis block
		return b.params.PowLimitBits, nil
	}

	// Check if this is a retarget height
	retargetInterval := int32(b.params.TargetTimespan / b.params.TargetTimePerBlock)
	if height%retargetInterval == 0 && height > 0 {
		// This is a retarget block - calculate new difficulty
		return b.CalcNextRequiredDifficulty(prevBlock, blockTime)
	}

	// Not a retarget block - use previous block's difficulty
	return prevBlock.Header.Bits, nil
}

// BigToCompact converts a big.Int to compact representation
func BigToCompact(n *big.Int) uint32 {
	if n.Sign() == 0 {
		return 0
	}

	bytes := n.Bytes()
	size := uint32(len(bytes))

	var compact uint32
	if size <= 3 {
		compact = uint32(bytes[0])
		if size > 1 {
			compact <<= 8
			compact |= uint32(bytes[1])
		}
		if size > 2 {
			compact <<= 8
			compact |= uint32(bytes[2])
		}
		compact <<= 8 * (3 - size)
	} else {
		compact = uint32(bytes[0])<<16 | uint32(bytes[1])<<8 | uint32(bytes[2])
	}

	compact |= size << 24

	if n.Sign() < 0 {
		compact |= 0x00800000
	}

	return compact
}

// CalcNextRequiredDifficulty calculates the required difficulty for the next block
// using Bitcoin's exact difficulty adjustment algorithm.
func (b *BlockChain) CalcNextRequiredDifficulty(lastBlock *wire.MsgBlock, newBlockTime int64) (uint32, error) {
	// Bitcoin adjusts difficulty every 2016 blocks
	// For 5-minute blocks: 2016 blocks = 1 week
	retargetInterval := int32(b.params.TargetTimespan / b.params.TargetTimePerBlock)

	// Genesis block or not at retarget interval - keep same difficulty
	if b.height == 0 || (b.height+1)%retargetInterval != 0 {
		// Check for minimum difficulty rules (testnet only)
		if b.params.ReduceMinDifficulty {
			return b.calcEasierDifficulty(lastBlock)
		}
		return lastBlock.Header.Bits, nil
	}

	// Get the first block of this difficulty period (2016 blocks ago)
	// Bitcoin uses: block[height - 2015] because it includes current block
	firstRetargetHeight := b.height - retargetInterval + 1
	firstBlock, err := b.getBlockByHeight(firstRetargetHeight)
	if err != nil {
		// If we can't find it, keep the same difficulty
		return lastBlock.Header.Bits, nil
	}

	// Calculate actual timespan between first and last block of this period
	actualTimespan := lastBlock.Header.Timestamp.Unix() - firstBlock.Header.Timestamp.Unix()

	// Bitcoin limits adjustment to prevent extreme changes
	// Min: 1/4 of target (if blocks found 4x faster)
	// Max: 4x of target (if blocks found 4x slower)
	targetTimespan := int64(b.params.TargetTimespan.Seconds())
	minTimespan := targetTimespan / b.params.RetargetAdjustmentFactor
	maxTimespan := targetTimespan * b.params.RetargetAdjustmentFactor

	adjustedTimespan := actualTimespan
	if adjustedTimespan < minTimespan {
		adjustedTimespan = minTimespan
	} else if adjustedTimespan > maxTimespan {
		adjustedTimespan = maxTimespan
	}

	// Bitcoin formula: new_target = old_target * (actual_time / target_time)
	lastTarget := CompactToBig(lastBlock.Header.Bits)
	newTarget := new(big.Int).Mul(lastTarget, big.NewInt(adjustedTimespan))
	newTarget.Div(newTarget, big.NewInt(targetTimespan))

	// Never allow difficulty to go below the minimum (PowLimit)
	if newTarget.Cmp(b.params.PowLimit) > 0 {
		newTarget.Set(b.params.PowLimit)
	}

	newDifficulty := BigToCompact(newTarget)

	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("Difficulty Retarget at Height %d\n", b.height+1)
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("  Blocks in period: %d (from height %d to %d)\n", retargetInterval, firstRetargetHeight, b.height)
	fmt.Printf("  Actual timespan:  %d seconds (%.2f days)\n", actualTimespan, float64(actualTimespan)/86400)
	fmt.Printf("  Target timespan:  %d seconds (%.2f days)\n", targetTimespan, float64(targetTimespan)/86400)
	fmt.Printf("  Adjusted (clamped): %d seconds\n", adjustedTimespan)
	fmt.Printf("  Adjustment ratio: %.2f%%\n", (float64(adjustedTimespan)/float64(targetTimespan))*100)
	fmt.Printf("  Old difficulty:   0x%08x\n", lastBlock.Header.Bits)
	fmt.Printf("  New difficulty:   0x%08x\n", newDifficulty)

	if adjustedTimespan < targetTimespan {
		fmt.Printf("  📈 Difficulty INCREASED (blocks found faster)\n")
	} else if adjustedTimespan > targetTimespan {
		fmt.Printf("  📉 Difficulty DECREASED (blocks found slower)\n")
	} else {
		fmt.Printf("  ➡️  Difficulty unchanged (perfect timing)\n")
	}
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")

	return newDifficulty, nil
}

// calcEasierDifficulty calculates an easier difficulty if too much time has passed
func (b *BlockChain) calcEasierDifficulty(lastBlock *wire.MsgBlock) (uint32, error) {
	// Get the minimum difficulty
	return b.params.PowLimitBits, nil
}

// getBlockByHeight retrieves a block by its height (inefficient, should use height index)
func (b *BlockChain) getBlockByHeight(height int32) (*wire.MsgBlock, error) {
	// This is a simplified implementation
	// In production, you'd want a height->hash index
	if height == 0 {
		return b.params.GenesisBlock, nil
	}

	// For now, we'll iterate backwards from the current best block
	// This is inefficient but works for demonstration
	currentBlock, err := b.BestBlock()
	if err != nil {
		return nil, err
	}

	for i := b.height; i > height; i-- {
		prevHash := currentBlock.Header.PrevBlock
		currentBlock, err = b.db.GetBlock(prevHash[:])
		if err != nil {
			return nil, fmt.Errorf("block at height %d not found: %v", i-1, err)
		}
	}

	return currentBlock, nil
}

// validateCheckpoint validates that a block matches the checkpoint at its height.
func (b *BlockChain) validateCheckpoint(height int32, blockHash wire.Hash) error {
	// Check if there's a checkpoint at this height
	for _, checkpoint := range b.params.Checkpoints {
		if checkpoint.Height == height {
			if !bytes.Equal(checkpoint.Hash[:], blockHash[:]) {
				return fmt.Errorf("checkpoint mismatch at height %d", height)
			}
			fmt.Printf("✓ Checkpoint validated at height %d\n", height)
		}
	}
	return nil
}

// GetLatestCheckpoint returns the latest checkpoint before or at the given height.
func (b *BlockChain) GetLatestCheckpoint(height int32) *chaincfg.Checkpoint {
	var latest *chaincfg.Checkpoint
	for i := range b.params.Checkpoints {
		checkpoint := &b.params.Checkpoints[i]
		if checkpoint.Height <= height {
			if latest == nil || checkpoint.Height > latest.Height {
				latest = checkpoint
			}
		}
	}
	return latest
}

// processTokenTransferOwnership processes a token ownership transfer transaction
func (b *BlockChain) processTokenTransferOwnership(tx *wire.MsgTx) error {
	// Extract token ID (first 32 bytes)
	tokenID := wire.Hash{}
	copy(tokenID[:], tx.Memo[:32])

	// Parse transfer data
	memoStr := string(tx.Memo[32:])
	newOwner := memoStr

	if newOwner == "" {
		return fmt.Errorf("new owner address is required")
	}

	// Check if token exists
	token, err := b.tokenStore.GetToken(tokenID)
	if err != nil {
		return fmt.Errorf("token does not exist: %v", err)
	}

	// Check sender is current owner (from output address)
	if len(tx.TxOut) == 0 {
		return fmt.Errorf("token ownership transfer transaction missing output")
	}
	sender := string(tx.TxOut[0].PkScript) // Simplified

	if sender != token.Owner {
		return fmt.Errorf("only current owner can transfer ownership")
	}

	// Transfer ownership
	oldOwner := token.Owner
	token.Owner = newOwner

	fmt.Printf("✓ Token ownership transferred: %s → %s for token %s\n", oldOwner, newOwner, token.Symbol)
	return nil
}

// validateSmartContractDeploy validates a smart contract deployment
func (b *BlockChain) validateSmartContractDeploy(tx *wire.MsgTx) error {
	// Basic validation: check memo contains contract code
	if len(tx.Memo) == 0 {
		return fmt.Errorf("smart contract deployment requires code in memo")
	}
	return nil
}

// validateSmartContractCall validates a smart contract call
func (b *BlockChain) validateSmartContractCall(tx *wire.MsgTx) error {
	// Basic validation: check memo contains call data
	if len(tx.Memo) == 0 {
		return fmt.Errorf("smart contract call requires data in memo")
	}
	return nil
}

// processTokenBurn processes a token burning transaction
func (b *BlockChain) processTokenBurn(tx *wire.MsgTx) error {
	// Parse token burn data from memo
	if len(tx.Memo) < 32 {
		return fmt.Errorf("token burn memo too short")
	}

	// Extract token ID (first 32 bytes)
	tokenID := wire.Hash{}
	copy(tokenID[:], tx.Memo[:32])

	// Parse burn data
	memoStr := string(tx.Memo[32:])
	parts := strings.Split(memoStr, "|")
	if len(parts) != 2 {
		return fmt.Errorf("invalid token burn memo format")
	}

	from := parts[0]
	amountStr := parts[1]

	amount, err := strconv.ParseInt(amountStr, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid burn amount: %v", err)
	}

	if amount <= 0 {
		return fmt.Errorf("burn amount must be positive")
	}

	// Check if token exists
	token, err := b.tokenStore.GetToken(tokenID)
	if err != nil {
		return fmt.Errorf("token does not exist: %v", err)
	}

	// Check sender has sufficient balance
	senderBalance := b.tokenStore.GetBalance(from, tokenID)
	if senderBalance < amount {
		return fmt.Errorf("insufficient token balance for burning: has %d, need %d", senderBalance, amount)
	}

	// Burn tokens (reduce balance and total supply)
	err = b.tokenStore.TransferToken(tokenID, from, "burn_address", amount)
	if err != nil {
		return fmt.Errorf("failed to burn tokens: %v", err)
	}

	// Update total supply
	token.Supply -= amount

	fmt.Printf("✓ Token burn: %d tokens burned from %s for token %s\n", amount, from, token.Symbol)
	return nil
}

// GetBalance returns the balance for a given address
func (b *BlockChain) GetBalance(address string) (int64, error) {
	// This is a simplified implementation that uses UTXO set
	// In production, you'd want proper address indexing
	return b.utxoSet.GetBalance(address)
}
