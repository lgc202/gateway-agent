package tool

import (
	"context"
	"fmt"
	"strings"

	einotool "github.com/cloudwego/eino/components/tool"
	toolutils "github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/schema"

	gatewayservice "github.com/lgc202/gateway-agent/internal/chatsvc/service/gateway"
)

const (
	routeCreateToolName        = "create_route"
	routeCreateToolDescription = "创建一条网关路由。该操作会修改网关，必须先向用户展示完整参数并获得审批。"
	routePathTypePrefix        = "PRE"
	routePathTypeExact         = "EQUAL"
	routePathTypeRegular       = "REGULAR"
	maxBackendPort             = 65535
	requiredBackendWeight      = 100
)

// routeCreateInput 是模型调用创建路由 Tool 时提交的参数
type routeCreateInput struct {
	Name     string                    `json:"name" jsonschema:"required" jsonschema_description:"路由名称"`
	Domains  []string                  `json:"domains,omitempty" jsonschema_description:"路由绑定的域名；不填写表示匹配全部域名"`
	Path     routeCreatePathInput      `json:"path" jsonschema:"required" jsonschema_description:"路径匹配条件"`
	Methods  []string                  `json:"methods,omitempty" jsonschema_description:"允许的 HTTP 方法；不填写表示匹配全部方法"`
	Backends []routeCreateBackendInput `json:"backends" jsonschema:"required" jsonschema_description:"路由转发到的后端服务"`
}

// routeCreatePathInput 是创建路由时的路径匹配条件
type routeCreatePathInput struct {
	Type          string `json:"type" jsonschema:"required,enum=PRE,enum=EQUAL,enum=REGULAR" jsonschema_description:"匹配类型：PRE 前缀、EQUAL 精确、REGULAR 正则"`
	Value         string `json:"value" jsonschema:"required" jsonschema_description:"需要匹配的路径"`
	CaseSensitive bool   `json:"case_sensitive,omitempty" jsonschema_description:"是否区分路径大小写"`
}

// routeCreateBackendInput 是创建路由时的后端服务
type routeCreateBackendInput struct {
	Name   string `json:"name" jsonschema:"required" jsonschema_description:"Higress 服务来源中的服务名称"`
	Port   int    `json:"port" jsonschema:"required" jsonschema_description:"后端服务端口"`
	Weight int    `json:"weight" jsonschema:"required" jsonschema_description:"流量权重；多个后端的权重总和应为 100"`
}

// Approval 是写操作 Tool 发给审批层的信息
type Approval struct {
	Operation string `json:"operation"`
	Arguments any    `json:"arguments"`
}

// ApprovalDecision 是审批恢复 Tool 时携带的用户决定
type ApprovalDecision struct {
	Approved bool   `json:"approved"`
	Reason   string `json:"reason,omitempty"`
}

// init 注册 Eino Checkpoint 中通过 interface 保存的具体类型
func init() {
	schema.Register[routeCreateInput]()
	schema.Register[*Approval]()
	schema.Register[*ApprovalDecision]()
}

// NewRouteCreate 创建一个必须在用户批准后才会执行的路由写入 Tool
func NewRouteCreate(writer gatewayservice.RouteWriter) (einotool.BaseTool, error) {
	return toolutils.InferTool(
		routeCreateToolName,
		routeCreateToolDescription,
		func(ctx context.Context, input routeCreateInput) (string, error) {
			wasInterrupted, _, storedInput := einotool.GetInterruptState[routeCreateInput](ctx)
			if !wasInterrupted {
				if err := input.normalize(); err != nil {
					return "", err
				}
				return "", interruptRouteCreate(ctx, input)
			}

			isResumeTarget, hasDecision, decision := einotool.GetResumeContext[*ApprovalDecision](ctx)
			if !isResumeTarget {
				return "", interruptRouteCreate(ctx, storedInput)
			}
			if !hasDecision {
				return "", fmt.Errorf("%s resumed without approval decision", routeCreateToolName)
			}
			if !decision.Approved {
				if decision.Reason != "" {
					return fmt.Sprintf("route creation rejected: %s", decision.Reason), nil
				}
				return "route creation rejected", nil
			}

			createdRoute, err := writer.CreateRoute(ctx, storedInput.toRoute())
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("route %q created successfully", createdRoute.Name), nil
		},
	)
}

// normalize 规范化并校验准备展示给用户审批的路由参数。
func (input *routeCreateInput) normalize() error {
	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" {
		return fmt.Errorf("route name is required")
	}

	for i := range input.Domains {
		input.Domains[i] = strings.TrimSpace(input.Domains[i])
		if input.Domains[i] == "" {
			return fmt.Errorf("route domain cannot be empty")
		}
	}

	input.Path.Type = strings.ToUpper(strings.TrimSpace(input.Path.Type))
	switch input.Path.Type {
	case routePathTypePrefix, routePathTypeExact, routePathTypeRegular:
	default:
		return fmt.Errorf("unsupported route path type %q", input.Path.Type)
	}
	input.Path.Value = strings.TrimSpace(input.Path.Value)
	if input.Path.Value == "" {
		return fmt.Errorf("route path value is required")
	}

	for i := range input.Methods {
		input.Methods[i] = strings.ToUpper(strings.TrimSpace(input.Methods[i]))
		if input.Methods[i] == "" {
			return fmt.Errorf("route method cannot be empty")
		}
	}

	if len(input.Backends) == 0 {
		return fmt.Errorf("at least one route backend is required")
	}
	var totalWeight int
	for i := range input.Backends {
		backend := &input.Backends[i]
		backend.Name = strings.TrimSpace(backend.Name)
		if backend.Name == "" {
			return fmt.Errorf("route backend name is required")
		}
		if backend.Port < 1 || backend.Port > maxBackendPort {
			return fmt.Errorf("route backend port must be between 1 and %d", maxBackendPort)
		}
		if backend.Weight < 1 {
			return fmt.Errorf("route backend weight must be greater than 0")
		}
		totalWeight += backend.Weight
	}
	if totalWeight != requiredBackendWeight {
		return fmt.Errorf("route backend weights must total %d", requiredBackendWeight)
	}

	return nil
}

// interruptRouteCreate 保存原始参数，并通知 Eino 暂停本次 Agent 执行
func interruptRouteCreate(ctx context.Context, input routeCreateInput) error {
	return einotool.StatefulInterrupt(ctx, &Approval{
		Operation: routeCreateToolName,
		Arguments: input,
	}, input)
}

// toRoute 将模型参数转换为网关适配器使用的统一路由结构
func (input routeCreateInput) toRoute() gatewayservice.Route {
	caseSensitive := input.Path.CaseSensitive
	backends := make([]gatewayservice.Backend, 0, len(input.Backends))
	for _, backend := range input.Backends {
		port := backend.Port
		weight := backend.Weight
		backends = append(backends, gatewayservice.Backend{
			Name:   backend.Name,
			Port:   &port,
			Weight: &weight,
		})
	}

	return gatewayservice.Route{
		Name:    input.Name,
		Domains: input.Domains,
		Path: &gatewayservice.RoutePredicate{
			Type:          input.Path.Type,
			Value:         input.Path.Value,
			CaseSensitive: &caseSensitive,
		},
		Methods:  input.Methods,
		Backends: backends,
	}
}
