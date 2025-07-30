package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestLineParse(t *testing.T) {
	// Test: Valid single header
	headers := make(Headers)
	data := []byte("Host: localhost:42069\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, 23, n)
	assert.False(t, done)

	// Test: Valid single header with extra whitespace
	headers = make(Headers)
	data = []byte("        Host: localhost:42069           \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, 42, n)
	assert.False(t, done)

	// Test: Valid 2 Headers with existing headers
	headers = make(Headers)
	data = []byte("   Host: localhost:42069  \r\n     Content: application/json    \r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, 28, n)
	assert.False(t, done)

	// Test: Valid done 
	headers = make(Headers)
	data = []byte("\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.Empty(t, headers)
	assert.Equal(t, 2, n)
	assert.True(t, done)

	// Test: Valid header with uppercase alphabet in header key
	headers = make(Headers)
	data = []byte("HOst: localhost:42069\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, 23, n)
	assert.False(t, done)

	// Test: Invalid header key character
	headers = make(Headers)
	data = []byte("H©st: localhost:42069\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, len(headers))
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: Invalid spacing header
	headers = make(Headers)
	data = []byte("       Host : localhost:42069       \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, len(headers))
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: Valid repeated header 
	headers = make(Headers)
	headers["host"] = "localhost:42069"
	data = []byte("Host:    localhost:69420  \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069, localhost:69420", headers["host"])
	assert.Equal(t, 28, n)
	assert.False(t, done)
}