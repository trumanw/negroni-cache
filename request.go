package negronicache

import (
    "time"
    "errors"
    "net/http"
)

type CacheRequest struct {
    *http.Request
	Key          Key
	Time         time.Time
	CacheControl CacheControl
}

func NewCacheRequest(r *http.Request) (*CacheRequest, error) {
	cc, err := ParseCacheControl(r.Header.Get("Cache-Control"))
	if err != nil {
		return nil, err
	}

	if r.Proto == "HTTP/1.1" && r.Host == "" {
		return nil, errors.New("Host header can't be empty")
	}

	return &CacheRequest{
		Request:      r,
		Key:          NewRequestKey(r),
		Time:         Clock(),
		CacheControl: cc,
	}, nil
}

func (r *CacheRequest) isCacheable() bool {
	if !(r.Method == "GET" || r.Method == "HEAD") {
		return false
	}

	if r.Header.Get("If-Match") != "" ||
		r.Header.Get("If-Unmodified-Since") != "" ||
		r.Header.Get("If-Range") != "" {
		return false
	}

	if maxAge, ok := r.CacheControl.Get("max-age"); ok && maxAge == "0" {
		return false
	}

	if r.CacheControl.Has("no-store") || r.CacheControl.Has("no-cache") {
		return false
	}

	return true
}
