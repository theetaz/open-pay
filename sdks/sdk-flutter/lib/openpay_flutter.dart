/// Open Pay Flutter SDK — crypto-to-fiat payment processing for mobile apps.
library openpay_flutter;

import 'dart:convert';

import 'package:crypto/crypto.dart';
import 'package:http/http.dart' as http;

import 'src/models.dart';

export 'src/models.dart';

const String _defaultBaseUrl = 'https://api.openpay.lk';

// ─── Errors ───

/// Base error class for Open Pay SDK errors.
class OpenPayError implements Exception {
  final String message;
  const OpenPayError(this.message);

  @override
  String toString() => 'OpenPayError: $message';
}

/// Thrown when API authentication fails.
class AuthenticationError extends OpenPayError {
  const AuthenticationError([String message = 'Authentication failed'])
      : super(message);

  @override
  String toString() => 'AuthenticationError: $message';
}

/// Thrown when the API returns an error response.
class APIError extends OpenPayError {
  final String code;
  final int statusCode;

  const APIError(this.code, String message, this.statusCode) : super(message);

  @override
  String toString() => 'APIError [$code]: $message (HTTP $statusCode)';
}

// ─── Auth ───

/// Parsed API key components.
class _ParsedKey {
  final String keyId;
  final String secret;
  const _ParsedKey(this.keyId, this.secret);
}

/// Parse a compound API key into key ID and secret.
/// Format: "ak_{env}_{id}.sk_{env}_{secret}"
_ParsedKey _parseAPIKey(String apiKey) {
  if (apiKey.isEmpty) {
    throw const AuthenticationError('API key is required');
  }

  final parts = apiKey.split('.');
  if (parts.length != 2) {
    throw const AuthenticationError('Invalid API key format');
  }

  final keyId = parts[0];
  final secret = parts[1];

  if (!keyId.startsWith('ak_live_') && !keyId.startsWith('ak_test_')) {
    throw const AuthenticationError('Invalid API key prefix');
  }
  if (!secret.startsWith('sk_live_') && !secret.startsWith('sk_test_')) {
    throw const AuthenticationError('Invalid API secret prefix');
  }

  return _ParsedKey(keyId, secret);
}

/// Sign an API request using HMAC-SHA256.
/// Matches the Go implementation: signing key = SHA256(secret), message = timestamp + METHOD + path + body.
String _signRequest(
  String secret,
  String timestamp,
  String method,
  String path,
  String body,
) {
  final signingKey = sha256.convert(utf8.encode(secret)).bytes;
  final message = utf8.encode(timestamp + method.toUpperCase() + path + body);
  final hmacSha256 = Hmac(sha256, signingKey);
  final digest = hmacSha256.convert(message);
  return digest.toString();
}

/// Build authentication headers for an API request.
Map<String, String> _buildAuthHeaders(
  String keyId,
  String secret,
  String method,
  String path,
  String body,
) {
  final timestamp = DateTime.now().millisecondsSinceEpoch.toString();
  final signature = _signRequest(secret, timestamp, method, path, body);

  return {
    'x-api-key': keyId,
    'x-timestamp': timestamp,
    'x-signature': signature,
  };
}

// ─── Webhook Verification ───

/// Verify an incoming webhook signature.
///
/// Validates the structure and timestamp freshness. For full ED25519 signature
/// verification, use the `pointycastle` package or verify on your backend.
///
/// [payload] - The raw request body string
/// [headers] - Map containing x-webhook-signature, x-webhook-timestamp, x-webhook-event, x-webhook-id
/// [publicKey] - The ED25519 public key (base64-encoded) — reserved for future use
WebhookEvent verifyWebhookSignature(
  String payload,
  Map<String, String> headers,
  String publicKey,
) {
  final signature = headers['x-webhook-signature'];
  final timestamp = headers['x-webhook-timestamp'];
  final event = headers['x-webhook-event'];
  final id = headers['x-webhook-id'];

  if (signature == null || timestamp == null || event == null || id == null) {
    throw const OpenPayError('Missing webhook signature headers');
  }

  // Check timestamp freshness (5 minutes)
  final ts = int.parse(timestamp);
  final now = DateTime.now().millisecondsSinceEpoch;
  if ((now - ts).abs() > 5 * 60 * 1000) {
    throw const OpenPayError('Webhook timestamp too old');
  }

  return WebhookEvent(
    id: id,
    event: event,
    timestamp: timestamp,
    data: jsonDecode(payload) as Map<String, dynamic>,
  );
}

// ─── Main Client ───

/// Open Pay Flutter SDK Client.
///
/// ```dart
/// final openpay = OpenPay('ak_live_xxx.sk_live_yyy');
///
/// final payment = await openpay.payments.create(CreatePaymentInput(
///   amount: '1000.00',
///   currency: 'LKR',
///   merchantTradeNo: 'ORDER-123',
/// ));
///
/// print(payment.id);
/// print(payment.checkoutLink);
/// ```
class OpenPay {
  final String _keyId;
  final String _secret;
  final String _baseURL;
  final Duration _timeout;
  final http.Client _httpClient;

  late final PaymentsResource payments;
  late final CheckoutResource checkout;

  OpenPay(
    String apiKey, {
    String? baseURL,
    Duration timeout = const Duration(seconds: 30),
    http.Client? httpClient,
  })  : _keyId = _parseAPIKey(apiKey).keyId,
        _secret = _parseAPIKey(apiKey).secret,
        _baseURL = (baseURL ?? _defaultBaseUrl).replaceAll(RegExp(r'/$'), ''),
        _timeout = timeout,
        _httpClient = httpClient ?? http.Client() {
    payments = PaymentsResource._(this);
    checkout = CheckoutResource._(this);
  }

  /// Make an authenticated API request.
  Future<Map<String, dynamic>> request(
    String method,
    String path, [
    Map<String, dynamic>? body,
  ]) async {
    final bodyStr = body != null ? jsonEncode(body) : '';
    final headers = _buildAuthHeaders(_keyId, _secret, method, path, bodyStr);
    headers['Content-Type'] = 'application/json';

    final url = Uri.parse('$_baseURL$path');

    http.Response response;
    try {
      switch (method.toUpperCase()) {
        case 'GET':
          response = await _httpClient
              .get(url, headers: headers)
              .timeout(_timeout);
          break;
        case 'POST':
          response = await _httpClient
              .post(url, headers: headers, body: bodyStr.isNotEmpty ? bodyStr : null)
              .timeout(_timeout);
          break;
        case 'PUT':
          response = await _httpClient
              .put(url, headers: headers, body: bodyStr.isNotEmpty ? bodyStr : null)
              .timeout(_timeout);
          break;
        case 'DELETE':
          response = await _httpClient
              .delete(url, headers: headers)
              .timeout(_timeout);
          break;
        default:
          throw OpenPayError('Unsupported HTTP method: $method');
      }
    } on Exception catch (e) {
      if (e.toString().contains('TimeoutException')) {
        throw const OpenPayError('Request timed out');
      }
      throw OpenPayError('Request failed: $e');
    }

    final json = jsonDecode(response.body) as Map<String, dynamic>;

    if (response.statusCode < 200 || response.statusCode >= 300) {
      final error = json['error'] as Map<String, dynamic>?;
      if (response.statusCode == 401) {
        throw AuthenticationError(
          error?['message'] as String? ?? 'Authentication failed',
        );
      }
      throw APIError(
        error?['code'] as String? ?? 'UNKNOWN_ERROR',
        error?['message'] as String? ?? 'HTTP ${response.statusCode}',
        response.statusCode,
      );
    }

    return json;
  }

  /// Close the underlying HTTP client.
  void close() {
    _httpClient.close();
  }
}

// ─── Payment Resource ───

/// Resource for managing payments.
class PaymentsResource {
  final OpenPay _client;
  PaymentsResource._(this._client);

  /// Create a new payment.
  Future<Payment> create(CreatePaymentInput input) async {
    final json = await _client.request('POST', '/v1/sdk/payments', input.toJson());
    return Payment.fromJson(json['data'] as Map<String, dynamic>);
  }

  /// Get a payment by ID.
  Future<Payment> get(String id) async {
    final json = await _client.request('GET', '/v1/sdk/payments/$id');
    return Payment.fromJson(json['data'] as Map<String, dynamic>);
  }

  /// List payments with optional filtering and pagination.
  Future<PaginatedResponse<Payment>> list([ListPaymentsParams? params]) async {
    final queryParams = params?.toQueryParams() ?? {};
    final queryString = queryParams.entries
        .map((e) => '${Uri.encodeComponent(e.key)}=${Uri.encodeComponent(e.value)}')
        .join('&');
    final path =
        '/v1/sdk/payments${queryString.isNotEmpty ? '?$queryString' : ''}';

    final json = await _client.request('GET', path);
    final dataList = json['data'] as List<dynamic>;
    final payments = dataList
        .map((item) => Payment.fromJson(item as Map<String, dynamic>))
        .toList();
    final meta =
        PaginationMeta.fromJson(json['meta'] as Map<String, dynamic>);

    return PaginatedResponse(data: payments, meta: meta);
  }
}

// ─── Checkout Resource ───

/// Resource for managing checkout sessions.
class CheckoutResource {
  final OpenPay _client;
  CheckoutResource._(this._client);

  /// Create a checkout session. Returns a hosted checkout URL.
  Future<CheckoutSession> createSession(CheckoutSessionInput input) async {
    final json = await _client.request(
      'POST',
      '/v1/sdk/checkout/sessions',
      input.toJson(),
    );
    return CheckoutSession.fromJson(json['data'] as Map<String, dynamic>);
  }
}
