.PHONY: up down ps logs build migrate api-dev web-dev

up:
	docker compose up -d --build

down:
	docker compose down

ps:
	docker compose ps

logs:
	docker compose logs -f --tail=200

build:
	docker compose build

migrate:
	docker compose run --rm migrate

api-dev:
	cd backend && go run ./cmd/api

web-dev:
	cd frontend && npm run dev

