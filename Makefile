.PHONY: build run test test-cover

all: build run

build:
	@echo "-- building"
	go build -o ./cmd/shortener/shortener ./cmd/shortener

run:
	@echo "-- running"
	./cmd/shortener/shortener

test:
	@echo "-- testing"
	go test ./...

test-cover:
	@echo "-- testing with cover"
	go test ./... -coverprofile cover.out && go tool cover -html cover.out
