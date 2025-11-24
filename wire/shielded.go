package wire

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"

	"github.com/btcsuite/btcutil/base58"
)

// Shielded address constants
const (
	ShieldedAddressPrefix = "zobs" // z-address prefix for Obsidian
	ShieldedAddressLength = 95     // Base58 encoded length
	NullifierSize         = 32     // Nullifier size in bytes
	CommitmentSize        = 32     // Note commitment size
	ProofSize             = 192    // Simplified proof size (real zk-SNARK is larger)
)

// ShieldedAddress represents a shielded (z-address) in Obsidian
type ShieldedAddress struct {
	Prefix     string // "zobs"
	PublicKey  []byte // 32 bytes
	ViewingKey []byte // 32 bytes (for viewing encrypted txs)
}

// Note represents a shielded note (value commitment)
type Note struct {
	Value     int64  // Amount in satoshis
	Recipient []byte // Recipient public key
	Rcm       []byte // Randomness for commitment
	Memo      []byte // 512 bytes memo
}

// NoteCommitment represents a commitment to a note
type NoteCommitment struct {
	Cm []byte // Commitment (hash of note)
}

// Nullifier represents a unique nullifier to prevent double-spending
type Nullifier struct {
	Nf []byte // Nullifier (hash of note + secret)
}

// NewShieldedAddress generates a new shielded address
func NewShieldedAddress() (*ShieldedAddress, error) {
	// Generate random keys
	publicKey := make([]byte, 32)
	viewingKey := make([]byte, 32)

	if _, err := rand.Read(publicKey); err != nil {
		return nil, err
	}
	if _, err := rand.Read(viewingKey); err != nil {
		return nil, err
	}

	return &ShieldedAddress{
		Prefix:     ShieldedAddressPrefix,
		PublicKey:  publicKey,
		ViewingKey: viewingKey,
	}, nil
}

// String returns the base58-encoded shielded address
func (addr *ShieldedAddress) String() string {
	// Combine prefix + public key + viewing key
	data := append([]byte(addr.Prefix), addr.PublicKey...)
	data = append(data, addr.ViewingKey...)

	// Add checksum
	checksum := sha256.Sum256(data)
	data = append(data, checksum[:4]...)

	return base58.Encode(data)
}

// ParseShieldedAddress parses a base58-encoded shielded address
func ParseShieldedAddress(address string) (*ShieldedAddress, error) {
	decoded := base58.Decode(address)

	if len(decoded) < 68 { // 4 (prefix) + 32 (pk) + 32 (vk) + 4 (checksum)
		return nil, ErrShieldedAddress
	}

	// Verify checksum
	checksumData := decoded[:len(decoded)-4]
	providedChecksum := decoded[len(decoded)-4:]
	computedChecksum := sha256.Sum256(checksumData)

	for i := 0; i < 4; i++ {
		if providedChecksum[i] != computedChecksum[i] {
			return nil, ErrShieldedAddress
		}
	}

	return &ShieldedAddress{
		Prefix:     string(decoded[:4]),
		PublicKey:  decoded[4:36],
		ViewingKey: decoded[36:68],
	}, nil
}

// CreateNote creates a new shielded note
func CreateNote(value int64, recipient []byte, memo []byte) (*Note, error) {
	if len(memo) > 512 {
		return nil, ErrMemoTooLarge
	}

	// Pad memo to 512 bytes
	paddedMemo := make([]byte, 512)
	copy(paddedMemo, memo)

	// Generate random commitment randomness
	rcm := make([]byte, 32)
	if _, err := rand.Read(rcm); err != nil {
		return nil, err
	}

	return &Note{
		Value:     value,
		Recipient: recipient,
		Rcm:       rcm,
		Memo:      paddedMemo,
	}, nil
}

// Commit creates a commitment to the note
func (n *Note) Commit() *NoteCommitment {
	// Simplified commitment: hash(value || recipient || rcm)
	h := sha256.New()

	// Write value (8 bytes, little-endian)
	valueBytes := make([]byte, 8)
	for i := 0; i < 8; i++ {
		valueBytes[i] = byte(n.Value >> (i * 8))
	}

	h.Write(valueBytes)
	h.Write(n.Recipient)
	h.Write(n.Rcm)

	cm := h.Sum(nil)

	return &NoteCommitment{
		Cm: cm,
	}
}

// ComputeNullifier computes a nullifier for the note
func (n *Note) ComputeNullifier(secret []byte) *Nullifier {
	// Simplified nullifier: hash(note_commitment || secret)
	cm := n.Commit()

	h := sha256.New()
	h.Write(cm.Cm)
	h.Write(secret)

	nf := h.Sum(nil)

	return &Nullifier{
		Nf: nf,
	}
}

// EncryptNote encrypts a note for the recipient
func EncryptNote(note *Note, sharedSecret []byte) ([]byte, error) {
	// Serialize note
	plaintext := make([]byte, 0, 8+32+32+512)

	// Value (8 bytes)
	valueBytes := make([]byte, 8)
	for i := 0; i < 8; i++ {
		valueBytes[i] = byte(note.Value >> (i * 8))
	}
	plaintext = append(plaintext, valueBytes...)

	// Recipient (32 bytes)
	plaintext = append(plaintext, note.Recipient...)

	// Rcm (32 bytes)
	plaintext = append(plaintext, note.Rcm...)

	// Memo (512 bytes)
	plaintext = append(plaintext, note.Memo...)

	// Encrypt with AES-256-GCM
	block, err := aes.NewCipher(sharedSecret)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)

	return ciphertext, nil
}

// DecryptNote decrypts an encrypted note
func DecryptNote(ciphertext []byte, sharedSecret []byte) (*Note, error) {
	block, err := aes.NewCipher(sharedSecret)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	// Parse plaintext
	if len(plaintext) < 8+32+32+512 {
		return nil, fmt.Errorf("invalid plaintext length")
	}

	// Value
	value := int64(0)
	for i := 0; i < 8; i++ {
		value |= int64(plaintext[i]) << (i * 8)
	}

	// Recipient
	recipient := plaintext[8:40]

	// Rcm
	rcm := plaintext[40:72]

	// Memo
	memo := plaintext[72:584]

	return &Note{
		Value:     value,
		Recipient: recipient,
		Rcm:       rcm,
		Memo:      memo,
	}, nil
}

// GenerateProof generates a simplified zk-SNARK proof
// In production, this would use a real zk-SNARK library like bellman
func GenerateProof(note *Note, secret []byte) ([]byte, error) {
	// Simplified proof: hash(note || secret)
	// In real implementation, use proper zk-SNARK proving system

	h := sha256.New()

	// Value
	valueBytes := make([]byte, 8)
	for i := 0; i < 8; i++ {
		valueBytes[i] = byte(note.Value >> (i * 8))
	}
	h.Write(valueBytes)

	h.Write(note.Recipient)
	h.Write(note.Rcm)
	h.Write(secret)

	proof := make([]byte, ProofSize)
	hash := h.Sum(nil)

	// Repeat hash to fill proof size
	for i := 0; i < ProofSize; i++ {
		proof[i] = hash[i%32]
	}

	return proof, nil
}

// VerifyProof verifies a zk-SNARK proof
func VerifyProof(proof []byte, commitment []byte, nullifier []byte) bool {
	// Simplified verification
	// In production, use proper zk-SNARK verification

	if len(proof) != ProofSize {
		return false
	}

	if len(commitment) != CommitmentSize {
		return false
	}

	if len(nullifier) != NullifierSize {
		return false
	}

	// For demo purposes, always return true if sizes are correct
	// Real implementation would verify the cryptographic proof
	return true
}

// DeriveSharedSecret derives a shared secret for encryption
func DeriveSharedSecret(recipientPublicKey []byte, senderPrivateKey []byte) []byte {
	// Simplified ECDH (in production, use proper curve25519 or secp256k1)
	h := sha256.New()
	h.Write(recipientPublicKey)
	h.Write(senderPrivateKey)
	return h.Sum(nil)
}
