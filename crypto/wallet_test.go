package crypto

import (
	"strings"
	"testing"
)

func TestBase62Encoding(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
	}{
		{"Empty", []byte{}},
		{"Single byte", []byte{42}},
		{"Multiple bytes", []byte{1, 2, 3, 4, 5}},
		{"Large number", []byte{255, 255, 255, 255}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if len(tt.input) == 0 {
				return // Skip empty test
			}

			encoded := EncodeBase62(tt.input)
			decoded, err := DecodeBase62(encoded)
			if err != nil {
				t.Errorf("DecodeBase62() error = %v", err)
				return
			}

			// Compare values (decoded may have leading zeros stripped)
			if len(decoded) == 0 && len(tt.input) == 1 && tt.input[0] == 0 {
				return // Zero case is fine
			}

			// Strip leading zeros from input for comparison
			inputTrimmed := tt.input
			for len(inputTrimmed) > 0 && inputTrimmed[0] == 0 {
				inputTrimmed = inputTrimmed[1:]
			}

			if len(decoded) != len(inputTrimmed) {
				t.Errorf("Length mismatch: got %d, want %d", len(decoded), len(inputTrimmed))
				return
			}

			for i := range decoded {
				if decoded[i] != inputTrimmed[i] {
					t.Errorf("Byte mismatch at index %d: got %d, want %d", i, decoded[i], inputTrimmed[i])
				}
			}
		})
	}
}

func TestBase62Alphabet(t *testing.T) {
	// Test that encoding uses only alphanumeric characters
	testData := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	encoded := EncodeBase62(testData)

	for _, char := range encoded {
		isValid := (char >= '0' && char <= '9') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= 'a' && char <= 'z')

		if !isValid {
			t.Errorf("Invalid character in Base62 encoding: %c", char)
		}
	}

	t.Logf("Base62 encoded: %s", encoded)
}

func TestGenerateSecureWallet(t *testing.T) {
	wallet, err := GenerateSecureWallet()
	if err != nil {
		t.Fatalf("GenerateSecureWallet() error = %v", err)
	}

	// Verify mnemonic
	if wallet.Mnemonic == "" {
		t.Error("Mnemonic is empty")
	}

	words := strings.Split(wallet.Mnemonic, " ")
	if len(words) != 24 {
		t.Errorf("Expected 24 words in mnemonic, got %d", len(words))
	}

	t.Logf("Mnemonic: %s", wallet.Mnemonic)

	// Verify seed
	if len(wallet.Seed) == 0 {
		t.Error("Seed is empty")
	}

	// Verify keys
	if wallet.PrivateKey == nil {
		t.Error("Private key is nil")
	}
	if wallet.PublicKey == nil {
		t.Error("Public key is nil")
	}

	// Verify transparent address
	if wallet.TransparentAddr == "" {
		t.Error("Transparent address is empty")
	}
	if !strings.HasPrefix(wallet.TransparentAddr, "obs") {
		t.Errorf("Transparent address should start with 'obs', got %s", wallet.TransparentAddr)
	}

	t.Logf("Transparent Address: %s", wallet.TransparentAddr)

	// Verify shielded address
	if wallet.ShieldedAddr == "" {
		t.Error("Shielded address is empty")
	}
	if !strings.HasPrefix(wallet.ShieldedAddr, "zobs") {
		t.Errorf("Shielded address should start with 'zobs', got %s", wallet.ShieldedAddr)
	}

	t.Logf("Shielded Address: %s", wallet.ShieldedAddr)

	// Check that address contains only alphanumeric characters
	checkAlphanumeric := func(addr string, prefix string) {
		remainder := strings.TrimPrefix(addr, prefix)
		for _, char := range remainder {
			isValid := (char >= '0' && char <= '9') ||
				(char >= 'A' && char <= 'Z') ||
				(char >= 'a' && char <= 'z')

			if !isValid {
				t.Errorf("Address contains invalid character: %c in %s", char, addr)
			}
		}
	}

	checkAlphanumeric(wallet.TransparentAddr, "obs")
	checkAlphanumeric(wallet.ShieldedAddr, "zobs")
}

func TestRestoreSecureWallet(t *testing.T) {
	// First generate a wallet
	original, err := GenerateSecureWallet()
	if err != nil {
		t.Fatalf("GenerateSecureWallet() error = %v", err)
	}

	// Restore from mnemonic
	restored, err := RestoreSecureWallet(original.Mnemonic)
	if err != nil {
		t.Fatalf("RestoreSecureWallet() error = %v", err)
	}

	// Verify addresses match
	if restored.TransparentAddr != original.TransparentAddr {
		t.Errorf("Transparent addresses don't match:\nOriginal:  %s\nRestored:  %s",
			original.TransparentAddr, restored.TransparentAddr)
	}

	if restored.ShieldedAddr != original.ShieldedAddr {
		t.Errorf("Shielded addresses don't match:\nOriginal:  %s\nRestored:  %s",
			original.ShieldedAddr, restored.ShieldedAddr)
	}

	t.Logf("Successfully restored wallet with matching addresses")
}

func TestValidateAddress(t *testing.T) {
	// Generate a valid wallet
	wallet, err := GenerateSecureWallet()
	if err != nil {
		t.Fatalf("GenerateSecureWallet() error = %v", err)
	}

	tests := []struct {
		name    string
		address string
		valid   bool
	}{
		{"Valid transparent", wallet.TransparentAddr, true},
		{"Valid shielded", wallet.ShieldedAddr, true},
		{"Invalid prefix", "xyz123456", false},
		{"Empty", "", false},
		{"Too short", "obs", false},
		{"Invalid characters", "obs!@#$%", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateAddress(tt.address)
			if result != tt.valid {
				t.Errorf("ValidateAddress(%s) = %v, want %v", tt.address, result, tt.valid)
			}
		})
	}
}

func TestMultipleWalletGeneration(t *testing.T) {
	// Generate multiple wallets and ensure they're unique
	wallets := make([]*SecureWallet, 5)
	addresses := make(map[string]bool)

	for i := 0; i < 5; i++ {
		wallet, err := GenerateSecureWallet()
		if err != nil {
			t.Fatalf("GenerateSecureWallet() #%d error = %v", i, err)
		}

		wallets[i] = wallet

		// Check uniqueness
		if addresses[wallet.TransparentAddr] {
			t.Errorf("Duplicate transparent address generated: %s", wallet.TransparentAddr)
		}
		addresses[wallet.TransparentAddr] = true

		if addresses[wallet.ShieldedAddr] {
			t.Errorf("Duplicate shielded address generated: %s", wallet.ShieldedAddr)
		}
		addresses[wallet.ShieldedAddr] = true

		t.Logf("Wallet #%d:", i+1)
		t.Logf("  Transparent: %s", wallet.TransparentAddr)
		t.Logf("  Shielded: %s", wallet.ShieldedAddr)
	}
}

func TestInvalidMnemonicRestore(t *testing.T) {
	invalidMnemonics := []string{
		"",
		"invalid mnemonic phrase",
		"word1 word2 word3",
		"abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon wrong",
	}

	for _, mnemonic := range invalidMnemonics {
		_, err := RestoreSecureWallet(mnemonic)
		if err == nil {
			t.Errorf("RestoreSecureWallet() should fail for invalid mnemonic: %s", mnemonic)
		}
	}
}
