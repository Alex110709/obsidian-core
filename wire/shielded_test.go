package wire

import (
	"bytes"
	"testing"
)

func TestNewShieldedAddress(t *testing.T) {
	addr, err := NewShieldedAddress()
	if err != nil {
		t.Fatalf("Failed to generate shielded address: %v", err)
	}

	if addr.Prefix != ShieldedAddressPrefix {
		t.Errorf("Expected prefix %s, got %s", ShieldedAddressPrefix, addr.Prefix)
	}

	if len(addr.PublicKey) != 32 {
		t.Errorf("Expected public key length 32, got %d", len(addr.PublicKey))
	}

	if len(addr.ViewingKey) != 32 {
		t.Errorf("Expected viewing key length 32, got %d", len(addr.ViewingKey))
	}
}

func TestShieldedAddressString(t *testing.T) {
	addr, err := NewShieldedAddress()
	if err != nil {
		t.Fatalf("Failed to generate shielded address: %v", err)
	}

	addrStr := addr.String()
	if addrStr == "" {
		t.Error("Address string is empty")
	}

	t.Logf("Generated address: %s", addrStr)
}

func TestParseShieldedAddress(t *testing.T) {
	// Generate address
	addr1, err := NewShieldedAddress()
	if err != nil {
		t.Fatalf("Failed to generate shielded address: %v", err)
	}

	// Convert to string
	addrStr := addr1.String()

	// Parse back
	addr2, err := ParseShieldedAddress(addrStr)
	if err != nil {
		t.Fatalf("Failed to parse shielded address: %v", err)
	}

	// Compare
	if addr2.Prefix != addr1.Prefix {
		t.Errorf("Prefix mismatch: %s != %s", addr2.Prefix, addr1.Prefix)
	}

	if !bytes.Equal(addr2.PublicKey, addr1.PublicKey) {
		t.Error("Public key mismatch")
	}

	if !bytes.Equal(addr2.ViewingKey, addr1.ViewingKey) {
		t.Error("Viewing key mismatch")
	}
}

func TestCreateNote(t *testing.T) {
	value := int64(100000000) // 1 OBS
	recipient := make([]byte, 32)
	memo := []byte("Test payment")

	note, err := CreateNote(value, recipient, memo)
	if err != nil {
		t.Fatalf("Failed to create note: %v", err)
	}

	if note.Value != value {
		t.Errorf("Expected value %d, got %d", value, note.Value)
	}

	if !bytes.Equal(note.Recipient, recipient) {
		t.Error("Recipient mismatch")
	}

	if len(note.Memo) != 512 {
		t.Errorf("Expected memo length 512, got %d", len(note.Memo))
	}

	if !bytes.Equal(note.Memo[:len(memo)], memo) {
		t.Error("Memo mismatch")
	}
}

func TestNoteCommitment(t *testing.T) {
	note, err := CreateNote(100000000, make([]byte, 32), []byte("test"))
	if err != nil {
		t.Fatalf("Failed to create note: %v", err)
	}

	cm := note.Commit()
	if len(cm.Cm) != 32 {
		t.Errorf("Expected commitment length 32, got %d", len(cm.Cm))
	}

	// Same note should produce same commitment
	cm2 := note.Commit()
	if !bytes.Equal(cm.Cm, cm2.Cm) {
		t.Error("Same note produced different commitments")
	}
}

func TestNoteNullifier(t *testing.T) {
	note, err := CreateNote(100000000, make([]byte, 32), []byte("test"))
	if err != nil {
		t.Fatalf("Failed to create note: %v", err)
	}

	secret := make([]byte, 32)
	secret[0] = 1

	nf := note.ComputeNullifier(secret)
	if len(nf.Nf) != 32 {
		t.Errorf("Expected nullifier length 32, got %d", len(nf.Nf))
	}

	// Same secret should produce same nullifier
	nf2 := note.ComputeNullifier(secret)
	if !bytes.Equal(nf.Nf, nf2.Nf) {
		t.Error("Same secret produced different nullifiers")
	}

	// Different secret should produce different nullifier
	secret2 := make([]byte, 32)
	secret2[0] = 2
	nf3 := note.ComputeNullifier(secret2)
	if bytes.Equal(nf.Nf, nf3.Nf) {
		t.Error("Different secrets produced same nullifier")
	}
}

func TestEncryptDecryptNote(t *testing.T) {
	value := int64(100000000)
	recipient := make([]byte, 32)
	copy(recipient, []byte("test_recipient_key"))
	memo := []byte("Secret message for testing encryption")

	note, err := CreateNote(value, recipient, memo)
	if err != nil {
		t.Fatalf("Failed to create note: %v", err)
	}

	sharedSecret := make([]byte, 32)
	copy(sharedSecret, []byte("shared_secret_key_for_testing"))

	// Encrypt
	ciphertext, err := EncryptNote(note, sharedSecret)
	if err != nil {
		t.Fatalf("Failed to encrypt note: %v", err)
	}

	if len(ciphertext) == 0 {
		t.Error("Ciphertext is empty")
	}

	// Decrypt
	decryptedNote, err := DecryptNote(ciphertext, sharedSecret)
	if err != nil {
		t.Fatalf("Failed to decrypt note: %v", err)
	}

	// Verify decrypted values
	if decryptedNote.Value != value {
		t.Errorf("Value mismatch: expected %d, got %d", value, decryptedNote.Value)
	}

	if !bytes.Equal(decryptedNote.Recipient, recipient) {
		t.Error("Recipient mismatch after decryption")
	}

	if !bytes.Equal(decryptedNote.Memo[:len(memo)], memo) {
		t.Errorf("Memo mismatch: expected %s, got %s", memo, decryptedNote.Memo[:len(memo)])
	}
}

func TestGenerateProof(t *testing.T) {
	note, err := CreateNote(100000000, make([]byte, 32), []byte("test"))
	if err != nil {
		t.Fatalf("Failed to create note: %v", err)
	}

	secret := make([]byte, 32)
	proof, err := GenerateProof(note, secret)
	if err != nil {
		t.Fatalf("Failed to generate proof: %v", err)
	}

	if len(proof) != ProofSize {
		t.Errorf("Expected proof size %d, got %d", ProofSize, len(proof))
	}
}

func TestVerifyProof(t *testing.T) {
	note, err := CreateNote(100000000, make([]byte, 32), []byte("test"))
	if err != nil {
		t.Fatalf("Failed to create note: %v", err)
	}

	secret := make([]byte, 32)
	proof, err := GenerateProof(note, secret)
	if err != nil {
		t.Fatalf("Failed to generate proof: %v", err)
	}

	cm := note.Commit()
	nf := note.ComputeNullifier(secret)

	// Verify proof
	valid := VerifyProof(proof, cm.Cm, nf.Nf)
	if !valid {
		t.Error("Valid proof was rejected")
	}

	// Invalid proof (wrong size)
	invalidProof := make([]byte, 10)
	valid = VerifyProof(invalidProof, cm.Cm, nf.Nf)
	if valid {
		t.Error("Invalid proof was accepted")
	}
}

func TestMemoTooLarge(t *testing.T) {
	largeMemo := make([]byte, 600) // Larger than 512
	_, err := CreateNote(100, make([]byte, 32), largeMemo)
	if err != ErrMemoTooLarge {
		t.Errorf("Expected ErrMemoTooLarge, got %v", err)
	}
}
