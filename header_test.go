package negronicache

import (
    "net/http"
    "testing"

    "github.com/stretchr/testify/assert"
)

func TestHeader_TimeHeader(t *testing.T) {
    h := make(http.Header)
    h.Add("Date", "Mon, 02 Jan 2006 15:04:05 GMT")

    date, err := timeHeader("Date", h)
    assert.Nil(t, err)

    assert.Equal(t, 2006, date.Year())
    assert.Equal(t, "January", date.Month().String())
    assert.Equal(t, 2, date.Day())
}

func TestHeader_IntHeader(t *testing.T) {
    h := make(http.Header)
    h.Add("Age", "65537")

    age, err := intHeader("Age", h)
    assert.Nil(t, err)

    assert.Equal(t, 65537, age)
}
