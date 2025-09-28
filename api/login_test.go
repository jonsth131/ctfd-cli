package api

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"testing"
)

func TestLogin_Success(t *testing.T) {
	loginPage := `
	<html>
		<head>
			<title>Login</title>
		</head>
		<body>
			<form>
				<input id="nonce" name="nonce" value="test-nonce">
			</form>
		</body>
	</html>`

	loginResponse := `<html><body>Welcome!</body></html>`

	mock := sequenceResponses(t, []*http.Response{
		newResponse(200, loginPage),
		newResponse(200, loginResponse),
	})

	base, _ := url.Parse("https://ctf.example.com")
	api := &ApiClient{
		client:  mock,
		baseUrl: base,
	}

	err := api.Login(context.Background(), "user", "pass")
	if err != nil {
		t.Fatalf("expected no error, but got %v", err)
	}
}

func TestLogin_InvalidCredentials(t *testing.T) {
	loginPage := `
	<html>
		<head>
			<title>Login</title>
		</head>
		<body>
			<input id="nonce" name="nonce" value="bad">
		</body>
	</html>
	`

	invalidResponse := "Your username or password is incorrect"

	mock := sequenceResponses(t, []*http.Response{
		newResponse(200, loginPage),
		newResponse(200, invalidResponse),
	})

	base, _ := url.Parse("https://ctf.example.com")
	api := &ApiClient{
		client:  mock,
		baseUrl: base,
	}

	err := api.Login(context.Background(), "user", "wrongpass")
	if err == nil {
		t.Fatalf("expected error for invalid credentials, got nil")
	}
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("expected error to be ErrInvalidCredentials, got %v", err)
	}
}

func TestLogin_CAPTCHA(t *testing.T) {
	loginPage := `
	<html>
		<head>
			<title>` + cloudflareCAPTCHATitle + `</title>
		</head>
	</html>
	`

	mock := mockResponse(t, newResponse(200, loginPage))

	base, _ := url.Parse("https://ctf.example.com")
	api := &ApiClient{
		client:  mock,
		baseUrl: base,
	}

	err := api.Login(context.Background(), "user", "pass")
	if err == nil {
		t.Fatalf("expected error, but got nil")
	}
	if !errors.Is(err, ErrCaptchaRequired) {
		t.Errorf("expected error to be ErrCaptchaRequired, got %v", err)
	}
}
