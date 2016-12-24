package negronicache

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

var tc = NewMemoryCache()
const (
    testKey = "GET:http://test.com"
)

func TestCache_NewMemoryCache(t *testing.T) {
	c := NewMemoryCache()
	assert.NotNil(t, c)
}

func TestCache_NewDiskCache(t *testing.T) {
	c, err := NewDiskCache("./cache")
	assert.Nil(t, err)
	assert.NotNil(t, c)
}

func TestCache_Store(t *testing.T) {
	// New a resource for caching
	h := make(http.Header)
	h.Set(CacheHeader, "SKIP")

	res := NewResourceBytes(404, nil, h)
	body := []byte("Here is the body for testing.")
	res.ReadSeekCloser = &ByteReadSeekCloser{bytes.NewReader(body)}

	// Store the resource
	err := tc.Store(res, testKey)
	assert.Nil(t, err)
}

func TestCache_Retrieve(t *testing.T) {
    res, err := tc.Retrieve(testKey)
	assert.Nil(t, err)

	assert.Equal(t, 404, res.statusCode)
}

func TestCache_Header(t *testing.T) {
	h, err := tc.Header(testKey)
	assert.Nil(t, err)

	assert.Equal(t, "SKIP", h.Get(CacheHeader))
}
