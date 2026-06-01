// Package wallet derives deterministic per-payment deposit addresses from a
// BIP39 mnemonic using the Ethereum BIP44 path m/44'/60'/0'/0/index.
//
// Only addresses are needed for payment detection. The private-key derivation
// is kept (unexported) so a future "sweep" feature can move funds out of the
// deposit addresses; nothing here persists private keys.
package wallet

import (
	"errors"
	"fmt"

	btcec "github.com/btcsuite/btcd/btcec/v2"
	bip32 "github.com/tyler-smith/go-bip32"
	bip39 "github.com/tyler-smith/go-bip39"
	"golang.org/x/crypto/sha3"
)

// hardened is the BIP32 hardened-key offset (2^31).
const hardened uint32 = 0x80000000

// HDWallet derives Ethereum addresses from a master seed.
type HDWallet struct {
	// accountKey is the node at m/44'/60'/0'/0 — children are external addresses.
	accountKey *bip32.Key
}

// NewHDWallet builds an HD wallet from a BIP39 mnemonic phrase.
func NewHDWallet(mnemonic string) (*HDWallet, error) {
	if !bip39.IsMnemonicValid(mnemonic) {
		return nil, errors.New("wallet: invalid BIP39 mnemonic")
	}
	seed := bip39.NewSeed(mnemonic, "")

	master, err := bip32.NewMasterKey(seed)
	if err != nil {
		return nil, fmt.Errorf("wallet: master key: %w", err)
	}

	// m/44'/60'/0'/0
	path := []uint32{hardened + 44, hardened + 60, hardened + 0, 0}
	key := master
	for _, idx := range path {
		key, err = key.NewChildKey(idx)
		if err != nil {
			return nil, fmt.Errorf("wallet: deriving path: %w", err)
		}
	}
	return &HDWallet{accountKey: key}, nil
}

// DeriveAddress returns the EIP-55 checksummed 0x address at the external chain
// index (m/44'/60'/0'/0/index).
func (w *HDWallet) DeriveAddress(index uint32) (string, error) {
	priv, err := w.derivePrivateKey(index)
	if err != nil {
		return "", err
	}
	return addressFromPriv(priv), nil
}

// derivePrivateKey returns the secp256k1 private key at the given index. Kept
// unexported and unused by the detection path — reserved for a future sweep.
func (w *HDWallet) derivePrivateKey(index uint32) (*btcec.PrivateKey, error) {
	child, err := w.accountKey.NewChildKey(index)
	if err != nil {
		return nil, fmt.Errorf("wallet: child key %d: %w", index, err)
	}
	priv, _ := btcec.PrivKeyFromBytes(child.Key)
	return priv, nil
}

// addressFromPriv computes the Ethereum address: keccak256 of the uncompressed
// public key (64 bytes, without the 0x04 prefix), last 20 bytes, EIP-55 cased.
func addressFromPriv(priv *btcec.PrivateKey) string {
	pub := priv.PubKey()
	uncompressed := pub.SerializeUncompressed() // 65 bytes: 0x04 || X(32) || Y(32)

	h := sha3.NewLegacyKeccak256()
	h.Write(uncompressed[1:]) // drop 0x04 prefix
	sum := h.Sum(nil)
	addr := sum[12:] // last 20 bytes

	return toChecksumAddress(addr)
}

// toChecksumAddress applies the EIP-55 mixed-case checksum.
func toChecksumAddress(addr []byte) string {
	const hexChars = "0123456789abcdef"
	lower := make([]byte, 40)
	for i, b := range addr {
		lower[i*2] = hexChars[b>>4]
		lower[i*2+1] = hexChars[b&0x0f]
	}

	h := sha3.NewLegacyKeccak256()
	h.Write(lower)
	hash := h.Sum(nil)

	out := make([]byte, 0, 42)
	out = append(out, '0', 'x')
	for i, c := range lower {
		if c >= 'a' && c <= 'f' {
			// uppercase if the corresponding hash nibble >= 8
			nibble := hash[i/2]
			if i%2 == 0 {
				nibble >>= 4
			} else {
				nibble &= 0x0f
			}
			if nibble >= 8 {
				c -= 32 // to uppercase
			}
		}
		out = append(out, c)
	}
	return string(out)
}
