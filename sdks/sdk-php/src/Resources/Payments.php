<?php

declare(strict_types=1);

namespace OpenPay\Resources;

use OpenPay\OpenPayClient;

class Payments
{
    public function __construct(private readonly OpenPayClient $client)
    {
    }

    /**
     * Create a new payment.
     *
     * @param array{
     *     amount: string,
     *     currency?: string,
     *     provider?: string,
     *     merchantTradeNo?: string,
     *     description?: string,
     *     webhookURL?: string,
     *     successURL?: string,
     *     cancelURL?: string,
     *     customerEmail?: string,
     * } $params
     * @return array<string, mixed>
     */
    public function create(array $params): array
    {
        $response = $this->client->request('POST', '/v1/sdk/payments', $params);
        return $response['data'];
    }

    /**
     * Get a payment by ID.
     *
     * @return array<string, mixed>
     */
    public function get(string $id): array
    {
        $response = $this->client->request('GET', "/v1/sdk/payments/{$id}");
        return $response['data'];
    }

    /**
     * List payments with optional filtering.
     *
     * @param array{
     *     page?: int,
     *     perPage?: int,
     *     status?: string,
     *     search?: string,
     * } $params
     * @return array{data: array<array<string, mixed>>, meta: array{page: int, perPage: int, total: int}}
     */
    public function list(array $params = []): array
    {
        $query = http_build_query(array_filter($params));
        $path = '/v1/sdk/payments' . ($query ? "?{$query}" : '');
        return $this->client->request('GET', $path);
    }
}
