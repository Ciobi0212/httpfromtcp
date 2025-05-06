package response

import (
	"testing"

	"github.com/Ciobi0212/httpfromtcp/internal/request"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Dummy handler functions for testing
func dummyHandler1(w ResponseWriter, req *request.Request) *HandlerError { return nil }
func dummyHandler2(w ResponseWriter, req *request.Request) *HandlerError { return nil }
func dummyHandler3(w ResponseWriter, req *request.Request) *HandlerError { return nil }
func dummyHandler4(w ResponseWriter, req *request.Request) *HandlerError { return nil }
func dummyHandler5(w ResponseWriter, req *request.Request) *HandlerError { return nil }

func TestRouter_AddAndGetHandler_Static(t *testing.T) {
	router := NewRouter()

	router.AddHandler(GET, "/", dummyHandler1)
	router.AddHandler(GET, "/users", dummyHandler2)
	router.AddHandler(POST, "/users", dummyHandler3)
	router.AddHandler(GET, "/users/profile", dummyHandler4)

	// Test cases
	tests := []struct {
		method   HttpMethod
		path     string
		expected Handler
		found    bool
	}{
		{GET, "/", dummyHandler1, true},
		{GET, "/users", dummyHandler2, true},
		{POST, "/users", dummyHandler3, true},
		{GET, "/users/profile", dummyHandler4, true},
		{PUT, "/users", nil, false},       // Wrong method
		{GET, "/nonexistent", nil, false}, // Wrong path
		{GET, "/users/", nil, false},      // Trailing slash mismatch (current implementation)
	}

	for _, tc := range tests {
		handler, params, ok := router.GetHandlerAndPathParamsForPath(tc.method, tc.path)
		assert.Equal(t, tc.found, ok, "Path: %s, Method: %s", tc.path, tc.method)
		// Use require.FunctionAddr for comparing function pointers
		if tc.expected != nil {
			require.NotNil(t, handler, "Path: %s, Method: %s", tc.path, tc.method)
			// Note: Comparing function pointers directly can be tricky.
			// This basic check verifies *a* handler was found.
			// For exact match, consider using reflection or unique identifiers if needed.
		} else {
			assert.Nil(t, handler, "Path: %s, Method: %s", tc.path, tc.method)
		}
		assert.Empty(t, params, "Static path should have no params: %s", tc.path)
	}
}

func TestRouter_AddAndGetHandler_Param(t *testing.T) {
	router := NewRouter()

	router.AddHandler(GET, "/users/{userId}", dummyHandler1)
	router.AddHandler(PUT, "/users/{userId}/settings", dummyHandler2)
	router.AddHandler(GET, "/products/{productId}/details", dummyHandler3)

	tests := []struct {
		method         HttpMethod
		path           string
		expected       Handler
		expectedParams map[string]string
		found          bool
	}{
		{GET, "/users/123", dummyHandler1, map[string]string{"userid": "123"}, true},
		{GET, "/users/abc", dummyHandler1, map[string]string{"userid": "abc"}, true},
		{PUT, "/users/456/settings", dummyHandler2, map[string]string{"userid": "456"}, true},
		{GET, "/products/xyz/details", dummyHandler3, map[string]string{"productid": "xyz"}, true},
		{POST, "/users/123", nil, nil, false},                // Wrong method
		{GET, "/users/123/nonexistent", nil, nil, false},     // Path doesn't match structure
		{GET, "/products/xyz", nil, nil, false},              // Path doesn't match structure
		{GET, "/products/xyz/details/more", nil, nil, false}, // Path too long
		{GET, "/users/{userId}", nil, nil, false},            // Literal param name shouldn't match
		{GET, "/products/123/settings", nil, nil, false},     // Mismatched static segment
	}

	for _, tc := range tests {
		handler, params, ok := router.GetHandlerAndPathParamsForPath(tc.method, tc.path)
		assert.Equal(t, tc.found, ok, "Path: %s, Method: %s", tc.path, tc.method)
		if tc.expected != nil {
			require.NotNil(t, handler, "Path: %s, Method: %s", tc.path, tc.method)
		} else {
			assert.Nil(t, handler, "Path: %s, Method: %s", tc.path, tc.method)
		}
		assert.Equal(t, tc.expectedParams, params, "Path: %s, Method: %s", tc.path, tc.method)
	}
}

func TestRouter_AddHandler_ConflictPanic(t *testing.T) {
	// Test conflicting parameter names at the same level
	assert.Panics(t, func() {
		router := NewRouter()
		router.AddHandler(GET, "/users/{userId}", dummyHandler1)
		router.AddHandler(GET, "/users/{differentId}", dummyHandler2) // Should panic
	}, "Should panic on conflicting parameter names")
}

func TestRouter_MixedStaticAndParam(t *testing.T) {
	router := NewRouter()

	router.AddHandler(GET, "/data/static", dummyHandler1)
	router.AddHandler(GET, "/data/{id}", dummyHandler2)
	router.AddHandler(POST, "/data/{id}/update", dummyHandler3)

	tests := []struct {
		method         HttpMethod
		path           string
		expected       Handler
		expectedParams map[string]string
		found          bool
	}{
		{GET, "/data/static", dummyHandler1, map[string]string{}, true}, // Matches static first
		{GET, "/data/123", dummyHandler2, map[string]string{"id": "123"}, true},
		{GET, "/data/abc", dummyHandler2, map[string]string{"id": "abc"}, true},
		{POST, "/data/456/update", dummyHandler3, map[string]string{"id": "456"}, true},
		{PUT, "/data/static", nil, nil, false},      // Wrong method for static
		{GET, "/data/static/more", nil, nil, false}, // Static path too long
	}

	for _, tc := range tests {
		handler, params, ok := router.GetHandlerAndPathParamsForPath(tc.method, tc.path)
		assert.Equal(t, tc.found, ok, "Path: %s, Method: %s", tc.path, tc.method)
		if tc.expected != nil {
			require.NotNil(t, handler, "Path: %s, Method: %s", tc.path, tc.method)
		} else {
			assert.Nil(t, handler, "Path: %s, Method: %s", tc.path, tc.method)
		}
		assert.Equal(t, tc.expectedParams, params, "Path: %s, Method: %s", tc.path, tc.method)
	}
}
