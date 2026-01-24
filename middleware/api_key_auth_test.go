package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestAPIKeyAuth(t *testing.T) {
	// Set up test API key (gt_ + 32 chars = 35 total)
	testKey := "gt_aaaabbbbccccddddeeeeffffgggghhhh"
	os.Setenv("GOTAIL_API_KEYS", testKey)
	ReloadAPIKeys()
	defer func() {
		os.Unsetenv("GOTAIL_API_KEYS")
		ReloadAPIKeys()
	}()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	tests := []struct {
		name           string
		headerKey      string
		headerValue    string
		expectedStatus int
	}{
		{
			name:           "valid key via X-API-Key header",
			headerKey:      "X-API-Key",
			headerValue:    testKey,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "valid key via Authorization Bearer",
			headerKey:      "Authorization",
			headerValue:    "Bearer " + testKey,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid key",
			headerKey:      "X-API-Key",
			headerValue:    "gt_invalidkey12345678901234567890ab",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "missing key",
			headerKey:      "",
			headerValue:    "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "wrong prefix",
			headerKey:      "X-API-Key",
			headerValue:    "xx_testkey12345678901234567890ab",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "too short",
			headerKey:      "X-API-Key",
			headerValue:    "gt_short",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
			if tt.headerKey != "" {
				req.Header.Set(tt.headerKey, tt.headerValue)
			}

			rec := httptest.NewRecorder()
			APIKeyAuth()(handler).ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}
		})
	}
}

func TestAPIKeyAuthWithBearerToken(t *testing.T) {
	testKey := "gt_aaaabbbbccccddddeeeeffffgggghhhh"
	os.Setenv("GOTAIL_API_KEYS", testKey)
	ReloadAPIKeys()
	defer func() {
		os.Unsetenv("GOTAIL_API_KEYS")
		ReloadAPIKeys()
	}()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Header.Set("Authorization", "Bearer "+testKey)

	rec := httptest.NewRecorder()
	APIKeyAuth()(handler).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestEitherAuth(t *testing.T) {
	testKey := "gt_aaaabbbbccccddddeeeeffffgggghhhh"
	testUser := "testuser"
	testPass := "testpass"

	os.Setenv("GOTAIL_API_KEYS", testKey)
	ReloadAPIKeys()
	defer func() {
		os.Unsetenv("GOTAIL_API_KEYS")
		ReloadAPIKeys()
	}()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	tests := []struct {
		name           string
		setupRequest   func(*http.Request)
		expectedStatus int
	}{
		{
			name: "valid API key",
			setupRequest: func(r *http.Request) {
				r.Header.Set("X-API-Key", testKey)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "valid basic auth",
			setupRequest: func(r *http.Request) {
				r.SetBasicAuth(testUser, testPass)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "invalid API key falls back to basic auth check",
			setupRequest: func(r *http.Request) {
				r.Header.Set("X-API-Key", "gt_invalidkey12345678901234567890ab")
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "invalid basic auth",
			setupRequest: func(r *http.Request) {
				r.SetBasicAuth("wrong", "wrong")
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "no auth",
			setupRequest:   func(r *http.Request) {},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/log", nil)
			tt.setupRequest(req)

			rec := httptest.NewRecorder()
			EitherAuth(testUser, testPass)(handler).ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}
		})
	}
}

func TestMultipleAPIKeys(t *testing.T) {
	key1 := "gt_firstkeyaabbccddee00112233445566"
	key2 := "gt_secondkeyabbccddee00112233445566"

	os.Setenv("GOTAIL_API_KEYS", key1+","+key2)
	ReloadAPIKeys()
	defer func() {
		os.Unsetenv("GOTAIL_API_KEYS")
		ReloadAPIKeys()
	}()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Test first key
	req1 := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req1.Header.Set("X-API-Key", key1)
	rec1 := httptest.NewRecorder()
	APIKeyAuth()(handler).ServeHTTP(rec1, req1)
	if rec1.Code != http.StatusOK {
		t.Errorf("first key: expected 200, got %d", rec1.Code)
	}

	// Test second key
	req2 := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req2.Header.Set("X-API-Key", key2)
	rec2 := httptest.NewRecorder()
	APIKeyAuth()(handler).ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Errorf("second key: expected 200, got %d", rec2.Code)
	}
}

func TestValidateAPIKeyFormat(t *testing.T) {
	testKey := "gt_aaaabbbbccccddddeeeeffffgggghhhh"
	os.Setenv("GOTAIL_API_KEYS", testKey)
	ReloadAPIKeys()
	defer func() {
		os.Unsetenv("GOTAIL_API_KEYS")
		ReloadAPIKeys()
	}()

	tests := []struct {
		name     string
		key      string
		expected bool
	}{
		{"valid key", testKey, true},
		{"wrong prefix", "xx_aaaabbbbccccddddeeeeffffgggghhhh", false},
		{"too short", "gt_short", false},
		{"too long", "gt_aaaabbbbccccddddeeeeffffgggghhhhiii", false},
		{"empty", "", false},
		{"just prefix", "gt_", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validateAPIKey(tt.key)
			if result != tt.expected {
				t.Errorf("validateAPIKey(%q) = %v, expected %v", tt.key, result, tt.expected)
			}
		})
	}
}
