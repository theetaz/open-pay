# openpay_flutter

Official Open Pay SDK for Flutter. Accept crypto payments with automatic LKR conversion in your mobile apps.

## Installation

Add to your `pubspec.yaml`:

```yaml
dependencies:
  openpay_flutter: ^0.1.0
```

Then run:

```bash
flutter pub get
```

## Quick Start

```dart
import 'package:openpay_flutter/openpay_flutter.dart';

final openpay = OpenPay('ak_live_xxx.sk_live_yyy');
```

## Create a Payment

```dart
final payment = await openpay.payments.create(CreatePaymentInput(
  amount: '1000.00',
  currency: 'LKR',
  merchantTradeNo: 'ORDER-123',
  description: 'Premium subscription',
));

print(payment.id);
print(payment.checkoutLink); // Redirect user here
print(payment.deepLink);     // Open wallet app directly
```

## Get Payment Status

```dart
final payment = await openpay.payments.get('pay_abc123');
print(payment.status); // PaymentStatus.paid, etc.
```

## List Payments

```dart
final result = await openpay.payments.list(ListPaymentsParams(
  page: 1,
  perPage: 20,
  status: PaymentStatus.paid,
));

for (final payment in result.data) {
  print('${payment.id}: ${payment.amount} ${payment.status.value}');
}
```

## Checkout Sessions

Create a hosted checkout session and open it in the browser:

```dart
import 'package:url_launcher/url_launcher.dart';

final session = await openpay.checkout.createSession(CheckoutSessionInput(
  amount: '2500.00',
  currency: 'LKR',
  successUrl: 'myapp://payment/success',
  cancelUrl: 'myapp://payment/cancel',
));

// Open checkout in browser
await launchUrl(Uri.parse(session.url));
```

## Webhook Verification

Verify incoming webhook signatures on your server:

```dart
import 'package:openpay_flutter/openpay_flutter.dart';

final event = verifyWebhookSignature(
  rawBody,
  {
    'x-webhook-signature': signature,
    'x-webhook-timestamp': timestamp,
    'x-webhook-event': eventType,
    'x-webhook-id': webhookId,
  },
  publicKey,
);

print(event.event); // 'payment.confirmed'
print(event.data);  // Payment data
```

## Configuration

```dart
final openpay = OpenPay(
  'ak_test_xxx.sk_test_yyy',
  baseURL: 'https://api.openpay.lk', // Custom API URL
  timeout: Duration(seconds: 15),     // Request timeout
);

// Don't forget to close when done
openpay.close();
```

## API Key Format

API keys follow the format `ak_{env}_{id}.sk_{env}_{secret}` where `env` is either `live` or `test`.

## Error Handling

```dart
try {
  final payment = await openpay.payments.create(CreatePaymentInput(
    amount: '1000.00',
  ));
} on AuthenticationError catch (e) {
  print('Bad API key: ${e.message}');
} on APIError catch (e) {
  print('API error [${e.code}]: ${e.message} (HTTP ${e.statusCode})');
} on OpenPayError catch (e) {
  print('SDK error: ${e.message}');
}
```

## License

MIT
