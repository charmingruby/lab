package httpx

import (
	"bytes"
	"context"
	"encoding/json/v2"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var (
	ErrCreateRequest        = errors.New("create request")
	ErrEncodeRequestBody    = errors.New("encode request body")
	ErrDoRequest            = errors.New("do request")
	ErrUnexpectedStatusCode = errors.New("unexpected status code")
	ErrDecodeResponse       = errors.New("decode response")
	ErrMissingBaseURL       = errors.New("missing base URL")
)

type Client struct {
	client  *http.Client
	baseURL string
}

type ClientOption func(*Client)

func NewClient(opts ...ClientOption) *Client {
	c := &Client{
		client: &http.Client{Timeout: 5 * time.Second},
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

func WithHTTPClient(hc *http.Client) ClientOption {
	return func(c *Client) {
		c.client = hc
	}
}

func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) {
		c.baseURL = baseURL
	}
}

type Request struct {
	Body           any
	Headers        http.Header
	Method         string
	URL            string
	ExpectedStatus int
}

func Do[T any](
	ctx context.Context,
	c *Client,
	r Request,
) (*T, error) {
	endpoint, err := c.resolveURL(r.URL)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateRequest, err)
	}

	var bodyReader io.Reader
	if r.Body != nil {
		buf := &bytes.Buffer{}
		if err := json.MarshalWrite(buf, r.Body); err != nil {
			return nil, fmt.Errorf("%w: %w", ErrEncodeRequestBody, err)
		}
		bodyReader = buf
	}

	req, err := http.NewRequestWithContext(ctx, r.Method, endpoint, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateRequest, err)
	}

	req.Header = r.Headers.Clone()
	if req.Header == nil {
		req.Header = make(http.Header)
	}

	if r.Body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	res, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrDoRequest, err)
	}

	defer func() {
		_ = res.Body.Close()
	}()

	if res.StatusCode != r.ExpectedStatus {
		body, _ := io.ReadAll(io.LimitReader(res.Body, 4096))
		return nil, fmt.Errorf(
			"%w: got %d, want %d, body: %s",
			ErrUnexpectedStatusCode,
			res.StatusCode,
			r.ExpectedStatus,
			body,
		)
	}

	if res.StatusCode == http.StatusNoContent {
		var result T
		return &result, nil
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrDecodeResponse, err)
	}

	var result T
	if len(bytes.TrimSpace(body)) == 0 {
		return &result, nil
	}

	if err := json.UnmarshalRead(bytes.NewReader(body), &result); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrDecodeResponse, err)
	}

	return &result, nil
}

func (c *Client) resolveURL(rawURL string) (string, error) {
	if u, err := url.Parse(rawURL); err != nil {
		return "", err
	} else if u.IsAbs() {
		return rawURL, nil
	}

	if c.baseURL == "" {
		return "", ErrMissingBaseURL
	}

	return strings.TrimSuffix(c.baseURL, "/") + "/" + strings.TrimPrefix(rawURL, "/"), nil
}
