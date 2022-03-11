include tools/tools.mk

build: deps lint
	go build -o ./bin/ cmd/main.go

docs:
	@ mkdir -p ./internal/docs
	@ go run ./hack/tools/docs/docs.go > ./internal/docs/generated.go
	@ go run ./internal/docs/generated.go
	@ rm -rf ./internal/docs

clean-tools:
	@-rm -rf $(TOOLS_BIN_DIR)

lint: $(GOLANGCI_LINT)
	@ $(GOLANGCI_LINT) run --config=.golangci.yaml --fix

deps:
	go mod tidy