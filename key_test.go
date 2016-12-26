package negronicache

import (
    "testing"
    "net/http"

    "github.com/stretchr/testify/assert"
)

func MockHTTPRequest(t *testing.T) *http.Request {
    req, err := http.NewRequest("GET", "https://api.example.com/me", nil)
    assert.Nil(t, err)

    req.RequestURI = "https://api.example.com/me"
    req.Method = "GET"
    req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("ETag", "15f0fff99ed5aae4edffdd6496d7131f")
    assert.Nil(t, err)

    return req
}

func MockVaryHTTPRequest(t *testing.T) *http.Request {
    req := MockHTTPRequest(t)
    req.Header.Set("ETag", "33a64df551425fcc55e4d42a148795d9")

    return req
}

func TestKey_NewRequestKey(t *testing.T) {
    // Mock a HTTP request
    r := MockHTTPRequest(t)

    // Init a caching key with HTTP request
    key := NewRequestKey(r)

    assert.NotNil(t, key)
    assert.Equal(t, "GET", key.method)

    // Parsing URL
    assert.Equal(t, "https://api.example.com/me", key.u.String())
}

func TestKey_NewRequestKeyWithContentLocation(t *testing.T) {
    // Mock a HTTP request with Content-Location
    r := MockHTTPRequest(t)
    r.Header.Set("Content-Location", "/me")

    // Init a caching key with HTTP request
    key := NewRequestKey(r)

    assert.NotNil(t, key)
    assert.Equal(t, "GET", key.method)

    // Parsing URL
    assert.Equal(t, "https://api.example.com/me", key.u.String())

    // Update the Content-Location to be other string instead of host
    r.Header.Set("Content-Location", "https://api.example.com/profile")
    // Init a caching key with HTTP request
    key = NewRequestKey(r)

    assert.NotNil(t, key)
    assert.Equal(t, "GET", key.method)
}

func TestKey_NewRequestKeyWithDiffContentLocation(t *testing.T) {
    // Mock a HTTP request with Content-Location
    r := MockHTTPRequest(t)
    r.Header.Set("Content-Location", "https://apis.example.com/me")

    // Init a caching key with HTTP request
    key := NewRequestKey(r)

    assert.NotNil(t, key)
    assert.Equal(t, "GET", key.method)
}

func TestKey_ForMethod(t *testing.T) {
    // Mock a HTTP request with Content-Location
    r := MockHTTPRequest(t)

    // Init a caching key with HTTP request
    key := NewRequestKey(r)
    newKey := key.ForMethod("POST")
    assert.Equal(t, "POST", newKey.method)
}

func TestKey_Vary(t *testing.T) {
    // Mock a HTTP request with Content-Location
    r := MockHTTPRequest(t)

    // Init a caching key with HTTP request
    key := NewRequestKey(r)
    varyReq := MockVaryHTTPRequest(t)

    assert.Empty(t, key.vary)
    varyKey := key.Vary("ETag", varyReq)
    assert.Equal(t, "ETag=33a64df551425fcc55e4d42a148795d9", varyKey.vary[0])
}
