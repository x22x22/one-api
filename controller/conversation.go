package controller

import (
	"fmt"
	"github.com/google/uuid"
	"strings"
)

type ContinueInfo struct {
	ConversationID string `json:"conversation_id"`
	ParentID       string `json:"parent_id"`
}

type APIRequest struct {
	Messages  []ApiMessage `json:"messages"`
	Stream    bool         `json:"stream"`
	Model     string       `json:"model"`
	PluginIDs []string     `json:"plugin_ids"`
}

type ApiMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type CreateConversationRequest struct {
	Action                     string                `json:"action"`
	Messages                   []ConversationMessage `json:"messages"`
	Model                      string                `json:"model"`
	ParentMessageID            string                `json:"parent_message_id"`
	ConversationID             *string               `json:"conversation_id"`
	PluginIDs                  []string              `json:"plugin_ids"`
	TimezoneOffsetMin          int                   `json:"timezone_offset_min"`
	ArkoseToken                string                `json:"arkose_token"`
	HistoryAndTrainingDisabled bool                  `json:"history_and_training_disabled"`
	AutoContinue               bool                  `json:"auto_continue"`
	Suggestions                []string              `json:"suggestions"`
}

func (c *CreateConversationRequest) AddMessage(role string, content string) {
	c.Messages = append(c.Messages, ConversationMessage{
		ID:       uuid.New().String(),
		Author:   Author{Role: role},
		Content:  Content{ContentType: "text", Parts: []interface{}{content}},
		Metadata: map[string]string{},
	})
}

type ConversationMessage struct {
	Author   Author      `json:"author"`
	Content  Content     `json:"content"`
	ID       string      `json:"id"`
	Metadata interface{} `json:"metadata"`
}

type Author struct {
	Role string `json:"role"`
}

type Content struct {
	ContentType string        `json:"content_type"`
	Parts       []interface{} `json:"parts"`
}

type CreateConversationResponse struct {
	Message struct {
		ID     string `json:"id"`
		Author struct {
			Role     string      `json:"role"`
			Name     interface{} `json:"name"`
			Metadata struct {
			} `json:"metadata"`
		} `json:"author"`
		CreateTime float64     `json:"create_time"`
		UpdateTime interface{} `json:"update_time"`
		Content    struct {
			ContentType string   `json:"content_type"`
			Parts       []string `json:"parts"`
		} `json:"content"`
		Status   string  `json:"status"`
		EndTurn  bool    `json:"end_turn"`
		Weight   float64 `json:"weight"`
		Metadata struct {
			MessageType   string `json:"message_type"`
			ModelSlug     string `json:"model_slug"`
			FinishDetails struct {
				Type string `json:"type"`
			} `json:"finish_details"`
		} `json:"metadata"`
		Recipient string `json:"recipient"`
	} `json:"message"`
	ConversationID string      `json:"conversation_id"`
	Error          interface{} `json:"error"`
}

type GetModelsResponse struct {
	Models []struct {
		Slug         string   `json:"slug"`
		MaxTokens    int      `json:"max_tokens"`
		Title        string   `json:"title"`
		Description  string   `json:"description"`
		Tags         []string `json:"tags"`
		Capabilities struct {
		} `json:"capabilities"`
		EnabledTools []string `json:"enabled_tools,omitempty"`
	} `json:"models"`
	Categories []struct {
		Category             string `json:"category"`
		HumanCategoryName    string `json:"human_category_name"`
		SubscriptionLevel    string `json:"subscription_level"`
		DefaultModel         string `json:"default_model"`
		CodeInterpreterModel string `json:"code_interpreter_model"`
		PluginsModel         string `json:"plugins_model"`
	} `json:"categories"`
}

func ConvertToString(chatgptResponse *ChatGPTResponse, previousText *StringStruct, role bool, id string, model string) string {
	var text string

	if len(chatgptResponse.Message.Content.Parts) == 1 {
		if part, ok := chatgptResponse.Message.Content.Parts[0].(string); ok {
			text = strings.ReplaceAll(part, previousText.Text, "")
			previousText.Text = part
		} else {
			text = fmt.Sprintf("%v", chatgptResponse.Message.Content.Parts[0])
		}
	} else {
		// When using GPT-4 messages with images (multimodal_text), the length of 'parts' might be 2.
		// Since the chatgpt API currently does not support multimodal content
		// and there is no official format for multimodal content,
		// the content is temporarily returned as is.
		var parts []string
		for _, part := range chatgptResponse.Message.Content.Parts {
			parts = append(parts, fmt.Sprintf("%v", part))
		}
		text = strings.Join(parts, ", ")
	}

	translatedResponse := NewChatCompletionChunk(text, id, model)
	if role {
		translatedResponse.Choices[0].Delta.Role = chatgptResponse.Message.Author.Role
	}

	return "data: " + translatedResponse.String() + "\n\n"
}
