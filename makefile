DB_CONTAINER=mysql-db
DB_NAME=mysql-db
DB_USER=root
DB_PASSWORD=123456
DB_COMMANDS_DIR=./db/commands/
CHARSET=--default-character-set=utf8mb4

init_db:
	cat $(DB_COMMANDS_DIR)init.sql | docker exec -i $(DB_CONTAINER) mysql -u$(DB_USER) -p$(DB_PASSWORD) $(DB_NAME) $(CHARSET)

q:
	cat $(DB_COMMANDS_DIR)q.sql | docker exec -i $(DB_CONTAINER) mysql -u$(DB_USER) -p$(DB_PASSWORD) $(DB_NAME) -t $(CHARSET)

import_data:
	cat $(DB_COMMANDS_DIR)import-data.sql | docker exec -i $(DB_CONTAINER) mysql -u$(DB_USER) -p$(DB_PASSWORD) $(DB_NAME) $(CHARSET)


.PHONY: init_db q import_data