GO ?= go
BIN_DIR ?= bin
BIN_NAME ?= lm
BIN := $(BIN_DIR)/$(BIN_NAME)

.PHONY: all build run test fmt vet tidy clean install

all: build

build: $(BIN_DIR)
	$(GO) build -o $(BIN) ./cmd/lm

run:
	$(GO) run ./cmd/lm

test:
	$(GO) test ./...

fmt:
	$(GO) fmt ./...

vet:
	$(GO) vet ./...

tidy:
	$(GO) mod tidy

install:
	$(GO) install ./cmd/lm

$(BIN_DIR):
	mkdir -p $(BIN_DIR)

clean:
	rm -rf $(BIN_DIR)
