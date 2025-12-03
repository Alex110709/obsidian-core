package network

import (
	"fmt"
	"net"
	"obsidian-core/chaincfg"
	"obsidian-core/tor"
	"time"
)

// PeerManager manages P2P connections.
type PeerManager struct {
	params          *chaincfg.Params
	torClient       *tor.Client
	peers           []string
	connectionCount int
}

// NewPeerManager creates a new peer manager.
func NewPeerManager(params *chaincfg.Params, torClient *tor.Client) *PeerManager {
	return &PeerManager{
		params:          params,
		torClient:       torClient,
		peers:           make([]string, 0),
		connectionCount: 0,
	}
}

// ConnectToPeer establishes a connection to a peer.
func (pm *PeerManager) ConnectToPeer(address string) (net.Conn, error) {
	// Use Tor for .onion addresses or if Tor is enabled
	var conn net.Conn
	var err error
	if pm.torClient.IsEnabled() || isOnionAddress(address) {
		fmt.Printf("Connecting to %s via Tor...\n", address)
		conn, err = pm.torClient.DialTimeout("tcp", address, 30*time.Second)
	} else {
		// Regular connection
		fmt.Printf("Connecting to %s directly...\n", address)
		conn, err = net.DialTimeout("tcp", address, 30*time.Second)
	}

	if err == nil {
		pm.connectionCount++
	}

	return conn, err
}

// DiscoverPeers attempts to discover peers from DNS seeds, onion seeds, or manually added seeds.
func (pm *PeerManager) DiscoverPeers() []string {
	peers := make([]string, 0)

	// Add manually added seed nodes first
	if len(pm.peers) > 0 {
		peers = append(peers, pm.peers...)
		fmt.Printf("Discovered %d manual seed peers\n", len(pm.peers))
	}

	// Add Tor onion seeds if available
	if pm.torClient.IsEnabled() && len(pm.params.TorOnionSeeds) > 0 {
		peers = append(peers, pm.params.TorOnionSeeds...)
		fmt.Printf("Discovered %d Tor onion peers\n", len(pm.params.TorOnionSeeds))
	}

	// Add DNS seeds
	if len(pm.params.DNSSeeds) > 0 {
		peers = append(peers, pm.params.DNSSeeds...)
		fmt.Printf("Discovered %d DNS peers\n", len(pm.params.DNSSeeds))
	}

	// Update peers list with all discovered peers
	pm.peers = peers
	return peers
}

// GetPeers returns the list of discovered peers.
func (pm *PeerManager) GetPeers() []string {
	return pm.peers
}

// AddSeedNodes manually adds seed nodes to the peer list.
func (pm *PeerManager) AddSeedNodes(seeds []string) {
	pm.peers = append(pm.peers, seeds...)
}

// isOnionAddress checks if an address is a Tor onion address.
func isOnionAddress(address string) bool {
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		return false
	}

	// Check if it ends with .onion
	return len(host) > 6 && host[len(host)-6:] == ".onion"
}

// GetConnectionCount returns the number of active connections
func (pm *PeerManager) GetConnectionCount() int {
	return pm.connectionCount
}

// GetConnectedPeers returns the list of connected peers
func (pm *PeerManager) GetConnectedPeers() []string {
	// For now, return discovered peers as a placeholder
	// In production, you'd track actual active connections
	if pm.connectionCount > 0 && len(pm.peers) > 0 {
		// Return up to connectionCount peers
		if pm.connectionCount < len(pm.peers) {
			return pm.peers[:pm.connectionCount]
		}
		return pm.peers
	}
	return []string{}
}

// DisconnectPeer disconnects from a specific peer
func (pm *PeerManager) DisconnectPeer(address string) {
	// Decrement connection count
	if pm.connectionCount > 0 {
		pm.connectionCount--
	}
	// In production, you'd close the actual connection here
}
