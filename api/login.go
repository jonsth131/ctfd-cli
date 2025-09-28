package api

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

func (c *ApiClient) Login(ctx context.Context, name string, password string) error {
	if strings.TrimSpace(name) == "" {
		return ErrInvalidUsername
	}
	if strings.TrimSpace(password) == "" {
		return ErrInvalidPassword
	}

	bodyString, err := c.getLoginPageBody(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", errFailedToGetLoginPage, err)
	}

	res, err := checkCAPTCHA(bodyString)
	if err != nil {
		return fmt.Errorf("%s: %w", errFailedToCheckCAPTCHA, err)
	}

	if res {
		return ErrCaptchaRequired
	}

	nonce, err := extractNonce(bodyString)
	if err != nil {
		return fmt.Errorf("%s: %w", errFailedToExtractNonce, err)
	}

	err = c.performLogin(ctx, name, password, nonce)
	if err != nil {
		return fmt.Errorf("%s: %w", errFailedToLogin, err)
	}

	return nil
}

func (c *ApiClient) getLoginPageBody(ctx context.Context) (string, error) {
	u := fmt.Sprintf("%s%s", c.baseUrl, loginURL)

	resp, err := c.get(ctx, u)

	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("%s: %v", errFailedToReadResponseBody, err)
	}

	bodyString := string(bodyBytes)

	if bodyString == "" {
		return "", fmt.Errorf(errEmptyResponseBody)
	}

	return bodyString, nil
}

func checkCAPTCHA(body string) (bool, error) {
	title, err := extractTitle(body)
	if err != nil {
		return false, fmt.Errorf("%s: %w", errFailedToExtractTitle, err)
	}

	if strings.Contains(title, cloudflareCAPTCHATitle) {
		return true, nil
	}

	return false, nil
}

func (c *ApiClient) performLogin(ctx context.Context, username string, password string, nonce string) error {
	body := url.Values{
		"name":     {username},
		"password": {password},
		"nonce":    {nonce},
	}

	resp, err := c.postForm(ctx, fmt.Sprintf("%s%s", c.baseUrl, loginURL), body)

	if err != nil {
		if errors.Is(err, context.Canceled) {
			return fmt.Errorf(errLoginCancelled)
		}
		if errors.Is(err, context.DeadlineExceeded) {
			return fmt.Errorf(errLoginTimeout)
		}
		return fmt.Errorf("%s: %v", errFailedToLogin, err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s: %v", errFailedToLogin, resp.Status)
	}

	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("%s: %v", errFailedToReadResponseBody, err)
	}
	bodyString := string(bodyBytes)

	if strings.Contains(bodyString, ctfdInvalidCredentials) {
		return ErrInvalidCredentials
	}

	return nil
}
