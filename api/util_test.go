package api

import (
	"strings"
	"testing"
)

func TestExtractTitle(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"<title>Just a moment...</title>", "Just a moment..."},
	}

	for _, test := range tests {
		title, err := extractTitle(test.input)
		if err != nil {
			t.Errorf("extractTitle(%q) returned an error: %v", test.input, err)
		}
		if title != test.expected {
			t.Errorf("extractTitle(%q) = %q, want %q", test.input, title, test.expected)
		}
	}
}

func TestExtractNonce(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		isError  bool
	}{
		{"<input id=\"nonce\" name=\"nonce\" value=\"nonce-value\">", "nonce-value", false},
		{"<input id=\"nonce\" name=\"nonce\">", "", true},
		{"<input>", "", true},
	}

	for _, test := range tests {
		nonce, err := extractNonce(test.input)

		if test.isError && err == nil {
			t.Errorf("extractNonce(%q) should have returned an error", test.input)
		}
		if err != nil && !test.isError {
			t.Errorf("extractNonce(%q) returned an error: %v", test.input, err)
		}
		if nonce != test.expected {
			t.Errorf("extractNonce(%q) = %q, want %q", test.input, nonce, test.expected)
		}
	}
}

func TestExtractCSRFToken(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		isError  bool
	}{
		{"<script>const data = { 'csrfNonce': \"abcdef1234567890\", };</script>", "abcdef1234567890", false},
		{"<input id=\"csrf_token\" name=\"csrf_token\">", "", true},
		{"<input>", "", true},
	}

	for _, test := range tests {
		token, err := extractCSRFToken(test.input)

		if test.isError && err == nil {
			t.Errorf("extractCSRFToken(%q) should have returned an error", test.input)
		}
		if err != nil && !test.isError {
			t.Errorf("extractCSRFToken(%q) returned an error: %v", test.input, err)
		}
		if token != test.expected {
			t.Errorf("extractCSRFToken(%q) = %q, want %q", test.input, token, test.expected)
		}
	}
}

func TestParseBaseUrl(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    string
		expectError bool
		errorMsg    string
	}{
		// Valid cases
		{
			name:     "HTTPS URL with domain",
			input:    "https://ctfd.example.com",
			expected: "https://ctfd.example.com",
		},
		{
			name:     "HTTP URL with domain",
			input:    "http://ctfd.example.com",
			expected: "http://ctfd.example.com",
		},
		{
			name:     "Domain without scheme (defaults to HTTPS)",
			input:    "ctfd.example.com",
			expected: "https://ctfd.example.com",
		},
		{
			name:     "URL with port",
			input:    "https://ctfd.example.com:8080",
			expected: "https://ctfd.example.com:8080",
		},
		{
			name:     "Domain with port without scheme",
			input:    "ctfd.example.com:8080",
			expected: "https://ctfd.example.com:8080",
		},
		{
			name:     "Localhost with port",
			input:    "localhost:8000",
			expected: "https://localhost:8000",
		},
		{
			name:     "HTTP localhost",
			input:    "http://localhost:8000",
			expected: "http://localhost:8000",
		},
		{
			name:     "IP address",
			input:    "192.168.1.100",
			expected: "https://192.168.1.100",
		},
		{
			name:     "HTTP IP address with port",
			input:    "http://192.168.1.100:3000",
			expected: "http://192.168.1.100:3000",
		},
		{
			name:     "URL with path (path should be stripped)",
			input:    "https://ctfd.example.com/some/path",
			expected: "https://ctfd.example.com",
		},
		{
			name:     "URL with query parameters (should be stripped)",
			input:    "https://ctfd.example.com?param=value",
			expected: "https://ctfd.example.com",
		},
		{
			name:     "URL with fragment (should be stripped)",
			input:    "https://ctfd.example.com#section",
			expected: "https://ctfd.example.com",
		},
		{
			name:     "URL with path, query, and fragment",
			input:    "https://ctfd.example.com/path?param=value#section",
			expected: "https://ctfd.example.com",
		},
		{
			name:     "URL with whitespace",
			input:    "  https://ctfd.example.com  ",
			expected: "https://ctfd.example.com",
		},
		{
			name:     "Domain with whitespace",
			input:    "  ctfd.example.com  ",
			expected: "https://ctfd.example.com",
		},
		{
			name:     "IPv6 address with brackets",
			input:    "http://[::1]:8080",
			expected: "http://[::1]:8080",
		},
		{
			name:     "IPv6 address without scheme",
			input:    "[2001:db8::1]:3000",
			expected: "https://[2001:db8::1]:3000",
		},
		{
			name:     "Domain with unusual but valid port",
			input:    "https://example.com:65535",
			expected: "https://example.com:65535",
		},
		{
			name:     "Subdomain with multiple levels",
			input:    "api.v2.ctfd.example.com",
			expected: "https://api.v2.ctfd.example.com",
		},
		{
			name:     "Domain with hyphen",
			input:    "https://ctf-platform.example-site.com",
			expected: "https://ctf-platform.example-site.com",
		},
		{
			name:     "URL with username and password (should be stripped for security)",
			input:    "https://user:pass@ctfd.example.com",
			expected: "https://ctfd.example.com",
		},

		// Error cases
		{
			name:        "Empty string",
			input:       "",
			expectError: true,
			errorMsg:    "URL cannot be empty",
		},
		{
			name:        "Only whitespace",
			input:       "   ",
			expectError: true,
			errorMsg:    "URL cannot be empty",
		},
		{
			name:        "Invalid scheme",
			input:       "ftp://ctfd.example.com",
			expectError: true,
			errorMsg:    "unsupported scheme 'ftp': only http and https are supported",
		},
		{
			name:        "File scheme",
			input:       "file:///path/to/file",
			expectError: true,
			errorMsg:    "unsupported scheme 'file': only http and https are supported",
		},
		{
			name:        "Invalid URL format",
			input:       "ht!tp://invalid-url",
			expectError: true,
			errorMsg:    "invalid URL format",
		},
		{
			name:        "URL with no host (scheme only)",
			input:       "https://",
			expectError: true,
			errorMsg:    "URL must contain a valid host",
		},
		{
			name:        "URL with invalid characters",
			input:       "https://invalid domain.com",
			expectError: true,
			errorMsg:    "invalid URL format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseBaseUrl(tt.input)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for input %q, but got none", tt.input)
					return
				}
				if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing %q, but got %q", tt.errorMsg, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error for input %q: %v", tt.input, err)
				return
			}

			if result == nil {
				t.Errorf("Expected non-nil result for input %q", tt.input)
				return
			}

			if result.String() != tt.expected {
				t.Errorf("For input %q, expected %q but got %q", tt.input, tt.expected, result.String())
			}

			// Additional validations for successful cases
			if result.Scheme == "" {
				t.Errorf("Result should have a scheme for input %q", tt.input)
			}
			if result.Host == "" {
				t.Errorf("Result should have a host for input %q", tt.input)
			}
			if result.Path != "" {
				t.Errorf("Base URL should not have a path, but got %q for input %q", result.Path, tt.input)
			}
			if result.RawQuery != "" {
				t.Errorf("Base URL should not have query parameters, but got %q for input %q", result.RawQuery, tt.input)
			}
			if result.Fragment != "" {
				t.Errorf("Base URL should not have a fragment, but got %q for input %q", result.Fragment, tt.input)
			}
		})
	}
}

func TestParseBaseUrlSchemeValidation(t *testing.T) {
	validSchemes := []string{"http", "https"}
	invalidSchemes := []string{"ftp", "file", "mailto", "ssh", "telnet"}

	for _, scheme := range validSchemes {
		t.Run("valid_scheme_"+scheme, func(t *testing.T) {
			input := scheme + "://example.com"
			result, err := parseBaseUrl(input)
			if err != nil {
				t.Errorf("Expected no error for valid scheme %q, but got: %v", scheme, err)
			}
			if result != nil && result.Scheme != scheme {
				t.Errorf("Expected scheme %q, but got %q", scheme, result.Scheme)
			}
		})
	}

	for _, scheme := range invalidSchemes {
		t.Run("invalid_scheme_"+scheme, func(t *testing.T) {
			input := scheme + "://example.com"
			_, err := parseBaseUrl(input)
			if err == nil {
				t.Errorf("Expected error for invalid scheme %q, but got none", scheme)
			}
			expectedErr := "unsupported scheme '" + scheme + "': only http and https are supported"
			if !strings.Contains(err.Error(), expectedErr) {
				t.Errorf("Expected error %q, but got %q", expectedErr, err.Error())
			}
		})
	}
}

func TestParseBaseUrlDefaultScheme(t *testing.T) {
	inputs := []string{
		"example.com",
		"www.example.com",
		"sub.example.com",
		"example.com:8080",
		"localhost",
		"localhost:3000",
		"192.168.1.1",
		"192.168.1.1:8080",
	}

	for _, input := range inputs {
		t.Run("default_scheme_"+input, func(t *testing.T) {
			result, err := parseBaseUrl(input)
			if err != nil {
				t.Errorf("Unexpected error for input %q: %v", input, err)
				return
			}
			if result.Scheme != "https" {
				t.Errorf("Expected default scheme 'https' for input %q, but got %q", input, result.Scheme)
			}
		})
	}
}

func BenchmarkParseBaseUrl(b *testing.B) {
	testCases := []string{
		"https://ctfd.example.com",
		"ctfd.example.com",
		"https://ctfd.example.com:8080/path?query=value#fragment",
		"localhost:3000",
	}

	for _, tc := range testCases {
		b.Run(tc, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = parseBaseUrl(tc)
			}
		})
	}
}
