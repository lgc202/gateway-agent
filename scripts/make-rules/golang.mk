GO ?= go
GOFMT ?= gofmt

CHAT_SVC_BIN := $(OUTPUT_DIR)/bin/chat-svc

.PHONY: fmt
fmt: ## 格式化 Go 代码
	@files="$$(find . -name '*.go' -not -path './_output/*')"; if [[ -n "$$files" ]]; then $(GOFMT) -w $$files; fi

.PHONY: vet
vet: ## 运行 go vet
	@$(GO) vet ./...

.PHONY: build
build: build-chat-svc ## 构建所有二进制

.PHONY: build-chat-svc
build-chat-svc: ## 构建 Chat Service
	@mkdir -p $(dir $(CHAT_SVC_BIN))
	@$(GO) build -trimpath -o $(CHAT_SVC_BIN) ./cmd/chat-svc
