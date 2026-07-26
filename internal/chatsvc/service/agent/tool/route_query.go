// Package tool 提供 Gateway Agent 可以调用的 Tool
package tool

import (
	"context"
	"errors"
	"fmt"
	"strings"

	einotool "github.com/cloudwego/eino/components/tool"
	toolutils "github.com/cloudwego/eino/components/tool/utils"

	gatewayservice "github.com/lgc202/gateway-agent/internal/chatsvc/service/gateway"
)

const (
	routeQueryToolName        = "query_routes"
	routeQueryToolDescription = "查询当前网关的实时路由配置。可以按精确路由名称查询，也可以按域名筛选。该工具只读，不会修改网关。"
	defaultRoutePageNumber    = 1
	defaultRoutePageSize      = 20
	maxRoutePageSize          = 100
)

// routeQueryInput 是模型调用路由查询 Tool 时提交的参数
type routeQueryInput struct {
	Name       string `json:"name,omitempty" jsonschema_description:"精确的路由名称；已知名称时优先使用"`
	Domain     string `json:"domain,omitempty" jsonschema_description:"路由绑定的完整域名"`
	PageNumber int    `json:"page_number,omitempty" jsonschema_description:"页码，从 1 开始，默认 1"`
	PageSize   int    `json:"page_size,omitempty" jsonschema_description:"每页数量，默认 20，最大 100"`
}

// NewRouteQuery 创建只读路由查询 Tool
func NewRouteQuery(reader gatewayservice.RouteReader) (einotool.BaseTool, error) {
	return toolutils.InferTool(
		routeQueryToolName,
		routeQueryToolDescription,
		func(ctx context.Context, input routeQueryInput) (gatewayservice.RoutePage, error) {
			if input.PageNumber < 0 {
				return gatewayservice.RoutePage{}, fmt.Errorf("page_number must be greater than 0")
			}
			if input.PageSize < 0 || input.PageSize > maxRoutePageSize {
				return gatewayservice.RoutePage{}, fmt.Errorf("page_size must be between 1 and %d", maxRoutePageSize)
			}

			name := strings.TrimSpace(input.Name)
			if name != "" {
				route, err := reader.GetRoute(ctx, name)
				if errors.Is(err, gatewayservice.ErrRouteNotFound) {
					return gatewayservice.RoutePage{Items: []gatewayservice.Route{}}, nil
				}
				if err != nil {
					return gatewayservice.RoutePage{}, err
				}
				return gatewayservice.RoutePage{
					Items: []gatewayservice.Route{route},
					Total: 1,
				}, nil
			}

			pageNumber := input.PageNumber
			if pageNumber == 0 {
				pageNumber = defaultRoutePageNumber
			}
			pageSize := input.PageSize
			if pageSize == 0 {
				pageSize = defaultRoutePageSize
			}

			return reader.ListRoutes(ctx, gatewayservice.RouteQuery{
				Domain:     strings.TrimSpace(input.Domain),
				PageNumber: pageNumber,
				PageSize:   pageSize,
			})
		},
	)
}
