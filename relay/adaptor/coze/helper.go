package coze

import "github.com/songquanpeng/one-api/relay/adaptor/coze/constant/event"

func event2StopReason(e *string) string {
	if e == nil || *e == event.Message {
		return ""
	}
	return "stop"
}

// V3 API事件转换为停止原因
func v3Event2StopReason(eventName string) string {
	if eventName == "conversation.chat.completed" ||
		eventName == "conversation.message.completed" {
		return "stop"
	}
	return ""
}
