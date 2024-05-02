package delivery

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	app "github.com/MisterMaks/go-yandex-shortener/internal/app"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	TestValidURL   string = "valid_url"
	TestInvalidURL string = "invalid_url"
	TestID         string = "1"
	TestHost       string = "http://example.com"
)

var (
	ErrTestInvalidURL = errors.New("invalid url")
	ErrTestIDNotFound = errors.New("ID not found")
)

type testAppUsecase struct{}

func (tau *testAppUsecase) GetOrCreateURL(rawURL string) (*app.URL, error) {
	switch rawURL {
	case TestValidURL:
		return &app.URL{
			ID:  TestID,
			URL: TestValidURL,
		}, nil
	case TestInvalidURL:
		return nil, ErrTestInvalidURL
	}
	return nil, ErrTestInvalidURL
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

func (tau *testAppUsecase) GenerateShortURL(addr, id string) string {
	return "http://" + addr + "/" + id
}

func TestAppHandler_GetOrCreateURL(t *testing.T) {
	type want struct {
		code        int
		response    string
		contentType string
	}
	type request struct {
		method      string
		contentType string
		path        string
		body        []byte
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
				path:        TestHost + "/",
				body:        []byte(TestValidURL),
			},
			want: want{
				code:        http.StatusCreated,
				response:    TestHost + "/" + TestID,
				contentType: TextPlainKey,
			},
		},
		{
			name: "invalid URL",
			request: request{
				method:      http.MethodPost,
				contentType: TextPlainKey,
				path:        TestHost + "/",
				body:        []byte(TestInvalidURL),
			},
			want: want{
				code:        http.StatusBadRequest,
				response:    "",
				contentType: "",
			},
		},
		{
			name: "invalid method",
			request: request{
				method:      http.MethodGet,
				contentType: TextPlainKey,
				path:        TestHost + "/",
				body:        []byte(TestValidURL),
			},
			want: want{
				code:        http.StatusMethodNotAllowed,
				response:    "",
				contentType: "",
			},
		},
		{
			name: "invalid Header Content-Type",
			request: request{
				method:      http.MethodPost,
				contentType: "invalid Header Content-Type",
				path:        TestHost + "/",
				body:        []byte(TestValidURL),
			},
			want: want{
				code:        http.StatusBadRequest,
				response:    "",
				contentType: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tau := &testAppUsecase{}
			appHandler := NewAppHandler(tau)

			bodyReader := bytes.NewReader(tt.request.body)

			req := httptest.NewRequest(tt.request.method, tt.request.path, bodyReader)
			req.Header.Add(ContentTypeKey, tt.request.contentType)
			w := httptest.NewRecorder()

			appHandler.GetOrCreateURL(w, req)

			res := w.Result()

			assert.Equal(t, tt.want.code, res.StatusCode)
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

func TestAppHandler_RedirectToURL(t *testing.T) {
	type want struct {
		code     int
		response string
	}
	type request struct {
		method string
		path   string
		id     string
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
				path:   TestHost + "/",
				id:     TestID,
			},
			want: want{
				code:     http.StatusTemporaryRedirect,
				response: "<a href=\"/" + TestValidURL + "\">Temporary Redirect</a>.\n\n",
			},
		},
		{
			name: "invalid ID",
			request: request{
				method: http.MethodGet,
				path:   TestHost + "/",
				id:     "2",
			},
			want: want{
				code:     http.StatusBadRequest,
				response: "",
			},
		},
		{
			name: "invalid method",
			request: request{
				method: http.MethodPost,
				path:   TestHost + "/",
				id:     TestID,
			},
			want: want{
				code:     http.StatusMethodNotAllowed,
				response: "",
			},
		},
		{
			name: "invalid path",
			request: request{
				method: http.MethodGet,
				path:   TestHost + "/",
				id:     "",
			},
			want: want{
				code:     http.StatusBadRequest,
				response: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tau := &testAppUsecase{}
			appHandler := NewAppHandler(tau)

			req := httptest.NewRequest(tt.request.method, tt.request.path, nil)
			req.SetPathValue("id", tt.request.id)
			w := httptest.NewRecorder()

			appHandler.RedirectToURL(w, req)

			res := w.Result()

			assert.Equal(t, tt.want.code, res.StatusCode)
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
