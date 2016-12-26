package negronicache

import (
    "testing"

    "github.com/stretchr/testify/assert"
)

func MockCacheControl(t *testing.T) CacheControl {
    cacheControlString := "max-age=0,no-cache,no-store"
    cacheControl, err := ParseCacheControl(cacheControlString)
    assert.Nil(t, err)

    return cacheControl
}

func TestControl_String(t *testing.T) {
    cacheControl := MockCacheControl(t)
    cacheControlString := cacheControl.String()
    assert.NotNil(t, cacheControlString)
}
