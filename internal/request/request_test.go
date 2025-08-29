package request

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type chunkReader struct {
	data            string
	numBytesPerRead int
	pos             int
}

// Read reads up to len(p) or numBytesPerRead bytes from the string per call.
// It's useful for simulating reading a variable number of bytes per chunk 
// from a network connection.
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

func TestRequestLineParse(t *testing.T) {
	// Test: Good GET Request line
	reader := &chunkReader{
		data: "GET / HTTP/1.1\r\n" + 
			"Host: localhost:42069\r\n" + 
			"User-Agent: curl/7.81.0\r\n" + 
			"Accept: */*\r\n" + 
			"\r\n",
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
		data: "GET /coffee HTTP/1.1\r\n" + 
			"Host: localhost:42069\r\n" + 
			"User-Agent: curl/7.81.0\r\n" + 
			"Accept: */*\r\n" + 
			"\r\n",
		numBytesPerRead: 5,
	}
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "GET", r.RequestLine.Method)
	assert.Equal(t, "/coffee", r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)

	// Test: Good POST Request line with path
	reader = &chunkReader{
		data: "POST /coffee/mediumroast HTTP/1.1\r\n" + 
			"Host: localhost:42069\r\n" + 
			"User-Agent: curl/7.81.0\r\n" + 
			"Accept: */*\r\n" + 
			"\r\n",
		numBytesPerRead: 6,
	}
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "POST", r.RequestLine.Method)
	assert.Equal(t, "/coffee/mediumroast", r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)

	// Test: Invalid number of parts in request line
	reader = &chunkReader{
		data: "/coffee HTTP/1.1\r\n" + 
			"Host: localhost:42069\r\n" + 
			"User-Agent: curl/7.81.0\r\n" + 
			"Accept: */*\r\n" + 
			"\r\n",
		numBytesPerRead: 4,
	}
	_, err = RequestFromReader(reader)
	require.Error(t, err)

	// Test: Invalid version in request line
	reader = &chunkReader{
		data: "GET /coffee HTTP/1.2\r\n" + 
			"Host: localhost:42069\r\n" + 
			"User-Agent: curl/7.81.0\r\n" + 
			"Accept: */*\r\n" + 
			"\r\n",
		numBytesPerRead: 7,
	}
	_, err = RequestFromReader(reader)
	require.Error(t, err)

	// Test: Out of order method in request line
	reader = &chunkReader{
		data: "/coffee GET HTTP/1.1\r\n" + 
			"Host: localhost:42069\r\n" + 
			"User-Agent: curl/7.81.0\r\n" + 
			"Accept: */*\r\n" + 
			"\r\n",
		numBytesPerRead: 5,
	}
	_, err = RequestFromReader(reader)
	require.Error(t, err)

	// Test: Incomplete request line
	reader = &chunkReader{
		data: "/coffee GET HTTP/1.1",
		numBytesPerRead: 5,
	}
	_, err = RequestFromReader(reader)
	require.Error(t, err)
}

func TestRequestHeadersParse(t *testing.T) {
	// Test: Standard Headers
	reader := &chunkReader{
		data: "GET / HTTP/1.1\r\nHost: localhost:42069\r\n" + 
			"User-Agent: curl/7.81.0\r\n" + 
			"Accept: */*\r\n" + 
			"\r\n",
		numBytesPerRead: 8,
	}
	r, err := RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "localhost:42069", r.Headers["host"])
	assert.Equal(t, "curl/7.81.0", r.Headers["user-agent"])
	assert.Equal(t, "*/*", r.Headers["accept"])

	// Test: Empty Headers
	reader = &chunkReader{
		data: "GET / HTTP/1.1\r\n" + 
			"\r\n",
		numBytesPerRead: 3,
	}
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, 0, len(r.Headers))

	// Test: Duplicate Headers
	reader = &chunkReader{
		data: "GET / HTTP/1.1\r\n" + 
			"Host: localhost:42069\r\n" + 
			"Content: text\r\n" + 
			"Content: application/json\r\n" + 
			"\r\n",
		numBytesPerRead: 10,
	}
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "localhost:42069", r.Headers["host"])
	assert.Equal(t, "text, application/json", r.Headers["content"])

	// Test: Case Insensitive Headers
	reader = &chunkReader{
		data: "GET / HTTP/1.1\r\n" + 
			"HoSt: localhost:42069\r\n" + 
			"ConTEnt: text\r\n" + 
			"\r\n",
		numBytesPerRead: 5,
	}
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "localhost:42069", r.Headers["host"])
	assert.Equal(t, "text", r.Headers["content"])

	// Test: Malformed Header
	reader = &chunkReader{
		data: "GET / HTTP/1.1\r\n" + 
			"Host localhost:42069\r\n" + 
			"\r\n",
		numBytesPerRead: 4,
	}
	_, err = RequestFromReader(reader)
	require.Error(t, err)

	// Test: Incomplete header section
	reader = &chunkReader{
		data: "GET / HTTP/1.1\r\n" + 
			"Host: localhost:42069\r\n" + 
			"User-Agent: curl/7.81.0",
		numBytesPerRead: 3,
	}
	_, err = RequestFromReader(reader)
	require.Error(t, err)

	// Test: Incomplete header section 
	reader = &chunkReader{
		data: "GET / HTTP/1.1\r\n" + 
			"Host: localhost:42069\r\n" + 
			"User-Agent: curl/7.81.0\r\n",
		numBytesPerRead: 3,
	}
	_, err = RequestFromReader(reader)
	require.Error(t, err)
}

func TestRequestBodyParse(t * testing.T) {
	// Test: Body equal to reported non-zero content length 
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

	// Test: Body shorter than reported non-zero content length
	reader = &chunkReader{
		data: "POST /submit HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"Content-Length: 20\r\n" +
			"\r\n" +
			"hello world!\n",
		numBytesPerRead: 5,
	}
	_, err = RequestFromReader(reader)
	require.Error(t, err)

	// Test: Body greater than reported non-zero content length
	reader = &chunkReader{
		data: "POST /submit HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"Content-Length: 10\r\n" +
			"\r\n" +
			"hello world!\n",
		numBytesPerRead: 5,
	}
	_, err = RequestFromReader(reader)
	require.Error(t, err)

	// Test: Empty Body with reported non-zero content length
	reader = &chunkReader{
		data: "POST /submit HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"Content-Length: 10\r\n" +
			"\r\n",
		numBytesPerRead: 5,
	}
	_, err = RequestFromReader(reader)
	require.Error(t, err)

	// Test: Empty Body with reported zero content length
	reader = &chunkReader{
		data: "POST /submit HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"Content-Length: 0\r\n" +
			"\r\n",
		numBytesPerRead: 4,
	}
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "", string(r.Body))

	// Test: Non empty Body with reported zero content length
	reader = &chunkReader{
		data: "POST /submit HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"Content-Length: 0\r\n" +
			"\r\n"+
			"hello world!\n",
		numBytesPerRead: 4,
	}
	_, err = RequestFromReader(reader)
	require.Error(t, err)

	// Test: Empty Body with no reported content length
	reader = &chunkReader{
		data: "POST /submit HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"\r\n",
		numBytesPerRead: 3,
	}
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "", string(r.Body))

	// Test: Non empty Body with no reported content length 
	reader = &chunkReader{
		data: "POST /submit HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"\r\n" +
			"hello world!\n",
		numBytesPerRead: 5,
	}
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "", string(r.Body))
}