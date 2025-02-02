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

test-cover-percentage:
	@echo "-- testing with cover percentage"
	go test -v -coverpkg=./... -coverprofile=cover.out -covermode=count ./... && go tool cover -func cover.out

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

mockgen-all: mockgen-main mockgen-app-usecase mockgen-app-delivery mockgen-user-usecase

.PHONY: mockgen-main
mockgen-main:
	@echo "-- mockgen main"
	mockgen -destination=cmd/shortener/mocks/main.go -package=mocks -source=cmd/shortener/main.go

.PHONY: mockgen-app-usecase
mockgen-app-usecase:
	@echo "-- mockgen app usecase"
	mockgen -destination=internal/app/usecase/mocks/usecase.go -package=mocks -source=internal/app/usecase/usecase.go

.PHONY: mockgen-app-delivery
mockgen-app-delivery:
	@echo "-- mockgen app delivery"
	mockgen -destination=internal/app/delivery/mocks/http.go -package=mocks -source=internal/app/delivery/http.go

.PHONY: mockgen-user-usecase
mockgen-user-usecase:
	@echo "-- mockgen user usecase"
	mockgen -destination=internal/user/usecase/mocks/usecase.go -package=mocks -source=internal/user/usecase/usecase.go
	mockgen -destination internal/user/usecase/mocks/grpc_server.go -package mocks google.golang.org/grpc ServerTransportStream

.PHONY: protoc
protoc:
	@echo "-- compiling .proto file"
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. \
	--go-grpc_opt=paths=source_relative api/proto/service.proto;
