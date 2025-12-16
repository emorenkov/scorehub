# syntax=docker/dockerfile:1

ARG GO_VERSION=1.25.4

FROM golang:${GO_VERSION}-alpine AS builder
WORKDIR /app

# Build arguments
ARG SERVICE_DIR=cmd/user
ARG BINARY_NAME=service

COPY go.mod go.sum ./
COPY vendor ./vendor
COPY cmd ./cmd
COPY pkg ./pkg

RUN apk add --no-cache ca-certificates && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod=vendor -o /out/${BINARY_NAME} ./${SERVICE_DIR}

FROM alpine:3.18
RUN apk add --no-cache ca-certificates
WORKDIR /app

ARG BINARY_NAME=service
COPY --from=builder /out/${BINARY_NAME} /app/service

EXPOSE 8080 50051 50052
ENTRYPOINT ["/app/service"]
