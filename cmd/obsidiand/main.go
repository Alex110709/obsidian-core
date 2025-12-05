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
	"obsidian-core/rpcserver"
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
	if cfg.LogFile != "" {
		file, err := os.OpenFile(cfg.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err == nil {
			logrus.SetOutput(file)
			defer file.Close()
		} else {
			logrus.Warnf("Failed to open log file %s: %v", cfg.LogFile, err)
		}
	}

	logrus.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
	})
	logrus.SetLevel(parseLogLevel(cfg.LogLevel))

	logrus.WithFields(logrus.Fields{
		"network":  cfg.Network,
		"p2p_addr": cfg.P2PAddr,
		"rpc_addr": cfg.RPCAddr,
	}).Info("Starting Obsidian Node")

	params := chaincfg.MainNetParams
	params.TorEnabled = cfg.TorEnabled

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
		logrus.Warnf("Tor not available: %v", err)
		// Create disabled Tor client
		torClient, _ = tor.NewClient(tor.Config{Enabled: false})
	}

	torStatus := "disabled"
	if torClient.IsEnabled() {
		torStatus = fmt.Sprintf("enabled (%s)", torClient.GetProxyAddr())
		defer torClient.Stop() // Stop Tor on shutdown
	}

	// Initialize P2P Network Manager
	peerManager := network.NewPeerManager(&params, torClient)

	// Add seed nodes from environment
	seedNodesEnv := os.Getenv("SEED_NODES")
	if seedNodesEnv != "" {
		logrus.Debugf("Seed nodes: %s", seedNodesEnv)
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

	// Single startup info line
	fmt.Printf("Obsidian Node [%s] | DarkMatter PoW | Tor: %s | Block: %.0fs/%dMB | Supply: %d/%d\n",
		params.Name, torStatus, params.TargetTimePerBlock.Seconds(),
		params.BlockMaxSize/(1024*1024), params.InitialSupply, params.MaxMoney)

	// Initialize Blockchain
	chain, err := blockchain.NewBlockchain(&params, pow)
	if err != nil {
		logrus.Fatalf("Failed to initialize blockchain: %v", err)
	}
	defer chain.Close()
	logrus.Debugf("Blockchain height: %d", chain.Height())

	// Initialize P2P Sync Manager
	syncManager := network.NewSyncManager(chain, peerManager, pow)
	if err := syncManager.Start(); err != nil {
		logrus.Errorf("Failed to start sync manager: %v", err)
	}

	// Start P2P server listener for inbound connections
	p2pAddr := os.Getenv("P2P_ADDR")
	if p2pAddr == "" {
		p2pAddr = "0.0.0.0:8333" // Default P2P port
	}
	go func() {
		if err := syncManager.StartListener(p2pAddr); err != nil {
			logrus.Errorf("Failed to start P2P listener: %v", err)
		}
	}()
	logrus.Infof("P2P listening on %s (peers: %d)", p2pAddr, syncManager.GetPeerCount())

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

	// Start Solo Miner if enabled
	var miner *mining.CPUMiner
	miningStatus := "disabled"
	if enableSoloMining {
		miner = mining.NewCPUMiner(chain, &params, pow, minerAddr)
		miner.SetSyncManager(syncManager)
		go miner.Start()
		miningStatus = "solo"
	}

	// Start Pool Server if enabled
	var poolServer *stratum.StratumPool
	if enablePoolServer {
		poolServer = stratum.NewStratumPool(chain, &params, pow, minerAddr, poolListenAddr)
		if err := poolServer.Start(); err != nil {
			logrus.Errorf("Failed to start pool server: %v", err)
		} else {
			if miningStatus == "solo" {
				miningStatus = "solo+pool"
			} else {
				miningStatus = "pool"
			}
			logrus.Infof("Pool server on %s", poolListenAddr)
		}
	}

	logrus.Infof("Mining: %s | Reward: %d OBS (halves every %s blocks)",
		miningStatus, params.BaseBlockReward, formatNumber(params.HalvingInterval))

	// Initialize and Start RPC Server
	rpcServer := rpcserver.NewServer(chain, miner, syncManager, cfg.RPCAddr)

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

	logrus.Info("Shutdown signal received, initiating graceful shutdown...")
	fmt.Println("\nShutting down gracefully...")

	// Stop components in reverse order
	if poolServer != nil {
		logrus.Info("Stopping pool server...")
		poolServer.Stop()
	}

	logrus.Info("Stopping RPC server...")
	if err := rpcServer.Stop(); err != nil {
		logrus.Errorf("Error stopping RPC server: %v", err)
	}

	if miner != nil {
		logrus.Info("Stopping miner...")
		miner.Stop()
	}

	logrus.Info("Stopping sync manager...")
	syncManager.Stop()

	logrus.Info("Closing blockchain database...")
	chain.Close()

	logrus.Info("Shutdown complete")
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
