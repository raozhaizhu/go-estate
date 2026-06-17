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
	mockgen -destination=$(DB_DIR)/mock/store.go -package=mock_db $(PROJECT_PATH)/db/sqlc Store && \
	mockgen -destination=$(SERVICE_DIR)/mock/service.go -package=mock_service $(PROJECT_PATH)/service DailyDataServiceInterface

.PHONY: migrate_create migrate_up migrate_up_1 migrate_down migrate_down_1 docker_down docker_up q sqlc_gen