package middleware_test

import (
	"bytes"
	"crypto/rand"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/medibloc/panacea-oracle/server/middleware"
	"github.com/stretchr/testify/require"
)

func TestBodySizeSmallerThanLimitSetting(t *testing.T) {
	testLimitMiddlewareHTTPRequest(
		t,
		newRequest(newRandomBody(1023)),
		1024,
		http.StatusOK,
		"",
	)
}

func TestBodySizeSameLimitSetting(t *testing.T) {
	testLimitMiddlewareHTTPRequest(
		t,
		newRequest(newRandomBody(1024)),
		1024,
		http.StatusOK,
		"",
	)
}

func TestBodySizeLargeThanLimitSetting(t *testing.T) {
	testLimitMiddlewareHTTPRequest(
		t,
		newRequest(newRandomBody(1025)),
		1024,
		http.StatusBadRequest,
		"request body too large",
	)
}

func TestDifferentBodySizeAndHeaderContentSize(t *testing.T) {
	req := newRequest(newRandomBody(1025))
	req.ContentLength = 1024

	testLimitMiddlewareHTTPRequest(
		t,
		req,
		1024,
		http.StatusBadRequest,
		"request body too large",
	)
}

func newRandomBody(size int) []byte {
	body := make([]byte, size)
	if _, err := rand.Read(body); err != nil {
		panic(err)
	}

	return body
}

func newRequest(body []byte) *http.Request {
	return httptest.NewRequest(
		"GET",
		"http://test.com",
		bytes.NewBuffer(body),
	)
}

func testLimitMiddlewareHTTPRequest(
	t *testing.T,
	req *http.Request,
	maxRequestBodySize int64,
	statusCode int,
	bodyMsg string,
) {
	w := httptest.NewRecorder()
	mw := middleware.NewLimitMiddleware(maxRequestBodySize)
	testHandler := mw.Middleware(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
			}
		}),
	)

	testHandler.ServeHTTP(w, req)

	resp := w.Result()
	require.Equal(t, statusCode, resp.StatusCode)
	if bodyMsg != "" {
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Contains(t, string(body), bodyMsg)
	}
}
