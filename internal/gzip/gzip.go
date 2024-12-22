package gzip

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"github.com/MisterMaks/go-yandex-shortener/internal/logger"
	"go.uber.org/zap"
)

// Used constants.
const (
	ContentTypeKey     string = "Content-Type"
	TextHTTPKey        string = "text/plain"
	ApplicationJSONKey string = "application/json"
	GzipKey            string = "gzip"
	ContentEncodingKey string = "Content-Encoding"
	AcceptEncodingKey  string = "Accept-Encoding"
)

// compressWriter реализует интерфейс http.ResponseWriter и позволяет прозрачно для сервера
// сжимать передаваемые данные и выставлять правильные HTTP-заголовки
type compressWriter struct {
	w  http.ResponseWriter
	zw *gzip.Writer
}

func newCompressWriter(w http.ResponseWriter) *compressWriter {
	return &compressWriter{
		w:  w,
		zw: gzip.NewWriter(w),
	}
}

// Header return response header.
func (c *compressWriter) Header() http.Header {
	return c.w.Header()
}

// Write write data.
func (c *compressWriter) Write(p []byte) (int, error) {
	return c.zw.Write(p)
}

// WriteHeader write header.
func (c *compressWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		c.w.Header().Set(ContentEncodingKey, GzipKey)
	}
	c.w.WriteHeader(statusCode)
}

// Close закрывает gzip.Writer и досылает все данные из буфера.
func (c *compressWriter) Close() error {
	return c.zw.Close()
}

// compressReader реализует интерфейс io.ReadCloser и позволяет прозрачно для сервера
// декомпрессировать получаемые от клиента данные
type compressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

func newCompressReader(r io.ReadCloser) (*compressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &compressReader{
		r:  r,
		zr: zr,
	}, nil
}

// Read read data.
func (c compressReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

// Close close reader.
func (c *compressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.zr.Close()
}

// GzipMiddleware middleware for zip data in response and unzip data from request.
func GzipMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctxLogger := logger.GetContextLogger(r.Context())

		contentType := r.Header.Get(ContentTypeKey)
		if !(strings.Contains(contentType, TextHTTPKey) || strings.Contains(contentType, ApplicationJSONKey) || strings.Contains(contentType, "application/x-gzip")) {
			h.ServeHTTP(w, r)
			return
		}

		// по умолчанию устанавливаем оригинальный http.ResponseWriter как тот,
		// который будем передавать следующей функции
		ow := w

		// проверяем, что клиент умеет получать от сервера сжатые данные в формате gzip
		acceptEncoding := r.Header.Get(AcceptEncodingKey)
		supportsGzip := strings.Contains(acceptEncoding, GzipKey)
		if supportsGzip {
			// оборачиваем оригинальный http.ResponseWriter новым с поддержкой сжатия
			cw := newCompressWriter(w)
			// меняем оригинальный http.ResponseWriter на новый
			ow = cw
			// не забываем отправить клиенту все сжатые данные после завершения middleware
			defer func() {
				err := cw.Close()
				ctxLogger.Warn("Failed to close compressWriter", zap.Error(err))
			}()
		}

		// проверяем, что клиент отправил серверу сжатые данные в формате gzip
		contentEncoding := r.Header.Get(ContentEncodingKey)
		sendsGzip := strings.Contains(contentEncoding, GzipKey)
		if sendsGzip {
			// оборачиваем тело запроса в io.Reader с поддержкой декомпрессии
			cr, err := newCompressReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			// меняем тело запроса на новое
			r.Body = cr
			defer func() {
				err = cr.Close()
				if err != nil {
					ctxLogger.Warn("Failed to close compressReader", zap.Error(err))
				}
			}()
		}

		// передаём управление хендлеру
		h.ServeHTTP(ow, r)
	})
}
