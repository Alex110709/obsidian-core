package blockchain

import (
	"fmt"
	"obsidian-core/wire"
	"sync"
	"time"
)

const (
	// MaxMempoolSize is the maximum number of transactions in the mempool
	MaxMempoolSize = 10000

	// MaxOrphanTxs is the maximum number of orphan transactions
	MaxOrphanTxs = 100

	// OrphanTxExpiry is the time after which orphan txs are removed
	OrphanTxExpiry = 20 * time.Minute
)

// TxDesc represents a transaction in the mempool
type TxDesc struct {
	Tx       *wire.MsgTx
	Added    time.Time
	Height   int32
	Fee      int64
	FeePerKB int64
}

// Mempool represents the transaction memory pool
type Mempool struct {
	mu sync.RWMutex

	// Pool of transactions
	pool map[wire.Hash]*TxDesc

	// Orphan transactions (transactions whose inputs reference unknown transactions)
	orphans map[wire.Hash]*TxDesc

	// Index of transactions by address
	outpoints map[wire.OutPoint]wire.Hash

	// Maximum size
	maxSize int
}

// NewMempool creates a new mempool
func NewMempool() *Mempool {
	return &Mempool{
		pool:      make(map[wire.Hash]*TxDesc),
		orphans:   make(map[wire.Hash]*TxDesc),
		outpoints: make(map[wire.OutPoint]wire.Hash),
		maxSize:   MaxMempoolSize,
	}
}

// AddTransaction adds a transaction to the mempool
func (m *Mempool) AddTransaction(tx *wire.MsgTx, height int32, fee int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if mempool is full
	if len(m.pool) >= m.maxSize {
		return fmt.Errorf("mempool is full")
	}

	txHash := tx.TxHash()

	// Check if transaction already exists
	if _, exists := m.pool[txHash]; exists {
		return fmt.Errorf("transaction already in mempool")
	}

	// Create transaction descriptor
	txDesc := &TxDesc{
		Tx:       tx,
		Added:    time.Now(),
		Height:   height,
		Fee:      fee,
		FeePerKB: calculateFeePerKB(tx, fee),
	}

	// Add to pool
	m.pool[txHash] = txDesc

	// Index outpoints
	for _, txIn := range tx.TxIn {
		m.outpoints[txIn.PreviousOutPoint] = txHash
	}

	return nil
}

// RemoveTransaction removes a transaction from the mempool
func (m *Mempool) RemoveTransaction(txHash wire.Hash) {
	m.mu.Lock()
	defer m.mu.Unlock()

	txDesc, exists := m.pool[txHash]
	if !exists {
		return
	}

	// Remove outpoint indexes
	for _, txIn := range txDesc.Tx.TxIn {
		delete(m.outpoints, txIn.PreviousOutPoint)
	}

	// Remove from pool
	delete(m.pool, txHash)
}

// GetTransaction retrieves a transaction from the mempool
func (m *Mempool) GetTransaction(txHash wire.Hash) (*wire.MsgTx, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	txDesc, exists := m.pool[txHash]
	if !exists {
		return nil, fmt.Errorf("transaction not found in mempool")
	}

	return txDesc.Tx, nil
}

// HasTransaction checks if a transaction exists in the mempool
func (m *Mempool) HasTransaction(txHash wire.Hash) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	_, exists := m.pool[txHash]
	return exists
}

// GetTransactions returns all transactions in the mempool
func (m *Mempool) GetTransactions() []*wire.MsgTx {
	m.mu.RLock()
	defer m.mu.RUnlock()

	txs := make([]*wire.MsgTx, 0, len(m.pool))
	for _, txDesc := range m.pool {
		txs = append(txs, txDesc.Tx)
	}

	return txs
}

// GetTransactionsByPriority returns transactions sorted by fee priority
func (m *Mempool) GetTransactionsByPriority(limit int) []*wire.MsgTx {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Create slice of transaction descriptors
	txDescs := make([]*TxDesc, 0, len(m.pool))
	for _, txDesc := range m.pool {
		txDescs = append(txDescs, txDesc)
	}

	// Sort by fee per KB (descending)
	for i := 0; i < len(txDescs)-1; i++ {
		for j := i + 1; j < len(txDescs); j++ {
			if txDescs[i].FeePerKB < txDescs[j].FeePerKB {
				txDescs[i], txDescs[j] = txDescs[j], txDescs[i]
			}
		}
	}

	// Return top N transactions
	count := limit
	if count > len(txDescs) {
		count = len(txDescs)
	}

	txs := make([]*wire.MsgTx, count)
	for i := 0; i < count; i++ {
		txs[i] = txDescs[i].Tx
	}

	return txs
}

// IsSpent checks if an outpoint is spent by a transaction in the mempool
func (m *Mempool) IsSpent(outpoint wire.OutPoint) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	_, exists := m.outpoints[outpoint]
	return exists
}

// RemoveDoubleSpends removes transactions that spend the same inputs
func (m *Mempool) RemoveDoubleSpends(tx *wire.MsgTx) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, txIn := range tx.TxIn {
		if conflictHash, exists := m.outpoints[txIn.PreviousOutPoint]; exists {
			m.removeTransactionLocked(conflictHash)
		}
	}
}

// removeTransactionLocked removes a transaction without acquiring the lock
func (m *Mempool) removeTransactionLocked(txHash wire.Hash) {
	txDesc, exists := m.pool[txHash]
	if !exists {
		return
	}

	// Remove outpoint indexes
	for _, txIn := range txDesc.Tx.TxIn {
		delete(m.outpoints, txIn.PreviousOutPoint)
	}

	// Remove from pool
	delete(m.pool, txHash)
}

// AddOrphan adds an orphan transaction
func (m *Mempool) AddOrphan(tx *wire.MsgTx) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.orphans) >= MaxOrphanTxs {
		// Remove oldest orphan
		var oldestHash wire.Hash
		var oldestTime time.Time
		for hash, desc := range m.orphans {
			if oldestTime.IsZero() || desc.Added.Before(oldestTime) {
				oldestHash = hash
				oldestTime = desc.Added
			}
		}
		delete(m.orphans, oldestHash)
	}

	txHash := tx.TxHash()
	m.orphans[txHash] = &TxDesc{
		Tx:    tx,
		Added: time.Now(),
	}
}

// ProcessOrphans processes orphan transactions that may now be valid
func (m *Mempool) ProcessOrphans(utxoSet *UTXOSet) []*wire.MsgTx {
	m.mu.Lock()
	defer m.mu.Unlock()

	var processedTxs []*wire.MsgTx

	for hash, desc := range m.orphans {
		// Check if all inputs are now available
		allInputsAvailable := true
		for _, txIn := range desc.Tx.TxIn {
			_, err := utxoSet.GetUTXO(txIn.PreviousOutPoint.Hash, txIn.PreviousOutPoint.Index)
			if err != nil {
				allInputsAvailable = false
				break
			}
		}

		if allInputsAvailable {
			// Move from orphans to pool
			delete(m.orphans, hash)
			m.pool[hash] = desc
			processedTxs = append(processedTxs, desc.Tx)
		}
	}

	return processedTxs
}

// RemoveExpiredOrphans removes orphan transactions that have expired
func (m *Mempool) RemoveExpiredOrphans() {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	for hash, desc := range m.orphans {
		if now.Sub(desc.Added) > OrphanTxExpiry {
			delete(m.orphans, hash)
		}
	}
}

// Count returns the number of transactions in the mempool
func (m *Mempool) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.pool)
}

// OrphanCount returns the number of orphan transactions
func (m *Mempool) OrphanCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.orphans)
}

// Reset clears the mempool
func (m *Mempool) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.pool = make(map[wire.Hash]*TxDesc)
	m.orphans = make(map[wire.Hash]*TxDesc)
	m.outpoints = make(map[wire.OutPoint]wire.Hash)
}

// Helper functions

func calculateFeePerKB(tx *wire.MsgTx, fee int64) int64 {
	// Estimate transaction size (simplified)
	size := estimateTxSize(tx)
	if size == 0 {
		return 0
	}

	return (fee * 1000) / int64(size)
}

func estimateTxSize(tx *wire.MsgTx) int {
	// Simplified size estimation
	// Version (4) + input count (1) + output count (1) + locktime (4)
	size := 10

	// Inputs: outpoint (36) + script length (1-9) + script + sequence (4)
	for _, txIn := range tx.TxIn {
		size += 36 + 1 + len(txIn.SignatureScript) + 4
	}

	// Outputs: value (8) + script length (1-9) + script
	for _, txOut := range tx.TxOut {
		size += 8 + 1 + len(txOut.PkScript)
	}

	return size
}
