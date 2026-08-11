GO := go
TOOLS_GOMOD := -modfile=./tools/go.mod
GO_TOOL := $(GO) run $(TOOLS_GOMOD) -mod=mod

.PHONY: update
update:
	@echo "Updating submodules..."
	git pull --recurse-submodules
	git submodule update --remote --recursive

.PHONY: build
build:
	@echo "Building..."
	go build -v ./...

.PHONY: install
install: build
	@echo "Installing..."
	go install -v ./...

.PHONY: lint
lint:
	@echo "Linting..."
	$(GO_TOOL) github.com/golangci/golangci-lint/v2/cmd/golangci-lint run --verbose -c .golangci.yml

.PHONY: generate
generate:
	@echo "Generating documentation..."
	cd tools; go generate ./...

.PHONY: fmt
fmt:
	@echo "Formating..."
	$(GO_TOOL) mvdan.cc/gofumpt -w .

.PHONY: test
test:
	@echo "Running unit tests..."
	go test -v -cover ./...

.PHONY: acc
acc:
	@echo "Running acceptance tests..."
	TF_ACC=1 TF_DEBUG=1 \
	  SECA_TEST_TOKEN=test \
	  SECA_TEST_TENANT=test-tenant \
	  SECA_TEST_REGION=itbg-bergamo \
	  SECA_TEST_REGION_ENDPOINT=http://172.18.0.2:30080/providers/seca.region \
	  SECA_TEST_AUTH_ENDPOINT=http://172.18.0.2:30080/providers/seca.authorization \
	  go test -v -timeout 60m ./internal/acctest
