package delivery

import (
	"bytes"
	"context"
	"errors"
	"github.com/MisterMaks/go-yandex-shortener/internal/user/usecase"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MisterMaks/go-yandex-shortener/internal/app"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	TestValidURL   string = "valid_url"
	TestInvalidURL string = "invalid_url"
	TestID         string = "1"
	TestHost       string = "http://example.com"
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
			ID:     TestID,
			URL:    TestValidURL,
			UserID: TestUserID,
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

// TODO
func (tau *testAppUsecase) SendDeleteUserURLsInChan(userID uint, urlIDs []string) {}

func TestAppHandler_GetOrCreateURL(t *testing.T) {
	type request struct {
		method      string
		contentType string
		url         string
		body        []byte
	}
	type want struct {
		statusCode  int
		contentType string
		response    string
	}

	tests := []struct {
		name    string
		request request
		want    want
	}{
		{
			name: "valid URL",
			request: request{
				method:      http.MethodPost,
				contentType: TextPlainKey,
				url:         TestHost + "/",
				body:        []byte(TestValidURL),
			},
			want: want{
				statusCode:  http.StatusCreated,
				response:    TestID,
				contentType: TextPlainKey,
			},
		},
		{
			name: "invalid URL",
			request: request{
				method:      http.MethodPost,
				contentType: TextPlainKey,
				url:         TestHost + "/",
				body:        []byte(TestInvalidURL),
			},
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "invalid method",
			request: request{
				method:      http.MethodGet,
				contentType: TextPlainKey,
				url:         TestHost + "/",
				body:        []byte(TestValidURL),
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
				url:         TestHost + "/",
				body:        []byte(TestValidURL),
			},
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tau := &testAppUsecase{}
			appHandler := NewAppHandler(tau)

			bodyReader := bytes.NewReader(tt.request.body)

			req := httptest.NewRequest(tt.request.method, tt.request.url, bodyReader)
			req.Header.Add(ContentTypeKey, tt.request.contentType)

			ctx := context.WithValue(req.Context(), usecase.UserIDKey, TestUserID)

			w := httptest.NewRecorder()

			appHandler.GetOrCreateURL(w, req.WithContext(ctx))

			res := w.Result()

			assert.Equal(t, tt.want.statusCode, res.StatusCode)
			switch res.StatusCode {
			case http.StatusCreated:
				assert.Contains(t, res.Header.Values(ContentTypeKey), tt.want.contentType)

				defer res.Body.Close()
				resBody, err := io.ReadAll(res.Body)
				require.NoError(t, err)
				assert.Equal(t, tt.want.response, string(resBody))
			}
		})
	}
}

func TestAppHandler_APIGetOrCreateURL(t *testing.T) {
	type request struct {
		method      string
		contentType string
		url         string
		body        []byte
	}
	type want struct {
		statusCode  int
		contentType string
		response    string
	}

	tests := []struct {
		name    string
		request request
		want    want
	}{
		{
			name: "valid URL",
			request: request{
				method:      http.MethodPost,
				contentType: ApplicationJSONKey,
				url:         TestHost + "/api/shorten",
				body:        []byte(`{"url": "` + TestValidURL + `"}`),
			},
			want: want{
				statusCode:  http.StatusCreated,
				response:    `{"result": "1"}`,
				contentType: ApplicationJSONKey,
			},
		},
		{
			name: "invalid URL",
			request: request{
				method:      http.MethodPost,
				contentType: ApplicationJSONKey,
				url:         TestHost + "/api/shorten",
				body:        []byte(`{"url": "` + TestInvalidURL + `"}`),
			},
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "invalid method",
			request: request{
				method:      http.MethodGet,
				contentType: ApplicationJSONKey,
				url:         TestHost + "/api/shorten",
				body:        []byte(`{"url": "` + TestValidURL + `"}`),
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
				url:         TestHost + "/api/shorten",
				body:        []byte(`{"url": "` + TestValidURL + `"}`),
			},
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tau := &testAppUsecase{}
			appHandler := NewAppHandler(tau)

			bodyReader := bytes.NewReader(tt.request.body)

			req := httptest.NewRequest(tt.request.method, tt.request.url, bodyReader)
			req.Header.Add(ContentTypeKey, tt.request.contentType)

			ctx := context.WithValue(req.Context(), usecase.UserIDKey, TestUserID)

			w := httptest.NewRecorder()

			appHandler.APIGetOrCreateURL(w, req.WithContext(ctx))

			res := w.Result()

			assert.Equal(t, tt.want.statusCode, res.StatusCode)
			switch res.StatusCode {
			case http.StatusCreated:
				assert.Contains(t, res.Header.Values(ContentTypeKey), tt.want.contentType)

				defer res.Body.Close()
				resBody, err := io.ReadAll(res.Body)
				require.NoError(t, err)
				assert.JSONEq(t, tt.want.response, string(resBody))
			}
		})
	}
}

func TestAppHandler_RedirectToURL(t *testing.T) {
	type request struct {
		method string
		url    string
		id     string
	}
	type want struct {
		statusCode int
		response   string
	}

	tests := []struct {
		name    string
		request request
		want    want
	}{
		{
			name: "valid ID",
			request: request{
				method: http.MethodGet,
				url:    TestHost + "/",
				id:     TestID,
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
				url:    TestHost + "/",
				id:     "2",
			},
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "invalid method",
			request: request{
				method: http.MethodPost,
				url:    TestHost + "/",
				id:     TestID,
			},
			want: want{
				statusCode: http.StatusMethodNotAllowed,
			},
		},
		{
			name: "invalid url",
			request: request{
				method: http.MethodGet,
				url:    TestHost + "/",
				id:     "",
			},
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tau := &testAppUsecase{}
			appHandler := NewAppHandler(tau)

			req := httptest.NewRequest(tt.request.method, tt.request.url+tt.request.id, nil)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.request.id)

			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			w := httptest.NewRecorder()

			appHandler.RedirectToURL(w, req)

			res := w.Result()

			assert.Equal(t, tt.want.statusCode, res.StatusCode)
			switch res.StatusCode {
			case http.StatusTemporaryRedirect:
				defer res.Body.Close()
				resBody, err := io.ReadAll(res.Body)
				require.NoError(t, err)
				assert.Equal(t, tt.want.response, string(resBody))
			}
		})
	}
}
