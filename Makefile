GO ?= go
BIN_DIR ?= bin
BIN_NAME ?= lm2
BIN := $(BIN_DIR)/$(BIN_NAME)

# make build VERSION=0.1.0 stamps the binaries; DATE defaults to today
VERSION ?=
DATE ?= $(shell date +%Y%m%d)
VERSION_PKG := github.com/msorc/languagemachine2/internal/version
LDFLAGS := $(if $(VERSION),-ldflags "-X $(VERSION_PKG).version=$(VERSION) -X $(VERSION_PKG).date=$(DATE)")

.PHONY: all build lm2n2go lm2n lmhl generate test race bench check ci lint fmt fmt-check generate-check fix vet tidy clean install

all: build lm2n2go lm2n lmhl

build: | $(BIN_DIR)
	$(GO) build $(LDFLAGS) -o $(BIN) ./cmd/lm2

lm2n2go: | $(BIN_DIR)
	$(GO) build $(LDFLAGS) -o $(BIN_DIR)/lm2n2go ./cmd/lm2n2go

lm2n: | $(BIN_DIR)
	$(GO) build $(LDFLAGS) -o $(BIN_DIR)/lm2n ./cmd/lm2n

lmhl: | $(BIN_DIR)
	$(GO) build $(LDFLAGS) -o $(BIN_DIR)/lmhl ./examples/highlight/lmhl

generate:
	$(GO) generate ./...

test:
	$(GO) test ./...

race:
	$(GO) test -race ./...

bench:
	$(GO) test ./internal/machine -run '^$$' -bench . -cpu 1,4,16

# check is what CI runs: formatting, vet, lint, the tests with and without
# the race detector, and generated files up to date with their sources
check: fmt-check vet lint test race generate-check

ci: check

lint:
	golangci-lint run ./...

fmt-check:
	@out=$$(gofmt -l .); if [ -n "$$out" ]; then echo "gofmt needed:"; echo "$$out"; exit 1; fi

generate-check: generate
	git diff --exit-code

fmt:
	$(GO) fmt ./...

fix:
	$(GO) fix ./...

vet:
	$(GO) vet ./...

tidy:
	$(GO) mod tidy

install:
	$(GO) install $(LDFLAGS) ./cmd/lm2 ./cmd/lm2n2go ./cmd/lm2n

$(BIN_DIR):
	mkdir -p $(BIN_DIR)

clean:
	rm -rf $(BIN_DIR)
