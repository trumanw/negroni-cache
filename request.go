package negronicache

import (
    "time"
    "errors"
    "net/http"
)

// CacheRequest extended http.Request to include more cache associated parameters.
type CacheRequest struct {
    *http.Request
	Key          Key
	Time         time.Time
	CacheControl CacheControl
}

// NewCacheRequest constructs an instance of CacheRequest
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

// isStateChanging returns true if HTTP methods are POST, PUT and DELETE
func (r *CacheRequest) isStateChanging() bool {
    if !(r.Method == "POST" || r.Method == "PUT" || r.Method == "DELETE") {
		return true
	}

	return false
}

// isCacheable returns true when request can be cacheable
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
