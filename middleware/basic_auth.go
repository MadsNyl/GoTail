package middleware

import (
  "crypto/sha256"
  "crypto/subtle"
  "net/http"
)

func BasicAuth(expectedUser, expectedPass string) func(http.Handler) http.Handler {
  userHash := sha256.Sum256([]byte(expectedUser))
  passHash := sha256.Sum256([]byte(expectedPass))
  return func(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
      u, p, ok := r.BasicAuth()
      if !ok {
        unauthorized(w)
        return
      }

      uHash := sha256.Sum256([]byte(u))
      pHash := sha256.Sum256([]byte(p))
      if subtle.ConstantTimeCompare(uHash[:], userHash[:]) == 1 &&
         subtle.ConstantTimeCompare(pHash[:], passHash[:]) == 1 {
        next.ServeHTTP(w, r)
        return
      }

      unauthorized(w)
    })
  }
}

func unauthorized(w http.ResponseWriter) {
  w.Header().Set("WWW-Authenticate", `Basic realm="GoTail UI", charset="UTF-8"`)
  http.Error(w, "Unauthorized", http.StatusUnauthorized)
}
