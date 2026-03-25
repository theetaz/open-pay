=== Open Pay for WooCommerce ===
Contributors: openpay
Tags: crypto, payments, woocommerce, lkr, usdt, bitcoin, ethereum
Requires at least: 6.0
Tested up to: 6.7
Requires PHP: 8.1
WC requires at least: 8.0
Stable tag: 0.1.0
License: MIT

Accept crypto payments (USDT, BTC, ETH) on your WooCommerce store. Payments settle in LKR to your bank account.

== Description ==

Open Pay for WooCommerce lets you accept cryptocurrency payments that automatically convert and settle in Sri Lankan Rupees (LKR).

**Features:**
* Accept USDT, BTC, ETH and more
* Automatic conversion to LKR at live exchange rates
* QR code + hosted checkout page for customers
* Webhook-based order status updates
* Support for Bybit, Binance Pay, and KuCoin

== Installation ==

1. Upload the plugin to `/wp-content/plugins/openpay-gateway`
2. Activate through the Plugins menu
3. Go to WooCommerce → Settings → Payments → Open Pay
4. Enter your API key (from Merchant Portal → Integrations)
5. Set the webhook URL in your Open Pay merchant portal

== Configuration ==

1. Get your API key from the [Merchant Portal](https://olp-merchant.nipuntheekshana.com/integrations)
2. In WooCommerce settings, paste your API key
3. Configure the webhook URL shown in the settings page
4. Enable the gateway and test with a small payment

== Changelog ==

= 0.1.0 =
* Initial release
* Checkout session creation with hosted payment page
* Webhook handling for payment.paid, payment.expired, payment.failed
* HMAC-SHA256 request signing
* WooCommerce HPOS compatibility
