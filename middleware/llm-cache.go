package middleware

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/common/logger"
)

type CacheData struct {
	Auth     *string                `json:"auth,omitempty"`
	Request  map[string]interface{} `json:"request"`
	Response interface{}            `json:"response"`
}

type CacheDataStream struct {
	Auth     *string                `json:"auth,omitempty"`
	Request  map[string]interface{} `json:"request"`
	Response interface{}            `string:"response"`
}

func LLMCache(cacheDir ...string) gin.HandlerFunc {
	dir := "./llm_cache"
	if len(cacheDir) > 0 {
		dir = cacheDir[0]
	}

	return func(c *gin.Context) {
		if c.GetHeader("Content-Type") != "application/json" {
			c.Next()
			return
		}

		reqBody, _ := io.ReadAll(c.Request.Body)
		c.Request.Body = io.NopCloser(bytes.NewBuffer(reqBody))

		var reqData map[string]interface{}
		json.Unmarshal(reqBody, &reqData)

		model := ""
		if m, ok := reqData["model"].(string); ok {
			model = m
		}

		reqHash := generateReqHash(c)
		cacheKey := fmt.Sprintf("llm_cache:%s:%s", model, reqHash)

		isStream := false
		if stream, ok := reqData["stream"].(bool); ok && stream {
			isStream = true
		}

		var auth *string
		authValue := c.GetHeader("Authorization")
		if authValue != "" {
			auth = &authValue
		}

		miss := true
		var cacheData []byte

		if common.RedisLLMEnabled {
			cacheDataStr, err := common.RedisGet(cacheKey)
			if err == nil {
				cacheData = []byte(cacheDataStr)
				miss = false
			}
		} else {
			cacheDir := filepath.Join(dir, model)
			_ = os.MkdirAll(cacheDir, 0755)
			cacheFile := filepath.Join(cacheDir, reqHash)
			if _, err := os.Stat(cacheFile); err == nil {
				cacheData, err = os.ReadFile(cacheFile)
				if err == nil {
					miss = false
				}
			}
		}

		if !miss {
			if isStream {
				var streamData CacheDataStream
				json.Unmarshal(cacheData, &streamData)
				c.Header("Content-Type", "text/event-stream")
				c.Header("Cache-Control", "no-cache")
				c.Header("Connection", "keep-alive")
				c.Stream(func(w io.Writer) bool {
					responseString := streamData.Response.(string)
					events := strings.Split(responseString, "\n\n")
					for _, event := range events {
						if event != "" {
							fmt.Fprintf(w, "%s\n\n", event)
							w.(http.Flusher).Flush()
						}
					}
					return false
				})
				c.Abort()
			} else {
				var normalData CacheData
				json.Unmarshal(cacheData, &normalData)
				c.JSON(200, normalData.Response)
				c.Abort()
			}
			return
		}

		writer := responseWriter{
			c.Writer,
			bytes.NewBuffer([]byte{}),
		}
		c.Writer = writer
		c.Next()
		statusCode := c.Writer.Status()
		respData, _ := io.ReadAll(writer.b)
		if statusCode == 200 {
			go func() {
				var cacheDataBytes []byte
				if isStream {
					respDataString := string(respData)
					if strings.Contains(respDataString, "[DONE]") {
						cacheData := CacheDataStream{
							Auth:     auth,
							Request:  reqData,
							Response: respDataString,
						}
						cacheDataBytes, _ = json.Marshal(cacheData)
					}
				} else {
					var respDataMap map[string]interface{}
					json.Unmarshal(respData, &respDataMap)
					cacheData := CacheData{
						Auth:     auth,
						Request:  reqData,
						Response: respDataMap,
					}
					cacheDataBytes, _ = json.Marshal(cacheData)
				}

				if len(cacheDataBytes) > 0 {
					if common.RedisLLMEnabled {
						err := common.RedisSet(cacheKey, string(cacheDataBytes), 24*time.Hour)
						if err != nil {
							logger.SysError("Failed to set Redis cache: " + err.Error())
						}
					} else {
						cacheDir := filepath.Join(dir, model)
						_ = os.MkdirAll(cacheDir, 0755)
						cacheFile := filepath.Join(cacheDir, reqHash)
						err := os.WriteFile(cacheFile, cacheDataBytes, 0644)
						if err != nil {
							logger.SysError("Failed to write file cache: " + err.Error())
						}
					}
				}
			}()
		}
	}
}

type responseWriter struct {
	gin.ResponseWriter
	b *bytes.Buffer
}

func (w responseWriter) Write(b []byte) (int, error) {
	w.b.Write(b)
	return w.ResponseWriter.Write(b)
}

func generateReqHash(c *gin.Context) string {
	var reqKey strings.Builder
	reqKey.WriteString(c.Request.Method)
	reqKey.WriteString(c.Request.URL.Path)

	reqBody, _ := io.ReadAll(c.Request.Body)
	c.Request.Body = io.NopCloser(bytes.NewBuffer(reqBody))
	reqKey.Write(reqBody)

	hash := sha256.Sum256([]byte(reqKey.String()))
	return hex.EncodeToString(hash[:])
}
