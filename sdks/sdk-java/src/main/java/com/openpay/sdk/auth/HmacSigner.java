package com.openpay.sdk.auth;

import com.openpay.sdk.exceptions.AuthenticationException;

import javax.crypto.Mac;
import javax.crypto.spec.SecretKeySpec;
import java.nio.charset.StandardCharsets;
import java.security.MessageDigest;
import java.util.Map;

/**
 * HMAC-SHA256 request signing for Open Pay API authentication.
 */
public final class HmacSigner {

    private HmacSigner() {}

    /**
     * Parse a compound API key into key ID and secret.
     * Format: "ak_{env}_{id}.sk_{env}_{secret}"
     */
    public static String[] parseApiKey(String apiKey) {
        if (apiKey == null || apiKey.isEmpty()) {
            throw new AuthenticationException("API key is required");
        }

        int dotIndex = apiKey.indexOf('.');
        if (dotIndex < 0) {
            throw new AuthenticationException("Invalid API key format");
        }

        String keyId = apiKey.substring(0, dotIndex);
        String secret = apiKey.substring(dotIndex + 1);

        if (!keyId.startsWith("ak_live_") && !keyId.startsWith("ak_test_")) {
            throw new AuthenticationException("Invalid API key prefix");
        }
        if (!secret.startsWith("sk_live_") && !secret.startsWith("sk_test_")) {
            throw new AuthenticationException("Invalid API secret prefix");
        }

        return new String[]{keyId, secret};
    }

    /**
     * Sign an API request using HMAC-SHA256.
     * Matches Go/TS/Python/PHP implementations.
     */
    public static String signRequest(String secret, String timestamp, String method, String path, String body) {
        try {
            // Derive signing key: SHA256(secret)
            byte[] signingKey = MessageDigest.getInstance("SHA-256")
                    .digest(secret.getBytes(StandardCharsets.UTF_8));

            // Build message: timestamp + METHOD + path + body
            String message = timestamp + method.toUpperCase() + path + (body != null ? body : "");

            Mac mac = Mac.getInstance("HmacSHA256");
            mac.init(new SecretKeySpec(signingKey, "HmacSHA256"));
            byte[] signature = mac.doFinal(message.getBytes(StandardCharsets.UTF_8));

            return bytesToHex(signature);
        } catch (Exception e) {
            throw new RuntimeException("Failed to sign request", e);
        }
    }

    /**
     * Get current timestamp as Unix milliseconds string.
     */
    public static String currentTimestamp() {
        return String.valueOf(System.currentTimeMillis());
    }

    /**
     * Build authentication headers for an API request.
     */
    public static Map<String, String> buildAuthHeaders(String keyId, String secret, String method, String path, String body) {
        String timestamp = currentTimestamp();
        String signature = signRequest(secret, timestamp, method, path, body);
        return Map.of(
                "x-api-key", keyId,
                "x-timestamp", timestamp,
                "x-signature", signature
        );
    }

    private static String bytesToHex(byte[] bytes) {
        StringBuilder sb = new StringBuilder(bytes.length * 2);
        for (byte b : bytes) {
            sb.append(String.format("%02x", b));
        }
        return sb.toString();
    }
}
