MYSQL_HOST ?= 127.0.0.1
MYSQL_PORT ?= 3306
MYSQL_USER ?= root
MYSQL_PWD ?= abc123
MYSQL_DB ?= sercherai

.PHONY: init-db seed-db
init-db:
	MYSQL_HOST=$(MYSQL_HOST) \
	MYSQL_PORT=$(MYSQL_PORT) \
	MYSQL_USER=$(MYSQL_USER) \
	MYSQL_PWD=$(MYSQL_PWD) \
	MYSQL_DB=$(MYSQL_DB) \
	./scripts/init_mysql.sh

seed-db:
	MYSQL_HOST=$(MYSQL_HOST) \
	MYSQL_PORT=$(MYSQL_PORT) \
	MYSQL_USER=$(MYSQL_USER) \
	MYSQL_PWD=$(MYSQL_PWD) \
	MYSQL_DB=$(MYSQL_DB) \
	./scripts/seed_mysql.sh
