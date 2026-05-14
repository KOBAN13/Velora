PROTOC ?= /usr/bin/protoc
PROTOC_GEN_GO ?= $(shell go env GOPATH)/bin/protoc-gen-go
PROTOC_GEN_GO_VERSION ?= v1.36.11

.PHONY: proto
proto: $(PROTOC_GEN_GO)
	$(PROTOC) -I=shared --plugin=protoc-gen-go=$(PROTOC_GEN_GO) --go_out=server shared/packets.proto

$(PROTOC_GEN_GO):
	go install google.golang.org/protobuf/cmd/protoc-gen-go@$(PROTOC_GEN_GO_VERSION)
