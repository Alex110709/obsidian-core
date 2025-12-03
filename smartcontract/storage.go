// Storage system for smart contract state
package smartcontract

import (
	"encoding/json"
	"fmt"
	"obsidian-core/database"
)

// ContractStorage manages persistent storage for contracts
type ContractStorage struct {
	db *database.Storage
}

// NewContractStorage creates a new contract storage
func NewContractStorage(db *database.Storage) *ContractStorage {
	return &ContractStorage{db: db}
}

// Store sets a key-value pair for a contract
func (cs *ContractStorage) Store(contractAddr, key string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	bucket := []byte("contracts")
	storageKey := []byte(fmt.Sprintf("%s_%s", contractAddr, key))
	return cs.db.Put(bucket, storageKey, data)
}

// Load retrieves a value for a contract
func (cs *ContractStorage) Load(contractAddr, key string) (interface{}, error) {
	bucket := []byte("contracts")
	storageKey := []byte(fmt.Sprintf("%s_%s", contractAddr, key))
	data, err := cs.db.Get(bucket, storageKey)
	if err != nil {
		return nil, err
	}

	var value interface{}
	err = json.Unmarshal(data, &value)
	return value, err
}

// Delete removes a key-value pair
func (cs *ContractStorage) Delete(contractAddr, key string) error {
	bucket := []byte("contracts")
	storageKey := []byte(fmt.Sprintf("%s_%s", contractAddr, key))
	return cs.db.Delete(bucket, storageKey)
}
