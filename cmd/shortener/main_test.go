package main

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/MisterMaks/go-yandex-shortener/internal/app/delivery/mocks"
	"github.com/MisterMaks/go-yandex-shortener/internal/gzip"
	"github.com/MisterMaks/go-yandex-shortener/internal/logger"
	"github.com/MisterMaks/go-yandex-shortener/internal/user/usecase"
	"github.com/golang/mock/gomock"

	"github.com/MisterMaks/go-yandex-shortener/internal/app"
	appDeliveryInternal "github.com/MisterMaks/go-yandex-shortener/internal/app/delivery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	TestValidURL   string = "http://valid_url.ru/"
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
	// создаём контроллер
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// создаём объект-заглушку
	m := mocks.NewMockAppUsecaseInterface(ctrl)

	// гарантируем, что заглушка
	// при вызове с аргументом "Key" вернёт "Value"
	m.EXPECT().GetOrCreateURL(TestValidURL, gomock.Any()).Return(&app.URL{
		ID:     TestID,
		URL:    TestValidURL,
		UserID: TestUserID,
	}, false, nil).AnyTimes()
	m.EXPECT().GetOrCreateURL(gomock.Any(), gomock.Any()).Return(nil, false, ErrTestInvalidURL).AnyTimes()

	m.EXPECT().GetURL(TestID).Return(&app.URL{
		ID:  TestID,
		URL: TestValidURL,
	}, nil).AnyTimes()
	m.EXPECT().GetURL(gomock.Any()).Return(nil, ErrTestIDNotFound).AnyTimes()

	m.EXPECT().GenerateShortURL(gomock.Any()).DoAndReturn(
		func(id string) string {
			return id
		},
	)

	appHandler := appDeliveryInternal.NewAppHandler(m)

	middlewares := &Middlewares{
		RequestLogger:  logger.RequestLogger,
		GzipMiddleware: gzip.GzipMiddleware,
		AuthenticateOrRegister: func(h http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				ctx := context.WithValue(r.Context(), usecase.UserIDKey, TestUserID)
				h.ServeHTTP(w, r.WithContext(ctx))
			})
		},
		Authenticate: func(h http.Handler) http.Handler { return h },
	}

	u, err := url.ParseRequestURI(ResultAddrPrefix)
	require.NoError(t, err)

	ts := httptest.NewServer(shortenerRouter(appHandler, u, middlewares))
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
				response:   "<a href=\"" + TestValidURL + "\">Temporary Redirect</a>.\n\n",
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

		assert.Equal(t, tt.want.statusCode, resp.StatusCode)
		if resp.StatusCode == http.StatusCreated {
			assert.Contains(t, resp.Header.Values(ContentTypeKey), tt.want.contentType)
			assert.Equal(t, tt.want.response, respBodyStr)
		}
		if resp.StatusCode == http.StatusTemporaryRedirect {
			assert.Equal(t, tt.want.response, respBodyStr)
		}
	}
}
