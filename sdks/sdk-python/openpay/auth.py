"""HMAC-SHA256 authentication for Open Pay API requests."""

import hashlib
import hmac
import time

from openpay.errors import AuthenticationError


def parse_api_key(api_key: str) -> tuple[str, str]:
    """Parse a compound API key into (key_id, secret).

    Format: "ak_{env}_{id}.sk_{env}_{secret}"
    """
    if not api_key:
        raise AuthenticationError("API key is required")

    parts = api_key.split(".", 1)
    if len(parts) != 2:
        raise AuthenticationError("Invalid API key format")

    key_id, secret = parts
    if not key_id.startswith(("ak_live_", "ak_test_")):
        raise AuthenticationError("Invalid API key prefix")
    if not secret.startswith(("sk_live_", "sk_test_")):
        raise AuthenticationError("Invalid API secret prefix")

    return key_id, secret


def sign_request(secret: str, timestamp: str, method: str, path: str, body: str) -> str:
    """Sign an API request using HMAC-SHA256.

    Matches the Go implementation: signing key = SHA256(secret),
    message = timestamp + METHOD + path + body.
    """
    signing_key = hashlib.sha256(secret.encode()).digest()
    message = (timestamp + method.upper() + path + body).encode()
    return hmac.new(signing_key, message, hashlib.sha256).hexdigest()


def current_timestamp() -> str:
    """Get current timestamp as Unix milliseconds string."""
    return str(int(time.time() * 1000))


def build_auth_headers(key_id: str, secret: str, method: str, path: str, body: str) -> dict[str, str]:
    """Build authentication headers for an API request."""
    timestamp = current_timestamp()
    signature = sign_request(secret, timestamp, method, path, body)
    return {
        "x-api-key": key_id,
        "x-timestamp": timestamp,
        "x-signature": signature,
    }
