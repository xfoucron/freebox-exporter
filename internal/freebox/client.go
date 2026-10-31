package freebox

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	maxBodyBytes   = 4 << 20
	sessionHeader  = "X-Fbx-App-Auth"
	errAuthMissing = "auth_required"
)

type APIError struct {
	Code    string
	Message string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("freebox api: %s: %s", e.Code, e.Message)
}

type envelope struct {
	Success   bool            `json:"success"`
	Result    json.RawMessage `json:"result"`
	ErrorCode string          `json:"error_code"`
	Message   string          `json:"msg"`
}

type Client struct {
	http    *http.Client
	baseURL string
	appID   string
	token   string

	mu      sync.Mutex
	session string
}

func New(ctx context.Context, host, appID, appToken string, httpClient *http.Client) (*Client, error) {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}
	c := &Client{http: httpClient, appID: appID, token: appToken}

	var info struct {
		BaseURL string `json:"api_base_url"`
		Version string `json:"api_version"`
	}
	root := strings.TrimRight(host, "/")
	if err := c.roundTrip(ctx, http.MethodGet, root+"/api_version", "", nil, &info, false); err != nil {
		return nil, fmt.Errorf("discover api version: %w", err)
	}
	major, _, _ := strings.Cut(info.Version, ".")
	if major == "" || info.BaseURL == "" {
		return nil, fmt.Errorf("unexpected api_version response: %+v", info)
	}
	c.baseURL = root + info.BaseURL + "v" + major + "/"
	return c, nil
}

func Get[T any](ctx context.Context, c *Client, path string) (T, error) {
	var out T
	err := c.call(ctx, http.MethodGet, path, nil, &out)
	return out, err
}

func (c *Client) call(ctx context.Context, method, path string, in, out any) error {
	for attempt := 0; ; attempt++ {
		token, err := c.sessionToken(ctx)
		if err != nil {
			return err
		}
		err = c.roundTrip(ctx, method, c.baseURL+path, token, in, out, true)
		var apiErr *APIError
		if attempt == 0 && errors.As(err, &apiErr) && apiErr.Code == errAuthMissing {
			c.invalidate(token)
			continue
		}
		if err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		return nil
	}
}

func (c *Client) roundTrip(ctx context.Context, method, url, session string, in, out any, wrapped bool) error {
	var body io.Reader
	if in != nil {
		payload, err := json.Marshal(in)
		if err != nil {
			return err
		}
		body = bytes.NewReader(payload)
	}
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return err
	}
	if in != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if session != "" {
		req.Header.Set(sessionHeader, session)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxBodyBytes))
	if err != nil {
		return err
	}

	if !wrapped {
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("unexpected status %d", resp.StatusCode)
		}
		return json.Unmarshal(data, out)
	}

	var env envelope
	if err := json.Unmarshal(data, &env); err != nil {
		return fmt.Errorf("decode response (status %d): %w", resp.StatusCode, err)
	}
	if !env.Success {
		return &APIError{Code: env.ErrorCode, Message: env.Message}
	}
	if out == nil || len(env.Result) == 0 {
		return nil
	}
	return json.Unmarshal(env.Result, out)
}

func (c *Client) sessionToken(ctx context.Context) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.session != "" {
		return c.session, nil
	}
	token, err := c.login(ctx)
	if err != nil {
		return "", fmt.Errorf("login: %w", err)
	}
	c.session = token
	return token, nil
}

func (c *Client) invalidate(token string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.session == token {
		c.session = ""
	}
}
