package chaincfg

import (
	"time"
)

// SecurityParams defines security-related parameters
type SecurityParams struct {
	// MaxBlockSize enforces a limit on block size
	MaxBlockSize uint32

	// MaxTxSize enforces a limit on transaction size
	MaxTxSize uint32

	// MaxOrphanBlocks is the maximum number of orphan blocks to keep
	MaxOrphanBlocks uint32

	// MaxOrphanTxs is the maximum number of orphan transactions to keep
	MaxOrphanTxs uint32

	// MinRelayTxFee is the minimum fee for a transaction to be relayed
	MinRelayTxFee int64

	// MaxSigOpsPerTx maximum signature operations per transaction
	MaxSigOpsPerTx uint32

	// BlockMaxAge is the maximum age of a block before it's considered stale
	BlockMaxAge time.Duration
}

var MainNetSecurityParams = SecurityParams{
	MaxBlockSize:    6000000, // 6MB
	MaxTxSize:       1000000, // 1MB
	MaxOrphanBlocks: 100,
	MaxOrphanTxs:    1000,
	MinRelayTxFee:   1000, // 0.00001 coins
	MaxSigOpsPerTx:  20000,
	BlockMaxAge:     24 * time.Hour,
}
