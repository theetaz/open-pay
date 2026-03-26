package com.openpay.sdk.resources;

import com.google.gson.JsonObject;
import com.openpay.sdk.OpenPayClient;

import java.util.Map;
import java.util.StringJoiner;

/**
 * Payment operations.
 */
public class PaymentsResource {

    private final OpenPayClient client;

    public PaymentsResource(OpenPayClient client) {
        this.client = client;
    }

    /**
     * Create a new payment.
     *
     * @param params Must include "amount". Optional: "currency", "merchantTradeNo", "description", etc.
     */
    public JsonObject create(Map<String, Object> params) {
        JsonObject resp = client.request("POST", "/v1/sdk/payments", params);
        return resp.getAsJsonObject("data");
    }

    /**
     * Get a payment by ID.
     */
    public JsonObject get(String id) {
        JsonObject resp = client.request("GET", "/v1/sdk/payments/" + id, null);
        return resp.getAsJsonObject("data");
    }

    /**
     * List payments with optional filtering.
     */
    public JsonObject list(Map<String, String> params) {
        String path = "/v1/sdk/payments";
        if (params != null && !params.isEmpty()) {
            StringJoiner joiner = new StringJoiner("&", "?", "");
            params.forEach((k, v) -> {
                if (v != null && !v.isEmpty()) {
                    joiner.add(k + "=" + v);
                }
            });
            path += joiner.toString();
        }
        return client.request("GET", path, null);
    }
}
