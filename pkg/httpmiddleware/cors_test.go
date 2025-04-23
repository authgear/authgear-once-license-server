package httpmiddleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCORSMiddleware(t *testing.T) {
	// Create a test handler that just returns 200 OK
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	tests := []struct {
		name                string
		allowedOrigins      string
		requestOrigin       string
		requestMethods      string
		requestHeaders      string
		method              string
		expectedStatus      int
		expectedOrigin      string
		expectedCredentials string
		expectedMethods     string
		expectedHeaders     string
		expectedMaxAge      string
	}{
		{
			name:           "No origin header",
			allowedOrigins: "http://example.com",

			requestOrigin:  "",
			requestMethods: "",
			requestHeaders: "",

			method:              http.MethodGet,
			expectedStatus:      http.StatusOK,
			expectedOrigin:      "",
			expectedCredentials: "",
			expectedMethods:     "",
			expectedHeaders:     "",
			expectedMaxAge:      "",
		},
		{
			name:           "Matching origin",
			allowedOrigins: "http://example.com",

			requestOrigin:  "http://example.com",
			requestMethods: "",
			requestHeaders: "",

			method:              http.MethodGet,
			expectedStatus:      http.StatusOK,
			expectedOrigin:      "http://example.com",
			expectedCredentials: "true",
			expectedMethods:     "",
			expectedHeaders:     "",
			expectedMaxAge:      "900",
		},
		{
			name:           "Non-matching origin",
			allowedOrigins: "http://example.com",

			requestOrigin:  "http://other.com",
			requestMethods: "",
			requestHeaders: "",

			method:              http.MethodGet,
			expectedStatus:      http.StatusOK,
			expectedOrigin:      "",
			expectedCredentials: "",
			expectedMethods:     "",
			expectedHeaders:     "",
			expectedMaxAge:      "",
		},
		{
			name:           "Multiple allowed origins",
			allowedOrigins: "http://example.com,http://other.com",

			requestOrigin:  "http://other.com",
			requestMethods: "",
			requestHeaders: "",

			method:              http.MethodGet,
			expectedStatus:      http.StatusOK,
			expectedOrigin:      "http://other.com",
			expectedCredentials: "true",
			expectedMethods:     "",
			expectedHeaders:     "",
			expectedMaxAge:      "900",
		},
		{
			name:           "Preflight request",
			allowedOrigins: "http://example.com",

			requestOrigin:  "http://example.com",
			requestMethods: "",
			requestHeaders: "",

			method:              http.MethodOptions,
			expectedStatus:      http.StatusNoContent,
			expectedOrigin:      "http://example.com",
			expectedCredentials: "true",
			expectedMethods:     "",
			expectedHeaders:     "",
			expectedMaxAge:      "900",
		},
		{
			name:           "No allowed origins",
			allowedOrigins: "",

			requestOrigin:  "http://example.com",
			requestMethods: "",
			requestHeaders: "",

			method:              http.MethodGet,
			expectedStatus:      http.StatusOK,
			expectedOrigin:      "",
			expectedCredentials: "",
			expectedMethods:     "",
			expectedHeaders:     "",
			expectedMaxAge:      "",
		},
		{
			name:           "Echo Access-Control-Request-Method",
			allowedOrigins: "http://example.com",

			requestOrigin:  "http://example.com",
			requestMethods: "POST",
			requestHeaders: "",

			method:              http.MethodOptions,
			expectedStatus:      http.StatusNoContent,
			expectedOrigin:      "http://example.com",
			expectedCredentials: "true",
			expectedMethods:     "POST",
			expectedHeaders:     "",
			expectedMaxAge:      "900",
		},
		{
			name:           "Echo Access-Control-Request-Headers",
			allowedOrigins: "http://example.com",

			requestOrigin:  "http://example.com",
			requestMethods: "",
			requestHeaders: "Content-Type",

			method:              http.MethodOptions,
			expectedStatus:      http.StatusNoContent,
			expectedOrigin:      "http://example.com",
			expectedCredentials: "true",
			expectedMethods:     "",
			expectedHeaders:     "Content-Type",
			expectedMaxAge:      "900",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Set environment variable for this test

			// Create a new request
			req := httptest.NewRequest(tc.method, "http://localhost", nil)
			if tc.requestOrigin != "" {
				req.Header.Set("Origin", tc.requestOrigin)
			}
			if tc.requestMethods != "" {
				req.Header.Set("Access-Control-Request-Method", tc.requestMethods)
			}
			if tc.requestHeaders != "" {
				req.Header.Set("Access-Control-Request-Headers", tc.requestHeaders)
			}

			// Create a response recorder
			rec := httptest.NewRecorder()

			// Apply the middleware to our test handler
			handler := CORSMiddleware(tc.allowedOrigins)(testHandler)
			handler.ServeHTTP(rec, req)

			// Check status code
			if rec.Code != tc.expectedStatus {
				t.Errorf("Expected status %d; got %d", tc.expectedStatus, rec.Code)
			}

			// Check headers
			if got := rec.Header().Get("Access-Control-Allow-Origin"); got != tc.expectedOrigin {
				t.Errorf("Expected Access-Control-Allow-Origin header %q; got %q", tc.expectedOrigin, got)
			}
			if got := rec.Header().Get("Access-Control-Allow-Credentials"); got != tc.expectedCredentials {
				t.Errorf("Expected Access-Control-Allow-Credentials header %q; got %q", tc.expectedCredentials, got)
			}
			if got := rec.Header().Get("Access-Control-Allow-Methods"); got != tc.expectedMethods {
				t.Errorf("Expected Access-Control-Allow-Methods header %q; got %q", tc.expectedMethods, got)
			}
			if got := rec.Header().Get("Access-Control-Allow-Headers"); got != tc.expectedHeaders {
				t.Errorf("Expected Access-Control-Allow-Headers header %q; got %q", tc.expectedHeaders, got)
			}
			if got := rec.Header().Get("Access-Control-Max-Age"); got != tc.expectedMaxAge {
				t.Errorf("Expected Access-Control-Max-Age header %q; got %q", tc.expectedMaxAge, got)
			}
		})
	}
}
