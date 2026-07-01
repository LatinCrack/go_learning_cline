package main

import (
	"fmt"
	"net/http"
	"strings"
)

// Detector analyzes HTTP responses to determine if a login attempt
// was successful or failed. It supports multiple detection strategies:
//   - Response body content matching (success/failure indicators)
//   - HTTP status code analysis
//   - Redirect detection (Location header)
type Detector struct {
	config *Config
}

// NewDetector creates a new login result Detector.
func NewDetector(cfg *Config) *Detector {
	return &Detector{config: cfg}
}

// DetectResult analyzes a LoginResponse and determines the outcome.
// Returns:
//   - "success": credentials are valid
//   - "failure": credentials are invalid
//   - "error": unable to determine (network error, ambiguous response, rate-limited, etc.)
//   - "blocked": IP blocked or rate-limited by server (429, 403, etc.)
func (d *Detector) DetectResult(resp *LoginResponse, username, password string) string {
	if resp.Error != nil {
		return "error"
	}

	// Check for rate limiting or blocking responses.
	if resp.StatusCode == http.StatusTooManyRequests { // 429
		return "blocked"
	}
	if resp.StatusCode == http.StatusForbidden { // 403
		return "blocked"
	}
	if resp.StatusCode == http.StatusServiceUnavailable { // 503
		return "blocked"
	}

	// Strategy 1: Check HTTP status code for success.
	if d.config.SuccessStatus > 0 && resp.StatusCode == d.config.SuccessStatus {
		return "success"
	}

	// Strategy 2: Check for success indicator in response body.
	if d.config.SuccessIndicator != "" && contains(resp.Body, d.config.SuccessIndicator) {
		return "success"
	}

	// Strategy 3: Check for failure indicator in response body.
	if d.config.FailureIndicator != "" && contains(resp.Body, d.config.FailureIndicator) {
		return "failure"
	}

	// Strategy 4: Redirect analysis — many login forms redirect on success.
	if resp.StatusCode >= 300 && resp.StatusCode < 400 {
		location := resp.Location
		if location != "" {
			// If redirected away from login page, consider it success.
			if !contains(location, "login") && !contains(location, "signin") && !contains(location, "error") {
				return "success"
			}
			// If redirected back to login or error page, it's a failure.
			return "failure"
		}
	}

	// Strategy 5: Status code heuristics.
	// 200 on login page usually means failure (form re-rendered).
	// 200 with no indicator — ambiguous, treat as failure.
	if resp.StatusCode == http.StatusOK {
		return "failure"
	}

	// Unable to determine.
	return "error"
}

// DetectLoginForm attempts to find login form fields in an HTML page.
// Returns the detected username field name, password field name, and form action URL.
type FormDetection struct {
	UsernameField string
	PasswordField string
	FormAction    string
	Found         bool
}

// DetectLoginForm parses HTML to find a login form.
// It looks for input fields with common names/types associated with authentication.
func DetectLoginForm(htmlBody string, baseURL string) FormDetection {
	result := FormDetection{}

	// Convert to lowercase for case-insensitive matching.
	lower := strings.ToLower(htmlBody)

	// Look for password input — this is the strongest indicator of a login form.
	passwordFields := []string{
		`type="password"`,
		`type='password'`,
		`type=password`,
	}

	hasPasswordField := false
	for _, pf := range passwordFields {
		if strings.Contains(lower, pf) {
			hasPasswordField = true
			break
		}
	}

	if !hasPasswordField {
		return result
	}

	// Try to extract the password field name.
	result.PasswordField = extractInputName(lower, "password")
	if result.PasswordField == "" {
		result.PasswordField = "password"
	}

	// Try to find the username field — look for common patterns.
	// Look for text/email input before the password field.
	pwIdx := strings.Index(lower, "type=\"password\"")
	if pwIdx == -1 {
		pwIdx = strings.Index(lower, "type='password'")
	}

	usernameCandidates := []string{"username", "user", "login", "email", "account", "usr", "name"}
	bestScore := 0
	bestField := ""

	// Search for input fields in the vicinity of the password field.
	searchArea := lower
	if pwIdx > 2000 {
		searchArea = lower[pwIdx-2000:]
	}

	for _, candidate := range usernameCandidates {
		score := 0
		if strings.Contains(searchArea, "name=\""+candidate+"\"") || strings.Contains(searchArea, "name='"+candidate+"'") {
			score = 10
		}
		if strings.Contains(searchArea, "id=\""+candidate+"\"") || strings.Contains(searchArea, "id='"+candidate+"'") {
			score += 5
		}
		if score > bestScore {
			bestScore = score
			bestField = candidate
		}
	}

	if bestField != "" {
		result.UsernameField = bestField
	} else {
		// Try to extract the first text/email input field name.
		result.UsernameField = extractFirstTextOrEmailInput(lower)
		if result.UsernameField == "" {
			result.UsernameField = "username"
		}
	}

	// Try to find form action.
	result.FormAction = extractFormAction(lower, baseURL)
	result.Found = true

	return result
}

// extractInputName finds the name attribute of an input field near the given type.
func extractInputName(html, inputType string) string {
	// Find the type attribute position.
	idx := strings.Index(html, `type="`+inputType+`"`)
	if idx == -1 {
		return ""
	}

	// Look backwards for the name attribute within the same <input> tag.
	// Go back up to 500 chars.
	start := idx - 500
	if start < 0 {
		start = 0 }
	segment := html[start : idx+50+len(inputType)]

	// Find the last name= before our type.
	nameIdx := strings.LastIndex(segment, `name="`)
	if nameIdx == -1 {
		nameIdx = strings.LastIndex(segment, `name='`)
		if nameIdx == -1 {
			return ""
		}
		// Extract value between single quotes.
		valueStart := nameIdx + 6
		valueEnd := strings.Index(segment[valueStart:], "'")
		if valueEnd == -1 {
			return ""
		}
		return segment[valueStart : valueStart+valueEnd]
	}

	// Extract value between double quotes.
	valueStart := nameIdx + 6
	valueEnd := strings.Index(segment[valueStart:], "\"")
	if valueEnd == -1 {
		return ""
	}
	return segment[valueStart : valueStart+valueEnd]
}

// extractFirstTextOrEmailInput finds the name of the first text or email input field.
func extractFirstTextOrEmailInput(html string) string {
	patterns := []string{`type="text"`, `type="email"`, `type="tel"`}
	bestIdx := len(html)

	for _, pattern := range patterns {
		idx := strings.Index(html, pattern)
		if idx != -1 && idx < bestIdx {
			bestIdx = idx
		}
	}

	if bestIdx == len(html) {
		return ""
	}

	segment := html[bestIdx:]
	nameIdx := strings.Index(segment, `name="`)
	if nameIdx == -1 {
		return ""
	}

	valueStart := nameIdx + 6
	valueEnd := strings.Index(segment[valueStart:], "\"")
	if valueEnd == -1 {
		return ""
	}
	return segment[valueStart : valueStart+valueEnd]
}

// extractFormAction finds the action attribute of the enclosing form.
func extractFormAction(html, baseURL string) string {
	// Find the last <form before the password field.
	pwIdx := strings.Index(html, "type=\"password\"")
	if pwIdx == -1 {
		return ""
	}

	searchArea := html[:pwIdx]
	formIdx := strings.LastIndex(searchArea, "<form")
	if formIdx == -1 {
		return ""
	}

	formTag := searchArea[formIdx:]

	// Look for action= attribute.
	actionIdx := strings.Index(formTag, "action=\"")
	if actionIdx == -1 {
		actionIdx = strings.Index(formTag, "action='")
		if actionIdx == -1 {
			return ""
		}
		valueStart := actionIdx + 8
		valueEnd := strings.Index(formTag[valueStart:], "'")
		if valueEnd == -1 {
			return ""
		}
		action := formTag[valueStart : valueStart+valueEnd]
		return resolveURL(baseURL, action)
	}

	valueStart := actionIdx + 8
	valueEnd := strings.Index(formTag[valueStart:], "\"")
	if valueEnd == -1 {
		return ""
	}
	action := formTag[valueStart : valueStart+valueEnd]
	return resolveURL(baseURL, action)
}

// resolveURL handles relative URLs by joining them with the base URL.
func resolveURL(base, relative string) string {
	if relative == "" {
		return base
	}
	if strings.HasPrefix(relative, "http://") || strings.HasPrefix(relative, "https://") {
		return relative
	}
	// Simple join — strip path from base and append relative.
	if !strings.HasPrefix(relative, "/") {
		relative = "/" + relative
	}
	// Find the host part of the base URL.
	idx := strings.Index(base, "://")
	if idx == -1 {
		return relative
	}
	hostPart := base[idx+3:]
	slashIdx := strings.Index(hostPart, "/")
	if slashIdx == -1 {
		return base + relative
	}
	return base[:idx+3+slashIdx] + relative
}

// String returns a human-readable summary of a FormDetection.
func (fd FormDetection) String() string {
	if !fd.Found {
		return "No login form detected"
	}
	return fmt.Sprintf("Login form detected: username=%q password=%q action=%q",
		fd.UsernameField, fd.PasswordField, fd.FormAction)
}