package types

type StreamChunk struct {
	Content string  `json:"content"`
	Done    bool    `json:"done"`
	Error   error   `json:"-"`
	Usage   Usage   `json:"usage,omitempty"`
	CostUSD float64 `json:"cost_usd,omitempty"`
}
