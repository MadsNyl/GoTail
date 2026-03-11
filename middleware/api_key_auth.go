package middleware

import (
	"crypto/sha256"
	"crypto/subtle"
	"log"
	"net/http"
	"os"
	"strings"
)

var apiKeyHashes [][32]byte

func init() {
	loadAPIKeys()
}

func loadAPIKeys() {
	keysEnv := os.Getenv("GOTAIL_API_KEYS")
	if keysEnv == "" {
		return
	}

	keys := strings.Split(keysEnv, ",")
	for _, key := range keys {
		key = strings.TrimSpace(key)
		if key != "" {
			hash := sha256.Sum256([]byte(key))
			apiKeyHashes = append(apiKeyHashes, hash)
		}
	}
}

// ReloadAPIKeys reloads API keys from environment (for use after env changes)
func ReloadAPIKeys() {
	apiKeyHashes = nil
	loadAPIKeys()
}

// APIKeyAuth returns middleware that validates API key authentication.
// Keys can be provided via X-API-Key header or Authorization: Bearer header.
func APIKeyAuth() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := extractAPIKey(r)
			if key == "" {
				log.Printf("AUTH FAILURE: api_key missing, ip=%s path=%s", r.RemoteAddr, r.URL.Path)
				unauthorizedJSON(w, "API key required")
				return
			}

			if !validateAPIKey(key) {
				log.Printf("AUTH FAILURE: api_key invalid, ip=%s path=%s", r.RemoteAddr, r.URL.Path)
				unauthorizedJSON(w, "Invalid API key")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// EitherAuth returns middleware that accepts either API key or basic auth.
// Useful for endpoints that need to support both authentication methods.
func EitherAuth(expectedUser, expectedPass string) func(http.Handler) http.Handler {
	userHash := sha256.Sum256([]byte(expectedUser))
	passHash := sha256.Sum256([]byte(expectedPass))

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Try API key first
			key := extractAPIKey(r)
			if key != "" && validateAPIKey(key) {
				next.ServeHTTP(w, r)
				return
			}

			// Fall back to basic auth
			u, p, ok := r.BasicAuth()
			if ok {
				uHash := sha256.Sum256([]byte(u))
				pHash := sha256.Sum256([]byte(p))
				if subtle.ConstantTimeCompare(uHash[:], userHash[:]) == 1 &&
					subtle.ConstantTimeCompare(pHash[:], passHash[:]) == 1 {
					next.ServeHTTP(w, r)
					return
				}
			}

			// Neither authentication method succeeded
			log.Printf("AUTH FAILURE: either_auth no valid method, ip=%s path=%s", r.RemoteAddr, r.URL.Path)
			unauthorized(w)
		})
	}
}

func extractAPIKey(r *http.Request) string {
	// Check X-API-Key header first
	if key := r.Header.Get("X-API-Key"); key != "" {
		return key
	}

	// Check Authorization: Bearer header
	auth := r.Header.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}

	return ""
}

func validateAPIKey(key string) bool {
	if len(apiKeyHashes) == 0 {
		return false
	}

	// Validate key format: must start with gt_ and be 35 chars total
	if !strings.HasPrefix(key, "gt_") || len(key) != 35 {
		return false
	}

	keyHash := sha256.Sum256([]byte(key))

	for _, validHash := range apiKeyHashes {
		if subtle.ConstantTimeCompare(keyHash[:], validHash[:]) == 1 {
			return true
		}
	}
	return false
}

func unauthorizedJSON(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte(`{"error":"` + message + `"}`))
}
