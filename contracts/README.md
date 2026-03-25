# Open Pay Smart Contracts

Solidity smart contracts for the Open Pay crypto-to-fiat payment platform. Handles payment escrow, multi-currency support with real-time price conversion via Chainlink oracles, and platform fee distribution.

## Architecture

```
┌──────────────────────────────────────────────────────────────┐
│                     Open Pay Backend                         │
│            (Go payment service orchestrator)                 │
└──────────────┬───────────────────────┬───────────────────────┘
               │ getQuote()            │ createPayment()
               │ (gas-free view)       │ confirmPayment()
               ▼                       ▼
┌──────────────────────────────────────────────────────────────┐
│                    OpenPayEscrow                             │
│  • Holds crypto until confirmed by platform                  │
│  • Supports ERC-20 (USDT, USDC, …) + native BNB             │
│  • On-chain price verification with slippage tolerance       │
│  • Platform fee deduction (configurable, max 10%)            │
│  • Payment expiration with auto-refund                       │
└──────────────┬───────────────────────────────────────────────┘
               │ getPrice() / getAmountInToken()
               ▼
┌──────────────────────────────────────────────────────────────┐
│                  OpenPayPriceFeed                             │
│  • Wraps Chainlink oracle aggregators                        │
│  • Stablecoin shortcut ($1.00 without oracle call)           │
│  • Staleness protection (rejects data > 1 hour old)          │
│  • Supports any ERC-20 + native currency                     │
└──────────────┬───────────────────────────────────────────────┘
               │
               ▼
┌──────────────────────────────────────────────────────────────┐
│              Chainlink / Mock Price Feeds                     │
│  • BNB/USD, USDT/USD, USDC/USD …                            │
│  • Testnet: MockPriceFeed with settable prices               │
│  • Mainnet: Real Chainlink AggregatorV3                      │
└──────────────────────────────────────────────────────────────┘
```

## Contracts

| Contract | Description |
|---|---|
| `OpenPayEscrow` | Core escrow — locks payments, distributes funds on confirmation, handles refunds and expiration |
| `OpenPayPriceFeed` | Oracle wrapper — maps tokens to Chainlink feeds, provides USD conversion with staleness checks |
| `MockUSDT` | Test USDT token (6 decimals, open mint) |
| `MockUSDC` | Test USDC token (6 decimals, open mint) |
| `MockPriceFeed` | Simulated Chainlink feed with settable price for testing |

## Payment Flow

### Volatile Token (BNB)
```
1. Merchant creates invoice: "Charge $100 USD"
2. Backend calls  escrow.getQuote(BNB, $100)
   → Returns: 0.1667 BNB, rate $600, min/max bounds
3. Frontend shows: "Pay 0.1667 BNB (~$100.00)"
4. Customer sends tx → Contract verifies amount ±2% of oracle price
5. Funds locked in escrow
6. Backend confirms → Merchant gets 98%, platform gets 2% fee
```

### Stablecoin (USDT/USDC)
```
1. Merchant creates invoice: "Charge $100 USD"
2. Amount = 100 USDT (1:1)
3. Pass usdAmount=0 to skip verification, or pass $100 for on-chain check
4. Same escrow → confirm → distribute flow
```

## BSC Testnet Deployment

| Contract | Address |
|---|---|
| **OpenPayEscrow** | `0xe50464081b781AFE101EB40bC7e68Fd017c5e8f2` |
| **OpenPayPriceFeed** | `0x1f34e070D4BB1eD3AaF37D8E3297b0a9A12a3399` |
| **MockUSDT** | `0x98e2146A4381C74708782D03dAd3913b0388954A` |
| **MockUSDC** | `0xE2b9aB57304C8AFc7068940c03EE202e0B8D4CEC` |
| **BNB/USD Feed** | `0x98e9C8Ff491C563AB87935C5F4C704Da2F39A80D` |
| **USDT/USD Feed** | `0x20f21e195d92c54993d6d64CBbFE4bF87a6293Ae` |

**Config:** 2% platform fee, 2% slippage tolerance, 1h staleness threshold

**Explorer:** [BscScan Testnet](https://testnet.bscscan.com)

## Development

### Prerequisites

- Node.js 18+
- npm

### Setup

```bash
cd contracts
npm install
```

### Compile

```bash
npm run compile
```

### Test (local Hardhat network)

```bash
npm test
```

Runs 26 tests covering:
- ERC-20 payments (create, confirm, refund, duplicate prevention, access control)
- Native BNB payments
- Payment expiration and auto-refund
- Admin functions (fee, slippage, token management)
- Price verification (quotes, slippage bounds, rejection)
- Oracle behavior (staleness, price updates, stablecoin shortcut)

### Deploy

```bash
# Local
npm run node                    # Terminal 1: start local node
npm run deploy:local            # Terminal 2: deploy

# BSC Testnet
export DEPLOYER_PRIVATE_KEY=your_key
npm run deploy:bsc-testnet

# Sepolia
export DEPLOYER_PRIVATE_KEY=your_key
export SEPOLIA_RPC_URL=https://your-rpc
npm run deploy:sepolia
```

### E2E Test (BSC Testnet)

```bash
npx hardhat run script/test-bsc-testnet.ts --network bscTestnet
```

## Supported Currencies

| Token | Type | Price Source | Decimals |
|---|---|---|---|
| BNB | Native | Chainlink oracle | 18 |
| USDT | ERC-20 | Chainlink oracle | 6 |
| USDC | ERC-20 | Stablecoin shortcut ($1.00) | 6 |

To add a new token:
1. `escrow.setSupportedToken(tokenAddress, true)` — enable in escrow
2. `priceFeed.configureToken(tokenAddress, chainlinkFeed, isStablecoin)` — configure pricing

## Security

- **ReentrancyGuard** on all state-changing payment functions
- **SafeERC20** for all token transfers
- **Ownable** access control on admin and confirmation functions
- **Slippage protection** rejects payments deviating >2% from oracle price
- **Staleness check** rejects oracle data older than 1 hour
- **Expiration** allows anyone to auto-refund expired payments
- **Platform fee cap** at 10% maximum (enforced in contract)
- **Slippage cap** at 5% maximum (enforced in contract)
