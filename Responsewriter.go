package main

import (
	"bytes"
	"net/http"

	"github.com/gin-gonic/gin"
)

// CompressResponseWriter is a custom writer that
// mimics Gin's ResponseWriter without actual I/O.
type CompressResponseWriter struct {
	headers            http.Header
	buffer             *bytes.Buffer
	status             int
	gin.ResponseWriter // risky as this value is nil and
	// will trigger an error if an um-implemented method is called.
}

func (w *CompressResponseWriter) Header() http.Header {
	return w.headers
}

func (w *CompressResponseWriter) Write(data []byte) (int, error) {
	return w.buffer.Write(data)
}

func (w *CompressResponseWriter) WriteHeader(statusCode int) {
	w.status = statusCode
}

func (w *CompressResponseWriter) Status() int {
	return w.status
}

func (w *CompressResponseWriter) Written() bool {
	return w.buffer.Len() > 0
}
