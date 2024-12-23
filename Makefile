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

.PHONY: docker
docker:
	@echo "-- building docker container"
	docker build -f build/Dockerfile -t shortener .

.PHONY: docker-run
docker-run:
	@echo "-- starting docker container"
	docker run -it -p 8080:8080 shortener

.PHONY: dcb
dcb:
	@echo "-- starting docker compose"
	docker-compose -f ./deployments/docker-compose.yml up --build

.PHONY: swagger
swagger:
	@echo "-- generating swagger"
	swag init --output ./api/ -g ./internal/app/delivery/http.go

.PHONY: godoc
godoc:
	@echo "-- running godoc server"
	godoc -http=:8000

build-staticlint:
	@echo "-- building staticlint"
	go build -o ./bin/staticlint ./cmd/staticlint

run-staticlint:
	@echo "-- running staticlint"
	./bin/staticlint -test=false -errcheck.exclude errcheck_excludes.txt ./...

test-staticlint:
	@echo "-- testing staticlint"
	go test ./cmd/staticlint/

.PHONY: unit-test
unit-test:
	@echo "-- unit testing"
	go test ./... -short
