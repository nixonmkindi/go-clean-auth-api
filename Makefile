APP_NAME=go-clean-auth-api

.PHONY: run test migrate migrate-down swagger docker-up docker-down

run:
	go run ./cmd/api/main.go

test:
	go test ./... -cover

migrate:
	go run ./cmd/migrate/main.go up

migrate-down:
	go run ./cmd/migrate/main.go down

swagger:
	@if command -v swag >/dev/null 2>&1; then \
		swag init -g cmd/api/main.go -o docs/swagger; \
	else \
		echo "swag CLI not installed. OpenAPI spec is available at docs/openapi.yaml"; \
	fi

docker-up:
	docker compose up --build -d

docker-down:
	docker compose down
