package higress

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	gatewayservice "github.com/lgc202/gateway-agent/internal/chatsvc/service/gateway"
)

// higressRoute 是 Higress Console API 使用的路由结构
type higressRoute struct {
	Name     string                 `json:"name"`
	Version  string                 `json:"version,omitempty"`
	Domains  []string               `json:"domains,omitempty"`
	Path     *higressRoutePredicate `json:"path,omitempty"`
	Methods  []string               `json:"methods,omitempty"`
	Services []higressUpstream      `json:"services"`
}

// higressRoutePredicate 是 Higress Console API 使用的路由匹配条件
type higressRoutePredicate struct {
	MatchType     string `json:"matchType"`
	MatchValue    string `json:"matchValue"`
	CaseSensitive *bool  `json:"caseSensitive,omitempty"`
}

// higressUpstream 是 Higress Console API 使用的路由后端
type higressUpstream struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
	Port    *int   `json:"port,omitempty"`
	Weight  *int   `json:"weight,omitempty"`
}

// routeListResponse 是 Higress Console 的路由列表响应
type routeListResponse struct {
	Success  bool           `json:"success"`
	Data     []higressRoute `json:"data"`
	Total    int            `json:"total"`
	PageNum  int            `json:"pageNum"`
	PageSize int            `json:"pageSize"`
}

// routeResponse 是 Higress Console 的单条路由响应
type routeResponse struct {
	Success bool         `json:"success"`
	Data    higressRoute `json:"data"`
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
	if _, err := c.request(ctx, http.MethodGet, requestURL, nil, &resp); err != nil {
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
	statusCode, err := c.request(ctx, http.MethodGet, requestURL, nil, &resp)
	if statusCode == http.StatusNotFound {
		return gatewayservice.Route{}, gatewayservice.ErrRouteNotFound
	}
	if err != nil {
		return gatewayservice.Route{}, err
	}
	if !resp.Success {
		return gatewayservice.Route{}, fmt.Errorf("higress console route query failed")
	}

	return toRoute(resp.Data), nil
}

// CreateRoute 在 Higress 中创建路由并返回服务端保存的配置
func (c *Client) CreateRoute(ctx context.Context, route gatewayservice.Route) (gatewayservice.Route, error) {
	requestURL := c.endpoint.JoinPath("v1", "routes")

	var resp routeResponse
	if _, err := c.request(ctx, http.MethodPost, requestURL, toHigressRoute(route), &resp); err != nil {
		return gatewayservice.Route{}, err
	}
	if !resp.Success {
		return gatewayservice.Route{}, fmt.Errorf("higress console route creation failed")
	}

	return toRoute(resp.Data), nil
}

// toRoutes 将 Higress 路由列表转换为 Agent 使用的统一结构
func toRoutes(values []higressRoute) []gatewayservice.Route {
	routes := make([]gatewayservice.Route, 0, len(values))
	for _, value := range values {
		routes = append(routes, toRoute(value))
	}
	return routes
}

// toRoute 将一条 Higress 路由转换为 Agent 使用的统一结构
func toRoute(value higressRoute) gatewayservice.Route {
	var path *gatewayservice.RoutePredicate
	if value.Path != nil {
		path = &gatewayservice.RoutePredicate{
			Type:          value.Path.MatchType,
			Value:         value.Path.MatchValue,
			CaseSensitive: value.Path.CaseSensitive,
		}
	}

	backends := make([]gatewayservice.Backend, 0, len(value.Services))
	for _, service := range value.Services {
		backends = append(backends, gatewayservice.Backend{
			Name:    service.Name,
			Version: service.Version,
			Port:    service.Port,
			Weight:  service.Weight,
		})
	}

	return gatewayservice.Route{
		Name:     value.Name,
		Version:  value.Version,
		Domains:  value.Domains,
		Path:     path,
		Methods:  value.Methods,
		Backends: backends,
	}
}

// toHigressRoute 将 Agent 的统一路由结构转换为 Higress 请求结构
func toHigressRoute(value gatewayservice.Route) higressRoute {
	var path *higressRoutePredicate
	if value.Path != nil {
		path = &higressRoutePredicate{
			MatchType:     value.Path.Type,
			MatchValue:    value.Path.Value,
			CaseSensitive: value.Path.CaseSensitive,
		}
	}

	services := make([]higressUpstream, 0, len(value.Backends))
	for _, backend := range value.Backends {
		services = append(services, higressUpstream{
			Name:    backend.Name,
			Version: backend.Version,
			Port:    backend.Port,
			Weight:  backend.Weight,
		})
	}

	return higressRoute{
		Name:     value.Name,
		Version:  value.Version,
		Domains:  value.Domains,
		Path:     path,
		Methods:  value.Methods,
		Services: services,
	}
}
