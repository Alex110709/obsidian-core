package crypto

import (
	"testing"
)

func TestGetAddressType(t *testing.T) {
	tests := []struct {
		name     string
		address  string
		expected AddressType
	}{
		{"Transparent address", "obs1abc123", AddressTypeTransparent},
		{"Shielded address", "zobs1abc123", AddressTypeShielded},
		{"Invalid short", "ob", AddressTypeUnknown},
		{"Invalid prefix", "xyz123", AddressTypeUnknown},
		{"Empty", "", AddressTypeUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetAddressType(tt.address)
			if result != tt.expected {
				t.Errorf("GetAddressType(%s) = %v, want %v", tt.address, result, tt.expected)
			}
		})
	}
}

func TestIsTransparentAddress(t *testing.T) {
	tests := []struct {
		name     string
		address  string
		expected bool
	}{
		{"Valid transparent", "obs1abc123", true},
		{"Shielded address", "zobs1abc123", false},
		{"Invalid", "xyz123", false},
		{"Empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsTransparentAddress(tt.address)
			if result != tt.expected {
				t.Errorf("IsTransparentAddress(%s) = %v, want %v", tt.address, result, tt.expected)
			}
		})
	}
}

func TestIsShieldedAddress(t *testing.T) {
	tests := []struct {
		name     string
		address  string
		expected bool
	}{
		{"Valid shielded", "zobs1abc123", true},
		{"Transparent address", "obs1abc123", false},
		{"Invalid", "xyz123", false},
		{"Empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsShieldedAddress(tt.address)
			if result != tt.expected {
				t.Errorf("IsShieldedAddress(%s) = %v, want %v", tt.address, result, tt.expected)
			}
		})
	}
}

func TestAddressTypeDetectionWithRealWallets(t *testing.T) {
	// Generate real wallets
	wallet1, err := GenerateSecureWallet()
	if err != nil {
		t.Fatalf("GenerateSecureWallet() error = %v", err)
	}

	wallet2, err := GenerateSecureWallet()
	if err != nil {
		t.Fatalf("GenerateSecureWallet() error = %v", err)
	}

	// Test transparent addresses
	if !IsTransparentAddress(wallet1.TransparentAddr) {
		t.Errorf("Failed to detect transparent address: %s", wallet1.TransparentAddr)
	}
	if !IsTransparentAddress(wallet2.TransparentAddr) {
		t.Errorf("Failed to detect transparent address: %s", wallet2.TransparentAddr)
	}

	// Test shielded addresses
	if !IsShieldedAddress(wallet1.ShieldedAddr) {
		t.Errorf("Failed to detect shielded address: %s", wallet1.ShieldedAddr)
	}
	if !IsShieldedAddress(wallet2.ShieldedAddr) {
		t.Errorf("Failed to detect shielded address: %s", wallet2.ShieldedAddr)
	}

	// Test cross-validation
	if IsShieldedAddress(wallet1.TransparentAddr) {
		t.Errorf("Transparent address incorrectly detected as shielded: %s", wallet1.TransparentAddr)
	}
	if IsTransparentAddress(wallet1.ShieldedAddr) {
		t.Errorf("Shielded address incorrectly detected as transparent: %s", wallet1.ShieldedAddr)
	}

	t.Logf("Wallet 1 - Transparent: %s, Shielded: %s", wallet1.TransparentAddr, wallet1.ShieldedAddr)
	t.Logf("Wallet 2 - Transparent: %s, Shielded: %s", wallet2.TransparentAddr, wallet2.ShieldedAddr)
}

func TestAddressTypeConstants(t *testing.T) {
	if AddressTypeUnknown == AddressTypeTransparent {
		t.Error("AddressTypeUnknown should not equal AddressTypeTransparent")
	}
	if AddressTypeUnknown == AddressTypeShielded {
		t.Error("AddressTypeUnknown should not equal AddressTypeShielded")
	}
	if AddressTypeTransparent == AddressTypeShielded {
		t.Error("AddressTypeTransparent should not equal AddressTypeShielded")
	}
}
