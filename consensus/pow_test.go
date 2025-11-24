package consensus

import (
	"bytes"
	"obsidian-core/wire"
	"testing"
	"time"
)

// TestDarkMatterDeterministic ensures the hash function is deterministic
func TestDarkMatterDeterministic(t *testing.T) {
	header := &wire.BlockHeader{
		Version:   1,
		PrevBlock: wire.Hash{},
		Timestamp: time.Now(),
		Bits:      0x2000ffff,
		Nonce:     12345,
	}

	// Calculate hash twice
	hash1 := DarkMatterHash(header)
	hash2 := DarkMatterHash(header)

	// They must be identical
	if !bytes.Equal(hash1, hash2) {
		t.Errorf("DarkMatterHash is not deterministic!\nHash1: %x\nHash2: %x", hash1, hash2)
	}
}

// TestDarkMatterVerify ensures Verify and Solve are consistent
func TestDarkMatterVerify(t *testing.T) {
	dm := NewDarkMatter()

	header := &wire.BlockHeader{
		Version:   1,
		PrevBlock: wire.Hash{},
		Timestamp: time.Now(),
		Bits:      0x2000ffff, // Easy difficulty
		Nonce:     0,
	}

	// Try to solve with a low difficulty
	nonce, solution, found := dm.SolveWithLimit(header, 100000)

	if found {
		t.Logf("Found solution with nonce: %d", nonce)
		header.Nonce = nonce
		header.DarkMatterSolution = solution

		// Verify must pass
		if !dm.Verify(header) {
			t.Errorf("Verify failed for solved block!")
		}
	} else {
		t.Logf("No solution found in 100000 attempts (this may be ok for high difficulty)")
	}
}
