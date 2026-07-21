COMPOSE_FILE := deploy/docker-compose.yaml

.PHONY: docker-up
docker-up: ## 启动本地 MySQL
	@docker compose -f $(COMPOSE_FILE) up -d

.PHONY: docker-down
docker-down: ## 停止本地 MySQL
	@docker compose -f $(COMPOSE_FILE) down

.PHONY: migrate-up
migrate-up: tools ## 执行数据库 Migration
	@test -n "$${PRODUCT_MIGRATION_URL:-}" || { echo 'PRODUCT_MIGRATION_URL is required'; exit 1; }
	@$(MIGRATE) -path db/migrations -database "$$PRODUCT_MIGRATION_URL" up

.PHONY: migrate-down
migrate-down: tools ## 回滚全部数据库 Migration
	@test -n "$${PRODUCT_MIGRATION_URL:-}" || { echo 'PRODUCT_MIGRATION_URL is required'; exit 1; }
	@$(MIGRATE) -path db/migrations -database "$$PRODUCT_MIGRATION_URL" down -all

.PHONY: demo-chat
demo-chat: ## 运行 Chat API 真实演示
	@./scripts/demo-chat.sh
