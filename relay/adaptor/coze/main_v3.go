package coze

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/common/conv"
	"github.com/songquanpeng/one-api/common/helper"
	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/common/render"
	"github.com/songquanpeng/one-api/relay/adaptor/coze/constant/messagetype"
	"github.com/songquanpeng/one-api/relay/adaptor/openai"
	"github.com/songquanpeng/one-api/relay/model"
)

// ConvertRequestV3 将OpenAI请求转换为Coze V3 API请求
func ConvertRequestV3(textRequest model.GeneralOpenAIRequest) *RequestV3 {
	cozeRequest := RequestV3{
		Stream:          textRequest.Stream,
		UserId:          textRequest.User,
		BotId:           strings.TrimPrefix(textRequest.Model, "bot-"),
		AutoSaveHistory: true,
	}

	// 注意：会话ID通过meta.Config.ConversationID获取，并在GetRequestURL方法中添加到URL参数

	var systemContent string
	var additionalMessages []EnterMessage

	for i, message := range textRequest.Messages {
		if message.Role == "system" {
			systemContent = message.StringContent()
			continue
		}

		// 将最后一条消息作为用户查询
		if i == len(textRequest.Messages)-1 {
			content := message.StringContent()
			// 如果有系统提示，将其添加到用户消息中
			if systemContent != "" {
				content = fmt.Sprintf("<s>\n%s\n</s>\n%s", systemContent, content)
			}

			// 添加用户消息
			userMessage := EnterMessage{
				Role:        "user",
				Content:     content,
				ContentType: "text",
			}
			additionalMessages = append(additionalMessages, userMessage)
			continue
		}

		// 处理历史消息
		cozeMessage := EnterMessage{
			Role:        message.Role,
			Content:     message.StringContent(),
			ContentType: "text",
		}
		additionalMessages = append(additionalMessages, cozeMessage)
	}

	cozeRequest.AdditionalMessages = additionalMessages
	return &cozeRequest
}

// StreamResponseV3ToOpenAI 将Coze V3流式响应转换为OpenAI格式
func StreamResponseV3ToOpenAI(event string, data []byte) (*openai.ChatCompletionsStreamResponse, error) {
	// 只处理消息事件
	if strings.HasPrefix(event, "conversation.message.") {
		var messageData MessageObject
		err := json.Unmarshal(data, &messageData)
		if err != nil {
			return nil, err
		}

		// 只处理answer类型的消息
		if messageData.Type != messagetype.Answer {
			return nil, nil
		}

		var choice openai.ChatCompletionsStreamResponseChoice
		choice.Delta.Content = messageData.Content
		choice.Delta.ReasoningContent = messageData.ReasoningContent
		choice.Delta.Role = "assistant"

		// 如果是已完成的消息，设置完成原因
		if event == "conversation.message.completed" {
			finishReason := "stop"
			choice.FinishReason = &finishReason
		}

		var openaiResponse openai.ChatCompletionsStreamResponse
		openaiResponse.Object = "chat.completion.chunk"
		openaiResponse.Choices = []openai.ChatCompletionsStreamResponseChoice{choice}
		openaiResponse.Id = messageData.ConversationId
		return &openaiResponse, nil
	}

	return nil, nil
}

// ResponseV3ToOpenAI 将Coze V3非流式响应转换为OpenAI格式
func ResponseV3ToOpenAI(cozeResponse *ResponseV3) *openai.TextResponse {
	choice := openai.TextResponseChoice{
		Index: 0,
		Message: model.Message{
			Role:    "assistant",
			Content: "", // 在非流式响应中，需要单独查询消息详情来获取内容
			Name:    nil,
		},
		FinishReason: "stop",
	}

	fullTextResponse := openai.TextResponse{
		Id:      fmt.Sprintf("chatcmpl-%s", cozeResponse.Data.Id),
		Model:   "coze-bot",
		Object:  "chat.completion",
		Created: helper.GetTimestamp(),
		Choices: []openai.TextResponseChoice{choice},
	}
	return &fullTextResponse
}

// StreamHandlerV3 处理V3 API流式响应
func StreamHandlerV3(c *gin.Context, resp *http.Response) (*model.ErrorWithStatusCode, *string) {
	var responseText string
	createdTime := helper.GetTimestamp()
	scanner := bufio.NewScanner(resp.Body)
	scanner.Split(bufio.ScanLines)

	common.SetEventStreamHeaders(c)
	var modelName string

	for scanner.Scan() {
		line := scanner.Text()
		if len(line) < 5 {
			continue
		}

		// 解析事件和数据
		var event, data string
		if strings.HasPrefix(line, "event:") {
			event = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
			// 下一行应该是数据
			if scanner.Scan() {
				dataLine := scanner.Text()
				if strings.HasPrefix(dataLine, "data:") {
					data = strings.TrimSpace(strings.TrimPrefix(dataLine, "data:"))
				}
			}
		} else if strings.HasPrefix(line, "data:") {
			data = strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		}

		if event == "done" {
			// 流结束
			break
		}

		if data == "" || event == "" {
			continue
		}

		// 转换事件
		response, err := StreamResponseV3ToOpenAI(event, []byte(data))
		if err != nil || response == nil {
			continue
		}

		for _, choice := range response.Choices {
			responseText += conv.AsString(choice.Delta.Content)
		}
		response.Model = modelName
		response.Created = createdTime

		err = render.ObjectData(c, response)
		if err != nil {
			logger.SysError(err.Error())
		}
	}

	if err := scanner.Err(); err != nil {
		logger.SysError("error reading stream: " + err.Error())
	}

	render.Done(c)

	err := resp.Body.Close()
	if err != nil {
		return openai.ErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}

	return nil, &responseText
}

// HandlerV3 处理V3 API非流式响应
func HandlerV3(c *gin.Context, resp *http.Response, promptTokens int, modelName string) (*model.ErrorWithStatusCode, *string) {
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return openai.ErrorWrapper(err, "read_response_body_failed", http.StatusInternalServerError), nil
	}
	err = resp.Body.Close()
	if err != nil {
		return openai.ErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}

	var cozeResponse ResponseV3
	err = json.Unmarshal(responseBody, &cozeResponse)
	if err != nil {
		return openai.ErrorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError), nil
	}

	if cozeResponse.Code != 0 {
		return &model.ErrorWithStatusCode{
			Error: model.Error{
				Message: cozeResponse.Msg,
				Code:    cozeResponse.Code,
			},
			StatusCode: resp.StatusCode,
		}, nil
	}

	// 对于V3 API，非流式响应需要额外调用获取消息详情的API
	// 这里简化处理，直接返回空内容，实际应该调用查看消息详情接口
	// TODO: 对于非流式响应，应该实现调用获取消息详情的API，
	// 并处理可能存在的reasoning_content字段（DeepSeek-R1模型的思维链）
	logger.SysLog(fmt.Sprintf("V3 API非流式响应: 要获取完整消息内容，需要调用查看消息详情接口，chatId=%s", cozeResponse.Data.Id))

	fullTextResponse := ResponseV3ToOpenAI(&cozeResponse)
	fullTextResponse.Model = modelName

	jsonResponse, err := json.Marshal(fullTextResponse)
	if err != nil {
		return openai.ErrorWrapper(err, "marshal_response_body_failed", http.StatusInternalServerError), nil
	}

	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(resp.StatusCode)
	_, err = c.Writer.Write(jsonResponse)

	// 这里简化处理返回空响应
	var responseText string
	if len(fullTextResponse.Choices) > 0 {
		responseText = fullTextResponse.Choices[0].Message.StringContent()
	}

	return nil, &responseText
}
