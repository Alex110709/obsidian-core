package network

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"math/rand"
	"net"
	"obsidian-core/blockchain"
	"obsidian-core/consensus"
	"obsidian-core/wire"
	"strings"
	"sync"
	"time"
)

// Connection limits
const (
	MaxOutboundPeers = 8
	MaxInboundPeers  = 125
	MaxTotalPeers    = MaxOutboundPeers + MaxInboundPeers
	MaxMessageSize   = 10 * 1024 * 1024 // 10MB max message size
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
	ScoreDuplicateTx       = -2 // Penalty for duplicate transactions
	ScoreStaleBlock        = -3 // Penalty for stale blocks
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
	MsgTypeReject       = "reject"
	MsgTypeFeeFilter    = "feefilter"
	MsgTypeSendHeaders  = "sendheaders"
	MsgTypeNotFound     = "notfound"
	MsgTypeMemPool      = "mempool"
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

// RejectMessage contains rejection information for invalid messages.
type RejectMessage struct {
	Message string // The command that was rejected
	CCode   string // Rejection code (e.g., "malformed", "invalid", "obsolete", "duplicate", "nonstandard", "dust", "insufficientfee", "checkpoint")
	Reason  string // Human-readable reason for rejection
	Data    []byte // Extra data (e.g., the hash of the rejected object)
}

// FeeFilterMessage contains the minimum fee rate.
type FeeFilterMessage struct {
	FeeRate int64 // Minimum fee rate in satoshis per kilobyte
}

// SendHeadersMessage requests headers-first block relay.
type SendHeadersMessage struct{}

// NotFoundMessage contains inventory that was not found.
type NotFoundMessage struct {
	Type   string // "block" or "tx"
	Hashes []wire.Hash
}

// MemPoolMessage requests mempool contents.
type MemPoolMessage struct{}

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
	feeFilter     int64 // Minimum fee rate in sat/kB
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

// ReceiveMessageWithTimeout receives a P2P message with a custom timeout.
func (p *Peer) ReceiveMessageWithTimeout(timeout time.Duration) (*P2PMessage, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if !p.connected {
		return nil, fmt.Errorf("peer not connected")
	}

	// Set read deadline
	p.conn.SetReadDeadline(time.Now().Add(timeout))

	// Create fresh decoder for each message to avoid gob type conflicts
	decoder := gob.NewDecoder(p.conn)
	msg := &P2PMessage{}
	if err := decoder.Decode(msg); err != nil {
		// Check for specific gob errors
		if strings.Contains(err.Error(), "duplicate type received") {
			p.AdjustScore(ScoreProtocolViolation)
			return nil, fmt.Errorf("gob type conflict: %v", err)
		}
		return nil, fmt.Errorf("failed to receive message: %v", err)
	}

	// Enhanced security checks
	if len(msg.Payload) > MaxMessageSize {
		p.AdjustScore(ScoreMisbehavior)
		return nil, fmt.Errorf("message payload too large: %d bytes (max: %d)", len(msg.Payload), MaxMessageSize)
	}

	// Check for empty or invalid message types
	if msg.Type == "" {
		p.AdjustScore(ScoreProtocolViolation)
		return nil, fmt.Errorf("empty message type")
	}

	// Validate message type
	validTypes := map[string]bool{
		MsgTypeVersion:      true,
		MsgTypeVerAck:       true,
		MsgTypeGetHeaders:   true,
		MsgTypeHeaders:      true,
		MsgTypeGetBlocks:    true,
		MsgTypeBlock:        true,
		MsgTypeInv:          true,
		MsgTypeGetData:      true,
		MsgTypeTx:           true,
		MsgTypePing:         true,
		MsgTypePong:         true,
		MsgTypeAddr:         true,
		MsgTypeGetAddr:      true,
		MsgTypeCompactBlock: true,
		MsgTypeGetBlockTxn:  true,
		MsgTypeBlockTxn:     true,
		MsgTypeSendCmpct:    true,
		MsgTypeReject:       true,
		MsgTypeFeeFilter:    true,
		MsgTypeSendHeaders:  true,
		MsgTypeNotFound:     true,
		MsgTypeMemPool:      true,
	}

	if !validTypes[msg.Type] {
		p.AdjustScore(ScoreProtocolViolation)
		return nil, fmt.Errorf("unknown message type: %s", msg.Type)
	}

	return msg, nil
}

// ReceiveMessage receives a P2P message from the peer.
func (p *Peer) ReceiveMessage() (*P2PMessage, error) {
	return p.ReceiveMessageWithTimeout(300 * time.Second)
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

	// Control
	stopChan chan struct{}
	running  bool
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
		stopChan:    make(chan struct{}),
		running:     false,
	}

	// Start background tasks
	go sm.peerMaintenanceLoop()

	return sm
}

// Stop stops the sync manager
func (sm *SyncManager) Stop() {
	if sm.running {
		close(sm.stopChan)
		sm.running = false

		// Disconnect all peers
		sm.mu.Lock()
		for addr, peer := range sm.peers {
			if peer.conn != nil {
				peer.conn.Close()
			}
			delete(sm.peers, addr)
		}
		sm.mu.Unlock()

		fmt.Println("Sync manager stopped")
	}
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

		fmt.Println("New peer connected")
		fmt.Printf("  Address:   %s\n", addr)
		fmt.Printf("  Direction: Outbound\n")
		fmt.Printf("  Peers:     %d/%d outbound\n", currentOutbound, MaxOutboundPeers)
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

		// Perform handshake
		if err := sm.performHandshake(peer); err != nil {
			fmt.Printf("❌ Handshake failed with %s: %v\n", addr, err)
			peer.Disconnect()
			// Clean up failed connection
			sm.mu.Lock()
			delete(sm.peers, addr)
			sm.outboundCount--
			sm.mu.Unlock()
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

		// Decay peer scores (more aggressive decay for negative scores)
		for addr, peer := range sm.peers {
			score := peer.GetScore()
			if score > InitialPeerScore {
				// Good peers decay slowly
				peer.AdjustScore(-ScoreDecayAmount / 2)
			} else if score < InitialPeerScore {
				// Bad peers decay towards neutral faster
				decay := ScoreDecayAmount
				if score < -10 {
					decay = ScoreDecayAmount * 2 // Faster decay for very bad peers
				}
				peer.AdjustScore(decay)
			}

			// Check for long-inactive peers
			if time.Since(peer.lastSeen) > 10*time.Minute && !peer.inbound {
				fmt.Printf("Disconnecting inactive outbound peer: %s\n", addr)
				peer.Disconnect()
			}
		}

		// Clean up expired bans with progressive unbanning
		now := time.Now()
		for addr, bannedUntil := range sm.bannedPeers {
			if now.After(bannedUntil) {
				delete(sm.bannedPeers, addr)
				fmt.Printf("✅ Unbanned peer: %s\n", addr)
				// Allow reconnection attempt after unbanning
				go sm.scheduleReconnection(addr, 1)
			}
		}

		// Disconnect and ban misbehaving peers
		for addr, peer := range sm.peers {
			score := peer.GetScore()
			if peer.IsBanned() || score <= BanThreshold {
				if !peer.IsBanned() {
					// Progressive banning: longer bans for repeated offenses
					banDuration := BanDuration
					if score <= BanThreshold-20 {
						banDuration = 7 * 24 * time.Hour // 1 week for very bad peers
					} else if score <= BanThreshold-10 {
						banDuration = 24 * time.Hour // 1 day for moderately bad peers
					}

					peer.Ban(banDuration)
					sm.bannedPeers[addr] = peer.bannedUntil
					fmt.Printf("🚫 Banned misbehaving peer %s for %v (score: %d)\n", addr, banDuration, score)
				}
				peer.Disconnect()
			}
		}

		// Clean up disconnected peers from the map
		for addr, peer := range sm.peers {
			if !peer.IsConnected() {
				delete(sm.peers, addr)
				if peer.inbound {
					sm.inboundCount--
				} else {
					sm.outboundCount--
				}
				fmt.Printf("Cleaned up disconnected peer: %s\n", addr)
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
	sm.running = true

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

	// Start periodic peer reconnection
	go sm.peerReconnectionLoop()

	return nil
}

// StartListener starts a P2P listener for inbound connections.
func (sm *SyncManager) StartListener(addr string) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to start P2P listener: %v", err)
	}
	defer listener.Close()

	fmt.Printf("P2P listener started on %s\n", addr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Failed to accept connection: %v\n", err)
			continue
		}

		// Check if we can accept this inbound connection
		if !sm.canAcceptPeer(conn.RemoteAddr().String(), true) {
			fmt.Printf("Rejecting inbound connection from %s (limit reached)\n", conn.RemoteAddr().String())
			conn.Close()
			continue
		}

		peer := NewPeer(conn, conn.RemoteAddr().String(), true) // true = inbound

		sm.mu.Lock()
		sm.peers[conn.RemoteAddr().String()] = peer
		sm.inboundCount++
		currentInbound := sm.inboundCount
		sm.mu.Unlock()

		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Println("🔗 NEW INBOUND PEER CONNECTED")
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Printf("  Address:   %s\n", conn.RemoteAddr().String())
		fmt.Printf("  Direction: Inbound\n")
		fmt.Printf("  Peers:     %d/%d inbound\n", currentInbound, MaxInboundPeers)
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

		// Perform handshake for inbound connections
		go func() {
			if err := sm.performHandshake(peer); err != nil {
				fmt.Printf("❌ Handshake failed with inbound peer %s: %v\n", conn.RemoteAddr().String(), err)
				peer.Disconnect()
				// Clean up failed connection
				sm.mu.Lock()
				delete(sm.peers, conn.RemoteAddr().String())
				sm.inboundCount--
				sm.mu.Unlock()
				return
			}

			// Start handling messages from this peer
			sm.handlePeerMessages(peer)
		}()
	}
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
		// Clean up failed connection
		sm.mu.Lock()
		delete(sm.peers, addr)
		sm.outboundCount--
		sm.mu.Unlock()
		return
	}

	// Send sendheaders to enable headers-first relay
	if err := peer.SendMessage(MsgTypeSendHeaders, &SendHeadersMessage{}); err != nil {
		fmt.Printf("Failed to send sendheaders to %s: %v\n", addr, err)
	}

	// Send sendcmpct to negotiate compact block relay (version 1)
	sendCmpctMsg := map[string]interface{}{
		"version":  1,
		"announce": true,
	}
	if err := peer.SendMessage(MsgTypeSendCmpct, sendCmpctMsg); err != nil {
		fmt.Printf("Failed to send sendcmpct to %s: %v\n", addr, err)
	}

	// Start message handler
	go sm.handlePeerMessages(peer)

	// Request initial sync
	sm.requestHeaderSync(peer)
}

// performHandshake performs the version handshake with a peer.
// Uses bidirectional handshake: both peers send version simultaneously,
// then both send verack after receiving the other's version.
func (sm *SyncManager) performHandshake(peer *Peer) error {
	// Validate peer is not banned before handshake
	if peer.IsBanned() {
		return fmt.Errorf("peer is banned")
	}

	version := &VersionMessage{
		Version:   1,
		Height:    sm.blockchain.Height(),
		Timestamp: time.Now().Unix(),
		UserAgent: "Obsidian/2.0.0",
	}

	handshakeTimeout := 30 * time.Second

	// Channel for send/receive errors
	sendErrCh := make(chan error, 1)
	recvErrCh := make(chan error, 1)
	versionCh := make(chan *VersionMessage, 1)

	// Send our version in a goroutine (non-blocking)
	go func() {
		// Small delay for inbound to avoid both sending at exact same time
		if peer.inbound {
			time.Sleep(10 * time.Millisecond)
		}
		sendErrCh <- peer.SendMessage(MsgTypeVersion, version)
	}()

	// Receive peer's version in a goroutine (non-blocking)
	go func() {
		msg, err := peer.ReceiveMessageWithTimeout(handshakeTimeout)
		if err != nil {
			recvErrCh <- fmt.Errorf("failed to receive version: %v", err)
			return
		}

		if msg.Type != MsgTypeVersion {
			peer.AdjustScore(ScoreProtocolViolation)
			recvErrCh <- fmt.Errorf("expected version message, got %s", msg.Type)
			return
		}

		// Decode peer version with fresh decoder to avoid gob type conflicts
		buf := bytes.NewBuffer(msg.Payload)
		decoder := gob.NewDecoder(buf)
		peerVersion := &VersionMessage{}
		if err := decoder.Decode(peerVersion); err != nil {
			peer.AdjustScore(ScoreProtocolViolation)
			recvErrCh <- fmt.Errorf("failed to decode version: %v", err)
			return
		}

		// Validate peer version
		if err := sm.validatePeerVersion(peerVersion); err != nil {
			peer.AdjustScore(ScoreProtocolViolation)
			recvErrCh <- fmt.Errorf("invalid peer version: %v", err)
			return
		}

		versionCh <- peerVersion
		recvErrCh <- nil
	}()

	// Wait for both send and receive to complete
	var peerVersion *VersionMessage
	var sendErr, recvErr error

	for i := 0; i < 2; i++ {
		select {
		case err := <-sendErrCh:
			sendErr = err
			if err != nil {
				return fmt.Errorf("failed to send version: %v", err)
			}
		case err := <-recvErrCh:
			recvErr = err
			if err != nil {
				return err
			}
			peerVersion = <-versionCh
		case <-time.After(handshakeTimeout):
			peer.AdjustScore(ScoreTimeout)
			return fmt.Errorf("handshake timeout")
		}
	}

	if sendErr != nil || recvErr != nil {
		if sendErr != nil {
			return sendErr
		}
		return recvErr
	}

	peer.mu.Lock()
	peer.version = peerVersion
	peer.mu.Unlock()

	// Now exchange verack messages (also bidirectional)
	sendErrCh = make(chan error, 1)
	recvErrCh = make(chan error, 1)

	// Send verack
	go func() {
		sendErrCh <- peer.SendMessage(MsgTypeVerAck, nil)
	}()

	// Receive verack
	go func() {
		msg, err := peer.ReceiveMessageWithTimeout(handshakeTimeout)
		if err != nil {
			peer.AdjustScore(ScoreTimeout)
			recvErrCh <- fmt.Errorf("failed to receive verack: %v", err)
			return
		}

		if msg.Type != MsgTypeVerAck {
			peer.AdjustScore(ScoreProtocolViolation)
			recvErrCh <- fmt.Errorf("expected verack message, got %s", msg.Type)
			return
		}

		recvErrCh <- nil
	}()

	// Wait for both verack operations
	for i := 0; i < 2; i++ {
		select {
		case err := <-sendErrCh:
			if err != nil {
				return fmt.Errorf("failed to send verack: %v", err)
			}
		case err := <-recvErrCh:
			if err != nil {
				return err
			}
		case <-time.After(handshakeTimeout):
			peer.AdjustScore(ScoreTimeout)
			return fmt.Errorf("verack timeout")
		}
	}

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("🤝 HANDSHAKE COMPLETE")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("  Peer:       %s\n", peer.addr)
	fmt.Printf("  Direction:  %s\n", map[bool]string{true: "Inbound", false: "Outbound"}[peer.inbound])
	fmt.Printf("  Version:    %d\n", peerVersion.Version)
	fmt.Printf("  Height:     %d\n", peerVersion.Height)
	fmt.Printf("  User Agent: %s\n", peerVersion.UserAgent)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	return nil
}

// validatePeerVersion validates the peer's version message
func (sm *SyncManager) validatePeerVersion(version *VersionMessage) error {
	// Check version compatibility
	if version.Version < 1 || version.Version > 1 {
		return fmt.Errorf("unsupported protocol version: %d", version.Version)
	}

	// Check timestamp (not too far in future or past)
	now := time.Now().Unix()
	maxTimeOffset := int64(24 * 60 * 60) // 24 hours
	if version.Timestamp > now+maxTimeOffset {
		return fmt.Errorf("peer timestamp too far in future: %d", version.Timestamp)
	}
	if version.Timestamp < now-maxTimeOffset {
		return fmt.Errorf("peer timestamp too far in past: %d", version.Timestamp)
	}

	// Check height (not unreasonably high)
	maxReasonableHeight := sm.blockchain.Height() + 10000 // Allow some leeway
	if version.Height > maxReasonableHeight {
		return fmt.Errorf("peer height too high: %d (our height: %d)", version.Height, sm.blockchain.Height())
	}

	// Check user agent length
	if len(version.UserAgent) > 256 {
		return fmt.Errorf("user agent too long: %d bytes", len(version.UserAgent))
	}

	return nil
}

// handlePeerMessages handles incoming messages from a peer.
func (sm *SyncManager) handlePeerMessages(peer *Peer) {
	defer peer.Disconnect()

	// Start keep-alive ping routine
	pingTicker := time.NewTicker(60 * time.Second) // Send ping every 60 seconds
	defer pingTicker.Stop()

	go func() {
		for range pingTicker.C {
			if peer.IsConnected() && !peer.IsBanned() {
				// Send ping to keep connection alive
				if err := peer.SendMessage(MsgTypePing, nil); err != nil {
					fmt.Printf("Failed to send ping to %s: %v\n", peer.addr, err)
					return
				}
			} else {
				return
			}
		}
	}()

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
			// Handle EOF as normal disconnection, not an error
			if err == io.EOF {
				fmt.Printf("Peer %s disconnected normally\n", peer.addr)
			} else {
				fmt.Printf("Error receiving message from %s: %v\n", peer.addr, err)
				peer.AdjustScore(ScoreTimeout)
			}
			break
		}

		if err := sm.handleMessage(peer, msg); err != nil {
			fmt.Printf("Error handling message from %s: %v\n", peer.addr, err)
			peer.AdjustScore(ScoreMisbehavior)

			// Send reject message for unknown message types
			if strings.Contains(err.Error(), "unknown message type") {
				sm.sendReject(peer, msg.Type, "malformed", "Unknown message type", nil)
			}
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
		// For outbound connections, attempt to reconnect with backoff
		if !peer.inbound && peer.GetScore() > -10 {
			go sm.scheduleReconnection(peer.addr, 1) // Start with 1 second delay
		}
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
	fmt.Printf("  Reason:      %s\n", getDisconnectReason(peer))
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
}

// handleMessage handles a specific message from a peer.
func (sm *SyncManager) handleMessage(peer *Peer, msg *P2PMessage) error {
	// Rate limiting check
	if peer.CheckRateLimit() {
		fmt.Printf("⚠️  Rate limit exceeded for peer %s\n", peer.addr)
		peer.AdjustScore(ScoreMisbehavior)
		if peer.GetScore() <= BanThreshold {
			peer.Ban(BanDuration)
		}
		return fmt.Errorf("rate limit exceeded")
	}

	// Update last seen time
	peer.lastSeen = time.Now()

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
		// Respond to ping with pong
		if err := peer.SendMessage(MsgTypePong, nil); err != nil {
			peer.AdjustScore(ScoreTimeout)
			return fmt.Errorf("failed to send pong: %v", err)
		}
		return nil
	case MsgTypePong:
		// Pong received, connection is alive
		return nil
	case MsgTypeGetAddr:
		return sm.handleGetAddr(peer)
	case MsgTypeAddr:
		return sm.handleAddr(peer, msg)
	case MsgTypeReject:
		return sm.handleReject(peer, msg)
	case MsgTypeFeeFilter:
		return sm.handleFeeFilter(peer, msg)
	case MsgTypeSendHeaders:
		return sm.handleSendHeaders(peer, msg)
	case MsgTypeNotFound:
		return sm.handleNotFound(peer, msg)
	case MsgTypeMemPool:
		return sm.handleMemPool(peer, msg)
	case MsgTypeVersion:
		// Version messages should be handled during handshake, penalize if received here
		peer.AdjustScore(ScoreProtocolViolation)
		return fmt.Errorf("unexpected version message after handshake")
	case MsgTypeVerAck:
		// VerAck messages should be handled during handshake, penalize if received here
		peer.AdjustScore(ScoreProtocolViolation)
		return fmt.Errorf("unexpected verack message after handshake")
	default:
		peer.AdjustScore(ScoreProtocolViolation)
		sm.sendReject(peer, msg.Type, "malformed", "Unknown message type", nil)
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
	// Basic payload size check
	if len(msg.Payload) < 80 { // Minimum block header size
		peer.AdjustScore(ScoreProtocolViolation)
		return fmt.Errorf("block message too small: %d bytes", len(msg.Payload))
	}

	buf := bytes.NewBuffer(msg.Payload)
	decoder := gob.NewDecoder(buf)
	block := &wire.MsgBlock{}
	if err := decoder.Decode(block); err != nil {
		peer.AdjustScore(ScoreProtocolViolation)
		return fmt.Errorf("failed to decode block: %v", err)
	}

	blockHash := block.BlockHash()

	// Check if we already know this block
	sm.mu.RLock()
	alreadyKnown := sm.knownBlocks[blockHash]
	sm.mu.RUnlock()

	if alreadyKnown {
		peer.AdjustScore(ScoreDuplicateTx) // Use duplicate penalty
		return fmt.Errorf("duplicate block: %s", blockHash.String())
	}

	// Mark as known
	sm.mu.Lock()
	sm.knownBlocks[blockHash] = true
	sm.mu.Unlock()

	// Basic block validation
	if len(block.Transactions) == 0 {
		peer.AdjustScore(ScoreInvalidBlock)
		return fmt.Errorf("block contains no transactions")
	}

	// Check block size against our limits
	blockSize := len(msg.Payload)
	maxBlockSize := 3200000 // 3.2MB, should match chaincfg.BlockMaxSize
	if blockSize > maxBlockSize {
		peer.AdjustScore(ScoreInvalidBlock)
		return fmt.Errorf("block too large: %d bytes (max: %d)", blockSize, maxBlockSize)
	}

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
	fmt.Printf("  Block Size:   %d bytes\n", blockSize)
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
	notFound := make([]wire.Hash, 0)

	for _, hash := range req.Hashes {
		if req.Type == "block" {
			block, err := sm.blockchain.GetBlock(hash[:])
			if err != nil {
				notFound = append(notFound, hash)
				continue
			}
			if err := peer.SendMessage(MsgTypeBlock, block); err != nil {
				return err
			}
		} else if req.Type == "tx" {
			// Get transaction from mempool
			tx, err := mempool.GetTransaction(hash)
			if err != nil {
				notFound = append(notFound, hash)
				continue
			}
			if err := peer.SendMessage(MsgTypeTx, tx); err != nil {
				return err
			}
		}
	}

	// Send notfound for missing items
	if len(notFound) > 0 {
		notFoundMsg := &NotFoundMessage{
			Type:   req.Type,
			Hashes: notFound,
		}
		if err := peer.SendMessage(MsgTypeNotFound, notFoundMsg); err != nil {
			return err
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

// handleReject processes a reject message.
func (sm *SyncManager) handleReject(peer *Peer, msg *P2PMessage) error {
	buf := bytes.NewBuffer(msg.Payload)
	decoder := gob.NewDecoder(buf)
	rejectMsg := &RejectMessage{}
	if err := decoder.Decode(rejectMsg); err != nil {
		return fmt.Errorf("failed to decode reject: %v", err)
	}

	fmt.Printf("Received reject from %s: message=%s, code=%s, reason=%s\n",
		peer.addr, rejectMsg.Message, rejectMsg.CCode, rejectMsg.Reason)

	// Adjust peer score for protocol violations
	if rejectMsg.CCode == "malformed" || rejectMsg.CCode == "invalid" {
		peer.AdjustScore(ScoreProtocolViolation)
	}

	return nil
}

// handleFeeFilter processes a feefilter message.
func (sm *SyncManager) handleFeeFilter(peer *Peer, msg *P2PMessage) error {
	buf := bytes.NewBuffer(msg.Payload)
	decoder := gob.NewDecoder(buf)
	feeFilterMsg := &FeeFilterMessage{}
	if err := decoder.Decode(feeFilterMsg); err != nil {
		return fmt.Errorf("failed to decode feefilter: %v", err)
	}

	fmt.Printf("Received feefilter from %s: fee rate=%d sat/kB\n", peer.addr, feeFilterMsg.FeeRate)

	// Store the fee filter for this peer
	peer.mu.Lock()
	peer.feeFilter = feeFilterMsg.FeeRate
	peer.mu.Unlock()

	return nil
}

// handleSendHeaders processes a sendheaders message.
func (sm *SyncManager) handleSendHeaders(peer *Peer, msg *P2PMessage) error {
	buf := bytes.NewBuffer(msg.Payload)
	decoder := gob.NewDecoder(buf)
	sendHeadersMsg := &SendHeadersMessage{}
	if err := decoder.Decode(sendHeadersMsg); err != nil {
		return fmt.Errorf("failed to decode sendheaders: %v", err)
	}

	fmt.Printf("Received sendheaders from %s\n", peer.addr)

	// Enable headers-first mode for this peer (not implemented yet)
	// In a full implementation, you'd send headers instead of inv for new blocks

	return nil
}

// handleNotFound processes a notfound message.
func (sm *SyncManager) handleNotFound(peer *Peer, msg *P2PMessage) error {
	buf := bytes.NewBuffer(msg.Payload)
	decoder := gob.NewDecoder(buf)
	notFoundMsg := &NotFoundMessage{}
	if err := decoder.Decode(notFoundMsg); err != nil {
		return fmt.Errorf("failed to decode notfound: %v", err)
	}

	fmt.Printf("Received notfound from %s: %d %ss not found\n",
		peer.addr, len(notFoundMsg.Hashes), notFoundMsg.Type)

	// Mark items as not found (not implemented yet)
	// In a full implementation, you'd remove from request queue

	return nil
}

// handleMemPool processes a mempool message.
func (sm *SyncManager) handleMemPool(peer *Peer, msg *P2PMessage) error {
	buf := bytes.NewBuffer(msg.Payload)
	decoder := gob.NewDecoder(buf)
	memPoolMsg := &MemPoolMessage{}
	if err := decoder.Decode(memPoolMsg); err != nil {
		return fmt.Errorf("failed to decode mempool: %v", err)
	}

	fmt.Printf("Received mempool request from %s\n", peer.addr)

	// Send mempool contents (simplified implementation)
	mempool := sm.blockchain.Mempool()
	transactions := mempool.GetTransactions()

	// Send inventory of all transactions in mempool
	hashes := make([]wire.Hash, 0, len(transactions))
	for _, tx := range transactions {
		hashes = append(hashes, tx.TxHash())
	}

	inv := &InvMessage{
		Type:   "tx",
		Hashes: hashes,
	}

	return peer.SendMessage(MsgTypeInv, inv)
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
	FeeFilter    int64
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
			FeeFilter:    peer.feeFilter,
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

// peerReconnectionLoop periodically attempts to reconnect to known peers
func (sm *SyncManager) peerReconnectionLoop() {
	ticker := time.NewTicker(30 * time.Second) // Check every 30 seconds
	defer ticker.Stop()

	for range ticker.C {
		sm.mu.RLock()
		peerCount := len(sm.peers)
		sm.mu.RUnlock()

		// If we have no peers, try to discover and connect to known peers
		if peerCount == 0 {
			peerAddrs := sm.peerManager.DiscoverPeers()
			for _, addr := range peerAddrs {
				sm.mu.RLock()
				_, exists := sm.peers[addr]
				sm.mu.RUnlock()

				if !exists {
					fmt.Printf("Attempting to reconnect to peer: %s\n", addr)
					go sm.connectAndSync(addr)
				}
			}
		}
	}
}

// scheduleReconnection attempts to reconnect to a peer with exponential backoff
func (sm *SyncManager) scheduleReconnection(addr string, attempt int) {
	maxAttempts := 10 // Increased max attempts
	if attempt > maxAttempts {
		fmt.Printf("Giving up reconnection attempts to peer: %s after %d attempts\n", addr, maxAttempts)
		return
	}

	// Exponential backoff with jitter: base delay with random variation
	baseDelay := time.Duration(1<<uint(attempt-1)) * time.Second
	jitter := time.Duration(rand.Int63n(int64(baseDelay / 4))) // Add up to 25% jitter
	delay := baseDelay + jitter

	// Cap at 5 minutes
	if delay > 5*time.Minute {
		delay = 5 * time.Minute
	}

	fmt.Printf("Scheduling reconnection to %s in %v (attempt %d/%d)\n", addr, delay, attempt, maxAttempts)

	time.Sleep(delay)

	// Check if we're already connected to this peer
	sm.mu.RLock()
	peer, exists := sm.peers[addr]
	sm.mu.RUnlock()

	if !exists {
		fmt.Printf("Attempting to reconnect to peer: %s (attempt %d/%d)\n", addr, attempt, maxAttempts)
		sm.ConnectToPeer(addr)
	} else if peer.IsBanned() {
		fmt.Printf("Peer %s is banned, not attempting reconnection\n", addr)
		return
	}

	// Schedule next attempt if this one fails
	// Note: In production, you'd want to track reconnection state per peer
	go func() {
		time.Sleep(60 * time.Second) // Wait before checking if connection succeeded
		sm.mu.RLock()
		_, stillExists := sm.peers[addr]
		sm.mu.RUnlock()

		if !stillExists {
			sm.scheduleReconnection(addr, attempt+1)
		}
	}()
}

// sendReject sends a reject message to a peer.
func (sm *SyncManager) sendReject(peer *Peer, message, ccode, reason string, data []byte) error {
	rejectMsg := &RejectMessage{
		Message: message,
		CCode:   ccode,
		Reason:  reason,
		Data:    data,
	}
	return peer.SendMessage(MsgTypeReject, rejectMsg)
}

// getDisconnectReason returns a human-readable reason for disconnection
func getDisconnectReason(peer *Peer) string {
	score := peer.GetScore()
	if score <= BanThreshold {
		return "Banned (low score)"
	}
	if score < 0 {
		return "Poor performance"
	}
	return "Normal disconnection"
}
