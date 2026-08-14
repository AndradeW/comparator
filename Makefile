.PHONY: build run test vet lint fmt

build:
	go build ./...

run:
	go run ./cmd

test:
	go test ./...

vet:
	go vet ./...

fmt:
	gofmt -l cmd internal

lint:
	golangci-lint run ./... || go vet ./...
