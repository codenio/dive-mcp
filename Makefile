BINARY     := dive-mcp
CMD        := ./cmd/dive-mcp
BIN_DIR    := bin
VERSION    := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS    := -s -w -X main.version=$(VERSION)

.PHONY: all
all: build

.PHONY: build
build:
	mkdir -p $(BIN_DIR)
	go build -ldflags "-X main.version=$(VERSION)" -o $(BIN_DIR)/$(BINARY) $(CMD)

.PHONY: install
install:
	go install -ldflags "-X main.version=$(VERSION)" $(CMD)

.PHONY: run
run:
	go run $(CMD)

.PHONY: test
test:
	go test ./...

.PHONY: lint
lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not found, falling back to go vet"; \
		go vet ./...; \
	fi

.PHONY: fmt
fmt:
	gofmt -w .
	go fmt ./...

.PHONY: fmt-check
fmt-check:
	@unformatted=$$(gofmt -l .); \
	if [ -n "$$unformatted" ]; then \
		echo "These files are not gofmt-formatted:"; \
		echo "$$unformatted"; \
		echo "Run: make fmt"; \
		exit 1; \
	fi

.PHONY: pre-commit
pre-commit:
	@command -v pre-commit >/dev/null 2>&1 || { \
		echo "pre-commit not found; install: https://pre-commit.com/#install"; \
		exit 1; \
	}
	pre-commit run --all-files

.PHONY: hooks-install
hooks-install:
	@command -v pre-commit >/dev/null 2>&1 || { \
		echo "pre-commit not found; install: https://pre-commit.com/#install"; \
		exit 1; \
	}
	@git config --unset-all core.hooksPath 2>/dev/null || true
	pre-commit install
	@echo "Installed pre-commit hooks from .pre-commit-config.yaml"

.PHONY: docs
docs:
	@command -v mkdocs >/dev/null 2>&1 || { \
		echo "mkdocs not found; install: pip install -r requirements-docs.txt"; \
		exit 1; \
	}
	mkdocs serve

.PHONY: docs-build
docs-build:
	@command -v mkdocs >/dev/null 2>&1 || { \
		echo "mkdocs not found; install: pip install -r requirements-docs.txt"; \
		exit 1; \
	}
	mkdocs build --strict

.PHONY: clean
clean:
	rm -rf $(BIN_DIR) dist site

.PHONY: release
release:
	mkdir -p dist
	GOOS=darwin  GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o dist/$(BINARY)-darwin-arm64 $(CMD)
	GOOS=darwin  GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/$(BINARY)-darwin-amd64 $(CMD)
	GOOS=linux   GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/$(BINARY)-linux-amd64  $(CMD)
	GOOS=linux   GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o dist/$(BINARY)-linux-arm64  $(CMD)
