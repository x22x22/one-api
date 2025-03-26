package completion

type ChatCompletionRequest struct {
	MaxTokens   *int      `json:"max_tokens,omitempty"`
	Messages    []Message `json:"messages"`
	Stream      *bool     `json:"stream,omitempty"`
	Model       *string   `json:"model,omitempty"`
	N           *int      `json:"n,omitempty"`
	Stop        *string   `json:"stop,omitempty"`
	Temperature *float64  `json:"temperature,omitempty"`
}
