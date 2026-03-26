"""Open Pay Python SDK client."""

from __future__ import annotations

from typing import Any, Optional

import httpx

from openpay.auth import build_auth_headers, parse_api_key
from openpay.errors import APIError, AuthenticationError, OpenPayError
from openpay.types import PaginatedPayments, Payment


DEFAULT_BASE_URL = "https://api.openpay.lk"


class OpenPay:
    """Open Pay API client.

    Usage::

        from openpay import OpenPay

        client = OpenPay("ak_live_xxx.sk_live_yyy", base_url="http://localhost:8080")

        # Create a payment
        payment = client.payments.create(amount="1000.00", currency="LKR")

        # List payments
        result = client.payments.list(status="PAID")
    """

    def __init__(self, api_key: str, *, base_url: str = DEFAULT_BASE_URL, timeout: float = 30.0):
        self._key_id, self._secret = parse_api_key(api_key)
        self._base_url = base_url.rstrip("/")
        self._http = httpx.Client(timeout=timeout)

        self.payments = PaymentsResource(self)
        self.checkout = CheckoutResource(self)
        self.webhooks = WebhooksResource(self)

    def close(self) -> None:
        """Close the HTTP client."""
        self._http.close()

    def __enter__(self) -> "OpenPay":
        return self

    def __exit__(self, *args: Any) -> None:
        self.close()

    def _request(self, method: str, path: str, body: Any = None) -> Any:
        """Make an authenticated API request."""
        import json as json_mod

        body_str = ""
        content = None
        if body is not None:
            body_str = json_mod.dumps(body, separators=(",", ":"))
            content = body_str.encode()

        headers = build_auth_headers(self._key_id, self._secret, method, path, body_str)
        headers["Content-Type"] = "application/json"

        try:
            resp = self._http.request(method, self._base_url + path, headers=headers, content=content)
        except httpx.TimeoutException:
            raise OpenPayError("Request timed out")
        except httpx.RequestError as e:
            raise OpenPayError(f"Request failed: {e}")

        data = resp.json()

        if resp.status_code >= 400:
            error = data.get("error", {})
            if resp.status_code == 401:
                raise AuthenticationError(error.get("message", "Authentication failed"))
            raise APIError(
                code=error.get("code", "UNKNOWN_ERROR"),
                message=error.get("message", f"HTTP {resp.status_code}"),
                status=resp.status_code,
            )

        return data


class PaymentsResource:
    """Payment operations."""

    def __init__(self, client: OpenPay):
        self._client = client

    def create(
        self,
        amount: str,
        currency: str = "LKR",
        *,
        provider: str = "",
        merchant_trade_no: str = "",
        description: str = "",
        webhook_url: str = "",
        success_url: str = "",
        cancel_url: str = "",
        customer_email: str = "",
    ) -> Payment:
        """Create a new payment."""
        body: dict[str, Any] = {"amount": amount, "currency": currency}
        if provider:
            body["provider"] = provider
        if merchant_trade_no:
            body["merchantTradeNo"] = merchant_trade_no
        if description:
            body["description"] = description
        if webhook_url:
            body["webhookURL"] = webhook_url
        if success_url:
            body["successURL"] = success_url
        if cancel_url:
            body["cancelURL"] = cancel_url
        if customer_email:
            body["customerEmail"] = customer_email

        resp = self._client._request("POST", "/v1/sdk/payments", body)
        return Payment.from_dict(resp["data"])

    def get(self, payment_id: str) -> Payment:
        """Get a payment by ID."""
        resp = self._client._request("GET", f"/v1/sdk/payments/{payment_id}")
        return Payment.from_dict(resp["data"])

    def list(
        self,
        *,
        page: int = 1,
        per_page: int = 20,
        status: str = "",
        search: str = "",
    ) -> PaginatedPayments:
        """List payments with optional filtering."""
        params = [f"page={page}", f"perPage={per_page}"]
        if status:
            params.append(f"status={status}")
        if search:
            params.append(f"search={search}")

        path = "/v1/sdk/payments?" + "&".join(params)
        resp = self._client._request("GET", path)
        return PaginatedPayments.from_dict(resp)


class CheckoutResource:
    """Checkout session operations."""

    def __init__(self, client: OpenPay):
        self._client = client

    def create_session(
        self,
        amount: str,
        currency: str = "LKR",
        *,
        success_url: str = "",
        cancel_url: str = "",
        customer_email: str = "",
        merchant_trade_no: str = "",
        description: str = "",
        expires_in_minutes: int = 0,
    ) -> dict[str, Any]:
        """Create a checkout session. Returns a hosted checkout URL."""
        body: dict[str, Any] = {"amount": amount, "currency": currency}
        if success_url:
            body["successUrl"] = success_url
        if cancel_url:
            body["cancelUrl"] = cancel_url
        if customer_email:
            body["customerEmail"] = customer_email
        if merchant_trade_no:
            body["merchantTradeNo"] = merchant_trade_no
        if description:
            body["description"] = description
        if expires_in_minutes > 0:
            body["expiresInMinutes"] = expires_in_minutes

        resp = self._client._request("POST", "/v1/sdk/checkout/sessions", body)
        return resp["data"]


class WebhooksResource:
    """Webhook operations."""

    def __init__(self, client: OpenPay):
        self._client = client

    def configure(self, url: str, events: Optional[list[str]] = None) -> None:
        """Configure the webhook endpoint."""
        body: dict[str, Any] = {"url": url}
        if events:
            body["events"] = events
        self._client._request("POST", "/v1/sdk/webhooks/configure", body)

    def get_public_key(self) -> str:
        """Get the ED25519 public key for webhook verification."""
        resp = self._client._request("GET", "/v1/sdk/webhooks/public-key")
        return resp["data"]["publicKey"]

    def test(self) -> None:
        """Send a test webhook."""
        self._client._request("POST", "/v1/sdk/webhooks/test")
