FROM golang:1.22.2 AS dependencies
WORKDIR /go/src/go-yandex-shortener
COPY go.mod .
COPY go.sum .
RUN go mod download

FROM dependencies AS build
COPY . /go/src/go-yandex-shortener
WORKDIR /go/src/go-yandex-shortener
RUN go build -o ./bin/shortener ./cmd/shortener

FROM debian
WORKDIR /app
COPY --from=build /go/src/go-yandex-shortener/migrations/ /app/migrations/
COPY --from=build /go/src/go-yandex-shortener/bin/shortener /app/
RUN chmod +x /app/*
EXPOSE 8080/tcp
CMD /app/shortener
