package wire

import (
	"fmt"
)

// Gas cost constants (Ethereum-style)
const (
	// Base transaction costs
	GasTxBase        uint64 = 21000 // Base cost of a transaction
	GasTxDataZero    uint64 = 4     // Cost per zero byte in transaction data
	GasTxDataNonZero uint64 = 68    // Cost per non-zero byte in transaction data
	GasTxCreate      uint64 = 32000 // Cost to create a contract

	// Shielded transaction costs (higher due to zk-SNARK verification)
	GasShieldedSpend  uint64 = 100000 // Cost per shielded spend
	GasShieldedOutput uint64 = 100000 // Cost per shielded output
	GasShieldedVerify uint64 = 500000 // Cost to verify zk-SNARK proof

	// Token operation costs
	GasTokenIssue    uint64 = 50000 // Cost to issue a token
	GasTokenTransfer uint64 = 30000 // Cost to transfer a token
	GasTokenMint     uint64 = 40000 // Cost to mint tokens
	GasTokenBurn     uint64 = 30000 // Cost to burn tokens

	// Smart contract costs
	GasContractDeploy  uint64 = 100000 // Base cost to deploy a contract
	GasContractCall    uint64 = 40000  // Base cost to call a contract
	GasContractStorage uint64 = 20000  // Cost per storage operation

	// Memory and computation costs
	GasMemory      uint64 = 3  // Cost per word of memory
	GasComputation uint64 = 10 // Cost per computation step
)

// CalculateIntrinsicGas calculates the intrinsic gas cost of a transaction
// This is the gas required before any execution (similar to Ethereum)
func (tx *MsgTx) CalculateIntrinsicGas() uint64 {
	gas := GasTxBase

	// Add data costs
	for _, txIn := range tx.TxIn {
		gas += calculateDataGas(txIn.SignatureScript)
	}

	for _, txOut := range tx.TxOut {
		gas += calculateDataGas(txOut.PkScript)
	}

	// Add costs based on transaction type
	switch tx.TxType {
	case TxTypeTransparent:
		// Base cost already included

	case TxTypeShielded:
		// Shielded transactions are more expensive
		gas += uint64(len(tx.ShieldedSpends)) * GasShieldedSpend
		gas += uint64(len(tx.ShieldedOutputs)) * GasShieldedOutput
		gas += GasShieldedVerify // zk-SNARK verification cost

	case TxTypeMixed:
		// Mixed transactions include both transparent and shielded
		gas += uint64(len(tx.ShieldedSpends)) * GasShieldedSpend
		gas += uint64(len(tx.ShieldedOutputs)) * GasShieldedOutput
		if len(tx.ShieldedSpends) > 0 || len(tx.ShieldedOutputs) > 0 {
			gas += GasShieldedVerify
		}

	case TxTypeTokenIssue:
		gas += GasTokenIssue

	case TxTypeTokenTransfer:
		gas += GasTokenTransfer

	case TxTypeTokenMint:
		gas += GasTokenMint

	case TxTypeTokenBurn:
		gas += GasTokenBurn

	case TxTypeTokenTransferOwnership:
		gas += GasTokenTransfer

	case TxTypeTokenShielded:
		gas += GasTokenTransfer
		gas += uint64(len(tx.ShieldedSpends)) * GasShieldedSpend
		gas += uint64(len(tx.ShieldedOutputs)) * GasShieldedOutput
		gas += GasShieldedVerify

	case TxTypeSmartContractDeploy:
		gas += GasContractDeploy
		// Add cost for contract code size
		if len(tx.Memo) > 0 {
			gas += calculateDataGas(tx.Memo) * 2 // Contract code is more expensive
		}

	case TxTypeSmartContractCall:
		gas += GasContractCall
		// Execution cost will be added during execution
	}

	// Add memo data cost if present
	if len(tx.Memo) > 0 && tx.TxType != TxTypeSmartContractDeploy {
		gas += calculateDataGas(tx.Memo)
	}

	return gas
}

// calculateDataGas calculates gas cost for data bytes
func calculateDataGas(data []byte) uint64 {
	var gas uint64
	for _, b := range data {
		if b == 0 {
			gas += GasTxDataZero
		} else {
			gas += GasTxDataNonZero
		}
	}
	return gas
}

// ValidateGas validates that the transaction has sufficient gas
func (tx *MsgTx) ValidateGas() error {
	intrinsicGas := tx.CalculateIntrinsicGas()

	if tx.GasLimit < intrinsicGas {
		return fmt.Errorf("gas limit %d is less than intrinsic gas %d", tx.GasLimit, intrinsicGas)
	}

	if tx.GasPrice <= 0 {
		return fmt.Errorf("gas price must be positive, got %d", tx.GasPrice)
	}

	return nil
}

// GetTransactionFee returns the transaction fee based on gas used
func (tx *MsgTx) GetTransactionFee() int64 {
	if tx.GasUsed > 0 {
		return int64(tx.GasUsed) * tx.GasPrice
	}
	// If not executed yet, estimate with gas limit
	return int64(tx.GasLimit) * tx.GasPrice
}

// SetDefaultGas sets default gas values for a transaction
func (tx *MsgTx) SetDefaultGas(minGasPrice int64) {
	if tx.GasLimit == 0 {
		tx.GasLimit = tx.CalculateIntrinsicGas() * 2 // 2x intrinsic gas as default
	}
	if tx.GasPrice == 0 {
		tx.GasPrice = minGasPrice
	}
}

// CalculateBlockGasLimit calculates the gas limit for a new block
// based on the parent block's gas used and limit (EIP-1559 style)
func CalculateBlockGasLimit(parentGasUsed, parentGasLimit, targetGasUsed, minGasLimit, maxGasLimit uint64, gasLimitBoundDiv uint64) uint64 {
	// Calculate adjustment
	delta := parentGasLimit / gasLimitBoundDiv

	if parentGasUsed > targetGasUsed {
		// Increase gas limit if usage is high
		newLimit := parentGasLimit + delta
		if newLimit > maxGasLimit {
			return maxGasLimit
		}
		return newLimit
	} else if parentGasUsed < targetGasUsed {
		// Decrease gas limit if usage is low
		if delta > parentGasLimit-minGasLimit {
			return minGasLimit
		}
		newLimit := parentGasLimit - delta
		if newLimit < minGasLimit {
			return minGasLimit
		}
		return newLimit
	}

	// No change if at target
	return parentGasLimit
}

// ValidateBlockGasUsage validates that total gas used in block doesn't exceed limit
func ValidateBlockGasUsage(txs []*MsgTx, blockGasLimit uint64) error {
	var totalGasUsed uint64

	for _, tx := range txs {
		intrinsicGas := tx.CalculateIntrinsicGas()
		if tx.GasUsed > 0 {
			totalGasUsed += tx.GasUsed
		} else {
			totalGasUsed += intrinsicGas
		}

		if totalGasUsed > blockGasLimit {
			return fmt.Errorf("block gas usage %d exceeds limit %d", totalGasUsed, blockGasLimit)
		}
	}

	return nil
}
