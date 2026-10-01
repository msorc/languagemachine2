GO ?= go
BIN_DIR ?= bin
BIN_NAME ?= lm
BIN := $(BIN_DIR)/$(BIN_NAME)

.PHONY: all build lmn2go lmn generate run test check fmt fix vet tidy clean install

all: build lmn2go lmn

build: $(BIN_DIR)
	$(GO) build -o $(BIN) ./cmd/lm

lmn2go: $(BIN_DIR)
	$(GO) build -o $(BIN_DIR)/lmn2go ./cmd/lmn2go

lmn: $(BIN_DIR)
	$(GO) build -o $(BIN_DIR)/lmn ./cmd/lmn

generate:
	$(GO) generate ./...

run:
	$(GO) run ./cmd/lm

test:
	$(GO) test ./...

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
	$(GO) install ./cmd/lm ./cmd/lmn2go ./cmd/lmn

$(BIN_DIR):
	mkdir -p $(BIN_DIR)

clean:
	rm -rf $(BIN_DIR)
