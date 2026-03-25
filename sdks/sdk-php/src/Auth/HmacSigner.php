<?php

declare(strict_types=1);

namespace OpenPay\Auth;

use OpenPay\Exceptions\AuthenticationException;

class HmacSigner
{
    /**
     * Parse a compound API key into key ID and secret.
     * Format: "ak_{env}_{id}.sk_{env}_{secret}"
     *
     * @return array{keyId: string, secret: string}
     */
    public static function parseApiKey(string $apiKey): array
    {
        if (empty($apiKey)) {
            throw new AuthenticationException('API key is required');
        }

        $parts = explode('.', $apiKey, 2);
        if (count($parts) !== 2) {
            throw new AuthenticationException('Invalid API key format');
        }

        [$keyId, $secret] = $parts;

        if (!str_starts_with($keyId, 'ak_live_') && !str_starts_with($keyId, 'ak_test_')) {
            throw new AuthenticationException('Invalid API key prefix');
        }
        if (!str_starts_with($secret, 'sk_live_') && !str_starts_with($secret, 'sk_test_')) {
            throw new AuthenticationException('Invalid API secret prefix');
        }

        return ['keyId' => $keyId, 'secret' => $secret];
    }

    /**
     * Sign an API request using HMAC-SHA256.
     * Matches Go/TS/Python implementations: signing key = SHA256(secret),
     * message = timestamp + METHOD + path + body.
     */
    public static function signRequest(
        string $secret,
        string $timestamp,
        string $method,
        string $path,
        string $body = '',
    ): string {
        $signingKey = hash('sha256', $secret, true);
        $message = $timestamp . strtoupper($method) . $path . $body;
        return hash_hmac('sha256', $message, $signingKey);
    }

    /**
     * Get current timestamp as Unix milliseconds string.
     */
    public static function currentTimestamp(): string
    {
        return (string) intval(microtime(true) * 1000);
    }

    /**
     * Build authentication headers for an API request.
     *
     * @return array<string, string>
     */
    public static function buildAuthHeaders(
        string $keyId,
        string $secret,
        string $method,
        string $path,
        string $body = '',
    ): array {
        $timestamp = self::currentTimestamp();
        $signature = self::signRequest($secret, $timestamp, $method, $path, $body);

        return [
            'x-api-key' => $keyId,
            'x-timestamp' => $timestamp,
            'x-signature' => $signature,
        ];
    }
}
