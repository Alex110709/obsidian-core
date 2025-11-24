package blockchain

import (
	"fmt"
	"obsidian-core/wire"
	"sync"
	"time"
)

// FeeEstimator estimates transaction fees based on recent blocks.
type FeeEstimator struct {
	mu sync.RWMutex

	// Recent blocks (last 100)
	recentBlocks []*BlockFeeData
	maxBlocks    int

	// Fee rate buckets (satoshis per KB)
	buckets []int64
}

// BlockFeeData contains fee statistics for a block.
type BlockFeeData struct {
	Height       int32
	Timestamp    time.Time
	Transactions int
	TotalFees    int64
	TotalSize    int64
	MedianFee    int64
	MinFee       int64
	MaxFee       int64
}

// NewFeeEstimator creates a new fee estimator.
func NewFeeEstimator() *FeeEstimator {
	return &FeeEstimator{
		recentBlocks: make([]*BlockFeeData, 0, 100),
		maxBlocks:    100,
		buckets: []int64{
			1000,   // 0.00001 OBS/KB (minimum)
			5000,   // 0.00005 OBS/KB (low)
			10000,  // 0.0001 OBS/KB (medium)
			25000,  // 0.00025 OBS/KB (high)
			50000,  // 0.0005 OBS/KB (very high)
			100000, // 0.001 OBS/KB (priority)
		},
	}
}

// AddBlock adds a block's fee data to the estimator.
func (fe *FeeEstimator) AddBlock(block *wire.MsgBlock, height int32) {
	fe.mu.Lock()
	defer fe.mu.Unlock()

	// Calculate block fee statistics
	data := fe.calculateBlockFees(block, height)

	// Add to recent blocks
	fe.recentBlocks = append(fe.recentBlocks, data)

	// Keep only last maxBlocks
	if len(fe.recentBlocks) > fe.maxBlocks {
		fe.recentBlocks = fe.recentBlocks[1:]
	}
}

// calculateBlockFees calculates fee statistics for a block.
func (fe *FeeEstimator) calculateBlockFees(block *wire.MsgBlock, height int32) *BlockFeeData {
	data := &BlockFeeData{
		Height:    height,
		Timestamp: block.Header.Timestamp,
	}

	fees := make([]int64, 0)
	totalSize := int64(0)

	// Skip coinbase transaction
	for i := 1; i < len(block.Transactions); i++ {
		tx := block.Transactions[i]

		// Calculate transaction size (simplified)
		txSize := int64(len(tx.TxIn)*180 + len(tx.TxOut)*34 + 10)
		totalSize += txSize

		// Calculate fee (input value - output value)
		// For now, we'll estimate based on output values
		// In production, you'd look up input values from UTXO set
		fee := int64(0)
		outputCount := len(tx.TxOut)
		if outputCount > 0 {
			// Simplified: assume some fee per output
			fee += int64(outputCount) * 10000 // 0.0001 OBS per output
		}

		fees = append(fees, fee)
		data.TotalFees += fee
	}

	data.Transactions = len(fees)
	data.TotalSize = totalSize

	if len(fees) > 0 {
		// Calculate median, min, max
		data.MinFee = fees[0]
		data.MaxFee = fees[0]
		sum := int64(0)

		for _, fee := range fees {
			if fee < data.MinFee {
				data.MinFee = fee
			}
			if fee > data.MaxFee {
				data.MaxFee = fee
			}
			sum += fee
		}

		// Simplified median (just use average for now)
		data.MedianFee = sum / int64(len(fees))
	}

	return data
}

// EstimateFee estimates the fee for a transaction to be confirmed within targetBlocks.
func (fe *FeeEstimator) EstimateFee(txSize int64, targetBlocks int) int64 {
	fe.mu.RLock()
	defer fe.mu.RUnlock()

	if len(fe.recentBlocks) == 0 {
		// No data, return minimum fee
		return 1000 * (txSize / 1024)
	}

	// Calculate average fee rate from recent blocks
	totalFeeRate := int64(0)
	count := 0

	for i := len(fe.recentBlocks) - 1; i >= 0 && count < targetBlocks; i-- {
		block := fe.recentBlocks[i]
		if block.TotalSize > 0 {
			// Fee rate in satoshis per KB
			feeRate := (block.TotalFees * 1024) / block.TotalSize
			totalFeeRate += feeRate
			count++
		}
	}

	if count == 0 {
		// Fallback to minimum
		return 1000 * (txSize / 1024)
	}

	avgFeeRate := totalFeeRate / int64(count)

	// Apply multiplier based on target blocks
	multiplier := float64(1.0)
	switch {
	case targetBlocks == 1:
		multiplier = 2.0 // High priority - double fee
	case targetBlocks <= 3:
		multiplier = 1.5 // Medium-high priority
	case targetBlocks <= 6:
		multiplier = 1.2 // Medium priority
	case targetBlocks <= 12:
		multiplier = 1.0 // Normal priority
	default:
		multiplier = 0.8 // Low priority
	}

	estimatedFeeRate := int64(float64(avgFeeRate) * multiplier)

	// Ensure minimum fee rate
	if estimatedFeeRate < 1000 {
		estimatedFeeRate = 1000
	}

	// Calculate total fee
	fee := (estimatedFeeRate * txSize) / 1024

	// Ensure minimum fee
	if fee < 10000 {
		fee = 10000
	}

	return fee
}

// EstimatePriority estimates fee for different priority levels.
func (fe *FeeEstimator) EstimatePriority(txSize int64) map[string]int64 {
	return map[string]int64{
		"minimum":  fe.EstimateFee(txSize, 24), // Confirm within 24 blocks (~2 hours)
		"low":      fe.EstimateFee(txSize, 12), // Confirm within 12 blocks (~1 hour)
		"medium":   fe.EstimateFee(txSize, 6),  // Confirm within 6 blocks (~30 min)
		"high":     fe.EstimateFee(txSize, 3),  // Confirm within 3 blocks (~15 min)
		"priority": fe.EstimateFee(txSize, 1),  // Confirm in next block
	}
}

// GetFeeStats returns fee statistics from recent blocks.
func (fe *FeeEstimator) GetFeeStats() string {
	fe.mu.RLock()
	defer fe.mu.RUnlock()

	if len(fe.recentBlocks) == 0 {
		return "No recent blocks"
	}

	totalTxs := 0
	totalFees := int64(0)

	for _, block := range fe.recentBlocks {
		totalTxs += block.Transactions
		totalFees += block.TotalFees
	}

	avgFee := int64(0)
	if totalTxs > 0 {
		avgFee = totalFees / int64(totalTxs)
	}

	return fmt.Sprintf("Recent %d blocks: %d txs, avg fee: %d satoshis",
		len(fe.recentBlocks), totalTxs, avgFee)
}
