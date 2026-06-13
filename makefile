DB_CONTAINER=mysql-db
DB_NAME=mysql-db
DB_USER=root
DB_PASSWORD=123456
DB_COMMANDS_DIR=./db/commands/

init_db:
	cat $(DB_COMMANDS_DIR)init.sql | docker exec -i $(DB_CONTAINER) mysql -u$(DB_USER) -p$(DB_PASSWORD) $(DB_NAME)

q:
	cat $(DB_COMMANDS_DIR)q.sql | docker exec -i $(DB_CONTAINER) mysql -u$(DB_USER) -p$(DB_PASSWORD) $(DB_NAME) -t

import_data:
	cat $(DB_COMMANDS_DIR)import-data.sql | docker exec -i $(DB_CONTAINER) mysql -u$(DB_USER) -p$(DB_PASSWORD) $(DB_NAME)


.PHONY: init_db q import_data