.PHONY: default build install test testacc kanidm-up kanidm-down generate clean fmt lint help

default: build

# Build the provider
build:
	go build -o terraform-provider-kanidm

# Install the provider locally for testing (OpenTofu dev_override)
install: build
	mkdir -p ~/.local/share/opentofu/plugins/registry.opentofu.org/tak-labo/kanidm/0.1.0/darwin_arm64/
	cp terraform-provider-kanidm ~/.local/share/opentofu/plugins/registry.opentofu.org/tak-labo/kanidm/0.1.0/darwin_arm64/

# Run unit tests
test:
	go test -v ./...

# Run acceptance tests from host (requires source .env.test first, or set KANIDM_URL/KANIDM_TOKEN)
testacc:
	TF_ACC=1 TF_ACC_PROVIDER_HOST=registry.opentofu.org TF_ACC_PROVIDER_NAMESPACE=tak-labo go test -v -timeout 30m ./internal/provider/

# Run acceptance tests inside Docker (matches production: separate kanidm + test containers)
testacc-docker:
	docker compose --profile test up --abort-on-container-exit --exit-code-from test

# Start Kanidm in Docker and initialize for acceptance testing
kanidm-up:
	@mkdir -p testdata/kanidm
	@if [ ! -f testdata/kanidm/cert.pem ]; then \
		echo "Generating TLS certificate..."; \
		openssl req -x509 -newkey rsa:2048 -nodes \
			-keyout testdata/kanidm/key.pem \
			-out testdata/kanidm/cert.pem \
			-days 365 \
			-subj "/CN=localhost" \
			-addext "subjectAltName=DNS:localhost,IP:127.0.0.1" 2>/dev/null; \
	fi
	docker compose up -d kanidm
	./scripts/setup-kanidm.sh

# Stop Kanidm and remove volumes
kanidm-down:
	docker compose down -v
	rm -f testdata/kanidm/cert.pem testdata/kanidm/key.pem .env.test

# Generate provider code from OpenAPI schema
generate:
	@echo "Generating provider code from OpenAPI schema..."
	tfplugingen-openapi generate \
		--config internal/spec/generator_config.yml \
		--output internal/spec/provider_code_spec.json \
		internal/spec/kanidm-openapi.json
	tfplugingen-framework generate all \
		--input internal/spec/provider_code_spec.json \
		--output internal/provider

# Generate documentation
docs:
	tfplugindocs generate --provider-name kanidm

# Format code
fmt:
	go fmt ./...

# Run linter
lint:
	golangci-lint run

# Clean build artifacts
clean:
	rm -f terraform-provider-kanidm
	rm -f internal/spec/provider_code_spec.json
	rm -rf dist/

# Show help
help:
	@echo "Available targets:"
	@echo "  build      - Build the provider binary"
	@echo "  install    - Install the provider locally for testing"
	@echo "  test       - Run unit tests"
	@echo "  testacc    - Run acceptance tests (requires KANIDM_URL and KANIDM_TOKEN)"
	@echo "  generate   - Regenerate provider code from OpenAPI schema"
	@echo "  docs       - Generate documentation"
	@echo "  fmt        - Format code"
	@echo "  lint       - Run linter"
	@echo "  clean      - Remove build artifacts"
	@echo "  help       - Show this help message"
