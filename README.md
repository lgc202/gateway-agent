# Gateway Agent

Gateway Agent 是一个面向企业 AI 网关运维场景的对话式 Agent。它把自然语言需求转换为可审查、可审批、可执行和可回滚的网关操作。一套 Gateway Agent 绑定一套逻辑 Gateway 控制面。

当前仓库已经提供一个可运行的 `gateway-agent-apiserver`，完成后续 Agent 能力依赖的最小对话入口：

- 创建和查询 Chat；
- 追加用户 Message；
- 按消息 ID 游标查询 Message；
- MySQL 事务内写入 Message 并更新 Chat 活跃时间；
- `/healthz` 和 `/readyz` 健康检查；
- 统一的 `code`、`message`、`data` HTTP 响应。

输入规范化、意图识别、上下文、检索、Plan、Replan、审批、Tool、Skill、MCP、Temporal 和 Eino 属于目标架构，但当前没有用空接口或预建表伪造这些能力。

## 技术栈

- Go 1.26.x
- Gin 1.12
- Google Wire 0.7
- Viper 1.20
- MySQL 8.4、database/sql、sqlc
- OpenAPI 3.1

目标架构还会按真实用例逐步接入 Redis、Temporal、Eino、Qdrant、OpenTelemetry 和 Prometheus。

## 本地运行

```bash
make tools
make generate
make docker-up

export PRODUCT_MIGRATION_URL='mysql://gateway_agent:gateway_agent@tcp(127.0.0.1:3306)/gateway_agent?multiStatements=true'
make migrate-up

make build
./_output/bin/gateway-agent-apiserver --config configs/gateway-agent-apiserver.yaml
```

在另一个终端执行真实 Chat API 演示：

```bash
make demo-chat
```

常用质量检查：

```bash
make verify
make verify-migrations
```

## 设计文档

- [00-文档阅读指南](.codex/docs/00-文档阅读指南.md)
- [01-产品需求文档](.codex/docs/01-企业级AI网关运维智能体产品需求文档.md)
- [02-目标产品与架构设计](.codex/docs/02-网关智能体目标产品与架构设计.md)
- [03-聊天与消息基础设计](.codex/docs/03-聊天与消息基础设计.md)
- [04-聊天与消息实施计划](.codex/docs/04-聊天与消息实施计划.md)

历史设计和已废弃实施计划统一保存在 `.codex/docs/archive/`，不得作为新代码的实现依据。

## 产品非目标

- 共享数据库多租户 SaaS；
- 集中管理大量 Gateway 的 Fleet Manager；
- 自研 IAM、通用消息平台和不必要微服务。
