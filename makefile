# 引入全局变量
include app.env
export $(shell sed 's/=.*//' app.env)

# migrate
migrate_create:
	migrate create -ext sql -dir $(DB_MIG_DIR) -seq $(name)
migrate_up:
	migrate -path $(DB_MIG_DIR) -database "$(DB_URL)" -verbose up
migrate_up_1:
	migrate -path $(DB_MIG_DIR) -database "$(DB_URL)" -verbose up 1
migrate_down:
	migrate -path $(DB_MIG_DIR) -database "$(DB_URL)" -verbose down
migrate_down_1:
	migrate -path $(DB_MIG_DIR) -database "$(DB_URL)" -verbose down 1
# docker
docker_down:
	docker compose down 
docker_up:
	docker compose up -d
# docker-mysql
q:
	cat $(DB_COMMANDS_DIR)q.sql | docker exec -i $(DB_CONTAINER) mysql -u$(DB_USER) -p$(DB_PASSWORD) $(DB_NAME) -t $(CHARSET)
# init_db:
# 	cat $(DB_COMMANDS_DIR)init.sql | docker exec -i $(DB_CONTAINER) mysql -u$(DB_USER) -p$(DB_PASSWORD) $(DB_NAME) $(CHARSET)
# import_data:
# 	cat $(DB_COMMANDS_DIR)import-data.sql | docker exec -i $(DB_CONTAINER) mysql -u$(DB_USER) -p$(DB_PASSWORD) $(DB_NAME) $(CHARSET)

# sqlc
sqlc_gen:
	sqlc generate
# mock
mock:
	mockgen -destination=$(DB_MOCK_DIR)/store.go -package=mock_db $(PROJECT_PATH)/db/sqlc Store

.PHONY: migrate_create migrate_up migrate_up_1 migrate_down migrate_down_1 docker_down docker_up q init_db import_data sqlc_gen