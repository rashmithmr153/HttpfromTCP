package request

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestLineParse(t *testing.T) {
	// Test: Good GET Request line
	// Test: Good GET Request line
	reader := &chunkReader{
		data:            "GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 3,
	}
	r, err := RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "GET", r.RequestLine.Method)
	assert.Equal(t, "/", r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)

	// Test: Good GET Request line with path
	reader = &chunkReader{
		data:            "GET /coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 1,
	}
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "GET", r.RequestLine.Method)
	assert.Equal(t, "/coffee", r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)

}

type chunkReader struct {
	data            string
	numBytesPerRead int
	pos             int
}

// Read reads up to len(p) or numBytesPerRead bytes from the string per call
// its useful for simulating reading a variable number of bytes per chunk from a network connection
func (cr *chunkReader) Read(p []byte) (n int, err error) {
	if cr.pos >= len(cr.data) {
		return 0, io.EOF
	}
	endIndex := cr.pos + cr.numBytesPerRead
	if endIndex > len(cr.data) {
		endIndex = len(cr.data)
	}
	n = copy(p, cr.data[cr.pos:endIndex])
	cr.pos += n

	return n, nil
}

func getValue(r *Request, name string) string {
	value, _ := r.Header.Get(name)
	return value
}
func TestHeaderParse(t *testing.T) {
	// Test: Standard Headers
	reader := &chunkReader{
		data:            "GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 3,
	}
	r, err := RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "localhost:42069", getValue(r, "Host"))
	assert.Equal(t, "curl/7.81.0", getValue(r, "user-agent"))
	assert.Equal(t, "*/*", getValue(r, "accept"))

	// Test: Malformed Header
	reader = &chunkReader{
		data:            "GET / HTTP/1.1\r\nHost localhost:42069\r\n\r\n",
		numBytesPerRead: 3,
	}
	r, err = RequestFromReader(reader)
	require.Error(t, err)
}

func TestParseBody(t *testing.T) {
	// Test: Standard Body
	reader := &chunkReader{
		data: "POST /submit HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"Content-Length: 13\r\n" +
			"\r\n" +
			"hello world!\n",
		numBytesPerRead: 3,
	}
	r, err := RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "hello world!\n", string(r.Body))

	// Test: Body shorter than reported content length
	reader = &chunkReader{
		data: "POST /submit HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"Content-Length: 20\r\n" +
			"\r\n" +
			"partial content",
		numBytesPerRead: 3,
	}
	r, err = RequestFromReader(reader)
	require.Error(t, err)
}
func TestGetRequestNoBody(t *testing.T) {
	// This is the core hang scenario — GET with no body, no content-length.
	// The parser must reach stateDone without waiting for EOF.
	reader := &chunkReader{
		data:            "GET / HTTP/1.1\r\nHost: localhost:42069\r\n\r\n",
		numBytesPerRead: 3,
	}
	r, err := RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "GET", r.RequestLine.Method)
	assert.Equal(t, "/", r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)
	assert.Nil(t, r.Body)
}

func TestGetRequestNoBodySingleByteReads(t *testing.T) {
	// Same scenario but with 1 byte per read — stresses state transitions
	// across many small reads, making it more likely to expose the hang.
	reader := &chunkReader{
		data:            "GET / HTTP/1.1\r\nHost: localhost:42069\r\n\r\n",
		numBytesPerRead: 1,
	}
	r, err := RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "GET", r.RequestLine.Method)
	assert.Nil(t, r.Body)
}

func TestGetRequestNoBodyLargeReads(t *testing.T) {
	// Entire request arrives in one read — parser must drain all state
	// transitions from a single buffer without blocking on next Read.
	reader := &chunkReader{
		data:            "GET /path HTTP/1.1\r\nHost: localhost:42069\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 1024,
	}
	r, err := RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "GET", r.RequestLine.Method)
	assert.Equal(t, "/path", r.RequestLine.RequestTarget)
	assert.Nil(t, r.Body)
}

func TestPostWithExactContentLength(t *testing.T) {
	// POST where body arrives split across reads — parser must not
	// block waiting for more data once content-length is satisfied.
	reader := &chunkReader{
		data: "POST /submit HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"Content-Length: 5\r\n" +
			"\r\n" +
			"hello",
		numBytesPerRead: 3,
	}
	r, err := RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "hello", string(r.Body))
}

// func TestPostBodyArrivesWithHeaders(t *testing.T) {
// 	// Body bytes arrive in the same read as the end of headers (\r\n\r\n).
// 	// Tests that stateParseBody correctly picks up leftover buffer bytes
// 	// rather than blocking on the next Read.
// 	reader := &chunkReader{
// 		data: "POST /submit HTTP/1.1\r\n" +
// 			"Host: localhost:42069\r\n" +
// 			"Content-Length: 5\r\n" +
// 			"\r\nhello",
// 		numBytesPerRead: 1024,
// 	}
// 	r, err := RequestFromReader(reader)
// 	require.NoError(t, err)
// 	require.NotNil(t, r)
// 	assert.Equal(t, "hello", string(r.Body))
// }

func TestMultipleHeadersNoBody(t *testing.T) {
	// Many headers, no body — all state transitions must complete
	// before blocking on Read.
	reader := &chunkReader{
		data: "GET /api HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"User-Agent: curl/7.81.0\r\n" +
			"Accept: */*\r\n" +
			"Connection: keep-alive\r\n" +
			"\r\n",
		numBytesPerRead: 5,
	}
	r, err := RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "GET", r.RequestLine.Method)
	assert.Equal(t, "localhost:42069", getValue(r, "Host"))
	assert.Equal(t, "keep-alive", getValue(r, "connection"))
	assert.Nil(t, r.Body)
}
