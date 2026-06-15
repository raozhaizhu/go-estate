# 变量
DB_CONTAINER=mysql_db
DB_NAME=mysql_db
DB_USER=root
DB_PASSWORD=123456
DB_COMMANDS_DIR=./db/commands/
CHARSET=--default-character-set=utf8mb4
DB_URL=mysql://root:123456@tcp(127.0.0.1:3306)/mysql_db
MIG_DIR=./db/migration

# migrate
migrate_create:
	migrate create -ext sql -dir $(MIG_DIR) -seq $(name)
migrate_up:
	migrate -path $(MIG_DIR) -database "$(DB_URL)" -verbose up
migrate_up_1:
	migrate -path $(MIG_DIR) -database "$(DB_URL)" -verbose up 1
migrate_down:
	migrate -path $(MIG_DIR) -database "$(DB_URL)" -verbose down
migrate_down_1:
	migrate -path $(MIG_DIR) -database "$(DB_URL)" -verbose down 1
# docker
docker_down:
	docker compose down 
docker_up:
	docker compose up -d
q:
	cat $(DB_COMMANDS_DIR)q.sql | docker exec -i $(DB_CONTAINER) mysql -u$(DB_USER) -p$(DB_PASSWORD) $(DB_NAME) -t $(CHARSET)
# init_db:
# 	cat $(DB_COMMANDS_DIR)init.sql | docker exec -i $(DB_CONTAINER) mysql -u$(DB_USER) -p$(DB_PASSWORD) $(DB_NAME) $(CHARSET)
# import_data:
# 	cat $(DB_COMMANDS_DIR)import-data.sql | docker exec -i $(DB_CONTAINER) mysql -u$(DB_USER) -p$(DB_PASSWORD) $(DB_NAME) $(CHARSET)

# sqlc
sqlc_gen:
	sqlc generate

.PHONY: init_db q import_data