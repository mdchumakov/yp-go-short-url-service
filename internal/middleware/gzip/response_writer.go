package gzip

import (
	"compress/gzip"
	"github.com/gin-gonic/gin"
)

type gzipResponseBodyWriter struct {
	gin.ResponseWriter
	writer *gzip.Writer
}

func (r *gzipResponseBodyWriter) Write(b []byte) (int, error) {
	return r.writer.Write(b)
}

func (r *gzipResponseBodyWriter) WriteString(s string) (int, error) {
	return r.Write([]byte(s))
}
