package provider

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/shopspring/decimal"

	"github.com/openlankapay/openlankapay/services/payment/internal/adapter/wallet"
	"github.com/openlankapay/openlankapay/services/payment/internal/domain"
)

// TokenConfig describes a supported ERC20 token.
type TokenConfig struct {
	Symbol   string
	Address  string
	Decimals int32
}

// ConfirmationLookup is satisfied by the watcher's ConfirmedStore. The provider
// reads it so a detected on-chain transfer surfaces as a Paid status through the
// normal provider abstraction.
type ConfirmationLookup interface {
	Lookup(addr string) (txHash string, blockNumber int64, found bool)
}

// IndexAllocator returns the next monotonic HD-wallet deposit index (backed by a
// Postgres sequence).
type IndexAllocator func(ctx context.Context) (int64, error)

// OnchainProvider implements domain.PaymentProvider for the scan-and-send model:
// each payment gets a unique HD-derived deposit address; the buyer sends a plain
// ERC20 transfer; the watcher detects it and records a confirmation.
type OnchainProvider struct {
	wallet    *wallet.HDWallet
	chainID   int64
	tokens    map[string]TokenConfig // by symbol (USDC, USDT)
	confirmed ConfirmationLookup
	allocIdx  IndexAllocator
}

func NewOnchainProvider(w *wallet.HDWallet, chainID int64, tokens []TokenConfig,
	confirmed ConfirmationLookup, allocIdx IndexAllocator) *OnchainProvider {
	m := make(map[string]TokenConfig, len(tokens))
	for _, t := range tokens {
		m[strings.ToUpper(t.Symbol)] = t
	}
	return &OnchainProvider{
		wallet:    w,
		chainID:   chainID,
		tokens:    m,
		confirmed: confirmed,
		allocIdx:  allocIdx,
	}
}

func (p *OnchainProvider) Name() string { return "ONCHAIN" }

func (p *OnchainProvider) CreatePayment(ctx context.Context, req domain.ProviderPaymentRequest) (*domain.ProviderPaymentResponse, error) {
	token, ok := p.tokens[strings.ToUpper(req.Currency)]
	if !ok {
		return nil, fmt.Errorf("onchain: unsupported currency %q", req.Currency)
	}

	// Allocate a unique deposit index and derive its address.
	idx, err := p.allocIdx(ctx)
	if err != nil {
		return nil, fmt.Errorf("onchain: allocating deposit index: %w", err)
	}
	addr, err := p.wallet.DeriveAddress(uint32(idx))
	if err != nil {
		return nil, fmt.Errorf("onchain: deriving address: %w", err)
	}

	// Convert the requested amount to token base units (stablecoin 1:1 with USD).
	amt, err := decimal.NewFromString(req.Amount)
	if err != nil {
		return nil, fmt.Errorf("onchain: invalid amount %q: %w", req.Amount, err)
	}
	baseUnits := amt.Shift(token.Decimals).BigInt() // amount * 10^decimals

	qr := buildEIP681(token.Address, p.chainID, addr, baseUnits)

	return &domain.ProviderPaymentResponse{
		ProviderPayID: addr, // deposit address doubles as the provider pay id
		QRContent:     qr,
		CheckoutLink:  "",
		DeepLink:      qr,
		WalletAddress: addr,
		DepositIndex:  idx,
	}, nil
}

func (p *OnchainProvider) GetPaymentStatus(_ context.Context, providerPayID string) (*domain.ProviderPaymentStatus, error) {
	// providerPayID is the deposit address.
	txHash, blockNumber, found := p.confirmed.Lookup(providerPayID)
	if !found {
		return &domain.ProviderPaymentStatus{Status: domain.StatusInitiated}, nil
	}
	return &domain.ProviderPaymentStatus{
		Status:      domain.StatusPaid,
		TxHash:      txHash,
		BlockNumber: blockNumber,
	}, nil
}

// buildEIP681 builds an EIP-681 ERC20 transfer request URI that any scanning
// wallet (MetaMask, Trust, …) understands:
//
//	ethereum:<token>@<chainId>/transfer?address=<deposit>&uint256=<amount>
func buildEIP681(token string, chainID int64, to string, amount *big.Int) string {
	return fmt.Sprintf("ethereum:%s@%d/transfer?address=%s&uint256=%s",
		token, chainID, to, amount.String())
}
