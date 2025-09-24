PROTOC_GEN_GO := $(GOPATH)/bin/protoc-gen-go
PROTOC_GEN_GO_GRPC := $(GOPATH)/bin/protoc-gen-go-grpc

.PHONY: proto
proto:
	protoc \
		-I api/proto \
		--go_out=internal/infrastructure/grpc/gen --go_opt=paths=source_relative \
		--go-grpc_out=internal/infrastructure/grpc/gen --go-grpc_opt=paths=source_relative \
		api/proto/users.proto


