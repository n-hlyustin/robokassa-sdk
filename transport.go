package robokassa

import (
	"bytes"
	"context"
	"io"
	"net/http"
)

type httpResponse struct {
	Body   []byte
	Status int
}

type transport struct {
	httpClient *http.Client
}

func (t *transport) get(ctx context.Context, url string, headers map[string]string) (*httpResponse, error) {
	return t.do(ctx, http.MethodGet, url, nil, headers)
}

func (t *transport) post(ctx context.Context, url string, body []byte, headers map[string]string) (*httpResponse, error) {
	return t.do(ctx, http.MethodPost, url, body, headers)
}

func (t *transport) do(ctx context.Context, method, url string, body []byte, headers map[string]string) (*httpResponse, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(body))
	if err != nil {
		return nil, &SDKError{Op: "transport.new_request", Message: "failed to build request", Err: err}
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return nil, &SDKError{Op: "transport.do", Message: "request failed", Err: err}
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &SDKError{Op: "transport.read_body", StatusCode: resp.StatusCode, Message: "failed to read response body", Err: err}
	}

	return &httpResponse{
		Body:   data,
		Status: resp.StatusCode,
	}, nil
}
