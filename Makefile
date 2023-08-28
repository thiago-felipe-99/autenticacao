POSTGRESQL_URL = postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable

.PHONY: verify_postgresql
verify_postgresql:
	docker exec -it autenticacao-postgres-1 sh /check_postgresql.sh

.PHONY: migrate_up
migrate_up:
	migrate -database $(POSTGRESQL_URL) -path data/migrations up

.PHONY: migrate_down
migrate_down:
	migrate -database $(POSTGRESQL_URL) -path data/migrations down -all

.PHONY: docker_up
docker_up:
	docker compose up -d

.PHONY: docker_down
docker_down:
	docker compose down

.PHONY: up
up: docker_up verify_postgresql migrate_up lint
	swag init
	go run ./...

.PHONY: down
down: migrate_down docker_down

.PHONY: down_clean 
down_clean: down
	docker compose down -v
	rm -rf coverage.out

.PHONY: lint
lint:
	golangci-lint run --fix ./...
	swag fmt .

.PHONY: test
test: docker_up verify_postgresql lint
	go test ./...

.PHONY: coverage
coverage: docker_up verify_postgresql migrate_up lint
	go test -coverprofile coverage.out ./...
	go tool cover -html=coverage.out

.PHONY: clean
clean: down_clean

.PHONY: all
all: up
