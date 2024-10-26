
LDFLAGS=

BUILD_ENVPARMS:=CGO_ENABLED=0

LOCAL_BIN:=$(CURDIR)/bin

##################### GOX #####################
GOX_BIN:=$(LOCAL_BIN)/gox

# local gox
ifeq ($(wildcard $(GOX_BIN)),)
GOX_BIN:=
endif

# Check global bin version
ifneq (, $(shell which gox))
GOX_BIN:=$(shell which gox)
endif

.PHONY: all
all: test build ## default scratch target: test and build

.PHONY: .lint-full
.lint-full: install-lint
	$(GOLANGCI_BIN) run --config=.golangci.yml ./...

.PHONY: .bin-deps
.bin-deps:
	mkdir -p bin
	$(info Installing binary dependencies...)
	GOBIN=$(LOCAL_BIN) go install github.com/mitchellh/gox@v1.0.1  && \
	GOBIN=$(LOCAL_BIN) go install golang.org/x/tools/cmd/goimports@v0.1.9 && \

.PHONY: .deps
.deps:
	$(info Install dependencies...)
	go mod download

.PHONY: update-deps
update-deps: .deps .bin-deps

.PHONY: .test
.test:
	$(info Running tests...)
	go test ./...

.PHONY: test
test: .test ## run unit tests

# CMD_LIST список таргетов (через пробел) которые надо собрать
# можно переопределить в Makefile, по дефолту все из ./cmd кроме основного пакета
# пример переопределения CMD_LIST:= ./cmd/example ./cmd/app ./cmd/cron
ifndef CMD_LIST
CMD_LIST:=$(shell ls ./cmd | sed -e 's/^/.\/cmd\//')
endif
# определение текущий ос
ifndef HOSTOS
HOSTOS:=$(shell go env GOHOSTOS)
endif
# определение текущий архитектуры
ifndef HOSTARCH
HOSTARCH:=$(shell go env GOHOSTARCH)
endif

ifndef BIN_DIR
BIN_DIR=./bin
endif

# если нужно собрать только основной сервис, можно указать в Makefile SINGLE_BUILD=1
DISABLE_CMD_LIST_BUILD?=0

.PHONY: .build
.build:
# сначала собирается основной сервис, скачиваются нужные пакеты и все кладется в кеш для дальнейшего использования
	$(info Building...)
	@if [ -n "$(CMD_LIST)" ] && [ "$(DISABLE_CMD_LIST_BUILD)" != 1 ]; then\
		$(BUILD_ENVPARMS) $(GOX_BIN) -output="$(BIN_DIR)" -osarch="$(HOSTOS)/$(HOSTARCH)" -ldflags "$(LDFLAGS)" $(CMD_LIST);\
	fi

.PHONY: build
build: .build ## build project