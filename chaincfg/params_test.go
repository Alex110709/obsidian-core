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

	// 4. Verify Block Time (20 seconds)
	if params.TargetTimePerBlock != time.Second*20 {
		t.Errorf("TargetTimePerBlock is %v, want 20s", params.TargetTimePerBlock)
	}

	// 5. Verify Gas Limits
	if params.BlockGasLimit != 30000000 {
		t.Errorf("BlockGasLimit is %d, want 30000000", params.BlockGasLimit)
	}

	// 5. Verify Network Name
	if params.Name != "mainnet" {
		t.Errorf("Name is %s, want mainnet", params.Name)
	}
}

func TestBlockReward(t *testing.T) {
	params := MainNetParams
	// Reset burn for clean test
	params.TotalBurned = 0

	// Test initial reward (25 OBS = 2,500,000,000 satoshis) - adjusted for 20 second blocks + yearly halving
	reward0 := params.CalcBlockSubsidy(0)
	if reward0 != 2500000000 {
		t.Errorf("Block 0 reward = %d, want 2500000000", reward0)
	}

	// Test before first halving
	reward100 := params.CalcBlockSubsidy(100)
	if reward100 != 2500000000 {
		t.Errorf("Block 100 reward = %d, want 2500000000", reward100)
	}

	// Test after first halving (1577000 blocks = ~1 year) - should be 12 OBS = 1,200,000,000 satoshis (bit shift: 25 >> 1 = 12)
	reward1577000 := params.CalcBlockSubsidy(1577000)
	if reward1577000 != 1200000000 {
		t.Errorf("Block 1577000 reward = %d, want 1200000000", reward1577000)
	}

	// Test after second halving (3154000 blocks = ~2 years) - should be 6 OBS = 600,000,000 satoshis (bit shift: 12 >> 1 = 6)
	reward3154000 := params.CalcBlockSubsidy(3154000)
	if reward3154000 != 600000000 {
		t.Errorf("Block 3154000 reward = %d, want 600000000", reward3154000)
	}

	// Test after third halving (4731000 blocks = ~3 years) - should be 3 OBS = 300,000,000 satoshis (bit shift: 6 >> 1 = 3)
	reward4731000 := params.CalcBlockSubsidy(4731000)
	if reward4731000 != 300000000 {
		t.Errorf("Block 4731000 reward = %d, want 300000000", reward4731000)
	}

	// Test minimum reward (after many halvings)
	rewardHigh := params.CalcBlockSubsidy(10000000)
	minReward := params.MinimumBlockReward * 100000000
	if rewardHigh < minReward {
		t.Errorf("Reward should not be less than minimum %d, got %d", minReward, rewardHigh)
	}

	t.Logf("Block rewards: 0=%f, 1.5M=%f, 3M=%f, 4.7M=%f OBS",
		float64(reward0)/100000000,
		float64(reward1577000)/100000000,
		float64(reward3154000)/100000000,
		float64(reward4731000)/100000000)
}

func TestBurnRedistribution(t *testing.T) {
	params := MainNetParams

	// Test with no burn
	redistribution0 := params.CalcBurnRedistribution()
	if redistribution0 != 0 {
		t.Errorf("CalcBurnRedistribution() with no burn = %d, want 0", redistribution0)
	}

	// Test with 1M OBS burned (100,000,000,000,000 satoshis)
	params.AddBurn(100000000000000)

	// With BurnRate of 10 (0.1%), should redistribute:
	// (100,000,000,000,000 * 10) / 10000 = 100,000,000,000 satoshis = 1,000 OBS per block
	expectedRedistribution := int64(100000000000)
	redistribution := params.CalcBurnRedistribution()

	if redistribution != expectedRedistribution {
		t.Errorf("CalcBurnRedistribution() = %d, want %d", redistribution, expectedRedistribution)
	}

	t.Logf("With 1M OBS burned, redistributing %f OBS per block", float64(redistribution)/100000000)

	// Test that block reward includes burn redistribution
	baseReward := int64(25 * 100000000) // 25 OBS base reward
	totalReward := params.CalcBlockSubsidy(0)
	expectedTotal := baseReward + redistribution

	if totalReward != expectedTotal {
		t.Errorf("CalcBlockSubsidy() with burn = %d, want %d (base %d + redistribution %d)",
			totalReward, expectedTotal, baseReward, redistribution)
	}

	t.Logf("Total block reward with burn redistribution: %f OBS", float64(totalReward)/100000000)
}

func TestCirculatingSupply(t *testing.T) {
	// Create a fresh params for this test
	testParams := MainNetParams
	testParams.TotalBurned = 0

	// Test with no burn
	supply := testParams.GetCirculatingSupply(100)
	minted := testParams.TotalSupplyAtHeight(100)

	if supply != minted {
		t.Errorf("GetCirculatingSupply() with no burn = %d, want %d", supply, minted)
	}

	// Test with burn
	burnAmount := int64(100000000000) // 1,000 OBS
	testParams.AddBurn(burnAmount)

	// After burning, the supply calculation includes burn redistribution
	// So total minted will be higher due to redistribution
	mintedWithRedistribution := testParams.TotalSupplyAtHeight(100)
	circulatingSupply := testParams.GetCirculatingSupply(100)
	expectedCirculating := mintedWithRedistribution - burnAmount

	if circulatingSupply != expectedCirculating {
		t.Errorf("GetCirculatingSupply() = %d, want %d (minted %d - burned %d)",
			circulatingSupply, expectedCirculating, mintedWithRedistribution, burnAmount)
	}

	t.Logf("Circulating supply: %f OBS (minted: %f, burned: %f)",
		float64(circulatingSupply)/100000000,
		float64(mintedWithRedistribution)/100000000,
		float64(burnAmount)/100000000)
}

func TestTotalSupply(t *testing.T) {
	params := MainNetParams

	// Test initial supply
	supply0 := params.TotalSupplyAtHeight(0)
	if supply0 != params.InitialSupply {
		t.Errorf("Supply at height 0 = %d, want %d", supply0, params.InitialSupply)
	}

	// Test supply after 100 blocks (100 blocks * 25 OBS reward = 2,500 OBS = 250,000,000,000 satoshis)
	// Note: This doesn't include burn redistribution in basic test
	// Reset burn for this test
	params.TotalBurned = 0
	supply100 := params.TotalSupplyAtHeight(100)
	expected := params.InitialSupply + (100 * 2500000000)
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
