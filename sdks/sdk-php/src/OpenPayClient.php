<?php

declare(strict_types=1);

namespace OpenPay;

use GuzzleHttp\Client;
use GuzzleHttp\Exception\RequestException;
use OpenPay\Auth\HmacSigner;
use OpenPay\Exceptions\ApiException;
use OpenPay\Exceptions\AuthenticationException;
use OpenPay\Exceptions\OpenPayException;
use OpenPay\Resources\Payments;
use OpenPay\Resources\Webhooks;

/**
 * Open Pay API Client.
 *
 * Usage:
 *   $client = new OpenPayClient('ak_live_xxx.sk_live_yyy', [
 *       'base_url' => 'https://olp-api.nipuntheekshana.com',
 *   ]);
 *
 *   $payment = $client->payments->create([
 *       'amount' => '1000.00',
 *       'currency' => 'LKR',
 *       'merchantTradeNo' => 'ORDER-123',
 *   ]);
 */
class OpenPayClient
{
    private string $keyId;
    private string $secret;
    private string $baseUrl;
    private Client $http;

    public readonly Payments $payments;
    public readonly Webhooks $webhooks;

    /**
     * @param string $apiKey Compound API key: "ak_{env}_{id}.sk_{env}_{secret}"
     * @param array{base_url?: string, timeout?: float} $options
     */
    public function __construct(string $apiKey, array $options = [])
    {
        $parsed = HmacSigner::parseApiKey($apiKey);
        $this->keyId = $parsed['keyId'];
        $this->secret = $parsed['secret'];
        $this->baseUrl = rtrim($options['base_url'] ?? 'https://api.openpay.lk', '/');

        $this->http = new Client([
            'base_uri' => $this->baseUrl,
            'timeout' => $options['timeout'] ?? 30.0,
            'http_errors' => false,
        ]);

        $this->payments = new Payments($this);
        $this->webhooks = new Webhooks($this);
    }

    /**
     * Make an authenticated API request.
     *
     * @return array<string, mixed>
     */
    public function request(string $method, string $path, ?array $body = null): array
    {
        $bodyStr = $body !== null ? json_encode($body, JSON_UNESCAPED_SLASHES) : '';
        $headers = HmacSigner::buildAuthHeaders($this->keyId, $this->secret, $method, $path, $bodyStr);
        $headers['Content-Type'] = 'application/json';

        $options = ['headers' => $headers];
        if ($body !== null) {
            $options['body'] = $bodyStr;
        }

        try {
            $response = $this->http->request($method, $path, $options);
        } catch (RequestException $e) {
            throw new OpenPayException("Request failed: {$e->getMessage()}", 0, $e);
        }

        $statusCode = $response->getStatusCode();
        $data = json_decode((string) $response->getBody(), true) ?? [];

        if ($statusCode >= 400) {
            $error = $data['error'] ?? [];
            $code = $error['code'] ?? 'UNKNOWN_ERROR';
            $message = $error['message'] ?? "HTTP {$statusCode}";

            if ($statusCode === 401) {
                throw new AuthenticationException($message);
            }

            throw new ApiException($code, $message, $statusCode);
        }

        return $data;
    }
}
