package watcher

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"time"
)

// Transfer(address,address,uint256) event signature hash (topic0).
const transferTopic0 = "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"

// RPCClient is a minimal Ethereum JSON-RPC client (only the methods the watcher
// needs). Avoids pulling in the full go-ethereum dependency tree.
type RPCClient struct {
	url    string
	http   *http.Client
	nextID int
}

func NewRPCClient(url string) *RPCClient {
	return &RPCClient{
		url:  url,
		http: &http.Client{Timeout: 20 * time.Second},
	}
}

type rpcRequest struct {
	JSONRPC string `json:"jsonrpc"`
	ID      int    `json:"id"`
	Method  string `json:"method"`
	Params  []any  `json:"params"`
}

type rpcResponse struct {
	Result json.RawMessage `json:"result"`
	Error  *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func (c *RPCClient) call(ctx context.Context, method string, params []any, out any) error {
	c.nextID++
	body, err := json.Marshal(rpcRequest{JSONRPC: "2.0", ID: c.nextID, Method: method, Params: params})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("rpc %s: %w", method, err)
	}
	defer resp.Body.Close()

	var r rpcResponse
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return fmt.Errorf("rpc %s decode: %w", method, err)
	}
	if r.Error != nil {
		return fmt.Errorf("rpc %s: %s (code %d)", method, r.Error.Message, r.Error.Code)
	}
	return json.Unmarshal(r.Result, out)
}

// BlockNumber returns the latest block height.
func (c *RPCClient) BlockNumber(ctx context.Context) (int64, error) {
	var hex string
	if err := c.call(ctx, "eth_blockNumber", []any{}, &hex); err != nil {
		return 0, err
	}
	return hexToInt64(hex)
}

// Log is a decoded eth_getLogs entry.
type Log struct {
	Address     string   `json:"address"`
	Topics      []string `json:"topics"`
	Data        string   `json:"data"`
	BlockNumber string   `json:"blockNumber"`
	TxHash      string   `json:"transactionHash"`
}

// FilterTransfers returns ERC20 Transfer logs for the given token addresses,
// restricted (by topic) to transfers whose `to` is one of toAddrs.
func (c *RPCClient) FilterTransfers(ctx context.Context, fromBlock, toBlock int64, tokens, toAddrs []string) ([]Log, error) {
	// topics: [Transfer, anyFrom, [toAddrs...]] — topic2 is the indexed `to`.
	toTopics := make([]string, 0, len(toAddrs))
	for _, a := range toAddrs {
		toTopics = append(toTopics, addressToTopic(a))
	}

	params := map[string]any{
		"fromBlock": int64ToHex(fromBlock),
		"toBlock":   int64ToHex(toBlock),
		"address":   tokens,
		"topics":    []any{transferTopic0, nil, toTopics},
	}

	var logs []Log
	if err := c.call(ctx, "eth_getLogs", []any{params}, &logs); err != nil {
		return nil, err
	}
	return logs, nil
}

// --- hex helpers ---

func hexToInt64(h string) (int64, error) {
	n, ok := new(big.Int).SetString(trim0x(h), 16)
	if !ok {
		return 0, fmt.Errorf("invalid hex int %q", h)
	}
	return n.Int64(), nil
}

// hexToBig parses a 0x hex string (e.g. log data) into a big.Int.
func hexToBig(h string) *big.Int {
	n, _ := new(big.Int).SetString(trim0x(h), 16)
	if n == nil {
		return big.NewInt(0)
	}
	return n
}

func int64ToHex(n int64) string {
	return "0x" + new(big.Int).SetInt64(n).Text(16)
}

func trim0x(h string) string {
	if len(h) >= 2 && (h[:2] == "0x" || h[:2] == "0X") {
		return h[2:]
	}
	return h
}

// addressToTopic left-pads a 20-byte address to a 32-byte topic.
func addressToTopic(addr string) string {
	a := trim0x(addr)
	return "0x" + leftPad(a, 64)
}

func leftPad(s string, width int) string {
	for len(s) < width {
		s = "0" + s
	}
	return s
}

// topicToAddress extracts the 20-byte address from a 32-byte indexed topic.
func topicToAddress(topic string) string {
	t := trim0x(topic)
	if len(t) < 40 {
		return ""
	}
	return "0x" + t[len(t)-40:]
}
