"""Type definitions for the Open Pay SDK."""

from dataclasses import dataclass, field
from typing import Any, Optional


@dataclass
class Payment:
    id: str
    merchant_id: str
    amount: str
    currency: str
    status: str
    provider: str
    merchant_trade_no: str
    created_at: str
    updated_at: str
    branch_id: str = ""
    provider_pay_id: str = ""
    qr_content: str = ""
    checkout_link: str = ""
    deep_link: str = ""
    webhook_url: str = ""
    success_url: str = ""
    cancel_url: str = ""
    customer_email: str = ""
    amount_lkr: str = ""
    exchange_rate: str = ""
    platform_fee_lkr: str = ""
    exchange_fee_lkr: str = ""
    net_amount_lkr: str = ""
    paid_at: Optional[str] = None
    confirmed_at: Optional[str] = None

    @classmethod
    def from_dict(cls, data: dict[str, Any]) -> "Payment":
        return cls(
            id=data.get("id", ""),
            merchant_id=data.get("merchantId", ""),
            amount=data.get("amount", ""),
            currency=data.get("currency", ""),
            status=data.get("status", ""),
            provider=data.get("provider", ""),
            merchant_trade_no=data.get("merchantTradeNo", ""),
            created_at=data.get("createdAt", ""),
            updated_at=data.get("updatedAt", ""),
            branch_id=data.get("branchId", ""),
            provider_pay_id=data.get("providerPayId", ""),
            qr_content=data.get("qrContent", ""),
            checkout_link=data.get("checkoutLink", ""),
            deep_link=data.get("deepLink", ""),
            webhook_url=data.get("webhookURL", ""),
            success_url=data.get("successURL", ""),
            cancel_url=data.get("cancelURL", ""),
            customer_email=data.get("customerEmail", ""),
            amount_lkr=data.get("amountLkr", ""),
            exchange_rate=data.get("exchangeRate", ""),
            platform_fee_lkr=data.get("platformFeeLkr", ""),
            exchange_fee_lkr=data.get("exchangeFeeLkr", ""),
            net_amount_lkr=data.get("netAmountLkr", ""),
            paid_at=data.get("paidAt"),
            confirmed_at=data.get("confirmedAt"),
        )


@dataclass
class PaginationMeta:
    page: int
    per_page: int
    total: int

    @classmethod
    def from_dict(cls, data: dict[str, Any]) -> "PaginationMeta":
        return cls(page=data.get("page", 1), per_page=data.get("perPage", 20), total=data.get("total", 0))


@dataclass
class PaginatedPayments:
    data: list[Payment]
    meta: PaginationMeta

    @classmethod
    def from_dict(cls, raw: dict[str, Any]) -> "PaginatedPayments":
        payments = [Payment.from_dict(p) for p in raw.get("data", [])]
        meta = PaginationMeta.from_dict(raw.get("meta", {}))
        return cls(data=payments, meta=meta)


@dataclass
class WebhookEvent:
    id: str
    event: str
    timestamp: str
    data: dict[str, Any] = field(default_factory=dict)
