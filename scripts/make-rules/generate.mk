WIRE_VERSION := v0.7.0
SQLC_VERSION := v1.29.0
MIGRATE_VERSION := v4.18.2
GO_TOOLCHAIN := go1.26.0

WIRE := $(TOOLS_DIR)/wire
SQLC := $(TOOLS_DIR)/sqlc
MIGRATE := $(TOOLS_DIR)/migrate

.PHONY: tools
tools: $(WIRE) $(SQLC) $(MIGRATE) ## 安装项目开发工具

$(WIRE):
	@mkdir -p $(TOOLS_DIR)
	@GOTOOLCHAIN=$(GO_TOOLCHAIN) GOBIN=$(TOOLS_DIR) go install github.com/google/wire/cmd/wire@$(WIRE_VERSION)

$(SQLC):
	@mkdir -p $(TOOLS_DIR)
	@GOTOOLCHAIN=$(GO_TOOLCHAIN) GOBIN=$(TOOLS_DIR) go install github.com/sqlc-dev/sqlc/cmd/sqlc@$(SQLC_VERSION)

$(MIGRATE):
	@mkdir -p $(TOOLS_DIR)
	@GOTOOLCHAIN=$(GO_TOOLCHAIN) GOBIN=$(TOOLS_DIR) go install -tags 'mysql' github.com/golang-migrate/migrate/v4/cmd/migrate@$(MIGRATE_VERSION)

.PHONY: generate
generate: tools ## 生成 sqlc 和 Wire 代码
	@$(SQLC) generate -f db/sqlc.yaml
	@$(WIRE) ./internal/apiserver
