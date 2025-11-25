package rpc

import (
	"encoding/hex"
	"fmt"
	"obsidian-core/wire"
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

	info := MiningInfo{
		Blocks:       currentHeight,
		CurrentHash:  hash.String(),
		Difficulty:   block.Header.Bits,
		MiningActive: s.miner != nil,
		HashesPerSec: 0, // TODO: Implement hash rate calculation
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
