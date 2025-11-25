package network

import (
	"net"
	"obsidian-core/chaincfg"
	"obsidian-core/tor"
	"testing"
	"time"
)

func TestNewPeerManager(t *testing.T) {
	params := &chaincfg.Params{
		DNSSeeds:      []string{"seed1.example.com", "seed2.example.com"},
		TorOnionSeeds: []string{"onion1.onion", "onion2.onion"},
	}
	torClient, err := tor.NewClient(tor.Config{Enabled: false})
	if err != nil {
		t.Fatalf("Failed to create tor client: %v", err)
	}

	pm := NewPeerManager(params, torClient)

	if pm.params != params {
		t.Errorf("Expected params to be set correctly")
	}
	if pm.torClient != torClient {
		t.Errorf("Expected torClient to be set correctly")
	}
	if len(pm.peers) != 0 {
		t.Errorf("Expected peers to be empty initially")
	}
}

func TestDiscoverPeers(t *testing.T) {
	params := &chaincfg.Params{
		DNSSeeds:      []string{"seed1.example.com", "seed2.example.com"},
		TorOnionSeeds: []string{"onion1.onion", "onion2.onion"},
	}
	torClient, err := tor.NewClient(tor.Config{Enabled: false})
	if err != nil {
		t.Fatalf("Failed to create tor client: %v", err)
	}
	pm := NewPeerManager(params, torClient)

	peers := pm.DiscoverPeers()

	expectedPeers := []string{"seed1.example.com", "seed2.example.com"}
	if len(peers) != len(expectedPeers) {
		t.Errorf("Expected %d peers, got %d", len(expectedPeers), len(peers))
	}
	for i, peer := range peers {
		if peer != expectedPeers[i] {
			t.Errorf("Expected peer %s, got %s", expectedPeers[i], peer)
		}
	}
}

func TestConnectToPeer(t *testing.T) {
	params := &chaincfg.Params{}
	torClient, err := tor.NewClient(tor.Config{Enabled: false})
	if err != nil {
		t.Fatalf("Failed to create tor client: %v", err)
	}
	pm := NewPeerManager(params, torClient)

	// Test with a mock server
	go func() {
		listener, _ := net.Listen("tcp", "127.0.0.1:0")
		defer listener.Close()
		conn, _ := listener.Accept()
		conn.Close()
	}()

	// Wait a bit for server to start
	time.Sleep(100 * time.Millisecond)

	// This is a basic test; in real scenarios, use mocks for Tor
	// For now, assume direct connection works
	_, err = pm.ConnectToPeer("127.0.0.1:12345") // Invalid address for test
	if err == nil {
		t.Errorf("Expected connection to fail for invalid address")
	}
}

func TestIsOnionAddress(t *testing.T) {
	tests := []struct {
		address  string
		expected bool
	}{
		{"example.onion:8333", true},
		{"192.168.1.1:8333", false},
		{"invalid", false},
	}

	for _, test := range tests {
		result := isOnionAddress(test.address)
		if result != test.expected {
			t.Errorf("isOnionAddress(%s) = %v; expected %v", test.address, result, test.expected)
		}
	}
}
