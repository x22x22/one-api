package coze

// V3 API 数据模型

// EnterMessage 定义v3 API的输入消息结构
type EnterMessage struct {
	Role        string `json:"role"`
	Type        string `json:"type,omitempty"`
	Content     string `json:"content"`
	ContentType string `json:"content_type"`
}

// RequestV3 定义v3 API的请求结构
type RequestV3 struct {
	BotId              string            `json:"bot_id"`
	UserId             string            `json:"user_id"`
	AdditionalMessages []EnterMessage    `json:"additional_messages,omitempty"`
	Stream             bool              `json:"stream"`
	CustomVariables    map[string]string `json:"custom_variables,omitempty"`
	AutoSaveHistory    bool              `json:"auto_save_history"`
	MetaData           map[string]string `json:"meta_data,omitempty"`
	ExtraParams        map[string]string `json:"extra_params,omitempty"`
}

// ChatObject 定义聊天对象结构
type ChatObject struct {
	Id             string            `json:"id"`
	ConversationId string            `json:"conversation_id"`
	BotId          string            `json:"bot_id"`
	CreatedAt      int64             `json:"created_at,omitempty"`
	CompletedAt    int64             `json:"completed_at,omitempty"`
	FailedAt       int64             `json:"failed_at,omitempty"`
	MetaData       map[string]string `json:"meta_data,omitempty"`
	LastError      *struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	} `json:"last_error,omitempty"`
	Status         string      `json:"status"`
	RequiredAction interface{} `json:"required_action,omitempty"`
	Usage          *struct {
		TokenCount  int `json:"token_count"`
		OutputCount int `json:"output_count"`
		InputCount  int `json:"input_count"`
	} `json:"usage,omitempty"`
}

// ResponseV3 定义v3 API非流式响应结构
type ResponseV3 struct {
	Data ChatObject `json:"data"`
	Code int        `json:"code"`
	Msg  string     `json:"msg"`
}

// MessageObject 定义消息对象结构
type MessageObject struct {
	Id               string            `json:"id"`
	ConversationId   string            `json:"conversation_id"`
	BotId            string            `json:"bot_id,omitempty"`
	ChatId           string            `json:"chat_id,omitempty"`
	MetaData         map[string]string `json:"meta_data,omitempty"`
	Role             string            `json:"role"`
	Content          string            `json:"content"`
	ContentType      string            `json:"content_type"`
	CreatedAt        int64             `json:"created_at,omitempty"`
	UpdatedAt        int64             `json:"updated_at,omitempty"`
	Type             string            `json:"type"`
	SectionId        string            `json:"section_id,omitempty"`
	ReasoningContent string            `json:"reasoning_content,omitempty"`
}

// StreamResponseV3 定义v3 API流式响应结构
type StreamResponseV3 struct {
	Event string      `json:"event"`
	Data  interface{} `json:"data"`
}
