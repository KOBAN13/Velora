PROTOC ?= /usr/bin/protoc
PROTOC_GEN_GO ?= $(shell go env GOPATH)/bin/protoc-gen-go

.PHONY: proto
proto:
	$(PROTOC) -I=shared --plugin=protoc-gen-go=$(PROTOC_GEN_GO) --go_out=server shared/packets.proto
