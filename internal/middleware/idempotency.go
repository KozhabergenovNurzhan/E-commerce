package middleware

import (
	"bytes"
	"net/http"

	"github.com/KozhabergenovNurzhan/E-commerce/internal/cache"
	"github.com/gin-gonic/gin"
)

// responseRecorder wraps gin.ResponseWriter to capture the status and body.
type responseRecorder struct {
	gin.ResponseWriter
	body   bytes.Buffer
	status int
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

func (r *responseRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

// Idempotency caches successful responses keyed by userID + Idempotency-Key header.
// If the header is absent the request passes through unchanged.
// If store is nil (Redis unavailable) the middleware is a no-op.
func Idempotency(store *cache.IdempotencyStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		if store == nil {
			c.Next()
			return
		}

		key := c.GetHeader("Idempotency-Key")
		if key == "" {
			c.Next()
			return
		}

		userID := MustUserID(c)

		if rec, err := store.Get(c.Request.Context(), userID, key); err == nil {
			c.Header("X-Idempotent-Replayed", "true")
			c.Data(rec.StatusCode, "application/json; charset=utf-8", rec.Body)
			c.Abort()
			return
		}

		recorder := &responseRecorder{ResponseWriter: c.Writer, status: http.StatusOK}
		c.Writer = recorder

		c.Next()

		if recorder.status >= 200 && recorder.status < 300 {
			_ = store.Set(c.Request.Context(), userID, key, &cache.IdempotencyRecord{
				StatusCode: recorder.status,
				Body:       recorder.body.Bytes(),
			})
		}
	}
}
