package rpcserver

// JSONRPCRequest represents a JSON-RPC 2.0 request.
type JSONRPCRequest struct {
	JSONRPC string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
	ID      interface{}   `json:"id"`
}

// JSONRPCResponse represents a JSON-RPC 2.0 response.
type JSONRPCResponse struct {
	JSONRPC string        `json:"jsonrpc"`
	Result  interface{}   `json:"result,omitempty"`
	Error   *JSONRPCError `json:"error,omitempty"`
	ID      interface{}   `json:"id"`
}

// JSONRPCError represents a JSON-RPC 2.0 error.
type JSONRPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// BlockInfo represents block information for RPC responses.
type BlockInfo struct {
	Hash         string `json:"hash"`
	Height       int32  `json:"height"`
	Version      int32  `json:"version"`
	PrevBlock    string `json:"previousblockhash"`
	MerkleRoot   string `json:"merkleroot"`
	Timestamp    int64  `json:"time"`
	Bits         uint32 `json:"bits"`
	Nonce        uint32 `json:"nonce"`
	Transactions int    `json:"tx_count"`
}

// BlockchainInfo represents blockchain information.
type BlockchainInfo struct {
	Chain         string `json:"chain"`
	Blocks        int32  `json:"blocks"`
	BestBlockHash string `json:"bestblockhash"`
	Difficulty    uint32 `json:"difficulty"`
	MaxMoney      int64  `json:"maxmoney"`
	InitialSupply int64  `json:"initialsupply"`
}

// MiningInfo represents mining information.
type MiningInfo struct {
	Blocks       int32  `json:"blocks"`
	CurrentHash  string `json:"currentblockhash"`
	Difficulty   uint32 `json:"difficulty"`
	MiningActive bool   `json:"generate"`
	HashesPerSec int64  `json:"hashespersec"`
	BlockReward  int64  `json:"blockreward"`
}

// ShieldedRecipient represents a recipient in a shielded transaction.
type ShieldedRecipient struct {
	Address string `json:"address"`
	Amount  int64  `json:"amount"`
	Memo    string `json:"memo,omitempty"`
}

// ShieldedTxInfo represents information about a shielded transaction.
type ShieldedTxInfo struct {
	TxID      string `json:"txid"`
	Amount    int64  `json:"amount"`
	Memo      string `json:"memo,omitempty"`
	Confirmed bool   `json:"confirmed"`
	BlockHash string `json:"blockhash,omitempty"`
	Time      int64  `json:"time"`
}

// MultiSigInfo represents information about a multisig address.
type MultiSigInfo struct {
	Address      string   `json:"address"`
	RedeemScript string   `json:"redeemscript"`
	M            int      `json:"m"`
	N            int      `json:"n"`
	PublicKeys   []string `json:"pubkeys"`
}

// MultiSigTx represents a multisig transaction.
type MultiSigTx struct {
	TxID        string   `json:"txid"`
	Hex         string   `json:"hex"`
	Complete    bool     `json:"complete"`
	MissingSigs int      `json:"missing_sigs"`
	Signatures  []string `json:"signatures,omitempty"`
}

// MultiSigSignature represents a signature for multisig.
type MultiSigSignature struct {
	PublicKey string `json:"pubkey"`
	Signature string `json:"signature"`
}

// HDWalletInfo represents HD wallet information.
type HDWalletInfo struct {
	MasterFingerprint string            `json:"master_fingerprint"`
	Addresses         map[string]string `json:"addresses"` // path -> address
	MiningAddress     string            `json:"mining_address"`
	SeedPhrase        string            `json:"seed_phrase,omitempty"` // Only for creation response
}
