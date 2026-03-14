package domain

import "context"

// ProviderPaymentRequest is the input for creating a payment with a provider.
type ProviderPaymentRequest struct {
	Amount   string
	Currency string
	OrderID  string
}

// ProviderPaymentResponse is the result of creating a payment with a provider.
type ProviderPaymentResponse struct {
	ProviderPayID string
	QRContent     string
	CheckoutLink  string
	DeepLink      string
}

// ProviderPaymentStatus is the status returned by a provider.
type ProviderPaymentStatus struct {
	Status PaymentStatus
	TxHash string
}

// PaymentProvider defines the abstraction over exchange partner APIs.
type PaymentProvider interface {
	CreatePayment(ctx context.Context, req ProviderPaymentRequest) (*ProviderPaymentResponse, error)
	GetPaymentStatus(ctx context.Context, providerPayID string) (*ProviderPaymentStatus, error)
	Name() string
}
