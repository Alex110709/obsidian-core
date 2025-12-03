package chaincfg

import "obsidian-core/wire"

// CalcBlockSubsidy calculates the block reward based on block height.
// The reward halves every HalvingInterval blocks until it reaches MinimumBlockReward.
// Returns reward in satoshis (1 OBS = 100,000,000 satoshis).
func (p *Params) CalcBlockSubsidy(height int32) int64 {
	// Calculate number of halvings
	halvings := height / p.HalvingInterval

	// Reward becomes zero after 64 halvings (extremely far in future)
	if halvings >= 64 {
		return p.MinimumBlockReward * 100000000
	}

	// Start with base reward
	subsidy := p.BaseBlockReward

	// Apply halvings
	subsidy >>= uint(halvings)

	// Ensure minimum reward
	if subsidy < p.MinimumBlockReward {
		subsidy = p.MinimumBlockReward
	}

	return subsidy * 100000000 // Convert OBS to satoshis
}

// TotalSupplyAtHeight calculates the total supply at a given height.
func (p *Params) TotalSupplyAtHeight(height int32) int64 {
	total := p.InitialSupply

	for h := int32(0); h < height; h++ {
		reward := p.CalcBlockSubsidy(h)
		total += reward
	}

	return total
}

// EstimatedBlocksToMaxSupply estimates how many blocks until max supply is reached.
func (p *Params) EstimatedBlocksToMaxSupply() int32 {
	remaining := p.MaxMoney - p.InitialSupply
	blocks := int32(0)

	for remaining > 0 && blocks < 10000000 {
		reward := p.CalcBlockSubsidy(blocks)
		if reward == 0 {
			break
		}
		remaining -= reward
		blocks++
	}

	return blocks
}

// CalcTxFee calculates the minimum fee for a transaction based on its size.
func (p *Params) CalcTxFee(txSizeBytes int) int64 {
	fee := int64(txSizeBytes) * p.FeePerByte

	// Ensure minimum fee
	if fee < p.MinTxFee {
		fee = p.MinTxFee
	}

	// Cap at maximum fee
	if fee > p.MaxTxFee {
		fee = p.MaxTxFee
	}

	return fee
}

// CalcBlockFees calculates total fees in a block.
func CalcBlockFees(txs []*wire.MsgTx) int64 {
	totalFees := int64(0)

	// Skip coinbase (first transaction)
	for i := 1; i < len(txs); i++ {
		tx := txs[i]

		// Calculate input total
		inputTotal := int64(0)
		for range tx.TxIn {
			// In real implementation, look up previous outputs
			// For now, assume inputs are valid
			inputTotal += 0 // TODO: Implement UTXO lookup
		}

		// Calculate output total
		outputTotal := int64(0)
		for _, txOut := range tx.TxOut {
			outputTotal += txOut.Value
		}

		// Fee = inputs - outputs
		fee := inputTotal - outputTotal
		if fee > 0 {
			totalFees += fee
		}
	}

	return totalFees
}
