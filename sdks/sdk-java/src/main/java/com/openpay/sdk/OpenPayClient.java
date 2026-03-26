package com.openpay.sdk;

import com.google.gson.Gson;
import com.google.gson.JsonElement;
import com.google.gson.JsonObject;
import com.openpay.sdk.auth.HmacSigner;
import com.openpay.sdk.exceptions.ApiException;
import com.openpay.sdk.exceptions.AuthenticationException;
import com.openpay.sdk.exceptions.OpenPayException;
import com.openpay.sdk.resources.PaymentsResource;
import com.openpay.sdk.resources.WebhooksResource;

import java.io.IOException;
import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.time.Duration;
import java.util.Map;

/**
 * Open Pay API Client.
 *
 * <pre>{@code
 * var client = new OpenPayClient("ak_live_xxx.sk_live_yyy");
 * var payment = client.payments().create(Map.of(
 *     "amount", "1000.00",
 *     "currency", "LKR"
 * ));
 * }</pre>
 */
public class OpenPayClient {

    private static final String DEFAULT_BASE_URL = "https://api.openpay.lk";
    private static final Gson GSON = new Gson();

    private final String keyId;
    private final String secret;
    private final String baseUrl;
    private final HttpClient http;

    private final PaymentsResource payments;
    private final WebhooksResource webhooks;

    public OpenPayClient(String apiKey) {
        this(apiKey, DEFAULT_BASE_URL);
    }

    public OpenPayClient(String apiKey, String baseUrl) {
        String[] parts = HmacSigner.parseApiKey(apiKey);
        this.keyId = parts[0];
        this.secret = parts[1];
        this.baseUrl = baseUrl.replaceAll("/+$", "");
        this.http = HttpClient.newBuilder()
                .connectTimeout(Duration.ofSeconds(10))
                .build();

        this.payments = new PaymentsResource(this);
        this.webhooks = new WebhooksResource(this);
    }

    public PaymentsResource payments() { return payments; }
    public WebhooksResource webhooks() { return webhooks; }

    /**
     * Make an authenticated API request.
     */
    public JsonObject request(String method, String path, Object body) {
        String bodyStr = body != null ? GSON.toJson(body) : "";
        Map<String, String> authHeaders = HmacSigner.buildAuthHeaders(keyId, secret, method, path, bodyStr);

        HttpRequest.Builder reqBuilder = HttpRequest.newBuilder()
                .uri(URI.create(baseUrl + path))
                .timeout(Duration.ofSeconds(30))
                .header("Content-Type", "application/json");

        for (var entry : authHeaders.entrySet()) {
            reqBuilder.header(entry.getKey(), entry.getValue());
        }

        if ("POST".equals(method) || "PUT".equals(method)) {
            reqBuilder.method(method, HttpRequest.BodyPublishers.ofString(bodyStr));
        } else {
            reqBuilder.method(method, HttpRequest.BodyPublishers.noBody());
        }

        try {
            HttpResponse<String> resp = http.send(reqBuilder.build(), HttpResponse.BodyHandlers.ofString());
            JsonObject data = GSON.fromJson(resp.body(), JsonObject.class);

            if (resp.statusCode() >= 400) {
                JsonElement errorElem = data != null ? data.get("error") : null;
                String code = "UNKNOWN_ERROR";
                String message = "HTTP " + resp.statusCode();
                if (errorElem != null && errorElem.isJsonObject()) {
                    JsonObject error = errorElem.getAsJsonObject();
                    if (error.has("code")) code = error.get("code").getAsString();
                    if (error.has("message")) message = error.get("message").getAsString();
                }
                if (resp.statusCode() == 401) {
                    throw new AuthenticationException(message);
                }
                throw new ApiException(code, message, resp.statusCode());
            }

            return data;
        } catch (IOException | InterruptedException e) {
            throw new OpenPayException("Request failed: " + e.getMessage(), e);
        }
    }
}
