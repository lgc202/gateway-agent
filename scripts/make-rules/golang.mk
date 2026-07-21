GO ?= go
GOFMT ?= gofmt

APISERVER_BIN := $(OUTPUT_DIR)/bin/gateway-agent-apiserver

.PHONY: fmt
fmt: ## 格式化 Go 代码
	@files="$$(find . -name '*.go' -not -path './_output/*')"; if [[ -n "$$files" ]]; then $(GOFMT) -w $$files; fi

.PHONY: vet
vet: ## 运行 go vet
	@$(GO) vet ./...

.PHONY: build
build: build-apiserver ## 构建所有二进制

.PHONY: build-apiserver
build-apiserver: ## 构建 apiserver
	@mkdir -p $(dir $(APISERVER_BIN))
	@$(GO) build -trimpath -o $(APISERVER_BIN) ./cmd/gateway-agent-apiserver
