# Obsidian Token System

Obsidian Core includes a built-in token system that allows users to create and transfer custom tokens without requiring smart contracts. This provides ERC-20 style functionality at the protocol level.

## Overview

The token system enables:
- **Token Creation**: Issue new tokens with custom parameters
- **Token Transfer**: Send tokens between addresses
- **Balance Queries**: Check token holdings
- **Token Discovery**: Browse available tokens
- **Network Security**: Token transfers require OB fees

## Token Structure

Each token in Obsidian has the following properties:

```go
type Token struct {
    ID          Hash   // Unique token identifier
    Name        string // Token name (max 32 chars)
    Symbol      string // Token symbol (max 8 chars)
    Decimals    uint8  // Decimal places (0-18)
    TotalSupply int64  // Total supply
    Owner       string // Token creator address
    Created     int64  // Creation timestamp
}
```

## Token Transactions

### Token Issuance (TxTypeTokenIssue)

Creates a new token with initial supply:

```go
// Create token issuance transaction
tokenIssue := &wire.TokenIssue{
    Name:     "My Token",
    Symbol:   "MTK",
    Decimals: 18,
    Supply:   1000000,
    Owner:    "creator_address",
}

tx := wire.NewTokenIssueTx("creator_address", tokenIssue)
```

**Characteristics:**
- Free operation (no OB fee required)
- Creates initial supply to owner address
- Symbol must be unique on network
- Transaction includes token metadata in memo field

### Token Transfer (TxTypeTokenTransfer)

Transfers tokens between addresses:

```go
// Create token transfer transaction
tx := wire.NewTokenTransferTx(
    "sender_address",
    "recipient_address",
    tokenID,
    1000, // amount to transfer
)
```

**Characteristics:**
- Requires OB fee for network security
- Validates sender balance before transfer
- Updates token balances atomically
- Includes transfer details in memo field

## RPC API

### issuetoken

Creates a new token.

**Parameters:**
- `name` (string): Token name
- `symbol` (string): Token symbol
- `decimals` (number): Decimal places
- `supply` (number): Initial supply
- `owner` (string, optional): Owner address

**Example:**
```bash
curl -X POST http://localhost:8545 -d '{
  "jsonrpc": "2.0",
  "method": "issuetoken",
  "params": ["My Token", "MTK", 18, 1000000, "owner_address"],
  "id": 1
}'
```

**Response:**
```json
{
  "jsonrpc": "2.0",
  "result": {
    "txid": "a1b2c3d4...",
    "name": "My Token",
    "symbol": "MTK",
    "supply": 1000000,
    "owner": "owner_address"
  },
  "id": 1
}
```

### transfertoken

Transfers tokens between addresses.

**Parameters:**
- `token_symbol` (string): Token symbol
- `from_address` (string): Sender address
- `to_address` (string): Recipient address
- `amount` (number): Amount to transfer

**Example:**
```bash
curl -X POST http://localhost:8545 -d '{
  "jsonrpc": "2.0",
  "method": "transfertoken",
  "params": ["MTK", "sender_addr", "recipient_addr", 1000],
  "id": 1
}'
```

### gettokenbalance

Gets token balance for an address.

**Parameters:**
- `address` (string): Address to query
- `token_symbol` (string): Token symbol

**Example:**
```bash
curl -X POST http://localhost:8545 -d '{
  "jsonrpc": "2.0",
  "method": "gettokenbalance",
  "params": ["user_address", "MTK"],
  "id": 1
}'
```

**Response:**
```json
{
  "jsonrpc": "2.0",
  "result": {
    "address": "user_address",
    "token": "MTK",
    "balance": 50000
  },
  "id": 1
}
```

### gettokeninfo

Gets information about a token.

**Parameters:**
- `token_symbol` (string): Token symbol

**Example:**
```bash
curl -X POST http://localhost:8545 -d '{
  "jsonrpc": "2.0",
  "method": "gettokeninfo",
  "params": ["MTK"],
  "id": 1
}'
```

### listtokens

Lists all tokens on the network.

**Example:**
```bash
curl -X POST http://localhost:8545 -d '{
  "jsonrpc": "2.0",
  "method": "listtokens",
  "params": [],
  "id": 1
}'
```

### getaddresstokens

Gets all tokens held by an address.

**Parameters:**
- `address` (string): Address to query

**Example:**
```bash
curl -X POST http://localhost:8545 -d '{
  "jsonrpc": "2.0",
  "method": "getaddresstokens",
  "params": ["user_address"],
  "id": 1
}'
```

### shieldtoken

Shields or unshield tokens using shielded transactions for privacy.

**Parameters:**
- `from_address` (string): Sender address (transparent or shielded)
- `to_address` (string): Recipient address (transparent or shielded)
- `token_symbol` (string): Token symbol
- `amount` (number): Amount to shield/unshield

**Shielding (Transparent → Shielded):**
```bash
curl -X POST http://localhost:8545 -d '{
  "jsonrpc": "2.0",
  "method": "shieldtoken",
  "params": ["transparent_addr", "zobs_shielded_addr", "MTK", 1000],
  "id": 1
}'
```

**Unshielding (Shielded → Transparent):**
```bash
curl -X POST http://localhost:8545 -d '{
  "jsonrpc": "2.0",
  "method": "shieldtoken",
  "params": ["zobs_shielded_addr", "transparent_addr", "MTK", 500],
  "id": 1
}'
```

**Response:**
```json
{
  "jsonrpc": "2.0",
  "result": {
    "txid": "a1b2c3d4...",
    "action": "shielding",
    "token": "MTK",
    "from": "transparent_addr",
    "to": "zobs_shielded_addr",
    "amount": 1000
  },
  "id": 1
}
```

## Token Shielded Transactions

Obsidian supports private token transfers using shielded transactions, providing the same privacy benefits as OB shielded transactions but for custom tokens.

### Shielding Process

1. **Transparent → Shielded**: Move tokens from public addresses to private shielded pool
2. **Shielded → Shielded**: Private transfers within shielded pool
3. **Shielded → Transparent**: Move tokens back to public addresses

### Privacy Features

- **Amount Hiding**: Transaction amounts are encrypted
- **Sender/Receiver Anonymity**: Identities are protected
- **Memo Encryption**: Optional encrypted messages
- **zk-SNARK Proofs**: Mathematical privacy guarantees

### Use Cases

- **Private Token Transfers**: Hide transaction amounts and participants
- **Confidential DeFi**: Private token swaps and lending
- **Gaming Privacy**: Hide in-game token movements
- **Corporate Privacy**: Confidential token distributions

## Implementation Details

### Token Storage

Tokens and balances are stored using the blockchain's database:

- **Tokens**: `token:{tokenID}` → Token struct
- **Balances**: `balance:{address}:{tokenID}` → balance amount
- **Symbols**: In-memory index for fast symbol lookup

### Validation Rules

**Token Issuance:**
- Symbol must be unique (case-sensitive)
- Name ≤ 32 characters
- Symbol ≤ 8 characters
- Decimals ≤ 18
- Supply > 0

**Token Transfer:**
- Sender must have sufficient balance
- Amount > 0
- Token must exist
- OB fee required for network security

### Transaction Processing

1. **Validation**: Check token rules and balances
2. **Execution**: Update token balances
3. **Persistence**: Save to database
4. **Broadcast**: Propagate to network

## Use Cases

### DeFi Tokens
- Lending protocol tokens
- Liquidity provider tokens
- Governance tokens

### Gaming Tokens
- In-game currency
- NFT collections
- Achievement tokens

### Utility Tokens
- Access tokens
- Service credits
- Membership tokens

### Asset Tokens
- Real estate tokens
- Commodity tokens
- Security tokens

## Examples

### Creating a Gaming Token

```javascript
// Issue a gaming token
const result = await rpcCall('issuetoken', [
  'GameCoin',
  'GAME',
  18,
  1000000000,
  playerAddress
]);

console.log('Token created:', result.txid);
```

### Token Transfer in Game

```javascript
// Transfer tokens as reward
await rpcCall('transfertoken', [
  'GAME',
  gameContractAddress,
  playerAddress,
  rewardAmount
]);
```

### Checking Player Balance

```javascript
// Get player token balance
const balance = await rpcCall('gettokenbalance', [
  playerAddress,
  'GAME'
]);

console.log('Player has', balance.balance, 'GAME tokens');
```

## Security Considerations

### Network Fees
- Token transfers require OB fees to prevent spam
- Fee amount scales with transaction size

### Balance Validation
- All transfers validate sender balance
- Double-spend protection via UTXO model

### Token Uniqueness
- Symbol uniqueness prevents conflicts
- Token ID based on issuance transaction hash

### Access Control
- Only token owner can modify certain properties
- Transfer validation prevents unauthorized moves

## Future Extensions

### Planned Features
- **Token Burning**: Destroy tokens permanently
- **Token Minting**: Increase token supply
- **Token Freezing**: Lock token transfers
- **Multi-Token Transfers**: Batch operations
- **Token Metadata**: Extended properties

### Integration Possibilities
- **DEX Integration**: Decentralized exchanges
- **Wallet Support**: Multi-token wallets
- **Explorer Features**: Token tracking
- **Bridge Support**: Cross-chain tokens

## Comparison with Smart Contracts

| Feature | Obsidian Tokens | Smart Contract Tokens |
|---------|-----------------|----------------------|
| Setup Cost | Free | Gas fees |
| Transfer Speed | Fast | Variable |
| Security Model | Protocol-level | Contract-dependent |
| Upgradeability | Protocol upgrades | Contract upgrades |
| Complexity | Simple | Complex |
| Interoperability | Native | Contract calls |

## Contributing

To contribute to the token system:

1. Test token operations thoroughly
2. Ensure backward compatibility
3. Follow existing code patterns
4. Update documentation
5. Add comprehensive tests

## License

Token functionality is part of Obsidian Core, licensed under MIT.</content>
<parameter name="filePath">/Users/yuchan/Desktop/Obsidian Chain/obsidian-core/docs/token-guide.md