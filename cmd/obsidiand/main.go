package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/sirupsen/logrus"
	"obsidian-core/blockchain"
	"obsidian-core/chaincfg"
	"obsidian-core/config"
	"obsidian-core/consensus"
	"obsidian-core/mining"
	"obsidian-core/network"
	"obsidian-core/rpc"
	"obsidian-core/stratum"
	"obsidian-core/tor"
)

func parseLogLevel(level string) logrus.Level {
	switch level {
	case "debug":
		return logrus.DebugLevel
	case "info":
		return logrus.InfoLevel
	case "warn":
		return logrus.WarnLevel
	case "error":
		return logrus.ErrorLevel
	default:
		return logrus.InfoLevel
	}
}

func main() {
	// Load configuration
	cfg := config.Load()

	// Setup logging
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetLevel(parseLogLevel(cfg.LogLevel))

	logrus.WithFields(logrus.Fields{
		"network":  cfg.Network,
		"p2p_addr": cfg.P2PAddr,
		"rpc_addr": cfg.RPCAddr,
	}).Info("Starting Obsidian Node")

	params := chaincfg.MainNetParams
	params.TorEnabled = cfg.TorEnabled

	fmt.Printf("Network: %s\n", params.Name)
	fmt.Printf("Block Size Limit: %d bytes\n", params.BlockMaxSize)
	fmt.Printf("Target Block Time: %v\n", params.TargetTimePerBlock)
	fmt.Printf("Max Supply: %d\n", params.MaxMoney)
	fmt.Printf("Initial Supply: %d\n", params.InitialSupply)

	// Initialize Tor
	torConfig := tor.Config{
		Enabled:   params.TorEnabled,
		ProxyAddr: params.TorProxyAddr,
	}
	torClient, err := tor.NewClient(torConfig)
	if err != nil {
		if params.TorEnabled {
			logrus.Fatalf("Failed to initialize Tor (required): %v", err)
		}
		fmt.Printf("Warning: Tor not available: %v\n", err)
		// Create disabled Tor client
		torClient, _ = tor.NewClient(tor.Config{Enabled: false})
	}
	if torClient.IsEnabled() {
		fmt.Printf("Tor enabled via proxy: %s\n", torClient.GetProxyAddr())
		defer torClient.Stop() // Stop Tor on shutdown
	} else {
		fmt.Println("Tor disabled - using direct connections")
	}

	// Initialize P2P Network Manager
	peerManager := network.NewPeerManager(&params, torClient)

	// Add seed nodes from environment
	seedNodesEnv := os.Getenv("SEED_NODES")
	if seedNodesEnv != "" {
		fmt.Printf("Seed nodes configured: %s\n", seedNodesEnv)
		// Parse comma-separated seed nodes
		var seedNodes []string
		for _, seed := range strings.Split(seedNodesEnv, ",") {
			seed = strings.TrimSpace(seed)
			if seed != "" {
				seedNodes = append(seedNodes, seed)
			}
		}
		if len(seedNodes) > 0 {
			peerManager.AddSeedNodes(seedNodes)
		}
	}

	// Initialize PoW
	pow := consensus.NewDarkMatter()
	fmt.Println("PoW Engine: DarkMatter (AES-SHA256 Hybrid)")

	// Initialize Blockchain
	chain, err := blockchain.NewBlockchain(&params, pow)
	if err != nil {
		logrus.Fatalf("Failed to initialize blockchain: %v", err)
	}
	defer chain.Close()
	fmt.Printf("Blockchain initialized. Height: %d\n", chain.Height())

	// Initialize P2P Sync Manager
	syncManager := network.NewSyncManager(chain, peerManager, pow)
	if err := syncManager.Start(); err != nil {
		logrus.Printf("Failed to start sync manager: %v", err)
	} else {
		fmt.Printf("P2P sync started. Connected peers: %d\n", syncManager.GetPeerCount())
	}

	// Start P2P server listener for inbound connections
	p2pAddr := os.Getenv("P2P_ADDR")
	if p2pAddr == "" {
		p2pAddr = "0.0.0.0:8333" // Default P2P port
	}
	go func() {
		if err := syncManager.StartListener(p2pAddr); err != nil {
			logrus.Printf("Failed to start P2P listener: %v", err)
		}
	}()
	fmt.Printf("P2P server listening on %s\n", p2pAddr)

	// Initialize mining configuration
	minerAddr := os.Getenv("MINER_ADDRESS")
	if minerAddr == "" {
		minerAddr = "ObsidianDefaultMinerAddress123456789" // Default address
	}

	enableSoloMining := os.Getenv("SOLO_MINING") != "false" // Default: enabled
	enablePoolServer := os.Getenv("POOL_SERVER") == "true"  // Default: disabled
	poolListenAddr := os.Getenv("POOL_LISTEN")
	if poolListenAddr == "" {
		poolListenAddr = "0.0.0.0:3333" // Default Stratum port
	}

	fmt.Println("Mining Configuration")
	fmt.Printf("  Miner Address:  %s\n", minerAddr)
	fmt.Printf("  Block Reward:   %d OBS (halves every %s blocks)\n",
		params.BaseBlockReward, formatNumber(params.HalvingInterval))
	fmt.Printf("  Solo Mining:    %v\n", enableSoloMining)
	fmt.Printf("  Pool Server:    %v\n", enablePoolServer)
	if enablePoolServer {
		fmt.Printf("  Pool Listen:    stratum+tcp://%s\n", poolListenAddr)
	}
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Start Solo Miner if enabled
	var miner *mining.CPUMiner
	if enableSoloMining {
		miner = mining.NewCPUMiner(chain, &params, pow, minerAddr)
		miner.SetSyncManager(syncManager)
		go miner.Start()
		fmt.Println("✓ Solo mining started")
	} else {
		fmt.Println("✗ Solo mining disabled")
	}

	// Start Pool Server if enabled
	var poolServer *stratum.StratumPool
	if enablePoolServer {
		poolServer = stratum.NewStratumPool(chain, &params, pow, minerAddr, poolListenAddr)
		if err := poolServer.Start(); err != nil {
			logrus.Printf("Failed to start pool server: %v", err)
		} else {
			fmt.Println("✓ Pool server started")
		}
	}

	// Initialize and Start RPC Server
	rpcAddr := os.Getenv("RPC_ADDR")
	if rpcAddr == "" {
		rpcAddr = "0.0.0.0:8545" // Default RPC address
	}
	rpcServer := rpc.NewServer(chain, miner, syncManager, rpcAddr)

	// Connect pool server to RPC if enabled
	if poolServer != nil {
		rpcServer.SetPoolServer(poolServer)
	}

	go func() {
		if err := rpcServer.Start(); err != nil {
			logrus.Printf("RPC server error: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	fmt.Println("\nShutting down...")

	if poolServer != nil {
		poolServer.Stop()
	}
	rpcServer.Stop()

	fmt.Println("Shutdown complete")
}

// formatNumber formats a number with thousand separators
func formatNumber(n int32) string {
	s := fmt.Sprintf("%d", n)
	var result string
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			result += ","
		}
		result += string(c)
	}
	return result
}
