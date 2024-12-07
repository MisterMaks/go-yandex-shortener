package delivery

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MisterMaks/go-yandex-shortener/internal/app/delivery/mocks"
	"github.com/MisterMaks/go-yandex-shortener/internal/user/usecase"
	"github.com/golang/mock/gomock"

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

	m.EXPECT().GenerateShortURL(gomock.Any()).DoAndReturn(
		func(id string) string {
			return id
		},
	)

	appHandler := NewAppHandler(m)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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

	m.EXPECT().GenerateShortURL(gomock.Any()).DoAndReturn(
		func(id string) string {
			return id
		},
	)

	appHandler := NewAppHandler(m)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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

	// создаём контроллер
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// создаём объект-заглушку
	m := mocks.NewMockAppUsecaseInterface(ctrl)

	// гарантируем, что заглушка
	// при вызове с аргументом "Key" вернёт "Value"
	m.EXPECT().GetURL(TestID).Return(&app.URL{
		ID:  TestID,
		URL: TestValidURL,
	}, nil).AnyTimes()
	m.EXPECT().GetURL(gomock.Any()).Return(nil, ErrTestIDNotFound).AnyTimes()

	appHandler := NewAppHandler(m)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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

func TestAppHandler_Ping(t *testing.T) {
	tests := []struct {
		name             string
		requestMethod    string
		usecasePingError error
		wantStatusCode   int
	}{
		{
			name:             "simple case",
			requestMethod:    http.MethodGet,
			usecasePingError: nil,
			wantStatusCode:   http.StatusOK,
		},
		{
			name:             "internal server error",
			requestMethod:    http.MethodGet,
			usecasePingError: fmt.Errorf("internal server error"),
			wantStatusCode:   http.StatusInternalServerError,
		},
		{
			name:             "invalid method",
			requestMethod:    http.MethodPost,
			usecasePingError: nil,
			wantStatusCode:   http.StatusMethodNotAllowed,
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := mocks.NewMockAppUsecaseInterface(ctrl)
			m.EXPECT().Ping().Return(tt.usecasePingError).AnyTimes()

			appHandler := NewAppHandler(m)

			req := httptest.NewRequest(tt.requestMethod, TestHost+"/ping", bytes.NewReader(nil))
			w := httptest.NewRecorder()
			appHandler.Ping(w, req)
			res := w.Result()
			err := res.Body.Close()
			require.NoError(t, err)
			assert.Equal(t, tt.wantStatusCode, res.StatusCode, "Invalid status code")
		})
	}
}

func TestAppHandler_APIGetOrCreateURLs(t *testing.T) {
	contextUserID := uint(1)

	type request struct {
		method      string
		body        []byte
		contentType string
		ctx         context.Context
	}

	type want struct {
		statusCode int
		body       []byte
	}

	tests := []struct {
		name    string
		request request
		want    want
	}{
		{
			name: "valid data",
			request: request{
				method:      http.MethodPost,
				body:        []byte(fmt.Sprintf(`[{"correlation_id": "%s", "original_url": "%s"}]`, TestID, TestValidURL)),
				contentType: ApplicationJSONKey,
				ctx:         context.WithValue(context.Background(), usecase.UserIDKey, contextUserID),
			},
			want: want{
				statusCode: http.StatusCreated,
				body:       []byte(fmt.Sprintf(`[{"correlation_id": "%s", "short_url": "%s"}]`, TestID, TestHost+"/"+TestID)),
			},
		},
		{
			name: "invalid method",
			request: request{
				method:      http.MethodGet,
				body:        []byte(fmt.Sprintf(`[{"correlation_id": "%s", "original_url": "%s"}]`, TestID, TestValidURL)),
				contentType: ApplicationJSONKey,
				ctx:         context.WithValue(context.Background(), usecase.UserIDKey, contextUserID),
			},
			want: want{
				statusCode: http.StatusMethodNotAllowed,
				body:       nil,
			},
		},
		{
			name: "invalid Content-Type",
			request: request{
				method:      http.MethodPost,
				body:        []byte(fmt.Sprintf(`[{"correlation_id": "%s", "original_url": "%s"}]`, TestID, TestValidURL)),
				contentType: "invalid Content-Type",
				ctx:         context.WithValue(context.Background(), usecase.UserIDKey, contextUserID),
			},
			want: want{
				statusCode: http.StatusBadRequest,
				body:       nil,
			},
		},
		{
			name: "invalid body",
			request: request{
				method:      http.MethodPost,
				body:        []byte(fmt.Sprintf(`[{"correlation_id": "%s", "original_url": "%s"`, TestID, TestValidURL)),
				contentType: ApplicationJSONKey,
				ctx:         context.WithValue(context.Background(), usecase.UserIDKey, contextUserID),
			},
			want: want{
				statusCode: http.StatusBadRequest,
				body:       nil,
			},
		},
		{
			name: "unauthorized user",
			request: request{
				method:      http.MethodPost,
				body:        []byte(fmt.Sprintf(`[{"correlation_id": "%s", "original_url": "%s"}]`, TestID, TestValidURL)),
				contentType: ApplicationJSONKey,
				ctx:         context.Background(),
			},
			want: want{
				statusCode: http.StatusUnauthorized,
				body:       nil,
			},
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mocks.NewMockAppUsecaseInterface(ctrl)

	requestBatchURLs := []app.RequestBatchURL{
		{
			CorrelationID: TestID,
			OriginalURL:   TestValidURL,
		},
	}

	responseBatchURLs := []app.ResponseBatchURL{
		{
			CorrelationID: TestID,
			ShortURL:      TestHost + "/" + TestID,
		},
	}

	m.EXPECT().GetOrCreateURLs(requestBatchURLs, contextUserID).Return(responseBatchURLs, nil).AnyTimes()

	appHandler := NewAppHandler(m)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.request.method, TestHost+"/api/shorten/batch", bytes.NewReader(tt.request.body))
			req.Header.Set(ContentTypeKey, tt.request.contentType)
			req = req.WithContext(tt.request.ctx)

			w := httptest.NewRecorder()

			appHandler.APIGetOrCreateURLs(w, req)

			res := w.Result()

			resBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			err = res.Body.Close()
			require.NoError(t, err)

			assert.Equal(t, tt.want.statusCode, res.StatusCode, "Invalid status code")

			if res.StatusCode == http.StatusCreated {
				assert.JSONEq(t, string(tt.want.body), string(resBody))
			}
		})
	}
}

func TestAppHandler_APIGetUserURLs(t *testing.T) {
	contextUserID := uint(1)

	type request struct {
		method string
		ctx    context.Context
	}

	type usecaseGetUserURLsResponse struct {
		userID   uint
		userURLs []app.ResponseUserURL
		err      error
	}

	type want struct {
		statusCode int
		body       []byte
	}

	tests := []struct {
		name                       string
		request                    request
		usecaseGetUserURLsResponse usecaseGetUserURLsResponse
		want                       want
	}{
		{
			name: "valid data",
			request: request{
				method: http.MethodGet,
				ctx:    context.WithValue(context.Background(), usecase.UserIDKey, contextUserID),
			},
			usecaseGetUserURLsResponse: usecaseGetUserURLsResponse{
				userID:   contextUserID,
				userURLs: []app.ResponseUserURL{{OriginalURL: TestValidURL, ShortURL: TestHost + "/" + TestID}},
				err:      nil,
			},
			want: want{
				statusCode: http.StatusOK,
				body:       []byte(fmt.Sprintf(`[{"original_url": "%s", "short_url": "%s"}]`, TestValidURL, TestHost+"/"+TestID)),
			},
		},
		{
			name: "no content",
			request: request{
				method: http.MethodGet,
				ctx:    context.WithValue(context.Background(), usecase.UserIDKey, contextUserID),
			},
			usecaseGetUserURLsResponse: usecaseGetUserURLsResponse{
				userID:   contextUserID,
				userURLs: nil,
				err:      nil,
			},
			want: want{
				statusCode: http.StatusNoContent,
				body:       nil,
			},
		},
		{
			name: "invalid method",
			request: request{
				method: http.MethodPost,
				ctx:    context.WithValue(context.Background(), usecase.UserIDKey, contextUserID),
			},
			usecaseGetUserURLsResponse: usecaseGetUserURLsResponse{
				userID:   contextUserID,
				userURLs: []app.ResponseUserURL{{OriginalURL: TestValidURL, ShortURL: TestHost + "/" + TestID}},
				err:      nil,
			},
			want: want{
				statusCode: http.StatusMethodNotAllowed,
				body:       nil,
			},
		},
		{
			name: "unauthorized user",
			request: request{
				method: http.MethodGet,
				ctx:    context.Background(),
			},
			usecaseGetUserURLsResponse: usecaseGetUserURLsResponse{
				userID:   contextUserID,
				userURLs: []app.ResponseUserURL{{OriginalURL: TestValidURL, ShortURL: TestHost + "/" + TestID}},
				err:      nil,
			},
			want: want{
				statusCode: http.StatusUnauthorized,
				body:       nil,
			},
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := mocks.NewMockAppUsecaseInterface(ctrl)
			m.EXPECT().GetUserURLs(tt.usecaseGetUserURLsResponse.userID).
				Return(
					tt.usecaseGetUserURLsResponse.userURLs,
					tt.usecaseGetUserURLsResponse.err,
				).AnyTimes()

			appHandler := NewAppHandler(m)

			req := httptest.NewRequest(tt.request.method, TestHost+"/api/user/urls", nil)
			req = req.WithContext(tt.request.ctx)

			w := httptest.NewRecorder()

			appHandler.APIGetUserURLs(w, req)

			res := w.Result()

			resBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			err = res.Body.Close()
			require.NoError(t, err)

			assert.Equal(t, tt.want.statusCode, res.StatusCode, "Invalid status code")

			if res.StatusCode == http.StatusCreated {
				assert.JSONEq(t, string(tt.want.body), string(resBody))
			}
		})
	}
}
