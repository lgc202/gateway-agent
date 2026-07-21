.PHONY: verify-generated
verify-generated: tools ## 验证生成代码无漂移
	@before="$$(find internal -path '*/sqlc/*.go' -o -name 'wire_gen.go' 2>/dev/null | sort | xargs shasum 2>/dev/null || true)"; \
	$(MAKE) generate >/dev/null; \
	after="$$(find internal -path '*/sqlc/*.go' -o -name 'wire_gen.go' 2>/dev/null | sort | xargs shasum 2>/dev/null || true)"; \
	[[ "$$before" == "$$after" ]] || { echo 'generated files are out of date'; exit 1; }

.PHONY: verify-openapi
verify-openapi: ## 验证 OpenAPI 文档存在
	@test -f api/openapi/gateway-agent.v1.yaml

.PHONY: verify-migrations
verify-migrations: tools ## 验证数据库 Migration
	@test -n "$${PRODUCT_MIGRATION_URL:-}" || { echo 'PRODUCT_MIGRATION_URL is required'; exit 1; }
	@$(MIGRATE) -path db/migrations -database "$$PRODUCT_MIGRATION_URL" up
	@$(MIGRATE) -path db/migrations -database "$$PRODUCT_MIGRATION_URL" down -all
	@$(MIGRATE) -path db/migrations -database "$$PRODUCT_MIGRATION_URL" up

.PHONY: verify
verify: fmt vet verify-generated verify-openapi build ## 运行当前阶段质量检查
	@go test ./...
