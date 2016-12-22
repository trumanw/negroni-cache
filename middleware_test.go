package negronicache

import (
    "testing"
    "net/http"
	"net/http/httptest"

    "github.com/urfave/negroni"
	"github.com/stretchr/testify/assert"
)

func setupServeHTTP(t *testing.T) (negroni.ResponseWriter, *http.Request) {
    req, err := http.NewRequest("GET", "http://example.com/stuff?rly=ya", nil)
    assert.Nil(t, err)

    req.RequestURI = "http://example.com/stuff?rly=ya"
    req.Method = "GET"
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Accept", "application/json")
    req.Header.Set("ETag", "15f0fff99ed5aae4edffdd6496d7131f")

    return negroni.NewResponseWriter(httptest.NewRecorder()), req
}

func TestMiddleware_ServeHTTP(t *testing.T) {
    mw := NewMiddleware(NewMemoryCache())

    recNoCache, reqNoCache := setupServeHTTP(t)
    mw.ServeHTTP(recNoCache, reqNoCache, func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(200)
    })
    assert.Equal(t, recNoCache.Status(), 200)
    assert.Equal(t, recNoCache.Header().Get(CacheHeader), "SKIP")

    recCache, reqCache := setupServeHTTP(t)
    mw.ServeHTTP(recCache, reqCache, func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(200)
    })
    assert.Equal(t, recCache.Status(), 200)
    assert.Equal(t, recCache.Header().Get(CacheHeader), "HIT")
}
