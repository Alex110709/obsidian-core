package chaincfg

import "obsidian-core/wire"

// CalcBlockSubsidy calculates the block reward based on block height.
// The reward halves every HalvingInterval blocks until it reaches MinimumBlockReward.
// Additionally adds redistribution of burned coins.
// Returns reward in satoshis (1 OBS = 100,000,000 satoshis).
func (p *Params) CalcBlockSubsidy(height int32) int64 {
	// Calculate number of halvings
	halvings := height / p.HalvingInterval

	// Reward becomes minimum after 64 halvings
	if halvings >= 64 {
		subsidy := p.MinimumBlockReward * 100000000
		return subsidy + p.CalcBurnRedistribution()
	}

	// Start with base reward
	subsidy := p.BaseBlockReward

	// Apply halvings
	subsidy >>= uint(halvings)

	// Ensure minimum reward
	if subsidy < p.MinimumBlockReward {
		subsidy = p.MinimumBlockReward
	}

	baseReward := subsidy * 100000000 // Convert OBS to satoshis

	// Add burned coin redistribution
	burnRedistribution := p.CalcBurnRedistribution()

	return baseReward + burnRedistribution
}

// CalcBurnRedistribution calculates how much burned OBS to redistribute in this block.
// Uses BurnRate (in basis points) to determine redistribution amount.
// For example: if 1M OBS burned and BurnRate is 10 (0.1%), redistribute 1,000 satoshis per block.
func (p *Params) CalcBurnRedistribution() int64 {
	if p.TotalBurned == 0 || p.BurnRate == 0 {
		return 0
	}

	// Calculate redistribution: (TotalBurned * BurnRate) / 10000
	// BurnRate is in basis points (1 basis point = 0.01%)
	redistribution := (p.TotalBurned * p.BurnRate) / 10000

	// Ensure we don't redistribute more than what's burned
	if redistribution < 0 {
		return 0
	}

	return redistribution
}

// AddBurn adds burned amount to the total burned counter.
// This increases the pool of coins available for redistribution.
func (p *Params) AddBurn(amount int64) {
	if amount > 0 {
		p.TotalBurned += amount
	}
}

// GetTotalBurned returns the total amount of OBS burned (in satoshis).
func (p *Params) GetTotalBurned() int64 {
	return p.TotalBurned
}

// GetCirculatingSupply returns the circulating supply (minted - burned).
func (p *Params) GetCirculatingSupply(height int32) int64 {
	totalMinted := p.TotalSupplyAtHeight(height)
	return totalMinted - p.TotalBurned
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
