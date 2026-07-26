// Package higress 通过 Higress Console API 管理网关配置
package higress

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/lgc202/gateway-agent/internal/chatsvc/config"
)

const requestTimeout = 10 * time.Second

// Client 是 Higress Console 客户端
type Client struct {
	endpoint   *url.URL
	username   string
	password   string
	httpClient *http.Client
}

// NewClient 创建使用 Basic Authentication 的 Higress Console 客户端
func NewClient(cfg *config.Config) (*Client, error) {
	parsedEndpoint, err := url.ParseRequestURI(strings.TrimSpace(cfg.Higress.Endpoint))
	if err != nil || parsedEndpoint.Host == "" ||
		(parsedEndpoint.Scheme != "http" && parsedEndpoint.Scheme != "https") {
		return nil, fmt.Errorf("invalid higress console endpoint")
	}
	if cfg.Higress.Username == "" || cfg.Higress.Password == "" {
		return nil, fmt.Errorf("higress console credentials are required")
	}

	return &Client{
		endpoint: parsedEndpoint,
		username: cfg.Higress.Username,
		password: cfg.Higress.Password,
		httpClient: &http.Client{
			Timeout: requestTimeout,
		},
	}, nil
}

// request 向 Higress Console 发送请求，并返回 HTTP 状态码供资源方法解释具体语义
func (c *Client) request(ctx context.Context, method string, requestURL *url.URL, body any, result any) (int, error) {
	var payload []byte
	if body != nil {
		var err error
		payload, err = json.Marshal(body)
		if err != nil {
			return 0, fmt.Errorf("encode higress console request: %w", err)
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, requestURL.String(), bytes.NewReader(payload))
	if err != nil {
		return 0, fmt.Errorf("create higress console request: %w", err)
	}
	req.SetBasicAuth(c.username, c.password)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("request higress console: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return resp.StatusCode, fmt.Errorf("higress console returned HTTP status %d", resp.StatusCode)
	}
	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return resp.StatusCode, fmt.Errorf("decode higress console response: %w", err)
	}

	return resp.StatusCode, nil
}
