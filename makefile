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
# docker
docker_down:
	docker compose down 
docker_up:
	docker compose up -d
# docker-mysql
q:
	cat $(DB_DIR)/commands/q.sql | docker exec -i $(DB_CONTAINER) mysql -u$(DB_USER) -p$(DB_PASSWORD) $(DB_NAME) -t $(CHARSET)

# sqlc
sqlc_gen:
	sqlc generate
# mock
mock:
	mockgen -destination=$(DB_DIR)/mock/store.go -package=mock_db $(PROJECT_INTERNAL_PATH)/db/sqlc Store && \
	mockgen -destination=$(DAILY_DATA_SERVICE_DIR)/mock/service.go -package=mock_service $(DAILY_DATA_SERVICE_PATH) Store && \
	mockgen -destination=$(DAILY_DATA_CONTROLLER_DIR)/mock/controller.go -package=mock_controller $(DAILY_DATA_CONTROLLER_PATH) Service && \
	mockgen -destination=$(USER_SERVICE_DIR)/mock/service.go -package=mock_service $(USER_SERVICE_PATH) Store && \
	mockgen -destination=$(USER_CONTROLLER_DIR)/mock/controller.go -package=mock_controller $(USER_CONTROLLER_PATH) Service
# test
test:
	go test -cover -race -count=1 $(DAILY_DATA_SERVICE_DIR)
	go test -cover -race -count=1 $(DAILY_DATA_CONTROLLER_DIR)
	go test -cover -race -count=1 $(DB_SQLC_DIR)


.PHONY: migrate_create migrate_up migrate_up_1 migrate_down migrate_down_1
.PHONY: docker_down docker_up q
.PHONY: sqlc_gen mock test