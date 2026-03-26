<?php
/**
 * Open Pay WooCommerce Payment Gateway.
 *
 * @package OpenPay
 */

defined('ABSPATH') || exit;

class WC_OpenPay_Gateway extends WC_Payment_Gateway
{
    private string $api_key;
    private string $api_base_url;

    public function __construct()
    {
        $this->id = 'openpay';
        $this->icon = OPENPAY_PLUGIN_URL . 'assets/icon.svg';
        $this->has_fields = false;
        $this->method_title = 'Open Pay';
        $this->method_description = 'Accept crypto payments (USDT, BTC, ETH) that settle in LKR via Open Pay.';
        $this->supports = ['products'];

        // Load settings
        $this->init_form_fields();
        $this->init_settings();

        $this->title = $this->get_option('title', 'Pay with Crypto');
        $this->description = $this->get_option('description', 'Pay securely with cryptocurrency. Scan the QR code with your crypto wallet.');
        $this->enabled = $this->get_option('enabled', 'no');
        $this->api_key = $this->get_option('api_key', '');
        $this->api_base_url = $this->get_option('api_base_url', 'https://olp-api.nipuntheekshana.com');

        // Save settings hook
        add_action('woocommerce_update_options_payment_gateways_' . $this->id, [$this, 'process_admin_options']);

        // Webhook handler
        add_action('woocommerce_api_openpay_webhook', [$this, 'handle_webhook']);

        // Thank you page
        add_action('woocommerce_thankyou_' . $this->id, [$this, 'thankyou_page']);
    }

    /**
     * Admin settings fields.
     */
    public function init_form_fields(): void
    {
        $this->form_fields = [
            'enabled' => [
                'title' => 'Enable/Disable',
                'type' => 'checkbox',
                'label' => 'Enable Open Pay Gateway',
                'default' => 'no',
            ],
            'title' => [
                'title' => 'Title',
                'type' => 'text',
                'description' => 'Payment method title shown to customers at checkout.',
                'default' => 'Pay with Crypto',
            ],
            'description' => [
                'title' => 'Description',
                'type' => 'textarea',
                'description' => 'Payment method description shown to customers.',
                'default' => 'Pay securely with cryptocurrency (USDT, BTC, ETH). You will be redirected to complete the payment.',
            ],
            'api_key' => [
                'title' => 'API Key',
                'type' => 'password',
                'description' => 'Your Open Pay API key (ak_live_xxx.sk_live_yyy). Get it from the Merchant Portal → Integrations.',
                'default' => '',
            ],
            'api_base_url' => [
                'title' => 'API Base URL',
                'type' => 'text',
                'description' => 'Open Pay API endpoint.',
                'default' => 'https://olp-api.nipuntheekshana.com',
            ],
            'webhook_section' => [
                'title' => 'Webhook',
                'type' => 'title',
                'description' => 'Set this as your webhook URL in Open Pay: <code>' . home_url('/wc-api/openpay_webhook') . '</code>',
            ],
        ];
    }

    /**
     * Process the payment — create a checkout session and redirect.
     */
    public function process_payment($order_id): array
    {
        $order = wc_get_order($order_id);
        if (!$order) {
            wc_add_notice('Order not found.', 'error');
            return ['result' => 'fail'];
        }

        try {
            $session = $this->create_checkout_session($order);

            // Store payment ID on the order
            $order->update_meta_data('_openpay_payment_id', $session['paymentId']);
            $order->update_meta_data('_openpay_session_id', $session['id']);
            $order->save();

            // Mark as pending
            $order->update_status('pending', 'Awaiting Open Pay crypto payment.');

            // Empty the cart
            WC()->cart->empty_cart();

            return [
                'result' => 'success',
                'redirect' => $session['url'],
            ];
        } catch (\Exception $e) {
            wc_add_notice('Payment error: ' . $e->getMessage(), 'error');
            return ['result' => 'fail'];
        }
    }

    /**
     * Create a checkout session via Open Pay API.
     */
    private function create_checkout_session(\WC_Order $order): array
    {
        $body = wp_json_encode([
            'amount' => $order->get_total(),
            'currency' => $order->get_currency(),
            'merchantTradeNo' => (string) $order->get_id(),
            'successUrl' => $this->get_return_url($order),
            'cancelUrl' => $order->get_cancel_order_url(),
            'customerEmail' => $order->get_billing_email(),
            'lineItems' => $this->get_line_items($order),
            'expiresInMinutes' => 15,
        ]);

        $path = '/v1/sdk/checkout/sessions';
        $timestamp = (string) intval(microtime(true) * 1000);
        $signing_key = hash('sha256', $this->get_secret(), true);
        $message = $timestamp . 'POST' . $path . $body;
        $signature = hash_hmac('sha256', $message, $signing_key);

        $response = wp_remote_post($this->api_base_url . $path, [
            'timeout' => 30,
            'headers' => [
                'Content-Type' => 'application/json',
                'x-api-key' => $this->get_key_id(),
                'x-timestamp' => $timestamp,
                'x-signature' => $signature,
            ],
            'body' => $body,
        ]);

        if (is_wp_error($response)) {
            throw new \Exception('API request failed: ' . $response->get_error_message());
        }

        $status = wp_remote_retrieve_response_code($response);
        $data = json_decode(wp_remote_retrieve_body($response), true);

        if ($status !== 201) {
            $error = $data['error']['message'] ?? 'Unknown error';
            throw new \Exception('Open Pay error: ' . $error);
        }

        return $data['data'];
    }

    /**
     * Handle incoming webhook from Open Pay.
     */
    public function handle_webhook(): void
    {
        $payload = file_get_contents('php://input');
        $event = sanitize_text_field($_SERVER['HTTP_X_WEBHOOK_EVENT'] ?? '');
        $webhookId = sanitize_text_field($_SERVER['HTTP_X_WEBHOOK_ID'] ?? '');

        if (empty($event) || empty($payload)) {
            status_header(400);
            exit('Missing webhook data');
        }

        $data = json_decode($payload, true);

        // TODO: Verify ED25519 signature in production
        // For now, process the event

        switch ($event) {
            case 'payment.paid':
                $this->handle_payment_paid($data);
                break;
            case 'payment.expired':
                $this->handle_payment_expired($data);
                break;
            case 'payment.failed':
                $this->handle_payment_failed($data);
                break;
        }

        status_header(200);
        exit('OK');
    }

    private function handle_payment_paid(array $data): void
    {
        $order_id = $data['merchantTradeNo'] ?? '';
        $order = wc_get_order($order_id);
        if (!$order) {
            return;
        }

        $order->payment_complete($data['id'] ?? '');
        $order->add_order_note('Open Pay payment confirmed. Transaction: ' . ($data['txHash'] ?? 'N/A'));
    }

    private function handle_payment_expired(array $data): void
    {
        $order_id = $data['merchantTradeNo'] ?? '';
        $order = wc_get_order($order_id);
        if (!$order) {
            return;
        }

        $order->update_status('cancelled', 'Open Pay payment expired.');
    }

    private function handle_payment_failed(array $data): void
    {
        $order_id = $data['merchantTradeNo'] ?? '';
        $order = wc_get_order($order_id);
        if (!$order) {
            return;
        }

        $order->update_status('failed', 'Open Pay payment failed.');
    }

    /**
     * Thank you page content.
     */
    public function thankyou_page(int $order_id): void
    {
        $order = wc_get_order($order_id);
        if (!$order) {
            return;
        }

        $payment_id = $order->get_meta('_openpay_payment_id');
        if ($payment_id) {
            echo '<p><strong>Open Pay Payment ID:</strong> ' . esc_html($payment_id) . '</p>';
        }
    }

    private function get_line_items(\WC_Order $order): array
    {
        $items = [];
        foreach ($order->get_items() as $item) {
            $items[] = [
                'name' => $item->get_name(),
                'description' => 'Qty: ' . $item->get_quantity(),
                'amount' => (string) $item->get_total(),
            ];
        }
        return $items;
    }

    private function get_key_id(): string
    {
        $parts = explode('.', $this->api_key, 2);
        return $parts[0] ?? '';
    }

    private function get_secret(): string
    {
        $parts = explode('.', $this->api_key, 2);
        return $parts[1] ?? '';
    }
}
