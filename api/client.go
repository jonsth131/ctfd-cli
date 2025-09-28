package api

import (
	"context"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
)

type CTFdClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type ApiClient struct {
	client  CTFdClient
	baseUrl *url.URL
}

func NewApiClient(u string) (*ApiClient, error) {
	ur, err := parseBaseUrl(u)
	if err != nil {
		return nil, err
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	httpClient := &http.Client{Jar: jar}

	return &ApiClient{client: httpClient, baseUrl: ur}, nil
}

func (c *ApiClient) get(ctx context.Context, fullURL string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
	if err != nil {
		return nil, err
	}
	return c.client.Do(req)
}

func (c *ApiClient) post(ctx context.Context, fullURL, bodyType string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", fullURL, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", bodyType)
	return c.client.Do(req)
}

func (c *ApiClient) postForm(ctx context.Context, fullURL string, data url.Values) (*http.Response, error) {
	return c.post(ctx, fullURL, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
}
