package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

// ServiceProxy holds reverse proxies to downstream services.
type ServiceProxy struct {
	MerchantProxy     *httputil.ReverseProxy
	PaymentProxy      *httputil.ReverseProxy
	ExchangeProxy     *httputil.ReverseProxy
	SettlementProxy   *httputil.ReverseProxy
	WebhookProxy      *httputil.ReverseProxy
	SubscriptionProxy *httputil.ReverseProxy
	NotificationProxy *httputil.ReverseProxy
}

// Config holds the URLs for downstream services.
type Config struct {
	MerchantServiceURL     string
	PaymentServiceURL      string
	ExchangeServiceURL     string
	SettlementServiceURL   string
	WebhookServiceURL      string
	SubscriptionServiceURL string
	NotificationServiceURL string
}

// NewServiceProxy creates a proxy that forwards requests to downstream services.
func NewServiceProxy(cfg Config) *ServiceProxy {
	return &ServiceProxy{
		MerchantProxy:     newProxy(cfg.MerchantServiceURL),
		PaymentProxy:      newProxy(cfg.PaymentServiceURL),
		ExchangeProxy:     newProxy(cfg.ExchangeServiceURL),
		SettlementProxy:   newProxy(cfg.SettlementServiceURL),
		WebhookProxy:      newProxy(cfg.WebhookServiceURL),
		SubscriptionProxy: newProxy(cfg.SubscriptionServiceURL),
		NotificationProxy: newProxy(cfg.NotificationServiceURL),
	}
}

func newProxy(targetURL string) *httputil.ReverseProxy {
	target, err := url.Parse(targetURL)
	if err != nil {
		panic("invalid proxy target URL: " + targetURL)
	}

	proxy := httputil.NewSingleHostReverseProxy(target)

	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		// Preserve the original path — don't strip prefixes
		req.Host = target.Host
	}

	proxy.ErrorHandler = func(w http.ResponseWriter, _ *http.Request, _ error) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(`{"error":{"code":"SERVICE_UNAVAILABLE","message":"downstream service unavailable"}}`))
	}

	return proxy
}

// ProxyToMerchant forwards the request to the merchant service.
func (p *ServiceProxy) ProxyToMerchant(w http.ResponseWriter, r *http.Request) {
	p.MerchantProxy.ServeHTTP(w, r)
}

// ProxyToPayment forwards the request to the payment service.
func (p *ServiceProxy) ProxyToPayment(w http.ResponseWriter, r *http.Request) {
	p.PaymentProxy.ServeHTTP(w, r)
}

// ProxyToExchange forwards the request to the exchange service.
func (p *ServiceProxy) ProxyToExchange(w http.ResponseWriter, r *http.Request) {
	p.ExchangeProxy.ServeHTTP(w, r)
}

// ProxyToSettlement forwards the request to the settlement service.
func (p *ServiceProxy) ProxyToSettlement(w http.ResponseWriter, r *http.Request) {
	p.SettlementProxy.ServeHTTP(w, r)
}

// ProxyToWebhook forwards the request to the webhook service.
func (p *ServiceProxy) ProxyToWebhook(w http.ResponseWriter, r *http.Request) {
	p.WebhookProxy.ServeHTTP(w, r)
}

// ProxyToSubscription forwards the request to the subscription service.
func (p *ServiceProxy) ProxyToSubscription(w http.ResponseWriter, r *http.Request) {
	p.SubscriptionProxy.ServeHTTP(w, r)
}

// ProxyToNotification forwards the request to the notification service.
func (p *ServiceProxy) ProxyToNotification(w http.ResponseWriter, r *http.Request) {
	p.NotificationProxy.ServeHTTP(w, r)
}
