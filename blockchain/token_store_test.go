package blockchain

import (
	"obsidian-core/wire"
	"testing"
)

func TestTokenStoreIssue(t *testing.T) {
	ts := NewTokenStore(nil)

	tokenID := wire.Hash{}
	copy(tokenID[:], []byte("test-token-id-1234567890123456"))

	err := ts.IssueToken(tokenID, "Test Token", "TEST", 8, 1000000, "owner-address")
	if err != nil {
		t.Fatalf("Failed to issue token: %v", err)
	}

	// Verify token exists
	token, err := ts.GetToken(tokenID)
	if err != nil {
		t.Fatalf("Failed to get token: %v", err)
	}

	if token.Name != "Test Token" {
		t.Errorf("Expected token name 'Test Token', got '%s'", token.Name)
	}

	// Verify owner balance
	balance := ts.GetBalance("owner-address", tokenID)
	if balance != 1000000 {
		t.Errorf("Expected balance 1000000, got %d", balance)
	}
}

func TestTokenStoreIssueDuplicate(t *testing.T) {
	ts := NewTokenStore(nil)

	tokenID := wire.Hash{}
	copy(tokenID[:], []byte("test-token-id-1234567890123456"))

	err := ts.IssueToken(tokenID, "Test Token", "TEST", 8, 1000000, "owner-address")
	if err != nil {
		t.Fatalf("Failed to issue token: %v", err)
	}

	// Try to issue duplicate symbol
	tokenID2 := wire.Hash{}
	copy(tokenID2[:], []byte("test-token-id-2234567890123456"))
	err = ts.IssueToken(tokenID2, "Test Token 2", "TEST", 8, 1000000, "owner-address")
	if err == nil {
		t.Fatal("Expected error when issuing duplicate symbol, got nil")
	}
}

func TestTokenStoreTransfer(t *testing.T) {
	ts := NewTokenStore(nil)

	tokenID := wire.Hash{}
	copy(tokenID[:], []byte("test-token-id-1234567890123456"))

	err := ts.IssueToken(tokenID, "Test Token", "TEST", 8, 1000000, "owner-address")
	if err != nil {
		t.Fatalf("Failed to issue token: %v", err)
	}

	// Transfer tokens
	err = ts.TransferToken(tokenID, "owner-address", "recipient-address", 50000)
	if err != nil {
		t.Fatalf("Failed to transfer token: %v", err)
	}

	// Verify balances
	ownerBalance := ts.GetBalance("owner-address", tokenID)
	if ownerBalance != 950000 {
		t.Errorf("Expected owner balance 950000, got %d", ownerBalance)
	}

	recipientBalance := ts.GetBalance("recipient-address", tokenID)
	if recipientBalance != 50000 {
		t.Errorf("Expected recipient balance 50000, got %d", recipientBalance)
	}
}

func TestTokenStoreTransferInsufficientBalance(t *testing.T) {
	ts := NewTokenStore(nil)

	tokenID := wire.Hash{}
	copy(tokenID[:], []byte("test-token-id-1234567890123456"))

	err := ts.IssueToken(tokenID, "Test Token", "TEST", 8, 1000000, "owner-address")
	if err != nil {
		t.Fatalf("Failed to issue token: %v", err)
	}

	// Try to transfer more than balance
	err = ts.TransferToken(tokenID, "owner-address", "recipient-address", 2000000)
	if err == nil {
		t.Fatal("Expected error when transferring more than balance, got nil")
	}
}

func TestTokenStoreTransferNegativeAmount(t *testing.T) {
	ts := NewTokenStore(nil)

	tokenID := wire.Hash{}
	copy(tokenID[:], []byte("test-token-id-1234567890123456"))

	err := ts.IssueToken(tokenID, "Test Token", "TEST", 8, 1000000, "owner-address")
	if err != nil {
		t.Fatalf("Failed to issue token: %v", err)
	}

	// Try to transfer negative amount
	err = ts.TransferToken(tokenID, "owner-address", "recipient-address", -100)
	if err == nil {
		t.Fatal("Expected error when transferring negative amount, got nil")
	}
}

func TestTokenStoreIssueNegativeSupply(t *testing.T) {
	ts := NewTokenStore(nil)

	tokenID := wire.Hash{}
	copy(tokenID[:], []byte("test-token-id-1234567890123456"))

	// Try to issue token with negative supply
	err := ts.IssueToken(tokenID, "Test Token", "TEST", 8, -1000000, "owner-address")
	if err == nil {
		t.Fatal("Expected error when issuing token with negative supply, got nil")
	}
}

func TestTokenStoreIssueZeroSupply(t *testing.T) {
	ts := NewTokenStore(nil)

	tokenID := wire.Hash{}
	copy(tokenID[:], []byte("test-token-id-1234567890123456"))

	// Try to issue token with zero supply
	err := ts.IssueToken(tokenID, "Test Token", "TEST", 8, 0, "owner-address")
	if err == nil {
		t.Fatal("Expected error when issuing token with zero supply, got nil")
	}
}

func TestTokenStoreTransferToSameAddress(t *testing.T) {
	ts := NewTokenStore(nil)

	tokenID := wire.Hash{}
	copy(tokenID[:], []byte("test-token-id-1234567890123456"))

	err := ts.IssueToken(tokenID, "Test Token", "TEST", 8, 1000000, "owner-address")
	if err != nil {
		t.Fatalf("Failed to issue token: %v", err)
	}

	// Try to transfer to same address
	err = ts.TransferToken(tokenID, "owner-address", "owner-address", 100)
	if err == nil {
		t.Fatal("Expected error when transferring to same address, got nil")
	}
}
