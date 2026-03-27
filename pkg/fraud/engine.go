package fraud

import (
	"time"

	"github.com/shopspring/decimal"
)

// RiskLevel represents the fraud risk assessment.
type RiskLevel string

const (
	RiskLow      RiskLevel = "LOW"
	RiskMedium   RiskLevel = "MEDIUM"
	RiskHigh     RiskLevel = "HIGH"
	RiskCritical RiskLevel = "CRITICAL"
)

// RiskAssessment contains the result of a fraud check.
type RiskAssessment struct {
	Score       int            `json:"score"`       // 0-100, higher = riskier
	Level       RiskLevel      `json:"level"`       // LOW, MEDIUM, HIGH, CRITICAL
	Flags       []string       `json:"flags"`       // Human-readable risk indicators
	ShouldBlock bool           `json:"shouldBlock"` // Whether to block the transaction
	CheckedAt   time.Time      `json:"checkedAt"`
}

// PaymentContext provides context about a payment for fraud analysis.
type PaymentContext struct {
	MerchantID          string
	Amount              decimal.Decimal
	Currency            string
	CustomerEmail       string
	CustomerIP          string
	Provider            string
	// Historical data
	PaymentsLast1Hour   int
	PaymentsLast24Hours int
	TotalAmountLast24H  decimal.Decimal
	AverageAmount       decimal.Decimal
	IsNewCustomer       bool
}

// Rule defines a single fraud detection rule.
type Rule struct {
	Name    string
	Check   func(ctx PaymentContext) (score int, flag string)
	Enabled bool
}

// Engine is the fraud detection engine that runs rules against payment context.
type Engine struct {
	rules      []Rule
	blockScore int // Score threshold to block transactions
}

// NewEngine creates a fraud detection engine with default rules.
func NewEngine() *Engine {
	e := &Engine{
		blockScore: 80,
	}
	e.rules = defaultRules()
	return e
}

// SetBlockThreshold sets the score threshold above which transactions are blocked.
func (e *Engine) SetBlockThreshold(score int) {
	e.blockScore = score
}

// Assess evaluates a payment context against all enabled rules and returns a risk assessment.
func (e *Engine) Assess(ctx PaymentContext) RiskAssessment {
	totalScore := 0
	var flags []string

	for _, rule := range e.rules {
		if !rule.Enabled {
			continue
		}
		score, flag := rule.Check(ctx)
		if score > 0 {
			totalScore += score
			if flag != "" {
				flags = append(flags, flag)
			}
		}
	}

	// Cap score at 100
	if totalScore > 100 {
		totalScore = 100
	}

	level := RiskLow
	switch {
	case totalScore >= 80:
		level = RiskCritical
	case totalScore >= 60:
		level = RiskHigh
	case totalScore >= 30:
		level = RiskMedium
	}

	return RiskAssessment{
		Score:       totalScore,
		Level:       level,
		Flags:       flags,
		ShouldBlock: totalScore >= e.blockScore,
		CheckedAt:   time.Now().UTC(),
	}
}

func defaultRules() []Rule {
	return []Rule{
		{
			Name:    "high_velocity_1h",
			Enabled: true,
			Check: func(ctx PaymentContext) (int, string) {
				if ctx.PaymentsLast1Hour >= 10 {
					return 30, "High payment velocity: 10+ payments in last hour"
				}
				if ctx.PaymentsLast1Hour >= 5 {
					return 15, "Elevated payment velocity: 5+ payments in last hour"
				}
				return 0, ""
			},
		},
		{
			Name:    "high_velocity_24h",
			Enabled: true,
			Check: func(ctx PaymentContext) (int, string) {
				if ctx.PaymentsLast24Hours >= 50 {
					return 25, "Excessive daily payment count: 50+ payments in 24 hours"
				}
				if ctx.PaymentsLast24Hours >= 20 {
					return 10, "High daily payment count: 20+ payments in 24 hours"
				}
				return 0, ""
			},
		},
		{
			Name:    "amount_anomaly",
			Enabled: true,
			Check: func(ctx PaymentContext) (int, string) {
				if ctx.AverageAmount.IsZero() {
					return 0, ""
				}
				// Flag if amount is 5x the average
				threshold := ctx.AverageAmount.Mul(decimal.NewFromInt(5))
				if ctx.Amount.GreaterThan(threshold) {
					return 25, "Amount significantly exceeds merchant average (5x+)"
				}
				// Flag if amount is 3x the average
				threshold3x := ctx.AverageAmount.Mul(decimal.NewFromInt(3))
				if ctx.Amount.GreaterThan(threshold3x) {
					return 10, "Amount above merchant average (3x+)"
				}
				return 0, ""
			},
		},
		{
			Name:    "large_amount",
			Enabled: true,
			Check: func(ctx PaymentContext) (int, string) {
				// Flag very large single transactions (> 10,000 USDT equivalent)
				if ctx.Amount.GreaterThan(decimal.NewFromInt(10000)) {
					return 20, "Large transaction amount (>10,000)"
				}
				if ctx.Amount.GreaterThan(decimal.NewFromInt(5000)) {
					return 10, "Substantial transaction amount (>5,000)"
				}
				return 0, ""
			},
		},
		{
			Name:    "daily_volume",
			Enabled: true,
			Check: func(ctx PaymentContext) (int, string) {
				// Flag if 24h volume is very high
				if ctx.TotalAmountLast24H.GreaterThan(decimal.NewFromInt(50000)) {
					return 20, "High daily volume (>50,000 in 24h)"
				}
				return 0, ""
			},
		},
		{
			Name:    "new_customer_large_amount",
			Enabled: true,
			Check: func(ctx PaymentContext) (int, string) {
				if ctx.IsNewCustomer && ctx.Amount.GreaterThan(decimal.NewFromInt(1000)) {
					return 15, "Large payment from new customer"
				}
				return 0, ""
			},
		},
	}
}
