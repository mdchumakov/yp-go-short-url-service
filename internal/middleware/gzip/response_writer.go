package gzip

import (
	"compress/gzip"

	"github.com/gin-gonic/gin"
)

type gzipResponseBodyWriter struct {
	gin.ResponseWriter
	writer *gzip.Writer
	closed bool
}

func (r *gzipResponseBodyWriter) Write(b []byte) (int, error) {
	if r.closed {
		return 0, nil
	}
	return r.writer.Write(b)
}

func (r *gzipResponseBodyWriter) WriteString(s string) (int, error) {
	return r.Write([]byte(s))
}

func (r *gzipResponseBodyWriter) Close() error {
	if r.closed {
		return nil
	}
	r.closed = true
	return r.writer.Close()
}
