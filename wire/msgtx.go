package wire

// TxVersion defines the version of the transaction.
const TxVersion = 1

// OutPoint defines a bitcoin data type that is used to track previous
// transaction outputs.
type OutPoint struct {
	Hash  Hash
	Index uint32
}

// TxIn defines a bitcoin transaction input.
type TxIn struct {
	PreviousOutPoint OutPoint
	SignatureScript  []byte
	Sequence         uint32
}

// TxOut defines a bitcoin transaction output.
type TxOut struct {
	Value    int64
	PkScript []byte
}

// TxType defines the type of transaction
type TxType uint8

const (
	TxTypeTransparent            TxType = 0 // Standard transparent transaction
	TxTypeShielded               TxType = 1 // Shielded transaction (private)
	TxTypeMixed                  TxType = 2 // Mixed (t-addr to z-addr or vice versa)
	TxTypeTokenIssue             TxType = 3 // Token issuance transaction
	TxTypeTokenTransfer          TxType = 4 // Token transfer transaction
	TxTypeTokenMint              TxType = 5 // Token minting transaction
	TxTypeTokenTransferOwnership TxType = 6 // Token ownership transfer transaction
	TxTypeTokenShielded          TxType = 7 // Token shielded transaction
	TxTypeTokenBurn              TxType = 8 // Token burning transaction
)

// ShieldedSpend represents a shielded input (spending from shielded pool)
type ShieldedSpend struct {
	Cv           []byte // Value commitment
	Anchor       []byte // Root of note commitment tree
	Nullifier    []byte // Unique nullifier (prevents double-spend)
	Rk           []byte // Randomized public key
	Proof        []byte // zk-SNARK proof
	SpendAuthSig []byte // Spend authorization signature
	TokenID      Hash   // Token identifier (zero hash for OB)
	TokenAmount  int64  // Token amount (0 for OB)
}

// ShieldedOutput represents a shielded output (adding to shielded pool)
type ShieldedOutput struct {
	Cv            []byte // Value commitment
	Cmu           []byte // Note commitment
	EphemeralKey  []byte // Ephemeral public key
	EncCiphertext []byte // Encrypted note ciphertext (580 bytes)
	OutCiphertext []byte // Encrypted outgoing ciphertext (80 bytes)
	Proof         []byte // zk-SNARK proof
	Memo          []byte // Encrypted memo (512 bytes)
	TokenID       Hash   // Token identifier (zero hash for OB)
	TokenAmount   int64  // Token amount (0 for OB)
}

// MsgTx implements the Message interface and represents a bitcoin tx message.
// It is used to deliver transaction information in response to a getdata
// message (MsgGetData) for a given transaction.
//
// Obsidian (Zcash-based) extensions:
// - Shielded transactions (zk-SNARKs)
// - Encrypted memos
// - Value privacy
type MsgTx struct {
	Version  int32
	TxIn     []*TxIn
	TxOut    []*TxOut
	LockTime uint32

	// Zcash/Obsidian specific fields
	TxType          TxType            // Transaction type
	ExpiryHeight    uint32            // Block height after which tx expires
	ValueBalance    int64             // Net value balance (shielded - transparent)
	ShieldedSpends  []*ShieldedSpend  // Shielded inputs
	ShieldedOutputs []*ShieldedOutput // Shielded outputs
	BindingSig      []byte            // Binding signature (proves value balance)

	// Transparent transaction memo (optional, for t-addr txs)
	Memo []byte // Up to 512 bytes (encrypted in shielded txs)
}

// AddTxIn adds a transaction input to the message.
func (msg *MsgTx) AddTxIn(ti *TxIn) {
	msg.TxIn = append(msg.TxIn, ti)
}

// AddTxOut adds a transaction output to the message.
func (msg *MsgTx) AddTxOut(to *TxOut) {
	msg.TxOut = append(msg.TxOut, to)
}

// NewMsgTx returns a new bitcoin tx message that conforms to the Message
// interface.  The return instance has a default version of TxVersion and there
// are no transaction inputs or outputs.
func NewMsgTx(version int32) *MsgTx {
	return &MsgTx{
		Version:         version,
		TxType:          TxTypeTransparent,
		TxIn:            make([]*TxIn, 0, 8),
		TxOut:           make([]*TxOut, 0, 8),
		ShieldedSpends:  make([]*ShieldedSpend, 0),
		ShieldedOutputs: make([]*ShieldedOutput, 0),
	}
}

// NewShieldedTx creates a new shielded transaction
func NewShieldedTx(version int32) *MsgTx {
	tx := NewMsgTx(version)
	tx.TxType = TxTypeShielded
	return tx
}

// AddShieldedSpend adds a shielded input to the transaction
func (msg *MsgTx) AddShieldedSpend(spend *ShieldedSpend) {
	msg.ShieldedSpends = append(msg.ShieldedSpends, spend)
}

// AddShieldedOutput adds a shielded output to the transaction
func (msg *MsgTx) AddShieldedOutput(output *ShieldedOutput) {
	msg.ShieldedOutputs = append(msg.ShieldedOutputs, output)
}

// IsShielded returns true if transaction contains any shielded components
func (msg *MsgTx) IsShielded() bool {
	return msg.TxType == TxTypeShielded || msg.TxType == TxTypeMixed ||
		len(msg.ShieldedSpends) > 0 || len(msg.ShieldedOutputs) > 0
}

// SetMemo sets the memo field (max 512 bytes)
func (msg *MsgTx) SetMemo(memo []byte) error {
	if len(memo) > 512 {
		return ErrMemoTooLarge
	}
	msg.Memo = memo
	return nil
}

// IsCoinbase returns true if this is a coinbase transaction
func (msg *MsgTx) IsCoinbase() bool {
	return len(msg.TxIn) == 1 &&
		msg.TxIn[0].PreviousOutPoint.Index == 0xffffffff &&
		msg.TxIn[0].PreviousOutPoint.Hash == (Hash{})
}

// TxHash generates the hash for the transaction
func (msg *MsgTx) TxHash() Hash {
	// Simplified hash calculation
	// In production, this should serialize the entire transaction and hash it
	data := make([]byte, 0, 1024)

	// Version
	data = append(data, byte(msg.Version), byte(msg.Version>>8), byte(msg.Version>>16), byte(msg.Version>>24))

	// Input count
	data = append(data, byte(len(msg.TxIn)))

	// Inputs
	for _, txIn := range msg.TxIn {
		data = append(data, txIn.PreviousOutPoint.Hash[:]...)
		data = append(data, byte(txIn.PreviousOutPoint.Index), byte(txIn.PreviousOutPoint.Index>>8),
			byte(txIn.PreviousOutPoint.Index>>16), byte(txIn.PreviousOutPoint.Index>>24))
		data = append(data, txIn.SignatureScript...)
	}

	// Output count
	data = append(data, byte(len(msg.TxOut)))

	// Outputs
	for _, txOut := range msg.TxOut {
		data = append(data, byte(txOut.Value), byte(txOut.Value>>8), byte(txOut.Value>>16), byte(txOut.Value>>24),
			byte(txOut.Value>>32), byte(txOut.Value>>40), byte(txOut.Value>>48), byte(txOut.Value>>56))
		data = append(data, txOut.PkScript...)
	}

	// LockTime
	data = append(data, byte(msg.LockTime), byte(msg.LockTime>>8), byte(msg.LockTime>>16), byte(msg.LockTime>>24))

	return DoubleHashH(data)
}

// NewCoinbaseTx creates a coinbase transaction for the given height and reward.
// The minerAddress is a simplified representation (in real implementation, use proper address format).
func NewCoinbaseTx(height int32, reward int64, minerAddress string) *MsgTx {
	// Create coinbase input with block height in signature script
	coinbaseScript := []byte{byte(height >> 8), byte(height & 0xff)}
	txIn := &TxIn{
		PreviousOutPoint: OutPoint{
			Hash:  Hash{}, // Null hash for coinbase
			Index: 0xffffffff,
		},
		SignatureScript: coinbaseScript,
		Sequence:        0xffffffff,
	}

	// Create output with reward to miner
	txOut := &TxOut{
		Value:    reward,
		PkScript: []byte(minerAddress), // Simplified - in production use proper script
	}

	tx := NewMsgTx(TxVersion)
	tx.AddTxIn(txIn)
	tx.AddTxOut(txOut)

	return tx
}

// Token represents a custom token on the Obsidian network
type Token struct {
	ID          Hash   // Unique token identifier (hash of issuance tx)
	Name        string // Token name (max 32 chars)
	Symbol      string // Token symbol (max 8 chars)
	Decimals    uint8  // Decimal places (0-18)
	TotalSupply int64  // Total supply
	Owner       string // Token owner address
	Mintable    bool   // Whether additional tokens can be minted
	Created     int64  // Creation timestamp
}

// TokenIssue represents a token issuance transaction
type TokenIssue struct {
	Name     string // Token name
	Symbol   string // Token symbol
	Decimals uint8  // Decimal places
	Supply   int64  // Initial supply
	Mintable bool   // Whether additional tokens can be minted
	Owner    string // Owner address
}

// TokenTransfer represents a token transfer transaction
type TokenTransfer struct {
	TokenID Hash   // Token identifier
	From    string // Sender address
	To      string // Recipient address
	Amount  int64  // Transfer amount
}

// TokenMint represents a token minting transaction
type TokenMint struct {
	TokenID Hash   // Token identifier
	Amount  int64  // Amount to mint
	To      string // Recipient address
}

// TokenTransferOwnership represents a token ownership transfer transaction
type TokenTransferOwnership struct {
	TokenID  Hash   // Token identifier
	NewOwner string // New owner address
}

// TokenBurn represents a token burning transaction
type TokenBurn struct {
	TokenID Hash   // Token identifier
	Amount  int64  // Amount to burn
	From    string // Address burning tokens
}

// TokenShielded represents a token shielded transaction
type TokenShielded struct {
	TokenID Hash   // Token identifier
	Amount  int64  // Amount to shield/unshield
	To      string // Recipient z-address (shield) or t-address (unshield)
}

// NewTokenIssueTx creates a new token issuance transaction
func NewTokenIssueTx(issuer string, tokenIssue *TokenIssue) *MsgTx {
	tx := NewMsgTx(TxVersion)
	tx.TxType = TxTypeTokenIssue

	// Add token issuance data to memo field (simplified)
	// In production, this should be properly serialized
	tokenData := []byte(tokenIssue.Name + "|" + tokenIssue.Symbol + "|" +
		string(rune(tokenIssue.Decimals)) + "|" + string(rune(tokenIssue.Supply)))
	tx.Memo = tokenData

	// Add a dummy output to make it a valid transaction
	txOut := &TxOut{
		Value:    0, // Token issuance doesn't require OB payment
		PkScript: []byte(issuer),
	}
	tx.AddTxOut(txOut)

	return tx
}

// NewTokenTransferTx creates a new token transfer transaction
func NewTokenTransferTx(from, to string, tokenID Hash, amount int64) *MsgTx {
	tx := NewMsgTx(TxVersion)
	tx.TxType = TxTypeTokenTransfer

	// Add token transfer data to memo field (simplified)
	transferData := append(tokenID[:], []byte("|"+from+"|"+to+"|"+string(rune(amount)))...)
	tx.Memo = transferData

	// Add a minimal OB output for transaction fee
	txOut := &TxOut{
		Value:    10000, // 0.0001 OB for fee
		PkScript: []byte(from),
	}
	tx.AddTxOut(txOut)

	return tx
}

// NewTokenShieldedTx creates a new token shielded transaction
func NewTokenShieldedTx(from, to string, tokenID Hash, amount int64) *MsgTx {
	tx := NewShieldedTx(TxVersion)
	tx.TxType = TxTypeTokenShielded

	// Add token shielded data to memo field (simplified)
	shieldedData := append(tokenID[:], []byte("|"+from+"|"+to+"|"+string(rune(amount)))...)
	tx.Memo = shieldedData

	// Add a minimal OB output for transaction fee
	txOut := &TxOut{
		Value:    10000, // 0.0001 OB for fee
		PkScript: []byte(from),
	}
	tx.AddTxOut(txOut)

	return tx
}

// NewTokenBurnTx creates a new token burning transaction
func NewTokenBurnTx(from string, tokenID Hash, amount int64) *MsgTx {
	tx := NewMsgTx(TxVersion)
	tx.TxType = TxTypeTokenBurn

	// Add token burn data to memo field (simplified)
	burnData := append(tokenID[:], []byte("|"+from+"|"+string(rune(amount)))...)
	tx.Memo = burnData

	// Add a minimal OB output for transaction fee
	txOut := &TxOut{
		Value:    10000, // 0.0001 OB for fee
		PkScript: []byte(from),
	}
	tx.AddTxOut(txOut)

	return tx
}
// NewTokenMintTx creates a new token minting transaction
func NewTokenMintTx(from, to string, tokenID Hash, amount int64) *MsgTx {
	tx := NewMsgTx(TxVersion)
	tx.TxType = TxTypeTokenMint

	// Add token mint data to memo field (simplified)
	mintData := append(tokenID[:], []byte("|"+string(rune(amount))+"|"+to+"|"+from)...)
	tx.Memo = mintData

	// Add a minimal OB output for transaction fee
	txOut := &TxOut{
		Value:    10000, // 0.0001 OB for fee
		PkScript: []byte(from),
	}
	tx.AddTxOut(txOut)

	return tx
}

// NewTokenTransferOwnershipTx creates a new token ownership transfer transaction
func NewTokenTransferOwnershipTx(from, newOwner string, tokenID Hash) *MsgTx {
	tx := NewMsgTx(TxVersion)
	tx.TxType = TxTypeTokenTransferOwnership

	// Add token ownership transfer data to memo field
	transferData := append(tokenID[:], []byte(newOwner)...)
	tx.Memo = transferData

	// Add a minimal OB output for transaction fee
	txOut := &TxOut{
		Value:    10000, // 0.0001 OB for fee
		PkScript: []byte(from),
	}
	tx.AddTxOut(txOut)

	return tx
}
