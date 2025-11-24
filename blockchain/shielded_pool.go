package blockchain

import (
	"bytes"
	"fmt"
	"obsidian-core/wire"
	"sync"
)

// ShieldedPool manages the shielded transaction pool
// This includes tracking note commitments and nullifiers
type ShieldedPool struct {
	mu sync.RWMutex

	// Set of all note commitments (unspent shielded outputs)
	commitments map[string]*wire.NoteCommitment

	// Set of all nullifiers (spent shielded outputs)
	nullifiers map[string]*wire.Nullifier

	// Merkle tree of commitments (simplified)
	commitmentTree [][]byte

	// Total shielded value in pool
	totalShieldedValue int64
}

// NewShieldedPool creates a new shielded pool
func NewShieldedPool() *ShieldedPool {
	return &ShieldedPool{
		commitments:    make(map[string]*wire.NoteCommitment),
		nullifiers:     make(map[string]*wire.Nullifier),
		commitmentTree: make([][]byte, 0),
	}
}

// AddCommitment adds a note commitment to the pool
func (sp *ShieldedPool) AddCommitment(cm *wire.NoteCommitment, value int64) error {
	sp.mu.Lock()
	defer sp.mu.Unlock()

	key := string(cm.Cm)

	// Check if commitment already exists
	if _, exists := sp.commitments[key]; exists {
		return fmt.Errorf("commitment already exists")
	}

	// Add to commitment map
	sp.commitments[key] = cm

	// Add to Merkle tree
	sp.commitmentTree = append(sp.commitmentTree, cm.Cm)

	// Update total value
	sp.totalShieldedValue += value

	return nil
}

// AddNullifier adds a nullifier to the pool (marks note as spent)
func (sp *ShieldedPool) AddNullifier(nf *wire.Nullifier) error {
	sp.mu.Lock()
	defer sp.mu.Unlock()

	key := string(nf.Nf)

	// Check if nullifier already exists (double-spend protection)
	if _, exists := sp.nullifiers[key]; exists {
		return wire.ErrInvalidNullifier
	}

	// Add to nullifier map
	sp.nullifiers[key] = nf

	return nil
}

// HasNullifier checks if a nullifier exists (is spent)
func (sp *ShieldedPool) HasNullifier(nf []byte) bool {
	sp.mu.RLock()
	defer sp.mu.RUnlock()

	_, exists := sp.nullifiers[string(nf)]
	return exists
}

// HasCommitment checks if a commitment exists
func (sp *ShieldedPool) HasCommitment(cm []byte) bool {
	sp.mu.RLock()
	defer sp.mu.RUnlock()

	_, exists := sp.commitments[string(cm)]
	return exists
}

// GetMerkleRoot returns the root of the commitment Merkle tree
func (sp *ShieldedPool) GetMerkleRoot() []byte {
	sp.mu.RLock()
	defer sp.mu.RUnlock()

	if len(sp.commitmentTree) == 0 {
		return make([]byte, 32) // Empty root
	}

	// Simplified Merkle root calculation
	return sp.commitmentTree[len(sp.commitmentTree)-1]
}

// GetTotalShieldedValue returns the total value in the shielded pool
func (sp *ShieldedPool) GetTotalShieldedValue() int64 {
	sp.mu.RLock()
	defer sp.mu.RUnlock()

	return sp.totalShieldedValue
}

// ValidateShieldedTransaction validates a shielded transaction
func (sp *ShieldedPool) ValidateShieldedTransaction(tx *wire.MsgTx) error {
	if !tx.IsShielded() {
		return nil // Not a shielded transaction
	}

	// 1. Verify all nullifiers are unique (no double-spends)
	for _, spend := range tx.ShieldedSpends {
		if sp.HasNullifier(spend.Nullifier) {
			return wire.ErrInvalidNullifier
		}
	}

	// 2. Verify all proofs
	for _, spend := range tx.ShieldedSpends {
		if !wire.VerifyProof(spend.Proof, spend.Anchor, spend.Nullifier) {
			return wire.ErrInvalidProof
		}
	}

	for _, output := range tx.ShieldedOutputs {
		if !wire.VerifyProof(output.Proof, output.Cmu, nil) {
			return wire.ErrInvalidProof
		}
	}

	// 3. Verify value balance
	if err := sp.validateValueBalance(tx); err != nil {
		return err
	}

	return nil
}

// validateValueBalance ensures the value balance equation holds
func (sp *ShieldedPool) validateValueBalance(tx *wire.MsgTx) error {
	// Calculate transparent outputs
	transparentOut := int64(0)
	for _, txOut := range tx.TxOut {
		transparentOut += txOut.Value
	}

	// Value balance equation:
	// transparent_in - transparent_out = value_balance + fees
	// Where value_balance = shielded_out - shielded_in

	// For now, just check that value balance is within reasonable bounds
	if tx.ValueBalance < -1e8 || tx.ValueBalance > 1e8 {
		return wire.ErrValueBalance
	}

	return nil
}

// ProcessShieldedTransaction processes a shielded transaction
func (sp *ShieldedPool) ProcessShieldedTransaction(tx *wire.MsgTx) error {
	if !tx.IsShielded() {
		return nil
	}

	// Validate transaction
	if err := sp.ValidateShieldedTransaction(tx); err != nil {
		return err
	}

	// Add nullifiers (mark inputs as spent)
	for _, spend := range tx.ShieldedSpends {
		nf := &wire.Nullifier{Nf: spend.Nullifier}
		if err := sp.AddNullifier(nf); err != nil {
			return err
		}
	}

	// Add commitments (create new shielded outputs)
	for _, output := range tx.ShieldedOutputs {
		cm := &wire.NoteCommitment{Cm: output.Cmu}
		// Value is encrypted, so we can't know it directly
		// In production, this would be tracked differently
		if err := sp.AddCommitment(cm, 0); err != nil {
			return err
		}
	}

	return nil
}

// GetShieldedBalance calculates the shielded balance for a viewing key
// This requires scanning all commitments and trying to decrypt them
func (sp *ShieldedPool) GetShieldedBalance(viewingKey []byte) (int64, error) {
	sp.mu.RLock()
	defer sp.mu.RUnlock()

	balance := int64(0)

	// For each commitment, try to decrypt and check if it belongs to this viewing key
	// This is a simplified implementation - real Zcash uses more sophisticated scanning

	for cmKey, cm := range sp.commitments {
		// Check if this commitment has been spent
		// (if its nullifier exists in the nullifier set)
		spent := false
		for nfKey := range sp.nullifiers {
			// Simplified check - in production, derive proper nullifier
			if bytes.Equal([]byte(cmKey), []byte(nfKey)) {
				spent = true
				break
			}
		}

		if !spent {
			// Try to decrypt this note
			// In production, use proper decryption with viewing key
			_ = cm

			// For demo, just add a fixed amount
			// balance += decryptedValue
		}
	}

	return balance, nil
}

// Stats returns statistics about the shielded pool
func (sp *ShieldedPool) Stats() map[string]interface{} {
	sp.mu.RLock()
	defer sp.mu.RUnlock()

	return map[string]interface{}{
		"total_commitments":    len(sp.commitments),
		"total_nullifiers":     len(sp.nullifiers),
		"total_shielded_value": sp.totalShieldedValue,
		"merkle_tree_size":     len(sp.commitmentTree),
	}
}
