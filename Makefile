PROTO_DIRS=pkg/user/proto pkg/event/proto pkg/notification/proto
OUT_DIR=.

PROTOC_GEN_GO?=protoc-gen-go
PROTOC_GEN_GRPC?=protoc-gen-go-grpc

.PHONY: proto
proto:
	@for dir in $(PROTO_DIRS); do \
	  for f in $$dir/*.proto; do \
	    echo "Generating $$f"; \
	    protoc -I . \
	      --go_out=. --go_opt=paths=source_relative \
	      --go-grpc_out=. --go-grpc_opt=paths=source_relative \
	      $$f; \
	  done; \
	done

.PHONY: build
build:
	@echo "Building services"
	@go build -o bin/user-service ./cmd/user-service
	@go build -o bin/event-service ./cmd/event-service
	@go build -o bin/notification-service ./cmd/notification-service
	@go build -o bin/email-service ./cmd/email-service

.PHONY: tidy
tidy:
	go mod tidy
