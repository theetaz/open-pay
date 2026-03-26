<?php

declare(strict_types=1);

namespace OpenPay\Webhook;

use OpenPay\Exceptions\OpenPayException;

class SignatureVerifier
{
    /**
     * Verify an incoming webhook signature using ED25519.
     *
     * @param string $payload The raw request body
     * @param string $signatureB64 The X-Webhook-Signature header (base64)
     * @param string $timestamp The X-Webhook-Timestamp header
     * @param string $event The X-Webhook-Event header
     * @param string $webhookId The X-Webhook-ID header
     * @param string $publicKeyB64 The ED25519 public key (base64)
     * @param int $maxAgeSeconds Maximum age of the webhook in seconds
     * @return array{id: string, event: string, timestamp: string, data: mixed}
     */
    public static function verify(
        string $payload,
        string $signatureB64,
        string $timestamp,
        string $event,
        string $webhookId,
        string $publicKeyB64,
        int $maxAgeSeconds = 300,
    ): array {
        if (empty($payload) || empty($signatureB64) || empty($timestamp) || empty($event) || empty($webhookId)) {
            throw new OpenPayException('Missing webhook signature parameters');
        }

        // Check timestamp freshness
        $ts = (int) $timestamp;
        $nowMs = (int) (microtime(true) * 1000);
        if (abs($nowMs - $ts) > $maxAgeSeconds * 1000) {
            throw new OpenPayException('Webhook timestamp too old');
        }

        // Verify ED25519 signature
        $publicKey = base64_decode($publicKeyB64, true);
        $signature = base64_decode($signatureB64, true);
        if ($publicKey === false || $signature === false) {
            throw new OpenPayException('Invalid base64 encoding');
        }

        $message = $timestamp . $payload;

        if (!sodium_crypto_sign_verify_detached($signature, $message, $publicKey)) {
            throw new OpenPayException('Invalid webhook signature');
        }

        return [
            'id' => $webhookId,
            'event' => $event,
            'timestamp' => $timestamp,
            'data' => json_decode($payload, true),
        ];
    }
}
