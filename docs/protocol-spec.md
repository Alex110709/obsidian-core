# Obsidian Protocol Specification

This document specifies the Obsidian cryptocurrency protocol, including wire format, message types, consensus rules, and network behavior. All implementations must conform to this specification for compatibility.

## Overview

Obsidian is a privacy-focused cryptocurrency with the following key features:

- **Proof of Work**: DarkMatter algorithm (SHA-256 + AES hybrid)
- **Shielded Transactions**: Zero-knowledge proofs for privacy
- **Encrypted Memos**: Optional encrypted messages attached to transactions
- **Tor Integration**: Optional anonymous networking
- **1-minute block time**: Optimized for faster confirmations

## Network Parameters

### Mainnet Parameters

| Parameter | Value | Description |
|-----------|-------|-------------|
| `Name` | "mainnet" | Network identifier |
| `BaseBlockReward` | 50 OBS | Initial block reward |
| `HalvingInterval` | 420,000 | Blocks between reward halvings |
| `MaxMoney` | 100,000,000 | Maximum supply |
| `BlockMaxSize` | 3,200,000 | Maximum block size in bytes |
| `TargetTimePerBlock` | 1 minute | Target block time |
| `DifficultyAdjustmentInterval` | 10,080 | Blocks between difficulty adjustments |

### Address Formats

#### Transparent Addresses (t-addresses)
- **Prefix**: `Obs`
- **Format**: Base58-encoded public key hash
- **Example**: `Obs1abc123def456...`

#### Shielded Addresses (z-addresses)
- **Prefix**: `zobs`
- **Format**: Bech32-encoded shielded address
- **Example**: `zobs1abc123def456...`

## Wire Protocol

### Message Structure

All network messages follow this format:

```
Message = Magic + Command + Length + Checksum + Payload
```

- **Magic**: 4 bytes (0x4F 0x42 0x53 0x00 for mainnet)
- **Command**: 12 bytes, null-padded command string
- **Length**: 4 bytes, little-endian payload length
- **Checksum**: 4 bytes, first 4 bytes of SHA-256(SHA-256(payload))
- **Payload**: Variable length message data

### Message Types

#### Version (version)

Handshake message sent when connecting to a peer.

```
Payload:
- Version: uint32 (protocol version)
- Services: uint64 (supported services bitfield)
- Timestamp: int64 (current timestamp)
- AddrRecv: NetworkAddress (receiver address)
- AddrFrom: NetworkAddress (sender address)
- Nonce: uint64 (random nonce)
- UserAgent: VarStr (client user agent)
- StartHeight: int32 (starting block height)
- Relay: bool (whether to relay transactions)
```

#### VerAck (verack)

Acknowledgment of version message.

```
Payload: (empty)
```

#### Ping (ping)

Keep-alive message.

```
Payload:
- Nonce: uint64
```

#### Pong (pong)

Response to ping.

```
Payload:
- Nonce: uint64
```

#### GetAddr (getaddr)

Request for known peer addresses.

```
Payload: (empty)
```

#### Addr (addr)

List of known peer addresses.

```
Payload:
- Count: VarInt
- Addresses: NetworkAddress[Count]
```

#### Inv (inv)

Inventory announcement.

```
Payload:
- Count: VarInt
- Inventory: InvVector[Count]
```

#### GetData (getdata)

Request for inventory data.

```
Payload:
- Count: VarInt
- Inventory: InvVector[Count]
```

#### Block (block)

Block data.

```
Payload:
- Block: Block
```

#### Tx (tx)

Transaction data.

```
Payload:
- Transaction: Transaction
```

#### GetBlocks (getblocks)

Request block headers.

```
Payload:
- Version: uint32
- BlockLocatorHashes: Hash[]
- HashStop: Hash
```

#### GetHeaders (getheaders)

Request block headers only.

```
Payload:
- Version: uint32
- BlockLocatorHashes: Hash[]
- HashStop: Hash
```

#### Headers (headers)

Block headers response.

```
Payload:
- Count: VarInt
- Headers: BlockHeader[Count]
```

## Data Structures

### NetworkAddress

```
- Time: uint32 (timestamp)
- Services: uint64 (services bitfield)
- IP: byte[16] (IPv6 address)
- Port: uint16 (network byte order)
```

### InvVector

```
- Type: uint32 (object type)
- Hash: Hash (object hash)
```

Object types:
- 1: Transaction
- 2: Block
- 3: Filtered Block

### Hash

32-byte hash value (little-endian).

### VarInt

Variable-length integer encoding:
- 0-252: 1 byte
- 253-65535: 3 bytes (0xfd + uint16)
- 65536-4294967295: 5 bytes (0xfe + uint32)
- >4294967295: 9 bytes (0xff + uint64)

### VarStr

Variable-length string:
- Length: VarInt
- String: byte[Length]

## Block Structure

### Block Header

```
- Version: int32
- PrevBlock: Hash
- MerkleRoot: Hash
- ShieldedRoot: Hash (root of shielded transaction tree)
- Timestamp: uint32
- Bits: uint32 (difficulty target)
- Nonce: uint32
- TransactionCount: VarInt
```

### Block

```
- Header: BlockHeader
- Transactions: Transaction[TransactionCount]
```

## Transaction Structure

### Transaction Types

1. **Transparent Transaction**: Standard public transaction
2. **Shielded Transaction**: Private transaction with zero-knowledge proof
3. **Shielding Transaction**: t-address to z-address
4. **Deshielding Transaction**: z-address to t-address

### Transparent Transaction

```
- Version: int32
- InputCount: VarInt
- Inputs: TxInput[InputCount]
- OutputCount: VarInt
- Outputs: TxOutput[OutputCount]
- LockTime: uint32
```

### Shielded Transaction

```
- Version: int32
- ShieldedSpends: VarInt
- ShieldedSpends: ShieldedSpend[ShieldedSpends]
- ShieldedOutputs: VarInt
- ShieldedOutputs: ShieldedOutput[ShieldedOutputs]
- TransparentInputs: VarInt
- TransparentInputs: TxInput[TransparentInputs]
- TransparentOutputs: VarInt
- TransparentOutputs: TxOutput[TransparentOutputs]
- BindingSig: byte[64] (binding signature)
- LockTime: uint32
```

### TxInput

```
- PrevTxHash: Hash
- PrevTxIndex: uint32
- ScriptSig: VarStr (unlocking script)
- Sequence: uint32
```

### TxOutput

```
- Value: int64 (amount in satoshis)
- ScriptPubKey: VarStr (locking script)
```

### ShieldedSpend

```
- Cv: byte[32] (value commitment)
- Anchor: byte[32] (merkle root)
- Nullifier: byte[32]
- Rk: byte[32] (randomized key)
- Proof: byte[192] (zero-knowledge proof)
- SpendAuthSig: byte[64] (spend authorization signature)
```

### ShieldedOutput

```
- Cv: byte[32] (value commitment)
- Cmu: byte[32] (commitment to value)
- EphemeralKey: byte[32]
- EncCiphertext: byte[580] (encrypted note)
- OutCiphertext: byte[80] (encrypted memo)
- Proof: byte[192] (zero-knowledge proof)
```

## Consensus Rules

### Proof of Work

Obsidian uses the DarkMatter PoW algorithm:

1. Take SHA-256 hash of block header
2. Use hash as AES key to encrypt a fixed plaintext
3. Take SHA-256 hash of ciphertext
4. Compare result against difficulty target

### Difficulty Adjustment

Uses Bitcoin's difficulty adjustment algorithm adapted for 5-minute blocks:

```
new_target = old_target × (actual_time / target_time)
```

Where:
- `target_time = 2016 × 5 minutes = 7 days`
- Adjustment clamped between 1/4 and 4x

### Block Validation

Blocks must satisfy:

1. **Header Validation**:
   - Valid PoW
   - Timestamp within reasonable range
   - Difficulty target correct
   - Previous block exists

2. **Transaction Validation**:
   - All inputs unspent
   - Signature verification
   - No double-spends
   - Fee calculation correct

3. **Shielded Transaction Validation**:
   - Zero-knowledge proofs valid
   - Nullifiers not previously used
   - Value commitments balance

### Transaction Validation

#### Transparent Transactions

1. Verify input scripts
2. Check amounts don't exceed inputs
3. Verify signatures
4. Check for double-spends

#### Shielded Transactions

1. Verify spend proofs
2. Verify output proofs
3. Check nullifiers are unique
4. Verify binding signature
5. Check value balance

## Network Behavior

### Peer Discovery

1. Connect to seed nodes
2. Request peer addresses with `getaddr`
3. Connect to discovered peers
4. Exchange version messages

### Block Propagation

1. Mine or receive new block
2. Validate block
3. Send `inv` message to all peers
4. Respond to `getdata` requests with `block` message

### Transaction Propagation

1. Receive or create transaction
2. Validate transaction
3. Send `inv` message to all peers
4. Respond to `getdata` requests with `tx` message

### Connection Management

- Maintain minimum 8 outbound connections
- Accept up to 125 inbound connections
- Disconnect peers sending invalid data
- Ban misbehaving peers

## RPC API

### JSON-RPC 2.0 Interface

All RPC calls use JSON-RPC 2.0 format:

```json
{
  "jsonrpc": "2.0",
  "method": "method_name",
  "params": [...],
  "id": 123
}
```

### Standard Methods

#### Blockchain Methods

- `getblockcount` - Get current block height
- `getbestblockhash` - Get best block hash
- `getblock` - Get block by hash
- `getblockhash` - Get block hash by height
- `getblockchaininfo` - Get blockchain information

#### Transaction Methods

- `getrawtransaction` - Get raw transaction
- `decoderawtransaction` - Decode raw transaction
- `sendrawtransaction` - Broadcast transaction
- `gettxout` - Get unspent transaction output

#### Mining Methods

- `getmininginfo` - Get mining information
- `getwork` - Get work for mining
- `submitblock` - Submit mined block

#### Wallet Methods

- `getnewaddress` - Generate new address
- `getbalance` - Get wallet balance
- `sendtoaddress` - Send to address
- `listtransactions` - List wallet transactions

### Shielded Methods

- `z_getnewaddress` - Generate shielded address
- `z_getbalance` - Get shielded balance
- `z_sendmany` - Send shielded transaction
- `z_listaddresses` - List shielded addresses
- `z_exportviewingkey` - Export viewing key
- `z_importviewingkey` - Import viewing key

## Tor Integration

### .onion Addresses

Obsidian supports Tor hidden services:

- **Format**: `hostname.onion:port`
- **Resolution**: Automatic through Tor SOCKS proxy
- **Authentication**: None required

### Tor Configuration

- **SOCKS Proxy**: `127.0.0.1:9050`
- **Control Port**: `127.0.0.1:9051`
- **Hidden Service**: Auto-generated on startup

## Security Considerations

### Privacy Features

1. **Shielded Addresses**: Hide transaction amounts and participants
2. **Encrypted Memos**: Optional encrypted messages
3. **Tor Networking**: Anonymous peer connections
4. **No Address Reuse**: Each transaction uses new addresses

### Network Security

1. **Proof of Work**: Prevents double-spending
2. **Checkpoints**: Hard-coded block checkpoints
3. **Banning**: Misbehaving peer banning
4. **Validation**: Comprehensive transaction/block validation

### Implementation Security

1. **Input Validation**: All network inputs validated
2. **Memory Safety**: No buffer overflows
3. **Cryptographic Security**: Secure random number generation
4. **Key Management**: Secure private key handling

## Implementation Notes

### Endianness

All integer values are little-endian unless specified otherwise.

### Hash Function

All hashes use double SHA-256:

```
hash = SHA-256(SHA-256(data))
```

### Merkle Trees

Transaction merkle trees use standard Bitcoin format with duplicate hashing for odd numbers of transactions.

### Time Values

All timestamps are UNIX timestamps (seconds since 1970-01-01 00:00:00 UTC).

### Version Numbers

- **Protocol Version**: 1 (current)
- **Block Version**: 1 (current)
- **Transaction Version**: 1 (transparent), 2 (shielded)

## Future Extensions

### Planned Features

1. **Sapling Upgrade**: Enhanced shielded transactions
2. **Orchard Upgrade**: Further privacy improvements
3. **Sidechains**: Cross-chain interoperability
4. **Smart Contracts**: Turing-complete scripting

### Extension Points

1. **New Transaction Types**: Extensible transaction format
2. **Consensus Rule Changes**: Hard fork mechanism
3. **New Network Messages**: Extensible message system
4. **Plugin Architecture**: Runtime extensibility

## References

- [Bitcoin Protocol Documentation](https://bitcoin.org/en/developer-reference)
- [Zcash Protocol Specification](https://zips.z.cash/protocol/protocol.pdf)
- [DarkMatter PoW Paper](https://example.com/darkmatter-paper)
- [Obsidian Whitepaper](https://obsidian.network/whitepaper)

## License

This specification is licensed under CC-BY-SA 4.0.</content>
<parameter name="filePath">/Users/yuchan/Desktop/Obsidian Chain/obsidian-core/docs/protocol-spec.md