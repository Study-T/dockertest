package middleware

import (
	"context"
	"crypto/rand"
	"net/http"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

const maxWindowSize = 10000

var _ = logx.Info // keep import

func RateLimitMiddleware(rate int) func(next http.HandlerFunc) http.HandlerFunc {
	var mu sync.Mutex
	window := make(map[string]map[int64]int)

	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			mu.Lock()
			now := time.Now().Unix()
			for ip, buckets := range window {
				for sec := range buckets {
					if now-sec > 1 {
						delete(buckets, sec)
					}
				}
				if len(buckets) == 0 {
					delete(window, ip)
				}
			}
			mu.Unlock()
		}
	}()

	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ip := parseClientIP(r)
			now := time.Now().Unix()

			mu.Lock()
			if len(window) >= maxWindowSize {
				mu.Unlock()
				http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
				return
			}
			if window[ip] == nil {
				window[ip] = make(map[int64]int)
			}
			count := window[ip][now]
			if count >= rate {
				mu.Unlock()
				http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
				return
			}
			window[ip][now] = count + 1
			mu.Unlock()

			next(w, r)
		}
	}
}

func parseClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if i := strings.IndexByte(xff, ','); i > 0 {
			return strings.TrimSpace(xff[:i])
		}
		return strings.TrimSpace(xff)
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}
	return r.RemoteAddr
}

func RequestIDMiddleware() func(next http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = generateRequestID()
			}
			w.Header().Set("X-Request-ID", requestID)
			ctx := context.WithValue(r.Context(), "request-id", requestID)
			next(w, r.WithContext(ctx))
		}
	}
}

func TimeoutMiddleware(timeout time.Duration) func(next http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		h := http.TimeoutHandler(next, timeout, "Request Timeout")
		return func(w http.ResponseWriter, r *http.Request) {
			h.ServeHTTP(w, r)
		}
	}
}

func RecoveryMiddleware() func(next http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					logx.Errorf("Panic recovered: %v\n%s", err, debug.Stack())
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte("Internal Server Error"))
				}
			}()
			next(w, r)
		}
	}
}

func CORSMiddleware(allowedOrigins, allowedMethods []string) func(next http.HandlerFunc) http.HandlerFunc {
	originSet := make(map[string]bool, len(allowedOrigins))
	for _, o := range allowedOrigins {
		originSet[o] = true
	}
	allowAllOrigins := originSet["*"]

	methods := "GET,POST,PUT,DELETE,OPTIONS"
	if len(allowedMethods) > 0 {
		parts := make([]string, 0, len(allowedMethods)+1)
		for _, m := range allowedMethods {
			parts = append(parts, m)
		}
		if !contains(allowedMethods, "OPTIONS") {
			parts = append(parts, "OPTIONS")
		}
		methods = strings.Join(parts, ",")
	}

	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if allowAllOrigins {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			} else if originSet[origin] {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Vary", "Origin")
			}
			w.Header().Set("Access-Control-Allow-Methods", methods)
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Authorization,X-Request-ID")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next(w, r)
		}
	}
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func generateRequestID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		for i := range b {
			b[i] = letters[byte(time.Now().UnixNano())%byte(len(letters))]
		}
		return string(b)
	}
	for i := range b {
		b[i] = letters[b[i]%byte(len(letters))]
	}
	return string(b)
}
