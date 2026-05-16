PROTOC ?= /usr/bin/protoc
PROTOC_GEN_GO ?= $(shell go env GOPATH)/bin/protoc-gen-go
PROTOC_GEN_GO_VERSION ?= v1.36.11
DEPLOY_HOST ?= root@apisfs.ru
DEPLOY_DIR ?= /opt/velora
RSYNC_EXCLUDES := \
	--exclude .git/ \
	--exclude .idea/ \
	--exclude .agents/ \
	--exclude .codex \
	--exclude config.env \
	--exclude '*.log' \
	--exclude tmp/ \
	--exclude vendor/

.PHONY: proto
proto: $(PROTOC_GEN_GO)
	$(PROTOC) -I=shared --plugin=protoc-gen-go=$(PROTOC_GEN_GO) --go_out=server shared/packets.proto

$(PROTOC_GEN_GO):
	go install google.golang.org/protobuf/cmd/protoc-gen-go@$(PROTOC_GEN_GO_VERSION)

.PHONY: deploy
deploy:
	docker compose up -d --build --force-recreate --remove-orphans

.PHONY: deploy-remote
deploy-remote:
	rsync -az --delete $(RSYNC_EXCLUDES) ./ $(DEPLOY_HOST):$(DEPLOY_DIR)/
	ssh $(DEPLOY_HOST) 'cd $(DEPLOY_DIR) && docker compose up -d --build --force-recreate --remove-orphans'
