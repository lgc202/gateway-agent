// Package higress 通过 Higress Console API 读取网关配置
package higress

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/lgc202/gateway-agent/internal/apiserver/config"
	gatewayservice "github.com/lgc202/gateway-agent/internal/apiserver/service/gateway"
)

const requestTimeout = 10 * time.Second

// Client 是 Higress Console 只读客户端
type Client struct {
	endpoint   *url.URL
	username   string
	password   string
	httpClient *http.Client
}

type routeListResponse struct {
	Success  bool           `json:"success"`
	Data     []higressRoute `json:"data"`
	Total    int            `json:"total"`
	PageNum  int            `json:"pageNum"`
	PageSize int            `json:"pageSize"`
}

type routeResponse struct {
	Success bool         `json:"success"`
	Data    higressRoute `json:"data"`
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

// ListRoutes 查询 Higress 路由列表
func (c *Client) ListRoutes(ctx context.Context, query gatewayservice.RouteQuery) (gatewayservice.RoutePage, error) {
	requestURL := c.endpoint.JoinPath("v1", "routes")
	values := requestURL.Query()
	if query.Domain != "" {
		values.Set("domainName", query.Domain)
	}
	if query.PageNumber > 0 {
		values.Set("pageNum", strconv.Itoa(query.PageNumber))
	}
	if query.PageSize > 0 {
		values.Set("pageSize", strconv.Itoa(query.PageSize))
	}
	requestURL.RawQuery = values.Encode()

	var resp routeListResponse
	if err := c.get(ctx, requestURL, &resp); err != nil {
		return gatewayservice.RoutePage{}, err
	}
	if !resp.Success {
		return gatewayservice.RoutePage{}, fmt.Errorf("higress console route query failed")
	}

	return gatewayservice.RoutePage{
		Items:      toRoutes(resp.Data),
		Total:      resp.Total,
		PageNumber: resp.PageNum,
		PageSize:   resp.PageSize,
	}, nil
}

// GetRoute 按名称查询一条 Higress 路由
func (c *Client) GetRoute(ctx context.Context, name string) (gatewayservice.Route, error) {
	requestURL := c.endpoint.JoinPath("v1", "routes", name)

	var resp routeResponse
	if err := c.get(ctx, requestURL, &resp); err != nil {
		return gatewayservice.Route{}, err
	}
	if !resp.Success {
		return gatewayservice.Route{}, fmt.Errorf("higress console route query failed")
	}

	return toRoute(resp.Data), nil
}

func (c *Client) get(ctx context.Context, requestURL *url.URL, result any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL.String(), nil)
	if err != nil {
		return fmt.Errorf("create higress console request: %w", err)
	}
	req.SetBasicAuth(c.username, c.password)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request higress console: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return gatewayservice.ErrRouteNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("higress console returned HTTP status %d", resp.StatusCode)
	}
	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return fmt.Errorf("decode higress console response: %w", err)
	}

	return nil
}
