package middleware

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"one-api/common"
	"reflect"
	"strconv"
	"time"
)

func Retry(group *gin.RouterGroup) gin.HandlerFunc {
	var retryMiddleware gin.HandlerFunc
	retryMiddleware = func(c *gin.Context) {
		// backup request header and body
		backupReqHeader := c.Request.Header.Clone()
		backupReqBody, err := io.ReadAll(c.Request.Body)
		if err != nil {
			abortWithMessage(c, http.StatusBadRequest, "无效的请求")
			return
		}
		_ = c.Request.Body.Close()
		c.Request.Body = io.NopCloser(bytes.NewBuffer(backupReqBody))

		// 获取Retry Middleware后续的中间件
		found := false
		filteredHandlers := make(gin.HandlersChain, 0)
		for _, handler := range group.Handlers {
			if reflect.ValueOf(handler).Pointer() == reflect.ValueOf(retryMiddleware).Pointer() {
				found = true
				continue
			}
			if found {
				filteredHandlers = append(filteredHandlers, handler)
			}
		}
		// 加入Relay处理函数 c.Handler() => c.handlers.Last() => controller.Relay
		filteredHandlers = append(filteredHandlers, c.Handler())

		// retry
		maxRetryStr := c.Query("retry")
		maxRetry, err := strconv.Atoi(maxRetryStr)
		if err != nil || maxRetryStr == "" || maxRetry < 0 || maxRetry > common.RetryTimes {
			maxRetry = common.RetryTimes
		}
		retryInterval := time.Duration(common.RetryInterval) * time.Millisecond
		for i := maxRetry; i >= 0; i-- {
			c.Set("retry", i)

			if i == maxRetry {
				// 第一次请求, 直接执行后续中间件
				c.Next()
			} else {
				// 重试, 恢复请求头和请求体, 并执行后续中间件
				c.Request.Header = backupReqHeader.Clone()
				c.Request.Body = io.NopCloser(bytes.NewBuffer(backupReqBody))
				for _, handler := range filteredHandlers {
					handler(c)
				}
			}
			// 无错误, 直接返回
			if len(c.Errors) == 0 {
				return
			}

			// 如果有错误，等待一段时间后重试
			time.Sleep(retryInterval)
			// 清理错误列表，以避免后续中间件误解这个错误
			c.Errors = c.Errors[:0]
		}
	}
	return retryMiddleware
}
