package middleware

import (
	"bytes"
	"github.com/gin-gonic/gin"
)

type traceWriter struct {
	gin.ResponseWriter
	b *bytes.Buffer
}

func (w traceWriter) Write(b []byte) (int, error) {
	w.b.Write(b)
	return w.ResponseWriter.Write(b)
}
