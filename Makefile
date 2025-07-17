
all: run

local:
	go run cmd/bot.go -env local

run:
	docker-compose up -d

build:
	docker compose build

restart:
	docker compose down && docker compose up -d

clean:
