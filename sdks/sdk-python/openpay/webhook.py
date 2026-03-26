"""Webhook signature verification using ED25519."""

import json
import time
from cryptography.hazmat.primitives.asymmetric.ed25519 import Ed25519PublicKey
from cryptography.hazmat.primitives.serialization import Encoding, PublicFormat
from cryptography.exceptions import InvalidSignature
import base64

from openpay.errors import OpenPayError
from openpay.types import WebhookEvent


def verify_webhook_signature(
    payload: str,
    signature_b64: str,
    timestamp: str,
    event: str,
    webhook_id: str,
    public_key_b64: str,
    max_age_seconds: int = 300,
) -> WebhookEvent:
    """Verify an incoming webhook signature using ED25519.

    Args:
        payload: The raw request body string
        signature_b64: The X-Webhook-Signature header (base64)
        timestamp: The X-Webhook-Timestamp header
        event: The X-Webhook-Event header
        webhook_id: The X-Webhook-ID header
        public_key_b64: The ED25519 public key (base64)
        max_age_seconds: Maximum age of the webhook in seconds (default 300)

    Returns:
        Parsed WebhookEvent

    Raises:
        OpenPayError: If verification fails
    """
    if not all([payload, signature_b64, timestamp, event, webhook_id, public_key_b64]):
        raise OpenPayError("Missing webhook signature parameters")

    # Check timestamp freshness
    try:
        ts = int(timestamp)
        if abs(time.time() * 1000 - ts) > max_age_seconds * 1000:
            raise OpenPayError("Webhook timestamp too old")
    except ValueError:
        raise OpenPayError("Invalid webhook timestamp")

    # Verify signature
    try:
        pub_bytes = base64.b64decode(public_key_b64)
        sig_bytes = base64.b64decode(signature_b64)
        message = timestamp.encode() + payload.encode()

        public_key = Ed25519PublicKey.from_public_bytes(pub_bytes)
        public_key.verify(sig_bytes, message)
    except InvalidSignature:
        raise OpenPayError("Invalid webhook signature")
    except Exception as e:
        raise OpenPayError(f"Webhook verification failed: {e}")

    return WebhookEvent(
        id=webhook_id,
        event=event,
        timestamp=timestamp,
        data=json.loads(payload),
    )
