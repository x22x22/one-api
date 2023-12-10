package controller

import (
	"bytes"
	"encoding/base64"
	"github.com/google/uuid"
	"io"
	"math/rand"
	"net"
	"net/http"
	"one-api/common"
	"regexp"
	"strings"
	"time"
)

var (
	reg *regexp.Regexp
)

func init() {
	reg, _ = regexp.Compile("[^a-zA-Z0-9]+")
}

func asyncHTTPDoWithOpenaiWeb(req *http.Request, isStream bool) (*http.Response, error) {
	backupBody, _ := io.ReadAll(req.Body)
	id := generateId()
	defer func(body io.ReadCloser) {
		err := body.Close()
		if err != nil {
			return
		}
	}(req.Body)

	respCh := make(chan *http.Response, 1)
	errCh := make(chan error, 1)

	go func() {
		defer func() {
			if r := recover(); r != nil {

			}
		}()

		var reqs []*http.Request

		defer func() {
			for _, req := range reqs {
				func(Body io.ReadCloser) {
					err := Body.Close()
					if err != nil {
						return
					}
				}(req.Body)
			}
		}()

		var fullResponse string
		var continueInfo *ContinueInfo = nil
		var responsePart string
		var pw *io.PipeWriter
		var model string = "gpt-web"
		defer func() {
			if pw != nil {
				_, _ = pw.Write([]byte("data: [DONE]\n\n"))
				err := pw.Close()
				if err != nil {
					return
				}
			}
		}()
		for i := 0; i < 3; i++ {
			var tmpReq *http.Request
			if continueInfo == nil {
				tmpReq = req.Clone(req.Context())
				tmpReq.Body = io.NopCloser(bytes.NewBuffer(backupBody))
			} else {
				var translatedRequest CreateConversationRequest
				_ = json.NewDecoder(bytes.NewBuffer(backupBody)).Decode(&translatedRequest)
				translatedRequest.Messages = nil
				translatedRequest.Action = "continue"
				translatedRequest.ConversationID = &continueInfo.ConversationID
				translatedRequest.ParentMessageID = continueInfo.ParentID
				jsonBytes, _ := json.Marshal(translatedRequest)
				tmpReq = req.Clone(req.Context())
				tmpReq.Body = io.NopCloser(bytes.NewBuffer(jsonBytes))
			}
			reqs = append(reqs, tmpReq)

			response, err := asyncHTTPDo(tmpReq, 1)
			if err != nil {
				errCh <- err
				return
			}

			if response.StatusCode != http.StatusOK {
				errCh <- err
				return
			}
			if i == 0 && isStream {
				var sseResp *http.Response
				sseResp, pw = createSSEResponse()
				respCh <- sseResp
			}
			responsePart, continueInfo = HandlerWithClose(pw, response, id, model)
			fullResponse += responsePart
			if continueInfo == nil {
				break
			}
		}

		if !isStream {
			resp := createSuccessResponse(newChatCompletion(fullResponse, model, id))
			respCh <- resp
		}
	}()
	for {
		select {
		case resp := <-respCh:
			return resp, nil
		case err := <-errCh:
			return nil, err
		}
	}
}

func HandlerWithClose(pw *io.PipeWriter, response *http.Response, id string, model string) (string, *ContinueInfo) {
	maxTokens := false

	eventStreamReader := NewEventStreamReader(response.Body.(io.Reader), 40960)
	defer response.Body.Close()

	var finishReason string
	var previousText StringStruct
	var originalResponse ChatGPTResponse
	var isRole = true
	for {
		data, err := eventStreamReader.ReadEvent()
		if err != nil {
			if err == io.EOF {
				break
			}
			common.SysError(err.Error())
			break
		}
		event, err := processEvent(data)
		if err != nil {
			break
		}
		// Check if line starts with [DONE]
		if !bytes.HasPrefix(event.Data, []byte("[DONE]")) {
			// Parse the line as JSON

			err = json.Unmarshal([]byte(event.Data), &originalResponse)
			if err != nil {
				continue
			}
			if originalResponse.Error != nil {
				return "", nil
			}
			if originalResponse.Message.Author.Role != "assistant" || originalResponse.Message.Content.Parts == nil {
				continue
			}
			if originalResponse.Message.Metadata.MessageType != "next" && originalResponse.Message.Metadata.MessageType != "continue" || originalResponse.Message.EndTurn != nil {
				continue
			}
			if (len(originalResponse.Message.Content.Parts) == 0 || originalResponse.Message.Content.Parts[0] == "") && !isRole {
				continue
			}
			responseString := ConvertToString(&originalResponse, &previousText, isRole, id, model)
			isRole = false
			if pw != nil {
				_, err := pw.Write([]byte(responseString + "\n\n"))
				if err != nil {
					return "", nil
				}
			}

			if originalResponse.Message.Metadata.FinishDetails != nil {
				if originalResponse.Message.Metadata.FinishDetails.Type == "max_tokens" {
					maxTokens = true
				}
				finishReason = originalResponse.Message.Metadata.FinishDetails.Type
			}

		} else {
			if finishReason == "" {
				finishReason = "stop"
			}
			finalLine := StopChunk(finishReason, id, model)
			if pw != nil {
				_, err := pw.Write([]byte("data: " + finalLine.String() + "\n\n"))
				if err != nil {
					return "", nil
				}
			}
		}
	}
	if !maxTokens {
		return previousText.Text, nil
	}
	return previousText.Text, &ContinueInfo{
		ConversationID: originalResponse.ConversationID,
		ParentID:       originalResponse.Message.ID,
	}
}

func generateId() string {
	id := uuid.NewString()
	id = strings.ReplaceAll(id, "-", "")
	id = base64.StdEncoding.EncodeToString([]byte(id))
	id = reg.ReplaceAllString(id, "")
	return "chatcmpl-" + id
}

var IPRanges = []string{
	"103.21.244.0/22",
	"103.22.200.0/22",
	"103.31.4.0/22",
	"104.16.0.0/13",
	"104.24.0.0/14",
	"108.162.192.0/18",
	"131.0.72.0/22",
	"141.101.64.0/18",
	"162.158.0.0/15",
	"172.64.0.0/13",
	"173.245.48.0/20",
	"188.114.96.0/20",
	"190.93.240.0/20",
	"197.234.240.0/22",
	"198.41.128.0/17",
}

func randomIPFromRanges(ranges []string) (net.IP, error) {
	// 随机选择一个IP段
	rand.Seed(time.Now().UnixNano())
	randomIndex := rand.Intn(len(ranges))
	selectedRange := ranges[randomIndex]

	// 解析CIDR以获取IP范围
	_, ipnet, err := net.ParseCIDR(selectedRange)
	if err != nil {
		return nil, err
	}

	// 随机生成IP地址
	randomIP := make(net.IP, len(ipnet.IP))
	for {
		copy(randomIP, ipnet.IP)
		for i := range randomIP {
			if ipnet.Mask[i] == 0xff {
				continue
			}
			randomIP[i] |= byte(rand.Intn(256) & ^int(ipnet.Mask[i]))
		}
		if ipnet.Contains(randomIP) {
			break
		}
	}

	return randomIP, nil
}

func NewChatGPTRequest() CreateConversationRequest {
	return CreateConversationRequest{
		Action:                     "next",
		ParentMessageID:            uuid.NewString(),
		Model:                      "text-davinci-002-render-sha",
		HistoryAndTrainingDisabled: true,
	}
}

func convertAPIRequest(apiRequest GeneralOpenAIRequest) (CreateConversationRequest, string) {
	chatgptRequest := NewChatGPTRequest()

	var model = "gpt-3.5-turbo-0613"

	if strings.HasPrefix(apiRequest.Model, "gpt-3.5") {
		chatgptRequest.Model = "text-davinci-002-render-sha"
	}

	if strings.HasPrefix(apiRequest.Model, "gpt-4") {
		chatgptRequest.Model = apiRequest.Model
		model = "gpt-4"
	}

	for _, apiMessage := range apiRequest.Messages {
		if apiMessage.Role == "system" {
			apiMessage.Role = "critic"
		}
		chatgptRequest.AddMessage(apiMessage.Role, apiMessage.Content.(string))
	}

	return chatgptRequest, model
}

// HandleRequestError 检查HTTP响应状态码并根据情况创建一个新的http.Response对象
func HandleRequestError(response *http.Response) *http.Response {
	if response.StatusCode != 200 {
		// 尝试将响应体作为JSON读取
		var errorResponse map[string]interface{}
		err := json.NewDecoder(response.Body).Decode(&errorResponse)
		if err != nil {
			// 读取响应体
			body, _ := io.ReadAll(response.Body)
			return NewErrorResponse(500, "Unknown error", "internal_server_error", string(body))
		}
		return NewErrorResponse(response.StatusCode, errorResponse["detail"].(string), response.Status, "")
	}
	return nil // 无错误，返回nil
}

func createSuccessResponse(data ChatCompletion) *http.Response {
	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(data)

	response := &http.Response{
		Status:        "200 OK",
		StatusCode:    200,
		Proto:         "HTTP/1.1",
		ProtoMajor:    1,
		ProtoMinor:    1,
		Header:        make(http.Header),
		Body:          io.NopCloser(&buf),
		ContentLength: int64(buf.Len()), // 未知长度
		Close:         false,
	}

	// 设置SSE相关的头信息
	response.Header.Set("Content-Type", "application/json")
	response.Header.Set("Cache-Control", "no-cache")

	return response
}

func createSSEResponse() (*http.Response, *io.PipeWriter) {
	pr, pw := io.Pipe()

	// 创建一个http.Response对象，其中Body是一个管道读取器
	response := &http.Response{
		Status:        "200 OK",
		StatusCode:    200,
		Proto:         "HTTP/1.1",
		ProtoMajor:    1,
		ProtoMinor:    1,
		Header:        make(http.Header),
		Body:          pr,
		ContentLength: -1, // 未知长度
		Close:         false,
	}

	// 设置SSE相关的头信息
	response.Header.Set("Content-Type", "text/event-stream")
	response.Header.Set("Cache-Control", "no-cache")
	response.Header.Set("Connection", "keep-alive")

	return response, pw
}

// NewErrorResponse 创建一个错误的http.Response对象
func NewErrorResponse(statusCode int, message, errorType, details string) *http.Response {
	errorJSON := map[string]interface{}{
		"error": map[string]interface{}{
			"message": message,
			"type":    errorType,
			"param":   nil,
			"code":    "error",
		},
	}
	if details != "" {
		errorJSON["error"].(map[string]interface{})["details"] = details
	}

	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(errorJSON)

	return &http.Response{
		Status:        http.StatusText(statusCode),
		StatusCode:    statusCode,
		Proto:         "HTTP/1.1",
		ProtoMajor:    1,
		ProtoMinor:    1,
		Header:        make(http.Header),
		Body:          io.NopCloser(&buf),
		ContentLength: int64(buf.Len()),
		Close:         false,
	}
}
