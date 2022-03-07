XDG_CACHE_HOME ?= $(HOME)/.cache
TOOLS_BIN_DIR := $(abspath $(XDG_CACHE_HOME)/keeping-docs-in-sync-with-code/bin)
TOOLS_MOD := tools/go.mod

GOLANGCI_LINT := $(TOOLS_BIN_DIR)/golangci-lint

$(TOOLS_BIN_DIR):
	@ mkdir -p $(TOOLS_BIN_DIR)

$(GOLANGCI_LINT): $(TOOLS_BIN_DIR)
	@ GOBIN=$(TOOLS_BIN_DIR) go install -modfile=$(TOOLS_MOD) github.com/golangci/golangci-lint/cmd/golangci-lint
