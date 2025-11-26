package blockchain

import (
	"fmt"
	"obsidian-core/database"
	"obsidian-core/wire"
	"sync"
)

// TokenStore manages token information and balances
type TokenStore struct {
	db         *database.Storage
	tokens     map[wire.Hash]*wire.Token      // tokenID -> token info
	balances   map[string]map[wire.Hash]int64 // address -> tokenID -> balance
	tokenIndex map[string]wire.Hash           // symbol -> tokenID
	mutex      sync.RWMutex
}

// NewTokenStore creates a new token store
func NewTokenStore(db *database.Storage) *TokenStore {
	return &TokenStore{
		db:         db,
		tokens:     make(map[wire.Hash]*wire.Token),
		balances:   make(map[string]map[wire.Hash]int64),
		tokenIndex: make(map[string]wire.Hash),
	}
}

// IssueToken creates a new token
func (ts *TokenStore) IssueToken(tokenID wire.Hash, name, symbol string, decimals uint8, supply int64, owner string) error {
	ts.mutex.Lock()
	defer ts.mutex.Unlock()

	// Input validation
	if supply < 0 {
		return fmt.Errorf("supply cannot be negative")
	}
	if supply == 0 {
		return fmt.Errorf("supply must be greater than zero")
	}
	if len(name) == 0 || len(name) > 32 {
		return fmt.Errorf("token name length must be between 1 and 32 characters")
	}
	if len(symbol) == 0 || len(symbol) > 8 {
		return fmt.Errorf("token symbol length must be between 1 and 8 characters")
	}
	if decimals > 18 {
		return fmt.Errorf("decimals cannot exceed 18")
	}
	if len(owner) == 0 {
		return fmt.Errorf("owner address cannot be empty")
	}

	// Check if symbol already exists
	if _, exists := ts.tokenIndex[symbol]; exists {
		return fmt.Errorf("token symbol %s already exists", symbol)
	}

	// Check if tokenID already exists
	if _, exists := ts.tokens[tokenID]; exists {
		return fmt.Errorf("token ID already exists")
	}

	token := &wire.Token{
		ID:          tokenID,
		Name:        name,
		Symbol:      symbol,
		Decimals:    decimals,
		TotalSupply: supply,
		Owner:       owner,
		Created:     0, // Will be set by caller
	}

	ts.tokens[tokenID] = token
	ts.tokenIndex[symbol] = tokenID

	// Initialize owner balance
	if ts.balances[owner] == nil {
		ts.balances[owner] = make(map[wire.Hash]int64)
	}
	ts.balances[owner][tokenID] = supply

	return nil
}

// TransferToken transfers tokens between addresses
func (ts *TokenStore) TransferToken(tokenID wire.Hash, from, to string, amount int64) error {
	ts.mutex.Lock()
	defer ts.mutex.Unlock()

	// Input validation
	if amount <= 0 {
		return fmt.Errorf("transfer amount must be positive")
	}
	if from == "" {
		return fmt.Errorf("from address cannot be empty")
	}
	if to == "" {
		return fmt.Errorf("to address cannot be empty")
	}
	if from == to {
		return fmt.Errorf("cannot transfer to same address")
	}

	// Check if token exists
	if _, exists := ts.tokens[tokenID]; !exists {
		return fmt.Errorf("token does not exist")
	}

	fromBalance := ts.getBalance(from, tokenID)
	if fromBalance < amount {
		return fmt.Errorf("insufficient balance: has %d, need %d", fromBalance, amount)
	}

	// Check for overflow on recipient balance
	toBalance := ts.getBalance(to, tokenID)
	if toBalance > 0 && amount > (int64(^uint64(0)>>1)-toBalance) {
		return fmt.Errorf("recipient balance would overflow")
	}

	ts.balances[from][tokenID] -= amount

	if ts.balances[to] == nil {
		ts.balances[to] = make(map[wire.Hash]int64)
	}
	ts.balances[to][tokenID] += amount

	return nil
}

// GetBalance returns the token balance for an address
func (ts *TokenStore) GetBalance(address string, tokenID wire.Hash) int64 {
	ts.mutex.RLock()
	defer ts.mutex.RUnlock()
	return ts.getBalance(address, tokenID)
}

// getBalance is the internal balance getter
func (ts *TokenStore) getBalance(address string, tokenID wire.Hash) int64 {
	if balances, exists := ts.balances[address]; exists {
		return balances[tokenID]
	}
	return 0
}

// GetToken returns token information by ID
func (ts *TokenStore) GetToken(tokenID wire.Hash) (*wire.Token, error) {
	ts.mutex.RLock()
	defer ts.mutex.RUnlock()

	token, exists := ts.tokens[tokenID]
	if !exists {
		return nil, fmt.Errorf("token not found")
	}
	return token, nil
}

// GetTokenBySymbol returns token information by symbol
func (ts *TokenStore) GetTokenBySymbol(symbol string) (*wire.Token, error) {
	ts.mutex.RLock()
	defer ts.mutex.RUnlock()

	tokenID, exists := ts.tokenIndex[symbol]
	if !exists {
		return nil, fmt.Errorf("token symbol not found")
	}
	return ts.tokens[tokenID], nil
}

// ListTokens returns all tokens
func (ts *TokenStore) ListTokens() []*wire.Token {
	ts.mutex.RLock()
	defer ts.mutex.RUnlock()

	tokens := make([]*wire.Token, 0, len(ts.tokens))
	for _, token := range ts.tokens {
		tokens = append(tokens, token)
	}
	return tokens
}

// GetAddressTokens returns all tokens held by an address
func (ts *TokenStore) GetAddressTokens(address string) map[wire.Hash]int64 {
	ts.mutex.RLock()
	defer ts.mutex.RUnlock()

	if balances, exists := ts.balances[address]; exists {
		result := make(map[wire.Hash]int64)
		for tokenID, balance := range balances {
			if balance > 0 {
				result[tokenID] = balance
			}
		}
		return result
	}
	return make(map[wire.Hash]int64)
}
