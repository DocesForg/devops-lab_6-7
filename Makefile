PROTO_MODULE := github.com/DocesForg/devops-lab_6-7
PROTO_GO_VERSION := v1.36.6
PROTO_GRPC_VERSION := v1.5.1

.PHONY: generate test build

generate:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@$(PROTO_GO_VERSION)
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@$(PROTO_GRPC_VERSION)
	PATH="$(shell go env GOPATH)/bin:$$PATH" protoc --proto_path=. \
		--go_out=. --go_opt=module=$(PROTO_MODULE) \
		--go-grpc_out=. --go-grpc_opt=module=$(PROTO_MODULE) employee.proto

test:
	go test -race ./...

build:
	go build ./...
