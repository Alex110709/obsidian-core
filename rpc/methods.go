package rpc

import (
	"encoding/hex"
	"fmt"
	"obsidian-core/smartcontract"
	"obsidian-core/wire"
	"strings"
)

// getBlockCount returns the current block height.
func (s *Server) getBlockCount(params []interface{}) (interface{}, error) {
	return s.chain.Height(), nil
}

// getBestBlockHash returns the hash of the best (tip) block.
func (s *Server) getBestBlockHash(params []interface{}) (interface{}, error) {
	block, err := s.chain.BestBlock()
	if err != nil {
		return nil, fmt.Errorf("failed to get best block: %v", err)
	}

	hash := block.BlockHash()
	return hash.String(), nil
}

// getBlock returns information about a block by hash.
func (s *Server) getBlock(params []interface{}) (interface{}, error) {
	if len(params) < 1 {
		return nil, fmt.Errorf("missing block hash parameter")
	}

	hashStr, ok := params[0].(string)
	if !ok {
		return nil, fmt.Errorf("invalid block hash parameter")
	}

	// Decode hex string to bytes
	hashBytes, err := hex.DecodeString(hashStr)
	if err != nil {
		return nil, fmt.Errorf("invalid hash format: %v", err)
	}

	// Get block from database
	block, err := s.chain.GetBlock(hashBytes)
	if err != nil {
		return nil, fmt.Errorf("block not found: %v", err)
	}

	// Convert to BlockInfo
	blockHash := block.BlockHash()
	blockInfo := BlockInfo{
		Hash:         blockHash.String(),
		Height:       s.chain.Height(),
		Version:      block.Header.Version,
		PrevBlock:    block.Header.PrevBlock.String(),
		MerkleRoot:   block.Header.MerkleRoot.String(),
		Timestamp:    block.Header.Timestamp.Unix(),
		Bits:         block.Header.Bits,
		Nonce:        block.Header.Nonce,
		Transactions: len(block.Transactions),
	}

	return blockInfo, nil
}

// getBlockchainInfo returns general blockchain information.
func (s *Server) getBlockchainInfo(params []interface{}) (interface{}, error) {
	block, err := s.chain.BestBlock()
	if err != nil {
		return nil, fmt.Errorf("failed to get best block: %v", err)
	}

	hash := block.BlockHash()
	info := BlockchainInfo{
		Chain:         s.chain.Params().Name,
		Blocks:        s.chain.Height(),
		BestBlockHash: hash.String(),
		Difficulty:    block.Header.Bits,
		MaxMoney:      s.chain.Params().MaxMoney,
		InitialSupply: s.chain.Params().InitialSupply,
	}

	return info, nil
}

// getMiningInfo returns mining-related information.
func (s *Server) getMiningInfo(params []interface{}) (interface{}, error) {
	block, err := s.chain.BestBlock()
	if err != nil {
		return nil, fmt.Errorf("failed to get best block: %v", err)
	}

	hash := block.BlockHash()
	currentHeight := s.chain.Height()
	blockReward := s.chain.Params().CalcBlockSubsidy(currentHeight + 1)

	var hashesPerSec int64
	if s.miner != nil {
		hashesPerSec = int64(s.miner.GetHashRate())
	}

	info := MiningInfo{
		Blocks:       currentHeight,
		CurrentHash:  hash.String(),
		Difficulty:   block.Header.Bits,
		MiningActive: s.miner != nil,
		HashesPerSec: hashesPerSec,
		BlockReward:  blockReward,
	}

	return info, nil
}

// getBlockReward returns the block reward for a given height.
func (s *Server) getBlockReward(params []interface{}) (interface{}, error) {
	var height int32

	if len(params) > 0 {
		// Height provided
		heightFloat, ok := params[0].(float64)
		if !ok {
			return nil, fmt.Errorf("invalid height parameter")
		}
		height = int32(heightFloat)
	} else {
		// Use next block height
		height = s.chain.Height() + 1
	}

	reward := s.chain.Params().CalcBlockSubsidy(height)

	result := map[string]interface{}{
		"height": height,
		"reward": reward,
	}

	return result, nil
}

// getnewaddress generates a new transparent t-address.
func (s *Server) getnewaddress(params []interface{}) (interface{}, error) {
	// Generate new transparent address
	addr, err := s.wallet.GetNewAddress()
	if err != nil {
		return nil, fmt.Errorf("failed to generate transparent address: %v", err)
	}

	return addr, nil
}

// getbalance returns the balance for a transparent address.
func (s *Server) getbalance(params []interface{}) (interface{}, error) {
	if len(params) < 1 {
		return nil, fmt.Errorf("missing address parameter")
	}

	address, ok := params[0].(string)
	if !ok {
		return nil, fmt.Errorf("invalid address parameter")
	}

	// Get transparent balance
	balance, err := s.wallet.GetBalance(address)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %v", err)
	}

	result := map[string]interface{}{
		"address":     address,
		"balance":     balance,
		"balance_obs": float64(balance) / 100000000.0,
	}

	return result, nil
}

// sendtoaddress sends funds from one transparent address to another.
func (s *Server) sendtoaddress(params []interface{}) (interface{}, error) {
	if len(params) < 3 {
		return nil, fmt.Errorf("missing parameters: sendtoaddress <from_address> <to_address> <amount>")
	}

	fromAddress, ok := params[0].(string)
	if !ok {
		return nil, fmt.Errorf("invalid from_address parameter")
	}

	toAddress, ok := params[1].(string)
	if !ok {
		return nil, fmt.Errorf("invalid to_address parameter")
	}

	amountFloat, ok := params[2].(float64)
	if !ok {
		return nil, fmt.Errorf("invalid amount parameter")
	}
	amount := int64(amountFloat * 100000000) // Convert to satoshis

	// Send transparent transaction
	txid, err := s.wallet.SendToAddress(fromAddress, toAddress, amount)
	if err != nil {
		return nil, fmt.Errorf("failed to send transaction: %v", err)
	}

	result := map[string]interface{}{
		"txid":   txid,
		"from":   fromAddress,
		"to":     toAddress,
		"amount": amount,
	}

	return result, nil
}

// listaddresses lists all transparent t-addresses in the wallet.
func (s *Server) listaddresses(params []interface{}) (interface{}, error) {
	addresses := s.wallet.ListAddresses()
	return addresses, nil
}

// estimateFee estimates the fee for a transaction of given size.
func (s *Server) estimateFee(params []interface{}) (interface{}, error) {
	var sizeBytes int

	if len(params) > 0 {
		// Size provided
		sizeFloat, ok := params[0].(float64)
		if !ok {
			return nil, fmt.Errorf("invalid size parameter")
		}
		sizeBytes = int(sizeFloat)
	} else {
		// Use default transaction size (250 bytes average)
		sizeBytes = 250
	}

	fee := s.chain.Params().CalcTxFee(sizeBytes)

	result := map[string]interface{}{
		"size_bytes":   sizeBytes,
		"fee_satoshis": fee,
		"fee_obs":      float64(fee) / 100000000.0,
		"fee_per_byte": s.chain.Params().FeePerByte,
	}

	return result, nil
}

// z_getnewaddress generates a new shielded z-address.
func (s *Server) z_getnewaddress(params []interface{}) (interface{}, error) {
	// Generate new shielded address
	addr, err := s.wallet.NewShieldedAddress()
	if err != nil {
		return nil, fmt.Errorf("failed to generate shielded address: %v", err)
	}

	return addr, nil
}

// z_listaddresses lists all shielded z-addresses in the wallet.
func (s *Server) z_listaddresses(params []interface{}) (interface{}, error) {
	addresses := s.wallet.ListShieldedAddresses()
	return addresses, nil
}

// z_getbalance returns the shielded balance for a z-address.
func (s *Server) z_getbalance(params []interface{}) (interface{}, error) {
	if len(params) < 1 {
		return nil, fmt.Errorf("missing z-address parameter")
	}

	address, ok := params[0].(string)
	if !ok {
		return nil, fmt.Errorf("invalid z-address parameter")
	}

	// Get shielded balance
	balance, err := s.wallet.GetShieldedBalance(address)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %v", err)
	}

	result := map[string]interface{}{
		"address":     address,
		"balance":     balance,
		"balance_obs": float64(balance) / 100000000.0,
	}

	return result, nil
}

// z_sendmany sends funds from a z-address to multiple recipients (transparent or shielded).
func (s *Server) z_sendmany(params []interface{}) (interface{}, error) {
	if len(params) < 2 {
		return nil, fmt.Errorf("missing parameters: z_sendmany <from_address> <amounts> [memo]")
	}

	fromAddress, ok := params[0].(string)
	if !ok {
		return nil, fmt.Errorf("invalid from_address parameter")
	}

	amounts, ok := params[1].([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid amounts parameter")
	}

	// Optional memo
	memo := ""
	if len(params) > 2 {
		memo, _ = params[2].(string)
	}

	// Parse recipients
	recipients := make([]ShieldedRecipient, 0, len(amounts))
	for _, amt := range amounts {
		amtMap, ok := amt.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid recipient format")
		}

		address, _ := amtMap["address"].(string)
		amount, _ := amtMap["amount"].(float64)
		recipientMemo, _ := amtMap["memo"].(string)

		if recipientMemo == "" {
			recipientMemo = memo
		}

		recipients = append(recipients, ShieldedRecipient{
			Address: address,
			Amount:  int64(amount * 100000000), // Convert to satoshis
			Memo:    recipientMemo,
		})
	}

	// Create and send shielded transaction
	txid, err := s.wallet.SendShielded(fromAddress, recipients)
	if err != nil {
		return nil, fmt.Errorf("failed to send transaction: %v", err)
	}

	result := map[string]interface{}{
		"txid":       txid,
		"from":       fromAddress,
		"recipients": len(recipients),
	}

	return result, nil
}

// z_listreceivedbyaddress lists amounts received by a z-address.
func (s *Server) z_listreceivedbyaddress(params []interface{}) (interface{}, error) {
	if len(params) < 1 {
		return nil, fmt.Errorf("missing z-address parameter")
	}

	address, ok := params[0].(string)
	if !ok {
		return nil, fmt.Errorf("invalid z-address parameter")
	}

	// Get received transactions
	received, err := s.wallet.ListReceivedShielded(address)
	if err != nil {
		return nil, fmt.Errorf("failed to get received transactions: %v", err)
	}

	return received, nil
}

// z_gettotalbalance returns the total balance (transparent + shielded).
func (s *Server) z_gettotalbalance(params []interface{}) (interface{}, error) {
	transparentBalance := s.wallet.GetTransparentBalance()
	shieldedBalance := s.wallet.GetTotalShieldedBalance()
	total := transparentBalance + shieldedBalance

	result := map[string]interface{}{
		"transparent": float64(transparentBalance) / 100000000.0,
		"shielded":    float64(shieldedBalance) / 100000000.0,
		"total":       float64(total) / 100000000.0,
	}

	return result, nil
}

// z_exportviewingkey exports the viewing key for a z-address.
func (s *Server) z_exportviewingkey(params []interface{}) (interface{}, error) {
	if len(params) < 1 {
		return nil, fmt.Errorf("missing z-address parameter")
	}

	address, ok := params[0].(string)
	if !ok {
		return nil, fmt.Errorf("invalid z-address parameter")
	}

	viewingKey, err := s.wallet.ExportViewingKey(address)
	if err != nil {
		return nil, fmt.Errorf("failed to export viewing key: %v", err)
	}

	return viewingKey, nil
}

// z_importviewingkey imports a viewing key (read-only access).
func (s *Server) z_importviewingkey(params []interface{}) (interface{}, error) {
	if len(params) < 1 {
		return nil, fmt.Errorf("missing viewing key parameter")
	}

	viewingKey, ok := params[0].(string)
	if !ok {
		return nil, fmt.Errorf("invalid viewing key parameter")
	}

	err := s.wallet.ImportViewingKey(viewingKey)
	if err != nil {
		return nil, fmt.Errorf("failed to import viewing key: %v", err)
	}

	return "Viewing key imported successfully", nil
}

// z_shieldcoinbase shields coinbase funds to a z-address.
func (s *Server) z_shieldcoinbase(params []interface{}) (interface{}, error) {
	if len(params) < 1 {
		return nil, fmt.Errorf("missing z-address parameter")
	}

	toAddress, ok := params[0].(string)
	if !ok {
		return nil, fmt.Errorf("invalid z-address parameter")
	}

	// Shield all transparent coinbase funds
	txid, err := s.wallet.ShieldCoinbase(toAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to shield coinbase: %v", err)
	}

	result := map[string]interface{}{
		"txid":         txid,
		"shielding_to": toAddress,
	}

	return result, nil
}

// getpoolinfo returns mining pool statistics
func (s *Server) getpoolinfo(params []interface{}, wallet interface{}) (interface{}, error) {
	if s.pool == nil {
		return map[string]interface{}{
			"enabled": false,
			"message": "Pool server not running",
		}, nil
	}

	stats := s.pool.GetStats()
	stats["enabled"] = true

	return stats, nil
}

// issueToken creates a new token
func (s *Server) issueToken(params []interface{}) (interface{}, error) {
	if len(params) < 4 {
		return nil, fmt.Errorf("insufficient parameters: need name, symbol, decimals, supply")
	}

	name, ok := params[0].(string)
	if !ok {
		return nil, fmt.Errorf("invalid name parameter")
	}

	symbol, ok := params[1].(string)
	if !ok {
		return nil, fmt.Errorf("invalid symbol parameter")
	}

	decimalsFloat, ok := params[2].(float64)
	if !ok {
		return nil, fmt.Errorf("invalid decimals parameter")
	}
	decimals := uint8(decimalsFloat)

	supplyFloat, ok := params[3].(float64)
	if !ok {
		return nil, fmt.Errorf("invalid supply parameter")
	}
	supply := int64(supplyFloat)

	// Get owner from params or use default
	owner := "default_owner"
	if len(params) > 4 {
		if ownerParam, ok := params[4].(string); ok {
			owner = ownerParam
		}
	}

	// Create token issuance transaction
	tokenIssue := &wire.TokenIssue{
		Name:     name,
		Symbol:   symbol,
		Decimals: decimals,
		Supply:   supply,
		Owner:    owner,
	}

	tx := wire.NewTokenIssueTx(owner, tokenIssue)

	// Add to mempool (simplified - in production, validate and broadcast)
	fmt.Printf("Token issuance transaction created: %s\n", tx.TxHash().String())

	return map[string]interface{}{
		"txid":   tx.TxHash().String(),
		"name":   name,
		"symbol": symbol,
		"supply": supply,
		"owner":  owner,
	}, nil
}

// transferToken transfers tokens between addresses
func (s *Server) transferToken(params []interface{}) (interface{}, error) {
	if len(params) < 4 {
		return nil, fmt.Errorf("insufficient parameters: need token_symbol, to_address, amount")
	}

	tokenSymbol, ok := params[0].(string)
	if !ok {
		return nil, fmt.Errorf("invalid token symbol parameter")
	}

	fromAddress, ok := params[1].(string)
	if !ok {
		return nil, fmt.Errorf("invalid from address parameter")
	}

	toAddress, ok := params[2].(string)
	if !ok {
		return nil, fmt.Errorf("invalid to address parameter")
	}

	amountFloat, ok := params[3].(float64)
	if !ok {
		return nil, fmt.Errorf("invalid amount parameter")
	}
	amount := int64(amountFloat)

	// Get token by symbol
	token, err := s.chain.GetTokenStore().GetTokenBySymbol(tokenSymbol)
	if err != nil {
		return nil, fmt.Errorf("token not found: %v", err)
	}

	// Create token transfer transaction
	tx := wire.NewTokenTransferTx(fromAddress, toAddress, token.ID, amount)

	// Add to mempool (simplified)
	fmt.Printf("Token transfer transaction created: %s\n", tx.TxHash().String())

	return map[string]interface{}{
		"txid":   tx.TxHash().String(),
		"token":  tokenSymbol,
		"from":   fromAddress,
		"to":     toAddress,
		"amount": amount,
	}, nil
}

// getTokenBalance returns the token balance for an address
func (s *Server) getTokenBalance(params []interface{}) (interface{}, error) {
	if len(params) < 2 {
		return nil, fmt.Errorf("insufficient parameters: need address, token_symbol")
	}

	address, ok := params[0].(string)
	if !ok {
		return nil, fmt.Errorf("invalid address parameter")
	}

	tokenSymbol, ok := params[1].(string)
	if !ok {
		return nil, fmt.Errorf("invalid token symbol parameter")
	}

	// Get token by symbol
	token, err := s.chain.GetTokenStore().GetTokenBySymbol(tokenSymbol)
	if err != nil {
		return nil, fmt.Errorf("token not found: %v", err)
	}

	// Get balance
	balance := s.chain.GetTokenStore().GetBalance(address, token.ID)

	return map[string]interface{}{
		"address": address,
		"token":   tokenSymbol,
		"balance": balance,
	}, nil
}

// getTokenInfo returns information about a token
func (s *Server) getTokenInfo(params []interface{}) (interface{}, error) {
	if len(params) < 1 {
		return nil, fmt.Errorf("insufficient parameters: need token_symbol")
	}

	tokenSymbol, ok := params[0].(string)
	if !ok {
		return nil, fmt.Errorf("invalid token symbol parameter")
	}

	// Get token by symbol
	token, err := s.chain.GetTokenStore().GetTokenBySymbol(tokenSymbol)
	if err != nil {
		return nil, fmt.Errorf("token not found: %v", err)
	}

	return map[string]interface{}{
		"id":          token.ID.String(),
		"name":        token.Name,
		"symbol":      token.Symbol,
		"decimals":    token.Decimals,
		"totalSupply": token.TotalSupply,
		"owner":       token.Owner,
		"created":     token.Created,
	}, nil
}

// listTokens returns all tokens
func (s *Server) listTokens(params []interface{}) (interface{}, error) {
	tokens := s.chain.GetTokenStore().ListTokens()

	result := make([]map[string]interface{}, len(tokens))
	for i, token := range tokens {
		result[i] = map[string]interface{}{
			"id":          token.ID.String(),
			"name":        token.Name,
			"symbol":      token.Symbol,
			"decimals":    token.Decimals,
			"totalSupply": token.TotalSupply,
			"owner":       token.Owner,
			"created":     token.Created,
		}
	}

	return result, nil
}

// getAddressTokens returns all tokens held by an address
func (s *Server) getAddressTokens(params []interface{}) (interface{}, error) {
	if len(params) < 1 {
		return nil, fmt.Errorf("insufficient parameters: need address")
	}

	address, ok := params[0].(string)
	if !ok {
		return nil, fmt.Errorf("invalid address parameter")
	}

	tokens := s.chain.GetTokenStore().GetAddressTokens(address)

	result := make(map[string]int64)
	for tokenID, balance := range tokens {
		if token, err := s.chain.GetTokenStore().GetToken(tokenID); err == nil {
			result[token.Symbol] = balance
		}
	}

	return map[string]interface{}{
		"address": address,
		"tokens":  result,
	}, nil
}

// shieldtoken shields or unshield tokens using shielded transactions
func (s *Server) shieldtoken(params []interface{}) (interface{}, error) {
	if len(params) < 4 {
		return nil, fmt.Errorf("insufficient parameters: need from_address, to_address, token_symbol, amount")
	}

	fromAddress, ok := params[0].(string)
	if !ok {
		return nil, fmt.Errorf("invalid from address parameter")
	}

	toAddress, ok := params[1].(string)
	if !ok {
		return nil, fmt.Errorf("invalid to address parameter")
	}

	tokenSymbol, ok := params[2].(string)
	if !ok {
		return nil, fmt.Errorf("invalid token symbol parameter")
	}

	amountFloat, ok := params[3].(float64)
	if !ok {
		return nil, fmt.Errorf("invalid amount parameter")
	}
	amount := int64(amountFloat)

	// Get token by symbol
	token, err := s.chain.GetTokenStore().GetTokenBySymbol(tokenSymbol)
	if err != nil {
		return nil, fmt.Errorf("token not found: %v", err)
	}

	// Create token shielded transaction
	tx := wire.NewTokenShieldedTx(fromAddress, toAddress, token.ID, amount)

	// Add to mempool (simplified)
	fmt.Printf("Token shielded transaction created: %s\n", tx.TxHash().String())

	isShielding := strings.HasPrefix(toAddress, "zobs")
	action := "shielding"
	if !isShielding {
		action = "unshielding"
	}

	return map[string]interface{}{
		"txid":   tx.TxHash().String(),
		"action": action,
		"token":  tokenSymbol,
		"from":   fromAddress,
		"to":     toAddress,
		"amount": amount,
	}, nil
}

// minttoken mints additional tokens for a mintable token
func (s *Server) minttoken(params []interface{}) (interface{}, error) {
	if len(params) < 4 {
		return nil, fmt.Errorf("insufficient parameters: need token_symbol, amount, to_address, from_address")
	}

	tokenSymbol, ok := params[0].(string)
	if !ok {
		return nil, fmt.Errorf("invalid token symbol parameter")
	}

	amountFloat, ok := params[1].(float64)
	if !ok {
		return nil, fmt.Errorf("invalid amount parameter")
	}
	amount := int64(amountFloat)

	toAddress, ok := params[2].(string)
	if !ok {
		return nil, fmt.Errorf("invalid to address parameter")
	}

	fromAddress, ok := params[3].(string)
	if !ok {
		return nil, fmt.Errorf("invalid from address parameter")
	}

	// Get token by symbol
	token, err := s.chain.GetTokenStore().GetTokenBySymbol(tokenSymbol)
	if err != nil {
		return nil, fmt.Errorf("token not found: %v", err)
	}

	if !token.Mintable {
		return nil, fmt.Errorf("token is not mintable")
	}

	if fromAddress != token.Owner {
		return nil, fmt.Errorf("only token owner can mint tokens")
	}

	// Create token mint transaction
	tx := wire.NewTokenMintTx(fromAddress, toAddress, token.ID, amount)

	// Add to mempool (simplified)
	fmt.Printf("Token mint transaction created: %s\n", tx.TxHash().String())

	return map[string]interface{}{
		"txid":   tx.TxHash().String(),
		"token":  tokenSymbol,
		"amount": amount,
		"to":     toAddress,
		"from":   fromAddress,
	}, nil
}

// transfertokenownership transfers token ownership to a new address
func (s *Server) transfertokenownership(params []interface{}) (interface{}, error) {
	if len(params) < 3 {
		return nil, fmt.Errorf("insufficient parameters: need token_symbol, new_owner_address, from_address")
	}

	tokenSymbol, ok := params[0].(string)
	if !ok {
		return nil, fmt.Errorf("invalid token symbol parameter")
	}

	newOwnerAddress, ok := params[1].(string)
	if !ok {
		return nil, fmt.Errorf("invalid new owner address parameter")
	}

	fromAddress, ok := params[2].(string)
	if !ok {
		return nil, fmt.Errorf("invalid from address parameter")
	}

	// Get token by symbol
	token, err := s.chain.GetTokenStore().GetTokenBySymbol(tokenSymbol)
	if err != nil {
		return nil, fmt.Errorf("token not found: %v", err)
	}

	if fromAddress != token.Owner {
		return nil, fmt.Errorf("only current owner can transfer ownership")
	}

	// Create token ownership transfer transaction
	tx := wire.NewTokenTransferOwnershipTx(fromAddress, newOwnerAddress, token.ID)

	// Add to mempool (simplified)
	fmt.Printf("Token ownership transfer transaction created: %s\n", tx.TxHash().String())

	return map[string]interface{}{
		"txid":     tx.TxHash().String(),
		"token":    tokenSymbol,
		"oldOwner": fromAddress,
		"newOwner": newOwnerAddress,
	}, nil
}

// burntoken burns tokens permanently
func (s *Server) burntoken(params []interface{}) (interface{}, error) {
	if len(params) < 3 {
		return nil, fmt.Errorf("insufficient parameters: need token_symbol, amount, from_address")
	}

	tokenSymbol, ok := params[0].(string)
	if !ok {
		return nil, fmt.Errorf("invalid token symbol parameter")
	}

	amountFloat, ok := params[1].(float64)
	if !ok {
		return nil, fmt.Errorf("invalid amount parameter")
	}
	amount := int64(amountFloat)

	fromAddress, ok := params[2].(string)
	if !ok {
		return nil, fmt.Errorf("invalid from address parameter")
	}

	// Get token by symbol
	token, err := s.chain.GetTokenStore().GetTokenBySymbol(tokenSymbol)
	if err != nil {
		return nil, fmt.Errorf("token not found: %v", err)
	}

	// Check sender has sufficient balance
	balance := s.chain.GetTokenStore().GetBalance(fromAddress, token.ID)
	if balance < amount {
		return nil, fmt.Errorf("insufficient balance: has %d, need %d", balance, amount)
	}

	// Create token burn transaction
	tx := wire.NewTokenBurnTx(fromAddress, token.ID, amount)

	// Add to mempool (simplified)
	fmt.Printf("Token burn transaction created: %s\n", tx.TxHash().String())

	return map[string]interface{}{
		"txid":   tx.TxHash().String(),
		"token":  tokenSymbol,
		"amount": amount,
		"from":   fromAddress,
		"action": "burn",
	}, nil
}

// getPeerInfo returns information about connected peers
func (s *Server) getPeerInfo(params []interface{}) (interface{}, error) {
	if s.syncManager == nil {
		return []interface{}{}, nil
	}

	// Type assertion to access methods
	sm, ok := s.syncManager.(interface {
		GetPeerInfo() []interface{}
	})
	if !ok {
		return []interface{}{}, nil
	}

	return sm.GetPeerInfo(), nil
}

// getConnectionCount returns the number of connections
func (s *Server) getConnectionCount(params []interface{}) (interface{}, error) {
	if s.syncManager == nil {
		return map[string]int{"inbound": 0, "outbound": 0}, nil
	}

	sm, ok := s.syncManager.(interface {
		GetConnectionStats() (int, int, int)
	})
	if !ok {
		return map[string]int{"inbound": 0, "outbound": 0}, nil
	}

	inbound, outbound, banned := sm.GetConnectionStats()
	return map[string]int{
		"inbound":  inbound,
		"outbound": outbound,
		"banned":   banned,
	}, nil
}

// createmultisig creates a multisig address
func (s *Server) createmultisig(params []interface{}) (interface{}, error) {
	if len(params) < 2 {
		return nil, fmt.Errorf("insufficient parameters: need nrequired, keys")
	}

	nRequiredFloat, ok := params[0].(float64)
	if !ok {
		return nil, fmt.Errorf("invalid nrequired parameter")
	}
	nRequired := int(nRequiredFloat)

	keys, ok := params[1].([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid keys parameter")
	}

	publicKeys := make([]string, len(keys))
	for i, key := range keys {
		publicKeys[i], ok = key.(string)
		if !ok {
			return nil, fmt.Errorf("invalid public key at index %d", i)
		}
	}

	multisigInfo, err := s.wallet.CreateMultiSigAddress(nRequired, publicKeys)
	if err != nil {
		return nil, fmt.Errorf("failed to create multisig address: %v", err)
	}

	return map[string]interface{}{
		"address":      multisigInfo.Address,
		"redeemscript": multisigInfo.RedeemScript,
		"m":            multisigInfo.M,
		"n":            multisigInfo.N,
		"pubkeys":      multisigInfo.PublicKeys,
	}, nil
}

// addmultisigaddress adds a multisig address to the wallet
func (s *Server) addmultisigaddress(params []interface{}) (interface{}, error) {
	if len(params) < 2 {
		return nil, fmt.Errorf("insufficient parameters: need nrequired, keys")
	}

	nRequiredFloat, ok := params[0].(float64)
	if !ok {
		return nil, fmt.Errorf("invalid nrequired parameter")
	}
	nRequired := int(nRequiredFloat)

	keys, ok := params[1].([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid keys parameter")
	}

	publicKeys := make([]string, len(keys))
	for i, key := range keys {
		publicKeys[i], ok = key.(string)
		if !ok {
			return nil, fmt.Errorf("invalid public key at index %d", i)
		}
	}

	account := ""
	if len(params) > 2 {
		account, _ = params[2].(string)
	}

	address, err := s.wallet.AddMultiSigAddress(nRequired, publicKeys, account)
	if err != nil {
		return nil, fmt.Errorf("failed to add multisig address: %v", err)
	}

	return address, nil
}

// signmultisigtx signs a multisig transaction
func (s *Server) signmultisigtx(params []interface{}) (interface{}, error) {
	if len(params) < 2 {
		return nil, fmt.Errorf("insufficient parameters: need txhex, redeemscript")
	}

	txHex, ok := params[0].(string)
	if !ok {
		return nil, fmt.Errorf("invalid txhex parameter")
	}

	redeemScript, ok := params[1].(string)
	if !ok {
		return nil, fmt.Errorf("invalid redeemscript parameter")
	}

	var privateKeys []string
	if len(params) > 2 {
		keys, ok := params[2].([]interface{})
		if ok {
			privateKeys = make([]string, len(keys))
			for i, key := range keys {
				privateKeys[i], _ = key.(string)
			}
		}
	}

	multisigTx, err := s.wallet.SignMultiSigTx(txHex, redeemScript, privateKeys)
	if err != nil {
		return nil, fmt.Errorf("failed to sign multisig transaction: %v", err)
	}

	return map[string]interface{}{
		"txid":         multisigTx.TxID,
		"hex":          multisigTx.Hex,
		"complete":     multisigTx.Complete,
		"missing_sigs": multisigTx.MissingSigs,
		"signatures":   multisigTx.Signatures,
	}, nil
}

// combinemultisigsigs combines multiple multisig signatures
func (s *Server) combinemultisigsigs(params []interface{}) (interface{}, error) {
	if len(params) < 2 {
		return nil, fmt.Errorf("insufficient parameters: need txhex, signatures")
	}

	txHex, ok := params[0].(string)
	if !ok {
		return nil, fmt.Errorf("invalid txhex parameter")
	}

	sigs, ok := params[1].([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid signatures parameter")
	}

	signatures := make([]string, len(sigs))
	for i, sig := range sigs {
		signatures[i], ok = sig.(string)
		if !ok {
			return nil, fmt.Errorf("invalid signature at index %d", i)
		}
	}

	multisigTx, err := s.wallet.CombineMultiSigSigs(txHex, signatures)
	if err != nil {
		return nil, fmt.Errorf("failed to combine multisig signatures: %v", err)
	}

	return map[string]interface{}{
		"txid":     multisigTx.TxID,
		"hex":      multisigTx.Hex,
		"complete": multisigTx.Complete,
	}, nil
}

// shield converts funds from a transparent address to a shielded address
func (s *Server) shield(params []interface{}) (interface{}, error) {
	if len(params) < 3 {
		return nil, fmt.Errorf("insufficient parameters: need from_address, to_shielded_address, amount")
	}

	fromAddress, ok := params[0].(string)
	if !ok {
		return nil, fmt.Errorf("invalid from_address parameter")
	}

	toShieldedAddress, ok := params[1].(string)
	if !ok {
		return nil, fmt.Errorf("invalid to_shielded_address parameter")
	}

	amountFloat, ok := params[2].(float64)
	if !ok {
		return nil, fmt.Errorf("invalid amount parameter")
	}
	amount := int64(amountFloat)

	// Validate addresses
	if !strings.HasPrefix(fromAddress, "obs") {
		return nil, fmt.Errorf("from_address must be a transparent address starting with 'obs'")
	}
	if !strings.HasPrefix(toShieldedAddress, "zobs") {
		return nil, fmt.Errorf("to_shielded_address must be a shielded address starting with 'zobs'")
	}

	// Check balance
	balance, err := s.wallet.GetBalance(fromAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %v", err)
	}
	if balance < amount {
		return nil, fmt.Errorf("insufficient balance: has %d, need %d", balance, amount)
	}

	// Create shielded transaction
	tx := wire.NewShieldTx(fromAddress, toShieldedAddress, amount)

	// Add to mempool (simplified)
	fmt.Printf("Shield transaction created: %s\n", tx.TxHash().String())

	return map[string]interface{}{
		"txid":   tx.TxHash().String(),
		"action": "shield",
		"from":   fromAddress,
		"to":     toShieldedAddress,
		"amount": amount,
	}, nil
}

// unshield converts funds from a shielded address to a transparent address
func (s *Server) unshield(params []interface{}) (interface{}, error) {
	if len(params) < 3 {
		return nil, fmt.Errorf("insufficient parameters: need from_shielded_address, to_address, amount")
	}

	fromShieldedAddress, ok := params[0].(string)
	if !ok {
		return nil, fmt.Errorf("invalid from_shielded_address parameter")
	}

	toAddress, ok := params[1].(string)
	if !ok {
		return nil, fmt.Errorf("invalid to_address parameter")
	}

	amountFloat, ok := params[2].(float64)
	if !ok {
		return nil, fmt.Errorf("invalid amount parameter")
	}
	amount := int64(amountFloat)

	// Validate addresses
	if !strings.HasPrefix(fromShieldedAddress, "zobs") {
		return nil, fmt.Errorf("from_shielded_address must be a shielded address starting with 'zobs'")
	}
	if !strings.HasPrefix(toAddress, "obs") {
		return nil, fmt.Errorf("to_address must be a transparent address starting with 'obs'")
	}

	// Check shielded balance
	balance, err := s.wallet.GetShieldedBalance(fromShieldedAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to get shielded balance: %v", err)
	}
	if balance < amount {
		return nil, fmt.Errorf("insufficient shielded balance: has %d, need %d", balance, amount)
	}

	// Create unshielded transaction
	tx := wire.NewUnshieldTx(fromShieldedAddress, toAddress, amount)

	// Add to mempool (simplified)
	fmt.Printf("Unshield transaction created: %s\n", tx.TxHash().String())

	return map[string]interface{}{
		"txid":   tx.TxHash().String(),
		"action": "unshield",
		"from":   fromShieldedAddress,
		"to":     toAddress,
		"amount": amount,
	}, nil
}

// deploycontract deploys a smart contract
func (s *Server) deploycontract(params []interface{}) (interface{}, error) {
	if len(params) < 1 {
		return nil, fmt.Errorf("insufficient parameters: need contract_code")
	}

	contractCode, ok := params[0].(string)
	if !ok {
		return nil, fmt.Errorf("invalid contract_code parameter")
	}

	// Parse and compile contract
	lexer := smartcontract.NewLexer(contractCode)
	tokens, err := lexer.Tokenize()
	if err != nil {
		return nil, fmt.Errorf("lexer error: %v", err)
	}

	parser := smartcontract.NewParser(tokens)
	ast, err := parser.Parse()
	if err != nil {
		return nil, fmt.Errorf("parser error: %v", err)
	}

	compiler := smartcontract.NewCompiler()
	bytecode := compiler.Compile(ast)

	// Create deployment transaction
	tx := wire.NewMsgTx(1)
	tx.TxType = wire.TxTypeSmartContractDeploy
	tx.Memo = []byte(contractCode) // Store code in memo

	// Add to mempool (simplified)
	fmt.Printf("Smart contract deployment transaction created: %s\n", tx.TxHash().String())

	return map[string]interface{}{
		"txid":     tx.TxHash().String(),
		"action":   "deploy",
		"code":     contractCode,
		"bytecode": fmt.Sprintf("%v", bytecode), // Simplified
	}, nil
}

// callcontract calls a smart contract function
func (s *Server) callcontract(params []interface{}) (interface{}, error) {
	if len(params) < 2 {
		return nil, fmt.Errorf("insufficient parameters: need contract_address, function_name, [args...]")
	}

	contractAddress, ok := params[0].(string)
	if !ok {
		return nil, fmt.Errorf("invalid contract_address parameter")
	}

	functionName, ok := params[1].(string)
	if !ok {
		return nil, fmt.Errorf("invalid function_name parameter")
	}

	// Simplified: assume bytecode is stored somewhere
	// In real implementation, retrieve contract bytecode from storage

	// Mock execution
	result := "contract executed successfully"

	return map[string]interface{}{
		"contract": contractAddress,
		"function": functionName,
		"result":   result,
	}, nil
}
