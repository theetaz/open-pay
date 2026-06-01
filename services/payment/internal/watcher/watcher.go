// Package watcher polls a BSC (EVM) chain for ERC20 Transfers into per-payment
// deposit addresses and drives the existing payment confirmation flow.
//
// It runs as a goroutine inside the payment service. On a qualifying transfer
// it records a Confirmation in the store and calls OnConfirm(paymentID), which
// is wired to PaymentService.HandleProviderCallback — reusing the same
// paid -> credit path a real exchange webhook would hit.
package watcher

import (
	"context"
	"math/big"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Logger is the minimal logging surface the watcher needs. Both *slog.Logger
// and a thin zerolog adapter satisfy it.
type Logger interface {
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

// PendingDeposit is an open on-chain payment awaiting a transfer.
type PendingDeposit struct {
	PaymentID uuid.UUID
	Address   string   // deposit address (lower-cased for matching)
	Token     string   // token contract address
	Expected  *big.Int // expected amount in token base units
}

// PendingRepo supplies open on-chain payments and persists the scan cursor.
type PendingRepo interface {
	ListPendingDeposits(ctx context.Context) ([]PendingDeposit, error)
	GetCursor(ctx context.Context, id string) (int64, bool, error)
	SetCursor(ctx context.Context, id string, block int64) error
}

// Config tunes the watcher.
type Config struct {
	Tokens        []string      // token contract addresses to watch
	Confirmations int64         // blocks to wait before crediting
	PollInterval  time.Duration // time between scans
	MaxBlockRange int64         // max blocks per eth_getLogs call (RPC limit)
	BootLookback  int64         // blocks to re-scan on first start
}

// Watcher polls the chain and confirms deposits.
type Watcher struct {
	rpc       *RPCClient
	repo      PendingRepo
	store     *ConfirmedStore
	onConfirm func(ctx context.Context, paymentID uuid.UUID) error
	cfg       Config
	log       Logger
}

const cursorID = "bsc-onchain"

func New(rpc *RPCClient, repo PendingRepo, store *ConfirmedStore,
	onConfirm func(context.Context, uuid.UUID) error, cfg Config, log Logger) *Watcher {
	if cfg.MaxBlockRange <= 0 {
		cfg.MaxBlockRange = 40 // public BSC RPCs are stingy; stay small
	}
	if cfg.PollInterval <= 0 {
		cfg.PollInterval = 15 * time.Second
	}
	if cfg.Confirmations <= 0 {
		cfg.Confirmations = 3
	}
	if cfg.BootLookback <= 0 {
		cfg.BootLookback = 200
	}
	return &Watcher{rpc: rpc, repo: repo, store: store, onConfirm: onConfirm, cfg: cfg, log: log}
}

// Run scans until ctx is cancelled.
func (w *Watcher) Run(ctx context.Context) {
	head, err := w.rpc.BlockNumber(ctx)
	if err != nil {
		w.log.Error("watcher: initial block number failed", "err", err)
	}

	// Seed the cursor if missing: start a small lookback behind head so a
	// payment that landed during a restart is still picked up.
	if cur, ok, _ := w.repo.GetCursor(ctx, cursorID); !ok {
		start := head - w.cfg.BootLookback
		if start < 0 {
			start = 0
		}
		_ = w.repo.SetCursor(ctx, cursorID, start)
		w.log.Info("watcher started", "cursor", start, "head", head)
	} else {
		w.log.Info("watcher started", "cursor", cur, "head", head)
	}

	ticker := time.NewTicker(w.cfg.PollInterval)
	defer ticker.Stop()

	for {
		if err := w.scanOnce(ctx); err != nil {
			w.log.Warn("watcher: scan error", "err", err)
		}
		select {
		case <-ctx.Done():
			w.log.Info("watcher stopped")
			return
		case <-ticker.C:
		}
	}
}

// scanOnce advances the cursor up to (head - confirmations), chunking the range
// to respect RPC limits, and credits any matching deposits.
func (w *Watcher) scanOnce(ctx context.Context) error {
	head, err := w.rpc.BlockNumber(ctx)
	if err != nil {
		return err
	}
	safe := head - w.cfg.Confirmations
	if safe <= 0 {
		return nil
	}

	cur, _, err := w.repo.GetCursor(ctx, cursorID)
	if err != nil {
		return err
	}
	if cur >= safe {
		return nil // nothing new to confirm yet
	}

	pending, err := w.repo.ListPendingDeposits(ctx)
	if err != nil {
		return err
	}

	// Index pending deposits by lower-cased address for O(1) matching.
	byAddr := make(map[string]PendingDeposit, len(pending))
	addrs := make([]string, 0, len(pending))
	for _, p := range pending {
		la := strings.ToLower(p.Address)
		byAddr[la] = p
		addrs = append(addrs, p.Address)
	}

	// With no open deposits there's nothing to match — just advance the cursor.
	if len(addrs) == 0 {
		return w.repo.SetCursor(ctx, cursorID, safe)
	}

	from := cur + 1
	for from <= safe {
		to := from + w.cfg.MaxBlockRange - 1
		if to > safe {
			to = safe
		}

		logs, err := w.rpc.FilterTransfers(ctx, from, to, w.cfg.Tokens, addrs)
		if err != nil {
			return err // leave cursor where it is; retry next tick
		}

		for _, lg := range logs {
			w.handleLog(ctx, lg, byAddr)
		}

		if err := w.repo.SetCursor(ctx, cursorID, to); err != nil {
			return err
		}
		from = to + 1
	}
	return nil
}

func (w *Watcher) handleLog(ctx context.Context, lg Log, byAddr map[string]PendingDeposit) {
	if len(lg.Topics) < 3 {
		return
	}
	toAddr := strings.ToLower(topicToAddress(lg.Topics[2]))
	dep, ok := byAddr[toAddr]
	if !ok {
		return // not one of our deposits
	}
	// Token must match the payment's expected token.
	if !strings.EqualFold(lg.Address, dep.Token) {
		w.log.Warn("watcher: transfer to deposit with wrong token",
			"addr", toAddr, "got", lg.Address, "want", dep.Token)
		return
	}
	// Idempotency: already confirmed.
	if w.store.Has(toAddr) {
		return
	}

	value := hexToBig(lg.Data)
	if value.Cmp(dep.Expected) < 0 {
		w.log.Warn("watcher: underpayment ignored",
			"payment", dep.PaymentID, "got", value.String(), "want", dep.Expected.String())
		return
	}
	if value.Cmp(dep.Expected) > 0 {
		w.log.Info("watcher: overpayment — crediting expected amount only",
			"payment", dep.PaymentID, "got", value.String(), "want", dep.Expected.String())
	}

	blockNum, _ := hexToInt64(lg.BlockNumber)
	w.store.Put(toAddr, Confirmation{TxHash: lg.TxHash, BlockNumber: blockNum, Amount: value})

	w.log.Info("watcher: deposit confirmed",
		"payment", dep.PaymentID, "tx", lg.TxHash, "block", blockNum, "amount", value.String())

	if err := w.onConfirm(ctx, dep.PaymentID); err != nil {
		w.log.Error("watcher: confirm callback failed", "payment", dep.PaymentID, "err", err)
	}
}
