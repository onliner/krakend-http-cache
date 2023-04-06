package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-http-utils/headers"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

const body string = `{"message": "Hello Wordl"}`
const conn string = "memory"

func setup() {
	httpmock.Activate()
	if GetCache(conn) == nil {
		if err := RegisterCache(conn, &CacheCnf{Driver: "memory"}); err != nil {
			println(err)
		}
	}
}

func teardown() {
	httpmock.DeactivateAndReset()
	GetCache(conn).Flush()
}

func newHandler() http.Handler {
	return NewCacheHandler(http.DefaultClient, noopLogger{}).Handle(&ClientConfig{Ttl: 1, Conn: conn})
}

func headerFromMap(input map[string]string) http.Header {
	headers := http.Header{}

	for key, value := range input {
		headers.Set(key, value)
	}

	return headers
}

func newRequest(method string, headers map[string]string) *http.Request {
	req, _ := http.NewRequest(method, "/test", nil)
	req.Header = headerFromMap(headers)

	return req
}

func registerResponse(method string, status int, headers map[string]string) {
	responder := httpmock.NewStringResponder(status, body).HeaderSet(headerFromMap(headers))
	httpmock.RegisterResponder(method, "/test", responder)
}

func testClient(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	newHandler().ServeHTTP(rr, req)

	return rr
}

func TestHandle(t *testing.T) {
	setup()
	defer teardown()

	requests := []struct {
		method    string
		status    int
		callCount int
	}{
		{http.MethodGet, http.StatusContinue, 2},
		{http.MethodGet, http.StatusOK, 1},
		{http.MethodGet, http.StatusNotModified, 2},
		{http.MethodGet, http.StatusBadRequest, 2},
		{http.MethodGet, http.StatusNotFound, 2},
		{http.MethodGet, http.StatusInternalServerError, 2},
		{http.MethodGet, http.StatusServiceUnavailable, 2},
	}

	for _, request := range requests {
		req := newRequest(request.method, nil)
		registerResponse(request.method, request.status, nil)

		rr := testClient(req)
		assert.Equal(t, request.status, rr.Result().StatusCode, "Status code mismatch")
		assert.Equal(t, body, rr.Body.String(), "Body mismatch")

		rr = testClient(req)
		assert.Equal(t, request.status, rr.Result().StatusCode, "Status code mismatch")
		assert.Equal(t, body, rr.Body.String(), "Body mismatch")
		assert.Equal(t, request.callCount, httpmock.GetTotalCallCount())

		httpmock.Reset()
		GetCache(conn).Flush()
	}
}

func TestHandleNotSupportedMethods(t *testing.T) {
	methods := []string{http.MethodPost, http.MethodPut}
	for _, method := range methods {
		req := newRequest(method, nil)
		registerResponse(method, http.StatusOK, nil)

		rr := testClient(req)
		assert.Equal(t, http.StatusNotImplemented, rr.Result().StatusCode, "Status code mismatch")
		assert.Empty(t, rr.Body.String(), "Body mismatch")
	}
}

func TestHandleStale(t *testing.T) {
	setup()
	defer teardown()

	req := newRequest(http.MethodGet, nil)
	registerResponse(http.MethodGet, http.StatusOK, nil)

	rr := testClient(req)
	assert.Equal(t, http.StatusOK, rr.Result().StatusCode, "Status code mismatch")
	assert.Equal(t, body, rr.Body.String(), "Body mismatch")

	GetCache(conn).Flush()

	rr = testClient(req)
	assert.Equal(t, http.StatusOK, rr.Result().StatusCode, "Status code mismatch")
	assert.Equal(t, body, rr.Body.String(), "Body mismatch")

	assert.Equal(t, 2, httpmock.GetTotalCallCount())
}

func TestHandleEtag(t *testing.T) {
	setup()
	defer teardown()

	req := newRequest(http.MethodGet, map[string]string{headers.IfNoneMatch: `W/"foo"`})
	registerResponse(http.MethodGet, http.StatusOK, map[string]string{headers.ETag: `W/"foo"`})

	rr := testClient(req)

	assert.Equal(t, http.StatusNotModified, rr.Result().StatusCode, "Status code mismatch")
	assert.Equal(t, "", rr.Body.String(), "Body mismatch")

	req = newRequest(http.MethodGet, map[string]string{headers.IfNoneMatch: `W/"bar"`})
	rr = testClient(req)

	assert.Equal(t, http.StatusOK, rr.Result().StatusCode, "Status code mismatch")
	assert.Equal(t, body, rr.Body.String(), "Body mismatch")

	assert.Equal(t, 1, httpmock.GetTotalCallCount())
}
