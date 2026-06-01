package postgres

import (
	"context"
	"errors"
	"fmt"
	"math/big"

	"github.com/jackc/pgx/v5"

	"github.com/openlankapay/openlankapay/services/payment/internal/watcher"
)

// NextDepositIndex reserves the next monotonic HD-wallet deposit index.
func (r *PaymentRepository) NextDepositIndex(ctx context.Context) (int64, error) {
	var idx int64
	if err := r.pool.QueryRow(ctx, `SELECT nextval('payment_deposit_index_seq')`).Scan(&idx); err != nil {
		return 0, fmt.Errorf("allocating deposit index: %w", err)
	}
	return idx, nil
}

// ListPendingDeposits returns open on-chain payments awaiting an inbound
// transfer (INITIATED, not expired). Token addresses are resolved by the caller
// via the currency→token map; here we expose the stored token by currency.
func (r *PaymentRepository) ListPendingDeposits(ctx context.Context) ([]watcher.PendingDeposit, error) {
	// We need the deposit address, the expected base-unit amount and the token.
	// The expected amount is amount_usdt scaled to the token's decimals; both
	// supported tokens use 6 decimals, so we scale by 1e6 here.
	const q = `
		SELECT id, wallet_address, currency, amount_usdt
		FROM payments
		WHERE provider = 'ONCHAIN'
		  AND status = 'INITIATED'
		  AND wallet_address <> ''
		  AND expire_time > NOW()
		  AND deleted_at IS NULL`

	rows, err := r.pool.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("listing pending deposits: %w", err)
	}
	defer rows.Close()

	var out []watcher.PendingDeposit
	for rows.Next() {
		var (
			pd        watcher.PendingDeposit
			currency  string
			amountStr string
		)
		if err := rows.Scan(&pd.PaymentID, &pd.Address, &currency, &amountStr); err != nil {
			return nil, fmt.Errorf("scanning pending deposit: %w", err)
		}
		// Resolve the token contract address from the payment's currency.
		pd.Token = r.tokenByCurrency(currency)
		pd.Expected = scaleToBaseUnits(amountStr, 6)
		out = append(out, pd)
	}
	return out, rows.Err()
}

// SetTokenResolver configures how currencies map to token contract addresses.
func (r *PaymentRepository) SetTokenResolver(f func(currency string) string) {
	r.tokenResolver = f
}

func (r *PaymentRepository) tokenByCurrency(currency string) string {
	if r.tokenResolver == nil {
		return ""
	}
	return r.tokenResolver(currency)
}

// GetCursor returns the last scanned block for a watcher id.
func (r *PaymentRepository) GetCursor(ctx context.Context, id string) (int64, bool, error) {
	var block int64
	err := r.pool.QueryRow(ctx, `SELECT last_scanned_block FROM watcher_cursor WHERE id = $1`, id).Scan(&block)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, fmt.Errorf("reading cursor: %w", err)
	}
	return block, true, nil
}

// SetCursor upserts the last scanned block for a watcher id.
func (r *PaymentRepository) SetCursor(ctx context.Context, id string, block int64) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO watcher_cursor (id, last_scanned_block, updated_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (id) DO UPDATE SET last_scanned_block = EXCLUDED.last_scanned_block, updated_at = NOW()`,
		id, block)
	if err != nil {
		return fmt.Errorf("writing cursor: %w", err)
	}
	return nil
}

// scaleToBaseUnits converts a decimal string (e.g. "5" or "5.5") to integer base
// units at the given number of decimals.
func scaleToBaseUnits(amount string, decimals int) *big.Int {
	// Parse via big.Rat to avoid float error, then scale.
	rat, ok := new(big.Rat).SetString(amount)
	if !ok {
		return big.NewInt(0)
	}
	scale := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
	rat.Mul(rat, new(big.Rat).SetInt(scale))
	// rat is now an integer value (assuming <= decimals places); take numerator/denominator.
	res := new(big.Int).Quo(rat.Num(), rat.Denom())
	return res
}

// nilIfZeroIndex returns nil for a zero deposit index so non-on-chain payments
// store NULL rather than 0.
func nilIfZeroIndex(idx int64) *int64 {
	if idx == 0 {
		return nil
	}
	return &idx
}
