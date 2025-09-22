APP := server

.PHONY: build run test fmt vet

build:
	go build -o bin/$(APP) ./...

run:
	go run ./...

test:
	go test ./tests -run ^TestApp$ -v

fmt:
	go fmt ./...

vet:
	go vet ./...
