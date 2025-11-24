package database

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"obsidian-core/wire"
	"os"
	"path/filepath"

	"go.etcd.io/bbolt"
)

const (
	defaultDbFile = "obsidian.db"
	blocksBucket  = "blocks"
)

type Storage struct {
	db *bbolt.DB
}

func NewStorage() (*Storage, error) {
	// Use DATA_DIR environment variable or default to current directory
	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = "."
	}

	// Create data directory if it doesn't exist
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %v", err)
	}

	dbFile := filepath.Join(dataDir, defaultDbFile)
	db, err := bbolt.Open(dbFile, 0600, nil)
	if err != nil {
		return nil, err
	}

	err = db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(blocksBucket))
		return err
	})
	if err != nil {
		return nil, err
	}

	return &Storage{db: db}, nil
}

func (s *Storage) Close() {
	s.db.Close()
}

// DB returns the underlying bolt database
func (s *Storage) DB() *bbolt.DB {
	return s.db
}

func (s *Storage) SaveBlock(block *wire.MsgBlock) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))

		// Serialize block
		var buf bytes.Buffer
		enc := gob.NewEncoder(&buf)
		if err := enc.Encode(block); err != nil {
			return err
		}

		// Key: Block Hash
		blockHash := block.BlockHash()
		key := blockHash[:]

		return b.Put(key, buf.Bytes())
	})
}

func (s *Storage) GetBlock(hash []byte) (*wire.MsgBlock, error) {
	var block wire.MsgBlock

	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		data := b.Get(hash)
		if data == nil {
			return fmt.Errorf("block not found")
		}

		dec := gob.NewDecoder(bytes.NewReader(data))
		return dec.Decode(&block)
	})

	return &block, err
}
