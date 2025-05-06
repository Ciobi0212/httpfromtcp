package request

import (
	"io"
	"strings"
	"testing"

	"github.com/Ciobi0212/httpfromtcp/headers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
	if n > cr.numBytesPerRead {
		n = cr.numBytesPerRead
		cr.pos -= n - cr.numBytesPerRead
	}
	return n, nil
}

func TestRequestFromReader(t *testing.T) {
	// Test: Good GET Request line with headers

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

	// Test headers using Get method
	assert.Equal(t, "localhost:42069", r.Headers.Get("Host"))
	assert.Equal(t, "curl/7.81.0", r.Headers.Get("User-Agent"))
	assert.Equal(t, "*/*", r.Headers.Get("Accept"))
	assert.Equal(t, 3, len(r.Headers))

	// Test: Good GET Request line with path and headers
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

	// Test headers using Get method
	assert.Equal(t, "localhost:42069", r.Headers.Get("Host"))
	assert.Equal(t, "curl/7.81.0", r.Headers.Get("User-Agent"))
	assert.Equal(t, "*/*", r.Headers.Get("Accept"))
	assert.Equal(t, 3, len(r.Headers))

	// Test: Invalid number of parts in request line
	_, err = RequestFromReader(strings.NewReader("/coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
	require.Error(t, err)
}

func TestHeaderParsing(t *testing.T) {
	// Test valid header
	h := headers.NewHeaders()

	str := "Host:  localhost:42069" + crlf

	_, _, err := h.ParseHeader([]byte(str))
	require.NoError(t, err)
	assert.Equal(t, "localhost:42069", h.Get("Host"))

	// Test invalid spacing header
	h = headers.NewHeaders()

	str = " Host:  localhost:42069" + crlf

	_, _, err = h.ParseHeader([]byte(str))
	require.Error(t, err)

	// Test invalid character in header key
	h = headers.NewHeaders()

	str = " H@st:  localhost:42069" + crlf

	_, _, err = h.ParseHeader([]byte(str))
	require.Error(t, err)

	// Test same header key
	h = headers.NewHeaders()

	h.Add("Cars", "dacia")

	str = "cars: renault" + crlf
	_, _, err = h.ParseHeader([]byte(str))
	require.NoError(t, err)
	assert.Equal(t, "dacia, renault", h.Get("Cars"))

	// Test: Standard Headers
	reader := &chunkReader{
		data:            "GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 3,
	}
	r, err := RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "localhost:42069", r.Headers["host"])
	assert.Equal(t, "curl/7.81.0", r.Headers["user-agent"])
	assert.Equal(t, "*/*", r.Headers["accept"])

	// Test: Malformed Header
	reader = &chunkReader{
		data:            "GET / HTTP/1.1\r\nHost localhost:42069\r\n\r\n",
		numBytesPerRead: 3,
	}
	_, err = RequestFromReader(reader)
	require.Error(t, err)
}

func TestBodyParsing(t *testing.T) {
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
	_, err = RequestFromReader(reader)
	require.Error(t, err)
}

func TestRequestFromReader_WithQueryParams(t *testing.T) {
	tests := []struct {
		name                  string
		rawRequest            string
		expectedRequestTarget string // Target after stripping query params
		expectedQueryParams   map[string]string
		expectError           bool
	}{
		{
			name:                  "Simple query params",
			rawRequest:            "GET /path?name=test&age=30 HTTP/1.1\r\nHost: example.com\r\n\r\n",
			expectedRequestTarget: "/path",
			expectedQueryParams:   map[string]string{"name": "test", "age": "30"},
			expectError:           false,
		},
		{
			name:                  "Query params with URL encoding",
			rawRequest:            "GET /search?query=hello%20world&lang=en%2FUS HTTP/1.1\r\nHost: example.com\r\n\r\n",
			expectedRequestTarget: "/search",
			expectedQueryParams:   map[string]string{"query": "hello world", "lang": "en/US"},
			expectError:           false,
		},
		{
			name:                  "Query param with no value",
			rawRequest:            "GET /filter?active&sort=asc HTTP/1.1\r\nHost: example.com\r\n\r\n",
			expectedRequestTarget: "/filter",
			expectedQueryParams:   map[string]string{"active": "", "sort": "asc"},
			expectError:           false,
		},
		{
			name:                  "Query param with empty value",
			rawRequest:            "GET /config?setting=&mode=test HTTP/1.1\r\nHost: example.com\r\n\r\n",
			expectedRequestTarget: "/config",
			expectedQueryParams:   map[string]string{"setting": "", "mode": "test"},
			expectError:           false,
		},
		{
			name:                  "No query params",
			rawRequest:            "GET /simplepath HTTP/1.1\r\nHost: example.com\r\n\r\n",
			expectedRequestTarget: "/simplepath",
			expectedQueryParams:   map[string]string{}, // Expect empty map, not nil
			expectError:           false,
		},
		{
			name:                  "Path ends with question mark",
			rawRequest:            "GET /path? HTTP/1.1\r\nHost: example.com\r\n\r\n",
			expectedRequestTarget: "/path",
			expectedQueryParams:   map[string]string{}, // Expect empty map
			expectError:           false,
		},
		{
			name:                  "Empty query string after question mark",
			rawRequest:            "GET /path?& HTTP/1.1\r\nHost: example.com\r\n\r\n", // Technically valid, results in no params
			expectedRequestTarget: "/path",
			expectedQueryParams:   map[string]string{},
			expectError:           false,
		},
		{
			name:                  "Query param with equals in value",
			rawRequest:            "GET /data?param=key%3Dvalue HTTP/1.1\r\nHost: example.com\r\n\r\n",
			expectedRequestTarget: "/data",
			expectedQueryParams:   map[string]string{"param": "key=value"},
			expectError:           false,
		},
		{
			name:                  "Multiple ampersands",
			rawRequest:            "GET /test?a=1&&b=2&c=3 HTTP/1.1\r\nHost: example.com\r\n\r\n",
			expectedRequestTarget: "/test",
			expectedQueryParams:   map[string]string{"a": "1", "b": "2", "c": "3"},
			expectError:           false,
		},
		// Add more test cases for edge cases or malformed query strings if addQueryParams is made more strict
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			reader := strings.NewReader(tc.rawRequest)
			req, err := RequestFromReader(reader)

			if tc.expectError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, req)

			assert.Equal(t, tc.expectedRequestTarget, req.RequestLine.RequestTarget)
			if len(tc.expectedQueryParams) == 0 {
				// If expected is empty, actual should be empty or nil.
				// Your current addQueryParams initializes QueryParams, so it won't be nil.
				assert.Empty(t, req.QueryParams)
			} else {
				assert.Equal(t, tc.expectedQueryParams, req.QueryParams)
			}
		})
	}
}

func TestAddQueryParams(t *testing.T) {
	tests := []struct {
		name                 string
		initialRequestTarget string
		expectedQueryParams  map[string]string
		expectedIdx          int // Changed from expectedFound (bool) to expectedIdx (int)
		expectError          bool
	}{
		{
			name:                 "Simple params",
			initialRequestTarget: "/path?name=test&age=30",
			expectedQueryParams:  map[string]string{"name": "test", "age": "30"},
			expectedIdx:          5, // Index of '?'
			expectError:          false,
		},
		{
			name:                 "URL encoded params",
			initialRequestTarget: "/search?query=hello%20world&lang=en%2FUS",
			expectedQueryParams:  map[string]string{"query": "hello world", "lang": "en/US"},
			expectedIdx:          7, // Index of '?'
			expectError:          false,
		},
		{
			name:                 "Param with no value",
			initialRequestTarget: "/filter?active&sort=asc",
			expectedQueryParams:  map[string]string{"active": "", "sort": "asc"},
			expectedIdx:          7, // Index of '?'
			expectError:          false,
		},
		{
			name:                 "Param with empty value",
			initialRequestTarget: "/config?setting=&mode=test",
			expectedQueryParams:  map[string]string{"setting": "", "mode": "test"},
			expectedIdx:          7, // Index of '?'
			expectError:          false,
		},
		{
			name:                 "No query params",
			initialRequestTarget: "/simplepath",
			expectedQueryParams:  nil,
			expectedIdx:          -1, // '?' not found
			expectError:          false,
		},
		{
			name:                 "Path ends with question mark",
			initialRequestTarget: "/path?",
			expectedQueryParams:  map[string]string{}, // No actual params parsed
			expectedIdx:          5,                   // '?' is found
			expectError:          false,
		},
		{
			name:                 "Empty key value pair",
			initialRequestTarget: "/path?=&foo=bar",
			expectedQueryParams:  map[string]string{"foo": "bar"},
			expectedIdx:          5, // Index of '?'
			expectError:          false,
		},
		{
			name:                 "Value contains equals sign",
			initialRequestTarget: "/data?param=key%3Dvalue&next=val",
			expectedQueryParams:  map[string]string{"param": "key=value", "next": "val"},
			expectedIdx:          5, // Index of '?'
			expectError:          false,
		},
		{
			name:                 "Multiple ampersands",
			initialRequestTarget: "/test?a=1&&b=2&c=3",
			expectedQueryParams:  map[string]string{"a": "1", "b": "2", "c": "3"},
			expectedIdx:          5, // Index of '?'
			expectError:          false,
		},
		{
			name:                 "Malformed pair - no equals",
			initialRequestTarget: "/test?keyonly&another=value",
			expectedQueryParams:  map[string]string{"keyonly": "", "another": "value"},
			expectedIdx:          5, // Index of '?'
			expectError:          false,
		},
		// Add tests for error cases from QueryUnescape if i make addQueryParams stricter
		{
			name:                 "Malformed key unescape",
			initialRequestTarget: "/test?key%=fail&value=ok",
			expectedQueryParams:  nil,
			expectedIdx:          5, // '?' is found, error occurs during parsing
			expectError:          true,
		},
		{
			name:                 "Malformed value unescape",
			initialRequestTarget: "/test?key=ok&value%=fail",
			expectedQueryParams:  nil,
			expectedIdx:          5, // '?' is found, error occurs during parsing
			expectError:          true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := NewRequest()
			req.RequestLine.RequestTarget = tc.initialRequestTarget

			idx, err := addQueryParams(req) // Renamed 'found' to 'idx'

			if tc.expectError {
				require.Error(t, err)
				// Optionally, assert the specific error type or message if needed
				// We still check idx because addQueryParams returns the '?' index even on parse errors
				assert.Equal(t, tc.expectedIdx, idx, "Index of '?' should be reported even on error")
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.expectedIdx, idx) // Assert the correct index

			if tc.expectedQueryParams == nil {
				assert.Nil(t, req.QueryParams, "Expected QueryParams to be nil")
			} else {
				assert.Equal(t, tc.expectedQueryParams, req.QueryParams)
			}
		})
	}
}
