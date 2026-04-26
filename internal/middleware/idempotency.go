package middleware

import (
	"bytes"
	"errors"
	"log/slog"
	"net/http"

	"github.com/KozhabergenovNurzhan/E-commerce/internal/cache"
	"github.com/KozhabergenovNurzhan/E-commerce/internal/pkg/response"
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
// Requests without the header pass through unchanged.
// Fail closed: if store is nil (Redis unavailable at startup) or Redis becomes unavailable
// at runtime, requests carrying Idempotency-Key are rejected with 503 rather than
// processed non-idempotently.
func Idempotency(store *cache.IdempotencyStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.GetHeader("Idempotency-Key")
		if key == "" {
			c.Next()
			return
		}

		if store == nil {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, response.Response{
				Success: false,
				Error:   "idempotency service unavailable",
			})
			return
		}

		userID := MustUserID(c)

		rec, err := store.Get(c.Request.Context(), userID, key)
		switch {
		case err == nil:
			c.Header("X-Idempotent-Replayed", "true")
			c.Data(rec.StatusCode, "application/json; charset=utf-8", rec.Body)
			c.Abort()
			return
		case errors.Is(err, cache.ErrNotFound):
			// new key — proceed to handler
		default:
			slog.Warn("idempotency store unavailable", "error", err)
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, response.Response{
				Success: false,
				Error:   "idempotency service unavailable",
			})
			return
		}

		recorder := &responseRecorder{ResponseWriter: c.Writer, status: http.StatusOK}
		c.Writer = recorder

		c.Next()

		if recorder.status >= 200 && recorder.status < 300 {
			if err := store.Set(c.Request.Context(), userID, key, &cache.IdempotencyRecord{
				StatusCode: recorder.status,
				Body:       recorder.body.Bytes(),
			}); err != nil {
				slog.Warn("failed to store idempotency record", "error", err)
			}
		}
	}
}
