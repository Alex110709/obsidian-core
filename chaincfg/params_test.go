package chaincfg

import (
	"testing"
	"time"
)

func TestObsidianParams(t *testing.T) {
	params := MainNetParams

	// 1. Verify Block Size (3.2MB)
	if params.BlockMaxSize != 3200000 {
		t.Errorf("BlockMaxSize is %d, want 3200000", params.BlockMaxSize)
	}

	// 2. Verify Max Supply (100M)
	if params.MaxMoney != 100000000 {
		t.Errorf("MaxMoney is %d, want 100000000", params.MaxMoney)
	}

	// 3. Verify Initial Supply (0 for fair launch)
	if params.InitialSupply != 0 {
		t.Errorf("InitialSupply is %d, want 0 (fair launch)", params.InitialSupply)
	}

	// 4. Verify Block Time (2 minutes)
	if params.TargetTimePerBlock != time.Minute*2 {
		t.Errorf("TargetTimePerBlock is %v, want 2m0s", params.TargetTimePerBlock)
	}

	// 5. Verify Network Name
	if params.Name != "mainnet" {
		t.Errorf("Name is %s, want mainnet", params.Name)
	}
}

func TestBlockReward(t *testing.T) {
	params := MainNetParams

	// Test initial reward (100 OBS = 10,000,000,000 satoshis)
	reward0 := params.CalcBlockSubsidy(0)
	if reward0 != 10000000000 {
		t.Errorf("Block 0 reward = %d, want 10000000000", reward0)
	}

	// Test before first halving
	reward100 := params.CalcBlockSubsidy(100)
	if reward100 != 10000000000 {
		t.Errorf("Block 100 reward = %d, want 10000000000", reward100)
	}

	// Test after first halving (420000 blocks) - should be 50 OBS = 5,000,000,000 satoshis
	reward420000 := params.CalcBlockSubsidy(420000)
	if reward420000 != 5000000000 {
		t.Errorf("Block 420000 reward = %d, want 5000000000", reward420000)
	}

	// Test after second halving (840000 blocks) - should be 25 OBS = 2,500,000,000 satoshis
	reward840000 := params.CalcBlockSubsidy(840000)
	if reward840000 != 2500000000 {
		t.Errorf("Block 840000 reward = %d, want 2500000000", reward840000)
	}

	// Test after third halving (1260000 blocks) - should be 12 OBS = 1,200,000,000 satoshis
	reward1260000 := params.CalcBlockSubsidy(1260000)
	if reward1260000 != 1200000000 {
		t.Errorf("Block 1260000 reward = %d, want 1200000000", reward1260000)
	}

	// Test minimum reward (after many halvings)
	rewardHigh := params.CalcBlockSubsidy(10000000)
	if rewardHigh < params.MinimumBlockReward {
		t.Errorf("Reward should not be less than minimum %d, got %d", params.MinimumBlockReward, rewardHigh)
	}
}

func TestTotalSupply(t *testing.T) {
	params := MainNetParams

	// Test initial supply
	supply0 := params.TotalSupplyAtHeight(0)
	if supply0 != params.InitialSupply {
		t.Errorf("Supply at height 0 = %d, want %d", supply0, params.InitialSupply)
	}

	// Test supply after 100 blocks (100 blocks * 100 OBS reward = 10,000 OBS = 1,000,000,000,000 satoshis)
	supply100 := params.TotalSupplyAtHeight(100)
	expected := params.InitialSupply + (100 * 10000000000)
	if supply100 != expected {
		t.Errorf("Supply at height 100 = %d, want %d", supply100, expected)
	}

	// Test that supply doesn't exceed max
	supplyMax := params.TotalSupplyAtHeight(10000000)
	maxSupplySatoshis := params.MaxMoney * 100000000
	if supplyMax > maxSupplySatoshis {
		t.Errorf("Supply %d exceeds max %d", supplyMax, maxSupplySatoshis)
	}
}
