package api

import (
	"fmt"
	"net/url"
	"strings"
)

func parseBaseUrl(u string) (*url.URL, error) {
	u = strings.TrimSpace(u)

	if u == "" {
		return nil, fmt.Errorf("URL cannot be empty")
	}

	if !strings.Contains(u, "://") {
		u = "https://" + u
	}

	ur, err := url.Parse(u)
	if err != nil {
		return nil, fmt.Errorf("invalid URL format: %w", err)
	}

	if ur.Scheme != "http" && ur.Scheme != "https" {
		return nil, fmt.Errorf("unsupported scheme '%s': only http and https are supported", ur.Scheme)
	}

	if ur.Host == "" {
		return nil, fmt.Errorf("URL must contain a valid host")
	}

	baseURL := &url.URL{
		Scheme: ur.Scheme,
		Host:   ur.Host,
	}

	return baseURL, nil
}
