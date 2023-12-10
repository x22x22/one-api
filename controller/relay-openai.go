package controller

import (
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
	eventStreamReader := NewEventStreamReader(resp.Body.(io.Reader), 40960)
	defer resp.Body.Close()
	dataChan := make(chan string)
	stopChan := make(chan bool)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop() // 确保 ticker 会被停止
	go func() {
		defer close(dataChan) // 确保 dataChan 会被关闭
		defer close(stopChan) // 确保 stopChan 会被关闭

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
			dataChan <- string(event.Data)

			if !bytes.HasPrefix(event.Data, []byte("[DONE]")) && len(event.Data) > 1 {
				switch relayMode {
				case RelayModeChatCompletions:
					var streamResponse ChatCompletionsStreamResponse
					err := json.Unmarshal(event.Data, &streamResponse)
					if err != nil {
						common.SysError(string(event.Data))
						common.SysError("error unmarshalling stream response: " + err.Error())
						continue // just ignore the error
					}
					for _, choice := range streamResponse.Choices {
						responseText += choice.Delta.Content
					}
				case RelayModeCompletions:
					var streamResponse CompletionsStreamResponse
					err := json.Unmarshal(event.Data, &streamResponse)
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
