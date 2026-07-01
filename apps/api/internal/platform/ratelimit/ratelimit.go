// Package ratelimit provides a simple in-memory per-IP token-bucket rate limiter.
package ratelimit

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

// Limiter enforces a per-IP token-bucket rate limit.
type Limiter struct {
	enabled  bool
	rps      float64
	burst    int
	ttl      time.Duration
	cleanup  time.Duration
	mu       sync.RWMutex
	clients  map[string]*client
	stopCh   chan struct{}
	stopOnce sync.Once
}

type client struct {
	mu       sync.Mutex
	tokens   float64
	lastSeen time.Time
}

// New creates a rate limiter. If enabled is false, Allow always returns true.
func New(enabled bool, rps float64, burst int, ttl, cleanup time.Duration) *Limiter {
	l := &Limiter{
		enabled: enabled,
		rps:     rps,
		burst:   burst,
		ttl:     ttl,
		cleanup: cleanup,
		clients: make(map[string]*client),
		stopCh:  make(chan struct{}),
	}
	if enabled {
		go l.cleanupLoop()
	}
	return l
}

// Stop halts the background cleanup goroutine.
func (l *Limiter) Stop() {
	l.stopOnce.Do(func() { close(l.stopCh) })
}

// Allow reports whether a request from ip is permitted under the rate limit.
func (l *Limiter) Allow(ip string) bool {
	if !l.enabled {
		return true
	}

	l.mu.RLock()
	c, ok := l.clients[ip]
	l.mu.RUnlock()
	now := time.Now()

	if !ok {
		l.mu.Lock()
		c, ok = l.clients[ip]
		if !ok {
			c = &client{tokens: float64(l.burst) - 1, lastSeen: now}
			l.clients[ip] = c
			l.mu.Unlock()
			return true
		}
		l.mu.Unlock()
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	elapsed := now.Sub(c.lastSeen).Seconds()
	c.lastSeen = now
	c.tokens += elapsed * l.rps
	if c.tokens > float64(l.burst) {
		c.tokens = float64(l.burst)
	}
	if c.tokens < 1 {
		return false
	}
	c.tokens--
	return true
}

func (l *Limiter) cleanupLoop() {
	ticker := time.NewTicker(l.cleanup)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			l.purgeStale()
		case <-l.stopCh:
			return
		}
	}
}

func (l *Limiter) purgeStale() {
	cutoff := time.Now().Add(-l.ttl)
	l.mu.Lock()
	defer l.mu.Unlock()
	for ip, c := range l.clients {
		c.mu.Lock()
		stale := c.lastSeen.Before(cutoff)
		c.mu.Unlock()
		if stale {
			delete(l.clients, ip)
		}
	}
}

// ClientIP returns the client IP from a request, stripping the port if present.
func ClientIP(r *http.Request) string {
	ip := r.RemoteAddr
	host, _, err := net.SplitHostPort(ip)
	if err == nil {
		ip = host
	}
	return ip
}

// Excluded reports whether a request path should bypass rate limiting.
func Excluded(r *http.Request) bool {
	if r.Method == http.MethodOptions {
		return true
	}
	path := r.URL.Path
	switch path {
	case "/healthz", "/readyz", "/api/v1/auth/csrf-token":
		return true
	}
	return false
}

// Middleware returns a chi middleware that applies the limiter.
func Middleware(l *Limiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if Excluded(r) {
				next.ServeHTTP(w, r)
				return
			}

			ip := ClientIP(r)
			if l.Allow(ip) {
				next.ServeHTTP(w, r)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			reqID := middleware.GetReqID(r.Context())
			_ = json.NewEncoder(w).Encode(map[string]any{
				"error": map[string]string{
					"code":       "rate_limited",
					"message":    fmt.Sprintf("rate limit exceeded for IP %s", ip),
					"request_id": reqID,
				},
			})
		})
	}
}
