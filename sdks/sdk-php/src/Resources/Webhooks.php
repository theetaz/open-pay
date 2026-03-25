<?php

declare(strict_types=1);

namespace OpenPay\Resources;

use OpenPay\OpenPayClient;

class Webhooks
{
    public function __construct(private readonly OpenPayClient $client)
    {
    }

    /**
     * Configure the webhook endpoint.
     *
     * @param string $url The webhook endpoint URL
     * @param string[] $events Optional event filter patterns
     */
    public function configure(string $url, array $events = []): void
    {
        $body = ['url' => $url];
        if (!empty($events)) {
            $body['events'] = $events;
        }
        $this->client->request('POST', '/v1/sdk/webhooks/configure', $body);
    }

    /**
     * Get the ED25519 public key for verifying webhook signatures.
     */
    public function getPublicKey(): string
    {
        $response = $this->client->request('GET', '/v1/sdk/webhooks/public-key');
        return $response['data']['publicKey'];
    }

    /**
     * Send a test webhook to the configured endpoint.
     */
    public function test(): void
    {
        $this->client->request('POST', '/v1/sdk/webhooks/test');
    }
}
