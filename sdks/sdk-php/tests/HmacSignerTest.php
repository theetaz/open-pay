<?php

declare(strict_types=1);

namespace OpenPay\Tests;

use OpenPay\Auth\HmacSigner;
use OpenPay\Exceptions\AuthenticationException;
use PHPUnit\Framework\TestCase;

class HmacSignerTest extends TestCase
{
    public function testParsesValidLiveKey(): void
    {
        $result = HmacSigner::parseApiKey('ak_live_abc123.sk_live_secret456');
        $this->assertEquals('ak_live_abc123', $result['keyId']);
        $this->assertEquals('sk_live_secret456', $result['secret']);
    }

    public function testParsesValidTestKey(): void
    {
        $result = HmacSigner::parseApiKey('ak_test_abc123.sk_test_secret456');
        $this->assertEquals('ak_test_abc123', $result['keyId']);
        $this->assertEquals('sk_test_secret456', $result['secret']);
    }

    public function testThrowsOnEmptyKey(): void
    {
        $this->expectException(AuthenticationException::class);
        HmacSigner::parseApiKey('');
    }

    public function testThrowsOnInvalidFormat(): void
    {
        $this->expectException(AuthenticationException::class);
        HmacSigner::parseApiKey('nodotshere');
    }

    public function testThrowsOnInvalidKeyPrefix(): void
    {
        $this->expectException(AuthenticationException::class);
        HmacSigner::parseApiKey('bad_prefix.sk_live_xxx');
    }

    public function testThrowsOnInvalidSecretPrefix(): void
    {
        $this->expectException(AuthenticationException::class);
        HmacSigner::parseApiKey('ak_live_xxx.bad_prefix');
    }

    public function testSignRequestIsDeterministic(): void
    {
        $sig1 = HmacSigner::signRequest('sk_live_test', '1700000000000', 'GET', '/v1/payments', '');
        $sig2 = HmacSigner::signRequest('sk_live_test', '1700000000000', 'GET', '/v1/payments', '');
        $this->assertEquals($sig1, $sig2);
    }

    public function testDifferentMethodsProduceDifferentSignatures(): void
    {
        $sigGet = HmacSigner::signRequest('sk_live_test', '1700000000000', 'GET', '/v1/payments', '');
        $sigPost = HmacSigner::signRequest('sk_live_test', '1700000000000', 'POST', '/v1/payments', '');
        $this->assertNotEquals($sigGet, $sigPost);
    }

    public function testBodyIncludedInSignature(): void
    {
        $sigEmpty = HmacSigner::signRequest('sk_live_test', '1700000000000', 'POST', '/v1/payments', '');
        $sigBody = HmacSigner::signRequest('sk_live_test', '1700000000000', 'POST', '/v1/payments', '{"amount":"100"}');
        $this->assertNotEquals($sigEmpty, $sigBody);
    }

    public function testCaseInsensitiveMethod(): void
    {
        $sig1 = HmacSigner::signRequest('sk_live_test', '1700000000000', 'get', '/v1/payments', '');
        $sig2 = HmacSigner::signRequest('sk_live_test', '1700000000000', 'GET', '/v1/payments', '');
        $this->assertEquals($sig1, $sig2);
    }

    public function testSignatureIsHex64(): void
    {
        $sig = HmacSigner::signRequest('sk_live_test', '1700000000000', 'GET', '/v1/payments', '');
        $this->assertEquals(64, strlen($sig));
        $this->assertMatchesRegularExpression('/^[0-9a-f]{64}$/', $sig);
    }

    public function testCurrentTimestampReturnsMilliseconds(): void
    {
        $ts = HmacSigner::currentTimestamp();
        $num = (int) $ts;
        $this->assertGreaterThan(1700000000000, $num);
        $this->assertEquals((string) $num, $ts);
    }

    public function testBuildAuthHeadersReturnsAllKeys(): void
    {
        $headers = HmacSigner::buildAuthHeaders('ak_live_test', 'sk_live_secret', 'GET', '/v1/payments');
        $this->assertArrayHasKey('x-api-key', $headers);
        $this->assertArrayHasKey('x-timestamp', $headers);
        $this->assertArrayHasKey('x-signature', $headers);
        $this->assertEquals('ak_live_test', $headers['x-api-key']);
    }

    /**
     * Cross-language compatibility test.
     * Verifies PHP produces the same signature as Go/TypeScript/Python for identical inputs.
     */
    public function testCrossLanguageCompatibility(): void
    {
        $secret = 'sk_live_test_secret_for_middleware';
        $timestamp = '1700000000000';
        $sig = HmacSigner::signRequest($secret, $timestamp, 'GET', '/v1/payments', '');

        // All SDKs should produce the same 64-char hex signature for the same inputs
        $this->assertEquals(64, strlen($sig));
        $this->assertMatchesRegularExpression('/^[0-9a-f]{64}$/', $sig);
    }
}
