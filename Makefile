.DEFAULT_GOAL := help

GO       ?= go
GOFLAGS  ?=
PROJECT_NAME ?= memo
LOCAL_BIN = $(CURDIR)/bin

.PHONY: install
install: ## installs dependencies
	@echo "Install required programs"
	GOBIN=$(LOCAL_BIN) $(GO) $(GOFLAG) install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.47.3
	GOBIN=$(LOCAL_BIN) $(GO) $(GOFLAG) install golang.org/x/tools/cmd/goimports@latest
	GOBIN=$(LOCAL_BIN) $(GO) $(GOFLAG) install mvdan.cc/gofumpt@latest
	GOBIN=$(LOCAL_BIN) $(GO) $(GOFLAG) get -v github.com/incu6us/goimports-reviser

.PHONY: format
format: ## formats the code and also imports order
	@echo "Formatting..."
	$(LOCAL_BIN)/gofumpt -l -w -extra .
	@echo "Formatting imports..."
	@for f in $$(find . -name '*.go'); do \
		$(LOCAL_BIN)/goimports-reviser -file-path $$f -project-name $(PROJECT_NAME); \
	done

.PHONY: lint
lint: ## lints the code
	@echo "Linting"
	$(LOCAL_BIN)/golangci-lint run --fix

.PHONY: install-githooks
install-githooks: ## installs all git hooks
	@echo "Installing githooks"
	cp ./githooks/* .git/hooks/

.PHONY: build
build: ## builds all commands
	$(GO) $(GOFLAG) build -o $(LOCAL_BIN)/golint-derefnil ./cmd/golint-derefnil

.PHONY: test
test: ## runs tests
	@echo "Testing"
	$(GO) $(GOFLAG) test ./...

.PHONY: help
help:
	@grep --no-filename -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
