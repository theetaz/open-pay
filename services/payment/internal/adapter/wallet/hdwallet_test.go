package wallet

import "testing"

// Known derivation from the standard Hardhat/Anvil test mnemonic, verified
// against ethers.js HDNodeWallet at m/44'/60'/0'/0/index.
func TestDeriveAddress_KnownVectors(t *testing.T) {
	const mnemonic = "test test test test test test test test test test test junk"
	want := []string{
		"0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266",
		"0x70997970C51812dc3A010C7d01b50e0d17dc79C8",
		"0x3C44CdDdB6a900fa2b585dd299e03d12FA4293BC",
	}

	w, err := NewHDWallet(mnemonic)
	if err != nil {
		t.Fatalf("NewHDWallet: %v", err)
	}

	for i, exp := range want {
		got, err := w.DeriveAddress(uint32(i))
		if err != nil {
			t.Fatalf("DeriveAddress(%d): %v", i, err)
		}
		if got != exp {
			t.Errorf("index %d: got %s, want %s", i, got, exp)
		}
	}
}

func TestNewHDWallet_InvalidMnemonic(t *testing.T) {
	if _, err := NewHDWallet("not a valid mnemonic phrase at all"); err == nil {
		t.Error("expected error for invalid mnemonic, got nil")
	}
}

// Deterministic: same index always yields the same address.
func TestDeriveAddress_Deterministic(t *testing.T) {
	const mnemonic = "test test test test test test test test test test test junk"
	w, _ := NewHDWallet(mnemonic)
	a, _ := w.DeriveAddress(42)
	b, _ := w.DeriveAddress(42)
	if a != b {
		t.Errorf("non-deterministic: %s != %s", a, b)
	}
	c, _ := w.DeriveAddress(43)
	if a == c {
		t.Errorf("different indexes gave same address: %s", a)
	}
}
