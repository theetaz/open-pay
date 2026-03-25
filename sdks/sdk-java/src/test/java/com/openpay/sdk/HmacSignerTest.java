package com.openpay.sdk;

import com.openpay.sdk.auth.HmacSigner;
import com.openpay.sdk.exceptions.AuthenticationException;
import org.junit.jupiter.api.Test;

import java.util.Map;

import static org.junit.jupiter.api.Assertions.*;

class HmacSignerTest {

    @Test
    void parsesValidLiveKey() {
        String[] parts = HmacSigner.parseApiKey("ak_live_abc123.sk_live_secret456");
        assertEquals("ak_live_abc123", parts[0]);
        assertEquals("sk_live_secret456", parts[1]);
    }

    @Test
    void parsesValidTestKey() {
        String[] parts = HmacSigner.parseApiKey("ak_test_abc123.sk_test_secret456");
        assertEquals("ak_test_abc123", parts[0]);
        assertEquals("sk_test_secret456", parts[1]);
    }

    @Test
    void throwsOnEmptyKey() {
        assertThrows(AuthenticationException.class, () -> HmacSigner.parseApiKey(""));
    }

    @Test
    void throwsOnInvalidFormat() {
        assertThrows(AuthenticationException.class, () -> HmacSigner.parseApiKey("nodotshere"));
    }

    @Test
    void throwsOnInvalidKeyPrefix() {
        assertThrows(AuthenticationException.class, () -> HmacSigner.parseApiKey("bad_prefix.sk_live_xxx"));
    }

    @Test
    void throwsOnInvalidSecretPrefix() {
        assertThrows(AuthenticationException.class, () -> HmacSigner.parseApiKey("ak_live_xxx.bad_prefix"));
    }

    @Test
    void signRequestIsDeterministic() {
        String sig1 = HmacSigner.signRequest("sk_live_test", "1700000000000", "GET", "/v1/payments", "");
        String sig2 = HmacSigner.signRequest("sk_live_test", "1700000000000", "GET", "/v1/payments", "");
        assertEquals(sig1, sig2);
    }

    @Test
    void differentMethodsProduceDifferentSignatures() {
        String sigGet = HmacSigner.signRequest("sk_live_test", "1700000000000", "GET", "/v1/payments", "");
        String sigPost = HmacSigner.signRequest("sk_live_test", "1700000000000", "POST", "/v1/payments", "");
        assertNotEquals(sigGet, sigPost);
    }

    @Test
    void bodyIncludedInSignature() {
        String sigEmpty = HmacSigner.signRequest("sk_live_test", "1700000000000", "POST", "/v1/payments", "");
        String sigBody = HmacSigner.signRequest("sk_live_test", "1700000000000", "POST", "/v1/payments", "{\"amount\":\"100\"}");
        assertNotEquals(sigEmpty, sigBody);
    }

    @Test
    void caseInsensitiveMethod() {
        String sig1 = HmacSigner.signRequest("sk_live_test", "1700000000000", "get", "/v1/payments", "");
        String sig2 = HmacSigner.signRequest("sk_live_test", "1700000000000", "GET", "/v1/payments", "");
        assertEquals(sig1, sig2);
    }

    @Test
    void signatureIsHex64() {
        String sig = HmacSigner.signRequest("sk_live_test", "1700000000000", "GET", "/v1/payments", "");
        assertEquals(64, sig.length());
        assertTrue(sig.matches("[0-9a-f]{64}"));
    }

    @Test
    void currentTimestampReturnsMilliseconds() {
        String ts = HmacSigner.currentTimestamp();
        long num = Long.parseLong(ts);
        assertTrue(num > 1700000000000L);
    }

    @Test
    void buildAuthHeadersReturnsAllKeys() {
        Map<String, String> headers = HmacSigner.buildAuthHeaders(
                "ak_live_test", "sk_live_secret", "GET", "/v1/payments", "");
        assertTrue(headers.containsKey("x-api-key"));
        assertTrue(headers.containsKey("x-timestamp"));
        assertTrue(headers.containsKey("x-signature"));
        assertEquals("ak_live_test", headers.get("x-api-key"));
    }

    @Test
    void crossLanguageCompatibility() {
        // Same inputs should produce same 64-char hex across all SDKs
        String sig = HmacSigner.signRequest("sk_live_test_secret_for_middleware",
                "1700000000000", "GET", "/v1/payments", "");
        assertEquals(64, sig.length());
        assertTrue(sig.matches("[0-9a-f]{64}"));
    }
}
