"""Tests for the OpenPay client."""

import pytest
from openpay import OpenPay, AuthenticationError


class TestOpenPayClient:
    def test_creates_client(self):
        client = OpenPay("ak_live_xxx.sk_live_yyy", base_url="http://localhost:8080")
        assert client.payments is not None
        assert client.webhooks is not None

    def test_empty_key_raises(self):
        with pytest.raises(AuthenticationError):
            OpenPay("")

    def test_invalid_key_raises(self):
        with pytest.raises(AuthenticationError):
            OpenPay("invalid-key")

    def test_context_manager(self):
        with OpenPay("ak_live_xxx.sk_live_yyy") as client:
            assert client is not None


class TestIntegration:
    """Integration tests — require local dev environment (make start)."""

    API_KEY = ""  # Set via env or hardcode for manual testing

    @pytest.fixture
    def client(self):
        import os
        api_key = os.environ.get("OPENPAY_API_KEY", self.API_KEY)
        if not api_key:
            pytest.skip("OPENPAY_API_KEY not set")
        base_url = os.environ.get("OPENPAY_BASE_URL", "http://localhost:8080")
        return OpenPay(api_key, base_url=base_url)

    def test_list_payments(self, client):
        result = client.payments.list()
        assert result.meta.page == 1

    def test_create_payment(self, client):
        import time
        payment = client.payments.create(
            amount="75.00",
            currency="LKR",
            merchant_trade_no=f"PY-SDK-{int(time.time())}",
        )
        assert payment.id
        assert payment.status == "INITIATED"

    def test_get_payment(self, client):
        import time
        created = client.payments.create(
            amount="30.00",
            currency="LKR",
            merchant_trade_no=f"PY-GET-{int(time.time())}",
        )
        fetched = client.payments.get(created.id)
        assert fetched.id == created.id
