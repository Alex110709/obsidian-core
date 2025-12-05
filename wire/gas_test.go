package wire

import (
	"testing"
)

func TestCalculateIntrinsicGas(t *testing.T) {
	tests := []struct {
		name     string
		tx       *MsgTx
		expected uint64
	}{
		{
			name: "Simple transparent transaction",
			tx: &MsgTx{
				Version:  1,
				TxType:   TxTypeTransparent,
				TxIn:     []*TxIn{{SignatureScript: []byte{1, 2, 3}}},
				TxOut:    []*TxOut{{Value: 1000, PkScript: []byte{4, 5, 6}}},
				LockTime: 0,
			},
			expected: GasTxBase + 6*GasTxDataNonZero, // 21000 + 6*68 = 21408
		},
		{
			name: "Shielded transaction with 1 spend and 1 output",
			tx: &MsgTx{
				Version:         1,
				TxType:          TxTypeShielded,
				ShieldedSpends:  []*ShieldedSpend{{}},
				ShieldedOutputs: []*ShieldedOutput{{}},
			},
			expected: GasTxBase + GasShieldedSpend + GasShieldedOutput + GasShieldedVerify,
			// 21000 + 100000 + 100000 + 500000 = 721000
		},
		{
			name: "Token issuance",
			tx: &MsgTx{
				Version: 1,
				TxType:  TxTypeTokenIssue,
			},
			expected: GasTxBase + GasTokenIssue, // 21000 + 50000 = 71000
		},
		{
			name: "Smart contract deployment",
			tx: &MsgTx{
				Version: 1,
				TxType:  TxTypeSmartContractDeploy,
				Memo:    make([]byte, 100), // 100 bytes of contract code
			},
			expected: GasTxBase + GasContractDeploy + 100*GasTxDataZero*2, // code is 2x expensive
			// 21000 + 100000 + 100*4*2 = 121800
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gas := tt.tx.CalculateIntrinsicGas()
			if gas != tt.expected {
				t.Errorf("CalculateIntrinsicGas() = %d, want %d", gas, tt.expected)
			}
		})
	}
}

func TestValidateGas(t *testing.T) {
	tests := []struct {
		name      string
		tx        *MsgTx
		shouldErr bool
	}{
		{
			name: "Valid gas",
			tx: &MsgTx{
				Version:  1,
				TxType:   TxTypeTransparent,
				GasLimit: 30000,
				GasPrice: 1000,
			},
			shouldErr: false,
		},
		{
			name: "Insufficient gas limit",
			tx: &MsgTx{
				Version:  1,
				TxType:   TxTypeTransparent,
				GasLimit: 10000, // Less than intrinsic gas (21000)
				GasPrice: 1000,
			},
			shouldErr: true,
		},
		{
			name: "Zero gas price",
			tx: &MsgTx{
				Version:  1,
				TxType:   TxTypeTransparent,
				GasLimit: 30000,
				GasPrice: 0,
			},
			shouldErr: true,
		},
		{
			name: "Negative gas price",
			tx: &MsgTx{
				Version:  1,
				TxType:   TxTypeTransparent,
				GasLimit: 30000,
				GasPrice: -1000,
			},
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.tx.ValidateGas()
			if (err != nil) != tt.shouldErr {
				t.Errorf("ValidateGas() error = %v, shouldErr %v", err, tt.shouldErr)
			}
		})
	}
}

func TestGetTransactionFee(t *testing.T) {
	tx := &MsgTx{
		Version:  1,
		GasLimit: 30000,
		GasPrice: 1000, // 0.00001 OBS per gas
		GasUsed:  25000,
	}

	// Fee = gasUsed * gasPrice = 25000 * 1000 = 25,000,000 satoshis = 0.25 OBS
	expectedFee := int64(25000000)
	fee := tx.GetTransactionFee()

	if fee != expectedFee {
		t.Errorf("GetTransactionFee() = %d, want %d", fee, expectedFee)
	}

	// Test with no gas used (should use gas limit)
	tx2 := &MsgTx{
		Version:  1,
		GasLimit: 30000,
		GasPrice: 1000,
		GasUsed:  0,
	}

	expectedFee2 := int64(30000000)
	fee2 := tx2.GetTransactionFee()

	if fee2 != expectedFee2 {
		t.Errorf("GetTransactionFee() with no gasUsed = %d, want %d", fee2, expectedFee2)
	}
}

func TestSetDefaultGas(t *testing.T) {
	tx := &MsgTx{
		Version: 1,
		TxType:  TxTypeTransparent,
	}

	minGasPrice := int64(1000)
	tx.SetDefaultGas(minGasPrice)

	if tx.GasLimit == 0 {
		t.Error("SetDefaultGas() did not set GasLimit")
	}

	if tx.GasPrice != minGasPrice {
		t.Errorf("SetDefaultGas() gasPrice = %d, want %d", tx.GasPrice, minGasPrice)
	}

	// GasLimit should be at least 2x intrinsic gas
	intrinsicGas := tx.CalculateIntrinsicGas()
	expectedMinGasLimit := intrinsicGas * 2

	if tx.GasLimit < expectedMinGasLimit {
		t.Errorf("SetDefaultGas() gasLimit %d is less than 2x intrinsic %d", tx.GasLimit, expectedMinGasLimit)
	}
}

func TestCalculateBlockGasLimit(t *testing.T) {
	tests := []struct {
		name             string
		parentGasUsed    uint64
		parentGasLimit   uint64
		targetGasUsed    uint64
		minGasLimit      uint64
		maxGasLimit      uint64
		gasLimitBoundDiv uint64
		expectedChange   string // "increase", "decrease", or "same"
	}{
		{
			name:             "High usage should increase limit",
			parentGasUsed:    20000000,
			parentGasLimit:   30000000,
			targetGasUsed:    15000000,
			minGasLimit:      5000000,
			maxGasLimit:      100000000,
			gasLimitBoundDiv: 1024,
			expectedChange:   "increase",
		},
		{
			name:             "Low usage should decrease limit",
			parentGasUsed:    10000000,
			parentGasLimit:   30000000,
			targetGasUsed:    15000000,
			minGasLimit:      5000000,
			maxGasLimit:      100000000,
			gasLimitBoundDiv: 1024,
			expectedChange:   "decrease",
		},
		{
			name:             "Target usage should keep same limit",
			parentGasUsed:    15000000,
			parentGasLimit:   30000000,
			targetGasUsed:    15000000,
			minGasLimit:      5000000,
			maxGasLimit:      100000000,
			gasLimitBoundDiv: 1024,
			expectedChange:   "same",
		},
		{
			name:             "Should not exceed max limit",
			parentGasUsed:    95000000,
			parentGasLimit:   99000000,
			targetGasUsed:    50000000,
			minGasLimit:      5000000,
			maxGasLimit:      100000000,
			gasLimitBoundDiv: 1024,
			expectedChange:   "increase",
		},
		{
			name:             "Should not go below min limit",
			parentGasUsed:    1000000,
			parentGasLimit:   6000000,
			targetGasUsed:    15000000,
			minGasLimit:      5000000,
			maxGasLimit:      100000000,
			gasLimitBoundDiv: 1024,
			expectedChange:   "decrease",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newLimit := CalculateBlockGasLimit(
				tt.parentGasUsed,
				tt.parentGasLimit,
				tt.targetGasUsed,
				tt.minGasLimit,
				tt.maxGasLimit,
				tt.gasLimitBoundDiv,
			)

			// Check bounds
			if newLimit > tt.maxGasLimit {
				t.Errorf("New gas limit %d exceeds max %d", newLimit, tt.maxGasLimit)
			}
			if newLimit < tt.minGasLimit {
				t.Errorf("New gas limit %d is below min %d", newLimit, tt.minGasLimit)
			}

			// Check direction of change
			switch tt.expectedChange {
			case "increase":
				if newLimit <= tt.parentGasLimit && newLimit < tt.maxGasLimit {
					t.Errorf("Expected increase, got %d (parent was %d)", newLimit, tt.parentGasLimit)
				}
			case "decrease":
				if newLimit >= tt.parentGasLimit && newLimit > tt.minGasLimit {
					t.Errorf("Expected decrease, got %d (parent was %d)", newLimit, tt.parentGasLimit)
				}
			case "same":
				if newLimit != tt.parentGasLimit {
					t.Errorf("Expected same, got %d (parent was %d)", newLimit, tt.parentGasLimit)
				}
			}

			t.Logf("Parent: %d, New: %d, Change: %s", tt.parentGasLimit, newLimit, tt.expectedChange)
		})
	}
}

func TestValidateBlockGasUsage(t *testing.T) {
	tests := []struct {
		name          string
		txs           []*MsgTx
		blockGasLimit uint64
		shouldErr     bool
	}{
		{
			name: "Valid block gas usage",
			txs: []*MsgTx{
				{Version: 1, TxType: TxTypeTransparent, GasUsed: 21000},
				{Version: 1, TxType: TxTypeTransparent, GasUsed: 21000},
				{Version: 1, TxType: TxTypeTransparent, GasUsed: 21000},
			},
			blockGasLimit: 100000,
			shouldErr:     false,
		},
		{
			name: "Exceeds block gas limit",
			txs: []*MsgTx{
				{Version: 1, TxType: TxTypeTransparent, GasUsed: 50000},
				{Version: 1, TxType: TxTypeTransparent, GasUsed: 50000},
				{Version: 1, TxType: TxTypeTransparent, GasUsed: 50000},
			},
			blockGasLimit: 100000,
			shouldErr:     true,
		},
		{
			name: "Exactly at limit",
			txs: []*MsgTx{
				{Version: 1, TxType: TxTypeTransparent, GasUsed: 50000},
				{Version: 1, TxType: TxTypeTransparent, GasUsed: 50000},
			},
			blockGasLimit: 100000,
			shouldErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateBlockGasUsage(tt.txs, tt.blockGasLimit)
			if (err != nil) != tt.shouldErr {
				t.Errorf("ValidateBlockGasUsage() error = %v, shouldErr %v", err, tt.shouldErr)
			}
		})
	}
}
