package negronicache

import (
    "net/http"
    "testing"

    "github.com/stretchr/testify/assert"
)

func MockCacheableRequest(t *testing.T) *CacheRequest {
    req, err := http.NewRequest("GET", "http://example.com", nil)
    assert.Nil(t, err)

    req.RequestURI = "http://example.com"
    req.Method = "GET"
    req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("ETag", "15f0fff99ed5aae4edffdd6496d7131f")
    cReq, err := NewCacheRequest(req)
    assert.Nil(t, err)

    return cReq
}

func MockNoCacheRequest(t *testing.T) *CacheRequest {
    req, err := http.NewRequest("GET", "http://example.com", nil)
    assert.Nil(t, err)

    req.RequestURI = "http://example.com"
    req.Method = "GET"
    req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("ETag", "15f0fff99ed5aae4edffdd6496d7131f")
    req.Header.Set("Cache-Control", "no-cache")
    cReq, err := NewCacheRequest(req)
    assert.Nil(t, err)

    return cReq
}

func MockNoStoreRequest(t *testing.T) *CacheRequest {
    req, err := http.NewRequest("GET", "http://example.com", nil)
    assert.Nil(t, err)

    req.RequestURI = "http://example.com"
    req.Method = "GET"
    req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("ETag", "15f0fff99ed5aae4edffdd6496d7131f")
    req.Header.Set("Cache-Control", "no-store")
    cReq, err := NewCacheRequest(req)
    assert.Nil(t, err)

    return cReq
}

func MockZeroMaxAgeRequest(t *testing.T) *CacheRequest {
    req, err := http.NewRequest("GET", "http://example.com", nil)
    assert.Nil(t, err)

    req.RequestURI = "http://example.com"
    req.Method = "GET"
    req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("ETag", "15f0fff99ed5aae4edffdd6496d7131f")
    req.Header.Set("Cache-Control", "max-age=0")
    cReq, err := NewCacheRequest(req)
    assert.Nil(t, err)

    return cReq
}

func MockEmptyHostRequest(t *testing.T) (*CacheRequest, error) {
    req, err := http.NewRequest("GET", "http://example.com", nil)
    assert.Nil(t, err)

    req.Host = ""
    cReq, err := NewCacheRequest(req)

    return cReq, err
}

func TestRequest_NewCacheRequest(t *testing.T) {
    cReq := MockCacheableRequest(t)
    assert.NotNil(t, cReq.Key)

    // Assert when host is empty
    cReqErr, err := MockEmptyHostRequest(t)
    assert.NotNil(t, err)
    assert.Nil(t, cReqErr)
}

func TestRequest_IsCacheable(t *testing.T) {
    cReq := MockCacheableRequest(t)

    // HTTP method is POST
    cReq.Method = "POST"
    isCacheableRequest := cReq.isCacheable()
    assert.False(t, isCacheableRequest)

    // HTTP method is DELETE
    cReq.Method = "DELETE"
    isCacheableRequest = cReq.isCacheable()
    assert.False(t, isCacheableRequest)
}

func TestRequest_IsCacheableIfMatch(t *testing.T) {
    cReq := MockCacheableRequest(t)

    // Headers contain "If-Match"
    cReq.Method = "GET"
    cReq.Header.Set("If-Match", "15f0fff99ed5aae4edffdd6496d7131f")
    isCacheableRequest := cReq.isCacheable()
    assert.False(t, isCacheableRequest)
}

func TestRequest_IsCacheableIfRange(t *testing.T) {
    cReq := MockCacheableRequest(t)

    // Headers contain "If-Range"
    cReq.Method = "GET"
    cReq.Header.Set("If-Range", "A023EF02BD589BC472A2D6774EAE3C58")
    isCacheableRequest := cReq.isCacheable()
    assert.False(t, isCacheableRequest)
}

func TestRequest_IsCacheableZeroMaxAge(t *testing.T) {
    cReq := MockZeroMaxAgeRequest(t)

    // Headers contain "max-age" and it equals to 0
    isCacheableRequest := cReq.isCacheable()
    assert.False(t, isCacheableRequest)
}

func TestRequest_IsCacheableWithNoCache(t *testing.T) {
    cReq := MockNoCacheRequest(t)

    // Headers contain "no-cache"
    isCacheableRequest := cReq.isCacheable()
    assert.False(t, isCacheableRequest)
}

func TestRequest_IsCacheableWithNoStore(t *testing.T) {
    cReq := MockNoStoreRequest(t)

    // Headers contain "no-store"
    isCacheableRequest := cReq.isCacheable()
    assert.False(t, isCacheableRequest)
}
