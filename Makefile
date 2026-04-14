# Build
build:
	go build -o bin/api-gateway ./cmd/server

run:
	go run ./cmd/server

fmt:
	gofmt -w ./cmd ./internal

tidy:
	go mod tidy
