package com.openpay.sdk.resources;

import com.google.gson.JsonObject;
import com.openpay.sdk.OpenPayClient;

import java.util.Map;

/**
 * Webhook operations.
 */
public class WebhooksResource {

    private final OpenPayClient client;

    public WebhooksResource(OpenPayClient client) {
        this.client = client;
    }

    /**
     * Configure the webhook endpoint.
     */
    public void configure(String url, String... events) {
        Map<String, Object> body = events.length > 0
                ? Map.of("url", url, "events", events)
                : Map.of("url", url);
        client.request("POST", "/v1/sdk/webhooks/configure", body);
    }

    /**
     * Get the ED25519 public key for webhook signature verification.
     */
    public String getPublicKey() {
        JsonObject resp = client.request("GET", "/v1/sdk/webhooks/public-key", null);
        return resp.getAsJsonObject("data").get("publicKey").getAsString();
    }

    /**
     * Send a test webhook.
     */
    public void test() {
        client.request("POST", "/v1/sdk/webhooks/test", null);
    }
}
