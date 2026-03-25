# SDK Quickstart

Open Pay provides official SDKs for all major languages. Each SDK handles HMAC authentication, request signing, and webhook verification automatically.

## Installation

### TypeScript / JavaScript
```bash
npm install @openpay/sdk
```

### Go
```bash
go get github.com/openlankapay/openlankapay/sdks/sdk-go
```

### Python
```bash
pip install openpay-sdk
```

### PHP
```bash
composer require openpay/sdk
```

### Java (Maven)
```xml
<dependency>
    <groupId>com.openpay</groupId>
    <artifactId>openpay-sdk</artifactId>
    <version>0.1.0</version>
</dependency>
```

### CLI
```bash
go install github.com/openlankapay/openlankapay/cmd/openpay@latest
```

---

## Quick Start

### 1. Get Your API Key

1. Log into the [Merchant Portal](https://olp-merchant.nipuntheekshana.com)
2. Go to **Settings > API Keys**
3. Click **Create API Key**
4. Copy the compound key (`ak_live_xxx.sk_live_yyy`) — it's shown only once

### 2. Create a Checkout Session

#### TypeScript
```typescript
import { OpenPay } from '@openpay/sdk'

const openpay = new OpenPay('ak_live_xxx.sk_live_yyy', {
  baseURL: 'https://olp-api.nipuntheekshana.com',
})

const session = await openpay.checkout.createSession({
  amount: '2500.00',
  currency: 'LKR',
  successUrl: 'https://mysite.com/success',
  cancelUrl: 'https://mysite.com/cancel',
})

// Redirect customer to payment page
console.log('Checkout URL:', session.url)
console.log('QR Code:', session.qrContent)
```

#### Go
```go
package main

import (
    "context"
    "fmt"
    openpay "github.com/openlankapay/openlankapay/sdks/sdk-go"
)

func main() {
    client, _ := openpay.NewClient("ak_live_xxx.sk_live_yyy",
        openpay.WithBaseURL("https://olp-api.nipuntheekshana.com"))

    session, _ := client.Checkout.CreateSession(context.Background(),
        openpay.CheckoutSessionInput{
            Amount:     "2500.00",
            Currency:   "LKR",
            SuccessURL: "https://mysite.com/success",
        })

    fmt.Println("Checkout URL:", session.URL)
}
```

#### Python
```python
from openpay import OpenPay

client = OpenPay(
    "ak_live_xxx.sk_live_yyy",
    base_url="https://olp-api.nipuntheekshana.com",
)

session = client.checkout.create_session(
    amount="2500.00",
    currency="LKR",
    success_url="https://mysite.com/success",
)

print("Checkout URL:", session["url"])
```

#### PHP
```php
use OpenPay\OpenPayClient;

$client = new OpenPayClient('ak_live_xxx.sk_live_yyy', [
    'base_url' => 'https://olp-api.nipuntheekshana.com',
]);

$session = $client->payments->create([
    'amount' => '2500.00',
    'currency' => 'LKR',
]);
```

#### Java
```java
var client = new OpenPayClient(
    "ak_live_xxx.sk_live_yyy",
    "https://olp-api.nipuntheekshana.com"
);

var payment = client.payments().create(Map.of(
    "amount", "2500.00",
    "currency", "LKR"
));
```

#### CLI
```bash
openpay config set-key ak_live_xxx.sk_live_yyy
openpay config set-url https://olp-api.nipuntheekshana.com
openpay checkout create-session --amount 2500 --currency LKR
```

### 3. Check Payment Status

```typescript
const payment = await openpay.payments.get(session.paymentId)
console.log(payment.status) // INITIATED → PAID → CONFIRMED
```

### 4. Handle Webhooks

```typescript
import { verifyWebhookSignature } from '@openpay/sdk'

app.post('/webhook', (req, res) => {
  const event = verifyWebhookSignature(
    req.body,                              // raw body string
    req.headers,                           // contains x-webhook-signature, etc.
    publicKey,                             // from openpay.webhooks.getPublicKey()
  )

  switch (event.event) {
    case 'payment.paid':
      // Fulfill the order
      break
    case 'payment.expired':
      // Cancel the order
      break
  }

  res.sendStatus(200)
})
```

---

## SDK Feature Matrix

| Feature | TypeScript | Go | Python | PHP | Java | CLI |
|---------|-----------|-----|--------|-----|------|-----|
| Checkout Sessions | ✅ | ✅ | ✅ | — | — | ✅ |
| Create Payment | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| List Payments | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| Get Payment | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| Configure Webhook | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| Verify Webhook | ✅ | ✅ | ✅ | ✅ | — | — |
| HMAC Auth | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
