FROM golang:1.24.4-alpine AS builder

RUN apk add --no-cache git bash curl


RUN go install github.com/pressly/goose/v3/cmd/goose@latest

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .


RUN go build -o ./bin/server ./cmd/server
RUN go build -o ./bin/grpc-server ./cmd/grpc-server


FROM alpine:3.18
RUN apk add --no-cache ca-certificates bash

WORKDIR /app
COPY --from=builder /app/bin ./bin
COPY .env .env

EXPOSE 8080 9090
