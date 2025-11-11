
prebuild:
	@go mod vendor && go mod tidy

fmt:
	@go fmt ./...

vet:
	@go vet ./...

build: prebuild fmt vet
	@go build -o bin/goapi

run: build
	@./bin/goapi

run-metrics:
	@docker compose -f ./docker-compose.metrics.yml up

run-db:
	@docker compose -f ./docker-compose.db.yml up

test:
	@go test -v ./...

.PHONY: prebuild fmt vet build run test run-metrics run-db
