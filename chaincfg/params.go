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

	// Gas parameters (Ethereum-style)
	BlockGasLimit    uint64 // Maximum gas per block
	MinGasLimit      uint64 // Minimum gas limit
	MaxGasLimit      uint64 // Maximum gas limit
	GasLimitBoundDiv uint64 // Gas limit adjustment divisor
	TargetGasUsed    uint64 // Target gas usage per block

	// Block reward parameters
	BaseBlockReward    int64 // Initial block reward
	HalvingInterval    int32 // Blocks between halvings
	MinimumBlockReward int64 // Minimum reward after halvings

	// Burn tracking
	TotalBurned int64 // Total amount burned (in satoshis)
	BurnRate    int64 // Percentage of burned coins to redistribute per block (basis points, e.g. 1 = 0.01%)

	// Transaction fee parameters (Gas-based)
	MinGasPrice   int64 // Minimum gas price (satoshis per gas unit)
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
	TargetTimespan:           time.Hour * 24,   // 1 day (4320 blocks at 20sec each)
	TargetTimePerBlock:       time.Second * 20, // 20 seconds per block (Ethereum-like)
	RetargetAdjustmentFactor: 4,                // Max 4x difficulty adjustment
	ReduceMinDifficulty:      false,            // No min difficulty reduction
	MinDiffReductionTime:     time.Minute * 5,  // Min time before difficulty reduction
	GenerateSupported:        true,             // Mining supported
	PowLimit:                 nil,              // Will be set in init()
	PowLimitBits:             0x2000ffff,       // Much lower difficulty for new blockchain

	// Obsidian Specifics
	BlockMaxSize:  3200000,   // 3.2MB max block size (still used as fallback)
	MaxMoney:      100000000, // 100 Million OBS total supply
	InitialSupply: 0,         // ZERO pre-mine - fair launch!

	// Gas Configuration (Ethereum-style)
	BlockGasLimit:    30000000,  // 30M gas per block (similar to Ethereum)
	MinGasLimit:      5000000,   // 5M gas minimum
	MaxGasLimit:      100000000, // 100M gas maximum
	GasLimitBoundDiv: 1024,      // Gas limit can adjust by 1/1024 per block
	TargetGasUsed:    15000000,  // Target 50% gas usage

	// Block Reward Configuration
	// With 20 second blocks: ~4,320 blocks per day, ~1,577,000 blocks per year
	// Total supply: 100M OBS distributed over time with halvings
	// Year 1-4: 50M OBS (25 OBS/block * ~6.3M blocks)
	// Year 5-8: 25M OBS (12.5 OBS/block * ~6.3M blocks)
	// Year 9-12: 12.5M OBS (6.25 OBS/block * ~6.3M blocks)
	// Year 13+: Remaining + burned coins redistributed
	BaseBlockReward:    25,      // 25 OBS per block initially
	HalvingInterval:    1577000, // Halve every ~1.577M blocks (~1 year at 20sec/block)
	MinimumBlockReward: 1,       // Minimum 1 OBS per block (plus burned redistribution)

	// Transaction Fee Configuration (Gas-based, in satoshis)
	MinGasPrice:   1000,      // 0.00001 OBS per gas unit minimum
	MinRelayTxFee: 1000,      // 0.00001 OBS per KB minimum to relay
	MinTxFee:      21000,     // 21,000 gas * minGasPrice = 0.00021 OBS minimum
	MaxTxFee:      100000000, // 1 OBS maximum transaction fee
	FeePerByte:    10,        // 0.0000001 OBS per byte default

	// Burn Configuration
	TotalBurned: 0,  // No initial burn
	BurnRate:    10, // 0.1% of total burned redistributed per block (10 basis points)

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
