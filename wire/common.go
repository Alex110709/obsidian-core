package wire

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
)

// Error definitions
var (
	ErrMemoTooLarge      = errors.New("memo exceeds 512 bytes")
	ErrInvalidProof      = errors.New("invalid zk-SNARK proof")
	ErrInvalidNullifier  = errors.New("invalid or duplicate nullifier")
	ErrValueBalance      = errors.New("value balance does not match")
	ErrInvalidCommitment = errors.New("invalid note commitment")
	ErrShieldedAddress   = errors.New("invalid shielded address")
)

// HashSize of array used to store hashes.  See Hash.
const HashSize = 32

// Hash is used in several of the bitcoin messages and common structures.  It
// typically represents the double sha256 of data.
type Hash [HashSize]byte

// String returns the Hash as the hexadecimal string of the byte-reversed
// hash.
func (hash Hash) String() string {
	for i := 0; i < HashSize/2; i++ {
		hash[i], hash[HashSize-1-i] = hash[HashSize-1-i], hash[i]
	}
	return hex.EncodeToString(hash[:])
}

// NewHashFromStr creates a Hash from a hash string.  The string should be
// the hexadecimal string of a byte-reversed hash, but any missing characters
// result in zero padding at the end of the Hash.
func NewHashFromStr(hash string) (*Hash, error) {
	ret := new(Hash)
	if len(hash) > HashSize*2 {
		return nil, fmt.Errorf("hash string too long")
	}

	// TODO: Implement proper decoding
	return ret, nil
}

// DoubleHashH calculates hash(hash(b)) and returns the resulting bytes as a
// Hash.
func DoubleHashH(b []byte) Hash {
	first := sha256.Sum256(b)
	return Hash(sha256.Sum256(first[:]))
}
