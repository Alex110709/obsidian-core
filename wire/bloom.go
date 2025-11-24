package wire

import (
	"crypto/sha256"
	"math"
)

// BloomFilter implements a bloom filter for SPV clients (BIP 37).
type BloomFilter struct {
	filter   []byte // Bit array
	hashFunc uint32 // Number of hash functions
	tweak    uint32 // Random tweak for hash functions
	flags    uint8  // Filter flags
}

const (
	// MaxBloomFilterSize is the maximum size of a bloom filter
	MaxBloomFilterSize = 36000 // bytes

	// MaxHashFuncs is the maximum number of hash functions
	MaxHashFuncs = 50

	// BloomUpdateNone indicates no auto-update
	BloomUpdateNone = 0

	// BloomUpdateAll indicates update on all matches
	BloomUpdateAll = 1

	// BloomUpdateP2PKOnly indicates update only on P2PK/P2PKH
	BloomUpdateP2PKOnly = 2
)

// NewBloomFilter creates a new bloom filter.
func NewBloomFilter(numElements uint32, falsePositiveRate float64, tweak uint32, flags uint8) *BloomFilter {
	// Calculate optimal filter size
	// m = -n * ln(p) / (ln(2)^2)
	size := uint32(-1.0 * float64(numElements) * math.Log(falsePositiveRate) / (math.Ln2 * math.Ln2))

	// Limit size
	if size > MaxBloomFilterSize*8 {
		size = MaxBloomFilterSize * 8
	}

	// Round up to nearest byte
	sizeBytes := (size + 7) / 8

	// Calculate optimal number of hash functions
	// k = (m/n) * ln(2)
	hashFuncs := uint32(float64(size) / float64(numElements) * math.Ln2)

	// Limit hash functions
	if hashFuncs > MaxHashFuncs {
		hashFuncs = MaxHashFuncs
	}
	if hashFuncs == 0 {
		hashFuncs = 1
	}

	return &BloomFilter{
		filter:   make([]byte, sizeBytes),
		hashFunc: hashFuncs,
		tweak:    tweak,
		flags:    flags,
	}
}

// Add adds data to the bloom filter.
func (bf *BloomFilter) Add(data []byte) {
	for i := uint32(0); i < bf.hashFunc; i++ {
		hash := bf.hash(i, data)
		index := hash % uint32(len(bf.filter)*8)
		bf.filter[index/8] |= 1 << (index % 8)
	}
}

// Contains checks if data might be in the filter.
// Returns true if possibly in set, false if definitely not in set.
func (bf *BloomFilter) Contains(data []byte) bool {
	for i := uint32(0); i < bf.hashFunc; i++ {
		hash := bf.hash(i, data)
		index := hash % uint32(len(bf.filter)*8)
		if (bf.filter[index/8] & (1 << (index % 8))) == 0 {
			return false // Definitely not in set
		}
	}
	return true // Possibly in set
}

// hash computes a hash of data with the given index.
func (bf *BloomFilter) hash(hashIndex uint32, data []byte) uint32 {
	// MurmurHash3-like hash (simplified)
	h := sha256.New()

	// Write seed (hashIndex * 0xFBA4C795 + tweak)
	seed := hashIndex*0xFBA4C795 + bf.tweak
	seedBytes := []byte{
		byte(seed),
		byte(seed >> 8),
		byte(seed >> 16),
		byte(seed >> 24),
	}
	h.Write(seedBytes)
	h.Write(data)

	hashResult := h.Sum(nil)

	// Return first 4 bytes as uint32
	return uint32(hashResult[0]) |
		uint32(hashResult[1])<<8 |
		uint32(hashResult[2])<<16 |
		uint32(hashResult[3])<<24
}

// Matches checks if a transaction matches the bloom filter.
func (bf *BloomFilter) MatchesTx(tx *MsgTx) bool {
	// Check transaction hash
	txHash := tx.TxHash()
	if bf.Contains(txHash[:]) {
		return true
	}

	// Check outputs
	for _, out := range tx.TxOut {
		// Check if output script matches
		if bf.Contains(out.PkScript) {
			return true
		}

		// Check if output script contains any data that matches
		// (simplified - real implementation would parse script)
	}

	// Check inputs
	for _, in := range tx.TxIn {
		// Check previous outpoint
		outpointData := append(in.PreviousOutPoint.Hash[:], byte(in.PreviousOutPoint.Index))
		if bf.Contains(outpointData) {
			return true
		}

		// Check signature script
		if bf.Contains(in.SignatureScript) {
			return true
		}
	}

	return false
}

// MerkleBlock represents a filtered block for SPV clients.
type MerkleBlock struct {
	Header       BlockHeader
	TxCount      uint32
	Hashes       []Hash
	Flags        []byte
	Transactions []*MsgTx // Matched transactions
}

// NewMerkleBlock creates a merkle block from a full block and bloom filter.
func NewMerkleBlock(block *MsgBlock, filter *BloomFilter) *MerkleBlock {
	mb := &MerkleBlock{
		Header:       block.Header,
		TxCount:      uint32(len(block.Transactions)),
		Hashes:       make([]Hash, 0),
		Flags:        make([]byte, 0),
		Transactions: make([]*MsgTx, 0),
	}

	// Build merkle tree with matched transactions
	matches := make([]bool, len(block.Transactions))
	hashes := make([]Hash, len(block.Transactions))

	for i, tx := range block.Transactions {
		hashes[i] = tx.TxHash()
		if filter.MatchesTx(tx) {
			matches[i] = true
			mb.Transactions = append(mb.Transactions, tx)
		}
	}

	// Build partial merkle tree
	mb.buildPartialTree(hashes, matches)

	return mb
}

// buildPartialTree builds a partial merkle tree for SPV proof.
func (mb *MerkleBlock) buildPartialTree(hashes []Hash, matches []bool) {
	// Simplified implementation - real version would build proper merkle tree
	// For now, just include all hashes of matched transactions
	for i, match := range matches {
		if match {
			mb.Hashes = append(mb.Hashes, hashes[i])
			mb.Flags = append(mb.Flags, 1)
		} else {
			mb.Flags = append(mb.Flags, 0)
		}
	}
}

// FilterLoad message for loading a bloom filter on a peer.
type FilterLoadMsg struct {
	Filter    []byte
	HashFuncs uint32
	Tweak     uint32
	Flags     uint8
}

// FilterAdd message for adding data to a bloom filter.
type FilterAddMsg struct {
	Data []byte
}

// FilterClear message for clearing a bloom filter.
type FilterClearMsg struct{}
