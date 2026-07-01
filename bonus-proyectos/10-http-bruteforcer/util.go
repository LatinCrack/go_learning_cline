package main

import (
	"fmt"
	"math"
	"math/rand"
	"time"
)

// applyJitter adds random jitter to a duration.
// factor is the fraction of the duration that can vary (0.0 to 1.0).
// e.g. factor=0.3 on 1000ms yields 700ms–1300ms.
func applyJitter(d time.Duration, factor float64) time.Duration {
	if factor <= 0 {
		return d
	}
	jitter := float64(d) * factor
	offset := (rand.Float64()*2 - 1) * jitter // range [-jitter, +jitter]
	result := float64(d) + offset
	if result < 0 {
		result = 0
	}
	return time.Duration(math.Round(result))
}

// truncateString truncates a string to maxLen characters, appending "..." if truncated.
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// contains checks if a string contains a substring (case-sensitive).
func contains(s, substr string) bool {
	return len(substr) > 0 && len(s) >= len(substr) && findSubstring(s, substr)
}

// findSubstring performs a simple substring search.
func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// formatDuration formats a duration into a human-readable string.
func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm%ds", int(d.Minutes()), int(d.Seconds())%60)
	}
	return fmt.Sprintf("%dh%dm%ds", int(d.Hours()), int(d.Minutes())%60, int(d.Seconds())%60)
}