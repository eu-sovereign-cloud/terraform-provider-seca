GO := go
TOOLS_GOMOD := -modfile=./tools/go.mod
GO_TOOL := $(GO) run $(TOOLS_GOMOD) -mod=mod

TF_CODEGEN_SPEC_DIR := ./config/codegen/
TF_CODEGEN_GENERATED_DIR := ./pkg/terraform

.PHONY: build
build:
	@echo "Building..."
	go build -v ./...

.PHONY: install
install: build
	@echo "Installing..."
	go install -v ./...

.PHONY: update
update:
	@echo "Updating submodules..."
	git pull --recurse-submodules
	git submodule update --remote --recursive

.PHONY: lint
lint:
	@echo "Linting..."
	$(GO_TOOL) github.com/golangci/golangci-lint/v2/cmd/golangci-lint run --verbose -c .golangci.yml

.PHONY: generate
generate:
	@echo "Generating documentation..."
	cd tools; go generate ./...

.PHONY: codegen
codegen:
	@for spec in $(TF_CODEGEN_SPEC_DIR)*.json; do \
		name=$$(basename $$spec .json); \
		pkg=$$(echo $$name | tr -cd '[:alnum:]'); \
		echo "Generating Terraform code from $$spec..."; \
		mkdir -p $(TF_CODEGEN_GENERATED_DIR)/$$name; \
		$(GO_TOOL) github.com/hashicorp/terraform-plugin-codegen-framework/cmd/tfplugingen-framework generate resources \
			--input $$spec \
			--output $(TF_CODEGEN_GENERATED_DIR)/$$name || exit 1; \
		$(GO_TOOL) github.com/hashicorp/terraform-plugin-codegen-framework/cmd/tfplugingen-framework generate data-sources \
			--input $$spec \
			--output $(TF_CODEGEN_GENERATED_DIR)/$$name || exit 1; \
	done

.PHONY: clean
clean:
	@echo "Cleaning..."
	rm -rf $(TF_CODEGEN_GENERATED_DIR)/*

.PHONY: fmt
fmt:
	@echo "Formating..."
	$(GO_TOOL) mvdan.cc/gofumpt -w .

.PHONY: test
test:
	@echo "Running unit tests..."
	go test -v -cover -timeout=120s -parallel=10 ./...

.PHONY: testacc
testacc:
	@echo "Running acceptance tests..."
	TF_ACC=1 go test -v -cover -timeout 120m ./...
