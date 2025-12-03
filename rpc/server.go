package rpc

import (
	"encoding/json"
	"fmt"
	"net/http"
	"obsidian-core/blockchain"
	"obsidian-core/mining"
	"strconv"
	"time"
)

// PoolServer interface for pool statistics
type PoolServer interface {
	GetStats() map[string]interface{}
}

// Wallet interface for shielded and transparent operations
type Wallet interface {
	// Transparent operations
	GetNewAddress() (string, error)
	GetBalance(address string) (int64, error)
	SendToAddress(from, to string, amount int64) (string, error)
	ListAddresses() []string

	// Shielded operations
	NewShieldedAddress() (string, error)
	ListShieldedAddresses() []string
	GetShieldedBalance(address string) (int64, error)
	SendShielded(from string, recipients []ShieldedRecipient) (string, error)
	ListReceivedShielded(address string) ([]ShieldedTxInfo, error)
	GetTransparentBalance() int64
	GetTotalShieldedBalance() int64
	ExportViewingKey(address string) (string, error)
	ImportViewingKey(key string) error
	ShieldCoinbase(toAddress string) (string, error)

	// Multisig operations
	CreateMultiSigAddress(nRequired int, publicKeys []string) (*MultiSigInfo, error)
	AddMultiSigAddress(nRequired int, publicKeys []string, account string) (string, error)
	SignMultiSigTx(txHex string, redeemScript string, privateKeys []string) (*MultiSigTx, error)
	CombineMultiSigSigs(txHex string, signatures []MultiSigSignature) (string, error)
}

// SimpleWallet is a simple implementation of Wallet interface
type SimpleWallet struct{}

func (w *SimpleWallet) GetNewAddress() (string, error) {
	return "Obs_demo_address_" + fmt.Sprintf("%d", 12345), nil
}

func (w *SimpleWallet) GetBalance(address string) (int64, error) {
	return 0, nil
}

func (w *SimpleWallet) SendToAddress(from, to string, amount int64) (string, error) {
	return "demo_txid_" + from + "_to_" + to, nil
}

func (w *SimpleWallet) ListAddresses() []string {
	return []string{}
}

func (w *SimpleWallet) NewShieldedAddress() (string, error) {
	return "zobs_demo_address_" + fmt.Sprintf("%d", 12345), nil
}

func (w *SimpleWallet) ListShieldedAddresses() []string {
	return []string{}
}

func (w *SimpleWallet) GetShieldedBalance(address string) (int64, error) {
	return 0, nil
}

func (w *SimpleWallet) SendShielded(from string, recipients []ShieldedRecipient) (string, error) {
	return "demo_txid_" + from, nil
}

func (w *SimpleWallet) ListReceivedShielded(address string) ([]ShieldedTxInfo, error) {
	return []ShieldedTxInfo{}, nil
}

func (w *SimpleWallet) GetTransparentBalance() int64 {
	return 0
}

func (w *SimpleWallet) GetTotalShieldedBalance() int64 {
	return 0
}

func (w *SimpleWallet) ExportViewingKey(address string) (string, error) {
	return "demo_viewing_key", nil
}

func (w *SimpleWallet) ImportViewingKey(key string) error {
	return nil
}

func (w *SimpleWallet) ShieldCoinbase(toAddress string) (string, error) {
	return "demo_shield_txid", nil
}

func (w *SimpleWallet) CreateMultiSigAddress(nRequired int, publicKeys []string) (*MultiSigInfo, error) {
	// Demo implementation
	redeemScript := fmt.Sprintf("multisig_%d_of_%d", nRequired, len(publicKeys))
	address := fmt.Sprintf("multi_%s", redeemScript)

	return &MultiSigInfo{
		Address:      address,
		RedeemScript: redeemScript,
		M:            nRequired,
		N:            len(publicKeys),
		PublicKeys:   publicKeys,
	}, nil
}

func (w *SimpleWallet) AddMultiSigAddress(nRequired int, publicKeys []string, account string) (string, error) {
	info, err := w.CreateMultiSigAddress(nRequired, publicKeys)
	if err != nil {
		return "", err
	}
	return info.Address, nil
}

func (w *SimpleWallet) SignMultiSigTx(txHex string, redeemScript string, privateKeys []string) (*MultiSigTx, error) {
	// Demo implementation - in real implementation, this would sign the transaction
	sigs := make([]string, len(privateKeys))
	for i := range privateKeys {
		sigs[i] = fmt.Sprintf("sig_%d", i+1)
	}

	return &MultiSigTx{
		TxID:        "demo_multisig_txid",
		Hex:         txHex,
		Complete:    len(privateKeys) >= 2, // Demo: assume 2-of-n
		MissingSigs: 2 - len(privateKeys),  // Demo
		Signatures:  sigs,
	}, nil
}

func (w *SimpleWallet) CombineMultiSigSigs(txHex string, signatures []MultiSigSignature) (string, error) {
	// Demo implementation
	return fmt.Sprintf("combined_tx_%s", txHex[:16]), nil
}

// Server represents the RPC server.
type Server struct {
	chain       *blockchain.BlockChain
	miner       *mining.CPUMiner
	pool        PoolServer
	wallet      Wallet
	syncManager interface{} // For peer info, will be *network.SyncManager
	addr        string
	server      *http.Server
}

// NewServer creates a new RPC server.
func NewServer(chain *blockchain.BlockChain, miner *mining.CPUMiner, syncManager interface{}, addr string) *Server {
	return &Server{
		chain:       chain,
		miner:       miner,
		pool:        nil,             // Pool is optional
		wallet:      &SimpleWallet{}, // Use simple wallet for now
		syncManager: syncManager,
		addr:        addr,
	}
}

// SetPoolServer sets the pool server for statistics
func (s *Server) SetPoolServer(pool PoolServer) {
	s.pool = pool
}

// Start starts the RPC server.
func (s *Server) Start() error {
	mux := http.NewServeMux()

	// RPC methods
	mux.HandleFunc("/", s.handleRequest)

	// Health check
	mux.HandleFunc("/health", s.healthHandler)

	// Metrics
	mux.HandleFunc("/metrics", s.metricsHandler)

	server := &http.Server{
		Addr:    s.addr,
		Handler: s.enableCORS(mux),
	}

	s.server = server

	fmt.Printf("RPC server listening on %s\n", s.addr)
	return server.ListenAndServe()
}

// healthHandler handles health check requests
func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	timestamp := time.Now().Format(time.RFC3339)
	w.Write([]byte(fmt.Sprintf(`{"status":"ok","timestamp":"%s"}`, timestamp)))
}

// metricsHandler handles metrics requests
func (s *Server) metricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	peerCount := 0
	if s.syncManager != nil {
		if sm, ok := s.syncManager.(interface{ GetPeerCount() int }); ok {
			peerCount = sm.GetPeerCount()
		}
	}

	jsonStr := `{"peers":` + strconv.Itoa(peerCount) + `,"height":` + strconv.Itoa(int(s.chain.Height())) + `}`
	w.Write([]byte(jsonStr))
}

// Stop stops the RPC server.
func (s *Server) Stop() error {
	if s.server != nil {
		return s.server.Close()
	}
	return nil
}

// enableCORS adds CORS headers to allow browser access.
func (s *Server) enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// handleRequest handles JSON-RPC requests.
func (s *Server) handleRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req JSONRPCRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, nil, -32700, "Parse error", nil)
		return
	}

	result, err := s.handleMethod(&req)
	if err != nil {
		sendError(w, req.ID, -32603, err.Error(), nil)
		return
	}

	sendResponse(w, req.ID, result)
}

// handleMethod routes the request to the appropriate handler.
func (s *Server) handleMethod(req *JSONRPCRequest) (interface{}, error) {
	switch req.Method {
	// Blockchain methods
	case "getblockcount":
		return s.getBlockCount(req.Params)
	case "getbestblockhash":
		return s.getBestBlockHash(req.Params)
	case "getblock":
		return s.getBlock(req.Params)
	case "getblockchaininfo":
		return s.getBlockchainInfo(req.Params)
	case "getmininginfo":
		return s.getMiningInfo(req.Params)
	case "getblockreward":
		return s.getBlockReward(req.Params)
	case "estimatefee":
		return s.estimateFee(req.Params)

	// Wallet methods
	case "getnewaddress":
		return s.getnewaddress(req.Params)
	case "getbalance":
		return s.getbalance(req.Params)
	case "sendtoaddress":
		return s.sendtoaddress(req.Params)
	case "listaddresses":
		return s.listaddresses(req.Params)

	// Multisig methods
	case "createmultisig":
		return s.createmultisig(req.Params)
	case "addmultisigaddress":
		return s.addmultisigaddress(req.Params)
	case "signmultisigtx":
		return s.signmultisigtx(req.Params)
	case "combinemultisigsigs":
		return s.combinemultisigsigs(req.Params)

	// Shielded transaction methods (Zcash-style)
	case "z_getnewaddress":
		return s.z_getnewaddress(req.Params)
	case "z_listaddresses":
		return s.z_listaddresses(req.Params)
	case "z_getbalance":
		return s.z_getbalance(req.Params)
	case "z_sendmany":
		return s.z_sendmany(req.Params)
	case "z_listreceivedbyaddress":
		return s.z_listreceivedbyaddress(req.Params)
	case "z_gettotalbalance":
		return s.z_gettotalbalance(req.Params)
	case "z_exportviewingkey":
		return s.z_exportviewingkey(req.Params)
	case "z_importviewingkey":
		return s.z_importviewingkey(req.Params)
	case "z_shieldcoinbase":
		return s.z_shieldcoinbase(req.Params)

	// Mining Pool
	case "getpoolinfo":
		return s.getpoolinfo(req.Params, s.wallet)

	// Token methods
	case "issuetoken":
		return s.issueToken(req.Params)
	case "transfertoken":
		return s.transferToken(req.Params)
	case "minttoken":
		return s.minttoken(req.Params)
	case "burntoken":
		return s.burntoken(req.Params)
	case "transfertokenownership":
		return s.transfertokenownership(req.Params)
	case "shieldtoken":
		return s.shieldtoken(req.Params)
	case "gettokenbalance":
		return s.getTokenBalance(req.Params)
	case "gettokeninfo":
		return s.getTokenInfo(req.Params)
	case "listtokens":
		return s.listTokens(req.Params)
	case "getaddresstokens":
		return s.getAddressTokens(req.Params)

	// Utility
	case "ping":
		return "pong", nil
	default:
		return nil, fmt.Errorf("method not found: %s", req.Method)
	}
}

// sendResponse sends a successful JSON-RPC response.
func sendResponse(w http.ResponseWriter, id interface{}, result interface{}) {
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// sendError sends an error JSON-RPC response.
func sendError(w http.ResponseWriter, id interface{}, code int, message string, data interface{}) {
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &JSONRPCError{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
