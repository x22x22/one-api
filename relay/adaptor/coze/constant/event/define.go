package event

// 通用事件
const (
	Message = "message"
	Done    = "done"
	Error   = "error"
)

// V3 事件
const (
	ChatCreated    = "conversation.chat.created"
	ChatInProgress = "conversation.chat.in_progress"
	ChatCompleted  = "conversation.chat.completed"
	ChatFailed     = "conversation.chat.failed"

	MessageDelta     = "conversation.message.delta"
	MessageCompleted = "conversation.message.completed"
)
