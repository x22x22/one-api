package completion_stream

type ChatCompletionChunk struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
}

type Choice struct {
	Index int   `json:"index"`
	Delta Delta `json:"delta"`
}

type Delta struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
