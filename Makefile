GO ?= go
BIN_DIR ?= bin
BIN_NAME ?= lm
BIN := $(BIN_DIR)/$(BIN_NAME)

# make build VERSION=0.1.0 stamps the binaries; DATE defaults to today
VERSION ?=
DATE ?= $(shell date +%Y%m%d)
VERSION_PKG := github.com/msorc/languagemachine2/internal/version
LDFLAGS := $(if $(VERSION),-ldflags "-X $(VERSION_PKG).version=$(VERSION) -X $(VERSION_PKG).date=$(DATE)")

.PHONY: all build lmn2go lmn generate run test race bench check fmt fix vet tidy clean install

all: build lmn2go lmn

build: $(BIN_DIR)
	$(GO) build $(LDFLAGS) -o $(BIN) ./cmd/lm

lmn2go: $(BIN_DIR)
	$(GO) build $(LDFLAGS) -o $(BIN_DIR)/lmn2go ./cmd/lmn2go

lmn: $(BIN_DIR)
	$(GO) build $(LDFLAGS) -o $(BIN_DIR)/lmn ./cmd/lmn

generate:
	$(GO) generate ./...

run:
	$(GO) run ./cmd/lm

test:
	$(GO) test ./...

race:
	$(GO) test -race ./...

bench:
	$(GO) test ./internal/machine -run '^$$' -bench . -cpu 1,4,16

check: vet test

fmt:
	$(GO) fmt ./...

fix:
	$(GO) fix ./...

vet:
	$(GO) vet ./...

tidy:
	$(GO) mod tidy

install:
	$(GO) install $(LDFLAGS) ./cmd/lm ./cmd/lmn2go ./cmd/lmn

$(BIN_DIR):
	mkdir -p $(BIN_DIR)

clean:
	rm -rf $(BIN_DIR)
