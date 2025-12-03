package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds all configuration for the Obsidian node
type Config struct {
	// Network
	Network string

	// Logging
	LogLevel string
	LogFile  string

	// P2P
	P2PAddr        string
	MaxPeers       int
	MinPeers       int
	ConnectTimeout time.Duration
	MessageTimeout time.Duration
	MaxMessageSize int

	// RPC
	RPCAddr string

	// Mining
	MinerAddress string
	SoloMining   bool
	PoolServer   bool
	PoolAddr     string

	// Database
	DataDir string

	// Tor
	TorEnabled   bool
	TorProxyAddr string

	// Security
	BanDuration time.Duration
}

// Load loads configuration from environment variables
func Load() *Config {
	return &Config{
		Network: getEnv("NETWORK", "mainnet"),

		LogLevel: getEnv("LOG_LEVEL", "info"),
		LogFile:  getEnv("LOG_FILE", ""),

		P2PAddr:        getEnv("P2P_ADDR", "0.0.0.0:8333"),
		MaxPeers:       getEnvInt("MAX_PEERS", 125),
		MinPeers:       getEnvInt("MIN_PEERS", 8),
		ConnectTimeout: getEnvDuration("CONNECT_TIMEOUT", 30*time.Second),
		MessageTimeout: getEnvDuration("MESSAGE_TIMEOUT", 300*time.Second),
		MaxMessageSize: getEnvInt("MAX_MESSAGE_SIZE", 10*1024*1024), // 10MB

		RPCAddr: getEnv("RPC_ADDR", "0.0.0.0:8545"),

		MinerAddress: getEnv("MINER_ADDRESS", "ObsidianDefaultMinerAddress123456789"),
		SoloMining:   getEnvBool("SOLO_MINING", true),
		PoolServer:   getEnvBool("POOL_SERVER", false),
		PoolAddr:     getEnv("POOL_ADDR", "0.0.0.0:3333"),

		DataDir: getEnv("DATA_DIR", "."),

		TorEnabled:   getEnvBool("TOR_ENABLED", false),
		TorProxyAddr: getEnv("TOR_PROXY_ADDR", "127.0.0.1:9050"),

		BanDuration: getEnvDuration("BAN_DURATION", 24*time.Hour),
	}
}

// getEnv gets an environment variable or returns default
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt gets an environment variable as int or returns default
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvBool gets an environment variable as bool or returns default
func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

// getEnvDuration gets an environment variable as duration or returns default
func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
