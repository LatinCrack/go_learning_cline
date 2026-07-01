package main

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// AttemptResult holds the outcome of a single bruteforce attempt.
type AttemptResult struct {
	Username  string
	Password  string
	Status    string        // "success", "failure", "error", "blocked"
	HTTPCode  int
	Duration  time.Duration
	Error     error
	Timestamp time.Time
}

// Stats holds real-time statistics for the bruteforce attack.
type Stats struct {
	TotalAttempts   atomic.Int64
	SuccessAttempts atomic.Int64
	FailedAttempts  atomic.Int64
	ErrorAttempts   atomic.Int64
	BlockedAttempts atomic.Int64
	TotalDuration   atomic.Int64 // nanoseconds
}

// Bruteforcer orchestrates the concurrent bruteforce attack.
type Bruteforcer struct {
	config    *Config
	client    *HTTPClient
	detector  *Detector
	rateLimit *RateLimiter
	reporter  *Reporter
	stats     *Stats

	// Results channel for the reporter to consume.
	results chan AttemptResult
	// Found credentials.
	found     []AttemptResult
	foundMu   sync.Mutex
	// Stop signal.
	stopped   atomic.Bool
}

// NewBruteforcer creates a new Bruteforcer from the given Config.
func NewBruteforcer(cfg *Config) (*Bruteforcer, error) {
	client, err := NewHTTPClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("create HTTP client: %w", err)
	}

	rateLimit := NewRateLimiter(
		time.Duration(cfg.Delay)*time.Millisecond,
		cfg.BurstSize,
		time.Duration(cfg.BurstDelay)*time.Millisecond,
		cfg.JitterFactor,
	)

	reporter := NewReporter(cfg)

	bf := &Bruteforcer{
		config:    cfg,
		client:    client,
		detector:  NewDetector(cfg),
		rateLimit: rateLimit,
		reporter:  reporter,
		stats:     &Stats{},
		results:   make(chan AttemptResult, cfg.Workers*2),
	}

	return bf, nil
}

// Run executes the bruteforce attack with the given credential pairs.
// It blocks until all pairs are processed, or the context is cancelled.
// Returns a slice of successful credentials.
func (bf *Bruteforcer) Run(ctx context.Context, pairs <-chan EntryPair) []AttemptResult {
	startTime := time.Now()

	// Print attack banner.
	bf.reporter.PrintBanner(bf.config)

	// Start the reporter goroutine.
	go bf.reporter.Start(bf.results, bf.stats, startTime)

	// Start worker goroutines.
	var wg sync.WaitGroup
	for i := 0; i < bf.config.Workers; i++ {
		wg.Add(1)
		go bf.worker(ctx, i, pairs, &wg)
	}

	// Wait for all workers to finish.
	wg.Wait()
	close(bf.results)

	// Give reporter a moment to flush.
	time.Sleep(100 * time.Millisecond)

	totalDuration := time.Since(startTime)
	bf.reporter.PrintSummary(bf.stats, bf.found, totalDuration)

	return bf.found
}

// worker processes credential pairs from the channel.
func (bf *Bruteforcer) worker(ctx context.Context, id int, pairs <-chan EntryPair, wg *sync.WaitGroup) {
	defer wg.Done()

	for pair := range pairs {
		// Check for context cancellation (Ctrl+C or timeout).
		if ctx.Err() != nil || bf.stopped.Load() {
			return
		}

		// Rate limit: wait for permission to send.
		bf.rateLimit.Wait()

		// Perform the login attempt with retries.
		result := bf.attemptWithRetry(ctx, pair)

		// Update statistics.
		bf.stats.TotalAttempts.Add(1)
		switch result.Status {
		case "success":
			bf.stats.SuccessAttempts.Add(1)
			bf.foundMu.Lock()
			bf.found = append(bf.found, result)
			bf.foundMu.Unlock()
		case "failure":
			bf.stats.FailedAttempts.Add(1)
		case "blocked":
			bf.stats.BlockedAttempts.Add(1)
		case "error":
			bf.stats.ErrorAttempts.Add(1)
		}
		bf.stats.TotalDuration.Add(result.Duration.Nanoseconds())

		// Send to reporter.
		bf.results <- result

		// If we found valid credentials and the config says to stop, stop.
		if result.Status == "success" {
			bf.stopped.Store(true)
			// Drain remaining pairs to unblock the generator.
			go func() {
				for range pairs {
				}
			}()
			return
		}

		// If blocked, apply extra backoff.
		if result.Status == "blocked" {
			backoff := time.Duration(bf.config.BurstDelay) * time.Millisecond * 3
			if bf.config.JitterEnabled {
				backoff = applyJitter(backoff, bf.config.JitterFactor)
			}
			select {
			case <-ctx.Done():
				return
			case <-time.After(backoff):
			}
		}
	}
}

// attemptWithRetry performs a login attempt with automatic retries on network errors.
func (bf *Bruteforcer) attemptWithRetry(ctx context.Context, pair EntryPair) AttemptResult {
	maxRetries := bf.config.MaxRetries
	if maxRetries < 1 {
		maxRetries = 1
	}

	var lastResult AttemptResult

	for attempt := 0; attempt < maxRetries; attempt++ {
		if ctx.Err() != nil {
			return AttemptResult{
				Username:  pair.Username,
				Password:  pair.Password,
				Status:    "error",
				Error:     ctx.Err(),
				Timestamp: time.Now(),
			}
		}

		loginResp := bf.client.AttemptLogin(pair.Username, pair.Password)

		result := AttemptResult{
			Username:  pair.Username,
			Password:  pair.Password,
			HTTPCode:  loginResp.StatusCode,
			Duration:  loginResp.Duration,
			Error:     loginResp.Error,
			Timestamp: time.Now(),
		}

		if loginResp.Error != nil {
			result.Status = "error"
			lastResult = result
			// Backoff before retry.
			if attempt < maxRetries-1 {
				retryDelay := time.Duration(500*(attempt+1)) * time.Millisecond
				select {
				case <-ctx.Done():
					return result
				case <-time.After(retryDelay):
				}
			}
			continue
		}

		result.Status = bf.detector.DetectResult(loginResp, pair.Username, pair.Password)
		return result
	}

	return lastResult
}

// Results returns the channel of results (for external consumers).
func (bf *Bruteforcer) Results() <-chan AttemptResult {
	return bf.results
}

// Close cleans up the Bruteforcer resources.
func (bf *Bruteforcer) Close() {
	bf.client.Close()
}