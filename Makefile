
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

test:
	@go test -v ./...

.PHONY: prebuild fmt vet build run test
