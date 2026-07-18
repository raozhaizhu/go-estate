# 引入全局变量
include app.env
export $(shell sed 's/=.*//' app.env)

# migrate
migrate_create:
	migrate create -ext sql -dir $(DB_DIR)/migration -seq $(name)
migrate_up:
	migrate -path $(DB_DIR)/migration -database "$(DB_URL)" -verbose up
migrate_up_1:
	migrate -path $(DB_DIR)/migration -database "$(DB_URL)" -verbose up 1
migrate_down:
	migrate -path $(DB_DIR)/migration -database "$(DB_URL)" -verbose down
migrate_down_1:
	migrate -path $(DB_DIR)/migration -database "$(DB_URL)" -verbose down 1
migrate_version:
	migrate -path $(DB_DIR)/migration -database "$(DB_URL)" version
migrate_clean_force:
	migrate -path $(DB_DIR)/migration -database "$(DB_URL)" force $(name)
# docker
docker_down:
	docker compose down 
docker_up:
	docker compose up -d
docker_rebuild:
	docker-compose down -v
	docker compose up -d --build
	sleep 10
	$(MAKE) migrate_up	

# docker-mysql
q:
	cat cmd/db_cmd/q.sql | docker exec -i $(DB_CONTAINER) mysql -u$(DB_USER) -p$(DB_PASSWORD) $(DB_NAME) -t $(CHARSET)

# sqlc
sqlc_gen:
	sqlc generate
# mock
mock:
	mockgen -destination=$(DB_DIR)/mock/store.go -package=mock_db $(PROJECT_INTERNAL_PATH)/dao/sqlc Store && \
	mockgen -destination=$(DB_DIR)/mock/cache.go -package=mock_db $(PROJECT_INTERNAL_PATH)/dao/cache SessionCache && \
	mockgen -destination=$(DAILY_DATA_CONTROLLER_DIR)/mock/controller.go -package=mock_controller $(DAILY_DATA_CONTROLLER_PATH) Service && \
	mockgen -destination=$(USER_CONTROLLER_DIR)/mock/controller.go -package=mock_controller $(USER_CONTROLLER_PATH) Service && \
	mockgen -destination=$(AUTH_CONTROLLER_DIR)/mock/controller.go -package=mock_controller $(AUTH_CONTROLLER_PATH) Service && \
	mockgen -source=internal/worker/types.go -destination=internal/worker/mock/task_distributor.go -package=mock_worker && \
	mockgen -destination=$(PKG_DIR)/token/mock/maker.go -package=mock_token $(MAKER_PATH) Maker
# test
test:
	gotestsum --format dots --no-summary=output

swag:
	swag init -g cmd/api/main.go --parseDependency --parseInternal

.PHONY: migrate_create migrate_up migrate_up_1 migrate_down migrate_down_1
.PHONY: docker_down docker_up q docker_rebuild
.PHONY: sqlc_gen mock test swag