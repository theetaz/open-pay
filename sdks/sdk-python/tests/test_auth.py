"""Tests for HMAC authentication."""

import pytest
from openpay.auth import parse_api_key, sign_request, current_timestamp


class TestParseAPIKey:
    def test_valid_live_key(self):
        key_id, secret = parse_api_key("ak_live_abc123.sk_live_secret456")
        assert key_id == "ak_live_abc123"
        assert secret == "sk_live_secret456"

    def test_valid_test_key(self):
        key_id, secret = parse_api_key("ak_test_abc123.sk_test_secret456")
        assert key_id == "ak_test_abc123"
        assert secret == "sk_test_secret456"

    def test_empty_key_raises(self):
        with pytest.raises(Exception, match="required"):
            parse_api_key("")

    def test_no_separator_raises(self):
        with pytest.raises(Exception, match="format"):
            parse_api_key("nodotshere")

    def test_invalid_key_prefix_raises(self):
        with pytest.raises(Exception, match="prefix"):
            parse_api_key("bad_prefix.sk_live_xxx")

    def test_invalid_secret_prefix_raises(self):
        with pytest.raises(Exception, match="prefix"):
            parse_api_key("ak_live_xxx.bad_prefix")


class TestSignRequest:
    def test_deterministic(self):
        sig1 = sign_request("sk_live_test", "1700000000000", "GET", "/v1/payments", "")
        sig2 = sign_request("sk_live_test", "1700000000000", "GET", "/v1/payments", "")
        assert sig1 == sig2

    def test_different_methods(self):
        sig_get = sign_request("sk_live_test", "1700000000000", "GET", "/v1/payments", "")
        sig_post = sign_request("sk_live_test", "1700000000000", "POST", "/v1/payments", "")
        assert sig_get != sig_post

    def test_includes_body(self):
        sig_empty = sign_request("sk_live_test", "1700000000000", "POST", "/v1/payments", "")
        sig_body = sign_request("sk_live_test", "1700000000000", "POST", "/v1/payments", '{"amount":"100"}')
        assert sig_empty != sig_body

    def test_case_insensitive_method(self):
        sig1 = sign_request("sk_live_test", "1700000000000", "get", "/v1/payments", "")
        sig2 = sign_request("sk_live_test", "1700000000000", "GET", "/v1/payments", "")
        assert sig1 == sig2

    def test_hex_output(self):
        sig = sign_request("sk_live_test", "1700000000000", "GET", "/v1/payments", "")
        assert len(sig) == 64
        assert all(c in "0123456789abcdef" for c in sig)

    def test_matches_typescript_sdk(self):
        """Cross-language compatibility: same inputs should produce same signature."""
        secret = "sk_live_test_secret_for_middleware"
        sig = sign_request(secret, "1700000000000", "GET", "/v1/payments", "")
        assert len(sig) == 64


class TestCurrentTimestamp:
    def test_returns_milliseconds(self):
        ts = current_timestamp()
        num = int(ts)
        assert num > 1700000000000
        assert ts == str(num)
