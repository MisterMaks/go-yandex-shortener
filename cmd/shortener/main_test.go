package main

import (
	"bytes"
	"context"
	"errors"
	"github.com/MisterMaks/go-yandex-shortener/internal/gzip"
	"github.com/MisterMaks/go-yandex-shortener/internal/logger"
	"github.com/MisterMaks/go-yandex-shortener/internal/user/usecase"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MisterMaks/go-yandex-shortener/internal/app"
	appDeliveryInternal "github.com/MisterMaks/go-yandex-shortener/internal/app/delivery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	TestValidURL   string = "valid_url"
	TestInvalidURL string = "invalid_url"
	TestID         string = "1"
	ContentTypeKey string = "Content-Type"
	TextPlainKey   string = "text/plain"
	TestUserID     uint   = 1
)

var (
	ErrTestInvalidURL = errors.New("invalid url")
	ErrTestIDNotFound = errors.New("ID not found")
)

type testAppUsecase struct{}

func (tau *testAppUsecase) GetOrCreateURL(rawURL string, userID uint) (*app.URL, bool, error) {
	switch rawURL {
	case TestValidURL:
		return &app.URL{
			ID:  TestID,
			URL: TestValidURL,
		}, false, nil
	}
	return nil, false, ErrTestInvalidURL
}

func (tau *testAppUsecase) GetURL(id string) (*app.URL, error) {
	switch id {
	case TestID:
		return &app.URL{
			ID:  TestID,
			URL: TestValidURL,
		}, nil
	}
	return nil, ErrTestIDNotFound
}

func (tau *testAppUsecase) GenerateShortURL(id string) string {
	return id
}

func (tau *testAppUsecase) Ping() error {
	return nil
}

// TODO
func (tau *testAppUsecase) GetOrCreateURLs(requestBatchURLs []app.RequestBatchURL, userID uint) ([]app.ResponseBatchURL, error) {
	return []app.ResponseBatchURL{}, nil
}

// TODO
func (tau *testAppUsecase) GetUserURLs(userID uint) ([]app.ResponseUserURL, error) {
	return []app.ResponseUserURL{}, nil
}

func testRequest(
	t *testing.T,
	ts *httptest.Server,
	method, path, contentType string,
	body []byte,
) (*http.Response, error) {
	req, err := http.NewRequest(method, ts.URL+path, bytes.NewReader(body))
	require.NoError(t, err)
	req.Header.Set(ContentTypeKey, contentType)
	return ts.Client().Do(req)
}

func TestRouter(t *testing.T) {
	tau := &testAppUsecase{}
	appHandler := appDeliveryInternal.NewAppHandler(tau)
	middlewares := &Middlewares{
		RequestLogger:  logger.RequestLogger,
		GzipMiddleware: gzip.GzipMiddleware,
		AuthenticateOrRegister: func(h http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				ctx := context.WithValue(r.Context(), usecase.UserIDKey, TestUserID)
				h.ServeHTTP(w, r.WithContext(ctx))
			})
		},
		Authenticate: func(h http.Handler) http.Handler {
			return h
		},
	}
	ts := httptest.NewServer(shortenerRouter(appHandler, "/", middlewares))
	defer ts.Close()
	client := ts.Client()
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	type request struct {
		method      string
		contentType string
		path        string
		body        []byte
	}
	type want struct {
		statusCode  int
		contentType string
		response    string
	}

	var testTable = []struct {
		name    string
		request request
		want    want
	}{
		{
			name: "create new valid url",
			request: request{
				method:      http.MethodPost,
				contentType: TextPlainKey,
				path:        "/",
				body:        []byte(TestValidURL),
			},
			want: want{
				statusCode:  http.StatusCreated,
				contentType: TextPlainKey,
				response:    TestID,
			},
		},
		{
			name: "create new invalid url",
			request: request{
				method:      http.MethodPost,
				contentType: TextPlainKey,
				path:        "/",
				body:        []byte(TestInvalidURL),
			},
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "invalid method",
			request: request{
				method: http.MethodGet,
				path:   "/",
			},
			want: want{
				statusCode: http.StatusMethodNotAllowed,
			},
		},
		{
			name: "invalid Header Content-Type",
			request: request{
				method:      http.MethodPost,
				contentType: "invalid Header Content-Type",
				path:        "/",
				body:        []byte(TestValidURL),
			},
			want: want{
				statusCode:  http.StatusBadRequest,
				contentType: "",
				response:    "",
			},
		},
		{
			name: "valid ID",
			request: request{
				method: http.MethodGet,
				path:   "/" + TestID,
			},
			want: want{
				statusCode: http.StatusTemporaryRedirect,
				response:   "<a href=\"/" + TestValidURL + "\">Temporary Redirect</a>.\n\n",
			},
		},
		{
			name: "invalid ID",
			request: request{
				method: http.MethodGet,
				path:   "/invalid_id",
			},
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "invalid method",
			request: request{
				method: http.MethodPost,
				path:   "/" + TestID,
			},
			want: want{
				statusCode: http.StatusMethodNotAllowed,
			},
		},
	}

	for _, tt := range testTable {
		resp, err := testRequest(
			t,
			ts,
			tt.request.method, tt.request.path, tt.request.contentType,
			tt.request.body,
		)
		require.NoError(t, err)

		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		require.NoError(t, err)
		respBodyStr := string(respBody)

		assert.Equal(t, tt.want.statusCode, resp.StatusCode, tt.name)
		if resp.StatusCode == http.StatusCreated {
			assert.Contains(t, resp.Header.Values(ContentTypeKey), tt.want.contentType)
			assert.Equal(t, tt.want.response, respBodyStr)
		}
		if resp.StatusCode == http.StatusTemporaryRedirect {
			assert.Equal(t, tt.want.response, respBodyStr)
		}
	}
}
