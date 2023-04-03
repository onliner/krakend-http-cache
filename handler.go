package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/google/uuid"

	"github.com/go-http-utils/fresh"
	"github.com/go-http-utils/headers"
)

type CacheHandler struct {
	client *http.Client
	logger Logger
}

type ClientConfig struct {
	Ttl  uint64
	Conn string `mapstructure:"connection"`
}

func NewCacheHandler(client *http.Client, logger Logger) *CacheHandler {
	return &CacheHandler{client, logger}
}

func (h *CacheHandler) Handle(cnf *ClientConfig) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		reqClone := cloneRequest(req)

		res := h.loadFromCache(reqClone, cnf)
		if res == nil {
			res = h.makeRequest(reqClone)

			if h.supportCaching(reqClone, res) {
				h.saveToCache(res, cnf)
			}
		}

		if fresh.IsFresh(req.Header, res.Header) {
			res.StatusCode = http.StatusNotModified
			res.Body = nil
		}

		h.writeResponse(w, res)
	})
}

func (h *CacheHandler) makeRequest(req *http.Request) *http.Response {
	h.logger.Debug(fmt.Sprintf("#%v", req.Header))
	res, err := h.client.Do(req)
	if err != nil {
		h.logger.Error(err)
		return &http.Response{StatusCode: http.StatusInternalServerError}
	}

	res.Body = io.NopCloser(ReusableReader(res.Body))

	return res
}

func (h *CacheHandler) loadFromCache(req *http.Request, cnf *ClientConfig) *http.Response {
	conn := GetCacheConn(cnf.Conn)
	if conn == nil {
		h.logger.Error(fmt.Sprintf("conn %s not found", cnf.Conn))
		return nil
	}

	v, err := conn.Fetch(cacheKey(req))
	if err != nil {
		return nil
	}

	buf := bufio.NewReader(bytes.NewReader([]byte(v)))
	r, err := http.ReadResponse(buf, nil)
	if err != nil {
		h.logger.Error("can't read from cache")
		return nil
	}

	return r
}

func (h *CacheHandler) saveToCache(res *http.Response, cnf *ClientConfig) {
	dump, err := httputil.DumpResponse(res, true)
	if err != nil {
		h.logger.Error("can't dump response")
		return
	}

	conn := GetCacheConn(cnf.Conn)
	if conn == nil {
		h.logger.Error(fmt.Sprintf("conn %s not found", cnf.Conn))
		return
	}

	err = conn.Save(cacheKey(res.Request), string(dump), time.Duration(cnf.Ttl)*time.Second)
	if err != nil {
		h.logger.Error(fmt.Sprintf("failed save to cache: %v", err))
	}
}

func (h *CacheHandler) supportCaching(req *http.Request, res *http.Response) bool {
	if req.Method != http.MethodGet {
		h.logger.Warning("can't non GET method for %s %s", req.Method, req.URL.RequestURI())
		return false
	}

	return res.StatusCode >= 200 && res.StatusCode <= 299
}

func (h *CacheHandler) writeResponse(w http.ResponseWriter, res *http.Response) {
	for k, hs := range res.Header {
		for _, h := range hs {
			w.Header().Add(k, h)
		}
	}

	w.WriteHeader(res.StatusCode)

	if res.Body == nil {
		return
	}

	_, err := io.Copy(w, res.Body)
	if err != nil {
		h.logger.Error(fmt.Sprintf("failed write response body: %v", err))
	}

	res.Body.Close()
}

func cloneRequest(req *http.Request) *http.Request {
	clone := req.Clone(req.Context())
	clone.Header.Del(headers.IfModifiedSince)
	clone.Header.Del(headers.IfUnmodifiedSince)
	clone.Header.Del(headers.IfNoneMatch)
	clone.Header.Del(headers.IfMatch)
	clone.Header.Del(headers.CacheControl)
	clone.Header.Del(headers.AcceptEncoding)

	return clone
}

func cacheKey(req *http.Request) string {
	url := req.URL.RequestURI()

	return fmt.Sprintf("krakend-hc:%s", uuid.NewSHA1(uuid.NameSpaceURL, []byte(url)))
}
