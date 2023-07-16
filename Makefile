postgresql_url = postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable

.PHONY: sleep
sleep:
	sleep 10

.PHONY: migrate_up
migrate_up:
	migrate -database $(postgresql_url) -path data/migrations up

.PHONY: migrate_down
migrate_down:
	migrate -database $(postgresql_url) -path data/migrations down -all

.PHONY: docker_up
docker_up:
	docker compose up postgres -d

.PHONY: docker_down
docker_down:
	docker compose down

.PHONY: up
up: docker_up sleep migrate_up

.PHONY: down
down: migrate_down docker_down

.PHONY: lint
lint:
	golangci-lint run --fix ./...

.PHONY: test
test:
	go test ./...
