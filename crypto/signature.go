package crypto

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"encoding/asn1"
	"fmt"
	"math/big"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcutil/base58"
	"github.com/tyler-smith/go-bip32"
	"github.com/tyler-smith/go-bip39"
	"golang.org/x/crypto/ripemd160"
)

// Signature represents an ECDSA signature
type Signature struct {
	R, S *big.Int
}

// GenerateKeyPair generates a new ECDSA key pair using secp256k1 curve
func GenerateKeyPair() (*ecdsa.PrivateKey, *ecdsa.PublicKey, error) {
	// Use secp256k1 curve (Bitcoin/Ethereum standard)
	privateKey, err := btcec.NewPrivateKey()
	if err != nil {
		return nil, nil, err
	}

	return privateKey.ToECDSA(), &privateKey.ToECDSA().PublicKey, nil
}

// Sign creates a signature for the given hash using the private key
func Sign(privateKey *ecdsa.PrivateKey, hash []byte) ([]byte, error) {
	r, s, err := ecdsa.Sign(rand.Reader, privateKey, hash)
	if err != nil {
		return nil, err
	}

	// Encode signature as DER
	return asn1.Marshal(Signature{R: r, S: s})
}

// Verify verifies a signature against a hash and public key
func Verify(publicKey *ecdsa.PublicKey, hash, signature []byte) bool {
	var sig Signature
	_, err := asn1.Unmarshal(signature, &sig)
	if err != nil {
		return false
	}

	return ecdsa.Verify(publicKey, hash, sig.R, sig.S)
}

// PublicKeyToBytes converts a public key to bytes
func PublicKeyToBytes(pubKey *ecdsa.PublicKey) []byte {
	// Compressed format: 0x02/0x03 + X coordinate
	x := pubKey.X.Bytes()
	prefix := byte(0x02)
	if pubKey.Y.Bit(0) == 1 {
		prefix = 0x03
	}

	// Pad X to 32 bytes
	paddedX := make([]byte, 33)
	paddedX[0] = prefix
	copy(paddedX[33-len(x):], x)

	return paddedX
}

// BytesToPublicKey converts bytes to a public key
func BytesToPublicKey(pubKeyBytes []byte) (*ecdsa.PublicKey, error) {
	if len(pubKeyBytes) != 33 {
		return nil, fmt.Errorf("invalid public key length")
	}

	pubKey, err := btcec.ParsePubKey(pubKeyBytes)
	if err != nil {
		return nil, err
	}

	return pubKey.ToECDSA(), nil
}

// Hash256 performs double SHA256 hash
func Hash256(data []byte) []byte {
	first := sha256.Sum256(data)
	second := sha256.Sum256(first[:])
	return second[:]
}

// Hash160 performs SHA256 followed by RIPEMD160
func Hash160(data []byte) []byte {
	hash := sha256.Sum256(data)
	ripemd := ripemd160.New()
	ripemd.Write(hash[:])
	return ripemd.Sum(nil)
}

// PrivateKeyToWIF converts a private key to Wallet Import Format
func PrivateKeyToWIF(privateKey *ecdsa.PrivateKey) string {
	// Simplified WIF encoding
	// In production, use proper base58check encoding
	d := privateKey.D.Bytes()
	return fmt.Sprintf("WIF_%x", d)
}

// WIFToPrivateKey converts WIF to private key
func WIFToPrivateKey(wif string) (*ecdsa.PrivateKey, error) {
	// Simplified WIF decoding
	// In production, use proper base58check decoding
	return nil, fmt.Errorf("WIF decoding not implemented")
}

// KeyToAddress generates an address from a public key
func KeyToAddress(pubKey *ecdsa.PublicKey) string {
	pubKeyBytes := PublicKeyToBytes(pubKey)
	hash := Hash160(pubKeyBytes)
	// Add version byte (0x00 for mainnet)
	versionedHash := append([]byte{0x00}, hash...)
	// Double hash for checksum
	checksum := Hash256(versionedHash)[:4]
	fullHash := append(versionedHash, checksum...)
	return "obs" + base58.Encode(fullHash)
}

// GenerateShieldedAddress generates a shielded address (simplified as zobs prefix)
func GenerateShieldedAddress(pubKey *ecdsa.PublicKey) string {
	transparentAddr := KeyToAddress(pubKey)
	// Simplified shielded address: zobs + transparent address
	return "zobs" + transparentAddr
}

// GenerateMnemonic generates a new BIP39 mnemonic (24 words)
func GenerateMnemonic() (string, error) {
	entropy, err := bip39.NewEntropy(256) // 24 words
	if err != nil {
		return "", err
	}
	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return "", err
	}
	return mnemonic, nil
}

// MnemonicToSeed converts mnemonic to seed
func MnemonicToSeed(mnemonic string) ([]byte, error) {
	return bip39.NewSeedWithErrorChecking(mnemonic, "")
}

// SeedToKeyPair derives ECDSA keypair from seed using BIP32/BIP44 path
func SeedToKeyPair(seed []byte) (*ecdsa.PrivateKey, *ecdsa.PublicKey, error) {
	masterKey, err := bip32.NewMasterKey(seed)
	if err != nil {
		return nil, nil, err
	}

	// BIP44 path: m/44'/0'/0'/0/0 for simplicity
	path := []uint32{44 + bip32.FirstHardenedChild, 0 + bip32.FirstHardenedChild, 0 + bip32.FirstHardenedChild, 0, 0}

	childKey := masterKey
	for _, childNum := range path {
		childKey, err = childKey.NewChildKey(childNum)
		if err != nil {
			return nil, nil, err
		}
	}

	privateKeyBytes := childKey.Key

	// Convert to ECDSA using secp256k1
	privKey, pubKey := btcec.PrivKeyFromBytes(privateKeyBytes)

	return privKey.ToECDSA(), pubKey.ToECDSA(), nil
}
