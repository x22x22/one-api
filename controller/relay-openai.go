package controller

import (
	"bufio"
	"bytes"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"one-api/common"
	"strings"
	"time"
)

type PreData struct {
	ID    string `json:"id"`
	Model string `json:"model"`
}

func openaiStreamHandler(c *gin.Context, resp *http.Response, relayMode int) (*OpenAIErrorWithStatusCode, string) {
	responseText := ""
	scanner := bufio.NewScanner(resp.Body)
	defer resp.Body.Close()
	scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}
		if i := strings.Index(string(data), "\n"); i >= 0 {
			return i + 1, data[0:i], nil
		}
		if atEOF {
			return len(data), data, nil
		}
		return 0, nil, nil
	})
	dataChan := make(chan string)
	stopChan := make(chan bool)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop() // 确保 ticker 会被停止
	go func() {
		defer close(dataChan) // 确保 dataChan 会被关闭
		defer close(stopChan) // 确保 stopChan 会被关闭
		for scanner.Scan() {
			data := scanner.Text()
			if len(data) < 6 { // ignore blank line or wrong format
				continue
			}
			event, err := processEvent([]byte(data))
			if err != nil {
				return
			}
			dataChan <- string(event.Data)
			data = strings.TrimPrefix(string(event.Data), " ")
			if !strings.HasPrefix(data, "[DONE]") && len(data) > 1 {
				switch relayMode {
				case RelayModeChatCompletions:
					var streamResponse ChatCompletionsStreamResponse
					err := json.Unmarshal([]byte(data), &streamResponse)
					if err != nil {
						common.SysError(data)
						common.SysError("error unmarshalling stream response: " + err.Error())
						continue // just ignore the error
					}
					for _, choice := range streamResponse.Choices {
						responseText += choice.Delta.Content
					}
				case RelayModeCompletions:
					var streamResponse CompletionsStreamResponse
					err := json.Unmarshal([]byte(data), &streamResponse)
					if err != nil {
						common.SysError("error unmarshalling stream response: " + err.Error())
						continue
					}
					for _, choice := range streamResponse.Choices {
						responseText += choice.Text
					}
				}
			}
		}
		stopChan <- true
	}()
	var pre = &ChatCompletionChunk{}
	setEventStreamHeaders(c)
	c.Stream(func(w io.Writer) bool {
		select {
		case <-ticker.C:
			if pre.ID != "" {
				pre.Choices[0].Delta.Content = ""
				c.Render(-1, common.CustomEvent{Data: "data: " + pre.String()})
			}
			return true
		case data := <-dataChan:
			data = strings.TrimSuffix(data, "\r")
			c.Render(-1, common.CustomEvent{Data: "data: " + data})
			_ = json.Unmarshal([]byte(data), &pre)
			return true
		case <-stopChan:
			return false
		}
	})
	return nil, responseText
}

func openaiHandler(c *gin.Context, resp *http.Response, promptTokens int, model string) (*OpenAIErrorWithStatusCode, *Usage) {
	var textResponse TextResponse
	responseBody, compressedResp, err := UnCompressResp(resp)
	if err != nil {
		return errorWrapper(err, "read_response_body_failed", http.StatusInternalServerError), nil
	}
	err = resp.Body.Close()
	if err != nil {
		return errorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}
	err = json.Unmarshal(responseBody, &textResponse)
	if err != nil {
		return errorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError), nil
	}
	if textResponse.Error.Type != "" {
		return &OpenAIErrorWithStatusCode{
			OpenAIError: textResponse.Error,
			StatusCode:  resp.StatusCode,
		}, nil
	}
	// Reset response body
	resp.Body = io.NopCloser(bytes.NewBuffer(compressedResp))
	// We shouldn't set the header before we parse the response body, because the parse part may fail.
	// And then we will have to send an error response, but in this case, the header has already been set.
	// So the httpClient will be confused by the response.
	// For example, Postman will report error, and we cannot check the response at all.
	for k, v := range resp.Header {
		c.Writer.Header().Set(k, v[0])
	}
	c.Writer.WriteHeader(resp.StatusCode)
	_, err = io.Copy(c.Writer, resp.Body)
	if err != nil {
		return errorWrapper(err, "copy_response_body_failed", http.StatusInternalServerError), nil
	}
	err = resp.Body.Close()
	if err != nil {
		return errorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}

	if textResponse.Usage.TotalTokens == 0 {
		completionTokens := 0
		for _, choice := range textResponse.Choices {
			completionTokens += countTokenText(choice.Message.StringContent(), model)
		}
		textResponse.Usage = Usage{
			PromptTokens:     promptTokens,
			CompletionTokens: completionTokens,
			TotalTokens:      promptTokens + completionTokens,
		}
	}
	return nil, &textResponse.Usage
}
