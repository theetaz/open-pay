package watcher

import (
	"math/big"
	"strings"
	"sync"
)

// Confirmation is a detected, sufficiently-confirmed inbound transfer to a
// deposit address.
type Confirmation struct {
	TxHash      string
	BlockNumber int64
	Amount      *big.Int // received amount in token base units
}

// ConfirmedStore records confirmed transfers keyed by lower-cased deposit
// address. The on-chain provider's GetPaymentStatus reads from it, so a
// detected payment surfaces as "Paid" through the normal provider flow.
type ConfirmedStore struct {
	mu sync.RWMutex
	m  map[string]Confirmation
}

func NewStore() *ConfirmedStore {
	return &ConfirmedStore{m: make(map[string]Confirmation)}
}

func (s *ConfirmedStore) Put(addr string, c Confirmation) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.m[strings.ToLower(addr)] = c
}

// Get returns the confirmation for an address, if any.
func (s *ConfirmedStore) Get(addr string) (Confirmation, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	c, ok := s.m[strings.ToLower(addr)]
	return c, ok
}

// Has reports whether a confirmation is already recorded (idempotency guard).
func (s *ConfirmedStore) Has(addr string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.m[strings.ToLower(addr)]
	return ok
}

// Lookup adapts the store to the provider's ConfirmationLookup interface,
// returning the tx hash and block number of a confirmed deposit.
func (s *ConfirmedStore) Lookup(addr string) (txHash string, blockNumber int64, found bool) {
	c, ok := s.Get(addr)
	if !ok {
		return "", 0, false
	}
	return c.TxHash, c.BlockNumber, true
}
