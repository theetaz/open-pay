package provider

import (
	"context"
	"testing"

	"github.com/openlankapay/openlankapay/services/payment/internal/adapter/wallet"
	"github.com/openlankapay/openlankapay/services/payment/internal/domain"
)

type fakeLookup struct{}

func (fakeLookup) Lookup(string) (string, int64, bool) { return "", 0, false }

func newTestProvider(t *testing.T) *OnchainProvider {
	t.Helper()
	hw, err := wallet.NewHDWallet("test test test test test test test test test test test junk")
	if err != nil {
		t.Fatalf("wallet: %v", err)
	}
	idx := int64(0)
	alloc := func(context.Context) (int64, error) { idx++; return idx, nil }
	tokens := []TokenConfig{
		{Symbol: "USDC", Address: "0x6c904e10631e199134891158B97311ed088A6a16", Decimals: 6},
		{Symbol: "USDT", Address: "0x5fB234eceBE3475E51F2A76bCE4a5d446D37CAB1", Decimals: 6},
	}
	return NewOnchainProvider(hw, 97, tokens, fakeLookup{}, alloc)
}

func TestOnchain_CreatePayment_EIP681(t *testing.T) {
	p := newTestProvider(t)

	resp, err := p.CreatePayment(context.Background(), domain.ProviderPaymentRequest{
		Amount:   "5",
		Currency: "USDC",
		OrderID:  "order-1",
	})
	if err != nil {
		t.Fatalf("CreatePayment: %v", err)
	}

	// 5 USDC at 6 decimals = 5000000 base units.
	want := "ethereum:0x6c904e10631e199134891158B97311ed088A6a16@97/transfer?address=" +
		resp.WalletAddress + "&uint256=5000000"
	if resp.QRContent != want {
		t.Errorf("QR\n got=%s\nwant=%s", resp.QRContent, want)
	}
	if resp.WalletAddress == "" || resp.WalletAddress != resp.ProviderPayID {
		t.Errorf("WalletAddress/ProviderPayID mismatch: %q vs %q", resp.WalletAddress, resp.ProviderPayID)
	}
	if resp.DepositIndex != 1 {
		t.Errorf("DepositIndex = %d, want 1", resp.DepositIndex)
	}
}

func TestOnchain_CreatePayment_FractionalUSDT(t *testing.T) {
	p := newTestProvider(t)
	resp, err := p.CreatePayment(context.Background(), domain.ProviderPaymentRequest{
		Amount:   "12.34",
		Currency: "USDT",
		OrderID:  "order-2",
	})
	if err != nil {
		t.Fatalf("CreatePayment: %v", err)
	}
	// 12.34 * 1e6 = 12340000
	if got := resp.QRContent; got[len(got)-8:] != "12340000" {
		t.Errorf("expected amount 12340000 in QR, got %s", got)
	}
}

func TestOnchain_CreatePayment_UnsupportedCurrency(t *testing.T) {
	p := newTestProvider(t)
	_, err := p.CreatePayment(context.Background(), domain.ProviderPaymentRequest{
		Amount: "5", Currency: "BTC", OrderID: "x",
	})
	if err == nil {
		t.Error("expected error for unsupported currency, got nil")
	}
}

func TestOnchain_GetPaymentStatus_Initiated(t *testing.T) {
	p := newTestProvider(t)
	st, err := p.GetPaymentStatus(context.Background(), "0xabc")
	if err != nil {
		t.Fatalf("GetPaymentStatus: %v", err)
	}
	if st.Status != domain.StatusInitiated {
		t.Errorf("status = %s, want INITIATED (no confirmation yet)", st.Status)
	}
}
