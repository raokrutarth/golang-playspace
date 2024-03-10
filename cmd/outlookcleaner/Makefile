.PHONY: clean critic security lint test build run

# app env vars
APP_NAME = irl-archive-cli
BUILD_DIR = $(PWD)/build

# db env vars
SYS_DB_URI := postgres://postgres:ep16N8oXEe4DymjR@localhost:32100/postgres?sslmode=disable
DEV_DB_URI := postgres://dev:Im9JpvW1SCsSE4G5@localhost:32100/irl_archive_dev?sslmode=disable
PROD_DB_URI := postgres://app:KOZ7dEl8yhpwiYon@localhost:32100/irl_archive?sslmode=disable

TIME_NOW := $(shell date +%Hh-%b-%d-%Y)

debug:
	-dlv debug cmd/cli/main.go prune
	rm /home/zee/sharp/newsSummarizer/archive/__debug_bin

serve-data-dir:
	# pwd: MLFIRobf8k68N212
	# docker run --rm -it -v $(PWD)/data:/srv:ro -v $(PWD)/iac/Caddyfile:/Caddyfile:ro \
	# 	-p 443:443 caddy caddy run --config /Caddyfile --watch

	# https://github.com/sigoden/dufs
	docker run --rm -it -v $(PWD)/data-copy:/data -p 3289:5000 sigoden/dufs /data -A

	# https://github.com/mufeedvh/binserve
	# https://github.com/mjpclab/go-http-file-server

	# https://github.com/syntaqx/serve
	# docker run --rm -it -v $(PWD)/data:/public -p 3289:8080 syntaqx/serve serve --dir /public

	# https://github.com/janiltonmaciel/statiks
	# docker run --rm -it -v $(PWD)/data-copy:/public -p 127.0.0.1:3289:9080 janilton/statiks /public

run:
	-rm -f data/*
	go run cmd/cli/main.go prune

auth-cli:
	go run /workspaces/golang-playspace/cmd/outlookcleaner/main.go

db-dev-in:
	psql $(DEV_DB_URI)

db-dev-flush:
	psql $(SYS_DB_URI) -c 'DROP DATABASE irl_archive_dev'
	psql $(SYS_DB_URI) -c 'CREATE DATABASE irl_archive_dev'


db-dev-mock: flush-db
	psql $(DEV_DB_URI) -f /home/dev/ws/backend/scripts/mock_1.sql
	make -s show-db

db-dev-show:
	psql $(DEV_DB_URI) -c "select * from email_accounts;"


