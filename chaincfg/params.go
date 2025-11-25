package chaincfg

import (
	"math/big"
	"obsidian-core/wire"
	"time"
)

// Params defines a network configuration.
type Params struct {
	Name                     string
	Net                      uint32
	DefaultPort              string
	DNSSeeds                 []string
	GenesisBlock             *wire.MsgBlock
	PowLimit                 *big.Int
	PowLimitBits             uint32
	TargetTimespan           time.Duration
	TargetTimePerBlock       time.Duration
	RetargetAdjustmentFactor int64
	ReduceMinDifficulty      bool
	MinDiffReductionTime     time.Duration
	GenerateSupported        bool

	// Obsidian specific parameters
	BlockMaxSize  uint32
	MaxMoney      int64
	InitialSupply int64

	// Block reward parameters
	BaseBlockReward    int64 // Initial block reward
	HalvingInterval    int32 // Blocks between halvings
	MinimumBlockReward int64 // Minimum reward after halvings

	// Transaction fee parameters
	MinRelayTxFee int64 // Minimum fee to relay a transaction (satoshis per KB)
	MinTxFee      int64 // Minimum transaction fee (satoshis)
	MaxTxFee      int64 // Maximum transaction fee (satoshis)
	FeePerByte    int64 // Default fee per byte

	// Tor configuration
	TorEnabled    bool
	TorProxyAddr  string
	TorOnionSeeds []string

	// Checkpoints
	Checkpoints []Checkpoint
}

// Checkpoint represents a known good block at a specific height.
type Checkpoint struct {
	Height int32
	Hash   wire.Hash
}

var MainNetParams = Params{
	Name:                     "mainnet",
	Net:                      0x0b51d1a5, // Magic bytes for Obsidian
	DefaultPort:              "8333",
	TargetTimespan:           time.Hour * 24 * 7, // 1 week (10080 blocks at 1min each)
	TargetTimePerBlock:       time.Minute * 2,    // 2 minutes per block
	RetargetAdjustmentFactor: 4,                  // Max 4x difficulty adjustment
	ReduceMinDifficulty:      false,              // No min difficulty reduction
	MinDiffReductionTime:     time.Minute * 20,   // Min time before difficulty reduction
	GenerateSupported:        true,               // Mining supported
	PowLimit:                 nil,                // Will be set in init()
	PowLimitBits:             0x2000ffff,         // Much lower difficulty for new blockchain

	// Obsidian Specifics
	BlockMaxSize:  3200000,   // 3.2MB max block size
	MaxMoney:      100000000, // 100 Million OBS total supply
	InitialSupply: 0,         // ZERO pre-mine - fair launch!

	// Block Reward Configuration
	// Total supply: 100,000,000 OBS distributed to miners
	// Distribution: 100M OBS over ~2.1M blocks (~20 years)
	// Halving schedule ensures controlled emission
	BaseBlockReward:    100,    // 100 OBS per block initially
	HalvingInterval:    420000, // Halve every 420,000 blocks (~4 years at 5min/block)
	MinimumBlockReward: 0,      // Eventually reaches 0 (pure fee market)

	// Transaction Fee Configuration (in satoshis, 1 OBS = 100,000,000 satoshis)
	MinRelayTxFee: 1000,      // 0.00001 OBS per KB minimum to relay
	MinTxFee:      10000,     // 0.0001 OBS minimum transaction fee
	MaxTxFee:      100000000, // 1 OBS maximum transaction fee
	FeePerByte:    10,        // 0.0000001 OBS per byte default

	// Tor Configuration
	TorEnabled:    false,            // Disabled by default
	TorProxyAddr:  "127.0.0.1:9050", // Default Tor SOCKS5 proxy
	TorOnionSeeds: []string{
		// Add .onion seed nodes here when available
	},

	// Checkpoints - Known good blocks at specific heights
	// These prevent long reorganizations and speed up initial sync
	Checkpoints: []Checkpoint{
		// Genesis block checkpoint
		// {Height: 0, Hash: [genesis hash]}, // Will be set after genesis is mined
		// Add more checkpoints as network matures
		// {Height: 10000, Hash: [block hash]},
		// {Height: 50000, Hash: [block hash]},
	},
}

func init() {
	// Set PowLimit to maximum target (minimum difficulty)
	// This is ~256 bits of 1s shifted to create a very large number
	MainNetParams.PowLimit = new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 224), big.NewInt(1))
}
