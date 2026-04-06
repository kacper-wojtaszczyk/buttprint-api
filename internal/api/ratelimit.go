package api

import (
	"log/slog"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type client struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type RateLimiter struct {
	clients map[string]*client
	mu      sync.Mutex
	rps     float64
	burst   int
	logger  *slog.Logger
	done    chan struct{}
}

func NewRateLimiter(rps float64, burst int, logger *slog.Logger) *RateLimiter {
	rl := &RateLimiter{
		clients: make(map[string]*client),
		rps:     rps,
		burst:   burst,
		logger:  logger,
		done:    make(chan struct{}),
	}

	go func() {
		ticker := time.NewTicker(3 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				rl.cleanup()
			case <-rl.done:
				return
			}
		}
	}()

	return rl
}

func (rl *RateLimiter) getLimiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	c, exists := rl.clients[ip]
	if !exists {
		c = &client{
			limiter: rate.NewLimiter(rate.Limit(rl.rps), rl.burst),
		}
		rl.clients[ip] = c
	}
	c.lastSeen = time.Now()
	return c.limiter
}

func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	evicted := 0
	for ip, c := range rl.clients {
		if time.Since(c.lastSeen) > 10*time.Minute {
			delete(rl.clients, ip)
			evicted++
		}
	}
	if evicted > 0 {
		rl.logger.Info("evicted stale rate limit entries", "count", evicted)
	}
}

func (rl *RateLimiter) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := clientIP(r)
		limiter := rl.getLimiter(ip)

		if !limiter.Allow() {
			w.Header().Set("Retry-After", "1")
			writeError(w, http.StatusTooManyRequests, "rate limit exceeded")
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (rl *RateLimiter) Stop() {
	close(rl.done)
}
