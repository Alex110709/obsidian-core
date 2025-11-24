package wire

import (
	"bytes"
	"encoding/gob"
	"fmt"
)

// CompactBlock represents a compact block (BIP 152 style).
// Instead of full transactions, it sends short transaction IDs.
type CompactBlock struct {
	Header       BlockHeader
	Nonce        uint64
	ShortIDs     []uint64 // Short transaction IDs (first 6 bytes of txid)
	PrefilledTxs []PrefilledTransaction
}

// PrefilledTransaction represents a transaction that must be included in full.
type PrefilledTransaction struct {
	Index int // Index in block
	Tx    *MsgTx
}

// NewCompactBlock creates a compact block from a full block.
func NewCompactBlock(block *MsgBlock, nonce uint64) *CompactBlock {
	cb := &CompactBlock{
		Header:       block.Header,
		Nonce:        nonce,
		ShortIDs:     make([]uint64, 0, len(block.Transactions)),
		PrefilledTxs: make([]PrefilledTransaction, 0),
	}

	for i, tx := range block.Transactions {
		// Coinbase is always included in full
		if i == 0 {
			cb.PrefilledTxs = append(cb.PrefilledTxs, PrefilledTransaction{
				Index: i,
				Tx:    tx,
			})
			continue
		}

		// Generate short ID for transaction
		txHash := tx.TxHash()
		shortID := computeShortID(txHash, nonce)
		cb.ShortIDs = append(cb.ShortIDs, shortID)
	}

	return cb
}

// computeShortID computes a short transaction ID.
// Uses first 6 bytes of hash(txid || nonce).
func computeShortID(txHash Hash, nonce uint64) uint64 {
	// Combine txHash and nonce
	data := make([]byte, 32+8)
	copy(data[:32], txHash[:])

	// Add nonce (little-endian)
	for i := 0; i < 8; i++ {
		data[32+i] = byte(nonce >> (i * 8))
	}

	// Hash it
	hash := DoubleHashH(data)

	// Return first 6 bytes as uint64
	shortID := uint64(0)
	for i := 0; i < 6; i++ {
		shortID |= uint64(hash[i]) << (i * 8)
	}

	return shortID
}

// ReconstructBlock attempts to reconstruct a full block from a compact block.
// Returns the full block and a list of missing transaction indices.
func (cb *CompactBlock) ReconstructBlock(mempool map[Hash]*MsgTx) (*MsgBlock, []int, error) {
	block := NewMsgBlock(&cb.Header)
	missing := make([]int, 0)

	// Build transaction index from mempool
	mempoolShortIDs := make(map[uint64]*MsgTx)
	for txHash, tx := range mempool {
		shortID := computeShortID(txHash, cb.Nonce)
		mempoolShortIDs[shortID] = tx
	}

	// Track prefilled transaction indices
	prefilledMap := make(map[int]*MsgTx)
	for _, pf := range cb.PrefilledTxs {
		prefilledMap[pf.Index] = pf.Tx
	}

	// Reconstruct transactions
	currentTxIndex := 0
	shortIDIndex := 0

	for currentTxIndex < len(cb.ShortIDs)+len(cb.PrefilledTxs) {
		// Check if this index has a prefilled transaction
		if prefilledTx, exists := prefilledMap[currentTxIndex]; exists {
			block.AddTransaction(prefilledTx)
		} else {
			// Look up short ID in mempool
			if shortIDIndex >= len(cb.ShortIDs) {
				return nil, nil, fmt.Errorf("short ID index out of bounds")
			}

			shortID := cb.ShortIDs[shortIDIndex]
			shortIDIndex++

			if tx, exists := mempoolShortIDs[shortID]; exists {
				block.AddTransaction(tx)
			} else {
				// Missing transaction - request it
				missing = append(missing, currentTxIndex)
				// Add placeholder nil (will be replaced)
				block.AddTransaction(&MsgTx{})
			}
		}

		currentTxIndex++
	}

	return block, missing, nil
}

// Encode encodes the compact block.
func (cb *CompactBlock) Encode() ([]byte, error) {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	if err := encoder.Encode(cb); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// DecodeCompactBlock decodes a compact block.
func DecodeCompactBlock(data []byte) (*CompactBlock, error) {
	buf := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buf)
	cb := &CompactBlock{}
	if err := decoder.Decode(cb); err != nil {
		return nil, err
	}
	return cb, nil
}

// BlockTxRequest requests missing transactions from a compact block.
type BlockTxRequest struct {
	BlockHash Hash
	Indices   []int // Indices of missing transactions
}

// BlockTxResponse contains the requested transactions.
type BlockTxResponse struct {
	BlockHash    Hash
	Transactions []*MsgTx
}
