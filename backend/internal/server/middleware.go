package server

import (
	"context"
	"mime"
	"net/http"
	"strings"
	"time"

	"github.com/go-redis/redis_rate/v10"
	"github.com/redis/go-redis/v9"
)

// requireJSONMiddleware ensures that the incoming request has a Content-Type of application/json
func requireJSONMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch {
			contentType := r.Header.Get("Content-Type")
			if contentType == "" {
				http.Error(w, "Content-Type header is required", http.StatusUnsupportedMediaType)
				return
			}

			mt, _, err := mime.ParseMediaType(contentType)
			if err != nil || mt != "application/json" {
				http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

// bodySizeLimitMiddleware restricts the request body to a maximum number of bytes
func bodySizeLimitMiddleware(maxBytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			next.ServeHTTP(w, r)
		})
	}
}

// rateLimitMiddleware applies a token bucket rate limit using Redis based on the client IP
func rateLimitMiddleware(redisClient *redis.Client, rate, burst int, period time.Duration) func(http.Handler) http.Handler {
	limiter := redis_rate.NewLimiter(redisClient)
	limit := redis_rate.Limit{
		Rate:   rate,
		Burst:  burst,
		Period: period,
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Basic IP extraction. In a real world scenario behind a proxy,
			// you'd look at X-Forwarded-For or X-Real-IP.
			ip := r.RemoteAddr
			if colonIdx := strings.LastIndex(ip, ":"); colonIdx != -1 {
				ip = ip[:colonIdx]
			}

			key := "rate_limit:" + ip
			res, err := limiter.Allow(context.Background(), key, limit)
			if err != nil {
				// Fail open on redis errors to prevent total outage
				next.ServeHTTP(w, r)
				return
			}

			if res.Allowed == 0 {
				http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
