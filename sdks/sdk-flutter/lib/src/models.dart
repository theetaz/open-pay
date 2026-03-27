/// Payment status enumeration.
enum PaymentStatus {
  initiated('INITIATED'),
  pending('PENDING'),
  paid('PAID'),
  confirmed('CONFIRMED'),
  expired('EXPIRED'),
  failed('FAILED'),
  refunded('REFUNDED');

  final String value;
  const PaymentStatus(this.value);

  static PaymentStatus fromString(String value) {
    return PaymentStatus.values.firstWhere(
      (s) => s.value == value,
      orElse: () => PaymentStatus.initiated,
    );
  }
}

/// Input parameters for creating a payment.
class CreatePaymentInput {
  final String amount;
  final String? currency;
  final String? provider;
  final String? merchantTradeNo;
  final String? description;
  final String? webhookURL;
  final String? successURL;
  final String? cancelURL;
  final String? customerEmail;
  final CustomerBilling? customerBilling;
  final List<Good>? goods;
  final String? orderExpireTime;

  const CreatePaymentInput({
    required this.amount,
    this.currency,
    this.provider,
    this.merchantTradeNo,
    this.description,
    this.webhookURL,
    this.successURL,
    this.cancelURL,
    this.customerEmail,
    this.customerBilling,
    this.goods,
    this.orderExpireTime,
  });

  Map<String, dynamic> toJson() {
    final map = <String, dynamic>{'amount': amount};
    if (currency != null) map['currency'] = currency;
    if (provider != null) map['provider'] = provider;
    if (merchantTradeNo != null) map['merchantTradeNo'] = merchantTradeNo;
    if (description != null) map['description'] = description;
    if (webhookURL != null) map['webhookURL'] = webhookURL;
    if (successURL != null) map['successURL'] = successURL;
    if (cancelURL != null) map['cancelURL'] = cancelURL;
    if (customerEmail != null) map['customerEmail'] = customerEmail;
    if (customerBilling != null) {
      map['customerBilling'] = customerBilling!.toJson();
    }
    if (goods != null) map['goods'] = goods!.map((g) => g.toJson()).toList();
    if (orderExpireTime != null) map['orderExpireTime'] = orderExpireTime;
    return map;
  }
}

/// Customer billing information.
class CustomerBilling {
  final String? firstName;
  final String? lastName;
  final String? phone;

  const CustomerBilling({this.firstName, this.lastName, this.phone});

  Map<String, dynamic> toJson() {
    final map = <String, dynamic>{};
    if (firstName != null) map['firstName'] = firstName;
    if (lastName != null) map['lastName'] = lastName;
    if (phone != null) map['phone'] = phone;
    return map;
  }
}

/// Good/item in a payment.
class Good {
  final String name;
  final String? description;
  final String? mccCode;

  const Good({required this.name, this.description, this.mccCode});

  Map<String, dynamic> toJson() {
    final map = <String, dynamic>{'name': name};
    if (description != null) map['description'] = description;
    if (mccCode != null) map['mccCode'] = mccCode;
    return map;
  }
}

/// Payment response model.
class Payment {
  final String id;
  final String merchantId;
  final String? branchId;
  final String amount;
  final String currency;
  final PaymentStatus status;
  final String provider;
  final String? providerPayId;
  final String merchantTradeNo;
  final String? qrContent;
  final String? checkoutLink;
  final String? deepLink;
  final String? webhookURL;
  final String? successURL;
  final String? cancelURL;
  final String? customerEmail;
  final String? amountLkr;
  final String? exchangeRate;
  final String? platformFeeLkr;
  final String? exchangeFeeLkr;
  final String? netAmountLkr;
  final String? paidAt;
  final String? confirmedAt;
  final String createdAt;
  final String updatedAt;

  const Payment({
    required this.id,
    required this.merchantId,
    this.branchId,
    required this.amount,
    required this.currency,
    required this.status,
    required this.provider,
    this.providerPayId,
    required this.merchantTradeNo,
    this.qrContent,
    this.checkoutLink,
    this.deepLink,
    this.webhookURL,
    this.successURL,
    this.cancelURL,
    this.customerEmail,
    this.amountLkr,
    this.exchangeRate,
    this.platformFeeLkr,
    this.exchangeFeeLkr,
    this.netAmountLkr,
    this.paidAt,
    this.confirmedAt,
    required this.createdAt,
    required this.updatedAt,
  });

  factory Payment.fromJson(Map<String, dynamic> json) {
    return Payment(
      id: json['id'] as String,
      merchantId: json['merchantId'] as String,
      branchId: json['branchId'] as String?,
      amount: json['amount'] as String,
      currency: json['currency'] as String,
      status: PaymentStatus.fromString(json['status'] as String),
      provider: json['provider'] as String,
      providerPayId: json['providerPayId'] as String?,
      merchantTradeNo: json['merchantTradeNo'] as String,
      qrContent: json['qrContent'] as String?,
      checkoutLink: json['checkoutLink'] as String?,
      deepLink: json['deepLink'] as String?,
      webhookURL: json['webhookURL'] as String?,
      successURL: json['successURL'] as String?,
      cancelURL: json['cancelURL'] as String?,
      customerEmail: json['customerEmail'] as String?,
      amountLkr: json['amountLkr'] as String?,
      exchangeRate: json['exchangeRate'] as String?,
      platformFeeLkr: json['platformFeeLkr'] as String?,
      exchangeFeeLkr: json['exchangeFeeLkr'] as String?,
      netAmountLkr: json['netAmountLkr'] as String?,
      paidAt: json['paidAt'] as String?,
      confirmedAt: json['confirmedAt'] as String?,
      createdAt: json['createdAt'] as String,
      updatedAt: json['updatedAt'] as String,
    );
  }
}

/// List payments query parameters.
class ListPaymentsParams {
  final int? page;
  final int? perPage;
  final PaymentStatus? status;
  final String? search;
  final String? branchId;
  final String? dateFrom;
  final String? dateTo;

  const ListPaymentsParams({
    this.page,
    this.perPage,
    this.status,
    this.search,
    this.branchId,
    this.dateFrom,
    this.dateTo,
  });

  Map<String, String> toQueryParams() {
    final map = <String, String>{};
    if (page != null) map['page'] = page.toString();
    if (perPage != null) map['perPage'] = perPage.toString();
    if (status != null) map['status'] = status!.value;
    if (search != null) map['search'] = search!;
    if (branchId != null) map['branchId'] = branchId!;
    if (dateFrom != null) map['dateFrom'] = dateFrom!;
    if (dateTo != null) map['dateTo'] = dateTo!;
    return map;
  }
}

/// Paginated response wrapper.
class PaginatedResponse<T> {
  final List<T> data;
  final PaginationMeta meta;

  const PaginatedResponse({required this.data, required this.meta});
}

/// Pagination metadata.
class PaginationMeta {
  final int page;
  final int perPage;
  final int total;

  const PaginationMeta({
    required this.page,
    required this.perPage,
    required this.total,
  });

  factory PaginationMeta.fromJson(Map<String, dynamic> json) {
    return PaginationMeta(
      page: json['page'] as int,
      perPage: json['perPage'] as int,
      total: json['total'] as int,
    );
  }
}

/// Checkout session input parameters.
class CheckoutSessionInput {
  final String amount;
  final String? currency;
  final String? provider;
  final String? merchantTradeNo;
  final String? description;
  final String? successUrl;
  final String? cancelUrl;
  final String? customerEmail;
  final List<LineItem>? lineItems;
  final int? expiresInMinutes;

  const CheckoutSessionInput({
    required this.amount,
    this.currency,
    this.provider,
    this.merchantTradeNo,
    this.description,
    this.successUrl,
    this.cancelUrl,
    this.customerEmail,
    this.lineItems,
    this.expiresInMinutes,
  });

  Map<String, dynamic> toJson() {
    final map = <String, dynamic>{'amount': amount};
    if (currency != null) map['currency'] = currency;
    if (provider != null) map['provider'] = provider;
    if (merchantTradeNo != null) map['merchantTradeNo'] = merchantTradeNo;
    if (description != null) map['description'] = description;
    if (successUrl != null) map['successUrl'] = successUrl;
    if (cancelUrl != null) map['cancelUrl'] = cancelUrl;
    if (customerEmail != null) map['customerEmail'] = customerEmail;
    if (lineItems != null) {
      map['lineItems'] = lineItems!.map((i) => i.toJson()).toList();
    }
    if (expiresInMinutes != null) map['expiresInMinutes'] = expiresInMinutes;
    return map;
  }
}

/// Line item in a checkout session.
class LineItem {
  final String name;
  final String? description;
  final String? amount;

  const LineItem({required this.name, this.description, this.amount});

  Map<String, dynamic> toJson() {
    final map = <String, dynamic>{'name': name};
    if (description != null) map['description'] = description;
    if (amount != null) map['amount'] = amount;
    return map;
  }
}

/// Checkout session response model.
class CheckoutSession {
  final String id;
  final String paymentId;
  final String url;
  final String amount;
  final String currency;
  final String amountUsdt;
  final String status;
  final String qrContent;
  final String deepLink;
  final String merchantTradeNo;
  final String successUrl;
  final String cancelUrl;
  final String? exchangeRate;
  final String expiresAt;
  final String createdAt;

  const CheckoutSession({
    required this.id,
    required this.paymentId,
    required this.url,
    required this.amount,
    required this.currency,
    required this.amountUsdt,
    required this.status,
    required this.qrContent,
    required this.deepLink,
    required this.merchantTradeNo,
    required this.successUrl,
    required this.cancelUrl,
    this.exchangeRate,
    required this.expiresAt,
    required this.createdAt,
  });

  factory CheckoutSession.fromJson(Map<String, dynamic> json) {
    return CheckoutSession(
      id: json['id'] as String,
      paymentId: json['paymentId'] as String,
      url: json['url'] as String,
      amount: json['amount'] as String,
      currency: json['currency'] as String,
      amountUsdt: json['amountUsdt'] as String,
      status: json['status'] as String,
      qrContent: json['qrContent'] as String,
      deepLink: json['deepLink'] as String,
      merchantTradeNo: json['merchantTradeNo'] as String,
      successUrl: json['successUrl'] as String,
      cancelUrl: json['cancelUrl'] as String,
      exchangeRate: json['exchangeRate'] as String?,
      expiresAt: json['expiresAt'] as String,
      createdAt: json['createdAt'] as String,
    );
  }
}

/// Webhook event model.
class WebhookEvent {
  final String id;
  final String event;
  final String timestamp;
  final Map<String, dynamic> data;

  const WebhookEvent({
    required this.id,
    required this.event,
    required this.timestamp,
    required this.data,
  });
}
