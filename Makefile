PROTOC ?= /usr/bin/protoc
GO ?= go
PROTOC_GEN_GO ?= $(shell $(GO) env GOPATH 2>/dev/null || printf '%s/go' "$$HOME")/bin/protoc-gen-go
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
	$(GO) install google.golang.org/protobuf/cmd/protoc-gen-go@$(PROTOC_GEN_GO_VERSION)

.PHONY: deploy
deploy:
	docker compose up -d --build --force-recreate --remove-orphans

.PHONY: deploy-remote
deploy-remote:
	rsync -az --delete $(RSYNC_EXCLUDES) ./ $(DEPLOY_HOST):$(DEPLOY_DIR)/
	ssh $(DEPLOY_HOST) 'cd $(DEPLOY_DIR) && docker compose up -d --build --force-recreate --remove-orphans'

KUKURUZKA_ESC_SCRIPT ?= ./scripts/kukuruzka-esc.sh
KUKURUZKA_ESC_PATH ?= ../kukuruzka-esc
KUKURUZKA_ESC_REF ?= latest
KUKURUZKA_ESC_INTERVAL ?= 10
KUKURUZKA_ESC_PUBLISH_REF ?=
KUKURUZKA_ESC_MESSAGE ?=

.PHONY: ecs-help
ecs-help:
	$(KUKURUZKA_ESC_SCRIPT) help

.PHONY: ecs-local
ecs-local:
	$(KUKURUZKA_ESC_SCRIPT) local "$(KUKURUZKA_ESC_PATH)"

.PHONY: ecs-update
ecs-update:
	$(KUKURUZKA_ESC_SCRIPT) update "$(KUKURUZKA_ESC_REF)"

.PHONY: ecs-watch-update
ecs-watch-update:
	$(KUKURUZKA_ESC_SCRIPT) watch-update "$(KUKURUZKA_ESC_REF)" "$(KUKURUZKA_ESC_INTERVAL)"

.PHONY: ecs-publish
ecs-publish:
	@test -n "$(KUKURUZKA_ESC_MESSAGE)" || (echo 'Set KUKURUZKA_ESC_MESSAGE="..."'; exit 1)
	$(KUKURUZKA_ESC_SCRIPT) publish "$(KUKURUZKA_ESC_PATH)" "$(KUKURUZKA_ESC_MESSAGE)" "$(KUKURUZKA_ESC_PUBLISH_REF)"
