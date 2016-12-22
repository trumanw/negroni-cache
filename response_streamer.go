package negronicache

import (
    "net/http"
    "io/ioutil"

    "gopkg.in/djherbis/stream.v1"
)

type ResponseStreamer struct {
    StatusCode int
    http.ResponseWriter
    *stream.Stream
    // C will be closed by WriteHeader to signal the headers' writing
    C chan struct{}
}

func NewResponseStreamer(w http.ResponseWriter) *ResponseStreamer {
	strm, err := stream.NewStream("responseBuffer", stream.NewMemFS())
	if err != nil {
		panic(err)
	}
	return &ResponseStreamer{
		ResponseWriter: w,
		Stream:         strm,
		C:              make(chan struct{}),
	}
}

// WaitHeaders returns iff and when WriteHeader has been called.
func (rs *ResponseStreamer) WaitHeaders() {
    for range rs.C {
    }
}

func(rs *ResponseStreamer) WriteHeader(status int) {
    defer close(rs.C)
    rs.StatusCode = status
    rs.ResponseWriter.WriteHeader(status)
}

func(rs *ResponseStreamer) Write(b []byte) (int, error) {
    rs.Stream.Write(b)
    return rs.ResponseWriter.Write(b)
}

func(rs *ResponseStreamer) Close() error {
    return rs.Stream.Close()
}

// Resource returns a copy of the responseStreamer as a Resource object
func (rs *ResponseStreamer) Resource() *Resource {
	r, err := rs.Stream.NextReader()
	if err == nil {
		b, err := ioutil.ReadAll(r)
		r.Close()
		if err == nil {
			return NewResourceBytes(rs.StatusCode, b, rs.Header())
		}
	}
	return &Resource{
		header:         rs.Header(),
		statusCode:     rs.StatusCode,
		ReadSeekCloser: errReadSeekCloser{err},
	}
}

type errReadSeekCloser struct {
	err error
}

func (e errReadSeekCloser) Error() string {
	return e.err.Error()
}
func (e errReadSeekCloser) Close() error                       { return e.err }
func (e errReadSeekCloser) Read(_ []byte) (int, error)         { return 0, e.err }
func (e errReadSeekCloser) Seek(_ int64, _ int) (int64, error) { return 0, e.err }
