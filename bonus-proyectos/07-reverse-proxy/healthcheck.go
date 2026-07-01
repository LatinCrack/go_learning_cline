package main

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"
)

// HealthChecker runs periodic health probes against all registered backends
// in its own goroutine. Unresponsive backends are temporarily removed from
// the load-balancing pool by marking them as unhealthy (thread-safe).
type HealthChecker struct {
	balancer *Balancer
	interval time.Duration
	timeout  time.Duration
	client   *http.Client
	stopCh   chan struct{}
	wg       sync.WaitGroup
}

// NewHealthChecker creates a HealthChecker that probes backends at the given
// interval using the provided timeout per probe.
func NewHealthChecker(balancer *Balancer, interval, timeout time.Duration) *HealthChecker {
	return &HealthChecker{
		balancer: balancer,
		interval: interval,
		timeout:  timeout,
		client: &http.Client{
			Timeout: timeout,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse // don't follow redirects
			},
		},
		stopCh: make(chan struct{}),
	}
}

// Start launches the background health-check goroutine.
func (hc *HealthChecker) Start() {
	hc.wg.Add(1)
	go hc.run()
	log.Printf("[healthcheck] started — interval: %s, timeout: %s", hc.interval, hc.timeout)
}

// Stop signals the health-check goroutine to exit and waits for it.
func (hc *HealthChecker) Stop() {
	close(hc.stopCh)
	hc.wg.Wait()
	log.Println("[healthcheck] stopped")
}

// run is the main loop of the health-check goroutine. It uses a Ticker to
// probe backends periodically until Stop is called.
func (hc *HealthChecker) run() {
	defer hc.wg.Done()

	// Perform an immediate check on startup.
	hc.checkAll()

	ticker := time.NewTicker(hc.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			hc.checkAll()
		case <-hc.stopCh:
			return
		}
	}
}

// checkAll probes every registered backend and updates its Alive state.
func (hc *HealthChecker) checkAll() {
	backends := hc.balancer.Backends()

	var wg sync.WaitGroup
	for _, backend := range backends {
		wg.Add(1)
		go func(b *Backend) {
			defer wg.Done()
			hc.checkBackend(b)
		}(backend)
	}
	wg.Wait()
}

// checkBackend sends an HTTP GET to the backend root and marks it up or down
// based on the response. Uses a short-lived context for timeout control.
func (hc *HealthChecker) checkBackend(b *Backend) {
	ctx, cancel := context.WithTimeout(context.Background(), hc.timeout)
	defer cancel()

	probeURL := b.URL.String()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, probeURL, nil)
	if err != nil {
		log.Printf("[healthcheck] failed to build request for %s: %v", probeURL, err)
		hc.balancer.MarkDown(b.URL)
		return
	}

	resp, err := hc.client.Do(req)
	if err != nil {
		if b.Alive.Load() {
			log.Printf("[healthcheck] backend %s is unreachable: %v", probeURL, err)
			hc.balancer.MarkDown(b.URL)
		}
		return
	}
	defer resp.Body.Close()

	// Any 2xx or 3xx response is considered healthy.
	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		if !b.Alive.Load() {
			log.Printf("[healthcheck] backend %s recovered (HTTP %d)", probeURL, resp.StatusCode)
			hc.balancer.MarkUp(b.URL)
		}
	} else {
		if b.Alive.Load() {
			log.Printf("[healthcheck] backend %s returned HTTP %d — marking DOWN", probeURL, resp.StatusCode)
			hc.balancer.MarkDown(b.URL)
		}
	}
}