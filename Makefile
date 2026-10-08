GO ?= go
BIN_DIR ?= bin
BIN_NAME ?= lm2
BIN := $(BIN_DIR)/$(BIN_NAME)

# make build VERSION=0.1.0 stamps the binaries; DATE defaults to today
VERSION ?=
DATE ?= $(shell date +%Y%m%d)
VERSION_PKG := github.com/msorc/languagemachine2/internal/version
LDFLAGS := $(if $(VERSION),-ldflags "-X $(VERSION_PKG).version=$(VERSION) -X $(VERSION_PKG).date=$(DATE)")

.PHONY: all build lm2n2go lm2n lmhl generate test race bench check ci lint fmt fmt-check generate-check fix vet tidy clean install release release-notes

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

# --- release ------------------------------------------------------------
# make release VERSION=0.2.0 verifies the tree, runs make check, then tags
# v0.2.0, pushes the tag and creates the GitHub release, with the notes
# generated from the commits since the last release tag.
#   REMOTE=origin      the remote to push to (GitHub, for gh)
#   MAIN_BRANCH=master the branch to release from
#   YES=1              skip the confirmation prompt
#   SKIP_CHECKS=1      do not run make check first
REMOTE ?= origin
MAIN_BRANCH ?= master
YES ?=
SKIP_CHECKS ?=
TAG = v$(VERSION)
PREV_TAG := $(shell git describe --tags --abbrev=0 --match 'v*' 2>/dev/null)

# release-notes prints the changelog of the commits since the last release tag
release-notes:
	@git log --no-merges --pretty=format:'- %s (%h)' $(if $(PREV_TAG),$(PREV_TAG)..)HEAD; echo

release:
	@set -e; \
	if [ -z "$(VERSION)" ]; then echo "usage: make release VERSION=x.y.z [YES=1] [SKIP_CHECKS=1]" >&2; exit 2; fi; \
	echo "$(VERSION)" | grep -Eq '^[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z.-]+)?$$' || { echo "VERSION '$(VERSION)' is not x.y.z[-prerelease]" >&2; exit 2; }; \
	command -v gh >/dev/null || { echo "the GitHub CLI (gh) is required: https://cli.github.com" >&2; exit 2; }; \
	gh auth status >/dev/null 2>&1 || { echo "gh is not authenticated: run gh auth login" >&2; exit 2; }; \
	test -z "$$(git status --porcelain)" || { echo "the working tree is not clean: commit or stash first" >&2; exit 2; }; \
	test "$$(git rev-parse --abbrev-ref HEAD)" = "$(MAIN_BRANCH)" || { echo "release from $(MAIN_BRANCH), not $$(git rev-parse --abbrev-ref HEAD)" >&2; exit 2; }; \
	built=$$(sed -n 's/^[[:space:]]*release[[:space:]]*=[[:space:]]*"\([^"]*\)".*/\1/p' internal/version/version.go); \
	test "$$built" = "$(VERSION)" || { echo "internal/version/version.go still says release $$built: bump it to $(VERSION) and commit first" >&2; exit 2; }; \
	git fetch --quiet $(REMOTE); \
	test "$$(git rev-parse HEAD)" = "$$(git rev-parse $(REMOTE)/$(MAIN_BRANCH))" || { echo "HEAD differs from $(REMOTE)/$(MAIN_BRANCH): push or pull first" >&2; exit 2; }; \
	if git rev-parse --quiet --verify "refs/tags/$(TAG)" >/dev/null; then echo "$(TAG) already exists locally" >&2; exit 2; fi; \
	if git ls-remote --exit-code --tags $(REMOTE) "refs/tags/$(TAG)" >/dev/null 2>&1; then echo "$(TAG) already exists on $(REMOTE)" >&2; exit 2; fi
ifneq ($(SKIP_CHECKS),)
	@echo "SKIP_CHECKS: not running make check"
else
	@$(MAKE) --no-print-directory check
endif
	@set -e; \
	notes=$$(mktemp); trap 'rm -f "$$notes"' EXIT; \
	{ \
		echo "## What's changed since $(if $(PREV_TAG),$(PREV_TAG),the first release)"; \
		echo; \
		git log --no-merges --pretty=format:'- %s (%h)' $(if $(PREV_TAG),$(PREV_TAG)..)HEAD; \
		echo; \
	} > "$$notes"; \
	echo; echo "Release notes for $(TAG):"; echo; cat "$$notes"; \
	if [ -z "$(YES)" ]; then printf 'Tag $(TAG), push it to $(REMOTE) and create the GitHub release? [y/N] '; read -r answer; case "$$answer" in [yY]) ;; *) echo aborted; exit 1;; esac; fi; \
	pre=""; case "$(VERSION)" in *-*) pre="--prerelease";; esac; \
	git tag -a $(TAG) -m "Release $(VERSION)"; \
	git push $(REMOTE) $(TAG); \
	gh release create $(TAG) --repo "$$(git remote get-url $(REMOTE))" --title "$(TAG)" --notes-file "$$notes" --target "$$(git rev-parse HEAD)" --verify-tag $$pre
