<?php
/**
 * Plugin Name: Open Pay for WooCommerce
 * Plugin URI: https://github.com/theetaz/open-pay
 * Description: Accept crypto payments that settle in LKR via Open Pay payment gateway.
 * Version: 0.1.0
 * Author: Open Pay
 * Author URI: https://github.com/theetaz/open-pay
 * License: MIT
 * Requires PHP: 8.1
 * Requires at least: 6.0
 * WC requires at least: 8.0
 *
 * @package OpenPay
 */

defined('ABSPATH') || exit;

define('OPENPAY_VERSION', '0.1.0');
define('OPENPAY_PLUGIN_DIR', plugin_dir_path(__FILE__));
define('OPENPAY_PLUGIN_URL', plugin_dir_url(__FILE__));

// Check if WooCommerce is active
add_action('plugins_loaded', 'openpay_init_gateway', 11);

function openpay_init_gateway(): void
{
    if (!class_exists('WC_Payment_Gateway')) {
        add_action('admin_notices', function () {
            echo '<div class="error"><p><strong>Open Pay</strong> requires WooCommerce to be installed and active.</p></div>';
        });
        return;
    }

    require_once OPENPAY_PLUGIN_DIR . 'includes/class-openpay-gateway.php';

    add_filter('woocommerce_payment_gateways', function (array $gateways): array {
        $gateways[] = 'WC_OpenPay_Gateway';
        return $gateways;
    });
}

// Declare HPOS compatibility
add_action('before_woocommerce_init', function () {
    if (class_exists('\Automattic\WooCommerce\Utilities\FeaturesUtil')) {
        \Automattic\WooCommerce\Utilities\FeaturesUtil::declare_compatibility('custom_order_tables', __FILE__, true);
    }
});
