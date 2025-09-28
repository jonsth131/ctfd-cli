package api

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

type mockClient struct {
	t      *testing.T
	doFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockClient) Do(req *http.Request) (*http.Response, error) {
	if m.doFunc == nil {
		m.t.Fatalf("mockClient.Do called but doFunc not set")
	}
	return m.doFunc(req)
}

func newResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     make(http.Header),
	}
}

func sequenceResponses(t *testing.T, responses []*http.Response) *mockClient {
	index := 0
	return &mockClient{
		t: t,
		doFunc: func(req *http.Request) (*http.Response, error) {
			if index >= len(responses) {
				t.Fatalf("unexpected extra HTTP call to %s", req.URL.Path)
			}
			resp := responses[index]
			index++
			resp.Request = req
			return resp, nil
		},
	}
}

func mockResponse(t *testing.T, resp *http.Response) *mockClient {
	return sequenceResponses(t, []*http.Response{resp})
}
