package chaincfg

import (
	"testing"
	"time"
)

func TestObsidianParams(t *testing.T) {
	params := MainNetParams

	// 1. Verify Block Size (6MB)
	if params.BlockMaxSize != 6000000 {
		t.Errorf("BlockMaxSize is %d, want 6000000", params.BlockMaxSize)
	}

	// 2. Verify Max Supply (100M)
	if params.MaxMoney != 100000000 {
		t.Errorf("MaxMoney is %d, want 100000000", params.MaxMoney)
	}

	// 3. Verify Initial Supply (0 for fair launch)
	if params.InitialSupply != 0 {
		t.Errorf("InitialSupply is %d, want 0 (fair launch)", params.InitialSupply)
	}

	// 4. Verify Block Time (5 minutes)
	if params.TargetTimePerBlock != time.Minute*5 {
		t.Errorf("TargetTimePerBlock is %v, want 5m0s", params.TargetTimePerBlock)
	}

	// 5. Verify Network Name
	if params.Name != "mainnet" {
		t.Errorf("Name is %s, want mainnet", params.Name)
	}
}

func TestBlockReward(t *testing.T) {
	params := MainNetParams

	// Test initial reward (100 OBS)
	reward0 := params.CalcBlockSubsidy(0)
	if reward0 != 100 {
		t.Errorf("Block 0 reward = %d, want 100", reward0)
	}

	// Test before first halving
	reward100 := params.CalcBlockSubsidy(100)
	if reward100 != 100 {
		t.Errorf("Block 100 reward = %d, want 100", reward100)
	}

	// Test after first halving (420000 blocks) - should be 50 OBS
	reward420000 := params.CalcBlockSubsidy(420000)
	if reward420000 != 50 {
		t.Errorf("Block 420000 reward = %d, want 50", reward420000)
	}

	// Test after second halving (840000 blocks) - should be 25 OBS
	reward840000 := params.CalcBlockSubsidy(840000)
	if reward840000 != 25 {
		t.Errorf("Block 840000 reward = %d, want 25", reward840000)
	}

	// Test after third halving (1260000 blocks) - should be 12 OBS
	reward1260000 := params.CalcBlockSubsidy(1260000)
	if reward1260000 != 12 {
		t.Errorf("Block 1260000 reward = %d, want 12", reward1260000)
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

	// Test supply after 100 blocks (100 blocks * 100 OBS reward)
	supply100 := params.TotalSupplyAtHeight(100)
	expected := params.InitialSupply + (100 * 100)
	if supply100 != expected {
		t.Errorf("Supply at height 100 = %d, want %d", supply100, expected)
	}

	// Test that supply doesn't exceed max
	supplyMax := params.TotalSupplyAtHeight(10000000)
	if supplyMax > params.MaxMoney {
		t.Errorf("Supply %d exceeds max %d", supplyMax, params.MaxMoney)
	}
}
