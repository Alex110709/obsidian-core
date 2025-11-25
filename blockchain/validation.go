package blockchain

import (
	"crypto/ecdsa"
	"fmt"
	"obsidian-core/crypto"
	"obsidian-core/wire"
	"strconv"
	"strings"
)

// ValidateTransaction validates a transaction against the UTXO set
func (b *BlockChain) ValidateTransaction(tx *wire.MsgTx, utxoSet *UTXOSet) error {
	// Skip validation for coinbase transactions
	if tx.IsCoinbase() {
		return nil
	}

	// Handle different transaction types
	switch tx.TxType {
	case wire.TxTypeTokenIssue:
		return b.validateTokenIssueTransaction(tx)
	case wire.TxTypeTokenTransfer:
		return b.validateTokenTransferTransaction(tx, utxoSet)
	}

	// 1. Check inputs exist and are unspent
	var totalInput int64
	for _, txIn := range tx.TxIn {
		utxo, err := utxoSet.GetUTXO(txIn.PreviousOutPoint.Hash, txIn.PreviousOutPoint.Index)
		if err != nil {
			return fmt.Errorf("input not found: %v", err)
		}
		totalInput += utxo.Value
	}

	// 2. Calculate total output value
	var totalOutput int64
	for _, txOut := range tx.TxOut {
		if txOut.Value < 0 {
			return fmt.Errorf("negative output value")
		}
		totalOutput += txOut.Value
	}

	// 3. Check input >= output (difference is fee)
	if totalInput < totalOutput {
		return fmt.Errorf("input value less than output value")
	}

	// 4. Check fee is reasonable
	fee := totalInput - totalOutput
	if fee < 0 {
		return fmt.Errorf("negative fee")
	}

	// Maximum fee check (prevent accidental high fees)
	maxFee := totalOutput / 10 // Max 10% fee
	if fee > maxFee && maxFee > 0 {
		return fmt.Errorf("fee too high: %d (max: %d)", fee, maxFee)
	}

	// 5. Verify signatures for each input
	for i, txIn := range tx.TxIn {
		utxo, _ := utxoSet.GetUTXO(txIn.PreviousOutPoint.Hash, txIn.PreviousOutPoint.Index)

		// Create signature hash
		sigHash := b.calculateSignatureHash(tx, i, utxo.PkScript)

		// Verify signature
		if err := b.verifyInputSignature(txIn, utxo.PkScript, sigHash); err != nil {
			return fmt.Errorf("invalid signature for input %d: %v", i, err)
		}
	}

	return nil
}

// calculateSignatureHash creates the hash to be signed
func (b *BlockChain) calculateSignatureHash(tx *wire.MsgTx, inputIndex int, prevPkScript []byte) []byte {
	// Simplified signature hash calculation
	// In production, implement proper SIGHASH types (ALL, NONE, SINGLE, ANYONECANPAY)

	data := make([]byte, 0, 1024)

	// Version
	data = append(data, byte(tx.Version), byte(tx.Version>>8), byte(tx.Version>>16), byte(tx.Version>>24))

	// Number of inputs
	data = append(data, byte(len(tx.TxIn)))

	// Inputs (with script from UTXO for the input being signed)
	for i, txIn := range tx.TxIn {
		data = append(data, txIn.PreviousOutPoint.Hash[:]...)
		data = append(data, byte(txIn.PreviousOutPoint.Index), byte(txIn.PreviousOutPoint.Index>>8),
			byte(txIn.PreviousOutPoint.Index>>16), byte(txIn.PreviousOutPoint.Index>>24))

		if i == inputIndex {
			// Use the previous output's script for this input
			data = append(data, prevPkScript...)
		} else {
			// Empty script for other inputs
			data = append(data, 0)
		}

		data = append(data, byte(txIn.Sequence), byte(txIn.Sequence>>8),
			byte(txIn.Sequence>>16), byte(txIn.Sequence>>24))
	}

	// Outputs
	data = append(data, byte(len(tx.TxOut)))
	for _, txOut := range tx.TxOut {
		data = append(data, byte(txOut.Value), byte(txOut.Value>>8), byte(txOut.Value>>16), byte(txOut.Value>>24),
			byte(txOut.Value>>32), byte(txOut.Value>>40), byte(txOut.Value>>48), byte(txOut.Value>>56))
		data = append(data, txOut.PkScript...)
	}

	// LockTime
	data = append(data, byte(tx.LockTime), byte(tx.LockTime>>8), byte(tx.LockTime>>16), byte(tx.LockTime>>24))

	// SIGHASH type (SIGHASH_ALL = 1)
	data = append(data, 0x01, 0x00, 0x00, 0x00)

	return crypto.Hash256(data)
}

// verifyInputSignature verifies the signature for a transaction input
func (b *BlockChain) verifyInputSignature(txIn *wire.TxIn, prevPkScript []byte, sigHash []byte) error {
	// Parse signature script
	// Format: <signature> <pubkey>
	sig, pubKey, err := parseSignatureScript(txIn.SignatureScript)
	if err != nil {
		return err
	}

	// Verify signature
	if !crypto.Verify(pubKey, sigHash, sig) {
		return fmt.Errorf("signature verification failed")
	}

	// Verify public key hash matches output script
	pubKeyHash := crypto.Hash160(crypto.PublicKeyToBytes(pubKey))
	if err := verifyPubKeyHash(prevPkScript, pubKeyHash); err != nil {
		return err
	}

	return nil
}

// parseSignatureScript extracts signature and public key from signature script
func parseSignatureScript(script []byte) ([]byte, *ecdsa.PublicKey, error) {
	// Simplified parsing
	// In production, implement full script parsing

	if len(script) < 40 {
		return nil, nil, fmt.Errorf("signature script too short")
	}

	// Extract signature length
	sigLen := int(script[0])
	if len(script) < sigLen+1 {
		return nil, nil, fmt.Errorf("invalid signature length")
	}

	signature := script[1 : sigLen+1]

	// Extract public key
	pubKeyStart := sigLen + 1
	if len(script) < pubKeyStart+1 {
		return nil, nil, fmt.Errorf("missing public key")
	}

	pubKeyLen := int(script[pubKeyStart])
	if len(script) < pubKeyStart+pubKeyLen+1 {
		return nil, nil, fmt.Errorf("invalid public key length")
	}

	pubKeyBytes := script[pubKeyStart+1 : pubKeyStart+pubKeyLen+1]

	pubKey, err := crypto.BytesToPublicKey(pubKeyBytes)
	if err != nil {
		return nil, nil, err
	}

	return signature, pubKey, nil
}

// verifyPubKeyHash verifies the public key hash matches the script
func verifyPubKeyHash(script []byte, pubKeyHash []byte) error {
	// Simplified P2PKH verification
	// Format: OP_DUP OP_HASH160 <pubKeyHash> OP_EQUALVERIFY OP_CHECKSIG

	if len(script) < 25 {
		return fmt.Errorf("invalid script length")
	}

	// Extract hash from script (simplified)
	scriptHash := script[3:23]

	// Compare hashes
	for i := 0; i < 20; i++ {
		if scriptHash[i] != pubKeyHash[i] {
			return fmt.Errorf("public key hash mismatch")
		}
	}

	return nil
}

// CreateP2PKHScript creates a Pay-to-PubKey-Hash script
func CreateP2PKHScript(pubKeyHash []byte) []byte {
	// OP_DUP OP_HASH160 <pubKeyHash> OP_EQUALVERIFY OP_CHECKSIG
	script := make([]byte, 25)
	script[0] = 0x76 // OP_DUP
	script[1] = 0xa9 // OP_HASH160
	script[2] = 0x14 // Push 20 bytes
	copy(script[3:23], pubKeyHash)
	script[23] = 0x88 // OP_EQUALVERIFY
	script[24] = 0xac // OP_CHECKSIG
	return script
}

// SignTransaction signs all inputs of a transaction
func (b *BlockChain) SignTransaction(tx *wire.MsgTx, privateKey *ecdsa.PrivateKey, utxoSet *UTXOSet) error {
	if tx.IsCoinbase() {
		return nil
	}

	for i, txIn := range tx.TxIn {
		// Get the UTXO being spent
		utxo, err := utxoSet.GetUTXO(txIn.PreviousOutPoint.Hash, txIn.PreviousOutPoint.Index)
		if err != nil {
			return fmt.Errorf("UTXO not found for input %d: %v", i, err)
		}

		// Calculate signature hash
		sigHash := b.calculateSignatureHash(tx, i, utxo.PkScript)

		// Sign the hash
		signature, err := crypto.Sign(privateKey, sigHash)
		if err != nil {
			return fmt.Errorf("failed to sign input %d: %v", i, err)
		}

		// Append SIGHASH type
		signature = append(signature, 0x01) // SIGHASH_ALL

		// Create signature script: <signature> <pubkey>
		pubKey := &privateKey.PublicKey
		pubKeyBytes := crypto.PublicKeyToBytes(pubKey)

		sigScript := make([]byte, 0, len(signature)+len(pubKeyBytes)+2)
		sigScript = append(sigScript, byte(len(signature)))
		sigScript = append(sigScript, signature...)
		sigScript = append(sigScript, byte(len(pubKeyBytes)))
		sigScript = append(sigScript, pubKeyBytes...)

		txIn.SignatureScript = sigScript
	}

	return nil
}

// CalculateTransactionFee calculates the fee for a transaction
func (b *BlockChain) CalculateTransactionFee(tx *wire.MsgTx, utxoSet *UTXOSet) (int64, error) {
	if tx.IsCoinbase() {
		return 0, nil
	}

	var totalInput int64
	for _, txIn := range tx.TxIn {
		utxo, err := utxoSet.GetUTXO(txIn.PreviousOutPoint.Hash, txIn.PreviousOutPoint.Index)
		if err != nil {
			return 0, err
		}
		totalInput += utxo.Value
	}

	var totalOutput int64
	for _, txOut := range tx.TxOut {
		totalOutput += txOut.Value
	}

	return totalInput - totalOutput, nil
}

// validateTokenIssueTransaction validates a token issuance transaction
func (b *BlockChain) validateTokenIssueTransaction(tx *wire.MsgTx) error {
	// Token issuance should have no inputs (free operation)
	if len(tx.TxIn) > 0 {
		return fmt.Errorf("token issuance should not have inputs")
	}

	// Should have exactly one output
	if len(tx.TxOut) != 1 {
		return fmt.Errorf("token issuance should have exactly one output")
	}

	// Parse token data from memo
	if len(tx.Memo) == 0 {
		return fmt.Errorf("token issuance missing memo data")
	}

	// Basic memo validation (simplified)
	// In production, implement proper deserialization
	memoStr := string(tx.Memo)
	parts := strings.Split(memoStr, "|")
	if len(parts) != 4 {
		return fmt.Errorf("invalid token issuance memo format")
	}

	name := parts[0]
	symbol := parts[1]

	// Validate token parameters
	if len(name) == 0 || len(name) > 32 {
		return fmt.Errorf("invalid token name length")
	}
	if len(symbol) == 0 || len(symbol) > 8 {
		return fmt.Errorf("invalid token symbol length")
	}

	// Check if symbol already exists
	if _, err := b.tokenStore.GetTokenBySymbol(symbol); err == nil {
		return fmt.Errorf("token symbol %s already exists", symbol)
	}

	return nil
}

// validateTokenTransferTransaction validates a token transfer transaction
func (b *BlockChain) validateTokenTransferTransaction(tx *wire.MsgTx, utxoSet *UTXOSet) error {
	// Should have at least one input (for fee payment)
	if len(tx.TxIn) == 0 {
		return fmt.Errorf("token transfer should have at least one input for fee")
	}

	// Should have at least one output (fee payment)
	if len(tx.TxOut) == 0 {
		return fmt.Errorf("token transfer should have at least one output for fee")
	}

	// Parse token transfer data from memo
	if len(tx.Memo) < 32 { // At least token ID (32 bytes)
		return fmt.Errorf("token transfer missing memo data")
	}

	// Extract token ID (first 32 bytes)
	tokenID := wire.Hash{}
	copy(tokenID[:], tx.Memo[:32])

	// Parse transfer data (simplified)
	// Format: tokenID + "|" + from + "|" + to + "|" + amount
	memoStr := string(tx.Memo[32:])
	parts := strings.Split(memoStr, "|")
	if len(parts) != 4 {
		return fmt.Errorf("invalid token transfer memo format")
	}

	from := parts[1]
	amountStr := parts[3]

	// Parse amount
	amount, err := strconv.ParseInt(amountStr, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid transfer amount: %v", err)
	}

	if amount <= 0 {
		return fmt.Errorf("transfer amount must be positive")
	}

	// Check if token exists
	_, err = b.tokenStore.GetToken(tokenID)
	if err != nil {
		return fmt.Errorf("token does not exist: %v", err)
	}

	// Check sender balance
	senderBalance := b.tokenStore.GetBalance(from, tokenID)
	if senderBalance < amount {
		return fmt.Errorf("insufficient token balance: has %d, need %d", senderBalance, amount)
	}

	// Validate fee payment (standard OB transaction validation)
	return b.validateStandardTransaction(tx, utxoSet)
}

// validateStandardTransaction validates a standard (non-token) transaction
func (b *BlockChain) validateStandardTransaction(tx *wire.MsgTx, utxoSet *UTXOSet) error {
	// 1. Check inputs exist and are unspent
	var totalInput int64
	for _, txIn := range tx.TxIn {
		utxo, err := utxoSet.GetUTXO(txIn.PreviousOutPoint.Hash, txIn.PreviousOutPoint.Index)
		if err != nil {
			return fmt.Errorf("input not found: %v", err)
		}
		totalInput += utxo.Value
	}

	// 2. Calculate total output value
	var totalOutput int64
	for _, txOut := range tx.TxOut {
		if txOut.Value < 0 {
			return fmt.Errorf("negative output value")
		}
		totalOutput += txOut.Value
	}

	// 3. Check input >= output (difference is fee)
	if totalInput < totalOutput {
		return fmt.Errorf("input value less than output value")
	}

	// 4. Check fee is reasonable
	fee := totalInput - totalOutput
	if fee < 0 {
		return fmt.Errorf("negative fee")
	}

	// Maximum fee check (prevent accidental high fees)
	maxFee := totalOutput / 10 // Max 10% fee
	if fee > maxFee && maxFee > 0 {
		return fmt.Errorf("fee too high: %d (max: %d)", fee, maxFee)
	}

	// 5. Verify signatures for each input
	for i, txIn := range tx.TxIn {
		utxo, _ := utxoSet.GetUTXO(txIn.PreviousOutPoint.Hash, txIn.PreviousOutPoint.Index)

		// Create signature hash
		sigHash := b.calculateSignatureHash(tx, i, utxo.PkScript)

		// Verify signature
		if err := b.verifyInputSignature(txIn, utxo.PkScript, sigHash); err != nil {
			return fmt.Errorf("invalid signature for input %d: %v", i, err)
		}
	}

	return nil
}
