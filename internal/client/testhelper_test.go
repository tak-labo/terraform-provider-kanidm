package client

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

type mockRoundTripper struct {
	handler func(*http.Request) (*http.Response, error)
}

func (m *mockRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	return m.handler(r)
}

func newTestClient(t *testing.T, handler func(*http.Request) (*http.Response, error)) *Client {
	t.Helper()
	httpClient := &http.Client{Transport: &mockRoundTripper{handler: handler}}
	return NewClient("https://idm.example.com", "test-token", WithHTTPClient(httpClient))
}

func jsonResponse(status int, body any) *http.Response {
	b, _ := json.Marshal(body)
	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(strings.NewReader(string(b))),
		Header:     make(http.Header),
	}
}

func emptyResponse(status int) *http.Response {
	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(strings.NewReader("")),
		Header:     make(http.Header),
	}
}

func stringBody(s string) io.ReadCloser {
	return io.NopCloser(strings.NewReader(s))
}
