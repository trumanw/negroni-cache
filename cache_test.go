package negronicache

func TestCache_NewMemoryCache(t testing.T) {
    c := NewMemoryCache()
    assert.NotNil(t, c.fs)
    assert.NotNil(t, c.stale)
}

func TestCache_NewDiskCache(t testing.T) {
    c, err := NewDiskCache("./cache")
    assert.Nil(t, err)

    assert.NotNil(t, c.fs)
    assert.NotNil(t, c.stale)
}
