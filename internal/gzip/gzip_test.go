package gzip

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MisterMaks/go-yandex-shortener/internal/logger"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

const (
	RequestBodyKey string = "request_body"
	TextPlainKey   string = "text/plain"
)

func TestGzipCompression(t *testing.T) {
	decodedRequestBodyStrCh := make(chan string, 2)

	testRequestBodyStr := "If I bring my army into your land, I will destroy your farms, slay your people, and raze your city! (Philip II)"
	testResponseBodyStr := "If! (This is sparta :) )"

	handler := http.Handler(GzipMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			logger.Log.Fatal("Unexpected error",
				zap.Any(RequestBodyKey, r.Body),
				zap.Error(err),
			)
		}
		bodyStr := string(body)

		decodedRequestBodyStrCh <- bodyStr

		w.Write([]byte(testResponseBodyStr))
	})))

	srv := httptest.NewServer(handler)
	defer srv.Close()

	t.Run("sends_gzip", func(t *testing.T) {
		buf := bytes.NewBuffer(nil)
		zb := gzip.NewWriter(buf)
		_, err := zb.Write([]byte(testRequestBodyStr))
		require.NoError(t, err)
		err = zb.Close()
		require.NoError(t, err)

		r := httptest.NewRequest(http.MethodPost, srv.URL, buf)
		r.RequestURI = ""
		r.Header.Set(ContentTypeKey, TextPlainKey)
		r.Header.Set(ContentEncodingKey, GzipKey)
		r.Header.Set(AcceptEncodingKey, "")

		resp, err := http.DefaultClient.Do(r)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		decodedRequestBodyStr := <-decodedRequestBodyStrCh
		require.Equal(t, testRequestBodyStr, decodedRequestBodyStr)

		defer resp.Body.Close()

		b, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Equal(t, testResponseBodyStr, string(b))
	})

	t.Run("accepts_gzip", func(t *testing.T) {
		buf := bytes.NewBufferString(testRequestBodyStr)
		r := httptest.NewRequest("POST", srv.URL, buf)
		r.RequestURI = ""
		r.Header.Set(ContentTypeKey, TextPlainKey)
		r.Header.Set(AcceptEncodingKey, GzipKey)

		resp, err := http.DefaultClient.Do(r)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		decodedRequestBodyStr := <-decodedRequestBodyStrCh
		require.Equal(t, testRequestBodyStr, decodedRequestBodyStr)

		defer resp.Body.Close()

		zr, err := gzip.NewReader(resp.Body)
		require.NoError(t, err)

		b, err := io.ReadAll(zr)
		require.NoError(t, err)

		require.Equal(t, testResponseBodyStr, string(b))
	})

	close(decodedRequestBodyStrCh)
}
