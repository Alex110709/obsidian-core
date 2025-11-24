package chaincfg

import (
	"obsidian-core/wire"
	"time"
)

// GenesisBlock defines the first block of the chain.
// FAIR LAUNCH: No pre-mine, genesis block must be mined!
var GenesisBlock = wire.MsgBlock{
	Header: wire.BlockHeader{
		Version:   1,
		Timestamp: time.Date(2025, 11, 23, 0, 0, 0, 0, time.UTC), // Fair launch date
		Bits:      0x2000ffff,                                    // Initial difficulty
		Nonce:     0,                                             // Must be mined!
	},
	Transactions: []*wire.MsgTx{GenesisCoinbaseTx},
}

// GenesisCoinbaseTx is the coinbase transaction in the genesis block.
// FAIR LAUNCH: Genesis block reward goes to the miner who finds it!
// No pre-allocation, no founder rewards - 100% fair distribution
var GenesisCoinbaseTx = &wire.MsgTx{
	Version: 1,
	TxIn: []*wire.TxIn{
		{
			PreviousOutPoint: wire.OutPoint{
				Hash:  wire.Hash{},
				Index: 0xffffffff,
			},
			SignatureScript: []byte{
				0x04, 0xff, 0xff, 0x00, 0x1d, 0x01, 0x04, // Block height 0
				// Timestamp message for genesis block
				'2', '3', '/', 'N', 'o', 'v', '/', '2', '0', '2', '5', ' ',
				'O', 'b', 's', 'i', 'd', 'i', 'a', 'n', ' ',
				'F', 'a', 'i', 'r', ' ', 'L', 'a', 'u', 'n', 'c', 'h', ' ',
				'-', ' ', 'N', 'o', ' ', 'P', 'r', 'e', 'm', 'i', 'n', 'e',
			},
			Sequence: 0xffffffff,
		},
	},
	TxOut: []*wire.TxOut{
		{
			// Genesis block reward: 100 OBS (same as all other blocks)
			// NO special allocation - miner gets same reward as any other block!
			Value:    10000000000, // 100 OBS (100M satoshis)
			PkScript: []byte{},    // Will be set by miner who finds genesis block
		},
	},
	LockTime: 0,
}

func init() {
	// Genesis block must be mined - no preset output!
	// Miner address will be set when genesis block is found
	MainNetParams.GenesisBlock = &GenesisBlock
}
