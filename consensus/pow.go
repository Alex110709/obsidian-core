package consensus

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/binary"
	"math/big"
	"obsidian-core/wire"
)

// PowEngine defines the interface for a Proof-of-Work algorithm.
type PowEngine interface {
	// Verify checks if the block header satisfies the PoW target.
	Verify(header *wire.BlockHeader) bool

	// Solve attempts to find a nonce that satisfies the PoW target.
	// Returns the nonce and the solution bytes (e.g. IV).
	Solve(header *wire.BlockHeader) (nonce uint32, solution []byte, found bool)

	// SolveWithLimit attempts to find a nonce with a custom attempt limit.
	SolveWithLimit(header *wire.BlockHeader, maxAttempts uint32) (nonce uint32, solution []byte, found bool)
}

// DarkMatter is a custom PoW algorithm for Obsidian.
// It uses a combination of SHA-256 and AES encryption to create a memory-hard
// and computation-heavy proof.
type DarkMatter struct {
	// Difficulty adjustment parameters could go here
}

func NewDarkMatter() *DarkMatter {
	return &DarkMatter{}
}

// Verify checks if the block header's hash meets the difficulty target
// using the DarkMatter algorithm.
func (d *DarkMatter) Verify(header *wire.BlockHeader) bool {
	// 1. Calculate DarkMatter hash
	hash := DarkMatterHash(header)

	// 2. Convert bits to target
	target := CompactToBig(header.Bits)

	// 3. Convert hash to big.Int
	hashNum := new(big.Int).SetBytes(hash)

	// 4. Check if hash <= target
	return hashNum.Cmp(target) <= 0
}

// Solve attempts to find a nonce.
func (d *DarkMatter) Solve(header *wire.BlockHeader) (uint32, []byte, bool) {
	return d.SolveWithLimit(header, 1000000)
}

// SolveWithLimit attempts to find a nonce with a custom limit.
func (d *DarkMatter) SolveWithLimit(header *wire.BlockHeader, maxAttempts uint32) (uint32, []byte, bool) {
	target := CompactToBig(header.Bits)

	// Try nonces up to the limit
	for nonce := uint32(0); nonce < maxAttempts; nonce++ {
		header.Nonce = nonce

		// Generate IV for this attempt
		iv := make([]byte, 16)
		binary.LittleEndian.PutUint32(iv[0:4], nonce)

		// Calculate hash
		hash := DarkMatterHash(header)
		hashNum := new(big.Int).SetBytes(hash)

		if hashNum.Cmp(target) <= 0 {
			return nonce, iv, true
		}
	}

	// Not found within limit
	return 0, nil, false
}

// DarkMatterHash calculates the PoW hash from a block header.
func DarkMatterHash(header *wire.BlockHeader) []byte {
	// Serialize header
	var headerBytes []byte

	// Version (4 bytes)
	vers := make([]byte, 4)
	binary.LittleEndian.PutUint32(vers, uint32(header.Version))
	headerBytes = append(headerBytes, vers...)

	// PrevBlock (32 bytes)
	headerBytes = append(headerBytes, header.PrevBlock[:]...)

	// MerkleRoot (32 bytes)
	headerBytes = append(headerBytes, header.MerkleRoot[:]...)

	// Timestamp (8 bytes)
	ts := make([]byte, 8)
	binary.LittleEndian.PutUint64(ts, uint64(header.Timestamp.Unix()))
	headerBytes = append(headerBytes, ts...)

	// Bits (4 bytes)
	bits := make([]byte, 4)
	binary.LittleEndian.PutUint32(bits, header.Bits)
	headerBytes = append(headerBytes, bits...)

	// Nonce (4 bytes)
	nonce := make([]byte, 4)
	binary.LittleEndian.PutUint32(nonce, header.Nonce)
	headerBytes = append(headerBytes, nonce...)

	return DarkMatterHashBytes(headerBytes)
}

// DarkMatterHashBytes calculates the PoW hash.
// Concept: SHA256(Data) -> AES Encrypt (using hash as key) -> SHA256(Ciphertext)
func DarkMatterHashBytes(data []byte) []byte {
	// 1. Initial Hash
	h1 := sha256.Sum256(data)

	// 2. AES Encryption Step (Memory hardness simulation)
	block, _ := aes.NewCipher(h1[:])

	ciphertext := make([]byte, len(h1))
	// CRITICAL: IV must be deterministic for verification to work
	// Derive IV from the first hash to ensure same input = same output
	iv := make([]byte, aes.BlockSize)
	copy(iv, h1[:aes.BlockSize])

	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(ciphertext, h1[:])

	// 3. Final Hash
	h2 := sha256.Sum256(ciphertext)
	return h2[:]
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

// BigToCompact converts a big.Int to compact representation
func BigToCompact(n *big.Int) uint32 {
	if n.Sign() == 0 {
		return 0
	}

	// Get the minimum number of bytes to represent the value
	bytes := n.Bytes()
	size := uint32(len(bytes))

	// Extract mantissa (top 3 bytes)
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

	// Set the exponent
	compact |= size << 24

	// Set the sign bit if negative
	if n.Sign() < 0 {
		compact |= 0x00800000
	}

	return compact
}
