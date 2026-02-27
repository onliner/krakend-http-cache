package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	uh "github.com/go-http-utils/headers"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

const body string = `{"message": "Hello World"}`
const conn string = "memory"

func setup(t *testing.T) {
	httpmock.Activate()
	if GetCache(conn) == nil {
		err := RegisterCache(conn, &CacheCnf{Driver: "memory"})
		assert.NoError(t, err)
	}
}

func teardown(t *testing.T) {
	httpmock.DeactivateAndReset()
	err := GetCache(conn).Flush()
	assert.NoError(t, err)
}

func newHandler() http.Handler {
	cnf, _ := NewClientConfig(map[string]interface{}{
		"onliner/krakend-http-cache": map[string]interface{}{
			"ttl":        1,
			"connection": "memory",
		},
	})

	return NewCacheHandler(http.DefaultClient, noopLogger{}).Handle(cnf)
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
	setup(t)
	defer teardown(t)

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
		err := GetCache(conn).Flush()
		assert.NoError(t, err)
	}
}

func TestHandleNotSupportedMethods(t *testing.T) {
	methods := []string{http.MethodHead, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete}
	for _, method := range methods {
		req := newRequest(method, nil)
		registerResponse(method, http.StatusOK, nil)

		rr := testClient(req)
		assert.Equal(t, http.StatusNotImplemented, rr.Result().StatusCode, "Status code mismatch")
		assert.Empty(t, rr.Body.String(), "Body mismatch")
	}
}

func TestHandleStale(t *testing.T) {
	setup(t)
	defer teardown(t)

	req := newRequest(http.MethodGet, nil)
	registerResponse(http.MethodGet, http.StatusOK, nil)

	rr := testClient(req)
	assert.Equal(t, http.StatusOK, rr.Result().StatusCode, "Status code mismatch")
	assert.Equal(t, body, rr.Body.String(), "Body mismatch")

	err := GetCache(conn).Flush()
	assert.NoError(t, err)

	rr = testClient(req)
	assert.Equal(t, http.StatusOK, rr.Result().StatusCode, "Status code mismatch")
	assert.Equal(t, body, rr.Body.String(), "Body mismatch")

	assert.Equal(t, 2, httpmock.GetTotalCallCount())
}

func TestHandleEtag(t *testing.T) {
	setup(t)
	defer teardown(t)

	req := newRequest(http.MethodGet, map[string]string{uh.IfNoneMatch: `W/"foo"`})
	registerResponse(http.MethodGet, http.StatusOK, map[string]string{uh.ETag: `W/"foo"`})

	rr := testClient(req)

	assert.Equal(t, http.StatusNotModified, rr.Result().StatusCode, "Status code mismatch")
	assert.Equal(t, "", rr.Body.String(), "Body mismatch")

	req = newRequest(http.MethodGet, map[string]string{uh.IfNoneMatch: `W/"bar"`})
	rr = testClient(req)

	assert.Equal(t, http.StatusOK, rr.Result().StatusCode, "Status code mismatch")
	assert.Equal(t, body, rr.Body.String(), "Body mismatch")

	assert.Equal(t, 1, httpmock.GetTotalCallCount())
}

func TestHandleHeaders(t *testing.T) {
	setup(t)
	defer teardown(t)

	requests := []struct {
		headers1  map[string]string
		headers2  map[string]string
		callCount int
	}{
		{map[string]string{"X-Custom-Header": "1"}, map[string]string{"x-custom-header": "1"}, 1},
		{map[string]string{"X-Custom-Header": "1"}, map[string]string{"X-Custom-Header": "2"}, 2},
		{map[string]string{"X-Custom-Header": "1"}, map[string]string{"X-Custom-Header": ""}, 2},
		{map[string]string{"X-Custom-Header": "1"}, nil, 2},
	}

	cnf, _ := NewClientConfig(map[string]interface{}{
		"onliner/krakend-http-cache": map[string]interface{}{
			"ttl":        1,
			"connection": "memory",
			"headers":    []string{"X-Custom-Header"},
		},
	})

	handler := NewCacheHandler(http.DefaultClient, noopLogger{}).Handle(cnf)

	for _, request := range requests {
		req := newRequest(http.MethodGet, request.headers1)
		registerResponse(http.MethodGet, http.StatusOK, nil)

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Result().StatusCode, "Status code mismatch")
		assert.Equal(t, body, rr.Body.String(), "Body mismatch")

		req = newRequest(http.MethodGet, request.headers2)
		rr = httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Result().StatusCode, "Status code mismatch")
		assert.Equal(t, body, rr.Body.String(), "Body mismatch")

		assert.Equal(t, request.callCount, httpmock.GetTotalCallCount())
		httpmock.Reset()
		err := GetCache(conn).Flush()
		assert.NoError(t, err)
	}
}
