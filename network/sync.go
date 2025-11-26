package network

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"net"
	"obsidian-core/blockchain"
	"obsidian-core/consensus"
	"obsidian-core/wire"
	"sync"
	"time"
)

// Connection limits
const (
	MaxOutboundPeers = 8
	MaxInboundPeers  = 125
	MaxTotalPeers    = MaxOutboundPeers + MaxInboundPeers
)

// Rate limiting constants
const (
	MaxMessagesPerSecond = 100
	RateLimitWindow      = time.Second
	BanDuration          = 24 * time.Hour
)

// Peer scoring constants
const (
	InitialPeerScore   = 0
	MaxPeerScore       = 100
	MinPeerScore       = -100
	BanThreshold       = -50
	ScoreDecayInterval = 10 * time.Minute
	ScoreDecayAmount   = 1
)

// Score adjustments
const (
	ScoreValidBlock        = 5
	ScoreValidTx           = 1
	ScoreInvalidBlock      = -25
	ScoreInvalidTx         = -10
	ScoreTimeout           = -5
	ScoreMisbehavior       = -20
	ScoreProtocolViolation = -50
)

// Message types for P2P communication
const (
	MsgTypeVersion      = "version"
	MsgTypeVerAck       = "verack"
	MsgTypeGetHeaders   = "getheaders"
	MsgTypeHeaders      = "headers"
	MsgTypeGetBlocks    = "getblocks"
	MsgTypeBlock        = "block"
	MsgTypeInv          = "inv"
	MsgTypeGetData      = "getdata"
	MsgTypeTx           = "tx"
	MsgTypePing         = "ping"
	MsgTypePong         = "pong"
	MsgTypeAddr         = "addr"
	MsgTypeGetAddr      = "getaddr"
	MsgTypeCompactBlock = "cmpctblock"
	MsgTypeGetBlockTxn  = "getblocktxn"
	MsgTypeBlockTxn     = "blocktxn"
	MsgTypeSendCmpct    = "sendcmpct"
)

// P2PMessage represents a message sent between peers.
type P2PMessage struct {
	Type    string
	Payload []byte
}

// VersionMessage contains information about a peer.
type VersionMessage struct {
	Version   int32
	Height    int32
	Timestamp int64
	UserAgent string
}

// HeadersMessage contains block headers.
type HeadersMessage struct {
	Headers []*wire.BlockHeader
}

// InvMessage announces known inventory (blocks or transactions).
type InvMessage struct {
	Type   string // "block" or "tx"
	Hashes []wire.Hash
}

// GetDataMessage requests specific inventory items.
type GetDataMessage struct {
	Type   string // "block" or "tx"
	Hashes []wire.Hash
}

// GetHeadersMessage requests block headers.
type GetHeadersMessage struct {
	StartHash wire.Hash
	StopHash  wire.Hash // Zero hash means get all after start
}

// GetBlocksMessage requests block hashes.
type GetBlocksMessage struct {
	StartHash wire.Hash
	StopHash  wire.Hash
}

// AddrMessage contains peer addresses.
type AddrMessage struct {
	Addresses []string
}

// Peer represents a connected peer.
type Peer struct {
	conn          net.Conn
	version       *VersionMessage
	connected     bool
	lastSeen      time.Time
	addr          string
	inbound       bool // true if peer connected to us, false if we connected to peer
	score         int
	messageCount  int
	lastRateReset time.Time
	bannedUntil   time.Time
	mu            sync.RWMutex
}

// NewPeer creates a new peer from a connection.
func NewPeer(conn net.Conn, addr string, inbound bool) *Peer {
	return &Peer{
		conn:          conn,
		connected:     true,
		lastSeen:      time.Now(),
		addr:          addr,
		inbound:       inbound,
		score:         InitialPeerScore,
		lastRateReset: time.Now(),
	}
}

// AdjustScore adjusts the peer's score by the given amount.
func (p *Peer) AdjustScore(adjustment int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.score += adjustment
	if p.score > MaxPeerScore {
		p.score = MaxPeerScore
	}
	if p.score < MinPeerScore {
		p.score = MinPeerScore
	}
}

// GetScore returns the peer's current score.
func (p *Peer) GetScore() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.score
}

// IsBanned returns whether the peer is currently banned.
func (p *Peer) IsBanned() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return time.Now().Before(p.bannedUntil)
}

// Ban bans the peer for the specified duration.
func (p *Peer) Ban(duration time.Duration) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.bannedUntil = time.Now().Add(duration)
	fmt.Printf("⛔ Banned peer %s until %v (score: %d)\n", p.addr, p.bannedUntil, p.score)
}

// CheckRateLimit checks if the peer has exceeded the rate limit.
func (p *Peer) CheckRateLimit() bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	now := time.Now()
	if now.Sub(p.lastRateReset) >= RateLimitWindow {
		p.messageCount = 0
		p.lastRateReset = now
	}

	p.messageCount++
	return p.messageCount > MaxMessagesPerSecond
}

// SendMessage sends a P2P message to the peer.
func (p *Peer) SendMessage(msgType string, payload interface{}) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.connected {
		return fmt.Errorf("peer not connected")
	}

	// Encode payload
	var payloadBytes []byte
	if payload != nil {
		buf := new(bytes.Buffer)
		encoder := gob.NewEncoder(buf)
		if err := encoder.Encode(payload); err != nil {
			return fmt.Errorf("failed to encode payload: %v", err)
		}
		payloadBytes = buf.Bytes()
	}

	// Create message
	msg := &P2PMessage{
		Type:    msgType,
		Payload: payloadBytes,
	}

	// Send message
	encoder := gob.NewEncoder(p.conn)
	if err := encoder.Encode(msg); err != nil {
		return fmt.Errorf("failed to send message: %v", err)
	}

	return nil
}

// ReceiveMessage receives a P2P message from the peer.
func (p *Peer) ReceiveMessage() (*P2PMessage, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if !p.connected {
		return nil, fmt.Errorf("peer not connected")
	}

	// Set read deadline to detect disconnections
	p.conn.SetReadDeadline(time.Now().Add(60 * time.Second))

	decoder := gob.NewDecoder(p.conn)
	msg := &P2PMessage{}
	if err := decoder.Decode(msg); err != nil {
		return nil, fmt.Errorf("failed to receive message: %v", err)
	}

	return msg, nil
}

// Disconnect closes the connection to the peer.
func (p *Peer) Disconnect() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.connected {
		p.conn.Close()
		p.connected = false
	}
}

// IsConnected returns whether the peer is connected.
func (p *Peer) IsConnected() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.connected
}

// SyncManager manages P2P synchronization.
type SyncManager struct {
	blockchain     *blockchain.BlockChain
	peerManager    *PeerManager
	pow            consensus.PowEngine
	peers          map[string]*Peer
	bannedPeers    map[string]time.Time
	knownBlocks    map[wire.Hash]bool
	knownTxs       map[wire.Hash]bool
	syncInProgress bool
	outboundCount  int
	inboundCount   int
	mu             sync.RWMutex
}

// NewSyncManager creates a new sync manager.
func NewSyncManager(bc *blockchain.BlockChain, pm *PeerManager, pow consensus.PowEngine) *SyncManager {
	sm := &SyncManager{
		blockchain:  bc,
		peerManager: pm,
		pow:         pow,
		peers:       make(map[string]*Peer),
		bannedPeers: make(map[string]time.Time),
		knownBlocks: make(map[wire.Hash]bool),
		knownTxs:    make(map[wire.Hash]bool),
	}

	// Start background tasks
	go sm.peerMaintenanceLoop()

	return sm
}

// ConnectToPeer manually connects to a peer.
func (sm *SyncManager) ConnectToPeer(addr string) {
	go func() {
		// Check if we can accept this peer (outbound)
		if !sm.canAcceptPeer(addr, false) {
			fmt.Printf("Cannot connect to peer %s (limit reached or banned)\n", addr)
			return
		}

		// Connect to peer
		conn, err := sm.peerManager.ConnectToPeer(addr)
		if err != nil {
			fmt.Printf("Failed to connect to peer %s: %v\n", addr, err)
			return
		}

		peer := NewPeer(conn, addr, false) // false = outbound

		sm.mu.Lock()
		sm.peers[addr] = peer
		sm.outboundCount++
		currentOutbound := sm.outboundCount
		sm.mu.Unlock()

		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Println("🔗 NEW PEER CONNECTED")
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Printf("  Address:   %s\n", addr)
		fmt.Printf("  Direction: Outbound\n")
		fmt.Printf("  Peers:     %d/%d outbound\n", currentOutbound, MaxOutboundPeers)
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

		// Perform handshake
		if err := sm.performHandshake(peer); err != nil {
			fmt.Printf("❌ Handshake failed with %s: %v\n", addr, err)
			peer.Disconnect()
			return
		}

		// Start message handler
		go sm.handlePeerMessages(peer)

		// Request initial sync
		sm.requestHeaderSync(peer)
	}()
}

// IsSyncing returns whether the node is currently syncing.
func (sm *SyncManager) IsSyncing() bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.syncInProgress
}

// peerMaintenanceLoop performs periodic peer maintenance tasks.
func (sm *SyncManager) peerMaintenanceLoop() {
	ticker := time.NewTicker(ScoreDecayInterval)
	defer ticker.Stop()

	for range ticker.C {
		sm.mu.Lock()

		// Decay peer scores
		for _, peer := range sm.peers {
			if peer.GetScore() > InitialPeerScore {
				peer.AdjustScore(-ScoreDecayAmount)
			} else if peer.GetScore() < InitialPeerScore {
				peer.AdjustScore(ScoreDecayAmount)
			}
		}

		// Clean up expired bans
		now := time.Now()
		for addr, bannedUntil := range sm.bannedPeers {
			if now.After(bannedUntil) {
				delete(sm.bannedPeers, addr)
				fmt.Printf("✅ Unbanned peer: %s\n", addr)
			}
		}

		// Disconnect banned peers
		for addr, peer := range sm.peers {
			if peer.IsBanned() || peer.GetScore() <= BanThreshold {
				if !peer.IsBanned() {
					peer.Ban(BanDuration)
					sm.bannedPeers[addr] = peer.bannedUntil
				}
				peer.Disconnect()
			}
		}

		sm.mu.Unlock()
	}
}

// canAcceptPeer checks if we can accept a new peer connection.
func (sm *SyncManager) canAcceptPeer(addr string, inbound bool) bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	// Check if peer is banned
	if bannedUntil, exists := sm.bannedPeers[addr]; exists {
		if time.Now().Before(bannedUntil) {
			return false
		}
	}

	// Check connection limits
	if inbound {
		return sm.inboundCount < MaxInboundPeers
	}
	return sm.outboundCount < MaxOutboundPeers
}

// Start begins the sync manager's operation.
func (sm *SyncManager) Start() error {
	// Discover peers
	peerAddrs := sm.peerManager.DiscoverPeers()
	if len(peerAddrs) == 0 {
		fmt.Println("No peers discovered, running in solo mode")
		return nil
	}

	// Connect to peers
	for _, addr := range peerAddrs {
		go sm.connectAndSync(addr)
	}

	return nil
}

// connectAndSync connects to a peer and starts synchronization.
func (sm *SyncManager) connectAndSync(addr string) {
	// Check if we can accept this peer (outbound)
	if !sm.canAcceptPeer(addr, false) {
		fmt.Printf("Cannot connect to peer %s (limit reached or banned)\n", addr)
		return
	}

	// Connect to peer
	conn, err := sm.peerManager.ConnectToPeer(addr)
	if err != nil {
		fmt.Printf("Failed to connect to peer %s: %v\n", addr, err)
		return
	}

	peer := NewPeer(conn, addr, false) // false = outbound

	sm.mu.Lock()
	sm.peers[addr] = peer
	sm.outboundCount++
	currentOutbound := sm.outboundCount
	sm.mu.Unlock()

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("🔗 NEW PEER CONNECTED")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("  Address:   %s\n", addr)
	fmt.Printf("  Direction: Outbound\n")
	fmt.Printf("  Peers:     %d/%d outbound\n", currentOutbound, MaxOutboundPeers)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Perform handshake
	if err := sm.performHandshake(peer); err != nil {
		fmt.Printf("Handshake failed with %s: %v\n", addr, err)
		peer.Disconnect()
		return
	}

	// Start message handler
	go sm.handlePeerMessages(peer)

	// Request initial sync
	sm.requestHeaderSync(peer)
}

// performHandshake performs the version handshake with a peer.
func (sm *SyncManager) performHandshake(peer *Peer) error {
	// Send version message
	version := &VersionMessage{
		Version:   1,
		Height:    sm.blockchain.Height(),
		Timestamp: time.Now().Unix(),
		UserAgent: "Obsidian/2.0.0",
	}

	if err := peer.SendMessage(MsgTypeVersion, version); err != nil {
		return fmt.Errorf("failed to send version: %v", err)
	}

	// Wait for version response
	msg, err := peer.ReceiveMessage()
	if err != nil {
		return fmt.Errorf("failed to receive version: %v", err)
	}

	if msg.Type != MsgTypeVersion {
		return fmt.Errorf("expected version message, got %s", msg.Type)
	}

	// Decode peer version
	buf := bytes.NewBuffer(msg.Payload)
	decoder := gob.NewDecoder(buf)
	peerVersion := &VersionMessage{}
	if err := decoder.Decode(peerVersion); err != nil {
		return fmt.Errorf("failed to decode version: %v", err)
	}

	peer.mu.Lock()
	peer.version = peerVersion
	peer.mu.Unlock()

	// Send verack
	if err := peer.SendMessage(MsgTypeVerAck, nil); err != nil {
		return fmt.Errorf("failed to send verack: %v", err)
	}

	// Wait for verack
	msg, err = peer.ReceiveMessage()
	if err != nil {
		return fmt.Errorf("failed to receive verack: %v", err)
	}

	if msg.Type != MsgTypeVerAck {
		return fmt.Errorf("expected verack message, got %s", msg.Type)
	}

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("🤝 HANDSHAKE COMPLETE")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("  Peer:       %s\n", peer.addr)
	fmt.Printf("  Version:    %d\n", peerVersion.Version)
	fmt.Printf("  Height:     %d\n", peerVersion.Height)
	fmt.Printf("  User Agent: %s\n", peerVersion.UserAgent)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	return nil
}

// handlePeerMessages handles incoming messages from a peer.
func (sm *SyncManager) handlePeerMessages(peer *Peer) {
	defer peer.Disconnect()

	for {
		if !peer.IsConnected() {
			break
		}

		// Check if peer is banned
		if peer.IsBanned() {
			fmt.Printf("Disconnecting banned peer: %s\n", peer.addr)
			break
		}

		// Check rate limit
		if peer.CheckRateLimit() {
			fmt.Printf("⚠️  Rate limit exceeded for peer %s\n", peer.addr)
			peer.AdjustScore(ScoreMisbehavior)
			if peer.GetScore() <= BanThreshold {
				peer.Ban(BanDuration)
			}
			break
		}

		msg, err := peer.ReceiveMessage()
		if err != nil {
			fmt.Printf("Error receiving message from %s: %v\n", peer.addr, err)
			peer.AdjustScore(ScoreTimeout)
			break
		}

		if err := sm.handleMessage(peer, msg); err != nil {
			fmt.Printf("Error handling message from %s: %v\n", peer.addr, err)
			peer.AdjustScore(ScoreMisbehavior)
		}

		peer.lastSeen = time.Now()
	}

	// Remove peer from active list
	sm.mu.Lock()
	delete(sm.peers, peer.addr)
	if peer.inbound {
		sm.inboundCount--
	} else {
		sm.outboundCount--
	}
	sm.mu.Unlock()

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("🔌 PEER DISCONNECTED")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("  Address:    %s\n", peer.addr)
	fmt.Printf("  Final Score: %d\n", peer.GetScore())
	if peer.inbound {
		fmt.Printf("  Direction:   Inbound\n")
	} else {
		fmt.Printf("  Direction:   Outbound\n")
	}
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
}

// handleMessage handles a specific message from a peer.
func (sm *SyncManager) handleMessage(peer *Peer, msg *P2PMessage) error {
	switch msg.Type {
	case MsgTypeGetHeaders:
		return sm.handleGetHeaders(peer, msg)
	case MsgTypeHeaders:
		return sm.handleHeaders(peer, msg)
	case MsgTypeGetBlocks:
		return sm.handleGetBlocks(peer, msg)
	case MsgTypeBlock:
		return sm.handleBlock(peer, msg)
	case MsgTypeInv:
		return sm.handleInv(peer, msg)
	case MsgTypeGetData:
		return sm.handleGetData(peer, msg)
	case MsgTypeTx:
		return sm.handleTx(peer, msg)
	case MsgTypePing:
		return peer.SendMessage(MsgTypePong, nil)
	case MsgTypePong:
		return nil
	case MsgTypeGetAddr:
		return sm.handleGetAddr(peer)
	case MsgTypeAddr:
		return sm.handleAddr(peer, msg)
	default:
		return fmt.Errorf("unknown message type: %s", msg.Type)
	}
}

// requestHeaderSync requests headers from a peer starting from our best block.
func (sm *SyncManager) requestHeaderSync(peer *Peer) error {
	bestBlock, err := sm.blockchain.BestBlock()
	if err != nil {
		return fmt.Errorf("failed to get best block: %v", err)
	}

	getHeaders := &GetHeadersMessage{
		StartHash: bestBlock.BlockHash(),
		StopHash:  wire.Hash{}, // Zero hash means get all
	}

	return peer.SendMessage(MsgTypeGetHeaders, getHeaders)
}

// handleGetHeaders responds to a getheaders request.
func (sm *SyncManager) handleGetHeaders(peer *Peer, msg *P2PMessage) error {
	buf := bytes.NewBuffer(msg.Payload)
	decoder := gob.NewDecoder(buf)
	req := &GetHeadersMessage{}
	if err := decoder.Decode(req); err != nil {
		return fmt.Errorf("failed to decode getheaders: %v", err)
	}

	// Get headers starting from the requested hash
	// For now, we'll just send our best block header
	bestBlock, err := sm.blockchain.BestBlock()
	if err != nil {
		return err
	}

	headers := &HeadersMessage{
		Headers: []*wire.BlockHeader{&bestBlock.Header},
	}

	return peer.SendMessage(MsgTypeHeaders, headers)
}

// handleHeaders processes received headers.
func (sm *SyncManager) handleHeaders(peer *Peer, msg *P2PMessage) error {
	buf := bytes.NewBuffer(msg.Payload)
	decoder := gob.NewDecoder(buf)
	headers := &HeadersMessage{}
	if err := decoder.Decode(headers); err != nil {
		return fmt.Errorf("failed to decode headers: %v", err)
	}

	fmt.Printf("Received %d headers from %s\n", len(headers.Headers), peer.addr)

	// Request blocks for each header we don't have
	var hashesToRequest []wire.Hash
	for _, header := range headers.Headers {
		blockHash := header.BlockHash()

		// Check if we already have this block
		if _, err := sm.blockchain.GetBlock(blockHash[:]); err != nil {
			hashesToRequest = append(hashesToRequest, blockHash)
		}
	}

	if len(hashesToRequest) > 0 {
		getData := &GetDataMessage{
			Type:   "block",
			Hashes: hashesToRequest,
		}
		return peer.SendMessage(MsgTypeGetData, getData)
	}

	return nil
}

// handleGetBlocks responds to a getblocks request.
func (sm *SyncManager) handleGetBlocks(peer *Peer, msg *P2PMessage) error {
	buf := bytes.NewBuffer(msg.Payload)
	decoder := gob.NewDecoder(buf)
	req := &GetBlocksMessage{}
	if err := decoder.Decode(req); err != nil {
		return fmt.Errorf("failed to decode getblocks: %v", err)
	}

	// Send inventory of blocks we have
	bestBlock, err := sm.blockchain.BestBlock()
	if err != nil {
		return err
	}

	inv := &InvMessage{
		Type:   "block",
		Hashes: []wire.Hash{bestBlock.BlockHash()},
	}

	return peer.SendMessage(MsgTypeInv, inv)
}

// handleBlock processes a received block.
func (sm *SyncManager) handleBlock(peer *Peer, msg *P2PMessage) error {
	buf := bytes.NewBuffer(msg.Payload)
	decoder := gob.NewDecoder(buf)
	block := &wire.MsgBlock{}
	if err := decoder.Decode(block); err != nil {
		peer.AdjustScore(ScoreProtocolViolation)
		return fmt.Errorf("failed to decode block: %v", err)
	}

	blockHash := block.BlockHash()

	// Mark as known
	sm.mu.Lock()
	sm.knownBlocks[blockHash] = true
	sm.mu.Unlock()

	// Process block
	if err := sm.blockchain.ProcessBlock(block, sm.pow); err != nil {
		peer.AdjustScore(ScoreInvalidBlock)
		fmt.Printf("❌ Invalid block from %s: %v\n", peer.addr, err)
		return fmt.Errorf("failed to process block: %v", err)
	}

	// Reward peer for valid block
	peer.AdjustScore(ScoreValidBlock)
	currentHeight := sm.blockchain.Height()

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("📦 NEW BLOCK RECEIVED")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("  Height:       %d\n", currentHeight)
	fmt.Printf("  Hash:         %s\n", blockHash.String())
	fmt.Printf("  From Peer:    %s\n", peer.addr)
	fmt.Printf("  Peer Score:   %d\n", peer.GetScore())
	fmt.Printf("  Transactions: %d\n", len(block.Transactions))
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Announce to other peers
	sm.announceBlock(block, peer.addr)
	peerCount := len(sm.peers) - 1 // Exclude source peer
	if peerCount > 0 {
		fmt.Printf("📡 Block relayed to %d other peer(s)\n", peerCount)
	}

	return nil
}

// handleInv processes an inventory announcement.
func (sm *SyncManager) handleInv(peer *Peer, msg *P2PMessage) error {
	buf := bytes.NewBuffer(msg.Payload)
	decoder := gob.NewDecoder(buf)
	inv := &InvMessage{}
	if err := decoder.Decode(inv); err != nil {
		return fmt.Errorf("failed to decode inv: %v", err)
	}

	fmt.Printf("Received inventory of %d %ss from %s\n", len(inv.Hashes), inv.Type, peer.addr)

	// Request items we don't have
	var hashesToRequest []wire.Hash
	for _, hash := range inv.Hashes {
		sm.mu.RLock()
		known := false
		if inv.Type == "block" {
			known = sm.knownBlocks[hash]
		} else {
			known = sm.knownTxs[hash]
		}
		sm.mu.RUnlock()

		if !known {
			hashesToRequest = append(hashesToRequest, hash)
		}
	}

	if len(hashesToRequest) > 0 {
		getData := &GetDataMessage{
			Type:   inv.Type,
			Hashes: hashesToRequest,
		}
		return peer.SendMessage(MsgTypeGetData, getData)
	}

	return nil
}

// handleGetData responds to a getdata request.
func (sm *SyncManager) handleGetData(peer *Peer, msg *P2PMessage) error {
	buf := bytes.NewBuffer(msg.Payload)
	decoder := gob.NewDecoder(buf)
	req := &GetDataMessage{}
	if err := decoder.Decode(req); err != nil {
		return fmt.Errorf("failed to decode getdata: %v", err)
	}

	mempool := sm.blockchain.Mempool()

	for _, hash := range req.Hashes {
		if req.Type == "block" {
			block, err := sm.blockchain.GetBlock(hash[:])
			if err != nil {
				continue
			}
			if err := peer.SendMessage(MsgTypeBlock, block); err != nil {
				return err
			}
		} else if req.Type == "tx" {
			// Get transaction from mempool
			tx, err := mempool.GetTransaction(hash)
			if err == nil {
				if err := peer.SendMessage(MsgTypeTx, tx); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// handleTx processes a received transaction.
func (sm *SyncManager) handleTx(peer *Peer, msg *P2PMessage) error {
	buf := bytes.NewBuffer(msg.Payload)
	decoder := gob.NewDecoder(buf)
	tx := &wire.MsgTx{}
	if err := decoder.Decode(tx); err != nil {
		peer.AdjustScore(ScoreProtocolViolation)
		return fmt.Errorf("failed to decode tx: %v", err)
	}

	txHash := tx.TxHash()
	fmt.Printf("Received transaction %s from %s\n", txHash.String(), peer.addr)

	// Mark as known
	sm.mu.Lock()
	sm.knownTxs[txHash] = true
	sm.mu.Unlock()

	// Add to mempool
	mempool := sm.blockchain.Mempool()
	currentHeight := sm.blockchain.Height()

	// Calculate fee (inputs - outputs)
	// For now, we'll use 0 fee as we don't have UTXO lookup in sync context
	// In production, you'd need to look up input values
	fee := int64(0)

	if err := mempool.AddTransaction(tx, currentHeight, fee); err != nil {
		fmt.Printf("Failed to add transaction to mempool: %v\n", err)
		peer.AdjustScore(ScoreInvalidTx)
		return nil // Don't fail on invalid tx, just log it
	}

	// Reward peer for valid transaction
	peer.AdjustScore(ScoreValidTx)
	fmt.Printf("Transaction %s added to mempool (peer score: %d)\n", txHash.String(), peer.GetScore())

	// Announce to other peers
	sm.announceTx(tx, peer.addr)

	return nil
}

// handleGetAddr responds to a getaddr request.
func (sm *SyncManager) handleGetAddr(peer *Peer) error {
	sm.mu.RLock()
	addrs := make([]string, 0, len(sm.peers))
	for addr := range sm.peers {
		addrs = append(addrs, addr)
	}
	sm.mu.RUnlock()

	addrMsg := &AddrMessage{
		Addresses: addrs,
	}

	return peer.SendMessage(MsgTypeAddr, addrMsg)
}

// handleAddr processes received peer addresses.
func (sm *SyncManager) handleAddr(peer *Peer, msg *P2PMessage) error {
	buf := bytes.NewBuffer(msg.Payload)
	decoder := gob.NewDecoder(buf)
	addrMsg := &AddrMessage{}
	if err := decoder.Decode(addrMsg); err != nil {
		return fmt.Errorf("failed to decode addr: %v", err)
	}

	fmt.Printf("Received %d peer addresses from %s\n", len(addrMsg.Addresses), peer.addr)

	// Connect to new peers
	for _, addr := range addrMsg.Addresses {
		sm.mu.RLock()
		_, exists := sm.peers[addr]
		sm.mu.RUnlock()

		if !exists && addr != peer.addr {
			go sm.connectAndSync(addr)
		}
	}

	return nil
}

// announceBlock announces a new block to all peers except the source.
func (sm *SyncManager) announceBlock(block *wire.MsgBlock, excludeAddr string) {
	blockHash := block.BlockHash()

	inv := &InvMessage{
		Type:   "block",
		Hashes: []wire.Hash{blockHash},
	}

	sm.mu.RLock()
	defer sm.mu.RUnlock()

	for addr, peer := range sm.peers {
		if addr != excludeAddr && peer.IsConnected() {
			go peer.SendMessage(MsgTypeInv, inv)
		}
	}
}

// announceTx announces a new transaction to all peers except the source.
func (sm *SyncManager) announceTx(tx *wire.MsgTx, excludeAddr string) {
	txHash := tx.TxHash()

	inv := &InvMessage{
		Type:   "tx",
		Hashes: []wire.Hash{txHash},
	}

	sm.mu.RLock()
	defer sm.mu.RUnlock()

	for addr, peer := range sm.peers {
		if addr != excludeAddr && peer.IsConnected() {
			go peer.SendMessage(MsgTypeInv, inv)
		}
	}
}

// BroadcastBlock broadcasts a newly mined block to all peers.
func (sm *SyncManager) BroadcastBlock(block *wire.MsgBlock) {
	sm.announceBlock(block, "")
}

// BroadcastTransaction broadcasts a new transaction to all peers.
func (sm *SyncManager) BroadcastTransaction(tx *wire.MsgTx) {
	sm.announceTx(tx, "")
}

// GetPeerCount returns the number of connected peers.
func (sm *SyncManager) GetPeerCount() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return len(sm.peers)
}

// PeerInfo contains information about a peer.
type PeerInfo struct {
	Addr         string
	Inbound      bool
	Score        int
	Height       int32
	LastSeen     time.Time
	MessageCount int
	Banned       bool
}

// GetPeerInfo returns information about all connected peers.
func (sm *SyncManager) GetPeerInfo() []PeerInfo {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	info := make([]PeerInfo, 0, len(sm.peers))
	for _, peer := range sm.peers {
		peer.mu.RLock()
		peerInfo := PeerInfo{
			Addr:         peer.addr,
			Inbound:      peer.inbound,
			Score:        peer.score,
			LastSeen:     peer.lastSeen,
			MessageCount: peer.messageCount,
			Banned:       peer.IsBanned(),
		}
		if peer.version != nil {
			peerInfo.Height = peer.version.Height
		}
		peer.mu.RUnlock()
		info = append(info, peerInfo)
	}

	return info
}

// GetConnectionStats returns connection statistics.
func (sm *SyncManager) GetConnectionStats() (inbound, outbound, banned int) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.inboundCount, sm.outboundCount, len(sm.bannedPeers)
}
