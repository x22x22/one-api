package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"runtime"
	"strings"
	"time"

	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/model/completion"
	completionStream "github.com/songquanpeng/one-api/model/completion-stream"
	"github.com/x22x22/langfuse-go"

	"github.com/gin-gonic/gin"
	lfModel "github.com/x22x22/langfuse-go/model"
)

func ChatCompletionsTrace() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		requestId := c.Keys["X-Oneapi-Request-Id"].(string)
		logger.SysLogf("[请求ID:%s] Trace开始 - Goroutine数量: %d, CPU使用数: %d", requestId, runtime.NumGoroutine(), runtime.NumCPU())

		if !strings.Contains(c.GetHeader("Content-Type"), "application/json") {
			c.Next()
			return
		}
		// Read request body
		reqBody, err := io.ReadAll(c.Request.Body)
		if err != nil {
			logger.Error(c, fmt.Sprintf("Failed to read request body: %v", err))
			c.Next()
			return
		}
		c.Request.Body = io.NopCloser(bytes.NewBuffer(reqBody))

		var reqData map[string]interface{}
		if err := json.Unmarshal(reqBody, &reqData); err != nil {
			logger.Error(c, fmt.Sprintf("Failed to unmarshal request data: %v", err))
			c.Next()
			return
		}

		// Check if stream mode
		isStream := false
		if stream, ok := reqData["stream"].(bool); ok && stream {
			isStream = true
		}

		// Create trace
		l := langfuse.New(context.Background())
		trace, err := l.Trace(&lfModel.Trace{
			Name: fmt.Sprintf("llm-gateway:%s %s", c.Request.Method, c.Request.URL.Path),
		})

		trace.SessionID = c.GetHeader("x-session-id")

		if err != nil {
			logger.Error(c, fmt.Sprintf("Failed to create trace: %v", err))
			c.Next()
			return
		}

		// Create span
		span, err := l.Span(&lfModel.Span{
			Name:    "request",
			TraceID: trace.ID,
		}, nil)
		if err != nil {
			logger.Error(c, fmt.Sprintf("Failed to create span: %v", err))
			c.Next()
			return
		}

		modelName := ""
		if m, ok := reqData["model"].(string); ok {
			modelName = m
		}
		maskToken := common.MaskToken(c.Keys["authorization"].(string))

		var request completion.ChatCompletionRequest

		requestJson, err := json.Marshal(reqData)
		if err != nil {
			logger.Error(c, fmt.Sprintf("%v", err))
		}
		err = json.Unmarshal(requestJson, &request)
		if err != nil {
			logger.Error(c, fmt.Sprintf("%v", err))
		}
		trace.Input = request

		// 获取最后一条用户消息的内容
		content := ""
		for i := len(request.Messages) - 1; i >= 0; i-- {
			if request.Messages[i].Content != "" {
				content = request.Messages[i].Content
				break
			}
		}

		content = strings.ReplaceAll(content, "\n", "")
		contentLen := len([]rune(content))
		role := request.Messages[0].Role
		SubString := common.SubString
		var msgSummary string
		if contentLen < 40 {
			msgSummary = fmt.Sprintf("m=%s:%s", role, content)
		} else if contentLen >= 40 && contentLen < 120 {
			msgSummary = fmt.Sprintf("m=%s:%s;%s", role, SubString(content, 0, 10), SubString(content, 30, 50))
		} else {
			msgSummary = fmt.Sprintf("m=%s:%s;%s;%s", role, SubString(content, 0, 10), SubString(content, 30, 50), SubString(content, 60, 80))
		}

		generation, err := l.Generation(
			&lfModel.Generation{
				TraceID: trace.ID,
				Name:    "llm_request",
				Model:   modelName,
				ModelParameters: lfModel.M{
					"temperature": reqData["temperature"],
					"max_tokens":  reqData["max_tokens"],
					"stream":      isStream,
				},
				Input: reqData,
				Metadata: lfModel.M{
					"path":                c.Request.URL.Path,
					"method":              c.Request.Method,
					"token_name":          c.Keys["token_name"],
					"token_id":            c.Keys["token_id"],
					"mask_token":          maskToken,
					"channel_id":          c.Keys["channel_id"],
					"channel_name":        c.Keys["channel_name"],
					"request_model":       c.Keys["request_model"],
					"original_model":      c.Keys["original_model"],
					"group":               c.Keys["group"],
					"X-Oneapi-Request-Id": c.Keys["X-Oneapi-Request-Id"],
					"msg_summary":         msgSummary,
				},
			},
			&span.ID,
		)
		if err != nil {
			logger.Error(c, fmt.Sprintf("Failed to create generation: %v", err))
			c.Next()
			return
		}

		trace.Tags = append(trace.Tags, msgSummary)
		trace.Metadata = generation.Metadata

		if isStream {
			// Handle streaming response
			originalWriter := c.Writer
			streamBuf := new(bytes.Buffer)
			writer := &traceWriter{
				ResponseWriter: originalWriter,
				b:              streamBuf,
			}
			c.Writer = writer

			c.Next()

			respData := streamBuf.String()
			statusCode := c.Writer.Status()

			if generation != nil {
				// For streaming responses, collect all events
				events := strings.Split(respData, "\n\n")
				var streamResponses []interface{}
				var content string

				for _, event := range events {
					if strings.TrimSpace(event) != "" && !strings.Contains(event, "data: [DONE]") {
						eventData := strings.TrimPrefix(event, "data: ")
						var resp completionStream.ChatCompletionChunk
						if err := json.Unmarshal([]byte(eventData), &resp); err != nil {
							logger.Error(c, fmt.Sprintf("Failed to unmarshal stream response: %v", err))
							continue
						}
						streamResponses = append(streamResponses, resp)
						if len(resp.Choices) > 0 {
							content += resp.Choices[0].Delta.Content
						}
					}
				}

				generation.Output = lfModel.M{
					"status_code":      statusCode,
					"stream_responses": streamResponses,
				}
				trace.Output = content
				if _, err := l.GenerationEnd(generation); err != nil {
					logger.Error(c, fmt.Sprintf("Failed to end generation: %v", err))
				}
			}
		} else {
			// Handle normal response
			writer := &traceWriter{
				ResponseWriter: c.Writer,
				b:              bytes.NewBuffer([]byte{}),
			}
			c.Writer = writer

			c.Next()

			statusCode := c.Writer.Status()
			respData, err := io.ReadAll(writer.b)
			if err != nil {
				logger.Error(c, fmt.Sprintf("Failed to read response data: %v", err))
				return
			}

			if generation != nil {
				if statusCode != 200 {
					// 对于非200状态码，直接记录原始错误响应
					var errorResp map[string]interface{}
					if err := json.Unmarshal(respData, &errorResp); err != nil {
						logger.Error(c, fmt.Sprintf("Failed to unmarshal error response: %v", err))
						errorResp = map[string]interface{}{
							"raw_response": string(respData),
						}
					}

					generation.Output = lfModel.M{
						"status_code": statusCode,
						"error":       errorResp,
					}
					trace.Output = fmt.Sprintf("Error: %v", errorResp)
				} else {
					var respDataMap completion.ChatCompletion
					if err := json.Unmarshal(respData, &respDataMap); err != nil {
						logger.Error(c, fmt.Sprintf("Failed to unmarshal response data: %v", err))
						return
					}

					generation.Output = lfModel.M{
						"status_code": statusCode,
						"response":    respDataMap,
					}
					generation.Usage = lfModel.Usage{
						Input:  respDataMap.Usage.PromptTokens,
						Output: respDataMap.Usage.CompletionTokens,
						Total:  respDataMap.Usage.TotalTokens,
					}
					trace.Output = respDataMap.Choices[0].Message.Content
				}

				if _, err := l.GenerationEnd(generation); err != nil {
					logger.Error(c, fmt.Sprintf("Failed to end generation: %v", err))
				}
			}
		}

		duration := time.Since(startTime)

		// Add score based on status code
		score := 1.0
		if c.Writer.Status() >= 400 {
			score = 0.0
		}
		if _, err := l.Score(&lfModel.Score{
			TraceID: trace.ID,
			Name:    "status_code",
			Value:   score,
		}); err != nil {
			logger.Error(c, fmt.Sprintf("Failed to create score: %v", err))
		}

		// End span with metadata
		if span != nil {
			span.Metadata = lfModel.M{
				"duration_ms": duration.Milliseconds(),
				"status":      c.Writer.Status(),
				"stream":      isStream,
			}
			if _, err := l.SpanEnd(span); err != nil {
				logger.Error(c, fmt.Sprintf("Failed to end span: %v", err))
			}
		}

		// 记录 trace
		l.Trace(trace)

		// 异步执行 Flush
		logger.SysLogf("[请求ID:%s] 准备执行Flush - Goroutine数量: %d", requestId, runtime.NumGoroutine())
		go func() {
			flushStart := time.Now()
			l.Flush(context.Background())
			logger.SysLogf("[请求ID:%s] Flush完成 - 耗时: %v, Goroutine数量: %d", requestId, time.Since(flushStart), runtime.NumGoroutine())
		}()

		runtime.ReadMemStats(&m)
		logger.SysLogf("[请求ID:%s] Trace结束 - 总耗时: %v, Goroutine数量: %d, 堆对象数: %d",
			requestId,
			time.Since(startTime),
			runtime.NumGoroutine(),
			m.HeapObjects)
	}
}
