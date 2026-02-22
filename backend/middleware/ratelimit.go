package middleware

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type ipLimiter struct {
	tokens   float64
	maxTok   float64
	rate     float64 // tokens per second
	lastSeen time.Time
	mu       sync.Mutex
}

var (
	limiters   sync.Map
	rlRate     = parseEnvFloat("RATE_LIMIT_RPS", 20)
	rlBurst    = parseEnvFloat("RATE_LIMIT_BURST", 40)
	rlDisabled = os.Getenv("RATE_LIMIT_DISABLED") == "1"
)

func parseEnvFloat(key string, def float64) float64 {
	if s := os.Getenv(key); s != "" {
		if n, err := strconv.ParseFloat(s, 64); err == nil && n > 0 {
			return n
		}
	}
	return def
}

func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return strings.TrimSpace(strings.SplitN(xff, ",", 2)[0])
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	host := r.RemoteAddr
	if i := strings.LastIndex(host, ":"); i != -1 {
		host = host[:i]
	}
	return host
}

func getLimiter(ip string) *ipLimiter {
	v, _ := limiters.LoadOrStore(ip, &ipLimiter{
		tokens:   rlBurst,
		maxTok:   rlBurst,
		rate:     rlRate,
		lastSeen: time.Now(),
	})
	return v.(*ipLimiter)
}

func (l *ipLimiter) allow() bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(l.lastSeen).Seconds()
	l.lastSeen = now

	l.tokens += elapsed * l.rate
	if l.tokens > l.maxTok {
		l.tokens = l.maxTok
	}

	if l.tokens >= 1 {
		l.tokens--
		return true
	}
	return false
}

// StartLimiterGC cleans up stale per-IP entries every 5 minutes.
func StartLimiterGC() {
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			cutoff := time.Now().Add(-10 * time.Minute)
			limiters.Range(func(k, v any) bool {
				l := v.(*ipLimiter)
				l.mu.Lock()
				stale := l.lastSeen.Before(cutoff)
				l.mu.Unlock()
				if stale {
					limiters.Delete(k)
				}
				return true
			})
		}
	}()
}

// RateLimit is a middleware that applies per-IP token-bucket rate limiting (HandlerFunc variant).
func RateLimit(next http.HandlerFunc) http.HandlerFunc {
	if rlDisabled {
		return next
	}
	return func(w http.ResponseWriter, r *http.Request) {
		ip := clientIP(r)
		if !getLimiter(ip).allow() {
			w.Header().Set("Retry-After", "1")
			writeJSON(w, 429, map[string]string{"error": "Too many requests"})
			return
		}
		next(w, r)
	}
}

// RateLimitHandler wraps any http.Handler with per-IP token-bucket rate limiting.
func RateLimitHandler(next http.Handler) http.Handler {
	if rlDisabled {
		return next
	}
	log.Printf("🛡️  Rate limiter active: %.0f req/s per IP, burst %.0f (set RATE_LIMIT_DISABLED=1 to turn off)", rlRate, rlBurst)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := clientIP(r)
		if !getLimiter(ip).allow() {
			w.Header().Set("Retry-After", "1")
			writeJSON(w, 429, map[string]string{"error": "Too many requests"})
			return
		}
		next.ServeHTTP(w, r)
	})
}
