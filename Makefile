.PHONY: fmt generate test vet run compose-up compose-down migrate-up migrate-down

fmt:
	go fmt ./...

generate:
	go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.30.0 generate

test:
	go test ./...

vet:
	go vet ./...

run:
	go run ./cmd/app

compose-up:
	docker compose -f deployments/compose.yaml up --build

compose-down:
	docker compose -f deployments/compose.yaml down

migrate-up:
	docker compose -f deployments/compose.yaml run --rm migrate

migrate-down:
	docker compose -f deployments/compose.yaml run --rm migrate -path=/migrations -database=postgres://hachi:hachi@postgres:5432/hachi?sslmode=disable down 1
