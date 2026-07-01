package main

import (
	"sync"
	"time"
)

// RateLimiter implements a token bucket rate limiter with burst support.
// It controls the rate at which requests are sent to avoid triggering
// WAF/IDS rate-based detection mechanisms.
type RateLimiter struct {
	mu          sync.Mutex
	tokens      float64   // Current number of tokens in the bucket.
	maxTokens   float64   // Maximum tokens (burst size).
	refillRate  float64   // Tokens added per second.
	lastRefill  time.Time // Last time tokens were refilled.
	burstDelay  time.Duration // Delay applied when burst is exhausted.
	perRequest  time.Duration // Base delay between requests.
	jitterFactor float64      // Jitter factor for delay randomization.
}

// NewRateLimiter creates a new rate limiter with the specified parameters.
//
// Parameters:
//   - perRequest: base delay between individual requests.
//   - burstSize: number of requests allowed in a burst before throttling.
//   - burstDelay: delay applied when burst tokens are exhausted.
//   - jitterFactor: random variation factor (0.0-1.0) for delays.
func NewRateLimiter(perRequest time.Duration, burstSize int, burstDelay time.Duration, jitterFactor float64) *RateLimiter {
	refillRate := 1.0 / perRequest.Seconds() // tokens per second

	return &RateLimiter{
		tokens:       float64(burstSize),
		maxTokens:    float64(burstSize),
		refillRate:   refillRate,
		lastRefill:   time.Now(),
		burstDelay:   burstDelay,
		perRequest:   perRequest,
		jitterFactor: jitterFactor,
	}
}

// Wait blocks until a token is available, enforcing rate limiting.
// This is the primary method called before each HTTP request.
func (rl *RateLimiter) Wait() {
	// Phase 1: Wait for an available token (may need to wait for refill + burst delay).
	for {
		waitDuration := rl.tryAcquire()
		if waitDuration == 0 {
			break // Token acquired.
		}
		time.Sleep(waitDuration)
	}

	// Phase 2: Apply per-request delay with jitter (pacing between requests).
	if rl.perRequest > 0 {
		delay := rl.perRequest
		if rl.jitterFactor > 0 {
			delay = applyJitter(delay, rl.jitterFactor)
		}
		time.Sleep(delay)
	}
}

// tryAcquire attempts to consume a token. Returns 0 if successful,
// or a duration to wait before retrying.
func (rl *RateLimiter) tryAcquire() time.Duration {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(rl.lastRefill).Seconds()
	rl.tokens += elapsed * rl.refillRate
	if rl.tokens > rl.maxTokens {
		rl.tokens = rl.maxTokens
	}
	rl.lastRefill = now

	// If we have at least one token, consume it.
	if rl.tokens >= 1.0 {
		rl.tokens -= 1.0
		return 0 // Token acquired; per-request delay is handled by Wait().
	}

	// No tokens available — burst exhausted.
	// Calculate how long until next token is available.
	deficit := 1.0 - rl.tokens
	waitSec := deficit / rl.refillRate
	wait := time.Duration(waitSec * float64(time.Second))

	// Add burst delay penalty.
	if rl.burstDelay > 0 {
		burst := rl.burstDelay
		if rl.jitterFactor > 0 {
			burst = applyJitter(burst, rl.jitterFactor)
		}
		wait += burst
	}

	return wait
}

// AvailableTokens returns the current number of available tokens (for monitoring).
func (rl *RateLimiter) AvailableTokens() float64 {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(rl.lastRefill).Seconds()
	tokens := rl.tokens + elapsed*rl.refillRate
	if tokens > rl.maxTokens {
		tokens = rl.maxTokens
	}
	return tokens
}