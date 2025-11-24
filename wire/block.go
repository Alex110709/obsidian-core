package wire

import (
	"bytes"
	"encoding/gob"
	"time"
)

// BlockVersion is the current block version.
const BlockVersion = 1

// BlockHeader defines information about a block and is used in the bitcoin
// block (MsgBlock) and headers (MsgHeaders) messages.
type BlockHeader struct {
	// Version of the block.  This is not the same as the protocol version.
	Version int32

	// Hash of the previous block header in the block chain.
	PrevBlock Hash

	// MerkleTreeHash is the double sha256 hash of all of the transaction
	// hashes in the block.
	MerkleRoot Hash

	// Timestamp the block was created.  This is, unfortunately, encoded as a
	// uint32 in the wire protocol which limits its range.
	Timestamp time.Time

	// Difficulty target for the block.
	Bits uint32

	// Nonce used to generate the block.
	Nonce uint32

	// DarkMatter solution bytes (Obsidian specific)
	// Contains the nonce and other proof data for the AES-SHA256 hybrid PoW.
	DarkMatterSolution []byte
}

// MsgBlock implements the Message interface and represents a bitcoin
// block message.  It is used to deliver block and transaction information in
// response to a getdata message (MsgGetData) for a given block hash.
type MsgBlock struct {
	Header       BlockHeader
	Transactions []*MsgTx
}

// AddTransaction adds a transaction to the message.
func (msg *MsgBlock) AddTransaction(tx *MsgTx) error {
	msg.Transactions = append(msg.Transactions, tx)
	return nil
}

// NewMsgBlock returns a new bitcoin block message that conforms to the
// Message interface.  The return instance has a default header version of
// BlockVersion and there are no transactions.
func NewMsgBlock(blockHeader *BlockHeader) *MsgBlock {
	return &MsgBlock{
		Header:       *blockHeader,
		Transactions: make([]*MsgTx, 0, 64),
	}
}

// BlockHash calculates the hash of the block header.
func (h *BlockHeader) BlockHash() Hash {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	enc.Encode(h)
	return DoubleHashH(buf.Bytes())
}

// BlockHash returns the hash of the block header.
func (msg *MsgBlock) BlockHash() Hash {
	return msg.Header.BlockHash()
}
