// Package gateway 定义 Chat Service 使用的网关能力和稳定领域对象
package gateway

import (
	"context"
	"errors"
)

// ErrRouteNotFound 表示网关中不存在指定路由
var ErrRouteNotFound = errors.New("gateway route not found")

// RouteReader 提供 Agent 当前需要的只读路由能力
type RouteReader interface {
	ListRoutes(context.Context, RouteQuery) (RoutePage, error)
	GetRoute(context.Context, string) (Route, error)
}

// RouteWriter 提供 Agent 当前需要的路由变更能力
type RouteWriter interface {
	CreateRoute(context.Context, Route) (Route, error)
}

// RouteQuery 是路由列表查询条件
type RouteQuery struct {
	Domain     string `json:"domain,omitempty"`
	PageNumber int    `json:"page_number,omitempty"`
	PageSize   int    `json:"page_size,omitempty"`
}

// RoutePage 是一页路由查询结果
type RoutePage struct {
	Items      []Route `json:"items"`
	Total      int     `json:"total"`
	PageNumber int     `json:"page_number,omitempty"`
	PageSize   int     `json:"page_size,omitempty"`
}

// Route 是不同网关适配器向 Agent 提供的统一路由信息
type Route struct {
	Name     string          `json:"name"`
	Version  string          `json:"version,omitempty"`
	Domains  []string        `json:"domains,omitempty"`
	Path     *RoutePredicate `json:"path,omitempty"`
	Methods  []string        `json:"methods,omitempty"`
	Backends []Backend       `json:"backends,omitempty"`
}

// RoutePredicate 描述路由的路径匹配规则
type RoutePredicate struct {
	Type          string `json:"type"`
	Value         string `json:"value"`
	CaseSensitive *bool  `json:"case_sensitive,omitempty"`
}

// Backend 描述路由指向的一个后端服务
type Backend struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
	Port    *int   `json:"port,omitempty"`
	Weight  *int   `json:"weight,omitempty"`
}
