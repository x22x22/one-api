package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/model/completion"
	"github.com/x22x22/langfuse-go"
	"io"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	lfModel "github.com/x22x22/langfuse-go/model"
)

func EmbeddingsTrace() gin.HandlerFunc {
	return func(c *gin.Context) {
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
				Input: []lfModel.M{
					{
						"request": reqData,
					},
				},
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
				},
			},
			&span.ID,
		)
		if err != nil {
			logger.Error(c, fmt.Sprintf("Failed to create generation: %v", err))
			c.Next()
			return
		}

		trace.Input = generation.Input
		trace.Metadata = generation.Metadata

		startTime := time.Now()

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
			var respDataMap completion.EmbeddingResponse
			if err := json.Unmarshal(respData, &respDataMap); err != nil {
				logger.Error(c, fmt.Sprintf("Failed to unmarshal response data: %v", err))
				return
			}

			generation.Output = lfModel.M{
				"status_code": statusCode,
				"response":    respDataMap,
			}
			generation.Usage = lfModel.Usage{
				Input:       respDataMap.Usage.PromptTokens,
				TotalTokens: respDataMap.Usage.TotalTokens,
			}
			trace.Output = respDataMap.Data[0].Embedding
			if _, err := l.GenerationEnd(generation); err != nil {
				logger.Error(c, fmt.Sprintf("Failed to end generation: %v", err))
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

		// Flush traces
		l.Trace(trace)
		l.Flush(context.Background())
	}
}
