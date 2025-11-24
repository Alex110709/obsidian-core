package blockchain

import (
	"encoding/binary"
	"fmt"
	"obsidian-core/wire"
	"sync"

	bolt "go.etcd.io/bbolt"
)

var (
	utxoBucketName = []byte("utxo")
)

// UTXO represents an unspent transaction output
type UTXO struct {
	TxHash   wire.Hash
	Index    uint32
	Value    int64
	PkScript []byte
	Height   int32
}

// UTXOSet represents the UTXO set
type UTXOSet struct {
	db *bolt.DB
	mu sync.RWMutex
}

// NewUTXOSet creates a new UTXO set
func NewUTXOSet(db *bolt.DB) *UTXOSet {
	return &UTXOSet{
		db: db,
	}
}

// AddUTXO adds a UTXO to the set
func (u *UTXOSet) AddUTXO(txHash wire.Hash, index uint32, value int64, pkScript []byte, height int32) error {
	u.mu.Lock()
	defer u.mu.Unlock()

	return u.db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(utxoBucketName)
		if err != nil {
			return err
		}

		key := makeUTXOKey(txHash, index)
		utxo := UTXO{
			TxHash:   txHash,
			Index:    index,
			Value:    value,
			PkScript: pkScript,
			Height:   height,
		}

		data, err := serializeUTXO(&utxo)
		if err != nil {
			return err
		}

		return bucket.Put(key, data)
	})
}

// RemoveUTXO removes a UTXO from the set
func (u *UTXOSet) RemoveUTXO(txHash wire.Hash, index uint32) error {
	u.mu.Lock()
	defer u.mu.Unlock()

	return u.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(utxoBucketName)
		if bucket == nil {
			return fmt.Errorf("utxo bucket not found")
		}

		key := makeUTXOKey(txHash, index)
		return bucket.Delete(key)
	})
}

// GetUTXO retrieves a UTXO
func (u *UTXOSet) GetUTXO(txHash wire.Hash, index uint32) (*UTXO, error) {
	u.mu.RLock()
	defer u.mu.RUnlock()

	var utxo *UTXO
	err := u.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(utxoBucketName)
		if bucket == nil {
			return fmt.Errorf("utxo bucket not found")
		}

		key := makeUTXOKey(txHash, index)
		data := bucket.Get(key)
		if data == nil {
			return fmt.Errorf("utxo not found")
		}

		var err error
		utxo, err = deserializeUTXO(data)
		return err
	})

	return utxo, err
}

// GetUTXOsForAddress returns all UTXOs for a given address
func (u *UTXOSet) GetUTXOsForAddress(address string) ([]*UTXO, error) {
	u.mu.RLock()
	defer u.mu.RUnlock()

	var utxos []*UTXO

	err := u.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(utxoBucketName)
		if bucket == nil {
			return nil // No UTXOs yet
		}

		return bucket.ForEach(func(k, v []byte) error {
			utxo, err := deserializeUTXO(v)
			if err != nil {
				return err
			}

			// Check if this UTXO belongs to the address
			// This is simplified - in production you'd properly decode the script
			if string(utxo.PkScript) == address {
				utxos = append(utxos, utxo)
			}

			return nil
		})
	})

	return utxos, err
}

// GetBalance returns the balance for an address
func (u *UTXOSet) GetBalance(address string) (int64, error) {
	utxos, err := u.GetUTXOsForAddress(address)
	if err != nil {
		return 0, err
	}

	var balance int64
	for _, utxo := range utxos {
		balance += utxo.Value
	}

	return balance, nil
}

// ApplyBlock applies a block's transactions to the UTXO set
func (u *UTXOSet) ApplyBlock(block *wire.MsgBlock, height int32) error {
	u.mu.Lock()
	defer u.mu.Unlock()

	return u.db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(utxoBucketName)
		if err != nil {
			return err
		}

		for _, msgTx := range block.Transactions {
			txHash := msgTx.TxHash()

			// Skip coinbase inputs (they don't spend UTXOs)
			if !msgTx.IsCoinbase() {
				// Remove spent UTXOs
				for _, txIn := range msgTx.TxIn {
					key := makeUTXOKey(txIn.PreviousOutPoint.Hash, txIn.PreviousOutPoint.Index)
					if err := bucket.Delete(key); err != nil {
						return err
					}
				}
			}

			// Add new UTXOs
			for i, txOut := range msgTx.TxOut {
				utxo := UTXO{
					TxHash:   txHash,
					Index:    uint32(i),
					Value:    txOut.Value,
					PkScript: txOut.PkScript,
					Height:   height,
				}

				data, err := serializeUTXO(&utxo)
				if err != nil {
					return err
				}

				key := makeUTXOKey(txHash, uint32(i))
				if err := bucket.Put(key, data); err != nil {
					return err
				}
			}
		}

		return nil
	})
}

// RollbackBlock removes a block's effects from the UTXO set
func (u *UTXOSet) RollbackBlock(block *wire.MsgBlock) error {
	u.mu.Lock()
	defer u.mu.Unlock()

	return u.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(utxoBucketName)
		if bucket == nil {
			return fmt.Errorf("utxo bucket not found")
		}

		// Process in reverse order
		for i := len(block.Transactions) - 1; i >= 0; i-- {
			msgTx := block.Transactions[i]
			txHash := msgTx.TxHash()

			// Remove UTXOs created by this transaction
			for j := range msgTx.TxOut {
				key := makeUTXOKey(txHash, uint32(j))
				if err := bucket.Delete(key); err != nil {
					return err
				}
			}

			// Restore spent UTXOs (if not coinbase)
			// Note: This requires storing spent UTXOs, which we'll implement later
			if !msgTx.IsCoinbase() {
				// TODO: Restore previously spent UTXOs
			}
		}

		return nil
	})
}

// Helper functions

func makeUTXOKey(txHash wire.Hash, index uint32) []byte {
	key := make([]byte, 32+4)
	copy(key[:32], txHash[:])
	binary.LittleEndian.PutUint32(key[32:], index)
	return key
}

func serializeUTXO(utxo *UTXO) ([]byte, error) {
	// Simple serialization: txHash(32) + index(4) + value(8) + height(4) + scriptLen(2) + script
	scriptLen := len(utxo.PkScript)
	data := make([]byte, 32+4+8+4+2+scriptLen)

	offset := 0
	copy(data[offset:], utxo.TxHash[:])
	offset += 32

	binary.LittleEndian.PutUint32(data[offset:], utxo.Index)
	offset += 4

	binary.LittleEndian.PutUint64(data[offset:], uint64(utxo.Value))
	offset += 8

	binary.LittleEndian.PutUint32(data[offset:], uint32(utxo.Height))
	offset += 4

	binary.LittleEndian.PutUint16(data[offset:], uint16(scriptLen))
	offset += 2

	copy(data[offset:], utxo.PkScript)

	return data, nil
}

func deserializeUTXO(data []byte) (*UTXO, error) {
	if len(data) < 32+4+8+4+2 {
		return nil, fmt.Errorf("invalid utxo data")
	}

	utxo := &UTXO{}
	offset := 0

	copy(utxo.TxHash[:], data[offset:offset+32])
	offset += 32

	utxo.Index = binary.LittleEndian.Uint32(data[offset:])
	offset += 4

	utxo.Value = int64(binary.LittleEndian.Uint64(data[offset:]))
	offset += 8

	utxo.Height = int32(binary.LittleEndian.Uint32(data[offset:]))
	offset += 4

	scriptLen := binary.LittleEndian.Uint16(data[offset:])
	offset += 2

	utxo.PkScript = make([]byte, scriptLen)
	copy(utxo.PkScript, data[offset:])

	return utxo, nil
}
