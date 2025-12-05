package stratum

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"obsidian-core/blockchain"
	"obsidian-core/chaincfg"
	"obsidian-core/consensus"
	"sync"
	"time"
)

// StratumPool implements a Stratum mining pool server
type StratumPool struct {
	chain        *blockchain.BlockChain
	params       *chaincfg.Params
	pow          consensus.PowEngine
	poolAddress  string
	listenAddr   string
	difficulty   float64
	listener     net.Listener
	clients      map[string]*PoolClient
	clientsMutex sync.RWMutex
	currentJob   *MiningJob
	jobMutex     sync.RWMutex
	jobCounter   uint64
	running      bool
}

// PoolClient represents a connected miner
type PoolClient struct {
	conn       net.Conn
	id         string
	authorized bool
	subscribed bool
	difficulty float64
	lastShare  time.Time
	shares     uint64
}

// MiningJob represents work to be done
type MiningJob struct {
	JobID        string
	PrevHash     string
	Coinbase1    string
	Coinbase2    string
	MerkleBranch []string
	Version      string
	NBits        string
	NTime        string
	CleanJobs    bool
	Height       int32
	Target       uint32
}

// StratumRequest represents a Stratum protocol request
type StratumRequest struct {
	ID     interface{}   `json:"id"`
	Method string        `json:"method"`
	Params []interface{} `json:"params"`
}

// StratumResponse represents a Stratum protocol response
type StratumResponse struct {
	ID     interface{} `json:"id"`
	Result interface{} `json:"result"`
	Error  interface{} `json:"error"`
}

// NewStratumPool creates a new Stratum mining pool
func NewStratumPool(chain *blockchain.BlockChain, params *chaincfg.Params, pow consensus.PowEngine, poolAddr string, listenAddr string) *StratumPool {
	return &StratumPool{
		chain:       chain,
		params:      params,
		pow:         pow,
		poolAddress: poolAddr,
		listenAddr:  listenAddr,
		difficulty:  1.0,
		clients:     make(map[string]*PoolClient),
		running:     false,
	}
}

// Start starts the Stratum pool server
func (p *StratumPool) Start() error {
	listener, err := net.Listen("tcp", p.listenAddr)
	if err != nil {
		return fmt.Errorf("failed to start pool listener: %v", err)
	}

	p.listener = listener
	p.running = true

	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("[MINING] Stratum Mining Pool Started\n")
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("  Listen Address: stratum+tcp://%s\n", p.listenAddr)
	fmt.Printf("  Pool Address:   %s\n", p.poolAddress)
	fmt.Printf("  Difficulty:     %.2f\n", p.difficulty)
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")

	// Start job generator
	go p.jobGenerator()

	// Accept connections
	go p.acceptConnections()

	return nil
}

// Stop stops the Stratum pool server
func (p *StratumPool) Stop() {
	p.running = false
	if p.listener != nil {
		p.listener.Close()
	}

	p.clientsMutex.Lock()
	for _, client := range p.clients {
		client.conn.Close()
	}
	p.clients = make(map[string]*PoolClient)
	p.clientsMutex.Unlock()

	fmt.Println("Stratum pool stopped")
}

// acceptConnections accepts incoming miner connections
func (p *StratumPool) acceptConnections() {
	for p.running {
		conn, err := p.listener.Accept()
		if err != nil {
			if p.running {
				fmt.Printf("Accept error: %v\n", err)
			}
			continue
		}

		go p.handleClient(conn)
	}
}

// handleClient handles a connected miner
func (p *StratumPool) handleClient(conn net.Conn) {
	clientID := conn.RemoteAddr().String()
	client := &PoolClient{
		conn:       conn,
		id:         clientID,
		difficulty: p.difficulty,
		lastShare:  time.Now(),
	}

	p.clientsMutex.Lock()
	p.clients[clientID] = client
	p.clientsMutex.Unlock()

	fmt.Printf("[MINING] New miner connected: %s\n", clientID)

	defer func() {
		p.clientsMutex.Lock()
		delete(p.clients, clientID)
		p.clientsMutex.Unlock()
		conn.Close()
		fmt.Printf("[MINING] Miner disconnected: %s (shares: %d)\n", clientID, client.shares)
	}()

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := scanner.Text()
		if err := p.handleMessage(client, line); err != nil {
			fmt.Printf("Error handling message from %s: %v\n", clientID, err)
			return
		}
	}
}

// handleMessage handles a Stratum message from a client
func (p *StratumPool) handleMessage(client *PoolClient, message string) error {
	var req StratumRequest
	if err := json.Unmarshal([]byte(message), &req); err != nil {
		return err
	}

	switch req.Method {
	case "mining.subscribe":
		return p.handleSubscribe(client, &req)
	case "mining.authorize":
		return p.handleAuthorize(client, &req)
	case "mining.submit":
		return p.handleSubmit(client, &req)
	default:
		return p.sendError(client, req.ID, fmt.Sprintf("unknown method: %s", req.Method))
	}
}

// handleSubscribe handles mining.subscribe
func (p *StratumPool) handleSubscribe(client *PoolClient, req *StratumRequest) error {
	client.subscribed = true

	result := []interface{}{
		[][]string{{"mining.notify", "ae6812eb4cd7735a302a8a9dd95c6578"}},
		"08000002",
		4,
	}

	if err := p.sendResponse(client, req.ID, result, nil); err != nil {
		return err
	}

	// Send difficulty
	if err := p.sendDifficulty(client); err != nil {
		return err
	}

	// Send current job
	if p.currentJob != nil {
		return p.sendJob(client, p.currentJob)
	}

	return nil
}

// handleAuthorize handles mining.authorize
func (p *StratumPool) handleAuthorize(client *PoolClient, req *StratumRequest) error {
	if len(req.Params) < 1 {
		return p.sendError(client, req.ID, "missing username")
	}

	username, ok := req.Params[0].(string)
	if !ok {
		return p.sendError(client, req.ID, "invalid username")
	}

	client.authorized = true
	fmt.Printf("[MINING] Miner authorized: %s (user: %s)\n", client.id, username)

	return p.sendResponse(client, req.ID, true, nil)
}

// handleSubmit handles mining.submit
func (p *StratumPool) handleSubmit(client *PoolClient, req *StratumRequest) error {
	if !client.authorized {
		return p.sendError(client, req.ID, "not authorized")
	}

	if len(req.Params) < 5 {
		return p.sendError(client, req.ID, "invalid submit parameters")
	}

	client.shares++
	client.lastShare = time.Now()

	fmt.Printf("[MINING] Share submitted by %s (total: %d)\n", client.id, client.shares)

	// TODO: Validate share and submit block if valid

	return p.sendResponse(client, req.ID, true, nil)
}

// sendResponse sends a Stratum response
func (p *StratumPool) sendResponse(client *PoolClient, id interface{}, result interface{}, err interface{}) error {
	resp := StratumResponse{
		ID:     id,
		Result: result,
		Error:  err,
	}

	data, jsonErr := json.Marshal(resp)
	if jsonErr != nil {
		return jsonErr
	}

	_, writeErr := client.conn.Write(append(data, '\n'))
	return writeErr
}

// sendError sends a Stratum error response
func (p *StratumPool) sendError(client *PoolClient, id interface{}, message string) error {
	return p.sendResponse(client, id, nil, []interface{}{21, message, nil})
}

// sendDifficulty sends mining.set_difficulty
func (p *StratumPool) sendDifficulty(client *PoolClient) error {
	method := map[string]interface{}{
		"id":     nil,
		"method": "mining.set_difficulty",
		"params": []interface{}{client.difficulty},
	}

	data, err := json.Marshal(method)
	if err != nil {
		return err
	}

	_, err = client.conn.Write(append(data, '\n'))
	return err
}

// sendJob sends mining.notify
func (p *StratumPool) sendJob(client *PoolClient, job *MiningJob) error {
	method := map[string]interface{}{
		"id":     nil,
		"method": "mining.notify",
		"params": []interface{}{
			job.JobID,
			job.PrevHash,
			job.Coinbase1,
			job.Coinbase2,
			job.MerkleBranch,
			job.Version,
			job.NBits,
			job.NTime,
			job.CleanJobs,
		},
	}

	data, err := json.Marshal(method)
	if err != nil {
		return err
	}

	_, err = client.conn.Write(append(data, '\n'))
	return err
}

// jobGenerator generates new mining jobs
func (p *StratumPool) jobGenerator() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for p.running {
		select {
		case <-ticker.C:
			if err := p.generateNewJob(); err != nil {
				fmt.Printf("Error generating job: %v\n", err)
			}
		}
	}
}

// generateNewJob creates a new mining job
func (p *StratumPool) generateNewJob() error {
	best, err := p.chain.BestBlock()
	if err != nil {
		return err
	}

	bestHash := best.BlockHash()
	currentHeight := p.chain.Height() + 1

	p.jobMutex.Lock()
	p.jobCounter++
	jobID := fmt.Sprintf("%016x", p.jobCounter)

	job := &MiningJob{
		JobID:        jobID,
		PrevHash:     bestHash.String(),
		Coinbase1:    "",
		Coinbase2:    "",
		MerkleBranch: []string{},
		Version:      fmt.Sprintf("%08x", 1),
		NBits:        fmt.Sprintf("%08x", best.Header.Bits),
		NTime:        fmt.Sprintf("%08x", time.Now().Unix()),
		CleanJobs:    true,
		Height:       currentHeight,
		Target:       best.Header.Bits,
	}

	p.currentJob = job
	p.jobMutex.Unlock()

	// Broadcast to all connected miners
	p.clientsMutex.RLock()
	clientCount := len(p.clients)
	for _, client := range p.clients {
		if client.subscribed {
			p.sendJob(client, job)
		}
	}
	p.clientsMutex.RUnlock()

	fmt.Printf("[MINING] New job generated: %s (height: %d, miners: %d)\n", jobID, currentHeight, clientCount)

	return nil
}

// GetStats returns pool statistics
func (p *StratumPool) GetStats() map[string]interface{} {
	p.clientsMutex.RLock()
	defer p.clientsMutex.RUnlock()

	totalShares := uint64(0)
	for _, client := range p.clients {
		totalShares += client.shares
	}

	return map[string]interface{}{
		"miners":       len(p.clients),
		"total_shares": totalShares,
		"difficulty":   p.difficulty,
		"pool_address": p.poolAddress,
	}
}
