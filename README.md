# Gateway Agent

Gateway Agent 是一个面向 AI 网关运维场景的对话式 Agent。一套 Gateway Agent 绑定一套逻辑 Gateway，用户可以用自然语言查询网关配置，并在对话中审批需要修改网关的操作。

当前仓库已经实现一条可运行的 Higress 路由垂直链路：

- 创建 Chat，并为 Chat 选择系统默认模型或用户保存的模型配置；
- 保存用户消息和完整的 Assistant 回复；
- 通过 SSE 实时返回模型文本、完成消息和待审批操作；
- 使用 Eino ADK 驱动模型与 Tool 的 ReAct 循环；
- 查询 Higress 路由；
- 创建路由前暂停 Agent，将完整 Tool 参数和 Eino Checkpoint 保存到 MySQL；
- 用户批准或拒绝后，从 Checkpoint 恢复原来的 Agent 调用；
- 管理 OpenAI、DeepSeek、Qwen、Claude、Gemini、Ollama 和 Ark 模型配置；
- 使用 AES-256-GCM 加密保存模型 API Key。

当前没有引入 Redis、Temporal、Qdrant、Task、Run、Outbox 或独立 Worker。输入规范化、意图识别、Plan/Replan、上下文压缩、智能检索、Skill、MCP、认证授权和可观测性仍是待设计能力，不应被理解为已经实现。

## 当前架构

```mermaid
flowchart LR
    UI["Web / API Client"] -->|"JSON / SSE"| CHAT["chat-svc"]
    CHAT --> MYSQL["MySQL 8.4"]
    CHAT --> EINO["Eino ADK Agent"]
    EINO --> MODEL["模型 Provider"]
    EINO --> QUERY["query_routes Tool"]
    EINO --> CREATE["create_route Tool"]
    QUERY --> HIGRESS["Higress Console API"]
    CREATE -->|"用户批准后"| HIGRESS
```

`chat-svc` 是当前唯一进程。MySQL 保存 Chat、Message、模型配置和审批恢复状态；文本分片只通过当前 SSE 连接传输，Assistant 完整消息完成后才写入 MySQL。

## 技术栈

- Go 1.26.x
- Gin 1.12
- Eino ADK
- Google Wire 0.7
- Viper 1.20
- MySQL 8.4、`database/sql`、sqlc
- OpenAPI 3.1
- Higress Console API

## 本地运行

启动 MySQL：

```bash
make tools
make docker-up

export PRODUCT_MIGRATION_URL='mysql://gateway_agent:gateway_agent@tcp(127.0.0.1:3306)/gateway_agent?multiStatements=true'
make migrate-up
```

启动 Chat Service 前至少需要提供模型配置加密主密钥和 Higress 凭证：

```bash
export MODEL_CONFIG_ENCRYPTION_KEY='<32 字节密钥的 Base64 编码>'
export HIGRESS_PASSWORD='<Higress Console 密码>'

# 使用配置文件中的默认模型时还需设置对应信息
export MODEL_API_KEY='<模型 API Key>'
export MODEL_NAME='<模型名称>'

make build
./_output/bin/chat-svc --config configs/chat-svc.yaml
```

也可以先通过 `/api/v1/model-configs` 保存模型配置，再在创建 Chat 时传入 `model_config_id`。完整 HTTP 契约见 [OpenAPI](api/openapi/gateway-agent.v1.yaml)。

## 文档

建议按以下顺序阅读：

1. [00-文档阅读指南](.codex/docs/00-文档阅读指南.md)
2. [01-产品需求文档](.codex/docs/01-企业级AI网关运维智能体产品需求文档.md)
3. [02-系统架构设计](.codex/docs/02-系统架构设计.md)
4. [03-接口与数据模型](.codex/docs/03-接口与数据模型.md)
5. [04-Agent运行与工具设计](.codex/docs/04-Agent运行与工具设计.md)
6. [05-路由审批与恢复设计](.codex/docs/05-路由审批与恢复设计.md)
7. [06-未实现能力与实现原则](.codex/docs/06-未实现能力与实现原则.md)

## 产品边界

- 一套 Gateway Agent 对应一套逻辑 Gateway，不做共享数据库多租户 SaaS；
- 当前只支持 Higress，不把未来 Adapter 写成已完成能力；
- 当前只实现路由查询和创建，不承诺通用网关运维能力；
- 未经用户明确批准，写 Tool 不得调用 Higress；
- 审批状态表示用户决定，不代表网关执行成功；执行结果通过恢复后的 SSE 返回。
