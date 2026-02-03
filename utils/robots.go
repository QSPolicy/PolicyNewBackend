package utils

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/temoto/robotstxt"
	"go.uber.org/zap"
)

// RobotsValidator checks if a URL is allowed to be visited by a specific user agent
type RobotsValidator struct {
	client    *http.Client
	cache     sync.Map // map[string]*robotstxt.RobotsData
	userAgent string
}

// NewRobotsValidator creates a new validator instance
func NewRobotsValidator(userAgent string, timeout time.Duration) *RobotsValidator {
	if timeout == 0 {
		timeout = 10 * time.Second
	}
	return &RobotsValidator{
		client: &http.Client{
			Timeout: timeout,
		},
		userAgent: userAgent,
	}
}

// IsAllowed checks if the given URL is allowed for the user agent
func (r *RobotsValidator) IsAllowed(targetURL string) (bool, error) {
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		return false, fmt.Errorf("invalid URL: %w", err)
	}

	host := parsedURL.Host
	scheme := parsedURL.Scheme
	if scheme == "" {
		scheme = "http"
	}

	// Check cache first
	if val, ok := r.cache.Load(host); ok {
		if data, ok := val.(*robotstxt.RobotsData); ok {
			return data.TestAgent(parsedURL.Path, r.userAgent), nil
		}
	}

	// Fetch robots.txt
	robotsURL := fmt.Sprintf("%s://%s/robots.txt", scheme, host)
	resp, err := r.client.Get(robotsURL)
	if err != nil {
		// If we can't fetch robots.txt, we assume we can crawl (or fail depending on policy)
		// Standard practice: if connection failure, maybe fail closed or open.
		// Here we log and fail open (allow) if it's just a network error on robots.txt?
		// Actually, Google treats 5xx as "disallow all" (temp), 4xx as "allow all".
		// But for a simple tool, let's just log and assume allowed if 404, disallowed if error.
		// For safety, let's treat network errors as 'allow' but log warning,
		// Or strictly follow:
		// 404 (Not Found) -> Allow all
		// 401/403 (Unauthorized) -> Disallow all
		// Other errors -> assume allowed or disallowed?

		// Let's take a safe approach:
		// If fetch fails (DNS, timeout), strict bots might stop.
		// But often simpler bots just proceed.
		// Let's return error so caller decides.
		return true, nil // Default to allow if robots.txt is unreachable/error?
		// Re-reading: "Treat other errors as full allow".
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 && resp.StatusCode < 500 {
		// 4xx implies no robots.txt or access denied to it.
		// If 401/403, we probably can't crawl the site anyway, but technically robots.txt isn't blocking.
		// Standard: 4xx => allow all.
		// We'll cache an empty/allow-all rule set effectively.
		r.cache.Store(host, allowAllRobots())
		return true, nil
	}

	if resp.StatusCode >= 500 {
		// Server error, should treat as full disallow to be polite
		return false, fmt.Errorf("server error fetching robots.txt: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("failed to read robots.txt body: %w", err)
	}

	data, err := robotstxt.FromBytes(body)
	if err != nil {
		// Parsing error, allow all?
		if Log != nil {
			Log.Warn("failed to parse robots.txt", zap.String("url", targetURL), zap.Error(err))
		}
		return true, nil
	}

	r.cache.Store(host, data)
	return data.TestAgent(parsedURL.Path, r.userAgent), nil
}

// Helper to create an allow-all config
func allowAllRobots() *robotstxt.RobotsData {
	data, _ := robotstxt.FromBytes([]byte("User-agent: *\nAllow: /"))
	return data
}
