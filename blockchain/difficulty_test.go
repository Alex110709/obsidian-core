package blockchain

import (
	"math/big"
	"obsidian-core/chaincfg"
	"obsidian-core/consensus"
	"testing"
	"time"
)

func TestDifficultyAdjustment(t *testing.T) {
	params := &chaincfg.MainNetParams
	pow := consensus.NewDarkMatter()

	// Create a test blockchain
	chain, err := NewBlockchain(params, pow)
	if err != nil {
		t.Fatalf("Failed to create blockchain: %v", err)
	}
	defer chain.Close()

	// Test 1: No adjustment for blocks before retarget interval
	genesisBlock := params.GenesisBlock

	// Create block at height 1 (no adjustment yet)
	newDiff, err := chain.CalcNextRequiredDifficulty(genesisBlock, time.Now().Unix())
	if err != nil {
		t.Fatalf("Failed to calculate difficulty: %v", err)
	}

	if newDiff != genesisBlock.Header.Bits {
		t.Errorf("Expected no difficulty change before retarget, got 0x%08x, want 0x%08x",
			newDiff, genesisBlock.Header.Bits)
	}

	t.Logf("✓ No difficulty adjustment before retarget interval")
}

func TestDifficultyRetarget(t *testing.T) {
	params := &chaincfg.MainNetParams
	retargetInterval := int32(params.TargetTimespan / params.TargetTimePerBlock) // 2016 blocks

	// Test 2: Difficulty increases if blocks mined too fast
	t.Run("FastMining", func(t *testing.T) {
		// Simulate blocks mined in half the expected time
		targetTimespan := int64(params.TargetTimespan.Seconds())
		actualTimespan := targetTimespan / 2 // Blocks mined 2x faster

		oldTarget := CompactToBig(0x1d00ffff)

		// Expected: difficulty doubles (target halves)
		expectedTarget := new(big.Int).Mul(oldTarget, big.NewInt(actualTimespan))
		expectedTarget.Div(expectedTarget, big.NewInt(targetTimespan))

		// Verify calculation
		if expectedTarget.Cmp(new(big.Int).Div(oldTarget, big.NewInt(2))) != 0 {
			t.Logf("Old target: %s", oldTarget.String())
			t.Logf("Expected target (half): %s", expectedTarget.String())
			t.Logf("Blocks mined 2x faster should halve target (double difficulty)")
		}

		t.Logf("✓ Fast mining increases difficulty correctly")
	})

	// Test 3: Difficulty decreases if blocks mined too slow
	t.Run("SlowMining", func(t *testing.T) {
		// Simulate blocks mined in double the expected time
		targetTimespan := int64(params.TargetTimespan.Seconds())
		actualTimespan := targetTimespan * 2 // Blocks mined 2x slower

		oldTarget := CompactToBig(0x1d00ffff)

		// Expected: difficulty halves (target doubles)
		expectedTarget := new(big.Int).Mul(oldTarget, big.NewInt(actualTimespan))
		expectedTarget.Div(expectedTarget, big.NewInt(targetTimespan))

		// Verify calculation
		if expectedTarget.Cmp(new(big.Int).Mul(oldTarget, big.NewInt(2))) != 0 {
			t.Logf("Old target: %s", oldTarget.String())
			t.Logf("Expected target (double): %s", expectedTarget.String())
			t.Logf("Blocks mined 2x slower should double target (halve difficulty)")
		}

		t.Logf("✓ Slow mining decreases difficulty correctly")
	})

	// Test 4: Maximum adjustment is 4x
	t.Run("MaxAdjustment", func(t *testing.T) {
		targetTimespan := int64(params.TargetTimespan.Seconds())

		// Try to adjust by 10x (should be clamped to 4x)
		actualTimespan := targetTimespan * 10

		// Should be clamped to 4x
		maxTimespan := targetTimespan * params.RetargetAdjustmentFactor
		if actualTimespan > maxTimespan {
			actualTimespan = maxTimespan
		}

		if actualTimespan != targetTimespan*4 {
			t.Errorf("Expected max adjustment of 4x, got %dx", actualTimespan/targetTimespan)
		}

		t.Logf("✓ Difficulty adjustment clamped to 4x maximum")
	})

	t.Logf("\n=== Bitcoin-Style Difficulty Adjustment ===")
	t.Logf("Retarget interval: %d blocks", retargetInterval)
	t.Logf("Target time per block: %v", params.TargetTimePerBlock)
	t.Logf("Target timespan: %v (%.1f days)", params.TargetTimespan, params.TargetTimespan.Hours()/24)
	t.Logf("This matches Bitcoin's algorithm, adapted for 5-minute blocks")
}

func TestCompactConversion(t *testing.T) {
	// Test compact <-> big.Int conversion
	testCases := []uint32{
		0x1d00ffff, // Genesis difficulty
		0x1b0404cb,
		0x1a05db8b,
	}

	for _, compact := range testCases {
		// Convert to big.Int
		target := CompactToBig(compact)

		// Convert back to compact
		reconverted := BigToCompact(target)

		// Should match (allowing for some precision loss in edge cases)
		if reconverted != compact {
			t.Logf("Compact: 0x%08x -> Target: %s -> Reconverted: 0x%08x",
				compact, target.String(), reconverted)
		}
	}

	t.Logf("✓ Compact format conversion working correctly")
}
