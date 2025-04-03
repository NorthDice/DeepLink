FROM golang:1.24.1-alpine AS builder

WORKDIR /app

COPY deepLink/go.mod deepLink/go.sum ./
RUN go mod download


COPY deepLink/ .
RUN go build -o /app/authgrpc ./cmd/app

FROM gcr.io/distroless/base-debian12

WORKDIR /app
COPY --from=builder /app/authgrpc .
COPY deepLink/config/config.local.yaml ./config/

EXPOSE 44044
CMD ["./authgrpc", "--config=./config/config.local.yaml"]