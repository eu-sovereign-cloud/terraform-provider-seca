GO := go
TOOLS_GOMOD := -modfile=./tools/go.mod
GO_TOOL := $(GO) run $(TOOLS_GOMOD) -mod=mod

WIREMOCK_PATH := $(shell pwd)/wiremock
WIREMOCK_MAPPINGS_PATH := $(WIREMOCK_PATH)/config/mappings

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

.PHONY: mock-run
mock-run:
	@echo "Running mock..."
	docker compose -f "$(WIREMOCK_PATH)/docker-compose.yml" -p seca-terraform-provider up

.PHONY: mock-start
mock-start:
	@echo "Starting mock..."
	docker compose -f "$(WIREMOCK_PATH)/docker-compose.yml" -p seca-terraform-provider up -d

.PHONY: mock-stop
mock-stop:
	@echo "Stopping mock..."
	docker compose -f "$(WIREMOCK_PATH)/docker-compose.yml" -p seca-terraform-provider down

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
	@echo "Formating code..."
	$(GO_TOOL) mvdan.cc/gofumpt -w .
	@echo "Formatting mock mappings..."
	find $(WIREMOCK_MAPPINGS_PATH) -name "*.json" -type f | while read -r file; do \
      jq '.' "$$file" > "$$file.tmp" && mv "$$file.tmp" "$$file"; \
	done	

.PHONY: test
test:
	@echo "Running unit tests..."
	go test -v -cover -timeout=120s -parallel=10 ./...

.PHONY: testacc
testacc:
	@echo "Running acceptance tests..."
	TF_ACC=1 go test -v -cover -timeout 120m ./...
