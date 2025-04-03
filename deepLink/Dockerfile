FROM golang:1.24.1-alpine AS builder

WORKDIR /app
COPY . .

RUN go mod download
RUN go build -o authgrpc ./cmd/app

FROM gcr.io/distroless/base-debian12

WORKDIR /app


COPY --from=builder /app/authgrpc .

COPY config/config.local.yaml ./config/

EXPOSE 8080

CMD ["./authgrpc", "--config=./config/config.local.yaml"]