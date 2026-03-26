// CLI tool for interacting with the Open Pay API.
//
// Usage:
//
//	openpay config set-key <api-key>
//	openpay config set-url <base-url>
//	openpay payments create --amount 100.00 --currency LKR
//	openpay payments list
//	openpay payments get <id>
//	openpay apikeys create --name "Production"
//	openpay apikeys list
//	openpay webhooks configure --url https://example.com/webhook
//	openpay webhooks test
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"
	"time"

	openpay "github.com/openlankapay/openlankapay/sdks/sdk-go"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "config":
		handleConfig()
	case "payments":
		handlePayments()
	case "checkout":
		handleCheckout()
	case "apikeys":
		handleAPIKeys()
	case "webhooks":
		handleWebhooks()
	case "version":
		fmt.Println("openpay v0.1.0")
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`Open Pay CLI — manage payments from your terminal

Usage:
  openpay <command> [subcommand] [flags]

Commands:
  config      Manage CLI configuration
  payments    Create and manage payments
  apikeys     Manage API keys (requires JWT auth)
  webhooks    Configure and test webhooks
  version     Show CLI version

Configuration:
  openpay config set-key <api-key>    Set your API key
  openpay config set-url <base-url>   Set API base URL
  openpay config show                 Show current config

Payments:
  openpay payments create --amount <amt> [--currency LKR] [--trade-no ORDER-1]
  openpay payments list [--status PAID] [--page 1]
  openpay payments get <payment-id>

Webhooks:
  openpay webhooks configure --url <webhook-url>
  openpay webhooks public-key
  openpay webhooks test

Environment:
  OPENPAY_API_KEY     API key (overrides config file)
  OPENPAY_BASE_URL    Base URL (overrides config file)`)
}

// ─── Config ───

type cliConfig struct {
	APIKey  string `json:"api_key"`
	BaseURL string `json:"base_url"`
}

func configPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".openpay", "config.json")
}

func loadConfig() cliConfig {
	var cfg cliConfig
	data, err := os.ReadFile(configPath())
	if err == nil {
		_ = json.Unmarshal(data, &cfg)
	}
	// Env overrides
	if v := os.Getenv("OPENPAY_API_KEY"); v != "" {
		cfg.APIKey = v
	}
	if v := os.Getenv("OPENPAY_BASE_URL"); v != "" {
		cfg.BaseURL = v
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api.openpay.lk"
	}
	return cfg
}

func saveConfig(cfg cliConfig) error {
	dir := filepath.Dir(configPath())
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	data, _ := json.MarshalIndent(cfg, "", "  ")
	return os.WriteFile(configPath(), data, 0600)
}

func handleConfig() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: openpay config <set-key|set-url|show>")
		return
	}
	switch os.Args[2] {
	case "set-key":
		if len(os.Args) < 4 {
			fatal("Usage: openpay config set-key <api-key>")
		}
		cfg := loadConfig()
		cfg.APIKey = os.Args[3]
		if err := saveConfig(cfg); err != nil {
			fatal("Failed to save config: %v", err)
		}
		fmt.Println("API key saved.")
	case "set-url":
		if len(os.Args) < 4 {
			fatal("Usage: openpay config set-url <base-url>")
		}
		cfg := loadConfig()
		cfg.BaseURL = os.Args[3]
		if err := saveConfig(cfg); err != nil {
			fatal("Failed to save config: %v", err)
		}
		fmt.Printf("Base URL set to: %s\n", os.Args[3])
	case "show":
		cfg := loadConfig()
		masked := cfg.APIKey
		if len(masked) > 20 {
			masked = masked[:20] + "..."
		}
		fmt.Printf("API Key:  %s\nBase URL: %s\nConfig:   %s\n", masked, cfg.BaseURL, configPath())
	default:
		fmt.Println("Usage: openpay config <set-key|set-url|show>")
	}
}

// ─── Client helper ───

func getClient() *openpay.Client {
	cfg := loadConfig()
	if cfg.APIKey == "" {
		fatal("No API key configured. Run: openpay config set-key <your-api-key>")
	}
	client, err := openpay.NewClient(cfg.APIKey, openpay.WithBaseURL(cfg.BaseURL))
	if err != nil {
		fatal("Invalid API key: %v", err)
	}
	return client
}

// ─── Payments ───

func handlePayments() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: openpay payments <create|list|get>")
		return
	}

	client := getClient()
	ctx := context.Background()

	switch os.Args[2] {
	case "create":
		amount, currency, tradeNo, desc := "", "LKR", "", ""
		for i := 3; i < len(os.Args); i++ {
			switch os.Args[i] {
			case "--amount":
				i++; amount = os.Args[i]
			case "--currency":
				i++; currency = os.Args[i]
			case "--trade-no":
				i++; tradeNo = os.Args[i]
			case "--description":
				i++; desc = os.Args[i]
			}
		}
		if amount == "" {
			fatal("--amount is required")
		}
		if tradeNo == "" {
			tradeNo = fmt.Sprintf("CLI-%d", time.Now().UnixMilli())
		}

		payment, err := client.Payments.Create(ctx, openpay.CreatePaymentInput{
			Amount:          amount,
			Currency:        currency,
			MerchantTradeNo: tradeNo,
			Description:     desc,
		})
		if err != nil {
			fatal("Failed: %v", err)
		}
		printJSON(payment)

	case "list":
		params := &openpay.ListPaymentsParams{Page: 1, PerPage: 20}
		for i := 3; i < len(os.Args); i++ {
			switch os.Args[i] {
			case "--status":
				i++; params.Status = os.Args[i]
			case "--page":
				i++; _, _ = fmt.Sscanf(os.Args[i], "%d", &params.Page)
			case "--per-page":
				i++; _, _ = fmt.Sscanf(os.Args[i], "%d", &params.PerPage)
			}
		}

		result, err := client.Payments.List(ctx, params)
		if err != nil {
			fatal("Failed: %v", err)
		}

		if len(result.Data) == 0 {
			fmt.Println("No payments found.")
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		_, _ = fmt.Fprintln(w, "ID\tAMOUNT\tCURRENCY\tSTATUS\tCREATED")
		for _, p := range result.Data {
			_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", p.ID, p.Amount, p.Currency, p.Status, p.CreatedAt[:10])
		}
		_ = w.Flush()
		fmt.Printf("\nPage %d of %d (%d total)\n", result.Meta.Page, (result.Meta.Total+result.Meta.PerPage-1)/result.Meta.PerPage, result.Meta.Total)

	case "get":
		if len(os.Args) < 4 {
			fatal("Usage: openpay payments get <payment-id>")
		}
		payment, err := client.Payments.Get(ctx, os.Args[3])
		if err != nil {
			fatal("Failed: %v", err)
		}
		printJSON(payment)

	default:
		fmt.Println("Usage: openpay payments <create|list|get>")
	}
}

// ─── API Keys ───

func handleAPIKeys() {
	fmt.Println("API key management requires JWT auth (use the merchant portal).")
	fmt.Println("To use the CLI, set your API key: openpay config set-key <key>")
}

// ─── Checkout ───

func handleCheckout() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: openpay checkout create-session --amount <amt> [--currency LKR] [--success-url URL] [--cancel-url URL]")
		return
	}

	client := getClient()
	ctx := context.Background()

	switch os.Args[2] {
	case "create-session":
		amount, currency, successURL, cancelURL := "", "LKR", "", ""
		for i := 3; i < len(os.Args); i++ {
			switch os.Args[i] {
			case "--amount":
				i++; amount = os.Args[i]
			case "--currency":
				i++; currency = os.Args[i]
			case "--success-url":
				i++; successURL = os.Args[i]
			case "--cancel-url":
				i++; cancelURL = os.Args[i]
			}
		}
		if amount == "" {
			fatal("--amount is required")
		}

		session, err := client.Checkout.CreateSession(ctx, openpay.CheckoutSessionInput{
			Amount:    amount,
			Currency:  currency,
			SuccessURL: successURL,
			CancelURL:  cancelURL,
		})
		if err != nil {
			fatal("Failed: %v", err)
		}
		printJSON(session)
		fmt.Printf("\nCheckout URL: %s\n", session.URL)
	default:
		fmt.Println("Usage: openpay checkout create-session --amount <amt>")
	}
}

// ─── Webhooks ───

func handleWebhooks() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: openpay webhooks <configure|public-key|test>")
		return
	}

	client := getClient()
	ctx := context.Background()

	switch os.Args[2] {
	case "configure":
		url := ""
		for i := 3; i < len(os.Args); i++ {
			if os.Args[i] == "--url" {
				i++; url = os.Args[i]
			}
		}
		if url == "" {
			fatal("--url is required")
		}
		if err := client.Webhooks.Configure(ctx, openpay.WebhookConfig{URL: url}); err != nil {
			fatal("Failed: %v", err)
		}
		fmt.Printf("Webhook configured: %s\n", url)

	case "public-key":
		key, err := client.Webhooks.GetPublicKey(ctx)
		if err != nil {
			fatal("Failed: %v", err)
		}
		fmt.Println(key)

	case "test":
		if err := client.Webhooks.Test(ctx); err != nil {
			fatal("Failed: %v", err)
		}
		fmt.Println("Test webhook sent.")

	default:
		fmt.Println("Usage: openpay webhooks <configure|public-key|test>")
	}
}

// ─── Helpers ───

func printJSON(v interface{}) {
	data, _ := json.MarshalIndent(v, "", "  ")
	fmt.Println(string(data))
}

func fatal(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "Error: "+format+"\n", args...)
	os.Exit(1)
}
