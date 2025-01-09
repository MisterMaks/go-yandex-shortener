package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"

	"github.com/MisterMaks/go-yandex-shortener/internal/app"
	appDeliveryInternal "github.com/MisterMaks/go-yandex-shortener/internal/app/delivery"
	"github.com/MisterMaks/go-yandex-shortener/internal/app/delivery/mocks"
	"github.com/MisterMaks/go-yandex-shortener/internal/gzip"
	"github.com/MisterMaks/go-yandex-shortener/internal/logger"
	"github.com/MisterMaks/go-yandex-shortener/internal/user/usecase"
	"github.com/golang/mock/gomock"
)

func newExampleServer(m *mocks.MockAppUsecaseInterface) (*httptest.Server, error) {
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
		Authenticate: func(h http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				ctx := context.WithValue(r.Context(), usecase.UserIDKey, TestUserID)
				h.ServeHTTP(w, r.WithContext(ctx))
			})
		},
		TrustedSubnetMiddleware: func(h http.Handler) http.Handler { return h },
	}

	u, err := url.ParseRequestURI(ResultAddrPrefix)
	if err != nil {
		return nil, err
	}

	r, err := shortenerRouter(appHandler, u, middlewares)
	if err != nil {
		return nil, err
	}

	ts := httptest.NewServer(r)

	client := ts.Client()
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	return ts, nil
}

func newExampleMock() *mocks.MockAppUsecaseInterface {
	ctrl := gomock.NewController(nil)

	m := mocks.NewMockAppUsecaseInterface(ctrl)

	m.EXPECT().GetOrCreateURL(TestValidURL, TestUserID).Return(&app.URL{
		ID:        TestID,
		URL:       TestValidURL,
		UserID:    TestUserID,
		IsDeleted: false,
	}, false, nil).AnyTimes()
	m.EXPECT().GenerateShortURL(TestID).Return("http://localhost:8080/" + TestID).AnyTimes()
	m.EXPECT().GetURL(TestID).Return(&app.URL{
		ID:        TestID,
		URL:       TestValidURL,
		UserID:    TestUserID,
		IsDeleted: false,
	}, nil)
	m.EXPECT().GetOrCreateURLs([]app.RequestBatchURL{
		{CorrelationID: TestID, OriginalURL: TestValidURL},
	}, TestUserID).Return([]app.ResponseBatchURL{
		{CorrelationID: TestID, ShortURL: "http://localhost:8080/" + TestID},
	}, nil).AnyTimes()
	m.EXPECT().GetUserURLs(TestUserID).Return([]app.ResponseUserURL{
		{ShortURL: "http://localhost:8080/" + TestID, OriginalURL: TestValidURL},
	}, nil).AnyTimes()
	m.EXPECT().SendDeleteUserURLsInChan(TestUserID, []string{TestID}).AnyTimes()
	m.EXPECT().Ping().Return(nil).AnyTimes()

	return m
}

func Example() {
	// preparing mocks and server
	m := newExampleMock()
	ts, err := newExampleServer(m)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer ts.Close()

	serverAddr := ts.URL

	fmt.Println("Get or create URL:")
	getOrCreateURLURL := serverAddr + "/"
	req, err := http.NewRequest(http.MethodPost, getOrCreateURLURL, bytes.NewReader([]byte(TestValidURL)))
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	req.Header.Set(ContentTypeKey, TextPlainKey)
	resp, err := ts.Client().Do(req)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	if err = resp.Body.Close(); err != nil {
		fmt.Println("Error:", err)
		return
	}

	respBodyStr := string(respBody)
	fmt.Println(resp.StatusCode)
	fmt.Println(respBodyStr)

	fmt.Println("Redirect to original URL:")
	redirectURL := serverAddr + "/" + TestID
	req, err = http.NewRequest(http.MethodGet, redirectURL, bytes.NewReader(nil))
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	resp, err = ts.Client().Do(req)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	respBody, err = io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	if err = resp.Body.Close(); err != nil {
		fmt.Println("Error:", err)
		return
	}

	respBodyStr, _ = strings.CutSuffix(string(respBody), "\n\n")
	fmt.Println(resp.StatusCode)
	fmt.Println(respBodyStr)

	fmt.Println("Get or create URL in JSON format:")
	apiGetOrCreateURLURL := serverAddr + "/api/shorten"
	bodyJSON := []byte(`{"url": "` + TestValidURL + `"}`)
	req, err = http.NewRequest(http.MethodPost, apiGetOrCreateURLURL, bytes.NewReader(bodyJSON))
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	req.Header.Set(ContentTypeKey, appDeliveryInternal.ApplicationJSONKey)
	resp, err = ts.Client().Do(req)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	respBody, err = io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	if err = resp.Body.Close(); err != nil {
		fmt.Println("Error:", err)
		return
	}

	respBodyStr = string(respBody)
	fmt.Println(resp.StatusCode)
	fmt.Println(respBodyStr)

	fmt.Println("Get or create URLs in JSON format:")
	apiGetOrCreateURLsURL := serverAddr + "/api/shorten/batch"
	bodyJSON = []byte(`[{"correlation_id": "` + TestID + `", "original_url": "` + TestValidURL + `"}]`)
	req, err = http.NewRequest(http.MethodPost, apiGetOrCreateURLsURL, bytes.NewReader(bodyJSON))
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	req.Header.Set(ContentTypeKey, appDeliveryInternal.ApplicationJSONKey)
	resp, err = ts.Client().Do(req)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	respBody, err = io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	if err = resp.Body.Close(); err != nil {
		fmt.Println("Error:", err)
		return
	}

	respBodyStr = string(respBody)
	fmt.Println(resp.StatusCode)
	fmt.Println(respBodyStr)

	fmt.Println("Get user URLs in JSON format:")
	apiGetUserURLsURL := serverAddr + "/api/user/urls"
	req, err = http.NewRequest(http.MethodGet, apiGetUserURLsURL, bytes.NewReader(nil))
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	resp, err = ts.Client().Do(req)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	respBody, err = io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	if err = resp.Body.Close(); err != nil {
		fmt.Println("Error:", err)
		return
	}

	respBodyStr = string(respBody)
	fmt.Println(resp.StatusCode)
	fmt.Println(respBodyStr)

	fmt.Println("Delete user URLs:")
	apiDeleteUserURLsURL := serverAddr + "/api/user/urls"
	bodyJSON = []byte(`["` + TestID + `"]`)
	req, err = http.NewRequest(http.MethodDelete, apiDeleteUserURLsURL, bytes.NewReader(bodyJSON))
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	resp, err = ts.Client().Do(req)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	if err = resp.Body.Close(); err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println(resp.StatusCode)

	fmt.Println("Ping DB:")
	pingURL := serverAddr + "/ping"
	req, err = http.NewRequest(http.MethodGet, pingURL, bytes.NewReader(nil))
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	resp, err = ts.Client().Do(req)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	if err = resp.Body.Close(); err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println(resp.StatusCode)

	// Output:
	// Get or create URL:
	// 201
	// http://localhost:8080/1
	// Redirect to original URL:
	// 307
	// <a href="http://valid_url.ru/">Temporary Redirect</a>.
	// Get or create URL in JSON format:
	// 201
	// {"result":"http://localhost:8080/1"}
	//
	// Get or create URLs in JSON format:
	// 201
	// [{"correlation_id":"1","short_url":"http://localhost:8080/1"}]
	//
	// Get user URLs in JSON format:
	// 200
	// [{"short_url":"http://localhost:8080/1","original_url":"http://valid_url.ru/"}]
	//
	// Delete user URLs:
	// 202
	// Ping DB:
	// 200
}
