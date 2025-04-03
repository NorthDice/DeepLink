FROM golang:1.24.1-alpine

WORKDIR /app

COPY deepLink/go.mod deepLink/go.sum ./
COPY protos/go.mod protos/go.sum ./

RUN go mod download

COPY deepLink .

RUN CGO_ENABLED=0 GOOS=linux go build -o /deeplink ./cmd/app/main.go

COPY deepLink/config/config.local.yaml /app/config/config.local.yaml

EXPOSE 8080

CMD ["/deeplink", "--config=./config/config.local.yaml"]