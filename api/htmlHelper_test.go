package api

import (
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
