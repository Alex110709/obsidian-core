# Difficulty Adjustment Algorithm

## Bitcoin Compatibility

Obsidian uses **Bitcoin's exact difficulty adjustment algorithm**, with parameters adjusted for 5-minute block times instead of Bitcoin's 10-minute blocks.

## Algorithm Comparison

| Parameter | Bitcoin | Obsidian | Notes |
|-----------|---------|----------|-------|
| Block Time | 10 minutes | 5 minutes | Target time per block |
| Retarget Interval | 2016 blocks | 2016 blocks | **Same as Bitcoin** |
| Retarget Period | ~2 weeks | ~1 week | Different duration, same logic |
| Max Adjustment | 4x | 4x | **Same as Bitcoin** |
| Formula | `new = old × (actual/target)` | `new = old × (actual/target)` | **Identical** |

## How It Works

### 1. Retarget Timing
- Difficulty adjusts every **2,016 blocks** (exactly like Bitcoin)
- At 5-minute block time: 2,016 × 5min = **7 days**
- Bitcoin's 10-minute blocks: 2,016 × 10min = 14 days

### 2. Calculation Steps

```go
// Step 1: Get time span of last 2016 blocks
firstBlock := blockchain.GetBlock(height - 2015)
lastBlock := blockchain.GetBlock(height)
actualTimespan := lastBlock.Timestamp - firstBlock.Timestamp

// Step 2: Clamp to prevent extreme adjustments
targetTimespan := 7 days (in seconds)
if actualTimespan < targetTimespan / 4:
    actualTimespan = targetTimespan / 4  // Max increase: 4x
if actualTimespan > targetTimespan * 4:
    actualTimespan = targetTimespan * 4  // Max decrease: 4x

// Step 3: Calculate new difficulty
newTarget = oldTarget * actualTimespan / targetTimespan
newDifficulty = CompactForm(newTarget)

// Step 4: Ensure not below minimum
if newTarget > PowLimit:
    newTarget = PowLimit
```

### 3. Real-World Example

```
Scenario: Network hashrate doubles

Block 0-2015:    Difficulty = 0x1d00ffff
                 Mining time = 7 days (target)

Block 2016-4031: Hashrate 2x faster
                 Mining time = 3.5 days (actual)
                 
At block 4032:
  actualTimespan = 3.5 days
  targetTimespan = 7 days
  ratio = 3.5 / 7 = 0.5
  
  newDifficulty = oldDifficulty * 0.5
  Result: Difficulty DOUBLES (target halves)
  
Block 4032-6047: Now takes ~7 days again
```

## Code Implementation

### Main Function
File: `obsidian-core/blockchain/chain.go:254`

```go
func (b *BlockChain) CalcNextRequiredDifficulty(lastBlock *wire.MsgBlock, 
                                                  newBlockTime int64) (uint32, error)
```

This function:
1. Checks if we're at a retarget boundary (every 2016 blocks)
2. Calculates actual time taken for last 2016 blocks
3. Applies Bitcoin's adjustment formula with 4x clamp
4. Returns new difficulty bits

### Parameters
File: `obsidian-core/chaincfg/params.go:47`

```go
TargetTimespan:           time.Hour * 24 * 7,  // 1 week
TargetTimePerBlock:       time.Minute * 5,     // 5 minutes
RetargetAdjustmentFactor: 4,                   // Max 4x change
PowLimit:                 2^224 - 1,           // Minimum difficulty
```

## Testing

Run difficulty adjustment tests:
```bash
go test ./blockchain -v -run TestDifficulty
```

Expected output:
```
✓ No difficulty adjustment before retarget interval
✓ Fast mining increases difficulty correctly
✓ Slow mining decreases difficulty correctly
✓ Difficulty adjustment clamped to 4x maximum

Retarget interval: 2016 blocks
Target time per block: 5m0s
Target timespan: 168h0m0s (7.0 days)
```

## Differences from Bitcoin

The **only** differences are in timing parameters:

1. **Block Time**: 5 minutes vs Bitcoin's 10 minutes
2. **Retarget Period**: 1 week vs Bitcoin's 2 weeks
3. **PoW Algorithm**: DarkMatter (AES/SHA256) vs Bitcoin's SHA256d

**Everything else is identical:**
- 2016 block retarget interval ✓
- 4x maximum adjustment ✓
- Same mathematical formula ✓
- Same compact target encoding ✓
- Same clamping logic ✓

## Visual Timeline

```
Bitcoin:
[────2016 blocks────][────2016 blocks────]
 0        14 days    2016     28 days    4032
          ↑                   ↑
      Retarget            Retarget

Obsidian:
[────2016 blocks────][────2016 blocks────]
 0         7 days    2016     14 days    4032
           ↑                   ↑
       Retarget            Retarget
```

Same number of blocks, different calendar duration.

## References

- Bitcoin Wiki: https://en.bitcoin.it/wiki/Difficulty
- Bitcoin Core Source: `src/pow.cpp::CalculateNextWorkRequired()`
- Obsidian Implementation: `blockchain/chain.go::CalcNextRequiredDifficulty()`
