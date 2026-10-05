package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Moreira-Henrique-Pedro/entregador/pkg/logger"
)

const (
	DefaultTimeout = 10 * time.Second

	maxResponseBodyBytes = 10 << 20
)

type Client struct {
	httpClient *http.Client
	baseURL    string
	headers    http.Header
	username   string
	password   string
}

type Option func(*Client)

func WithBaseURL(baseURL string) Option {
	return func(c *Client) { c.baseURL = strings.TrimRight(baseURL, "/") }
}

func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) { c.httpClient.Timeout = timeout }
}

func WithBasicAuth(username, password string) Option {
	return func(c *Client) { c.username, c.password = username, password }
}

func WithHeader(key, value string) Option {
	return func(c *Client) { c.headers.Set(key, value) }
}

func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) { c.httpClient = httpClient }
}

func New(opts ...Option) *Client {
	c := &Client{
		httpClient: &http.Client{Timeout: DefaultTimeout},
		headers:    http.Header{},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

type Response struct {
	StatusCode int
	Header     http.Header
	Body       []byte
}

func (r *Response) IsSuccess() bool {
	return r.StatusCode >= 200 && r.StatusCode < 300
}

func (r *Response) DecodeJSON(target any) error {
	if err := json.Unmarshal(r.Body, target); err != nil {
		return fmt.Errorf("decode response body: %w", err)
	}
	return nil
}

func (c *Client) Get(ctx context.Context, path string) (*Response, error) {
	return c.Do(ctx, http.MethodGet, path, nil, nil)
}

func (c *Client) PostJSON(ctx context.Context, path string, payload any) (*Response, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("encode request body: %w", err)
	}
	return c.Do(ctx, http.MethodPost, path, bytes.NewReader(body), http.Header{"Content-Type": {"application/json"}})
}

func (c *Client) PostForm(ctx context.Context, path string, form url.Values) (*Response, error) {
	return c.Do(ctx, http.MethodPost, path, strings.NewReader(form.Encode()),
		http.Header{"Content-Type": {"application/x-www-form-urlencoded"}})
}

func (c *Client) Do(ctx context.Context, method, path string, body io.Reader, headers http.Header) (*Response, error) {
	endpoint := c.baseURL + path

	request, err := http.NewRequestWithContext(ctx, method, endpoint, body)
	if err != nil {
		return nil, fmt.Errorf("build request %s %s: %w", method, endpoint, err)
	}
	maps.Copy(request.Header, c.headers)
	maps.Copy(request.Header, headers)
	if c.username != "" || c.password != "" {
		request.SetBasicAuth(c.username, c.password)
	}

	start := time.Now()
	response, err := c.httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("%s %s: %w", method, request.URL.Redacted(), err)
	}
	defer func() { _ = response.Body.Close() }()

	responseBody, err := io.ReadAll(io.LimitReader(response.Body, maxResponseBodyBytes))
	if err != nil {
		return nil, fmt.Errorf("read response of %s %s: %w", method, request.URL.Redacted(), err)
	}

	logger.GetLoggerFromContext(ctx).Debug("HTTP request completed",
		"method", method,
		"host", request.URL.Host,
		"path", request.URL.Path,
		"status", response.StatusCode,
		"duration_ms", time.Since(start).Milliseconds(),
	)

	return &Response{
		StatusCode: response.StatusCode,
		Header:     response.Header,
		Body:       responseBody,
	}, nil
}
