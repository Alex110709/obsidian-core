# Obsidian Emission Schedule

## Overview

Obsidian has a total supply cap of **100,000,000 OBS** with a fair distribution model.

## Supply Breakdown

- **Pre-mine**: 1,000,000 OBS (1% - initial supply)
- **Minable**: 99,000,000 OBS (99% - through block rewards)

## Block Reward Schedule

### Initial Reward
- **100 OBS** per block

### Halving Schedule

Block rewards halve every **420,000 blocks** (~4 years at 5-minute block time):

| Period | Block Range | Reward | Duration | Total Mined |
|--------|-------------|--------|----------|-------------|
| 1 | 0 - 419,999 | 100 OBS | ~4 years | 42,000,000 OBS |
| 2 | 420,000 - 839,999 | 50 OBS | ~4 years | 21,000,000 OBS |
| 3 | 840,000 - 1,259,999 | 25 OBS | ~4 years | 10,500,000 OBS |
| 4 | 1,260,000 - 1,679,999 | 12 OBS | ~4 years | 5,040,000 OBS |
| 5 | 1,680,000 - 2,099,999 | 6 OBS | ~4 years | 2,520,000 OBS |
| 6 | 2,100,000 - 2,519,999 | 3 OBS | ~4 years | 1,260,000 OBS |
| 7 | 2,520,000 - 2,939,999 | 1 OBS | ~4 years | 420,000 OBS |
| 8+ | 2,940,000+ | 0 OBS | Forever | 0 OBS |

**Total Minable**: ~82,740,000 OBS  
**Remaining**: Can be distributed via tail emission if needed

## Timeline

```
Year 0-4:    100 OBS/block  →  42M OBS mined   (42% of max)
Year 4-8:     50 OBS/block  →  21M OBS mined   (21% of max)
Year 8-12:    25 OBS/block  →  10.5M OBS mined (10.5% of max)
Year 12-16:   12 OBS/block  →  5M OBS mined    (5% of max)
Year 16-20:    6 OBS/block  →  2.5M OBS mined  (2.5% of max)
Year 20-24:    3 OBS/block  →  1.3M OBS mined  (1.3% of max)
Year 24-28:    1 OBS/block  →  0.4M OBS mined  (0.4% of max)
Year 28+:      0 OBS/block  →  No new emission
```

## Comparison with Bitcoin

| Feature | Bitcoin | Obsidian |
|---------|---------|----------|
| Total Supply | 21,000,000 BTC | 100,000,000 OBS |
| Initial Reward | 50 BTC | 100 OBS |
| Block Time | 10 minutes | 5 minutes |
| Halving Interval | 210,000 blocks (~4 years) | 420,000 blocks (~4 years) |
| Pre-mine | 0% | 1% (1M OBS) |

**Note**: Obsidian has 2x faster blocks but same halving period in real time as Bitcoin.

## Supply Curve

```
100M ┤                                        ___________________
     │                                   ____/
 80M ┤                              ____/
     │                         ____/
 60M ┤                    ____/
     │               ____/
 40M ┤          ____/
     │     ____/
 20M ┤____/
     │
  0M └─────┬─────┬─────┬─────┬─────┬─────┬─────┬─────>
         4yr   8yr   12yr  16yr  20yr  24yr  28yr   Time
```

## Inflation Rate

Year-over-year inflation decreases due to halvings:

| Year | Annual Inflation | Circulating Supply |
|------|------------------|--------------------|
| 1 | ~26% | 26M OBS |
| 4 | ~13% | 43M OBS |
| 8 | ~6% | 64M OBS |
| 12 | ~3% | 75M OBS |
| 16 | ~1.5% | 80M OBS |
| 20 | ~0.7% | 83M OBS |
| 24 | ~0.3% | 84M OBS |
| 28+ | 0% | ~83.7M OBS (capped) |

## Fair Launch Principles

1. **Transparent Pre-mine**: 1M OBS (1%) disclosed upfront
2. **No ICO**: All remaining supply mined fairly
3. **Public Mining**: Anyone can mine from day 1
4. **Predictable Schedule**: Bitcoin-style halvings
5. **Long-term Security**: 28+ years of mining rewards

## Future Considerations

### Tail Emission (Optional)

After year 28, a small tail emission could be added to ensure long-term miner incentives:

- **Option 1**: Fixed 1 OBS/block forever (inflation: ~0.1% annually)
- **Option 2**: 0.5 OBS/block forever (inflation: ~0.05% annually)
- **Option 3**: Transaction fees only (no emission)

This would be decided by community governance before the final halving.

## Economic Model

### Block Rewards vs Transaction Fees

As block rewards decrease, transaction fees become increasingly important:

```
Early Years (0-8):  99% rewards, 1% fees
Mid Years (8-16):   95% rewards, 5% fees
Late Years (16-24): 80% rewards, 20% fees
Final Years (24+):  0% rewards, 100% fees
```

### Miner Revenue

With 5-minute blocks:
- **Year 1**: 100 OBS/block × 105,120 blocks/year = 10,512,000 OBS/year
- **Year 5**: 50 OBS/block × 105,120 blocks/year = 5,256,000 OBS/year
- **Year 10**: 25 OBS/block × 105,120 blocks/year = 2,628,000 OBS/year

## Verification

You can verify the emission schedule:

```bash
# Get current block reward
curl -X POST http://localhost:8545 -d '{
  "jsonrpc":"2.0",
  "method":"getblocktemplate",
  "params":[],
  "id":1
}'

# Or check in code
go test ./chaincfg -v -run TestBlockReward
```

## Summary

- ✅ **100M total supply** (like many modern cryptocurrencies)
- ✅ **100 OBS initial reward** (appropriate for supply cap)
- ✅ **99M minable** (99% fair distribution)
- ✅ **1M pre-mine** (1% for development/ecosystem)
- ✅ **Bitcoin-style halvings** (proven economic model)
- ✅ **Long-term sustainability** (28+ years of rewards)

This emission schedule balances fair distribution, miner incentives, and long-term sustainability.
