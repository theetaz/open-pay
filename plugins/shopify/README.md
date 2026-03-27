# Open Pay Shopify Plugin

Accept crypto payments (USDT, BTC, ETH, USDC, BNB) in your Shopify store via Open Pay.

## Setup

```bash
cd plugins/shopify
npm install
```

## Configuration

Set environment variables:

```bash
OPENPAY_API_URL=https://olp-api.nipuntheekshana.com
OPENPAY_API_KEY=ak_live_xxx
OPENPAY_API_SECRET=your_secret
OPENPAY_PROVIDER=BYBIT          # BYBIT, BINANCE, KUCOIN, or TEST
APP_URL=https://your-app.com    # Public URL for webhooks
PORT=3100
```

## Run

```bash
npm start
# or for development:
npm run dev
```

## Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/payment/create` | Create payment and get checkout URL |
| `GET` | `/payment/status/:id` | Check payment status |
| `POST` | `/webhook/payment` | Receive payment webhooks |
| `GET` | `/health` | Health check |
| `GET` | `/config` | View current configuration |

## Shopify Integration

1. In Shopify Admin, go to Settings → Payments → Alternative Payments
2. Add a custom payment method pointing to your deployed plugin URL
3. Configure the redirect URL to `POST /payment/create`
4. Set the return URL to your Shopify order confirmation page

## Payment Flow

```
Customer → Shopify Checkout → POST /payment/create → Open Pay Checkout
    ↓
Customer pays crypto → Open Pay webhook → POST /webhook/payment
    ↓
Plugin updates Shopify order status via Admin API
```
