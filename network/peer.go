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
	params    *chaincfg.Params
	torClient *tor.Client
	peers     []string
}

// NewPeerManager creates a new peer manager.
func NewPeerManager(params *chaincfg.Params, torClient *tor.Client) *PeerManager {
	return &PeerManager{
		params:    params,
		torClient: torClient,
		peers:     make([]string, 0),
	}
}

// ConnectToPeer establishes a connection to a peer.
func (pm *PeerManager) ConnectToPeer(address string) (net.Conn, error) {
	// Use Tor for .onion addresses or if Tor is enabled
	if pm.torClient.IsEnabled() || isOnionAddress(address) {
		fmt.Printf("Connecting to %s via Tor...\n", address)
		return pm.torClient.DialTimeout("tcp", address, 30*time.Second)
	}

	// Regular connection
	fmt.Printf("Connecting to %s directly...\n", address)
	return net.DialTimeout("tcp", address, 30*time.Second)
}

// DiscoverPeers attempts to discover peers from DNS seeds or onion seeds.
func (pm *PeerManager) DiscoverPeers() []string {
	peers := make([]string, 0)

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

	pm.peers = peers
	return peers
}

// GetPeers returns the list of discovered peers.
func (pm *PeerManager) GetPeers() []string {
	return pm.peers
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
